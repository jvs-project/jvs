package engine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/jvs-project/jvs/pkg/fsutil"
	"github.com/jvs-project/jvs/pkg/model"
)

// ReflinkEngine performs reflink-based copy (O(1) CoW) on supported filesystems.
// Falls back to regular copy for files that cannot be reflinked.
type ReflinkEngine struct {
	CopyEngine *CopyEngine // Fallback for unsupported cases
}

// NewReflinkEngine creates a new ReflinkEngine.
func NewReflinkEngine() *ReflinkEngine {
	return &ReflinkEngine{
		CopyEngine: NewCopyEngine(),
	}
}

// Name returns the engine type.
func (e *ReflinkEngine) Name() model.EngineType {
	return model.EngineReflinkCopy
}

// Clone performs a reflink copy if supported, falls back to regular copy otherwise.
// Returns a degraded result if any files could not be reflinked.
func (e *ReflinkEngine) Clone(src, dst string) (*CloneResult, error) {
	result := &CloneResult{}

	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return nil, fmt.Errorf("create dst directory: %w", err)
	}

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("relative path: %w", err)
		}
		dstPath := filepath.Join(dst, rel)

		switch {
		case info.IsDir():
			return e.copyDir(path, dstPath, info)

		case info.Mode()&os.ModeSymlink != 0:
			return e.copySymlink(path, dstPath, info)

		default:
			// Try reflink first
			if err := e.reflinkFile(path, dstPath, info); err != nil {
				// Reflink failed, fall back to copy
				result.Degraded = true
				result.Degradations = append(result.Degradations, "reflink")
				return e.copyFile(path, dstPath, info)
			}
			return nil
		}
	})

	if err != nil {
		return nil, fmt.Errorf("reflink clone: %w", err)
	}

	// Fsync the destination directory
	if err := fsutil.FsyncDir(dst); err != nil {
		return nil, fmt.Errorf("fsync dst: %w", err)
	}

	return result, nil
}

// reflinkFile attempts to create a reflink copy of a file.
func (e *ReflinkEngine) reflinkFile(src, dst string, info os.FileInfo) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("create dst: %w", err)
	}
	defer dstFile.Close()

	// Try FICLONE ioctl (Linux)
	// FICLONE: ioctl(dest_fd, FICLONE, src_fd)
	const FICLONE = 0x40049409
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, dstFile.Fd(), FICLONE, srcFile.Fd())
	if errno != 0 {
		dstFile.Close()
		os.Remove(dst)
		return fmt.Errorf("ficlone failed: %v", errno)
	}

	// Preserve mod time
	return os.Chtimes(dst, info.ModTime(), info.ModTime())
}

func (e *ReflinkEngine) copyDir(src, dst string, info os.FileInfo) error {
	return os.MkdirAll(dst, info.Mode())
}

func (e *ReflinkEngine) copySymlink(src, dst string, info os.FileInfo) error {
	target, err := os.Readlink(src)
	if err != nil {
		return fmt.Errorf("readlink: %w", err)
	}
	return os.Symlink(target, dst)
}

func (e *ReflinkEngine) copyFile(src, dst string, info os.FileInfo) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("create dst: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return os.Chtimes(dst, info.ModTime(), info.ModTime())
}
