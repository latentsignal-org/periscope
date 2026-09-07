package sync

import (
	"log"
	"sync/atomic"
	"time"
)

// PhaseStats accumulates wall-clock time spent in each phase of the
// bulk-write hot path during a single sync pass. Each field is a
// nanosecond total. Fields are exported and read after the pass to keep
// the per-batch increment fast.
//
// The bulk path serializes a producer-consumer pipeline: worker
// goroutines parse files and the single collectAndBatch goroutine
// preps each result, scans for secrets, computes signals, and writes
// the batch. PhaseStats measures the time spent in each step of that
// consumer so a profile-driven analysis can attribute the wall clock.
type PhaseStats struct {
	PrepNanos      atomic.Int64 // prepareSessionWrite (toDBMessages, toDBSession, ...)
	ScanNanos      atomic.Int64 // computeSignalsAndSecrets (regex scan)
	WriteNanos     atomic.Int64 // db.WriteSessionBatch (one DB tx per batch)
	Batches        atomic.Int64
	BatchedWrites  atomic.Int64 // sessions written via bulk batches
	WriteBatchSize atomic.Int64 // sum of len(writes) across all batches
}

// Reset zeroes every counter. Called at the start of each sync pass so
// stats reflect only that pass.
func (p *PhaseStats) Reset() {
	p.PrepNanos.Store(0)
	p.ScanNanos.Store(0)
	p.WriteNanos.Store(0)
	p.Batches.Store(0)
	p.BatchedWrites.Store(0)
	p.WriteBatchSize.Store(0)
}

// Log emits a single-line summary of accumulated phase totals. It is
// a no-op when no batch ran (so non-bulk syncs stay quiet).
func (p *PhaseStats) Log(label string) {
	batches := p.Batches.Load()
	if batches == 0 {
		return
	}
	prep := time.Duration(p.PrepNanos.Load())
	scan := time.Duration(p.ScanNanos.Load())
	write := time.Duration(p.WriteNanos.Load())
	avg := float64(p.WriteBatchSize.Load()) / float64(batches)
	log.Printf(
		"%s phase totals: prep=%s scan=%s write=%s "+
			"batches=%d avg_batch=%.1f sessions_written=%d",
		label,
		prep.Round(time.Millisecond),
		scan.Round(time.Millisecond),
		write.Round(time.Millisecond),
		batches,
		avg,
		p.BatchedWrites.Load(),
	)
}
