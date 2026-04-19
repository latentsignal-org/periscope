package summarize

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wesm/agentsview/internal/llm"
)

// systemPrompt is cached on the API side, so it can be long without
// incurring the full cost on every call.
const systemPrompt = `You analyze a single turn of a coding-agent session and produce a compact JSON summary.

A "turn" is one user message plus the assistant's response, including any tool calls and their outputs.

Return ONLY a single JSON object with these keys:
- summary: one or two sentences describing what the user asked and what the assistant did (past tense, concrete).
- intent: one of "implement", "debug", "explore", "read", "test", "refactor", "plan", "answer", "other".
- outcome: one of "progress", "stuck", "failed", "off-topic", "info", "mixed".
- topic: 2-6 words naming the subject (e.g. "auth middleware refactor", "flaky login test").
- files_touched: array of file paths the turn read or modified, deduplicated.
- tags: 0-4 short lowercase tags (e.g. ["retry-loop", "large-read", "tangent", "new-topic"]).

Rules:
- Do not invent file paths. Only include files that appear in the input.
- Prefer nouns over verbs in topic.
- If the turn is clearly a tangent from the session's main task (e.g. pivoting to deployment warnings during an auth refactor), add the "tangent" tag.
- If the assistant retried the same failing approach, add "retry-loop".
- If a single tool produced a large low-value read, add "large-read".
- JSON only, no prose, no code fences.`

// SummarizeTurn calls the LLM with a single turn bundle and returns
// a parsed summary. The caller decides whether to persist it.
func SummarizeTurn(
	ctx context.Context,
	client llm.Client,
	model string,
	bundle TurnBundle,
) (Summary, error) {
	if client == nil {
		return Summary{}, fmt.Errorf("summarize: nil client")
	}
	userPrompt := renderTurnPrompt(bundle)
	if model == "" {
		model = llm.DefaultSummaryModel
	}
	resp, err := client.Complete(ctx, llm.Request{
		Model:        model,
		MaxTokens:    400,
		Temperature:  0,
		SystemCached: systemPrompt,
		Messages: []llm.Message{
			{Role: "user", Content: userPrompt},
		},
	})
	if err != nil {
		return Summary{}, err
	}
	parsed, err := parseSummary(resp.Text)
	if err != nil {
		return Summary{}, fmt.Errorf(
			"summarize: parse: %w text=%s",
			err, truncate(resp.Text, 300),
		)
	}
	parsed.Model = resp.Model
	if parsed.Model == "" {
		parsed.Model = model
	}
	return parsed, nil
}

// Summary is the parsed JSON returned by the LLM.
type Summary struct {
	Summary      string   `json:"summary"`
	Intent       string   `json:"intent"`
	Outcome      string   `json:"outcome"`
	Topic        string   `json:"topic"`
	FilesTouched []string `json:"files_touched"`
	Tags         []string `json:"tags"`
	Model        string   `json:"-"`
}

func renderTurnPrompt(b TurnBundle) string {
	var sb strings.Builder
	fmt.Fprintf(&sb,
		"Turn %d (messages %d-%d).\n\n",
		b.TurnIndex, b.StartOrdinal, b.EndOrdinal,
	)
	if b.UserMessage != "" {
		sb.WriteString("USER:\n")
		sb.WriteString(b.UserMessage)
		sb.WriteString("\n\n")
	}
	if b.Thinking != "" {
		sb.WriteString("THINKING (truncated):\n")
		sb.WriteString(truncate(b.Thinking, 600))
		sb.WriteString("\n\n")
	}
	if b.AssistantMessage != "" {
		sb.WriteString("ASSISTANT:\n")
		sb.WriteString(b.AssistantMessage)
		sb.WriteString("\n\n")
	}
	if len(b.ToolCalls) > 0 {
		sb.WriteString("TOOL CALLS:\n")
		for i, tc := range b.ToolCalls {
			fmt.Fprintf(&sb,
				"%d) %s (%s) status=%s\n",
				i+1, tc.ToolName, tc.Category, ifEmpty(tc.Status, "-"),
			)
			if tc.InputText != "" {
				fmt.Fprintf(&sb, "   input: %s\n", tc.InputText)
			}
			if tc.Result != "" {
				fmt.Fprintf(&sb, "   result: %s\n", tc.Result)
			}
		}
		sb.WriteString("\n")
	}
	sb.WriteString("Return the JSON object now.")
	return sb.String()
}

func parseSummary(text string) (Summary, error) {
	text = strings.TrimSpace(text)
	text = stripCodeFences(text)
	// The model may still wrap the JSON in prose. Extract the first
	// balanced object.
	if i := strings.IndexByte(text, '{'); i > 0 {
		text = text[i:]
	}
	if j := strings.LastIndexByte(text, '}'); j >= 0 && j+1 < len(text) {
		text = text[:j+1]
	}
	var s Summary
	if err := json.Unmarshal([]byte(text), &s); err != nil {
		return Summary{}, err
	}
	if s.FilesTouched == nil {
		s.FilesTouched = []string{}
	}
	if s.Tags == nil {
		s.Tags = []string{}
	}
	return s, nil
}

func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func ifEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
