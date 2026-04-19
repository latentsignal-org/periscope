package signals

import (
	"fmt"
	"math"
	"strings"
)

// RewindSignal holds the result of analyzing whether the last
// turn is a candidate for rewind.
type RewindSignal struct {
	ShouldRewind      bool     `json:"should_rewind"`
	Confidence        string   `json:"confidence"` // "high", "medium", "low"
	Reasons           []string `json:"reasons"`
	TokensRecoverable int      `json:"tokens_recoverable"`
	Score             int      `json:"score"` // 0-100, higher = stronger rewind signal

	// Rewind target: which turn to rewind to
	RewindToTurn   int    `json:"rewind_to_turn,omitempty"`   // 1-based turn number
	RewindToReason string `json:"rewind_to_reason,omitempty"` // why this turn
	BadStretchFrom int    `json:"bad_stretch_from,omitempty"` // first bad turn
	BadStretchTo   int    `json:"bad_stretch_to,omitempty"`   // last bad turn (usually last)
}

// RewindTurn is the per-turn data needed by the rewind detector.
// Callers build this from the context timeline and message data.
type RewindTurn struct {
	Turn             int
	DeltaTokens      int
	CumulativeTokens int
	DominantCategory string
	Categories       map[string]int // category -> tokens

	// Tool call details for this turn
	ToolCalls []RewindToolCall

	// Whether any Edit/Write tool call succeeded in this turn
	HasSuccessfulEdit bool
}

// RewindToolCall is the minimal tool call info needed for rewind
// heuristics.
type RewindToolCall struct {
	ToolName      string
	Category      string
	InputJSON     string
	ResultContent string
	EventStatus   string
}

// DetectRewindCandidate analyzes the last turn of a session and
// returns a signal indicating whether the user should rewind.
// maxContextTokens is the session's context window capacity (0 if unknown).
func DetectRewindCandidate(
	turns []RewindTurn,
	maxContextTokens int,
) RewindSignal {
	if len(turns) == 0 {
		return RewindSignal{}
	}

	last := turns[len(turns)-1]
	var reasons []string
	totalScore := 0

	// Heuristic 1: Failed tool loop
	if score, reason := checkFailedToolLoop(last); score > 0 {
		totalScore += score
		reasons = append(reasons, reason)
	}

	// Heuristic 2: Edit churn on same file
	if score, reason := checkEditChurn(last); score > 0 {
		totalScore += score
		reasons = append(reasons, reason)
	}

	// Heuristic 3: Large context delta with low value
	if score, reason := checkLargeNoValueDelta(last, turns); score > 0 {
		totalScore += score
		reasons = append(reasons, reason)
	}

	// Heuristic 4: Retry of previous turn
	if len(turns) >= 2 {
		prev := turns[len(turns)-2]
		if score, reason := checkRetryOfPrevious(last, prev); score > 0 {
			totalScore += score
			reasons = append(reasons, reason)
		}
	}

	// Heuristic 6: High context cost relative to remaining budget
	if maxContextTokens > 0 {
		if score, reason := checkHighCostLowBudget(last, maxContextTokens); score > 0 {
			totalScore += score
			reasons = append(reasons, reason)
		}
	}

	if totalScore > 100 {
		totalScore = 100
	}

	confidence := "low"
	shouldRewind := false
	if totalScore >= 60 {
		confidence = "high"
		shouldRewind = true
	} else if totalScore >= 35 {
		confidence = "medium"
		shouldRewind = true
	} else if totalScore >= 15 {
		confidence = "low"
		shouldRewind = true
	}

	if !shouldRewind {
		return RewindSignal{}
	}

	// Find the rewind target: scan backwards to find where the
	// bad stretch started, then recommend rewinding to the turn
	// just before it.
	rewindTo, badFrom, badTo, rewindReason, tokensRecoverable :=
		findRewindTarget(turns, maxContextTokens)

	return RewindSignal{
		ShouldRewind:      true,
		Confidence:        confidence,
		Reasons:           reasons,
		TokensRecoverable: tokensRecoverable,
		Score:             totalScore,
		RewindToTurn:      rewindTo,
		RewindToReason:    rewindReason,
		BadStretchFrom:    badFrom,
		BadStretchTo:      badTo,
	}
}

