package summarize

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/wesm/agentsview/internal/db"
	"github.com/wesm/agentsview/internal/llm"
)

type fakeClient struct {
	mu    sync.Mutex
	calls int
	resp  string
}

func (f *fakeClient) Complete(
	_ context.Context, _ llm.Request,
) (llm.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	r := f.resp
	if r == "" {
		r = `{"summary":"ok","intent":"implement","outcome":"progress","topic":"test"}`
	}
	return llm.Response{Text: r, Model: "claude-haiku-4-5"}, nil
}

type fakeStore struct {
	mu         sync.Mutex
	messages   []db.Message
	starred    []string
	summaries  []db.TurnSummary
	hasCalls   int
	upsertErrs error
}

func (s *fakeStore) GetAllMessages(
	_ context.Context, _ string,
) ([]db.Message, error) {
	return s.messages, nil
}

func (s *fakeStore) ListStarredSessionIDs(
	_ context.Context,
) ([]string, error) {
	return s.starred, nil
}

func (s *fakeStore) HasTurnSummary(
	_ context.Context, _ string, turn int, hash string,
) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hasCalls++
	for _, ts := range s.summaries {
		if ts.TurnIndex == turn && ts.ContentHash == hash {
			return true, nil
		}
	}
	return false, nil
}

func (s *fakeStore) UpsertTurnSummary(ts db.TurnSummary) error {
	if s.upsertErrs != nil {
		return s.upsertErrs
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.summaries = append(s.summaries, ts)
	return nil
}

// TestWorkerProcessStarred runs the worker against a fake client and
// store, verifying every turn is summarised once.
func TestWorkerProcessStarred(t *testing.T) {
	store := &fakeStore{
		messages: []db.Message{
			{Ordinal: 1, Role: "user", Content: "hello"},
			{Ordinal: 2, Role: "assistant", Content: "hi"},
			{Ordinal: 3, Role: "user", Content: "next"},
			{Ordinal: 4, Role: "assistant", Content: "ok"},
		},
		starred: []string{"sess-1"},
	}
	client := &fakeClient{}

	w := NewWorker(store, client, WorkerOptions{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()

	w.Enqueue("sess-1")

	// Wait for summaries to be produced. Bail fast if the worker is
	// misbehaving rather than hanging the whole test suite.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		store.mu.Lock()
		n := len(store.summaries)
		store.mu.Unlock()
		if n >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	<-done

	if len(store.summaries) != 2 {
		t.Fatalf("summaries = %d, want 2", len(store.summaries))
	}
	if client.calls != 2 {
		t.Fatalf("client calls = %d, want 2", client.calls)
	}
}

// TestWorkerSkipsExistingSummary verifies that a turn with a matching
// content hash is not re-summarised.
func TestWorkerSkipsExistingSummary(t *testing.T) {
	msgs := []db.Message{
		{Ordinal: 1, Role: "user", Content: "hello"},
		{Ordinal: 2, Role: "assistant", Content: "hi"},
	}
	bundle := BuildTurns(msgs)[0]
	store := &fakeStore{
		messages: msgs,
		summaries: []db.TurnSummary{{
			TurnIndex:   1,
			ContentHash: ContentHash(bundle),
		}},
	}
	client := &fakeClient{}

	w := NewWorker(store, client, WorkerOptions{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { w.Run(ctx); close(done) }()

	w.Enqueue("sess-1")

	// Give the worker a beat to claim/process/release.
	time.Sleep(100 * time.Millisecond)
	cancel()
	<-done

	if client.calls != 0 {
		t.Fatalf("client called %d times, want 0 (turn already summarised)",
			client.calls)
	}
	if len(store.summaries) != 1 {
		t.Fatalf("summaries = %d, want 1 unchanged", len(store.summaries))
	}
}

// TestWorkerEnabled reports enabled state based on client presence.
func TestWorkerEnabled(t *testing.T) {
	var nilClient llm.Client
	w := NewWorker(&fakeStore{}, nilClient, WorkerOptions{})
	if w.Enabled() {
		t.Fatalf("Enabled() = true with nil client")
	}
	w2 := NewWorker(&fakeStore{}, &fakeClient{}, WorkerOptions{})
	if !w2.Enabled() {
		t.Fatalf("Enabled() = false with non-nil client")
	}
}

func TestWorkerProcess_PersistsFallbackSummaryOnParseFailure(t *testing.T) {
	store := &fakeStore{
		messages: []db.Message{
			{Ordinal: 1, Role: "user", Content: "Debug the failing test in internal/server/context.go"},
			{Ordinal: 2, Role: "assistant", Content: "Investigated the failing test."},
		},
	}
	client := &fakeClient{resp: "```json"}
	w := NewWorker(store, client, WorkerOptions{})

	hadErrors := w.process(context.Background(), "sess-1")

	if hadErrors {
		t.Fatal("process reported errors, want fallback summary to avoid retry loop")
	}
	if len(store.summaries) != 1 {
		t.Fatalf("summaries = %d, want 1", len(store.summaries))
	}
	if store.summaries[0].Summary == "" {
		t.Fatal("persisted fallback summary is empty")
	}
	if store.summaries[0].Model == "" {
		t.Fatal("persisted fallback model is empty")
	}
}
