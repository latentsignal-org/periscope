package server

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/wesm/agentsview/internal/db"
)

// ProjectTotal holds range-wide token and cost totals per project.
type ProjectTotal struct {
	Project             string  `json:"project"`
	InputTokens         int     `json:"inputTokens"`
	OutputTokens        int     `json:"outputTokens"`
	CacheCreationTokens int     `json:"cacheCreationTokens"`
	CacheReadTokens     int     `json:"cacheReadTokens"`
	Cost                float64 `json:"cost"`
}

// ModelTotal holds range-wide token and cost totals per model.
type ModelTotal struct {
	Model               string  `json:"model"`
	InputTokens         int     `json:"inputTokens"`
	OutputTokens        int     `json:"outputTokens"`
	CacheCreationTokens int     `json:"cacheCreationTokens"`
	CacheReadTokens     int     `json:"cacheReadTokens"`
	Cost                float64 `json:"cost"`
}

// AgentTotal holds range-wide token and cost totals per agent.
type AgentTotal struct {
	Agent               string  `json:"agent"`
	InputTokens         int     `json:"inputTokens"`
	OutputTokens        int     `json:"outputTokens"`
	CacheCreationTokens int     `json:"cacheCreationTokens"`
	CacheReadTokens     int     `json:"cacheReadTokens"`
	Cost                float64 `json:"cost"`
}

// CacheStats summarizes cache hit/miss for the period.
type CacheStats struct {
	CacheReadTokens     int     `json:"cacheReadTokens"`
	CacheCreationTokens int     `json:"cacheCreationTokens"`
	UncachedInputTokens int     `json:"uncachedInputTokens"`
	OutputTokens        int     `json:"outputTokens"`
	HitRate             float64 `json:"hitRate"`
	SavingsVsUncached   float64 `json:"savingsVsUncached"`
}

// Comparison holds the prior-period cost comparison.
type Comparison struct {
	PriorFrom      string  `json:"priorFrom"`
	PriorTo        string  `json:"priorTo"`
	PriorTotalCost float64 `json:"priorTotalCost"`
	DeltaPct       float64 `json:"deltaPct"`
}

// UsageSummaryResponse is the JSON shape for
// GET /api/v1/usage/summary.
type UsageSummaryResponse struct {
	From          string                `json:"from"`
	To            string                `json:"to"`
	Totals        db.UsageTotals        `json:"totals"`
	Daily         []db.DailyUsageEntry  `json:"daily"`
	ProjectTotals []ProjectTotal        `json:"projectTotals"`
	ModelTotals   []ModelTotal          `json:"modelTotals"`
	AgentTotals   []AgentTotal          `json:"agentTotals"`
	SessionCounts db.UsageSessionCounts `json:"sessionCounts"`
	CacheStats    CacheStats            `json:"cacheStats"`
	Comparison    *Comparison           `json:"comparison,omitempty"`
}

// parseUsageFilter extracts usage filter params from a request.
// Returns the filter and true on success; writes an error
// response and returns false on failure.
func parseUsageFilter(
	w http.ResponseWriter, r *http.Request,
) (db.UsageFilter, bool) {
	q := r.URL.Query()
	tz := q.Get("timezone")
	if tz == "" {
		tz = "UTC"
	}
	if _, err := time.LoadLocation(tz); err != nil {
		writeError(w, http.StatusBadRequest,
			"invalid timezone: "+tz)
		return db.UsageFilter{}, false
	}

	from, to := defaultDateRange(q.Get("from"), q.Get("to"))

	if !isValidDate(from) || !isValidDate(to) {
		writeError(w, http.StatusBadRequest,
			"invalid date format: use YYYY-MM-DD")
		return db.UsageFilter{}, false
	}
	if from > to {
		writeError(w, http.StatusBadRequest,
			"from must not be after to")
		return db.UsageFilter{}, false
	}

	return db.UsageFilter{
		From:           from,
		To:             to,
		Agent:          q.Get("agent"),
		Project:        q.Get("project"),
		ExcludeProject: q.Get("exclude_project"),
		ExcludeAgent:   q.Get("exclude_agent"),
		ExcludeModel:   q.Get("exclude_model"),
		Model:          q.Get("model"),
		Timezone:       tz,
		Breakdowns:     true,
	}, true
}

