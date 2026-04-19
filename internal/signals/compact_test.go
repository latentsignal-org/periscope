package signals

import "testing"

func TestDetectCompactCandidate_NoCapacity(t *testing.T) {
	sig := DetectCompactCandidate(CompactInput{})
	if sig.ShouldCompact {
		t.Fatal("expected no compact for empty input")
	}
}

func TestDetectCompactCandidate_LowOccupancy(t *testing.T) {
	sig := DetectCompactCandidate(CompactInput{
		TokensInUse:      20_000,
		MaxContextTokens: 200_000,
		TurnCount:        5,
	})
	if sig.ShouldCompact {
		t.Fatalf("expected no compact at 10%% occupancy, got score=%d reasons=%v",
			sig.Score, sig.Reasons)
	}
}

func TestDetectCompactCandidate_HighOccupancy(t *testing.T) {
	sig := DetectCompactCandidate(CompactInput{
		TokensInUse:      170_000,
		MaxContextTokens: 200_000,
		TurnCount:        25,
		Composition: map[string]int{
			"tool_outputs":    60_000,
			"search_results":  30_000,
			"assistant_messages": 50_000,
			"user_messages":   30_000,
		},
		OlderTurnTokens: 130_000,
	})
	if !sig.ShouldCompact {
		t.Fatal("expected compact at 85% occupancy")
	}
	if sig.Confidence != "high" {
		t.Fatalf("expected high confidence, got %s (score=%d)", sig.Confidence, sig.Score)
	}
}

func TestDetectCompactCandidate_GrowthProjection(t *testing.T) {
	sig := DetectCompactCandidate(CompactInput{
		TokensInUse:       150_000,
		MaxContextTokens:  200_000,
		TurnCount:         15,
		MedianDeltaTokens: 5_000,
		RecentGrowthRate:  1.5, // recent turns are 1.5x median
		OlderTurnTokens:   100_000,
	})
	if !sig.ShouldCompact {
		t.Fatal("expected compact with fast growth projection")
	}
	found := false
	for _, r := range sig.Reasons {
		if contains(r, "Projected") || contains(r, "occupancy") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected growth projection reason, got %v", sig.Reasons)
	}
}

func TestDetectCompactCandidate_LowValueCategories(t *testing.T) {
	sig := DetectCompactCandidate(CompactInput{
		TokensInUse:      140_000,
		MaxContextTokens: 200_000,
		TurnCount:        20,
		Composition: map[string]int{
			"tool_outputs":       50_000,
			"search_results":     30_000,
			"file_reads":         20_000,
			"assistant_messages": 25_000,
			"user_messages":      15_000,
		},
		OlderTurnTokens: 100_000,
	})
	if !sig.ShouldCompact {
		t.Fatal("expected compact with high low-value ratio")
	}
	found := false
	for _, r := range sig.Reasons {
		if contains(r, "compactable categories") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected low-value category reason, got %v", sig.Reasons)
	}
	if len(sig.CompactFocus) == 0 {
		t.Fatal("expected compact focus suggestions")
	}
}

func TestDetectCompactCandidate_ManyTurns(t *testing.T) {
	sig := DetectCompactCandidate(CompactInput{
		TokensInUse:      130_000,
		MaxContextTokens: 200_000,
		TurnCount:        45,
		OlderTurnTokens:  100_000,
	})
	if !sig.ShouldCompact {
		t.Fatal("expected compact for 45 turns")
	}
	found := false
	for _, r := range sig.Reasons {
		if contains(r, "turns") && contains(r, "compaction") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected turn accumulation reason, got %v", sig.Reasons)
	}
}

func TestDetectCompactCandidate_StaleContext(t *testing.T) {
	sig := DetectCompactCandidate(CompactInput{
		TokensInUse:      140_000,
		MaxContextTokens: 200_000,
		TurnCount:        20,
		OlderTurnTokens:  120_000,
	})
	if !sig.ShouldCompact {
		t.Fatal("expected compact with high stale ratio")
	}
	found := false
	for _, r := range sig.Reasons {
		if contains(r, "older turns") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected stale context reason, got %v", sig.Reasons)
	}
}

func TestDetectCompactCandidate_AlreadyCompacted(t *testing.T) {
	sig := DetectCompactCandidate(CompactInput{
		TokensInUse:      50_000,
		MaxContextTokens: 200_000,
		TurnCount:        8,
		AlreadyCompacted: true,
	})
	if sig.ShouldCompact {
		t.Fatalf("expected no compact for recently compacted session with few turns, got score=%d",
			sig.Score)
	}
}

func TestDetectCompactCandidate_Reclaimable(t *testing.T) {
	sig := DetectCompactCandidate(CompactInput{
		TokensInUse:      170_000,
		MaxContextTokens: 200_000,
		TurnCount:        30,
		OlderTurnTokens:  130_000,
	})
	if !sig.ShouldCompact {
		t.Fatal("expected compact")
	}
	// 80% of 130k = 104k
	expected := int(float64(130_000) * 0.8)
	if sig.EstimatedReclaimable != expected {
		t.Fatalf("expected reclaimable ~%d, got %d", expected, sig.EstimatedReclaimable)
	}
}
