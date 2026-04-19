package summarize

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"

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
		fallback := fallbackSummary(bundle)
		fallback.Model = fallbackModel(resp.Model, model)
		return fallback, nil
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

const maxUserPromptLen = 1000
const maxAssistantPromptLen = 1400
const maxToolCallCount = 8
const maxToolPromptInputLen = 80
const maxToolPromptResultLen = 80
const maxFallbackSummaryLen = 180

func renderTurnPrompt(b TurnBundle) string {
	var sb strings.Builder
	fmt.Fprintf(&sb,
		"Turn %d (messages %d-%d).\n\n",
		b.TurnIndex, b.StartOrdinal, b.EndOrdinal,
	)
	if b.UserMessage != "" {
		sb.WriteString("USER:\n")
		sb.WriteString(truncate(b.UserMessage, maxUserPromptLen))
		sb.WriteString("\n\n")
	}
	if b.Thinking != "" {
		sb.WriteString("THINKING (truncated):\n")
		sb.WriteString(truncate(b.Thinking, 600))
		sb.WriteString("\n\n")
	}
	if b.AssistantMessage != "" {
		sb.WriteString("ASSISTANT:\n")
		sb.WriteString(truncate(b.AssistantMessage, maxAssistantPromptLen))
		sb.WriteString("\n\n")
	}
	if len(b.ToolCalls) > 0 {
		sb.WriteString("TOOL CALLS:\n")
		limit := min(len(b.ToolCalls), maxToolCallCount)
		for i, tc := range b.ToolCalls[:limit] {
			fmt.Fprintf(&sb,
				"%d) %s (%s) status=%s\n",
				i+1, tc.ToolName, tc.Category, ifEmpty(tc.Status, "-"),
			)
			if tc.InputText != "" {
				fmt.Fprintf(&sb, "   input: %s\n",
					truncate(tc.InputText, maxToolPromptInputLen))
			}
			if tc.Result != "" {
				fmt.Fprintf(&sb, "   result: %s\n",
					truncate(tc.Result, maxToolPromptResultLen))
			}
		}
		if len(b.ToolCalls) > limit {
			fmt.Fprintf(&sb,
				"... %d more tool calls omitted for brevity\n",
				len(b.ToolCalls)-limit,
			)
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

func fallbackModel(respModel, requestedModel string) string {
	base := strings.TrimSpace(respModel)
	if base == "" {
		base = strings.TrimSpace(requestedModel)
	}
	if base == "" {
		return "fallback-local"
	}
	return base + "+fallback-local"
}

func fallbackSummary(b TurnBundle) Summary {
	topic := inferTopic(b)
	intent := inferIntent(b)
	outcome := inferOutcome(b)
	files := inferFilesTouched(b)
	tags := inferTags(b)

	request := summariseSnippet(b.UserMessage)
	action := summariseSnippet(b.AssistantMessage)
	if request == "" {
		request = "continued the current task"
	}
	if action == "" {
		action = "made progress on the task"
	}
	return Summary{
		Summary: truncate(
			fmt.Sprintf(
				"User requested %s. Assistant %s.",
				request, lowerFirst(action),
			),
			maxFallbackSummaryLen,
		),
		Intent:       intent,
		Outcome:      outcome,
		Topic:        topic,
		FilesTouched: files,
		Tags:         tags,
	}
}

func summariseSnippet(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "\n", " ")
	if idx := strings.IndexAny(s, ".!?"); idx >= 0 {
		s = s[:idx]
	}
	return strings.TrimSpace(truncate(s, 90))
}

func inferIntent(b TurnBundle) string {
	user := strings.ToLower(b.UserMessage)
	switch {
	case strings.Contains(user, "debug") || strings.Contains(user, "fix"):
		return "debug"
	case strings.Contains(user, "plan"):
		return "plan"
	case strings.Contains(user, "test"):
		return "test"
	case strings.Contains(user, "refactor"):
		return "refactor"
	case strings.Contains(user, "implement") || strings.Contains(user, "add") ||
		strings.Contains(user, "build"):
		return "implement"
	case strings.Contains(user, "review"):
		return "answer"
	}
	for _, tc := range b.ToolCalls {
		switch tc.Category {
		case "Read", "Glob", "Grep":
			return "read"
		case "Edit", "Write":
			return "implement"
		case "Bash":
			if strings.Contains(strings.ToLower(tc.InputText), "test") {
				return "test"
			}
		}
	}
	return "other"
}

func inferOutcome(b TurnBundle) string {
	assistant := strings.ToLower(b.AssistantMessage)
	if strings.Contains(assistant, "no progress") ||
		strings.Contains(assistant, "stuck") {
		return "stuck"
	}
	for _, tc := range b.ToolCalls {
		status := strings.ToLower(tc.Status)
		result := strings.ToLower(tc.Result)
		if status == "errored" || status == "cancelled" ||
			strings.Contains(result, "error") ||
			strings.Contains(result, "failed") {
			if strings.Contains(assistant, "fixed") ||
				strings.Contains(assistant, "resolved") {
				return "mixed"
			}
			return "failed"
		}
	}
	if assistant == "" && len(b.ToolCalls) > 0 {
		return "info"
	}
	return "progress"
}

func inferTopic(b TurnBundle) string {
	words := topicWords(b.UserMessage)
	if len(words) == 0 {
		words = topicWords(b.AssistantMessage)
	}
	if len(words) == 0 {
		return "coding task"
	}
	if len(words) > 6 {
		words = words[:6]
	}
	return strings.Join(words, " ")
}

func topicWords(s string) []string {
	fields := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) ||
			r == '/' || r == '_' || r == '-')
	})
	stop := map[string]struct{}{
		"the": {}, "and": {}, "that": {}, "this": {}, "with": {},
		"from": {}, "into": {}, "then": {}, "have": {}, "your": {},
		"make": {}, "look": {}, "through": {}, "these": {}, "them": {},
		"also": {}, "want": {}, "need": {}, "able": {}, "show": {},
		"about": {}, "just": {}, "very": {}, "more": {}, "will": {},
		"would": {}, "could": {}, "should": {}, "what": {}, "when": {},
		"where": {}, "which": {}, "there": {}, "their": {}, "working": {},
		"implemented": {}, "features": {}, "commits": {}, "last": {},
		"user": {}, "assistant": {}, "please": {}, "review": {},
		"write": {}, "detailed": {}, "proper": {}, "handoff": {},
	}
	out := make([]string, 0, 6)
	for _, field := range fields {
		if len(field) < 3 {
			continue
		}
		if _, ok := stop[field]; ok {
			continue
		}
		out = append(out, field)
		if len(out) >= 6 {
			break
		}
	}
	return out
}

