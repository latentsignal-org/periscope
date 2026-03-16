//go:build linux

package server

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// systemBootTime returns the system boot time on Linux by
// reading the btime field from /proc/stat.
func systemBootTime() (time.Time, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return time.Time{}, err
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if strings.HasPrefix(line, "btime ") {
			var btime int64
			if _, err := fmt.Sscanf(
				line, "btime %d", &btime,
			); err != nil {
				return time.Time{}, err
			}
			return time.Unix(btime, 0), nil
		}
	}
	return time.Time{}, fmt.Errorf(
		"btime not found in /proc/stat",
	)
}
