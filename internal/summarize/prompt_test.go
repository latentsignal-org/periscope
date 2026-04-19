package summarize

import (
	"context"
	"strings"
	"testing"

	"github.com/wesm/agentsview/internal/llm"
)

type captureClient struct {
	req  llm.Request
	resp llm.Response
}

func (c *captureClient) Complete(
	_ context.Context, req llm.Request,
) (llm.Response, error) {
	c.req = req
	if c.resp.Text == "" {
		c.resp.Text = `{"summary":"ok","intent":"implement","outcome":"progress","topic":"test","files_touched":[],"tags":[]}`
	}
	if c.resp.Model == "" {
		c.resp.Model = "claude-haiku-4-5"
	}
	return c.resp, nil
}

func TestSummarizeTurn_FallsBackWhenModelReturnsMalformedJSON(t *testing.T) {
	client := &captureClient{
		resp: llm.Response{
			Text:  "```json",
			Model: "claude-haiku-4-5-20251001",
		},
	}
	bundle := TurnBundle{
		TurnIndex:        1,
		StartOrdinal:     1,
		EndOrdinal:       3,
		UserMessage:      "Debug the failing auth middleware tests in internal/server/context.go.",
		AssistantMessage: "Inspected the test failure and traced it back to internal/server/context.go.",
		ToolCalls: []ToolCallBundle{
			{
				ToolName:  "Read",
				Category:  "Read",
				InputText: `{"file_path":"internal/server/context.go"}`,
			},
			{
				ToolName:  "Bash",
				Category:  "Bash",
				InputText: `{"command":"go test ./internal/server -run TestComputeSessionContextView"}`,
				Result:    "FAIL",
				Status:    "errored",
			},
		},
	}

	got, err := SummarizeTurn(
		context.Background(), client, llm.DefaultSummaryModel, bundle,
	)
	if err != nil {
		t.Fatalf("SummarizeTurn returned error: %v", err)
	}
	if got.Summary == "" {
		t.Fatal("fallback summary is empty")
	}
	if got.Intent != "debug" {
		t.Fatalf("intent = %q, want debug", got.Intent)
	}
	if got.Outcome != "failed" {
		t.Fatalf("outcome = %q, want failed", got.Outcome)
	}
	if got.Model != "claude-haiku-4-5-20251001+fallback-local" {
		t.Fatalf("model = %q, want fallback model suffix", got.Model)
	}
	foundContextFile := false
	for _, path := range got.FilesTouched {
		if strings.Contains(path, "context.go") {
			foundContextFile = true
			break
		}
	}
	if !foundContextFile {
		t.Fatalf("files_touched = %#v, want a path containing context.go", got.FilesTouched)
	}
}

func TestRenderTurnPrompt_TruncatesLargeBundles(t *testing.T) {
	client := &captureClient{}
	largeAssistant := strings.Repeat("assistant output ", 400)
	toolCalls := make([]ToolCallBundle, 0, 40)
	for i := 0; i < 40; i++ {
		toolCalls = append(toolCalls, ToolCallBundle{
			ToolName:  "Read",
			Category:  "Read",
			InputText: strings.Repeat("x", 100),
			Result:    strings.Repeat("y", 100),
		})
	}
	_, err := SummarizeTurn(context.Background(), client, "", TurnBundle{
		TurnIndex:        1,
		StartOrdinal:     1,
		EndOrdinal:       50,
		UserMessage:      strings.Repeat("user request ", 200),
		AssistantMessage: largeAssistant,
		ToolCalls:        toolCalls,
	})
	if err != nil {
		t.Fatalf("SummarizeTurn returned error: %v", err)
	}
	if strings.Count(client.req.Messages[0].Content, "\n") == 0 {
		t.Fatal("captured prompt is unexpectedly empty")
	}
	if strings.Count(client.req.Messages[0].Content, ") Read (Read)") > maxToolCallCount {
		t.Fatalf("tool calls in prompt exceeded cap: %d", strings.Count(client.req.Messages[0].Content, ") Read (Read)"))
	}
	if !strings.Contains(client.req.Messages[0].Content, "more tool calls omitted for brevity") {
		t.Fatal("prompt missing omission marker for capped tool calls")
	}
	if len(client.req.Messages[0].Content) > 5500 {
		t.Fatalf("prompt length = %d, want bounded prompt", len(client.req.Messages[0].Content))
	}
}
