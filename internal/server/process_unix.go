//go:build !windows

package server

import (
	"os"
	"syscall"
)

// processAlive reports whether a process with the given PID
// exists. On Unix, signal 0 probes existence without
// delivering a signal.
func processAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