var filePathRE = regexp.MustCompile(`(?:/|[A-Za-z]:\\)[A-Za-z0-9._/\-\\]+|[A-Za-z0-9._-]+\.(?:go|ts|tsx|js|jsx|json|md|sql|svelte|toml|yaml|yml|py|rb|rs|java|kt|swift|sh)`)

func inferFilesTouched(b TurnBundle) []string {
	seen := map[string]struct{}{}
	var out []string
	addMatches := func(s string) {
		for _, match := range filePathRE.FindAllString(s, -1) {
			match = strings.Trim(match, `"'.,:;()[]{}<>`)
			if match == "" {
				continue
			}
			if _, ok := seen[match]; ok {
				continue
			}
			seen[match] = struct{}{}
			out = append(out, match)
			if len(out) >= 8 {
				return
			}
		}
	}
	addMatches(b.UserMessage)
	addMatches(b.AssistantMessage)
	for _, tc := range b.ToolCalls {
		if len(out) >= 8 {
			break
		}
		addMatches(tc.InputText)
		addMatches(tc.Result)
	}
	return out
}

func inferTags(b TurnBundle) []string {
	var tags []string
	readCount := 0
	errorCount := 0
	for _, tc := range b.ToolCalls {
		if tc.Category == "Read" {
			readCount++
		}
		status := strings.ToLower(tc.Status)
		result := strings.ToLower(tc.Result)
		if status == "errored" || strings.Contains(result, "failed") ||
			strings.Contains(result, "error") {
			errorCount++
		}
	}
	if readCount >= 8 {
		tags = append(tags, "large-read")
	}
	if errorCount >= 2 {
		tags = append(tags, "retry-loop")
	}
	slices.Sort(tags)
	return tags
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
