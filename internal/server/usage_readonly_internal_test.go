package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wesm/agentsview/internal/db"
)

// readOnlyUsageSpy stubs the Store interface and returns
// db.ErrReadOnly from the three usage queries. It lets us
// verify the handlers map the sentinel error to 501 Not
// Implemented without spinning up a real PG instance.
type readOnlyUsageSpy struct {
	db.Store
}

func (readOnlyUsageSpy) GetDailyUsage(
	_ context.Context, _ db.UsageFilter,
) (db.DailyUsageResult, error) {
	return db.DailyUsageResult{}, db.ErrReadOnly
}

func (readOnlyUsageSpy) GetTopSessionsByCost(
	_ context.Context, _ db.UsageFilter, _ int,
) ([]db.TopSessionEntry, error) {
	return nil, db.ErrReadOnly
}

func (readOnlyUsageSpy) GetUsageSessionCounts(
	_ context.Context, _ db.UsageFilter,
) (db.UsageSessionCounts, error) {
	return db.UsageSessionCounts{}, db.ErrReadOnly
}

// TestUsageHandlers_ReturnNotImplementedOnReadOnlyStore locks
// in the Postgres-backend contract: when the underlying Store
// reports a usage query as unavailable (db.ErrReadOnly), both
// usage HTTP endpoints must surface 501 Not Implemented rather
// than silently returning an empty body, which would look like
// "no usage data" to the user.
func TestUsageHandlers_ReturnNotImplementedOnReadOnlyStore(
	t *testing.T,
) {
	s := &Server{db: readOnlyUsageSpy{}}

	cases := []struct {
		name    string
		path    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{
			name: "summary",
			path: "/api/v1/usage/summary?" +
				"from=2024-06-01&to=2024-06-03",
			handler: s.handleUsageSummary,
		},
		{
			name: "top-sessions",
			path: "/api/v1/usage/top-sessions?" +
				"from=2024-06-01&to=2024-06-03",
			handler: s.handleUsageTopSessions,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet, tc.path, nil,
			)
			w := httptest.NewRecorder()
			tc.handler(w, req)
			if w.Code != http.StatusNotImplemented {
				t.Errorf(
					"status = %d, want 501; body=%s",
					w.Code, w.Body.String(),
				)
			}
		})
	}
}
