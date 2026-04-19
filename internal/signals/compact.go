package signals

import "fmt"

// CompactSignal holds the result of analyzing whether a session
// should be compacted.
type CompactSignal struct {
	ShouldCompact bool     `json:"should_compact"`
	Confidence    string   `json:"confidence"` // "high", "medium", "low"
	Reasons       []string `json:"reasons"`
	Score         int      `json:"score"` // 0-100, higher = stronger compact signal

	// How much context could potentially be reclaimed
	EstimatedReclaimable int `json:"estimated_reclaimable"`

	// Suggested focus areas for a targeted compact
	CompactFocus []string `json:"compact_focus,omitempty"`

	KeepItems        []string `json:"keep_items,omitempty"`
	DropItems        []string `json:"drop_items,omitempty"`
	CompactFocusText string   `json:"compact_focus_text,omitempty"`
	FocusProvenance  string   `json:"focus_provenance,omitempty"`
	FocusModel       string   `json:"focus_model,omitempty"`
	EvidenceTurns    []int    `json:"evidence_turns,omitempty"`
}

// CompactInput holds session-level data for the compact detector.
type CompactInput struct {
	// Occupancy
	TokensInUse      int
	MaxContextTokens int

	// Composition by category (category name -> tokens)
	Composition map[string]int

	// Timeline summary
	TurnCount        int
	AlreadyCompacted bool // whether session has been compacted before
	RecentTurnCount  int  // turns in last ~20% of session
	RecentTurnTokens int  // tokens from recent turns
	OlderTurnTokens  int  // tokens from older turns

	// Growth pattern
	MedianDeltaTokens int
	RecentGrowthRate  float64 // avg delta of last 5 turns / median
}

// DetectCompactCandidate analyzes a session and returns a signal
// indicating whether the user should compact.
func DetectCompactCandidate(in CompactInput) CompactSignal {
	if in.MaxContextTokens <= 0 || in.TokensInUse <= 0 {
		return CompactSignal{}
	}

	var reasons []string
	var focus []string
	totalScore := 0

	// Heuristic 1: Occupancy threshold
	if score, reason := checkOccupancy(in); score > 0 {
		totalScore += score
		reasons = append(reasons, reason)
	}

	// Heuristic 2: Projected turns until critical
	if score, reason := checkGrowthProjection(in); score > 0 {
		totalScore += score
		reasons = append(reasons, reason)
	}

	// Heuristic 3: Low-value category ratio
	if score, reason, cats := checkLowValueRatio(in); score > 0 {
		totalScore += score
		reasons = append(reasons, reason)
		focus = append(focus, cats...)
	}

	// Heuristic 4: Many turns without compaction
	if score, reason := checkTurnAccumulation(in); score > 0 {
		totalScore += score
		reasons = append(reasons, reason)
	}

	// Heuristic 5: Stale context ratio
	if score, reason := checkStaleContextRatio(in); score > 0 {
		totalScore += score
		reasons = append(reasons, reason)
	}

	if totalScore > 100 {
		totalScore = 100
	}

	confidence := "low"
	shouldCompact := false
	if totalScore >= 60 {
		confidence = "high"
		shouldCompact = true
	} else if totalScore >= 35 {
		confidence = "medium"
		shouldCompact = true
	} else if totalScore >= 15 {
		confidence = "low"
		shouldCompact = true
	}

	if !shouldCompact {
		return CompactSignal{}
	}

	// Estimate reclaimable tokens: older turn tokens minus a
	// conservative estimate for what the compact summary retains (~20%).
	reclaimable := int(float64(in.OlderTurnTokens) * 0.8)
	if reclaimable < 0 {
		reclaimable = 0
	}

	return CompactSignal{
		ShouldCompact:        true,
		Confidence:           confidence,
		Reasons:              reasons,
		Score:                totalScore,
		EstimatedReclaimable: reclaimable,
		CompactFocus:         focus,
	}
}

// Heuristic 1: Occupancy threshold
// Pure percentage-based: >80% is urgent, >65% is notable.
func checkOccupancy(in CompactInput) (int, string) {
	occupancy := float64(in.TokensInUse) / float64(in.MaxContextTokens)

	if occupancy >= 0.85 {
		return 35, fmt.Sprintf(
			"Critical occupancy: session is %.0f%% full (%dk / %dk tokens)",
			occupancy*100, in.TokensInUse/1000, in.MaxContextTokens/1000,
		)
	}
	if occupancy >= 0.75 {
		return 25, fmt.Sprintf(
			"High occupancy: session is %.0f%% full (%dk / %dk tokens)",
			occupancy*100, in.TokensInUse/1000, in.MaxContextTokens/1000,
		)
	}
	if occupancy >= 0.65 {
		return 15, fmt.Sprintf(
			"Elevated occupancy: session is %.0f%% full (%dk / %dk tokens)",
			occupancy*100, in.TokensInUse/1000, in.MaxContextTokens/1000,
		)
	}
	return 0, ""
}

