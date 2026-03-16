//go:build windows

package server

import "syscall"

const processQueryLimitedInformation = 0x1000

// processAliveAccess combines SYNCHRONIZE (required by
// WaitForSingleObject) with limited query rights.
const processAliveAccess = syscall.SYNCHRONIZE |
	processQueryLimitedInformation

// processAlive reports whether a process with the given PID
// exists. On Windows, signal-0 is not supported, so we open
// a process handle and use WaitForSingleObject with a zero
// timeout. WAIT_TIMEOUT means the process is still running.
// This avoids the GetExitCodeProcess pitfall where a process
// that exits with code 259 (STILL_ACTIVE) is falsely
// detected as alive.
func processAlive(pid int) bool {
	h, err := syscall.OpenProcess(
		processAliveAccess, false, uint32(pid),
	)
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(h)

	event, _ := syscall.WaitForSingleObject(h, 0)
	return event == syscall.WAIT_TIMEOUT
}