// foldProjectTotals sums daily project breakdowns into
// range-wide totals sorted by cost descending.
func foldProjectTotals(
	daily []db.DailyUsageEntry,
) []ProjectTotal {
	m := make(map[string]*ProjectTotal)
	for _, d := range daily {
		for _, pb := range d.ProjectBreakdowns {
			pt, ok := m[pb.Project]
			if !ok {
				pt = &ProjectTotal{Project: pb.Project}
				m[pb.Project] = pt
			}
			pt.InputTokens += pb.InputTokens
			pt.OutputTokens += pb.OutputTokens
			pt.CacheCreationTokens += pb.CacheCreationTokens
			pt.CacheReadTokens += pb.CacheReadTokens
			pt.Cost += pb.Cost
		}
	}
	out := make([]ProjectTotal, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Cost != out[j].Cost {
			return out[i].Cost > out[j].Cost
		}
		return out[i].Project < out[j].Project
	})
	return out
}

// foldModelTotals sums daily model breakdowns into range-wide
// totals sorted by cost descending.
func foldModelTotals(
	daily []db.DailyUsageEntry,
) []ModelTotal {
	m := make(map[string]*ModelTotal)
	for _, d := range daily {
		for _, mb := range d.ModelBreakdowns {
			mt, ok := m[mb.ModelName]
			if !ok {
				mt = &ModelTotal{Model: mb.ModelName}
				m[mb.ModelName] = mt
			}
			mt.InputTokens += mb.InputTokens
			mt.OutputTokens += mb.OutputTokens
			mt.CacheCreationTokens += mb.CacheCreationTokens
			mt.CacheReadTokens += mb.CacheReadTokens
			mt.Cost += mb.Cost
		}
	}
	out := make([]ModelTotal, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Cost != out[j].Cost {
			return out[i].Cost > out[j].Cost
		}
		return out[i].Model < out[j].Model
	})
	return out
}

// foldAgentTotals sums daily agent breakdowns into range-wide
// totals sorted by cost descending.
func foldAgentTotals(
	daily []db.DailyUsageEntry,
) []AgentTotal {
	m := make(map[string]*AgentTotal)
	for _, d := range daily {
		for _, ab := range d.AgentBreakdowns {
			at, ok := m[ab.Agent]
			if !ok {
				at = &AgentTotal{Agent: ab.Agent}
				m[ab.Agent] = at
			}
			at.InputTokens += ab.InputTokens
			at.OutputTokens += ab.OutputTokens
			at.CacheCreationTokens += ab.CacheCreationTokens
			at.CacheReadTokens += ab.CacheReadTokens
			at.Cost += ab.Cost
		}
	}
	out := make([]AgentTotal, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Cost != out[j].Cost {
			return out[i].Cost > out[j].Cost
		}
		return out[i].Agent < out[j].Agent
	})
	return out
}

// computeCacheStats derives cache hit/miss metrics from totals.
// SavingsVsUncached passes through totals.CacheSavings, which
// the DB layer computes per-message using each row's actual
// per-model rates — so mixed-model periods (e.g. Opus + Sonnet)
// report the right net delta instead of a single hard-coded
// proxy rate.
func computeCacheStats(t db.UsageTotals) CacheStats {
	// Anthropic reports input_tokens as the NON-cached portion
	// of the input (cache_read and cache_creation are separate
	// fields), so UncachedInputTokens is just t.InputTokens
	// directly — no subtraction.
	cs := CacheStats{
		CacheReadTokens:     t.CacheReadTokens,
		CacheCreationTokens: t.CacheCreationTokens,
		UncachedInputTokens: t.InputTokens,
		OutputTokens:        t.OutputTokens,
		SavingsVsUncached:   t.CacheSavings,
	}
	denominator := t.CacheReadTokens + t.InputTokens
	if denominator > 0 {
		cs.HitRate = float64(t.CacheReadTokens) /
			float64(denominator)
	}
	return cs
}

