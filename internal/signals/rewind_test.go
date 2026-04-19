package signals

import (
	"testing"
)

func TestDetectRewindCandidate_EmptyTurns(t *testing.T) {
	sig := DetectRewindCandidate(nil, 200_000)
	if sig.ShouldRewind {
		t.Fatal("expected no rewind for empty turns")
	}
}

func TestDetectRewindCandidate_FailedToolLoop(t *testing.T) {
	turns := []RewindTurn{
		{Turn: 1, DeltaTokens: 1000, CumulativeTokens: 1000},
		{Turn: 2, DeltaTokens: 1000, CumulativeTokens: 2000},
		{Turn: 3, DeltaTokens: 1000, CumulativeTokens: 3000},
		{
			Turn: 4, DeltaTokens: 5000, CumulativeTokens: 8000,
			ToolCalls: []RewindToolCall{
				{ToolName: "Bash", Category: "Bash", EventStatus: "errored"},
				{ToolName: "Bash", Category: "Bash", EventStatus: "errored"},
				{ToolName: "Bash", Category: "Bash", EventStatus: "errored"},
				{ToolName: "Edit", Category: "Edit", EventStatus: "errored"},
			},
		},
	}
	sig := DetectRewindCandidate(turns, 200_000)
	if !sig.ShouldRewind {
		t.Fatal("expected rewind for failed tool loop")
	}
	if sig.Confidence != "medium" && sig.Confidence != "high" {
		t.Fatalf("expected medium or high confidence, got %s", sig.Confidence)
	}
	if len(sig.Reasons) == 0 {
		t.Fatal("expected at least one reason")
	}
	if sig.RewindToTurn != 3 {
		t.Fatalf("expected rewind to turn 3, got %d", sig.RewindToTurn)
	}
	if sig.BadStretchFrom != 4 || sig.BadStretchTo != 4 {
		t.Fatalf("expected bad stretch 4-4, got %d-%d", sig.BadStretchFrom, sig.BadStretchTo)
	}
}

func TestDetectRewindCandidate_EditChurn(t *testing.T) {
	turns := []RewindTurn{
		{Turn: 1, DeltaTokens: 1000, CumulativeTokens: 1000},
		{Turn: 2, DeltaTokens: 1000, CumulativeTokens: 2000},
		{Turn: 3, DeltaTokens: 1000, CumulativeTokens: 3000},
		{
			Turn: 4, DeltaTokens: 2000, CumulativeTokens: 5000,
			ToolCalls: []RewindToolCall{
				{ToolName: "Edit", Category: "Edit", InputJSON: `{"file_path":"/src/main.go","old_string":"a","new_string":"b"}`},
				{ToolName: "Edit", Category: "Edit", InputJSON: `{"file_path":"/src/main.go","old_string":"b","new_string":"c"}`},
				{ToolName: "Edit", Category: "Edit", InputJSON: `{"file_path":"/src/main.go","old_string":"c","new_string":"d"}`},
			},
		},
	}
	sig := DetectRewindCandidate(turns, 200_000)
	if !sig.ShouldRewind {
		t.Fatal("expected rewind for edit churn")
	}
	found := false
	for _, r := range sig.Reasons {
		if contains(r, "Edit churn") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected edit churn reason, got %v", sig.Reasons)
	}
}

func TestDetectRewindCandidate_LargeLowValueDelta(t *testing.T) {
	turns := []RewindTurn{
		{Turn: 1, DeltaTokens: 1000, CumulativeTokens: 1000},
		{Turn: 2, DeltaTokens: 1200, CumulativeTokens: 2200},
		{Turn: 3, DeltaTokens: 800, CumulativeTokens: 3000},
		{
			Turn: 4, DeltaTokens: 15000, CumulativeTokens: 18000,
			DominantCategory: "search_results",
			Categories: map[string]int{
				"search_results": 12000,
				"tool_outputs":   2000,
				"assistant_messages": 1000,
			},
			HasSuccessfulEdit: false,
		},
	}
	sig := DetectRewindCandidate(turns, 200_000)
	if !sig.ShouldRewind {
		t.Fatal("expected rewind for large low-value delta")
	}
	found := false
	for _, r := range sig.Reasons {
		if contains(r, "Large low-value delta") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected large delta reason, got %v", sig.Reasons)
	}
}

func TestDetectRewindCandidate_RetryOfPrevious(t *testing.T) {
	sharedCalls := []RewindToolCall{
		{ToolName: "Bash", Category: "Bash", InputJSON: `{"command":"go test ./..."}`},
		{ToolName: "Edit", Category: "Edit", InputJSON: `{"file_path":"/src/main.go","old_string":"x"}`},
		{ToolName: "Bash", Category: "Bash", InputJSON: `{"command":"go test ./..."}`},
	}
	turns := []RewindTurn{
		{Turn: 1, DeltaTokens: 1000, CumulativeTokens: 1000},
		{Turn: 2, DeltaTokens: 1000, CumulativeTokens: 2000},
		{Turn: 3, DeltaTokens: 3000, CumulativeTokens: 5000, ToolCalls: sharedCalls},
		{Turn: 4, DeltaTokens: 3000, CumulativeTokens: 8000, ToolCalls: sharedCalls},
	}
	sig := DetectRewindCandidate(turns, 200_000)
	if !sig.ShouldRewind {
		t.Fatal("expected rewind for retry of previous turn")
	}
	found := false
	for _, r := range sig.Reasons {
		if contains(r, "Retry detected") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected retry reason, got %v", sig.Reasons)
	}
}

