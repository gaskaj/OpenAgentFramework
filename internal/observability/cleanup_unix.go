//go:build !windows

package observability

import (
	"fmt"
	"syscall"
)

func checkDiskSpace(dir string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return 0, fmt.Errorf("getting disk usage: %w", err)
	}
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}
