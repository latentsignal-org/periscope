package server

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"slices"
	"strings"

	"github.com/wesm/agentsview/internal/db"
	"github.com/wesm/agentsview/internal/signals"
)

const (
	contextCategorySystem          = "system_prompt_and_tool_definitions"
	contextCategoryUser            = "user_messages"
	contextCategoryAssistant       = "assistant_messages"
	contextCategoryThinking        = "thinking"
	contextCategoryToolCalls       = "tool_calls"
	contextCategoryToolOutputs     = "tool_outputs"
	contextCategoryFileReads       = "file_reads"
	contextCategorySearchResults   = "search_results"
	contextCategorySummaries       = "summaries_and_compacted_handoffs"
	contextCategorySubagentOutputs = "subagent_outputs"
	contextCategoryFreeSpace       = "free_space"
	contextCategoryOther           = "other"

	contextProvenanceMeasured  = "measured"
	contextProvenanceInferred  = "inferred"
	contextProvenanceEstimated = "estimated"
	contextProvenanceUnknown   = "unknown"

	contextRowGranularityMessage = "message"
)

type contextCapacity struct {
	MaxTokens  int    `json:"max_tokens"`
	Provenance string `json:"provenance"`
	Model      string `json:"model,omitempty"`
	Agent      string `json:"agent,omitempty"`
}

type contextSummary struct {
	TokensInUse         int     `json:"tokens_in_use"`
	TokensProvenance    string  `json:"tokens_provenance"`
	PercentConsumed     float64 `json:"percent_consumed"`
	PercentProvenance   string  `json:"percent_provenance"`
	RemainingTokens     int     `json:"remaining_tokens"`
	RemainingKnown      bool    `json:"remaining_known"`
	VisibleRowCount     int     `json:"visible_row_count"`
	VisibleSinceOrdinal int     `json:"visible_since_ordinal"`
	LastUpdatedAt       string  `json:"last_updated_at,omitempty"`
	RowGranularity      string  `json:"row_granularity"`
}

type contextCompositionItem struct {
	Category   string  `json:"category"`
	Tokens     int     `json:"tokens"`
	Percentage float64 `json:"percentage"`
	Provenance string  `json:"provenance"`
}

type contextCategoryValue struct {
	Category string `json:"category"`
	Tokens   int    `json:"tokens"`
}

type contextTimelineRow struct {
	Ordinal              int                    `json:"ordinal"`
	Timestamp            string                 `json:"timestamp,omitempty"`
	Label                string                 `json:"label"`
	Granularity          string                 `json:"granularity"`
	DeltaTokens          int                    `json:"delta_tokens"`
	DeltaProvenance      string                 `json:"delta_provenance"`
	CumulativeTokens     int                    `json:"cumulative_tokens"`
	CumulativeProvenance string                 `json:"cumulative_provenance"`
	DominantCategory     string                 `json:"dominant_category,omitempty"`
	Categories           []contextCategoryValue `json:"categories"`
	Markers              []string               `json:"markers,omitempty"`
	Annotations          []string               `json:"annotations,omitempty"`
}

type contextSupports struct {
	LiveUpdates       bool   `json:"live_updates"`
	StandaloneRoute   bool   `json:"standalone_route"`
	EmbeddedTab       bool   `json:"embedded_tab"`
	TranscriptJump    bool   `json:"transcript_jump"`
	RowGranularity    string `json:"row_granularity"`
	CompactionTrimmed bool   `json:"compaction_trimmed"`
}

type sessionContextResponse struct {
	Summary     contextSummary           `json:"summary"`
	Capacity    contextCapacity          `json:"capacity"`
	Composition []contextCompositionItem `json:"composition"`
	Supports    contextSupports          `json:"supports"`
	Warnings    []string                 `json:"warnings,omitempty"`
}

type sessionContextTimelineResponse struct {
	Timeline []contextTimelineRow `json:"timeline"`
	Supports contextSupports      `json:"supports"`
	Warnings []string             `json:"warnings,omitempty"`
}

type sessionContextView struct {
	Summary     contextSummary
	Capacity    contextCapacity
	Composition []contextCompositionItem
	Timeline    []contextTimelineRow
	Supports    contextSupports
	Warnings    []string
}

