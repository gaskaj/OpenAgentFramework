//go:build windows

package workspace

import (
	"fmt"
	"syscall"
	"unsafe"
)

func getDiskStats(path string) (*DiskStats, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")

	dirPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("converting path: %w", err)
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes int64
	ret, _, callErr := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(dirPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetDiskFreeSpaceExW for %s: %w", path, callErr)
	}

	totalMB := totalBytes / (1024 * 1024)
	availableMB := freeBytesAvailable / (1024 * 1024)
	usedMB := totalMB - (totalFreeBytes / (1024 * 1024))

	usagePercent := 0.0
	if totalMB > 0 {
		usagePercent = float64(usedMB) / float64(totalMB) * 100.0
	}

	return &DiskStats{
		TotalMB:      totalMB,
		UsedMB:       usedMB,
		AvailableMB:  availableMB,
		UsagePercent: usagePercent,
	}, nil
}
