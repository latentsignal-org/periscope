package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.kenn.io/agentsview/internal/config"
	"go.kenn.io/agentsview/internal/db"
	"go.kenn.io/agentsview/internal/dbtest"
	"go.kenn.io/agentsview/internal/llm"
	"go.kenn.io/agentsview/internal/summarize"
)

const contextHTTPTestSessionID = "ctx-http-sess"

func seedContextHTTPSession(t *testing.T, d *db.DB) {
	t.Helper()
	require.NoError(t, d.UpsertSession(db.Session{
		ID:                          contextHTTPTestSessionID,
		Project:                     "proj",
		Agent:                       "claude",
		HasModelContextWindowTokens: true,
		ModelContextWindowTokens:    200000,
	}))
	require.NoError(t, d.InsertMessages([]db.Message{
		{
			SessionID:        contextHTTPTestSessionID,
			Ordinal:          1,
			Role:             "user",
			Content:          "Investigate the failing context tests.",
			ContentLength:    36,
			ContextTokens:    100,
			HasContextTokens: true,
		},
		{
			SessionID:        contextHTTPTestSessionID,
			Ordinal:          2,
			Role:             "assistant",
			Content:          "I will inspect the handler wiring first.",
			ContentLength:    38,
			ContextTokens:    240,
			HasContextTokens: true,
			Model:            "claude-sonnet-4-5",
		},
	}))
}

func newContextRoutesHTTPServer(
	t *testing.T, cfg config.Config, opts ...Option,
) (*Server, http.Handler) {
	t.Helper()
	database := dbtest.OpenTestDB(t)
	seedContextHTTPSession(t, database)
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 30 * time.Second
	}
	srv := New(cfg, database, nil, opts...)
	return srv, srv.Handler()
}

type contextHTTPRequestOption func(*http.Request)

func withContextBearer(token string) contextHTTPRequestOption {
	return func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer "+token)
	}
}

func serveContextHTTP(
	t *testing.T,
	h http.Handler,
	method, path string,
	opts ...contextHTTPRequestOption,
) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	req.RemoteAddr = "127.0.0.1:1234"
	for _, opt := range opts {
		opt(req)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

type stubSummarizeClient struct{}

func (stubSummarizeClient) Complete(
	_ context.Context, _ llm.Request,
) (llm.Response, error) {
	return llm.Response{
		Text:  `{"summary":"ok","intent":"implement","outcome":"progress","topic":"test"}`,
		Model: "claude-haiku-4-5",
	}, nil
}

func TestGetSessionContext_HTTPReturnsJSONContract(t *testing.T) {
	database := dbtest.OpenTestDB(t)
	seedContextHTTPSession(t, database)
	s := newRoutedTestServerWithStore(t, database)
	w := serveGet(
		t, s,
		"/api/v1/sessions/"+contextHTTPTestSessionID+"/context",
	)
	assertRecorderStatus(t, w, http.StatusOK)
	assertContentType(t, w, "application/json")

	var body sessionContextResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 240, body.Summary.TokensInUse)
	assert.Equal(t, 200000, body.Capacity.MaxTokens)
	assert.NotEmpty(t, body.Composition)
	assert.True(t, body.Supports.StandaloneRoute)
	require.NotNil(t, body.SummaryCoverage)
	assert.Equal(t, "disabled", body.SummaryCoverage.Status)
	assert.Equal(t, 1, body.SummaryCoverage.TotalTurns)
}

func TestGetSessionContext_HTTPNotFound(t *testing.T) {
	s := newRoutedTestServerWithStore(t, dbtest.OpenTestDB(t))
	w := serveGet(t, s, "/api/v1/sessions/missing-session/context")
	assertRecorderStatus(t, w, http.StatusNotFound)
	assertContentType(t, w, "application/json")

	var errBody map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errBody))
	assert.Equal(t, "session not found", errBody["error"])
}

