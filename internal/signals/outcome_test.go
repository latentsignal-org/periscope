package signals

import (
	"testing"
	"time"
)

func TestClassifyOutcome(t *testing.T) {
	pastTime := time.Now().Add(-1 * time.Hour)
	recentTime := time.Now().Add(-1 * time.Minute)

	tests := []struct {
		name       string
		input      OutcomeInput
		wantResult OutcomeResult
	}{
		{
			name: "automated session",
			input: OutcomeInput{
				IsAutomated:   true,
				MessageCount:  20,
				EndedWithRole: "assistant",
			},
			wantResult: OutcomeResult{
				Outcome:    "unknown",
				Confidence: "low",
			},
		},
		{
			name: "one-shot Q&A",
			input: OutcomeInput{
				MessageCount:  2,
				EndedWithRole: "assistant",
				LastActivity:  pastTime,
			},
			wantResult: OutcomeResult{
				Outcome:    "completed",
				Confidence: "medium",
			},
		},
		{
			name: "too few messages",
			input: OutcomeInput{
				MessageCount:  1,
				EndedWithRole: "user",
				LastActivity:  pastTime,
			},
			wantResult: OutcomeResult{
				Outcome:    "unknown",
				Confidence: "low",
			},
		},
		{
			name: "recent session",
			input: OutcomeInput{
				MessageCount:  5,
				EndedWithRole: "assistant",
				LastActivity:  recentTime,
			},
			wantResult: OutcomeResult{
				Outcome:    "unknown",
				Confidence: "low",
				IsRecent:   true,
			},
		},
		{
			name: "ended with user few msgs",
			input: OutcomeInput{
				MessageCount:  5,
				EndedWithRole: "user",
				LastActivity:  pastTime,
			},
			wantResult: OutcomeResult{
				Outcome:    "abandoned",
				Confidence: "medium",
			},
		},
		{
			name: "ended with user many msgs",
			input: OutcomeInput{
				MessageCount:  10,
				EndedWithRole: "user",
				LastActivity:  pastTime,
			},
			wantResult: OutcomeResult{
				Outcome:    "abandoned",
				Confidence: "high",
			},
		},
		{
			name: "final failure streak",
			input: OutcomeInput{
				MessageCount:       6,
				EndedWithRole:      "assistant",
				FinalFailureStreak: 3,
				LastActivity:       pastTime,
			},
			wantResult: OutcomeResult{
				Outcome:    "errored",
				Confidence: "medium",
			},
		},
		{
			name: "ended with assistant normal",
			input: OutcomeInput{
				MessageCount:      4,
				EndedWithRole:     "assistant",
				LastAssistantText: "Done! All tests pass.",
				LastActivity:      pastTime,
			},
			wantResult: OutcomeResult{
				Outcome:    "completed",
				Confidence: "medium",
			},
		},
		{
			name: "give-up pattern lowers confidence",
			input: OutcomeInput{
				MessageCount:      4,
				EndedWithRole:     "assistant",
				LastAssistantText: "I'm unable to complete this task.",
				LastActivity:      pastTime,
			},
			wantResult: OutcomeResult{
				Outcome:    "completed",
				Confidence: "low",
			},
		},
		{
			name: "zero-value LastActivity not recent",
			input: OutcomeInput{
				MessageCount:  5,
				EndedWithRole: "assistant",
			},
			wantResult: OutcomeResult{
				Outcome:    "completed",
				Confidence: "medium",
			},
		},
		{
			name: "give-up I cannot proceed",
			input: OutcomeInput{
				MessageCount:      4,
				EndedWithRole:     "assistant",
				LastAssistantText: "I cannot proceed without credentials.",
				LastActivity:      pastTime,
			},
			wantResult: OutcomeResult{
				Outcome:    "completed",
				Confidence: "low",
			},
		},
		{
			name: "give-up I am unable to",
			input: OutcomeInput{
				MessageCount:      4,
				EndedWithRole:     "assistant",
				LastAssistantText: "I am unable to access that.",
				LastActivity:      pastTime,
			},
			wantResult: OutcomeResult{
				Outcome:    "completed",
				Confidence: "low",
			},
		},
		{
			name: "give-up I don't have access",
			input: OutcomeInput{
				MessageCount:      4,
				EndedWithRole:     "assistant",
				LastAssistantText: "I don't have access to the repo.",
				LastActivity:      pastTime,
			},
			wantResult: OutcomeResult{
				Outcome:    "completed",
				Confidence: "low",
			},
		},
		{
			name: "give-up I can't proceed",
			input: OutcomeInput{
				MessageCount:      4,
				EndedWithRole:     "assistant",
				LastAssistantText: "I can't proceed further.",
				LastActivity:      pastTime,
			},
			wantResult: OutcomeResult{
				Outcome:    "completed",
				Confidence: "low",
			},
		},
		{
			name: "no role set falls through",
			input: OutcomeInput{
				MessageCount: 5,
				LastActivity: pastTime,
			},
			wantResult: OutcomeResult{
				Outcome:    "unknown",
				Confidence: "low",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyOutcome(tt.input)
			if got != tt.wantResult {
				t.Errorf(
					"ClassifyOutcome() = %+v, want %+v",
					got, tt.wantResult,
				)
			}
		})
	}
}
