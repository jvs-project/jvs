//go:build conformance

package conformance

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var jvsBinary string

func init() {
	// Find the jvs binary
	cwd, _ := os.Getwd()
	// Walk up to find bin/jvs
	for {
		binPath := filepath.Join(cwd, "bin", "jvs")
		if _, err := os.Stat(binPath); err == nil {
			jvsBinary = binPath
			return
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}
	// Fallback to PATH
	jvsBinary = "jvs"
}

// initTestRepo creates a temp repo and returns its path and cleanup function.
func initTestRepo(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "testrepo")

	runJVS(t, dir, "init", "testrepo")

	cleanup := func() {
		// Temp dir is auto-cleaned by testing package
	}
	return repoPath, cleanup
}

// runJVS executes the jvs binary with args in the given working directory.
func runJVS(t *testing.T, cwd string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(jvsBinary, args...)
	cmd.Dir = cwd
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	} else {
		exitCode = 0
	}
	return
}

// runJVSInRepo runs jvs from within the repo's main worktree.
func runJVSInRepo(t *testing.T, repoPath string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cwd := filepath.Join(repoPath, "main")
	return runJVS(t, cwd, args...)
}
