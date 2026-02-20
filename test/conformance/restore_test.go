//go:build conformance

package conformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test 6: Safe restore creates new worktree
func TestRestore_SafeRestore(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create content and snapshot
	dataPath := filepath.Join(repoPath, "main", "data.txt")
	os.WriteFile(dataPath, []byte("original"), 0644)

	runJVSInRepo(t, repoPath, "snapshot", "v1")

	// Get snapshot ID from history
	stdout, _, _ := runJVSInRepo(t, repoPath, "history", "--json")
	snapshotID := extractSnapshotID(stdout)
	if snapshotID == "" {
		t.Fatal("could not get snapshot ID")
	}

	// Safe restore
	stdout, stderr, code := runJVSInRepo(t, repoPath, "restore", snapshotID)
	if code != 0 {
		t.Fatalf("restore failed: %s", stderr)
	}
	if !strings.Contains(stdout, "Restored snapshot") {
		t.Errorf("expected restore message, got: %s", stdout)
	}

	// Verify new worktree exists
	stdout, _, _ = runJVSInRepo(t, repoPath, "worktree", "list")
	if !strings.Contains(stdout, "restore-") {
		t.Error("restore worktree should exist")
	}
}

// Test 7: Restore by tag
func TestRestore_ByTag(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create content and snapshot with tag
	dataPath := filepath.Join(repoPath, "main", "data.txt")
	os.WriteFile(dataPath, []byte("tagged content"), 0644)

	runJVSInRepo(t, repoPath, "snapshot", "release v1", "--tag", "v1.0")

	// Restore by tag using --latest-tag
	stdout, stderr, code := runJVSInRepo(t, repoPath, "restore", "--latest-tag", "v1.0")
	if code != 0 {
		t.Fatalf("restore by tag failed: %s", stderr)
	}
	if !strings.Contains(stdout, "Restored snapshot") {
		t.Errorf("expected restore message, got: %s", stdout)
	}
}

// Test 8: Inplace restore requires force
func TestRestore_InplaceNeedsForce(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create snapshot
	runJVSInRepo(t, repoPath, "snapshot", "v1")

	// Get snapshot ID
	stdout, _, _ := runJVSInRepo(t, repoPath, "history", "--json")
	snapshotID := extractSnapshotID(stdout)

	// Try inplace restore without force
	_, _, code := runJVSInRepo(t, repoPath, "restore", snapshotID, "--inplace", "--reason", "test")
	if code == 0 {
		t.Error("inplace restore should require --force")
	}
}

// Test 9: Inplace restore requires reason
func TestRestore_InplaceRequiresReason(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create snapshot
	runJVSInRepo(t, repoPath, "snapshot", "v1")

	// Get snapshot ID
	stdout, _, _ := runJVSInRepo(t, repoPath, "history", "--json")
	snapshotID := extractSnapshotID(stdout)

	// Try inplace restore without reason
	_, _, code := runJVSInRepo(t, repoPath, "restore", snapshotID, "--inplace", "--force")
	if code == 0 {
		t.Error("inplace restore should require --reason")
	}
}

func extractSnapshotID(historyJSON string) string {
	// Simple extraction from JSON output
	lines := strings.Split(historyJSON, "\n")
	for _, line := range lines {
		if strings.Contains(line, `"snapshot_id"`) {
			parts := strings.Split(line, `"`)
			for i, p := range parts {
				if p == "snapshot_id" && i+2 < len(parts) {
					return parts[i+2]
				}
			}
		}
	}
	return ""
}
