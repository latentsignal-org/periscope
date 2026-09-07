//go:build sshtest

package ssh

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.kenn.io/agentsview/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHSyncEndToEnd(t *testing.T) {
	host := testSSHHost(t)
	port := testSSHPort(t)
	user := testSSHUser(t)
	opts := testSSHOpts(t)
	database := testDB(t)

	rs := &RemoteSync{
		Host:    host,
		User:    user,
		Port:    port,
		Full:    true,
		DB:      database,
		SSHOpts: opts,
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), 30*time.Second,
	)
	defer cancel()

	stats, err := rs.Run(ctx)
	require.NoError(t, err, "remote sync")

	require.NotZero(t, stats.SessionsSynced, "expected at least 1 session synced")

	// Verify session landed in DB.
	page, err := database.ListSessions(
		context.Background(), db.SessionFilter{Limit: 100},
	)
	require.NoError(t, err, "listing sessions")
	require.NotEmpty(t, page.Sessions, "no sessions in database")

	// Session ID should carry the host prefix.
	found := false
	for _, s := range page.Sessions {
		if s.Machine == host {
			found = true
			assert.True(t, strings.HasPrefix(s.ID, host+"~"),
				"session ID %q missing host prefix", s.ID)
			break
		}
	}
	assert.True(t, found, "no session with machine=%q", host)
}

func TestSSHSyncIncremental(t *testing.T) {
	host := testSSHHost(t)
	port := testSSHPort(t)
	user := testSSHUser(t)
	opts := testSSHOpts(t)
	database := testDB(t)

	rs := &RemoteSync{
		Host:    host,
		User:    user,
		Port:    port,
		Full:    false,
		DB:      database,
		SSHOpts: opts,
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), 30*time.Second,
	)
	defer cancel()

	// First sync: should pull sessions.
	stats1, err := rs.Run(ctx)
	require.NoError(t, err, "first sync")
	require.NotZero(t, stats1.SessionsSynced, "first sync: expected sessions")

	// Second sync: nothing changed, should skip all.
	stats2, err := rs.Run(ctx)
	require.NoError(t, err, "second sync")
	assert.Equal(t, 0, stats2.SessionsSynced, "second sync: expected 0 synced")
	assert.NotZero(t, stats2.Skipped, "second sync: expected skipped > 0")
}

func TestSSHSyncFull(t *testing.T) {
	host := testSSHHost(t)
	port := testSSHPort(t)
	user := testSSHUser(t)
	opts := testSSHOpts(t)
	database := testDB(t)

	ctx, cancel := context.WithTimeout(
		context.Background(), 30*time.Second,
	)
	defer cancel()

	// Incremental first.
	rs := &RemoteSync{
		Host:    host,
		User:    user,
		Port:    port,
		Full:    false,
		DB:      database,
		SSHOpts: opts,
	}
	_, err := rs.Run(ctx)
	require.NoError(t, err, "first sync")

	// Full flag clears the remote skip cache but the engine
	// still skips unchanged sessions via DB lookup. Verify
	// it completes without error.
	rs.Full = true
	stats, err := rs.Run(ctx)
	require.NoError(t, err, "full sync")
	// Session was already synced and unchanged, so it may
	// be skipped by the engine's own DB-based detection.
	assert.NotZero(t, stats.SessionsSynced+stats.Skipped,
		"full sync: expected sessions processed")
}
