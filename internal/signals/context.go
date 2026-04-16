package signals

import (
	"sort"
	"strings"
)

// ContextTokenRow represents a single message's context token
// measurement.
type ContextTokenRow struct {
	ContextTokens    int
	HasContextTokens bool
}

// midTaskWindowBefore is the count of tools immediately before a
// boundary used to decide whether post-boundary work overlaps.
const midTaskWindowBefore = 10

// midTaskWindowAfter is the count of tools immediately after a
// boundary checked for overlap with the pre-boundary window.
const midTaskWindowAfter = 5

// midTaskOverlapThreshold is the minimum number of tool names
// shared between the pre- and post-boundary windows that flags a
// boundary as mid-task (i.e., the agent likely lost context and
// is repeating earlier work).
const midTaskOverlapThreshold = 2

// CountMidTaskCompactions returns the number of compact boundary
// messages where the first few tool calls after the boundary
// share names with the tool calls immediately before it. A
// boundary surrounded by overlapping tool work is a strong
// signal that compaction interrupted active work and the agent
// is repeating itself.
//
// boundaryOrdinals must be sorted ascending. toolNamesByOrdinal
// holds the ordinal of the message that issued each call paired
// with its tool name, in chronological order.
func CountMidTaskCompactions(
	boundaryOrdinals []int,
	toolCalls []ToolCallOrdinal,
) int {
	if len(boundaryOrdinals) == 0 || len(toolCalls) == 0 {
		return 0
	}
	count := 0
	for _, b := range boundaryOrdinals {
		before := toolWindowBefore(toolCalls, b, midTaskWindowBefore)
		after := toolWindowAfter(toolCalls, b, midTaskWindowAfter)
		if len(before) == 0 || len(after) == 0 {
			continue
		}
		beforeSet := map[string]struct{}{}
		for _, name := range before {
			beforeSet[name] = struct{}{}
		}
		// Count distinct shared names so a single tool repeated
		// many times after the boundary doesn't inflate the
		// overlap into a false mid-task signal.
		matched := map[string]struct{}{}
		for _, name := range after {
			if _, ok := beforeSet[name]; ok {
				matched[name] = struct{}{}
			}
		}
		if len(matched) >= midTaskOverlapThreshold {
			count++
		}
	}
	return count
}

// ToolCallOrdinal pairs a tool call with the ordinal of the
// message that emitted it. Used by the mid-task compaction
// detector to locate calls relative to a boundary.
type ToolCallOrdinal struct {
	MessageOrdinal int
	ToolName       string
}

// toolWindowBefore returns up to n tool names from calls strictly
// before the given ordinal, taking the most recent ones.
func toolWindowBefore(
	calls []ToolCallOrdinal, ordinal, n int,
) []string {
	var names []string
	for _, c := range calls {
		if c.MessageOrdinal < ordinal {
			names = append(names, c.ToolName)
		}
	}
	if len(names) > n {
		names = names[len(names)-n:]
	}
	return names
}

// toolWindowAfter returns up to n tool names from calls strictly
// after the given ordinal, taking the earliest ones.
func toolWindowAfter(
	calls []ToolCallOrdinal, ordinal, n int,
) []string {
	var names []string
	for _, c := range calls {
		if c.MessageOrdinal > ordinal {
			names = append(names, c.ToolName)
			if len(names) >= n {
				break
			}
		}
	}
	return names
}

// ContextPressureResult holds computed context pressure metrics.
type ContextPressureResult struct {
	CompactionCount int
	PressureMax     *float64 // nil when data unavailable
}

// contextWindowSizes maps model name prefixes to their context
// window sizes in tokens.
var contextWindowSizes = map[string]int{
	"claude-opus-4-6":   1_000_000,
	"claude-sonnet-4-6": 200_000,
	"claude-sonnet-4-5": 200_000,
	"claude-haiku-4-5":  200_000,
	"claude-3-5-sonnet": 200_000,
	"claude-3-opus":     200_000,
	"claude-3-haiku":    200_000,
	"gpt-4o-mini":       128_000,
	"gpt-4o":            128_000,
	"gpt-4-turbo":       128_000,
	"o3":                200_000,
	"o4-mini":           200_000,
	"gemini-2.5-pro":    1_000_000,
	"gemini-2.5-flash":  1_000_000,
	"gemini-2.0-flash":  1_000_000,
}

// sortedPrefixes holds model prefixes sorted longest-first so
// "gpt-4o-mini" matches before "gpt-4o".
var sortedPrefixes []string

func init() {
	sortedPrefixes = make([]string, 0, len(contextWindowSizes))
	for k := range contextWindowSizes {
		sortedPrefixes = append(sortedPrefixes, k)
	}
	sort.Slice(sortedPrefixes, func(i, j int) bool {
		return len(sortedPrefixes[i]) > len(sortedPrefixes[j])
	})
}

// ComputeContextPressure computes compaction count and peak context
// pressure from an ordered slice of token rows. Pure computation,
// no DB access.
func ComputeContextPressure(
	tokens []ContextTokenRow,
	peakContextTokens int,
	model string,
) ContextPressureResult {
	var r ContextPressureResult
	r.CompactionCount = countCompactions(tokens)
	r.PressureMax = computePressure(peakContextTokens, model)
	return r
}

// countCompactions counts >30% drops between consecutive entries
// that both have HasContextTokens = true.
func countCompactions(tokens []ContextTokenRow) int {
	count := 0
	prevTokens := -1
	for _, t := range tokens {
		if !t.HasContextTokens {
			continue
		}
		if prevTokens > 0 {
			threshold := float64(prevTokens) * 0.7
			if float64(t.ContextTokens) < threshold {
				count++
			}
		}
		prevTokens = t.ContextTokens
	}
	return count
}

// computePressure returns ratio of peakContextTokens to model's
// context window, or nil when model is unknown or peak <= 0.
func computePressure(
	peakContextTokens int,
	model string,
) *float64 {
	if peakContextTokens <= 0 || model == "" {
		return nil
	}
	windowSize := lookupWindowSize(model)
	if windowSize == 0 {
		return nil
	}
	ratio := float64(peakContextTokens) / float64(windowSize)
	return &ratio
}

// lookupWindowSize finds context window size for a model. Tries
// exact match first, then prefix match (longest prefix wins).
func lookupWindowSize(model string) int {
	if size, ok := contextWindowSizes[model]; ok {
		return size
	}
	for _, prefix := range sortedPrefixes {
		if strings.HasPrefix(model, prefix) &&
			(len(model) == len(prefix) ||
				model[len(prefix)] == '-') {
			return contextWindowSizes[prefix]
		}
	}
	return 0
}
