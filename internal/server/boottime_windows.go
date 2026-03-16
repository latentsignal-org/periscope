//go:build windows

package server

import (
	"fmt"
	"time"
)

// systemBootTime is intentionally not implemented on Windows.
// GetTickCount64 (the usual source) excludes time spent in
// sleep/hibernation, so the calculated boot time drifts
// forward after each suspend cycle. This would cause
// hasLiveStateFile to incorrectly delete state files for
// long-running servers that survived a sleep.
//
// On Windows, processStartTime (via GetProcessTimes) provides
// a reliable absolute creation timestamp, so the boot-time
// layer is unnecessary.
func systemBootTime() (time.Time, error) {
	return time.Time{}, fmt.Errorf(
		"boot time not available on Windows " +
			"(GetTickCount64 is sleep-aware); " +
			"use processStartTime instead",
	)
}
