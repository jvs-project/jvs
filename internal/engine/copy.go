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

// CopyEngine performs a full recursive copy of directories.
type CopyEngine struct{}

// NewCopyEngine creates a new CopyEngine.
func NewCopyEngine() *CopyEngine {
	return &CopyEngine{}
}

// Name returns the engine type.
func (e *CopyEngine) Name() model.EngineType {
	return model.EngineCopy
}

// Clone recursively copies src to dst.
func (e *CopyEngine) Clone(src, dst string) (*CloneResult, error) {
	result := &CloneResult{}

	// Track hardlinks to detect degradation
	seenInodes := make(map[uint64]string)

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("relative path: %w", err)
		}
		dstPath := filepath.Join(dst, rel)

		// Check if this is a hardlink to a previously seen file
		if !info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			stat, ok := info.Sys().(*syscall.Stat_t)
			if ok {
				if seenInodes[stat.Ino] != "" {
					// This is a hardlink, copy engine cannot preserve it
					result.Degraded = true
					result.Degradations = append(result.Degradations, "hardlink")
				} else {
					seenInodes[stat.Ino] = path
				}
			}
		}

		switch {
		case info.IsDir():
			return e.copyDir(path, dstPath, info)

		case info.Mode()&os.ModeSymlink != 0:
			return e.copySymlink(path, dstPath, info)

		default:
			return e.copyFile(path, dstPath, info)
		}
	})

	if err != nil {
		return nil, fmt.Errorf("copy: %w", err)
	}

	// Fsync the destination directory
	if err := fsutil.FsyncDir(dst); err != nil {
		return nil, fmt.Errorf("fsync dst: %w", err)
	}

	return result, nil
}

func (e *CopyEngine) copyDir(src, dst string, info os.FileInfo) error {
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return fmt.Errorf("mkdir %s: %w", dst, err)
	}
	return nil
}

func (e *CopyEngine) copyFile(src, dst string, info os.FileInfo) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src %s: %w", src, err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("create dst %s: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy %s to %s: %w", src, dst, err)
	}

	// Sync file content
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("sync %s: %w", dst, err)
	}

	// Preserve mod time
	return os.Chtimes(dst, info.ModTime(), info.ModTime())
}

func (e *CopyEngine) copySymlink(src, dst string, info os.FileInfo) error {
	target, err := os.Readlink(src)
	if err != nil {
		return fmt.Errorf("readlink %s: %w", src, err)
	}
	return os.Symlink(target, dst)
}