// findRewindTarget scans backwards from the last turn to find
// the contiguous stretch of "bad" turns at the tail. Returns the
// recommended rewind-to turn (the last clean turn) and the range.
func findRewindTarget(
	turns []RewindTurn, maxContextTokens int,
) (rewindTo, badFrom, badTo int, reason string, tokensRecoverable int) {
	if len(turns) == 0 {
		return 0, 0, 0, "", 0
	}

	// Compute median delta for heuristic 3
	medianDelta := computeMedianDelta(turns)

	// Walk backwards: each turn that triggers any heuristic is
	// part of the bad stretch.
	lastTurn := turns[len(turns)-1]
	badTo = lastTurn.Turn
	badFrom = lastTurn.Turn
	tokensRecoverable = lastTurn.DeltaTokens

	for i := len(turns) - 2; i >= 0; i-- {
		t := turns[i]
		if isTurnBad(t, turns, i, maxContextTokens, medianDelta) {
			badFrom = t.Turn
			tokensRecoverable += t.DeltaTokens
		} else {
			break
		}
	}

	// Rewind to the turn just before the bad stretch
	rewindTo = badFrom - 1
	if rewindTo < 1 {
		rewindTo = 1
	}

	badCount := badTo - badFrom + 1
	if badCount == 1 {
		reason = fmt.Sprintf(
			"Turn %d is the last clean turn before the problematic turn %d",
			rewindTo, badTo,
		)
	} else {
		reason = fmt.Sprintf(
			"Turn %d is the last clean turn before the bad stretch (turns %d–%d, %d turns, ~%dk tokens)",
			rewindTo, badFrom, badTo, badCount, tokensRecoverable/1000,
		)
	}
	return rewindTo, badFrom, badTo, reason, tokensRecoverable
}

// isTurnBad checks if a single turn triggers any rewind heuristic.
// Used when scanning backwards to find where the bad stretch starts.
func isTurnBad(
	turn RewindTurn,
	allTurns []RewindTurn,
	turnIdx int,
	maxContextTokens int,
	medianDelta int,
) bool {
	// H1: Failed tool loop
	if score, _ := checkFailedToolLoop(turn); score > 0 {
		return true
	}
	// H2: Edit churn
	if score, _ := checkEditChurn(turn); score > 0 {
		return true
	}
	// H3: Large low-value delta (use pre-computed median)
	if medianDelta > 0 && turn.DeltaTokens > 0 {
		ratio := float64(turn.DeltaTokens) / float64(medianDelta)
		if ratio >= 2.5 && !turn.HasSuccessfulEdit {
			lowValue := turn.Categories["tool_outputs"] +
				turn.Categories["search_results"] +
				turn.Categories["file_reads"]
			if float64(lowValue)/float64(turn.DeltaTokens) >= 0.6 {
				return true
			}
		}
	}
	// H4: Retry of previous turn
	if turnIdx > 0 {
		prev := allTurns[turnIdx-1]
		if score, _ := checkRetryOfPrevious(turn, prev); score > 0 {
			return true
		}
	}
	// H6: High cost vs budget
	if maxContextTokens > 0 {
		if score, _ := checkHighCostLowBudget(turn, maxContextTokens); score > 0 {
			return true
		}
	}
	return false
}

func computeMedianDelta(turns []RewindTurn) int {
	deltas := make([]int, 0, len(turns))
	for _, t := range turns {
		if t.DeltaTokens > 0 {
			deltas = append(deltas, t.DeltaTokens)
		}
	}
	if len(deltas) < 3 {
		return 0
	}
	sortInts(deltas)
	return deltas[len(deltas)/2]
}