func TestGetSessionContextTimeline_HTTPReturnsTimeline(t *testing.T) {
	database := dbtest.OpenTestDB(t)
	seedContextHTTPSession(t, database)
	s := newRoutedTestServerWithStore(t, database)
	w := serveGet(
		t, s,
		"/api/v1/sessions/"+contextHTTPTestSessionID+"/context/timeline",
	)
	assertRecorderStatus(t, w, http.StatusOK)
	assertContentType(t, w, "application/json")

	var body sessionContextTimelineResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Timeline, 1)
	assert.Equal(t, 1, body.Timeline[0].Turn)
	assert.Equal(t, 1, body.Timeline[0].StartOrdinal)
	assert.Equal(t, 2, body.Timeline[0].EndOrdinal)
	assert.True(t, body.Supports.EmbeddedTab)
}

func TestPostSessionSummarize_HTTPDisabled(t *testing.T) {
	database := dbtest.OpenTestDB(t)
	seedContextHTTPSession(t, database)
	s := newRoutedTestServerWithStore(t, database)
	w := serveContextHTTP(
		t, s.mux,
		http.MethodPost,
		"/api/v1/sessions/"+contextHTTPTestSessionID+"/summarize",
	)
	assertRecorderStatus(t, w, http.StatusServiceUnavailable)
	assertContentType(t, w, "application/json")

	var errBody map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errBody))
	assert.Contains(
		t, errBody["error"],
		"summaries disabled: set ANTHROPIC_API_KEY and restart",
	)
}

func TestPostSessionSummarize_HTTPEnabled(t *testing.T) {
	database := dbtest.OpenTestDB(t)
	seedContextHTTPSession(t, database)
	worker := summarize.NewWorker(
		database, stubSummarizeClient{}, summarize.WorkerOptions{},
	)
	s := newRoutedTestServerWithStore(t, database)
	s.summarizer = worker

	w := serveContextHTTP(
		t, s.mux,
		http.MethodPost,
		"/api/v1/sessions/"+contextHTTPTestSessionID+"/summarize",
	)
	assertRecorderStatus(t, w, http.StatusNoContent)
	assert.Empty(t, w.Body.String())
}

func TestContextRoutes_HTTPRequireAuthWhenEnabled(t *testing.T) {
	_, handler := newContextRoutesHTTPServer(t, config.Config{
		RequireAuth: true,
		AuthToken:   "ctx-secret",
	})
	base := "/api/v1/sessions/" + contextHTTPTestSessionID

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"GetContext", http.MethodGet, base + "/context"},
		{"GetTimeline", http.MethodGet, base + "/context/timeline"},
		{"PostSummarize", http.MethodPost, base + "/summarize"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := serveContextHTTP(t, handler, tt.method, tt.path)
			assert.Equal(t, http.StatusUnauthorized, w.Code)

			w = serveContextHTTP(
				t, handler, tt.method, tt.path,
				withContextBearer("ctx-secret"),
			)
			if tt.method == http.MethodPost {
				assert.Equal(t, http.StatusServiceUnavailable, w.Code)
				return
			}
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestContextRoutes_HTTPTimeoutWrapped(t *testing.T) {
	t.Parallel()

	srv := testServer(
		t, 10*time.Millisecond,
		withHandlerDelay(100*time.Millisecond),
	)
	database, ok := srv.db.(*db.DB)
	require.True(t, ok)
	seedContextHTTPSession(t, database)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	srv.SetPort(port)
	ts := httptest.NewUnstartedServer(srv.Handler())
	ts.Listener = ln
	ts.Start()
	t.Cleanup(ts.Close)

	tests := []struct {
		name   string
		method string
		path   string
		detail string
	}{
		{
			name:   "GetContext",
			method: http.MethodGet,
			path:   "/api/v1/sessions/" + contextHTTPTestSessionID + "/context",
			detail: "GET /api/v1/sessions/{id}/context",
		},
		{
			name:   "GetTimeline",
			method: http.MethodGet,
			path:   "/api/v1/sessions/" + contextHTTPTestSessionID + "/context/timeline",
			detail: "GET /api/v1/sessions/{id}/context/timeline",
		},
		{
			name:   "PostSummarize",
			method: http.MethodPost,
			path:   "/api/v1/sessions/" + contextHTTPTestSessionID + "/summarize",
			detail: "POST /api/v1/sessions/{id}/summarize",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, ts.URL+tt.path, nil)
			require.NoError(t, err)
			req.Header.Set("Origin", ts.URL)
			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assertTimeoutResponse(
				t, resp, tt.detail, "10ms", "--write-timeout",
			)
		})
	}
}
