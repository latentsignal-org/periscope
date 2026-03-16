//go:build !darwin && !linux && !windows

package server

import (
	"fmt"
	"time"
)

// systemBootTime is not implemented on this platform.
// Returns an error so callers fall back to PID-only checks.
func systemBootTime() (time.Time, error) {
	return time.Time{}, fmt.Errorf(
		"boot time not available on this platform",
	)
}