// Heuristic 2: Projected turns until critical
// At current growth rate, how many turns until 90% full?
func checkGrowthProjection(in CompactInput) (int, string) {
	if in.MedianDeltaTokens <= 0 || in.RecentGrowthRate <= 0 {
		return 0, ""
	}

	remaining := in.MaxContextTokens - in.TokensInUse
	if remaining <= 0 {
		return 20, "Context window is effectively full at current growth rate"
	}

	// Use recent growth rate to project
	avgRecentDelta := float64(in.MedianDeltaTokens) * in.RecentGrowthRate
	if avgRecentDelta <= 0 {
		return 0, ""
	}

	// How many turns until 90% full?
	target90 := int(float64(in.MaxContextTokens) * 0.9)
	remaining90 := target90 - in.TokensInUse
	if remaining90 <= 0 {
		return 20, fmt.Sprintf(
			"Already past 90%% occupancy, compaction overdue",
		)
	}

	turnsUntil90 := float64(remaining90) / avgRecentDelta

	if turnsUntil90 <= 3 {
		return 25, fmt.Sprintf(
			"Projected to hit 90%% in ~%.0f turns at current growth rate",
			turnsUntil90,
		)
	}
	if turnsUntil90 <= 8 {
		return 15, fmt.Sprintf(
			"Projected to hit 90%% in ~%.0f turns at current growth rate",
			turnsUntil90,
		)
	}
	return 0, ""
}

// Heuristic 3: Low-value category ratio
// If tool_outputs + search_results + file_reads are a large
// fraction of total context, compaction can reclaim significant
// space by summarizing those away.
func checkLowValueRatio(in CompactInput) (int, string, []string) {
	if in.TokensInUse <= 0 {
		return 0, "", nil
	}

	lowValueCats := []struct {
		key   string
		label string
	}{
		{"tool_outputs", "tool outputs"},
		{"search_results", "search results"},
		{"file_reads", "file reads"},
	}

	lowValueTotal := 0
	var bigCategories []string
	for _, cat := range lowValueCats {
		tokens := in.Composition[cat.key]
		if tokens <= 0 {
			continue
		}
		catPct := float64(tokens) / float64(in.TokensInUse)
		lowValueTotal += tokens
		if catPct >= 0.15 {
			bigCategories = append(bigCategories, fmt.Sprintf(
				"%s (%.0f%%)", cat.label, catPct*100,
			))
		}
	}

	ratio := float64(lowValueTotal) / float64(in.TokensInUse)

	if ratio >= 0.50 && len(bigCategories) > 0 {
		return 20, fmt.Sprintf(
			"%.0f%% of context is compactable categories: %s",
			ratio*100, joinStrings(bigCategories),
		), bigCategories
	}
	if ratio >= 0.35 && len(bigCategories) > 0 {
		return 10, fmt.Sprintf(
			"%.0f%% of context is compactable categories: %s",
			ratio*100, joinStrings(bigCategories),
		), bigCategories
	}
	return 0, "", nil
}

// Heuristic 4: Many turns without compaction
// More turns = more accumulated intermediate state that
// compaction can collapse.
func checkTurnAccumulation(in CompactInput) (int, string) {
	if in.TurnCount < 10 {
		return 0, ""
	}

	if in.AlreadyCompacted {
		// After a compaction, the bar is lower since we're
		// counting turns since last compaction
		if in.TurnCount >= 30 {
			return 15, fmt.Sprintf(
				"Session has %d turns since last compaction",
				in.TurnCount,
			)
		}
		return 0, ""
	}

	if in.TurnCount >= 40 {
		return 20, fmt.Sprintf(
			"Long session: %d turns with no compaction",
			in.TurnCount,
		)
	}
	if in.TurnCount >= 20 {
		return 10, fmt.Sprintf(
			"Growing session: %d turns with no compaction",
			in.TurnCount,
		)
	}
	return 0, ""
}

// Heuristic 5: Stale context ratio
// If the older turns hold a large share of tokens compared to
// recent turns, most of that older context is likely summarizable.
func checkStaleContextRatio(in CompactInput) (int, string) {
	if in.OlderTurnTokens <= 0 || in.TokensInUse <= 0 {
		return 0, ""
	}

	staleRatio := float64(in.OlderTurnTokens) / float64(in.TokensInUse)

	if staleRatio >= 0.75 {
		return 15, fmt.Sprintf(
			"%.0f%% of context is from older turns that could be summarized",
			staleRatio*100,
		)
	}
	return 0, ""
}

// --- helpers ---

func joinStrings(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if i == len(parts)-1 {
			result += " and " + parts[i]
		} else {
			result += ", " + parts[i]
		}
	}
	return result
}
