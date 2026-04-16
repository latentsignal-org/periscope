package sync

import (
	"reflect"
	"testing"

	"github.com/wesm/agentsview/internal/db"
	"github.com/wesm/agentsview/internal/signals"
)

func TestExtractToolCallRows(t *testing.T) {
	msgs := []db.Message{
		{Ordinal: 0, Role: "user", Content: "do stuff"},
		{
			Ordinal: 1,
			Role:    "assistant",
			ToolCalls: []db.ToolCall{
				{
					ToolName:      "Bash",
					Category:      "Bash",
					InputJSON:     `{"command":"ls"}`,
					ResultContent: "/tmp",
					ResultEvents: []db.ToolResultEvent{
						{Status: "completed", EventIndex: 0},
					},
				},
				{
					ToolName:      "Edit",
					Category:      "Edit",
					InputJSON:     `{"file":"/a.go"}`,
					ResultContent: "ok",
					// Multiple events: latest wins.
					ResultEvents: []db.ToolResultEvent{
						{Status: "running", EventIndex: 0},
						{Status: "errored", EventIndex: 1},
					},
				},
			},
		},
		{
			Ordinal: 2,
			Role:    "assistant",
			ToolCalls: []db.ToolCall{
				{
					ToolName:  "Read",
					Category:  "Read",
					InputJSON: `{"file":"/b.go"}`,
				},
			},
		},
	}

	got := extractToolCallRows(msgs)
	want := []signals.ToolCallRow{
		{
			ToolName:       "Bash",
			Category:       "Bash",
			InputJSON:      `{"command":"ls"}`,
			ResultContent:  "/tmp",
			MessageOrdinal: 1,
			CallIndex:      0,
			EventStatus:    "completed",
		},
		{
			ToolName:       "Edit",
			Category:       "Edit",
			InputJSON:      `{"file":"/a.go"}`,
			ResultContent:  "ok",
			MessageOrdinal: 1,
			CallIndex:      1,
			EventStatus:    "errored",
		},
		{
			ToolName:       "Read",
			Category:       "Read",
			InputJSON:      `{"file":"/b.go"}`,
			MessageOrdinal: 2,
			CallIndex:      0,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractToolCallRows mismatch\n got = %+v\nwant = %+v", got, want)
	}
}

func TestExtractContextTokens(t *testing.T) {
	msgs := []db.Message{
		{Ordinal: 0, Role: "user"},
		{
			Ordinal: 1, Role: "assistant",
			ContextTokens: 1000, HasContextTokens: true,
		},
		{Ordinal: 2, Role: "user"},
		{
			Ordinal: 3, Role: "assistant",
			ContextTokens: 2000, HasContextTokens: true,
		},
		// Zero/missing tokens are still emitted (caller cares).
		{Ordinal: 4, Role: "assistant"},
	}
	got := extractContextTokens(msgs)
	want := []signals.ContextTokenRow{
		{ContextTokens: 1000, HasContextTokens: true},
		{ContextTokens: 2000, HasContextTokens: true},
		{ContextTokens: 0, HasContextTokens: false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractContextTokens mismatch\n got = %+v\nwant = %+v", got, want)
	}
}

func TestExtractCompactBoundaryOrdinals(t *testing.T) {
	msgs := []db.Message{
		{Ordinal: 0, Role: "user"},
		{Ordinal: 1, Role: "assistant"},
		{Ordinal: 2, Role: "user", IsCompactBoundary: true},
		{Ordinal: 3, Role: "assistant"},
		{Ordinal: 4, Role: "user", IsCompactBoundary: true},
	}
	got := extractCompactBoundaryOrdinals(msgs)
	want := []int{2, 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractCompactBoundaryOrdinals = %v, want %v", got, want)
	}

	if extractCompactBoundaryOrdinals(nil) != nil {
		t.Error("extractCompactBoundaryOrdinals(nil) should return nil")
	}
}

func TestExtractMostCommonModel(t *testing.T) {
	msgs := []db.Message{
		{Role: "user"},
		{Role: "assistant", Model: "claude-sonnet-4-5"},
		{Role: "assistant", Model: "claude-sonnet-4-5"},
		{Role: "assistant", Model: "claude-opus-4-6"},
		{Role: "assistant", Model: ""}, // ignored
	}
	if got := extractMostCommonModel(msgs); got != "claude-sonnet-4-5" {
		t.Errorf("extractMostCommonModel = %q, want claude-sonnet-4-5", got)
	}

	// Tie broken by chronological-first.
	tied := []db.Message{
		{Role: "assistant", Model: "claude-sonnet-4-5"},
		{Role: "assistant", Model: "claude-opus-4-6"},
	}
	if got := extractMostCommonModel(tied); got != "claude-sonnet-4-5" {
		t.Errorf("tied: extractMostCommonModel = %q, want claude-sonnet-4-5", got)
	}

	if got := extractMostCommonModel(nil); got != "" {
		t.Errorf("empty: extractMostCommonModel = %q, want empty", got)
	}
}

func TestExtractLastMessageRole(t *testing.T) {
	msgs := []db.Message{
		{Ordinal: 0, Role: "user", Content: "hi"},
		{Ordinal: 1, Role: "assistant", Content: "hello"},
		{Ordinal: 2, Role: "user", Content: "thanks"},
		{Ordinal: 3, Role: "user", Content: "system noise", IsSystem: true},
	}
	role, content := extractLastMessageRole(msgs)
	if role != "user" || content != "thanks" {
		t.Errorf("extractLastMessageRole = (%q, %q), want (user, thanks)", role, content)
	}

	role, content = extractLastMessageRole(nil)
	if role != "" || content != "" {
		t.Errorf("nil case: got (%q, %q), want empty", role, content)
	}
}

func TestComputeSignalsFromMessages_Errors(t *testing.T) {
	// Session with a final tool failure: outcome should be
	// "errored" (recent enough to be pending), penalties should
	// reflect the failure streak, and HasToolCalls is true.
	endedAt := "2099-12-31T00:00:00Z"
	sess := db.Session{
		ID:                "s1",
		MessageCount:      4,
		EndedAt:           &endedAt,
		PeakContextTokens: 50_000,
	}
	msgs := []db.Message{
		{Ordinal: 0, Role: "user", Content: "go"},
		{
			Ordinal: 1, Role: "assistant", Model: "claude-sonnet-4-5",
			ContextTokens: 10_000, HasContextTokens: true,
			ToolCalls: []db.ToolCall{{
				ToolName: "Bash", Category: "Bash",
				ResultEvents: []db.ToolResultEvent{
					{Status: "errored", EventIndex: 0},
				},
			}},
		},
		{
			Ordinal: 2, Role: "assistant", Model: "claude-sonnet-4-5",
			ContextTokens: 12_000, HasContextTokens: true,
			ToolCalls: []db.ToolCall{{
				ToolName: "Bash", Category: "Bash",
				ResultEvents: []db.ToolResultEvent{
					{Status: "errored", EventIndex: 0},
				},
			}},
		},
		{Ordinal: 3, Role: "assistant", Content: "I give up"},
	}

	got := computeSignalsFromMessages(sess, msgs)

	if !got.HasToolCalls {
		t.Error("HasToolCalls = false, want true")
	}
	if !got.HasContextData {
		t.Error("HasContextData = false, want true")
	}
	if got.ToolFailureSignalCount == 0 {
		t.Error("ToolFailureSignalCount = 0, want > 0")
	}
	if got.FinalFailureStreak == 0 {
		t.Errorf("FinalFailureStreak = 0, want > 0")
	}
	if got.HealthScore == nil {
		t.Fatal("HealthScore is nil; want a value")
	}
	if *got.HealthScore >= 100 {
		t.Errorf("HealthScore = %d, want < 100", *got.HealthScore)
	}
	if got.HealthGrade == nil || *got.HealthGrade == "" {
		t.Errorf("HealthGrade = %v, want non-empty", got.HealthGrade)
	}
	if got.EndedWithRole != "assistant" {
		t.Errorf("EndedWithRole = %q, want assistant", got.EndedWithRole)
	}
}

func TestComputeSignalsFromMessages_ExplicitBoundariesOverrideHeuristic(t *testing.T) {
	// Two explicit boundaries should win over zero token-drops.
	sess := db.Session{ID: "s1", MessageCount: 5}
	msgs := []db.Message{
		{Ordinal: 0, Role: "user"},
		{Ordinal: 1, Role: "assistant", Model: "claude-sonnet-4-5"},
		{Ordinal: 2, Role: "user", IsCompactBoundary: true},
		{Ordinal: 3, Role: "assistant", Model: "claude-sonnet-4-5"},
		{Ordinal: 4, Role: "user", IsCompactBoundary: true},
	}
	got := computeSignalsFromMessages(sess, msgs)
	if got.CompactionCount != 2 {
		t.Errorf("CompactionCount = %d, want 2", got.CompactionCount)
	}
}
