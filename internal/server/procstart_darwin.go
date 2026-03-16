//go:build darwin

package server

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

// processStartTime returns the wall-clock start time of the
// process with the given PID using sysctl KERN_PROC on macOS.
func processStartTime(pid int) (time.Time, error) {
	mib := [4]int32{
		1,  // CTL_KERN
		14, // KERN_PROC
		1,  // KERN_PROC_PID
		int32(pid),
	}

	// First call to get the required buffer size.
	var size uintptr
	_, _, errno := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		4,
		0,
		uintptr(unsafe.Pointer(&size)),
		0,
		0,
	)
	if errno != 0 {
		return time.Time{}, fmt.Errorf(
			"sysctl KERN_PROC size: %w", errno,
		)
	}
	if size == 0 {
		return time.Time{}, fmt.Errorf(
			"process %d not found", pid,
		)
	}

	buf := make([]byte, size)
	_, _, errno = syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		4,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
		0,
		0,
	)
	if errno != 0 {
		return time.Time{}, fmt.Errorf(
			"sysctl KERN_PROC: %w", errno,
		)
	}

	// kinfo_proc starts with extern_proc whose first
	// field (p_un.__p_starttime) is a timeval at offset 0.
	const startTimeOff = 0
	if int(size) < startTimeOff+16 {
		return time.Time{}, fmt.Errorf(
			"kinfo_proc too small: %d bytes", size,
		)
	}

	sec := int64(binary.LittleEndian.Uint64(
		buf[startTimeOff : startTimeOff+8],
	))
	usec := int64(binary.LittleEndian.Uint32(
		buf[startTimeOff+8 : startTimeOff+12],
	))
	if sec == 0 {
		return time.Time{}, fmt.Errorf(
			"process %d: zero start time", pid,
		)
	}
	return time.Unix(sec, usec*1000), nil
}