type contextRowCalc struct {
	row               contextTimelineRow
	estimatedTotal    int
	categoryTotals    map[string]int
	hasMeasuredTokens bool
	measuredTokens    int
}

func (s *Server) handleGetSessionContext(
	w http.ResponseWriter, r *http.Request,
) {
	view, err := s.buildSessionContextView(r.Context(), r.PathValue("id"))
	if err != nil {
		if handleContextError(w, err) {
			return
		}
		if errors.Is(err, db.ErrReadOnly) && handleReadOnly(w, err) {
			return
		}
		if errors.Is(err, errSessionNotFound) {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, sessionContextResponse{
		Summary:     view.Summary,
		Capacity:    view.Capacity,
		Composition: view.Composition,
		Supports:    view.Supports,
		Warnings:    view.Warnings,
	})
}

func (s *Server) handleGetSessionContextTimeline(
	w http.ResponseWriter, r *http.Request,
) {
	view, err := s.buildSessionContextView(r.Context(), r.PathValue("id"))
	if err != nil {
		if handleContextError(w, err) {
			return
		}
		if errors.Is(err, db.ErrReadOnly) && handleReadOnly(w, err) {
			return
		}
		if errors.Is(err, errSessionNotFound) {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, sessionContextTimelineResponse{
		Timeline: view.Timeline,
		Supports: view.Supports,
		Warnings: view.Warnings,
	})
}

var errSessionNotFound = errors.New("session not found")

func (s *Server) buildSessionContextView(
	ctx context.Context, sessionID string,
) (sessionContextView, error) {
	session, err := s.db.GetSession(ctx, sessionID)
	if err != nil {
		return sessionContextView{}, err
	}
	if session == nil {
		return sessionContextView{}, errSessionNotFound
	}
	msgs, err := s.db.GetAllMessages(ctx, sessionID)
	if err != nil {
		return sessionContextView{}, err
	}
	return computeSessionContextView(*session, msgs), nil
}

func computeSessionContextView(
	session db.Session, msgs []db.Message,
) sessionContextView {
	visible, compactionTrimmed, visibleSince := trimContextMessages(msgs)
	model := pickPrimaryModel(visible, msgs)
	capacity := resolveContextCapacity(session, visible, model)

	rows := make([]contextRowCalc, 0, len(visible))
	compositionTotals := map[string]int{}
	lastUpdatedAt := ""
	warnings := []string{}

	prevMeasured := 0
	prevMeasuredValid := false
	cumulative := 0
	cumulativeProv := contextProvenanceEstimated
	hasAnyMeasured := false

	for _, msg := range visible {
		rc := buildContextRow(msg)
		lastUpdatedAt = latestNonEmpty(lastUpdatedAt, msg.Timestamp)

		if rc.hasMeasuredTokens {
			hasAnyMeasured = true
			rc.row.CumulativeTokens = rc.measuredTokens
			rc.row.CumulativeProvenance = contextProvenanceMeasured
			if prevMeasuredValid {
				rc.row.DeltaTokens = max(0, rc.measuredTokens-prevMeasured)
			} else {
				rc.row.DeltaTokens = max(0, rc.measuredTokens)
			}
			rc.row.DeltaProvenance = contextProvenanceMeasured
			prevMeasured = rc.measuredTokens
			prevMeasuredValid = true
			cumulative = rc.measuredTokens
			cumulativeProv = contextProvenanceMeasured
		} else {
			rc.row.DeltaTokens = rc.estimatedTotal
			rc.row.DeltaProvenance = contextProvenanceEstimated
			cumulative += rc.estimatedTotal
			rc.row.CumulativeTokens = cumulative
			rc.row.CumulativeProvenance = contextProvenanceEstimated
			cumulativeProv = contextProvenanceEstimated
		}

		for category, tokens := range rc.categoryTotals {
			compositionTotals[category] += tokens
		}
		rows = append(rows, rc)
	}

	applySpikeMarkers(rows)

	currentTokens := cumulative
	tokenProv := cumulativeProv
	if len(rows) == 0 {
		tokenProv = contextProvenanceUnknown
	}

	remaining := 0
	remainingKnown := false
	percent := 0.0
	percentProv := contextProvenanceUnknown
	if capacity.MaxTokens > 0 {
		remainingKnown = true
		remaining = max(0, capacity.MaxTokens-currentTokens)
		percent = min(100, (float64(currentTokens)/float64(capacity.MaxTokens))*100)
		percentProv = weakerProvenance(tokenProv, capacity.Provenance)
	}

	if !hasAnyMeasured {
		warnings = append(warnings,
			"This session has no recorded context-token snapshots after the visible boundary; totals are estimated from stored content and tool payload sizes.")
	}
	if compactionTrimmed {
		warnings = append(warnings,
			"Visible context starts at the latest compaction boundary. Earlier history is intentionally excluded in V1.")
	}
	if capacity.Provenance == contextProvenanceUnknown {
		warnings = append(warnings,
			"Context capacity is unknown for this session; occupancy and free-space values are omitted.")
	}
	warnings = append(warnings,
		"Timeline rows are rendered at message granularity in V1.")

	composition := buildComposition(compositionTotals, currentTokens, capacity)

	return sessionContextView{
		Summary: contextSummary{
			TokensInUse:         currentTokens,
			TokensProvenance:    tokenProv,
			PercentConsumed:     percent,
			PercentProvenance:   percentProv,
			RemainingTokens:     remaining,
			RemainingKnown:      remainingKnown,
			VisibleRowCount:     len(rows),
			VisibleSinceOrdinal: visibleSince,
			LastUpdatedAt:       lastUpdatedAt,
			RowGranularity:      contextRowGranularityMessage,
		},
		Capacity:    capacity,
		Composition: composition,
		Timeline:    mapTimelineRows(rows),
		Supports: contextSupports{
			LiveUpdates:       true,
			StandaloneRoute:   true,
			EmbeddedTab:       true,
			TranscriptJump:    false,
			RowGranularity:    contextRowGranularityMessage,
			CompactionTrimmed: compactionTrimmed,
		},
		Warnings: warnings,
	}
}

func mapTimelineRows(rows []contextRowCalc) []contextTimelineRow {
	out := make([]contextTimelineRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.row)
	}
	return out
}

func buildComposition(
	totals map[string]int,
	currentTokens int,
	capacity contextCapacity,
) []contextCompositionItem {
	categories := make([]string, 0, len(totals))
	sum := 0
	for category, tokens := range totals {
		if tokens <= 0 {
			continue
		}
		sum += tokens
		categories = append(categories, category)
	}
	slices.SortStableFunc(categories, func(a, b string) int {
		if totals[a] == totals[b] {
			return strings.Compare(a, b)
		}
		return totals[b] - totals[a]
	})

	displayTotals := map[string]int{}
	for _, category := range categories {
		displayTotals[category] = totals[category]
	}
	if sum > 0 && currentTokens > 0 && sum != currentTokens {
		remaining := currentTokens
		for i, category := range categories {
			value := int(math.Round(float64(totals[category]) * float64(currentTokens) / float64(sum)))
			if i == len(categories)-1 {
				value = remaining
			}
			if value < 0 {
				value = 0
			}
			displayTotals[category] = value
			remaining -= value
		}
	}

	items := make([]contextCompositionItem, 0, len(categories)+1)
	for _, category := range categories {
		tokens := displayTotals[category]
		pct := 0.0
		if currentTokens > 0 {
			pct = (float64(tokens) / float64(currentTokens)) * 100
		}
		items = append(items, contextCompositionItem{
			Category:   category,
			Tokens:     tokens,
			Percentage: pct,
			Provenance: contextProvenanceEstimated,
		})
	}
	if capacity.MaxTokens > 0 && capacity.MaxTokens > currentTokens {
		free := capacity.MaxTokens - currentTokens
		items = append(items, contextCompositionItem{
			Category:   contextCategoryFreeSpace,
			Tokens:     free,
			Percentage: (float64(free) / float64(capacity.MaxTokens)) * 100,
			Provenance: capacity.Provenance,
		})
	}
	return items
}

func applySpikeMarkers(rows []contextRowCalc) {
	deltas := make([]int, 0, len(rows))
	for _, row := range rows {
		if row.row.DeltaTokens > 0 {
			deltas = append(deltas, row.row.DeltaTokens)
		}
	}
	if len(deltas) < 3 {
		return
	}
	slices.Sort(deltas)
	median := deltas[len(deltas)/2]
	if median <= 0 {
		return
	}
	threshold := int(math.Ceil(float64(median) * 2.5))
	for i := range rows {
		if rows[i].row.DeltaTokens < threshold {
			continue
		}
		rows[i].row.Markers = append(rows[i].row.Markers, "spike")
		switch rows[i].row.DominantCategory {
		case contextCategorySearchResults:
			rows[i].row.Annotations = append(rows[i].row.Annotations,
				"Search-heavy row increased visible context sharply.")
		case contextCategoryFileReads:
			rows[i].row.Annotations = append(rows[i].row.Annotations,
				"File-read payloads dominate this growth spike.")
		case contextCategoryToolOutputs:
			rows[i].row.Annotations = append(rows[i].row.Annotations,
				"Tool output volume dominates this growth spike.")
		case contextCategorySubagentOutputs:
			rows[i].row.Annotations = append(rows[i].row.Annotations,
				"Subagent output dominates this growth spike.")
		}
	}
}

func buildContextRow(msg db.Message) contextRowCalc {
	categoryTotals := map[string]int{}
	markers := []string{}
	annotations := []string{}

	addTokens := func(category string, tokens int) {
		if tokens <= 0 {
			return
		}
		categoryTotals[category] += tokens
	}

	if msg.IsCompactBoundary || msg.SourceSubtype == "compact_boundary" {
		addTokens(contextCategorySummaries, estimateTokens(msg.ContentLength))
		markers = append(markers, "compaction")
		annotations = append(annotations, "Compaction boundary: visible context starts here.")
	} else if msg.IsSystem {
		addTokens(contextCategorySystem, estimateTokens(msg.ContentLength))
	} else if msg.Role == "user" {
		addTokens(contextCategoryUser, estimateTokens(msg.ContentLength))
	} else {
		thinkingLen, assistantLen := splitThinkingContent(msg.Content)
		addTokens(contextCategoryThinking, estimateTokens(thinkingLen))
		addTokens(contextCategoryAssistant, estimateTokens(assistantLen))
	}

	for _, tc := range msg.ToolCalls {
		addTokens(contextCategoryToolCalls, estimateTokens(len(tc.InputJSON)))
		outputCategory := classifyToolOutputCategory(tc)
		outputLen := tc.ResultContentLength
		if len(tc.ResultEvents) > 0 {
			outputLen = 0
			for _, ev := range tc.ResultEvents {
				outputLen += ev.ContentLength
			}
		}
		addTokens(outputCategory, estimateTokens(outputLen))
		if outputCategory == contextCategorySubagentOutputs && !slices.Contains(markers, "subagent") {
			markers = append(markers, "subagent")
		}
	}

	estimatedTotal := 0
	dominantCategory := ""
	dominantTokens := 0
	categories := make([]contextCategoryValue, 0, len(categoryTotals))
	for category, tokens := range categoryTotals {
		estimatedTotal += tokens
		categories = append(categories, contextCategoryValue{
			Category: category,
			Tokens:   tokens,
		})
		if tokens > dominantTokens {
			dominantTokens = tokens
			dominantCategory = category
		}
	}
	slices.SortStableFunc(categories, func(a, b contextCategoryValue) int {
		if a.Tokens == b.Tokens {
			return strings.Compare(a.Category, b.Category)
		}
		return b.Tokens - a.Tokens
	})

	hasCtx, _ := msg.TokenPresence()
	row := contextTimelineRow{
		Ordinal:          msg.Ordinal,
		Timestamp:        msg.Timestamp,
		Label:            rowLabel(msg),
		Granularity:      contextRowGranularityMessage,
		DominantCategory: dominantCategory,
		Categories:       categories,
		Markers:          markers,
		Annotations:      annotations,
	}

	return contextRowCalc{
		row:               row,
		estimatedTotal:    estimatedTotal,
		categoryTotals:    categoryTotals,
		hasMeasuredTokens: hasCtx,
		measuredTokens:    msg.ContextTokens,
	}
}

func rowLabel(msg db.Message) string {
	if msg.IsCompactBoundary || msg.SourceSubtype == "compact_boundary" {
		return "Compaction seed"
	}
	switch msg.Role {
	case "user":
		return "User message"
	case "assistant":
		if msg.HasToolUse {
			return "Assistant tool turn"
		}
		return "Assistant message"
	case "system":
		return "System message"
	default:
		return "Message"
	}
}

func classifyToolOutputCategory(tc db.ToolCall) string {
	if tc.SubagentSessionID != "" || tc.Category == "Task" {
		return contextCategorySubagentOutputs
	}
	for _, ev := range tc.ResultEvents {
		if ev.SubagentSessionID != "" {
			return contextCategorySubagentOutputs
		}
	}
	switch tc.Category {
	case "Read":
		return contextCategoryFileReads
	case "Grep", "Glob":
		return contextCategorySearchResults
	default:
		name := strings.ToLower(tc.ToolName)
		if strings.Contains(name, "search") || strings.Contains(name, "grep") || strings.Contains(name, "glob") {
			return contextCategorySearchResults
		}
		if strings.Contains(name, "read") || strings.Contains(name, "view") {
			return contextCategoryFileReads
		}
		return contextCategoryToolOutputs
	}
}

func trimContextMessages(msgs []db.Message) ([]db.Message, bool, int) {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].IsCompactBoundary || msgs[i].SourceSubtype == "compact_boundary" {
			return msgs[i:], true, msgs[i].Ordinal
		}
	}
	if len(msgs) == 0 {
		return msgs, false, 0
	}
	return msgs, false, msgs[0].Ordinal
}

