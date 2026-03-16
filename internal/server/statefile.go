// ABOUTME: Manages server state files so CLI commands can detect
// ABOUTME: a running agentsview server instance.
package server

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StateFile records a running server instance.
type StateFile struct {
	PID       int    `json:"pid"`
	Port      int    `json:"port"`
	Host      string `json:"host"`
	Version   string `json:"version"`
	StartedAt string `json:"started_at"`
}

// stateFileName returns the filename for a given port.
func stateFileName(port int) string {
	return fmt.Sprintf("server.%d.json", port)
}

// WriteStateFile writes a state file to dataDir for the
// running server. Returns the path written. StartedAt is
// set to the actual process creation time so it passes
// processStartTime validation even when startup is slow.
func WriteStateFile(
	dataDir string, host string, port int, version string,
) (string, error) {
	started := time.Now()
	if ps, err := processStartTime(os.Getpid()); err == nil {
		started = ps
	}
	sf := StateFile{
		PID:       os.Getpid(),
		Port:      port,
		Host:      host,
		Version:   version,
		StartedAt: started.UTC().Format(time.RFC3339Nano),
	}
	data, err := json.Marshal(sf)
	if err != nil {
		return "", fmt.Errorf("marshaling state file: %w", err)
	}
	path := filepath.Join(dataDir, stateFileName(port))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("writing state file: %w", err)
	}
	return path, nil
}

// RemoveStateFile removes the state file for the given port.
func RemoveStateFile(dataDir string, port int) {
	os.Remove(filepath.Join(dataDir, stateFileName(port)))
}

// FindRunningServer scans dataDir for server state files and
// returns the first one whose process is still alive and whose
// port is accepting connections. Stale state files are cleaned
// up automatically.
func FindRunningServer(dataDir string) *StateFile {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil
	}

	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "server.") ||
			!strings.HasSuffix(name, ".json") {
			continue
		}

		path := filepath.Join(dataDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var sf StateFile
		if err := json.Unmarshal(data, &sf); err != nil {
			continue
		}

		// Check if the process is still running.
		if !processAlive(sf.PID) {
			os.Remove(path)
			continue
		}

		// Verify the port is actually listening. If the
		// process is alive but the dial fails (transient
		// timeout, GC pause, full backlog), keep the state
		// file — only a dead PID justifies removal.
		probeHost := probeHostForDial(sf.Host)
		conn, err := net.DialTimeout(
			"tcp",
			net.JoinHostPort(probeHost, fmt.Sprint(sf.Port)),
			500*time.Millisecond,
		)
		if err != nil {
			continue
		}
		conn.Close()

		return &sf
	}

	return nil
}

// stateFileStartTolerance is the maximum acceptable
// difference between the state file's StartedAt and the
// actual process start time. Accounts for the delay between
// process creation and WriteStateFile being called.
const stateFileStartTolerance = 120 * time.Second

// hasLiveStateFile reports whether any server state file in
// dataDir has a live PID, regardless of port connectivity.
// Unlike FindRunningServer, this returns true even during
// transient TCP probe failures.
//
// Staleness detection uses two layers:
//  1. Boot time: state files from before the last reboot are
//     removed (PID reuse across reboots).
//  2. Process start time: if available, the recorded
//     StartedAt is compared against the actual process
//     creation time. A mismatch beyond the tolerance
//     indicates same-boot PID reuse.
func hasLiveStateFile(dataDir string) bool {
	bootTime, _ := systemBootTime()

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "server.") ||
			!strings.HasSuffix(name, ".json") {
			continue
		}
		path := filepath.Join(dataDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var sf StateFile
		if err := json.Unmarshal(data, &sf); err != nil {
			continue
		}
		if !processAlive(sf.PID) {
			os.Remove(path)
			continue
		}

		started, parseErr := time.Parse(
			time.RFC3339Nano, sf.StartedAt,
		)

		// Layer 1: boot time check.
		if !bootTime.IsZero() && parseErr == nil {
			if started.Before(bootTime) {
				os.Remove(path)
				continue
			}
		}

		// Layer 2: process start time check.
		if parseErr == nil {
			if isStaleByProcessStart(
				sf.PID, started,
			) {
				os.Remove(path)
				continue
			}
		}

		return true
	}
	return false
}

