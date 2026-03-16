//go:build !darwin && !linux && !windows

package server

import (
	"fmt"
	"time"
)

// processStartTime is not implemented on this platform.
// Callers fall back to boot-time + PID checks only.
func processStartTime(_ int) (time.Time, error) {
	return time.Time{}, fmt.Errorf(
		"process start time not available",
	)
}