func resolveContextCapacity(
	session db.Session, msgs []db.Message, model string,
) contextCapacity {
	if recorded := extractRecordedCapacity(msgs); recorded > 0 {
		return contextCapacity{
			MaxTokens:  recorded,
			Provenance: contextProvenanceMeasured,
			Model:      model,
			Agent:      session.Agent,
		}
	}
	if inferred := signals.LookupContextWindowSize(model); inferred > 0 {
		return contextCapacity{
			MaxTokens:  inferred,
			Provenance: contextProvenanceInferred,
			Model:      model,
			Agent:      session.Agent,
		}
	}
	return contextCapacity{
		Provenance: contextProvenanceUnknown,
		Model:      model,
		Agent:      session.Agent,
	}
}

func extractRecordedCapacity(msgs []db.Message) int {
	keys := []string{
		"max_context_tokens",
		"max_input_tokens",
		"context_window",
		"context_window_tokens",
	}
	for i := len(msgs) - 1; i >= 0; i-- {
		if len(msgs[i].TokenUsage) == 0 {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal(msgs[i].TokenUsage, &payload); err != nil {
			continue
		}
		for _, key := range keys {
			if v, ok := payload[key]; ok {
				if n := anyToInt(v); n > 0 {
					return n
				}
			}
		}
	}
	return 0
}

func anyToInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case float32:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}