// Heuristic 1: Failed tool loop
// If the majority of tool calls in the last turn failed, it's a
// strong signal the turn added noise without progress.
func checkFailedToolLoop(turn RewindTurn) (int, string) {
	if len(turn.ToolCalls) == 0 {
		return 0, ""
	}

	failures := 0
	consecutive := 0
	maxConsecutive := 0
	for _, tc := range turn.ToolCalls {
		if isRewindFailure(tc) {
			failures++
			consecutive++
			if consecutive > maxConsecutive {
				maxConsecutive = consecutive
			}
		} else {
			consecutive = 0
		}
	}

	total := len(turn.ToolCalls)
	failRate := float64(failures) / float64(total)

	// 3+ consecutive failures is strong
	if maxConsecutive >= 3 {
		return 35, reasonFailedLoop(failures, total, maxConsecutive)
	}
	// >50% failure rate with 2+ failures
	if failures >= 2 && failRate > 0.5 {
		return 25, reasonFailedLoop(failures, total, maxConsecutive)
	}

	return 0, ""
}

func reasonFailedLoop(failures, total, maxConsecutive int) string {
	if maxConsecutive >= 3 {
		return fmt.Sprintf(
			"Failed tool loop: %d of %d tool calls failed (%d consecutive)",
			failures, total, maxConsecutive,
		)
	}
	return fmt.Sprintf(
		"Failed tool loop: %d of %d tool calls failed",
		failures, total,
	)
}

// Heuristic 2: Edit churn on same file
// If the turn edited the same file 2+ times, the agent is likely
// flailing rather than converging.
func checkEditChurn(turn RewindTurn) (int, string) {
	fileCounts := map[string]int{}
	for _, tc := range turn.ToolCalls {
		if tc.Category != "Edit" && tc.Category != "Write" {
			continue
		}
		path := extractFilePathFromJSON(tc.InputJSON)
		if path == "" {
			continue
		}
		fileCounts[path]++
	}

	maxEdits := 0
	maxFile := ""
	for path, count := range fileCounts {
		if count > maxEdits {
			maxEdits = count
			maxFile = path
		}
	}

	if maxEdits >= 3 {
		return 30, fmt.Sprintf(
			"Edit churn: %s was edited %d times in this turn",
			shortPath(maxFile), maxEdits,
		)
	}
	if maxEdits >= 2 {
		return 15, fmt.Sprintf(
			"Edit churn: %s was edited %d times in this turn",
			shortPath(maxFile), maxEdits,
		)
	}
	return 0, ""
}

// Heuristic 3: Large context delta with low value
// A spike in context consumption dominated by tool output / search
// results, with no successful edits, suggests the turn added bulk
// but no durable progress.
func checkLargeNoValueDelta(
	turn RewindTurn, allTurns []RewindTurn,
) (int, string) {
	if len(allTurns) < 3 || turn.DeltaTokens <= 0 {
		return 0, ""
	}

	// Compute median delta across all turns
	deltas := make([]int, 0, len(allTurns))
	for _, t := range allTurns {
		if t.DeltaTokens > 0 {
			deltas = append(deltas, t.DeltaTokens)
		}
	}
	if len(deltas) < 3 {
		return 0, ""
	}
	sortInts(deltas)
	median := deltas[len(deltas)/2]
	if median <= 0 {
		return 0, ""
	}

	ratio := float64(turn.DeltaTokens) / float64(median)
	if ratio < 2.5 {
		return 0, ""
	}

	// Check if the delta is dominated by low-value categories
	lowValue := turn.Categories["tool_outputs"] +
		turn.Categories["search_results"] +
		turn.Categories["file_reads"]
	if turn.DeltaTokens > 0 &&
		float64(lowValue)/float64(turn.DeltaTokens) < 0.6 {
		return 0, ""
	}

	// If the turn also produced successful edits, it's less
	// clear-cut — the bulk might have been necessary research
	if turn.HasSuccessfulEdit {
		return 0, ""
	}

	score := int(math.Min(30, ratio*8))
	return score, fmt.Sprintf(
		"Large low-value delta: this turn added %dk tokens (%.1fx median), dominated by %s with no successful edits",
		turn.DeltaTokens/1000, ratio, turn.DominantCategory,
	)
}

