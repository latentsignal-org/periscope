//go:build linux

package server

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// atClkTck is the AT_CLKTCK auxiliary vector tag.
const atClkTck = 17

// clockTick caches the runtime CLK_TCK value.
var (
	clockTickOnce sync.Once
	clockTickVal  int64 = 100 // fallback
)

// getClockTick reads the kernel clock tick rate from
// /proc/self/auxv (AT_CLKTCK). Falls back to 100 if the
// auxiliary vector is unreadable.
func getClockTick() int64 {
	clockTickOnce.Do(func() {
		data, err := os.ReadFile("/proc/self/auxv")
		if err != nil {
			return
		}
		ptrSize := int(unsafe.Sizeof(uintptr(0)))
		entrySize := ptrSize * 2
		for i := 0; i+entrySize <= len(data); i += entrySize {
			var tag, val uint64
			if ptrSize == 8 {
				tag = binary.NativeEndian.Uint64(
					data[i : i+8],
				)
				val = binary.NativeEndian.Uint64(
					data[i+8 : i+16],
				)
			} else {
				tag = uint64(
					binary.NativeEndian.Uint32(
						data[i : i+4],
					),
				)
				val = uint64(
					binary.NativeEndian.Uint32(
						data[i+4 : i+8],
					),
				)
			}
			if tag == atClkTck && val > 0 {
				clockTickVal = int64(val)
				return
			}
			if tag == 0 { // AT_NULL
				return
			}
		}
	})
	return clockTickVal
}

// processStartTime returns the wall-clock start time of the
// process with the given PID by reading /proc/<pid>/stat.
// Field 22 (1-indexed) is starttime in clock ticks since boot.
func processStartTime(pid int) (time.Time, error) {
	data, err := os.ReadFile(
		fmt.Sprintf("/proc/%d/stat", pid),
	)
	if err != nil {
		return time.Time{}, err
	}

	// The comm field (field 2) is in parentheses and may
	// contain spaces or parens, so find the last ')' to
	// skip it reliably.
	s := string(data)
	idx := strings.LastIndex(s, ") ")
	if idx < 0 {
		return time.Time{}, fmt.Errorf(
			"malformed /proc/%d/stat", pid,
		)
	}
	// Fields after comm start at field 3 (state).
	// starttime is field 22 = index 19 after comm.
	fields := strings.Fields(s[idx+2:])
	if len(fields) < 20 {
		return time.Time{}, fmt.Errorf(
			"too few fields in /proc/%d/stat", pid,
		)
	}

	var startTicks int64
	if _, err := fmt.Sscanf(
		fields[19], "%d", &startTicks,
	); err != nil {
		return time.Time{}, fmt.Errorf(
			"parsing starttime: %w", err,
		)
	}

	bootTime, err := systemBootTime()
	if err != nil {
		return time.Time{}, err
	}
	hz := getClockTick()
	startSec := startTicks / hz
	startNsec := (startTicks % hz) *
		(int64(time.Second) / hz)
	return bootTime.Add(
		time.Duration(startSec)*time.Second +
			time.Duration(startNsec),
	), nil
}