func pickPrimaryModel(visible, all []db.Message) string {
	for i := len(visible) - 1; i >= 0; i-- {
		if visible[i].Model != "" {
			return visible[i].Model
		}
	}
	for i := len(all) - 1; i >= 0; i-- {
		if all[i].Model != "" {
			return all[i].Model
		}
	}
	return ""
}

func splitThinkingContent(content string) (thinkingLen, assistantLen int) {
	const startTag = "[Thinking]\n"
	const endTag = "\n[/Thinking]"
	remaining := content
	for {
		start := strings.Index(remaining, startTag)
		if start < 0 {
			assistantLen += len(remaining)
			break
		}
		assistantLen += start
		remaining = remaining[start+len(startTag):]
		end := strings.Index(remaining, endTag)
		if end < 0 {
			thinkingLen += len(remaining)
			break
		}
		thinkingLen += end
		remaining = remaining[end+len(endTag):]
	}
	return thinkingLen, assistantLen
}

func estimateTokens(length int) int {
	if length <= 0 {
		return 0
	}
	return max(1, int(math.Ceil(float64(length)/4.0)))
}

func weakerProvenance(a, b string) string {
	rank := map[string]int{
		contextProvenanceMeasured:  0,
		contextProvenanceInferred:  1,
		contextProvenanceEstimated: 2,
		contextProvenanceUnknown:   3,
	}
	if rank[a] >= rank[b] {
		return a
	}
	return b
}

func latestNonEmpty(current, next string) string {
	if next != "" {
		return next
	}
	return current
}
