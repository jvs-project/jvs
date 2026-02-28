//go:build windows

package audit

import "os"

// lockFile is a no-op on Windows; the in-process mutex provides sufficient
// protection for a single-user CLI tool.
func lockFile(_ *os.File) error   { return nil }
func unlockFile(_ *os.File) error { return nil }
