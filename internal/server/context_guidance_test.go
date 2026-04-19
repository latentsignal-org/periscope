package server

import (
	"context"
	"testing"

	"github.com/wesm/agentsview/internal/db"
	"github.com/wesm/agentsview/internal/llm"
	"github.com/wesm/agentsview/internal/signals"
)

type stubGuidanceClient struct {
	responses []llm.Response
	calls     int
}

func (s *stubGuidanceClient) Complete(
	_ context.Context, _ llm.Request,
) (llm.Response, error) {
	s.calls++
	if len(s.responses) == 0 {
		return llm.Response{}, nil
	}
	resp := s.responses[0]
	s.responses = s.responses[1:]
	return resp, nil
}

func TestEnrichGuidanceSignals_RewindAddsGeneratedTextAndCaches(t *testing.T) {
	client := &stubGuidanceClient{
		responses: []llm.Response{{
			Text:  `{"tangent_label":"deployment warning branch","rewind_reprompt_text":"Ignore the deployment warning branch and resume the auth refactor."}`,
			Model: "claude-haiku-4-5-20251001",
		}},
	}
	srv := &Server{
		guidanceClient: client,
		guidanceModel:  llm.DefaultGenerateModel,
		guidanceCache:  map[string]guidanceCacheEntry{},
	}
	view := sessionContextView{
		RewindSignal: &signals.RewindSignal{
			ShouldRewind:   true,
			RewindToTurn:   4,
			BadStretchFrom: 5,
			BadStretchTo:   6,
		},
	}
	summaries := []db.TurnSummary{
		{TurnIndex: 1, Topic: "auth refactor", Summary: "Started the auth refactor."},
		{TurnIndex: 4, Topic: "auth refactor", Summary: "Validated the auth constraint and had a clean plan."},
		{TurnIndex: 5, Topic: "deployment warning", Summary: "Pivoted into deployment warning debugging."},
		{TurnIndex: 6, Topic: "deployment warning", Summary: "Kept debugging the warning with no progress."},
	}

	srv.enrichGuidanceSignals(context.Background(), &view, summaries)

	if got := view.RewindSignal.RewindRepromptText; got == "" {
		t.Fatal("rewind_reprompt_text = empty, want generated text")
	}
	if got := view.RewindSignal.TangentLabel; got != "deployment warning branch" {
		t.Fatalf("tangent_label = %q, want deployment warning branch", got)
	}
	if got := view.RewindSignal.RepromptProvenance; got != "model-generated" {
		t.Fatalf("reprompt_provenance = %q, want model-generated", got)
	}
	if got := client.calls; got != 1 {
		t.Fatalf("client calls after first enrichment = %d, want 1", got)
	}

	secondView := sessionContextView{
		RewindSignal: &signals.RewindSignal{
			ShouldRewind:   true,
			RewindToTurn:   4,
			BadStretchFrom: 5,
			BadStretchTo:   6,
		},
	}
	srv.enrichGuidanceSignals(context.Background(), &secondView, summaries)

	if got := client.calls; got != 1 {
		t.Fatalf("client calls after cached enrichment = %d, want 1", got)
	}
	if got := secondView.RewindSignal.RewindRepromptText; got == "" {
		t.Fatal("cached rewind_reprompt_text = empty, want cached text")
	}
}

func TestEnrichGuidanceSignals_CompactAddsGeneratedText(t *testing.T) {
	client := &stubGuidanceClient{
		responses: []llm.Response{{
			Text:  `{"keep_items":["auth refactor plan","validated constraint B"],"drop_items":["deployment warning branch"],"compact_focus_text":"Preserve the auth refactor plan and validated constraint B. Drop the deployment warning branch before continuing."}`,
			Model: "claude-haiku-4-5-20251001",
		}},
	}
	srv := &Server{
		guidanceClient: client,
		guidanceModel:  llm.DefaultGenerateModel,
		guidanceCache:  map[string]guidanceCacheEntry{},
	}
	view := sessionContextView{
		Timeline: []contextTimelineTurn{
			{Turn: 1},
			{Turn: 2},
			{Turn: 3},
			{Turn: 4},
			{Turn: 5},
		},
		CompactSignal: &signals.CompactSignal{
			ShouldCompact: true,
			CompactFocus: []string{
				"tool outputs (32%)",
				"file reads (18%)",
			},
		},
	}
	summaries := []db.TurnSummary{
		{TurnIndex: 1, Topic: "auth refactor", Summary: "Sketched the auth refactor plan."},
		{TurnIndex: 2, Topic: "auth refactor", Summary: "Validated constraint A."},
		{TurnIndex: 3, Topic: "auth refactor", Summary: "Validated constraint B."},
		{TurnIndex: 4, Topic: "deployment warning", Summary: "Dove into deployment warning debugging."},
		{TurnIndex: 5, Topic: "deployment warning", Summary: "Repeated the warning investigation."},
	}

	srv.enrichGuidanceSignals(context.Background(), &view, summaries)

	if got := view.CompactSignal.CompactFocusText; got == "" {
		t.Fatal("compact_focus_text = empty, want generated text")
	}
	if len(view.CompactSignal.KeepItems) != 2 {
		t.Fatalf("keep_items len = %d, want 2", len(view.CompactSignal.KeepItems))
	}
	if len(view.CompactSignal.DropItems) != 1 {
		t.Fatalf("drop_items len = %d, want 1", len(view.CompactSignal.DropItems))
	}
	if got := view.CompactSignal.FocusProvenance; got != "model-generated" {
		t.Fatalf("focus_provenance = %q, want model-generated", got)
	}
}

func TestEnrichGuidanceSignals_FallsBackWhenSummariesDoNotCoverRange(t *testing.T) {
	client := &stubGuidanceClient{
		responses: []llm.Response{{
			Text: `{"tangent_label":"should not be used","rewind_reprompt_text":"should not be used"}`,
		}},
	}
	srv := &Server{
		guidanceClient: client,
		guidanceModel:  llm.DefaultGenerateModel,
		guidanceCache:  map[string]guidanceCacheEntry{},
	}
	view := sessionContextView{
		RewindSignal: &signals.RewindSignal{
			ShouldRewind:   true,
			RewindToTurn:   3,
			BadStretchFrom: 4,
			BadStretchTo:   5,
		},
	}
	summaries := []db.TurnSummary{
		{TurnIndex: 3, Topic: "auth refactor", Summary: "Still on the main task."},
		{TurnIndex: 5, Topic: "deployment warning", Summary: "Explored the tangent."},
	}

	srv.enrichGuidanceSignals(context.Background(), &view, summaries)

	if got := client.calls; got != 0 {
		t.Fatalf("client calls = %d, want 0 when summary coverage is incomplete", got)
	}
	if got := view.RewindSignal.RewindRepromptText; got != "" {
		t.Fatalf("rewind_reprompt_text = %q, want empty fallback", got)
	}
	if got := view.RewindSignal.TangentLabel; got != "" {
		t.Fatalf("tangent_label = %q, want empty fallback", got)
	}
}
