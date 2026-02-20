//go:build conformance

package conformance

import (
	"strings"
	"testing"
)

// Test 4: Snapshot creates successfully
func TestSnapshot_Basic(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create snapshot
	stdout, stderr, code := runJVSInRepo(t, repoPath, "snapshot", "test snapshot")
	if code != 0 {
		t.Fatalf("snapshot failed: %s", stderr)
	}
	if !strings.Contains(stdout, "Created snapshot") {
		t.Errorf("expected 'Created snapshot' in output, got: %s", stdout)
	}
}

// Test 5: Snapshot with tags
func TestSnapshot_WithTags(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create snapshot with tags
	stdout, stderr, code := runJVSInRepo(t, repoPath, "snapshot", "release", "--tag", "v1.0", "--tag", "release")
	if code != 0 {
		t.Fatalf("snapshot with tags failed: %s", stderr)
	}
	if !strings.Contains(stdout, "Created snapshot") {
		t.Errorf("expected 'Created snapshot' in output, got: %s", stdout)
	}

	// Verify tags appear in history
	historyOut, _, _ := runJVSInRepo(t, repoPath, "history")
	if !strings.Contains(historyOut, "release") {
		t.Error("expected tag in history output")
	}
}
