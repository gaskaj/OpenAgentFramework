//go:build !windows

package workspace

import (
	"fmt"
	"syscall"
)

func getDiskStats(path string) (*DiskStats, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, fmt.Errorf("getting disk stats for %s: %w", path, err)
	}

	blockSize := int64(stat.Bsize)
	totalBlocks := int64(stat.Blocks)
	freeBlocks := int64(stat.Bavail)
	usedBlocks := totalBlocks - int64(stat.Bfree)

	totalMB := (totalBlocks * blockSize) / (1024 * 1024)
	usedMB := (usedBlocks * blockSize) / (1024 * 1024)
	availableMB := (freeBlocks * blockSize) / (1024 * 1024)

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
