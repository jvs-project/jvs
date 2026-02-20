//go:build conformance

package conformance

import (
	"strings"
	"testing"
)

// Test 4: Snapshot requires lock
func TestSnapshot_RequiresLock(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Try snapshot without lock
	_, _, code := runJVSInRepo(t, repoPath, "snapshot", "test")
	if code == 0 {
		t.Error("snapshot should require lock")
	}
}

// Test 5: Snapshot succeeds with lock
func TestSnapshot_WithLock(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Acquire lock
	runJVSInRepo(t, repoPath, "lock", "acquire")

	// Create snapshot
	stdout, stderr, code := runJVSInRepo(t, repoPath, "snapshot", "test snapshot")
	if code != 0 {
		t.Fatalf("snapshot failed: %s", stderr)
	}
	if !strings.Contains(stdout, "Created snapshot") {
		t.Errorf("expected 'Created snapshot' in output, got: %s", stdout)
	}
}
