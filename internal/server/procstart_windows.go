//go:build windows

package server

import (
	"fmt"
	"syscall"
	"time"
)

// processStartTime returns the wall-clock start time of the
// process with the given PID using GetProcessTimes on Windows.
func processStartTime(pid int) (time.Time, error) {
	h, err := syscall.OpenProcess(
		processQueryLimitedInformation,
		false,
		uint32(pid),
	)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"OpenProcess(%d): %w", pid, err,
		)
	}
	defer syscall.CloseHandle(h)

	var creation, exit, kernel, user syscall.Filetime
	err = syscall.GetProcessTimes(
		h, &creation, &exit, &kernel, &user,
	)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"GetProcessTimes(%d): %w", pid, err,
		)
	}

	return time.Unix(
		0, creation.Nanoseconds(),
	), nil
}
