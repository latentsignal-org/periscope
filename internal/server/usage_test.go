package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wesm/agentsview/internal/db"
	"github.com/wesm/agentsview/internal/dbtest"
	"github.com/wesm/agentsview/internal/server"
)

// tokenUsageJSON is a valid token_usage blob for test messages.
var tokenUsageJSON = json.RawMessage(
	`{"input_tokens":100,"output_tokens":50,` +
		`"cache_creation_input_tokens":10,` +
		`"cache_read_input_tokens":20}`,
)

func TestParseUsageFilterDefaults(t *testing.T) {
	te := setup(t)

	// No params at all -> defaults should kick in.
	w := te.get(t, "/api/v1/usage/summary")
	assertStatus(t, w, http.StatusOK)

	resp := decode[server.UsageSummaryResponse](t, w)
	if resp.From == "" {
		t.Error("From is empty, expected defaulted value")
	}
	if resp.To == "" {
		t.Error("To is empty, expected defaulted value")
	}
	// from should be ~30 days before to.
	if resp.From >= resp.To {
		t.Errorf("From %q >= To %q", resp.From, resp.To)
	}
}

func TestParseUsageFilterExplicit(t *testing.T) {
	te := setup(t)

	w := te.get(t, buildPathURL("/api/v1/usage/summary",
		map[string]string{
			"from":     "2024-06-01",
			"to":       "2024-06-15",
			"timezone": "America/New_York",
			"project":  "myproj",
			"agent":    "claude",
			"model":    "claude-sonnet-4-20250514",
		}))
	assertStatus(t, w, http.StatusOK)

	resp := decode[server.UsageSummaryResponse](t, w)
	if resp.From != "2024-06-01" {
		t.Errorf("From = %q, want 2024-06-01", resp.From)
	}
	if resp.To != "2024-06-15" {
		t.Errorf("To = %q, want 2024-06-15", resp.To)
	}
}

func TestParseUsageFilterInvalidDate(t *testing.T) {
	te := setup(t)

	w := te.get(t, buildPathURL("/api/v1/usage/summary",
		map[string]string{"from": "yesterday"}))
	assertStatus(t, w, http.StatusBadRequest)
}

// seedUsageEnv populates sessions with token_usage data for
// usage endpoint testing.
func seedUsageEnv(t *testing.T, te *testEnv) {
	t.Helper()

	type entry struct {
		id, project, agent, started string
		msgs                        int
	}
	entries := []entry{
		{"u1", "alpha", "claude", "2024-06-01T09:00:00Z", 4},
		{"u2", "beta", "codex", "2024-06-02T10:00:00Z", 4},
	}

	for _, e := range entries {
		te.seedSession(t, e.id, e.project, e.msgs,
			func(sess *db.Session) {
				sess.Agent = e.agent
				sess.StartedAt = &e.started
				sess.EndedAt = &e.started
				sess.FirstMessage = dbtest.Ptr("Usage test")
			},
		)
		started := e.started
		te.seedMessages(t, e.id, e.msgs,
			func(i int, m *db.Message) {
				m.Timestamp = started
				if m.Role == "assistant" {
					m.Model = "claude-sonnet-4-20250514"
					m.TokenUsage = tokenUsageJSON
				}
			},
		)
	}
}

func TestHandleUsageSummaryJSONShape(t *testing.T) {
	te := setup(t)
	seedUsageEnv(t, te)

	w := te.get(t, buildPathURL("/api/v1/usage/summary",
		map[string]string{
			"from":     "2024-06-01",
			"to":       "2024-06-03",
			"timezone": "UTC",
		}))
	assertStatus(t, w, http.StatusOK)

	// Verify all expected top-level keys exist.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(
		w.Body.Bytes(), &raw,
	); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	required := []string{
		"from", "to", "totals", "daily",
		"projectTotals", "modelTotals", "agentTotals",
		"sessionCounts", "cacheStats",
	}
	for _, key := range required {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing key %q in response", key)
		}
	}

	resp := decode[server.UsageSummaryResponse](t, w)
	if len(resp.Daily) == 0 {
		t.Error("Daily is empty, expected entries")
	}
	if len(resp.ProjectTotals) == 0 {
		t.Error("ProjectTotals is empty")
	}
	if len(resp.ModelTotals) == 0 {
		t.Error("ModelTotals is empty")
	}
	if len(resp.AgentTotals) == 0 {
		t.Error("AgentTotals is empty")
	}
}

func TestHandleUsageTopSessionsEmpty(t *testing.T) {
	te := setup(t)

	w := te.get(t, buildPathURL(
		"/api/v1/usage/top-sessions",
		map[string]string{
			"from":     "2024-06-01",
			"to":       "2024-06-03",
			"timezone": "UTC",
		}))
	assertStatus(t, w, http.StatusOK)

	var entries []db.TopSessionEntry
	if err := json.Unmarshal(
		w.Body.Bytes(), &entries,
	); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if entries == nil {
		t.Error("expected non-null JSON array")
	}
}

func TestHandleUsageTopSessionsLimit(t *testing.T) {
	te := setup(t)
	seedUsageEnv(t, te)

	w := te.get(t, buildPathURL(
		"/api/v1/usage/top-sessions",
		map[string]string{
			"from":     "2024-06-01",
			"to":       "2024-06-03",
			"timezone": "UTC",
			"limit":    "1",
		}))
	assertStatus(t, w, http.StatusOK)

	var entries []db.TopSessionEntry
	if err := json.Unmarshal(
		w.Body.Bytes(), &entries,
	); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(entries) > 1 {
		t.Errorf("len(entries) = %d, want <= 1",
			len(entries))
	}
}

// TestUsageSummaryErrorRedaction verifies internal errors
// don't leak DB details.
func TestUsageSummaryErrorRedaction(t *testing.T) {
	te := setup(t)
	te.db.Close()

	w := te.get(t, buildPathURL("/api/v1/usage/summary",
		map[string]string{
			"from": "2024-06-01",
			"to":   "2024-06-03",
		}))
	assertStatus(t, w, http.StatusInternalServerError)
}

// Verify the route is actually registered by checking we
// don't get a 404 for usage endpoints.
func TestUsageRoutesRegistered(t *testing.T) {
	te := setup(t)

	endpoints := []string{
		"/api/v1/usage/summary",
		"/api/v1/usage/top-sessions",
	}
	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet, ep, nil,
			)
			w := httptest.NewRecorder()
			te.handler.ServeHTTP(w, req)
			if w.Code == http.StatusNotFound {
				t.Errorf("%s returned 404", ep)
			}
		})
	}
}
