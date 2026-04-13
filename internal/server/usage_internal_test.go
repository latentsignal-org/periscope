package server

import (
	"math"
	"testing"

	"github.com/wesm/agentsview/internal/db"
)

// approxEqual returns true if a and b are within eps (for
// float comparisons that have rounding from division).
func approxEqual(a, b, eps float64) bool {
	return math.Abs(a-b) <= eps
}

func TestComputeCacheStats_SavingsPassThrough(t *testing.T) {
	// SavingsVsUncached is now computed per-model in the DB
	// layer; computeCacheStats just forwards totals.CacheSavings.
	// Verify the pass-through at the positive, negative, and
	// zero boundaries so a future refactor that drops the field
	// trips a test.
	cases := []struct {
		name string
		in   float64
	}{
		{"positive", 4.65},
		{"negative", -0.75},
		{"zero", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cs := computeCacheStats(db.UsageTotals{
				CacheSavings: tc.in,
			})
			if !approxEqual(cs.SavingsVsUncached, tc.in, 1e-9) {
				t.Errorf(
					"SavingsVsUncached = %v, want %v",
					cs.SavingsVsUncached, tc.in,
				)
			}
		})
	}
}

func TestComputeCacheStats_ZeroTotalsIsZero(t *testing.T) {
	cs := computeCacheStats(db.UsageTotals{})
	if cs.SavingsVsUncached != 0 {
		t.Errorf("SavingsVsUncached = %v, want 0",
			cs.SavingsVsUncached)
	}
	if cs.HitRate != 0 {
		t.Errorf("HitRate = %v, want 0", cs.HitRate)
	}
}

func TestComputeCacheStats_HitRate(t *testing.T) {
	// 800 cache reads, 200 uncached inputs -> 0.80 hit rate.
	// (The HitRate denominator in this code is
	// cacheRead + input where input is already the uncached
	// portion — see the pass-through test below.)
	cs := computeCacheStats(db.UsageTotals{
		InputTokens:     200,
		CacheReadTokens: 800,
	})
	// denom = 800 + 200 = 1000; hit = 800/1000 = 0.80.
	if !approxEqual(cs.HitRate, 0.80, 1e-9) {
		t.Errorf("HitRate = %v, want ~0.80", cs.HitRate)
	}
}

func TestComputeCacheStats_UncachedPassesInputThrough(t *testing.T) {
	// Anthropic's input_tokens field is the NON-cached portion
	// of the input; cache_read and cache_creation are tracked
	// separately. UncachedInputTokens must therefore equal
	// InputTokens directly — not input minus the cache buckets,
	// which would double-subtract and wrongly drive the value
	// toward zero for any cached workload.
	cs := computeCacheStats(db.UsageTotals{
		InputTokens:         100,
		CacheReadTokens:     200,
		CacheCreationTokens: 50,
	})
	if cs.UncachedInputTokens != 100 {
		t.Errorf("UncachedInputTokens = %d, want 100",
			cs.UncachedInputTokens)
	}
	// And the cache buckets are reported verbatim alongside it.
	if cs.CacheReadTokens != 200 {
		t.Errorf("CacheReadTokens = %d, want 200",
			cs.CacheReadTokens)
	}
	if cs.CacheCreationTokens != 50 {
		t.Errorf("CacheCreationTokens = %d, want 50",
			cs.CacheCreationTokens)
	}
}
