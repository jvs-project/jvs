// Package fsutil provides filesystem utilities for atomic operations and syncing.
package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// AtomicWrite writes data to a temporary file, fsyncs, then renames to target path.
func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".jvs-tmp-*")
	if err != nil {
		return fmt.Errorf("atomic write create tmp: %w", err)
	}
	tmpPath := tmp.Name()

	// Clean up on failure
	success := false
	defer func() {
		if !success {
			tmp.Close()
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("atomic write: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		return fmt.Errorf("atomic write chmod: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("atomic write fsync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("atomic write close: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("atomic write rename: %w", err)
	}
	if err := FsyncDir(dir); err != nil {
		return fmt.Errorf("atomic write fsync dir: %w", err)
	}

	success = true
	return nil
}

// RenameAndSync renames old to new and fsyncs the parent directory.
func RenameAndSync(oldpath, newpath string) error {
	if err := os.Rename(oldpath, newpath); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return FsyncDir(filepath.Dir(newpath))
}

// FsyncDir fsyncs a directory to ensure rename visibility is durable.
func FsyncDir(dirPath string) error {
	d, err := os.Open(dirPath)
	if err != nil {
		return fmt.Errorf("fsync dir open: %w", err)
	}
	defer d.Close()
	return d.Sync()
}

// FsyncTree recursively fsyncs all files under the given root directory.
// This ensures all data is durably written to disk before marking an operation complete.
func FsyncTree(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Only sync files, not directories (directories are synced via FsyncDir)
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("open %s for fsync: %w", path, err)
			}
			if err := f.Sync(); err != nil {
				f.Close()
				return fmt.Errorf("fsync %s: %w", path, err)
			}
			f.Close()
		}
		return nil
	})
}