// computeComparison runs a prior-period query and computes
// the cost delta. Best-effort: logs and returns nil on error.
func (s *Server) computeComparison(
	r *http.Request, f db.UsageFilter, currentCost float64,
) *Comparison {
	fromT, err := time.Parse("2006-01-02", f.From)
	if err != nil {
		return nil
	}
	toT, err := time.Parse("2006-01-02", f.To)
	if err != nil {
		return nil
	}
	days := int(toT.Sub(fromT).Hours()/24) + 1
	priorTo := fromT.AddDate(0, 0, -1)
	priorFrom := priorTo.AddDate(0, 0, -(days - 1))

	priorFilter := db.UsageFilter{
		From:           priorFrom.Format("2006-01-02"),
		To:             priorTo.Format("2006-01-02"),
		Agent:          f.Agent,
		Project:        f.Project,
		Model:          f.Model,
		ExcludeProject: f.ExcludeProject,
		ExcludeAgent:   f.ExcludeAgent,
		ExcludeModel:   f.ExcludeModel,
		Timezone:       f.Timezone,
		Breakdowns:     false,
	}
	priorResult, err := s.db.GetDailyUsage(
		r.Context(), priorFilter,
	)
	if err != nil {
		log.Printf("usage comparison error: %v", err)
		return nil
	}

	c := &Comparison{
		PriorFrom:      priorFilter.From,
		PriorTo:        priorFilter.To,
		PriorTotalCost: priorResult.Totals.TotalCost,
	}
	if c.PriorTotalCost > 0 {
		c.DeltaPct = (currentCost - c.PriorTotalCost) /
			c.PriorTotalCost
	}
	return c
}

func (s *Server) handleUsageSummary(
	w http.ResponseWriter, r *http.Request,
) {
	f, ok := parseUsageFilter(w, r)
	if !ok {
		return
	}

	ctx := r.Context()

	result, err := s.db.GetDailyUsage(ctx, f)
	if err != nil {
		if handleContextError(w, err) {
			return
		}
		if handleReadOnly(w, err) {
			return
		}
		log.Printf("usage summary error: %v", err)
		writeError(w, http.StatusInternalServerError,
			"internal server error")
		return
	}

	// Session counts don't use breakdowns.
	scFilter := f
	scFilter.Breakdowns = false
	sessionCounts, err := s.db.GetUsageSessionCounts(
		ctx, scFilter,
	)
	if err != nil {
		if handleContextError(w, err) {
			return
		}
		if handleReadOnly(w, err) {
			return
		}
		log.Printf("usage session counts error: %v", err)
		writeError(w, http.StatusInternalServerError,
			"internal server error")
		return
	}

	cacheStats := computeCacheStats(result.Totals)

	comparison := s.computeComparison(
		r, f, result.Totals.TotalCost,
	)

	resp := UsageSummaryResponse{
		From:          f.From,
		To:            f.To,
		Totals:        result.Totals,
		Daily:         result.Daily,
		ProjectTotals: foldProjectTotals(result.Daily),
		ModelTotals:   foldModelTotals(result.Daily),
		AgentTotals:   foldAgentTotals(result.Daily),
		SessionCounts: sessionCounts,
		CacheStats:    cacheStats,
		Comparison:    comparison,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleUsageTopSessions(
	w http.ResponseWriter, r *http.Request,
) {
	f, ok := parseUsageFilter(w, r)
	if !ok {
		return
	}
	f.Breakdowns = false

	limit := 20
	if s := r.URL.Query().Get("limit"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			writeError(w, http.StatusBadRequest,
				"invalid limit parameter")
			return
		}
		limit = v
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	entries, err := s.db.GetTopSessionsByCost(
		r.Context(), f, limit,
	)
	if err != nil {
		if handleContextError(w, err) {
			return
		}
		if handleReadOnly(w, err) {
			return
		}
		log.Printf("usage top sessions error: %v", err)
		writeError(w, http.StatusInternalServerError,
			"internal server error")
		return
	}

	writeJSON(w, http.StatusOK, entries)
}
