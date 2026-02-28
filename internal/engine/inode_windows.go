//go:build windows

package engine

import "os"

// fileInode is a no-op on Windows; hardlink detection is not supported.
func fileInode(_ os.FileInfo) (uint64, bool) {
	return 0, false
}