func TestDetectRewindCandidate_HighCostLowBudget(t *testing.T) {
	maxTokens := 200_000
	turns := []RewindTurn{
		{Turn: 1, DeltaTokens: 50000, CumulativeTokens: 50000},
		{Turn: 2, DeltaTokens: 50000, CumulativeTokens: 100000},
		{Turn: 3, DeltaTokens: 50000, CumulativeTokens: 150000},
		{Turn: 4, DeltaTokens: 30000, CumulativeTokens: 180000},
	}
	sig := DetectRewindCandidate(turns, maxTokens)
	if !sig.ShouldRewind {
		t.Fatal("expected rewind for high cost low budget")
	}
	found := false
	for _, r := range sig.Reasons {
		if contains(r, "High cost vs budget") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected high cost reason, got %v", sig.Reasons)
	}
}

func TestDetectRewindCandidate_HealthySession(t *testing.T) {
	turns := []RewindTurn{
		{Turn: 1, DeltaTokens: 1000, CumulativeTokens: 1000},
		{Turn: 2, DeltaTokens: 1200, CumulativeTokens: 2200},
		{Turn: 3, DeltaTokens: 900, CumulativeTokens: 3100},
		{
			Turn: 4, DeltaTokens: 1100, CumulativeTokens: 4200,
			HasSuccessfulEdit: true,
			ToolCalls: []RewindToolCall{
				{ToolName: "Read", Category: "Read", EventStatus: ""},
				{ToolName: "Edit", Category: "Edit", EventStatus: ""},
			},
		},
	}
	sig := DetectRewindCandidate(turns, 200_000)
	if sig.ShouldRewind {
		t.Fatalf("expected no rewind for healthy session, got score=%d reasons=%v",
			sig.Score, sig.Reasons)
	}
}

func TestDetectRewindCandidate_MultiTurnBadStretch(t *testing.T) {
	failedCalls := []RewindToolCall{
		{ToolName: "Bash", Category: "Bash", EventStatus: "errored"},
		{ToolName: "Bash", Category: "Bash", EventStatus: "errored"},
		{ToolName: "Bash", Category: "Bash", EventStatus: "errored"},
	}
	turns := []RewindTurn{
		{Turn: 1, DeltaTokens: 1000, CumulativeTokens: 1000},
		{Turn: 2, DeltaTokens: 1000, CumulativeTokens: 2000,
			HasSuccessfulEdit: true,
			ToolCalls: []RewindToolCall{
				{ToolName: "Edit", Category: "Edit"},
			},
		},
		{Turn: 3, DeltaTokens: 3000, CumulativeTokens: 5000, ToolCalls: failedCalls},
		{Turn: 4, DeltaTokens: 3000, CumulativeTokens: 8000, ToolCalls: failedCalls},
		{Turn: 5, DeltaTokens: 3000, CumulativeTokens: 11000, ToolCalls: failedCalls},
	}
	sig := DetectRewindCandidate(turns, 200_000)
	if !sig.ShouldRewind {
		t.Fatal("expected rewind")
	}
	if sig.RewindToTurn != 2 {
		t.Fatalf("expected rewind to turn 2 (last clean), got %d", sig.RewindToTurn)
	}
	if sig.BadStretchFrom != 3 {
		t.Fatalf("expected bad stretch from turn 3, got %d", sig.BadStretchFrom)
	}
	if sig.BadStretchTo != 5 {
		t.Fatalf("expected bad stretch to turn 5, got %d", sig.BadStretchTo)
	}
	if sig.TokensRecoverable != 9000 {
		t.Fatalf("expected 9000 tokens recoverable, got %d", sig.TokensRecoverable)
	}
}

func TestDetectRewindCandidate_RetryChainRewindsToBeforeFirst(t *testing.T) {
	sharedCalls := []RewindToolCall{
		{ToolName: "Bash", Category: "Bash", InputJSON: `{"command":"npm test"}`},
		{ToolName: "Edit", Category: "Edit", InputJSON: `{"file_path":"/src/app.ts","old_string":"x"}`},
		{ToolName: "Bash", Category: "Bash", InputJSON: `{"command":"npm test"}`},
	}
	turns := []RewindTurn{
		{Turn: 1, DeltaTokens: 1000, CumulativeTokens: 1000},
		{Turn: 2, DeltaTokens: 1000, CumulativeTokens: 2000},
		{Turn: 3, DeltaTokens: 2000, CumulativeTokens: 4000, ToolCalls: sharedCalls},
		{Turn: 4, DeltaTokens: 2000, CumulativeTokens: 6000, ToolCalls: sharedCalls},
		{Turn: 5, DeltaTokens: 2000, CumulativeTokens: 8000, ToolCalls: sharedCalls},
	}
	sig := DetectRewindCandidate(turns, 200_000)
	if !sig.ShouldRewind {
		t.Fatal("expected rewind for retry chain")
	}
	// Turn 4 retries 3, turn 5 retries 4 → bad stretch is 4-5,
	// turn 3 is the first occurrence (not a retry of 2), so it's clean.
	if sig.RewindToTurn != 3 {
		t.Fatalf("expected rewind to turn 3, got %d", sig.RewindToTurn)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
