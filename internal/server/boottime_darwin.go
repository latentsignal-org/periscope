//go:build darwin

package server

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

// systemBootTime returns the system boot time on macOS using
// sysctl kern.boottime.
func systemBootTime() (time.Time, error) {
	mib := [2]int32{1 /* CTL_KERN */, 21 /* KERN_BOOTTIME */}
	var tv syscall.Timeval
	size := uintptr(unsafe.Sizeof(tv))

	_, _, errno := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		2,
		uintptr(unsafe.Pointer(&tv)),
		uintptr(unsafe.Pointer(&size)),
		0,
		0,
	)
	if errno != 0 {
		return time.Time{}, fmt.Errorf(
			"sysctl kern.boottime: %w", errno,
		)
	}
	return time.Unix(tv.Sec, int64(tv.Usec)*1000), nil
}