// Heuristic 4: Retry of previous turn
// If tool call signatures in the last turn heavily overlap with
// the previous turn, the agent is retrying the same approach.
func checkRetryOfPrevious(
	current, previous RewindTurn,
) (int, string) {
	if len(current.ToolCalls) == 0 || len(previous.ToolCalls) == 0 {
		return 0, ""
	}

	// Build signature set for previous turn
	prevSigs := map[string]struct{}{}
	for _, tc := range previous.ToolCalls {
		prevSigs[toolSignature(tc)] = struct{}{}
	}

	// Count how many current tool calls match
	matched := 0
	for _, tc := range current.ToolCalls {
		if _, ok := prevSigs[toolSignature(tc)]; ok {
			matched++
		}
	}

	overlapRate := float64(matched) / float64(len(current.ToolCalls))

	if overlapRate >= 0.6 && matched >= 2 {
		score := 25
		if overlapRate >= 0.8 {
			score = 35
		}
		return score, fmt.Sprintf(
			"Retry detected: %d of %d tool calls repeat the previous turn (%.0f%% overlap)",
			matched, len(current.ToolCalls), overlapRate*100,
		)
	}
	return 0, ""
}

// Heuristic 6: High context cost relative to remaining budget
// If the turn consumed >10% of the total context window while
// the session is already >60% full, the turn is expensive.
func checkHighCostLowBudget(
	turn RewindTurn, maxTokens int,
) (int, string) {
	if maxTokens <= 0 {
		return 0, ""
	}

	occupancy := float64(turn.CumulativeTokens) / float64(maxTokens)
	turnCost := float64(turn.DeltaTokens) / float64(maxTokens)

	if occupancy < 0.6 || turnCost < 0.10 {
		return 0, ""
	}

	score := 0
	if turnCost >= 0.20 && occupancy >= 0.75 {
		score = 30
	} else if turnCost >= 0.15 && occupancy >= 0.70 {
		score = 20
	} else if turnCost >= 0.10 && occupancy >= 0.60 {
		score = 10
	}

	if score == 0 {
		return 0, ""
	}

	return score, fmt.Sprintf(
		"High cost vs budget: this turn used %.0f%% of context window, session is now %.0f%% full",
		turnCost*100, occupancy*100,
	)
}

// --- helpers ---

func isRewindFailure(tc RewindToolCall) bool {
	if tc.EventStatus != "" {
		return tc.EventStatus == "errored" ||
			tc.EventStatus == "cancelled"
	}
	return isContentFailure(tc.Category, tc.ResultContent)
}

// toolSignature produces a coarse identity for a tool call:
// tool name + the first 200 chars of input. This is enough to
// detect retries without exact-match sensitivity.
func toolSignature(tc RewindToolCall) string {
	input := tc.InputJSON
	if len(input) > 200 {
		input = input[:200]
	}
	return tc.ToolName + "|" + input
}

func extractFilePathFromJSON(input string) string {
	marker := `"file_path":"`
	idx := strings.Index(input, marker)
	if idx < 0 {
		return ""
	}
	start := idx + len(marker)
	end := strings.Index(input[start:], `"`)
	if end < 0 {
		return ""
	}
	return input[start : start+end]
}

func shortPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) <= 3 {
		return path
	}
	return "…/" + strings.Join(parts[len(parts)-2:], "/")
}

func sortInts(a []int) {
	// Simple insertion sort — turn counts are small
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j] < a[j-1]; j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}

