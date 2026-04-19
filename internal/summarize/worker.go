package summarize

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/wesm/agentsview/internal/db"
	"github.com/wesm/agentsview/internal/llm"
)

// Store is the subset of db.Store the worker needs.
type Store interface {
	GetAllMessages(ctx context.Context, sessionID string) ([]db.Message, error)
	ListStarredSessionIDs(ctx context.Context) ([]string, error)
	HasTurnSummary(ctx context.Context, sessionID string, turnIndex int, contentHash string) (bool, error)
	UpsertTurnSummary(s db.TurnSummary) error
}

// Notifier is called after a session finishes a summarisation pass so
// connected Context pages can refresh. Optional.
type Notifier func(sessionID string)

// Worker summarises turns for starred sessions in the background.
//
// Lifecycle:
//
//	w := NewWorker(store, client, WorkerOptions{...})
//	go w.Run(ctx)
//	w.Enqueue(sessionID)   // fire and forget
//	w.ReconcileStarred(ctx) // on boot
type Worker struct {
	store    Store
	client   llm.Client
	model    string
	notifier Notifier
	logger   *log.Logger

	queue chan string

	mu       sync.Mutex
	inflight map[string]bool
}

// WorkerOptions configures the worker.
type WorkerOptions struct {
	Model     string
	QueueSize int
	Notifier  Notifier
	Logger    *log.Logger
}

// NewWorker builds a worker. A nil client is allowed — the worker
// will simply drop jobs (used when ANTHROPIC_API_KEY is unset).
func NewWorker(store Store, client llm.Client, opts WorkerOptions) *Worker {
	qs := opts.QueueSize
	if qs <= 0 {
		qs = 32
	}
	model := opts.Model
	if model == "" {
		model = llm.DefaultSummaryModel
	}
	logger := opts.Logger
	if logger == nil {
		logger = log.Default()
	}
	return &Worker{
		store:    store,
		client:   client,
		model:    model,
		notifier: opts.Notifier,
		logger:   logger,
		queue:    make(chan string, qs),
		inflight: map[string]bool{},
	}
}

// Enabled reports whether the worker has an LLM client configured.
func (w *Worker) Enabled() bool {
	return w != nil && w.client != nil
}

// Enqueue schedules a session for summarisation. Duplicate enqueues
// for the same session are coalesced. Returns without blocking when
// the queue is full; the reconcile pass will catch it next time.
func (w *Worker) Enqueue(sessionID string) {
	if w == nil || sessionID == "" {
		return
	}
	select {
	case w.queue <- sessionID:
	default:
		w.logger.Printf(
			"summarize: queue full, dropping %s", sessionID,
		)
	}
}

// Run processes the queue until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	if w == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case sessionID := <-w.queue:
			if !w.claim(sessionID) {
				continue
			}
			hadErrors := w.process(ctx, sessionID)
			w.release(sessionID)
			if w.notifier != nil {
				w.notifier(sessionID)
			}
			// Re-enqueue after a backoff when turns failed (e.g.
			// rate-limited turns that exhausted retries). This
			// ensures progress is eventually made rather than
			// leaving sessions permanently stuck.
			if hadErrors {
				go func(id string) {
					if err := sleepCtx(ctx, retryDelay); err != nil {
						return
					}
					w.Enqueue(id)
				}(sessionID)
			}
		}
	}
}

// retryDelay is how long to wait before re-enqueueing a session
// that had one or more turn-level failures (e.g. rate limits).
const retryDelay = 2 * time.Minute

// ReconcileStarred enqueues every starred session whose summaries are
// incomplete. Intended to run on boot.
func (w *Worker) ReconcileStarred(ctx context.Context) {
	if w == nil || !w.Enabled() {
		return
	}
	ids, err := w.store.ListStarredSessionIDs(ctx)
	if err != nil {
		w.logger.Printf("summarize: reconcile list starred: %v", err)
		return
	}
	for _, id := range ids {
		w.Enqueue(id)
	}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (w *Worker) claim(id string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.inflight[id] {
		return false
	}
	w.inflight[id] = true
	return true
}

func (w *Worker) release(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.inflight, id)
}

// process summarises every unseen turn of the session. Returns true
// when at least one turn failed (API error, upsert error) so the
// caller can schedule a retry pass.
func (w *Worker) process(ctx context.Context, sessionID string) (hadErrors bool) {
	if !w.Enabled() {
		return false
	}
	msgs, err := w.store.GetAllMessages(ctx, sessionID)
	if err != nil {
		w.logger.Printf(
			"summarize: load messages %s: %v", sessionID, err,
		)
		return true
	}
	turns := BuildTurns(msgs)
	if len(turns) == 0 {
		return false
	}

	generated := 0
	start := time.Now()
	for _, bundle := range turns {
		if ctx.Err() != nil {
			return false
		}
		hash := ContentHash(bundle)
		exists, err := w.store.HasTurnSummary(
			ctx, sessionID, bundle.TurnIndex, hash,
		)
		if err != nil {
			w.logger.Printf(
				"summarize: check %s:%d: %v",
				sessionID, bundle.TurnIndex, err,
			)
			hadErrors = true
			continue
		}
		if exists {
			continue
		}
		summary, err := SummarizeTurn(ctx, w.client, w.model, bundle)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return false
			}
			w.logger.Printf(
				"summarize: generate %s:%d: %v",
				sessionID, bundle.TurnIndex, err,
			)
			hadErrors = true
			continue
		}
		files, _ := json.Marshal(summary.FilesTouched)
		tags, _ := json.Marshal(summary.Tags)
		if err := w.store.UpsertTurnSummary(db.TurnSummary{
			SessionID:     sessionID,
			TurnIndex:     bundle.TurnIndex,
			StartOrdinal:  bundle.StartOrdinal,
			EndOrdinal:    bundle.EndOrdinal,
			ContentHash:   hash,
			Summary:       summary.Summary,
			Intent:        summary.Intent,
			Outcome:       summary.Outcome,
			Topic:         summary.Topic,
			FilesTouched:  string(files),
			Tags:          string(tags),
			Model:         summary.Model,
			PromptVersion: PromptVersion,
		}); err != nil {
			w.logger.Printf(
				"summarize: upsert %s:%d: %v",
				sessionID, bundle.TurnIndex, err,
			)
			hadErrors = true
			continue
		}
		generated++
	}
	if generated > 0 {
		w.logger.Printf(
			"summarize: %s generated %d turn(s) in %s",
			sessionID, generated, time.Since(start).Round(time.Millisecond),
		)
	}
	return hadErrors
}
