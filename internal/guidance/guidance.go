package guidance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wesm/agentsview/internal/db"
	"github.com/wesm/agentsview/internal/llm"
)

const (
	rewindSystemPrompt = `You analyze a rewind candidate in a coding-agent session and produce concise operational guidance.

Return ONLY a single JSON object with these keys:
- tangent_label: 2-8 words naming the tangent, failing branch, or low-value branch to drop.
- rewind_reprompt_text: a direct prompt the user can paste after rewinding. It should tell the agent what to ignore/drop, what useful state to keep in mind, and what task to resume.

Rules:
- Base the answer only on the provided task topic and turn summaries.
- Do not invent files, commands, or facts not present in the input.
- Keep tangent_label short and concrete.
- Keep rewind_reprompt_text under 90 words.
- JSON only, no prose, no code fences.`

	compactSystemPrompt = `You analyze a compaction candidate in a coding-agent session and produce concise operational guidance.

Return ONLY a single JSON object with these keys:
- keep_items: 1-4 short phrases the compact summary must preserve.
- drop_items: 1-4 short phrases the compact summary can discard.
- compact_focus_text: a direct prompt the user can paste when compacting. It should say what to preserve, what to drop, and how to continue.

Rules:
- Base the answer only on the provided summaries and low-value categories.
- Do not invent files, commands, or facts not present in the input.
- keep_items and drop_items should be short noun phrases, not full sentences.
- compact_focus_text must be under 90 words.
- JSON only, no prose, no code fences.`
)

type RewindInput struct {
	TaskTopic     string
	LastCleanTurn db.TurnSummary
	BadStretch    []db.TurnSummary
}

type RewindOutput struct {
	TangentLabel       string `json:"tangent_label"`
	RewindRepromptText string `json:"rewind_reprompt_text"`
	Model              string `json:"-"`
}

type CompactInput struct {
	TaskTopic          string
	OlderTurns         []db.TurnSummary
	RecentTurns        []db.TurnSummary
	LowValueCategories []string
}

type CompactOutput struct {
	KeepItems        []string `json:"keep_items"`
	DropItems        []string `json:"drop_items"`
	CompactFocusText string   `json:"compact_focus_text"`
	Model            string   `json:"-"`
}

func GenerateRewind(
	ctx context.Context,
	client llm.Client,
	model string,
	in RewindInput,
) (RewindOutput, error) {
	if client == nil {
		return RewindOutput{}, fmt.Errorf("guidance: nil client")
	}
	if model == "" {
		model = llm.DefaultGenerateModel
	}
	resp, err := client.Complete(ctx, llm.Request{
		Model:        model,
		MaxTokens:    350,
		Temperature:  0,
		SystemCached: rewindSystemPrompt,
		Messages: []llm.Message{{
			Role:    "user",
			Content: renderRewindPrompt(in),
		}},
	})
	if err != nil {
		return RewindOutput{}, err
	}
	out, err := parseRewindOutput(resp.Text)
	if err != nil {
		return RewindOutput{}, fmt.Errorf(
			"guidance: parse rewind: %w text=%s",
			err, truncate(resp.Text, 300),
		)
	}
	out.Model = resp.Model
	if out.Model == "" {
		out.Model = model
	}
	return out, nil
}

func GenerateCompact(
	ctx context.Context,
	client llm.Client,
	model string,
	in CompactInput,
) (CompactOutput, error) {
	if client == nil {
		return CompactOutput{}, fmt.Errorf("guidance: nil client")
	}
	if model == "" {
		model = llm.DefaultGenerateModel
	}
	resp, err := client.Complete(ctx, llm.Request{
		Model:        model,
		MaxTokens:    400,
		Temperature:  0,
		SystemCached: compactSystemPrompt,
		Messages: []llm.Message{{
			Role:    "user",
			Content: renderCompactPrompt(in),
		}},
	})
	if err != nil {
		return CompactOutput{}, err
	}
	out, err := parseCompactOutput(resp.Text)
	if err != nil {
		return CompactOutput{}, fmt.Errorf(
			"guidance: parse compact: %w text=%s",
			err, truncate(resp.Text, 300),
		)
	}
	out.Model = resp.Model
	if out.Model == "" {
		out.Model = model
	}
	return out, nil
}

func renderRewindPrompt(in RewindInput) string {
	var sb strings.Builder
	if in.TaskTopic != "" {
		fmt.Fprintf(&sb, "Overall task topic: %s\n\n", in.TaskTopic)
	}
	sb.WriteString("Last clean turn:\n")
	writeTurnSummary(&sb, in.LastCleanTurn)
	sb.WriteString("\nBad stretch:\n")
	for _, turn := range in.BadStretch {
		writeTurnSummary(&sb, turn)
	}
	sb.WriteString("\nReturn the JSON object now.")
	return sb.String()
}

func renderCompactPrompt(in CompactInput) string {
	var sb strings.Builder
	if in.TaskTopic != "" {
		fmt.Fprintf(&sb, "Overall task topic: %s\n\n", in.TaskTopic)
	}
	if len(in.LowValueCategories) > 0 {
		fmt.Fprintf(&sb, "Dominant low-value categories: %s\n\n", strings.Join(in.LowValueCategories, ", "))
	}
	sb.WriteString("Older turn summaries:\n")
	for _, turn := range in.OlderTurns {
		writeTurnSummary(&sb, turn)
	}
	sb.WriteString("\nRecent turn summaries:\n")
	for _, turn := range in.RecentTurns {
		writeTurnSummary(&sb, turn)
	}
	sb.WriteString("\nReturn the JSON object now.")
	return sb.String()
}

func writeTurnSummary(sb *strings.Builder, turn db.TurnSummary) {
	fmt.Fprintf(sb, "- Turn %d", turn.TurnIndex)
	if turn.Topic != "" {
		fmt.Fprintf(sb, " | topic=%s", turn.Topic)
	}
	if turn.Intent != "" {
		fmt.Fprintf(sb, " | intent=%s", turn.Intent)
	}
	if turn.Outcome != "" {
		fmt.Fprintf(sb, " | outcome=%s", turn.Outcome)
	}
	sb.WriteString("\n")
	if turn.Summary != "" {
		fmt.Fprintf(sb, "  summary: %s\n", turn.Summary)
	}
}

func parseRewindOutput(text string) (RewindOutput, error) {
	text = cleanJSONText(text)
	var out RewindOutput
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return RewindOutput{}, err
	}
	out.TangentLabel = strings.TrimSpace(out.TangentLabel)
	out.RewindRepromptText = strings.TrimSpace(out.RewindRepromptText)
	return out, nil
}

func parseCompactOutput(text string) (CompactOutput, error) {
	text = cleanJSONText(text)
	var out CompactOutput
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return CompactOutput{}, err
	}
	out.CompactFocusText = strings.TrimSpace(out.CompactFocusText)
	if out.KeepItems == nil {
		out.KeepItems = []string{}
	}
	if out.DropItems == nil {
		out.DropItems = []string{}
	}
	for i := range out.KeepItems {
		out.KeepItems[i] = strings.TrimSpace(out.KeepItems[i])
	}
	for i := range out.DropItems {
		out.DropItems[i] = strings.TrimSpace(out.DropItems[i])
	}
	return out, nil
}

func cleanJSONText(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	if i := strings.IndexByte(s, '{'); i > 0 {
		s = s[i:]
	}
	if j := strings.LastIndexByte(s, '}'); j >= 0 && j+1 < len(s) {
		s = s[:j+1]
	}
	return s
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
