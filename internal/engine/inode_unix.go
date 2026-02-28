//go:build !windows

package engine

import (
	"os"
	"syscall"
)

// fileInode extracts the inode number from file info on Unix systems.
func fileInode(info os.FileInfo) (uint64, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return stat.Ino, true
}
