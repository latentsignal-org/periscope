package service

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errReader fails every read with err, simulating a broken connection.
type errReader struct{ err error }

func (e errReader) Read(_ []byte) (int, error) { return 0, e.err }

func TestParseSSE_SingleEvent(t *testing.T) {
	t.Parallel()
	raw := "event: hello\ndata: world\n\n"
	var got []Event
	parseSSE(strings.NewReader(raw), func(ev Event) bool {
		got = append(got, ev)
		return true
	})
	require.Len(t, got, 1)
	assert.Equal(t, "hello", got[0].Event)
	assert.Equal(t, "world", got[0].Data)
}

func TestParseSSE_MultipleEvents(t *testing.T) {
	t.Parallel()
	raw := "event: a\ndata: 1\n\nevent: b\ndata: 2\n\n"
	var got []Event
	parseSSE(strings.NewReader(raw), func(ev Event) bool {
		got = append(got, ev)
		return true
	})
	require.Len(t, got, 2)
	assert.Equal(t, "a", got[0].Event)
	assert.Equal(t, "1", got[0].Data)
	assert.Equal(t, "b", got[1].Event)
	assert.Equal(t, "2", got[1].Data)
}

func TestParseSSE_EmitStopsParsing(t *testing.T) {
	t.Parallel()
	raw := "event: a\ndata: 1\n\nevent: b\ndata: 2\n\n"
	var got []Event
	parseSSE(strings.NewReader(raw), func(ev Event) bool {
		got = append(got, ev)
		return false // stop after first event
	})
	require.Len(t, got, 1)
	assert.Equal(t, "a", got[0].Event)
}

func TestParseSSE_IgnoresIncompleteTrailingEvent(t *testing.T) {
	t.Parallel()
	// No blank line after the final "data:" line — the event must
	// not be emitted because the delimiter hasn't been seen.
	raw := "event: a\ndata: 1\n\nevent: b\ndata: 2\n"
	var got []Event
	parseSSE(strings.NewReader(raw), func(ev Event) bool {
		got = append(got, ev)
		return true
	})
	require.Len(t, got, 1)
	assert.Equal(t, "a", got[0].Event)
}

func TestParseSSE_SkipsUnknownFields(t *testing.T) {
	t.Parallel()
	// Lines that are neither "event: " nor "data: " must be ignored
	// without breaking parsing.
	raw := "id: 42\nevent: hello\nretry: 1000\ndata: world\n\n"
	var got []Event
	parseSSE(strings.NewReader(raw), func(ev Event) bool {
		got = append(got, ev)
		return true
	})
	require.Len(t, got, 1)
	assert.Equal(t, "hello", got[0].Event)
	assert.Equal(t, "world", got[0].Data)
}

func TestParseSSE_SurfacesReadError(t *testing.T) {
	t.Parallel()
	boom := errors.New("connection reset")
	err := parseSSE(errReader{err: boom}, func(Event) bool { return true })
	require.ErrorIs(t, err, boom)
}

const summaryEvent = "event: summary\n" +
	"data: {\"scanned\":2,\"with_secrets\":1,\"total_findings\":3," +
	"\"definite_findings\":2,\"candidate_findings\":1}\n\n"

func TestParseScanStream_SummaryReturned(t *testing.T) {
	t.Parallel()
	sum, err := parseScanStream(strings.NewReader(summaryEvent), nil)
	require.NoError(t, err)
	require.NotNil(t, sum)
	assert.Equal(t, 2, sum.Scanned)
	assert.Equal(t, 1, sum.WithSecrets)
	assert.Equal(t, 3, sum.TotalFindings)
	assert.Equal(t, 2, sum.DefiniteFindings)
	assert.Equal(t, 1, sum.CandidateFindings)
}

func TestParseScanStream_NoSummaryIsError(t *testing.T) {
	t.Parallel()
	// Stream ends cleanly after a progress tick but before any summary.
	raw := "event: progress\ndata: {\"scanned\":1,\"total\":2}\n\n"
	_, err := parseScanStream(strings.NewReader(raw), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stream ended before summary")
}

func TestParseScanStream_ReadErrorBeforeSummary(t *testing.T) {
	t.Parallel()
	_, err := parseScanStream(errReader{err: errors.New("boom")}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading stream")
}

func TestParseScanStream_ErrorEventWins(t *testing.T) {
	t.Parallel()
	raw := "event: error\ndata: scan aborted\n\n"
	_, err := parseScanStream(strings.NewReader(raw), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scan aborted")
}

func TestParseScanStream_SummaryDespiteTrailingReadError(t *testing.T) {
	t.Parallel()
	// A complete summary arrives, then the connection drops. The result
	// is valid, so the trailing read error must not mask it.
	r := io.MultiReader(strings.NewReader(summaryEvent),
		errReader{err: errors.New("late boom")})
	sum, err := parseScanStream(r, nil)
	require.NoError(t, err)
	require.NotNil(t, sum)
	assert.Equal(t, 2, sum.Scanned)
}

func TestParseScanStream_MalformedSummaryIsError(t *testing.T) {
	t.Parallel()
	raw := "event: summary\ndata: {not valid json}\n\n"
	_, err := parseScanStream(strings.NewReader(raw), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decoding summary")
}