// isStaleByProcessStart returns true if the process's actual
// creation time differs from the recorded StartedAt by more
// than stateFileStartTolerance, indicating PID reuse.
func isStaleByProcessStart(
	pid int, startedAt time.Time,
) bool {
	actual, err := processStartTime(pid)
	if err != nil {
		// Platform doesn't support start time — can't tell.
		return false
	}
	diff := actual.Sub(startedAt)
	if diff < 0 {
		diff = -diff
	}
	return diff > stateFileStartTolerance
}

const startupLockPrefix = "server.starting."

// startupLockFile returns the lock filename for a given PID.
func startupLockFile(pid int) string {
	return fmt.Sprintf("%s%d", startupLockPrefix, pid)
}

// WriteStartupLock creates a lock file indicating a server is
// starting up (syncing, binding port). Each server uses a
// PID-specific filename so concurrent startups on different
// ports don't clobber each other. Written via a temp file and
// atomic rename to prevent partial reads.
func WriteStartupLock(dataDir string) {
	name := startupLockFile(os.Getpid())
	target := filepath.Join(dataDir, name)
	tmp := target + ".tmp"
	data := fmt.Appendf(nil, "%d", os.Getpid())
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		_ = os.WriteFile(target, data, 0o644)
		return
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.WriteFile(target, data, 0o644)
		os.Remove(tmp)
	}
}

// RemoveStartupLock removes the startup lock file for the
// current process.
func RemoveStartupLock(dataDir string) {
	name := startupLockFile(os.Getpid())
	os.Remove(filepath.Join(dataDir, name))
}

// isServerStarting reports whether any server is currently
// starting up by scanning for lock files with live PIDs.
// Stale locks (dead PIDs) are cleaned up automatically.
func isServerStarting(dataDir string) bool {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, startupLockPrefix) {
			continue
		}
		path := filepath.Join(dataDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var pid int
		if _, err := fmt.Sscanf(
			string(data), "%d", &pid,
		); err != nil {
			continue
		}
		if !processAlive(pid) {
			os.Remove(path)
			continue
		}
		return true
	}
	return false
}

// IsStartupLocked reports whether the startup lock file exists
// with a live PID. Callers use this to distinguish "server is
// starting up" from "server is running but TCP probe failed".
func IsStartupLocked(dataDir string) bool {
	return isServerStarting(dataDir)
}

// IsServerActive reports whether a server process is managing
// the database in dataDir. Returns true if:
//   - a state file with a live PID exists (even if the port
//     probe fails due to a transient issue), or
//   - a startup lock with a live PID exists (server is still
//     syncing / binding its port).
//
// This is the check CLI commands should use to decide whether
// to skip on-demand sync.
func IsServerActive(dataDir string) bool {
	return hasLiveStateFile(dataDir) || isServerStarting(dataDir)
}

// WaitForStartup polls until the startup lock clears or a
// running server is detected, up to the given timeout.
// Returns true if a server became ready, false on timeout.
func WaitForStartup(dataDir string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if FindRunningServer(dataDir) != nil {
			return true
		}
		if !isServerStarting(dataDir) {
			// Lock gone but no running server — startup
			// may have failed. Caller should try on-demand
			// sync.
			return false
		}
		time.Sleep(250 * time.Millisecond)
	}
	return false
}

// probeHostForDial converts a bind-all address to a loopback
// address suitable for a TCP readiness probe, matching the
// normalization used by the server startup checks.
func probeHostForDial(host string) string {
	switch host {
	case "", "0.0.0.0":
		return "127.0.0.1"
	case "::":
		return "::1"
	default:
		return host
	}
}
