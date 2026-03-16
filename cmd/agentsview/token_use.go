// ABOUTME: CLI subcommand that returns token usage data for a
// ABOUTME: session, syncing on-demand if no server is running.
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/wesm/agentsview/internal/config"
	"github.com/wesm/agentsview/internal/db"
	"github.com/wesm/agentsview/internal/parser"
	"github.com/wesm/agentsview/internal/server"
	"github.com/wesm/agentsview/internal/sync"
)

// tokenUseOutput is the JSON structure written to stdout.
// This format is experimental and may change.
type tokenUseOutput struct {
	SessionID         string `json:"session_id"`
	Agent             string `json:"agent"`
	Project           string `json:"project"`
	TotalOutputTokens int    `json:"total_output_tokens"`
	PeakContextTokens int    `json:"peak_context_tokens"`
	ServerRunning     bool   `json:"server_running"`
}

// startupWaitTimeout is how long token-use will wait for a
// starting server to become ready before falling back to
// on-demand sync.
const startupWaitTimeout = 30 * time.Second

func runTokenUse(args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr,
			"usage: agentsview token-use <session-id>")
		os.Exit(1)
	}

	if err := tokenUse(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func tokenUse(sessionID string) error {
	appCfg, err := config.LoadMinimal()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := os.MkdirAll(appCfg.DataDir, 0o755); err != nil {
		return fmt.Errorf("creating data dir: %w", err)
	}

	serverActive := server.IsServerActive(appCfg.DataDir)

	// If a server is actively starting up (startup lock
	// present), wait for it to finish so we read fresh data
	// rather than returning stale results or "not found".
	// We only wait when the startup lock is the reason
	// IsServerActive returned true — if a state file has a
	// live PID but the TCP probe is transiently failing,
	// the server is running and we should just read the DB.
	if serverActive &&
		server.FindRunningServer(appCfg.DataDir) == nil {
		if server.IsStartupLocked(appCfg.DataDir) {
			fmt.Fprintf(os.Stderr,
				"server is starting up, waiting...\n")
			if !server.WaitForStartup(
				appCfg.DataDir, startupWaitTimeout,
			) {
				if server.IsStartupLocked(appCfg.DataDir) {
					// Lock still live after timeout:
					// the server is active (still
					// syncing, or state file write
					// failed). Don't compete — read
					// the DB as-is.
					fmt.Fprintf(os.Stderr,
						"server still starting after "+
							"%s, reading DB as-is\n",
						startupWaitTimeout,
					)
				} else {
					// Lock cleared but no running
					// server. Re-check in case of
					// transient TCP failure.
					serverActive = server.IsServerActive(
						appCfg.DataDir,
					)
				}
			}
		} else if !server.IsServerActive(appCfg.DataDir) {
			// The server that was alive at the first check
			// has since exited. Fall back to on-demand sync.
			serverActive = false
		}
	}

	database, err := db.Open(appCfg.DBPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer database.Close()

	if appCfg.CursorSecret != "" {
		secret, decErr := base64.StdEncoding.DecodeString(
			appCfg.CursorSecret,
		)
		if decErr != nil {
			return fmt.Errorf(
				"invalid cursor secret: %w", decErr,
			)
		}
		database.SetCursorSecret(secret)
	}

	// If no server is managing the DB, do an on-demand sync
	// for this session so the data is fresh. Re-check right
	// before syncing to close the TOCTOU window where a
	// server could have started since our initial probe.
	// If the re-check detects a starting server, wait for
	// it rather than reading potentially stale data.
	if !serverActive {
		serverActive = server.IsServerActive(appCfg.DataDir)
		if serverActive &&
			server.FindRunningServer(appCfg.DataDir) == nil &&
			server.IsStartupLocked(appCfg.DataDir) {
			fmt.Fprintf(os.Stderr,
				"server is starting up, waiting...\n")
			if server.WaitForStartup(
				appCfg.DataDir, startupWaitTimeout,
			) {
				// Server is ready; read DB below.
			} else if !server.IsStartupLocked(
				appCfg.DataDir,
			) {
				// Lock cleared, no running server
				// via TCP. Re-check: a live state
				// file (transient probe failure)
				// still means the server is active.
				serverActive = server.IsServerActive(
					appCfg.DataDir,
				)
			}
			// Lock still live after timeout: server is
			// active but slow. Read DB as-is.
		}
	}
	if !serverActive {
		engine := sync.NewEngine(database, sync.EngineConfig{
			AgentDirs:               appCfg.AgentDirs,
			Machine:                 "local",
			BlockedResultCategories: appCfg.ResultContentBlockedCategories,
		})
		if syncErr := engine.SyncSingleSession(
			sessionID,
		); syncErr != nil {
			// Not fatal: session may already be in the DB
			// from a previous sync, or may not exist at all.
			fmt.Fprintf(os.Stderr,
				"warning: sync failed: %v\n", syncErr)
		}
	}

	sess, err := database.GetSession(
		context.Background(), sessionID,
	)
	if err != nil {
		return fmt.Errorf("querying session: %w", err)
	}
	if sess == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	agent := sess.Agent
	if agent == "" {
		if def, ok := parser.AgentByPrefix(sessionID); ok {
			agent = string(def.Type)
		}
	}

	out := tokenUseOutput{
		SessionID:         sess.ID,
		Agent:             agent,
		Project:           sess.Project,
		TotalOutputTokens: sess.TotalOutputTokens,
		PeakContextTokens: sess.PeakContextTokens,
		ServerRunning:     serverActive,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
