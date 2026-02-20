//go:build conformance

package conformance

import (
	"path/filepath"
	"strings"
	"testing"
)

// Test 26: Inplace restore requires lock
func TestRestore_InplaceRequiresLock(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create snapshot
	runJVSInRepo(t, repoPath, "lock", "acquire")
	runJVSInRepo(t, repoPath, "snapshot", "v1")
	runJVSInRepo(t, repoPath, "lock", "release")

	// Get snapshot ID
	stdout, _, _ := runJVSInRepo(t, repoPath, "history", "--json")
	snapshotID := extractSnapshotID(stdout)

	// Try inplace restore without lock
	_, _, code := runJVSInRepo(t, repoPath, "restore", snapshotID, "--inplace", "--force", "--reason", "test")
	if code == 0 {
		t.Error("inplace restore should require lock")
	}
}

// Test 27: Inplace restore requires force flag
func TestRestore_InplaceRequiresForce(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create snapshot with lock
	runJVSInRepo(t, repoPath, "lock", "acquire")
	runJVSInRepo(t, repoPath, "snapshot", "v1")

	// Get snapshot ID
	stdout, _, _ := runJVSInRepo(t, repoPath, "history", "--json")
	snapshotID := extractSnapshotID(stdout)

	// Try inplace restore without force
	_, _, code := runJVSInRepo(t, repoPath, "restore", snapshotID, "--inplace", "--reason", "test")
	if code == 0 {
		t.Error("inplace restore should require --force")
	}

	runJVSInRepo(t, repoPath, "lock", "release")
}

// Test 28: History limit works
func TestHistory_Limit(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create 5 snapshots
	runJVSInRepo(t, repoPath, "lock", "acquire")
	for i := 0; i < 5; i++ {
		runJVSInRepo(t, repoPath, "snapshot", "test")
	}
	runJVSInRepo(t, repoPath, "lock", "release")

	// History with limit 2
	stdout, _, code := runJVSInRepo(t, repoPath, "history", "--limit", "2")
	if code != 0 {
		t.Fatal("history --limit failed")
	}

	lines := strings.Count(stdout, "\n")
	if lines > 3 { // 2 snapshots + possible header
		t.Errorf("expected at most 2 snapshots, got %d lines", lines)
	}
}

// Test 29: Invalid name rejected
func TestValidation_InvalidName(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Try to create worktree with invalid name
	_, _, code := runJVSInRepo(t, repoPath, "worktree", "create", "../evil")
	if code == 0 {
		t.Error("should reject path traversal in name")
	}
}

// Test 30: Verify with payload hash
func TestVerify_WithPayloadHash(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create snapshot
	runJVSInRepo(t, repoPath, "lock", "acquire")
	runJVSInRepo(t, repoPath, "snapshot", "v1")
	runJVSInRepo(t, repoPath, "lock", "release")

	// Get snapshot ID
	stdout, _, _ := runJVSInRepo(t, repoPath, "history", "--json")
	snapshotID := extractSnapshotID(stdout)

	// Verify with payload hash (single snapshot)
	_, stderr, code := runJVSInRepo(t, repoPath, "verify", snapshotID)
	if code != 0 {
		t.Fatalf("verify failed: %s", stderr)
	}
}

// Test 31: Multiple snapshots maintain lineage
func TestSnapshot_Lineage(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create multiple snapshots
	runJVSInRepo(t, repoPath, "lock", "acquire")
	runJVSInRepo(t, repoPath, "snapshot", "first")
	runJVSInRepo(t, repoPath, "snapshot", "second")
	runJVSInRepo(t, repoPath, "snapshot", "third")
	runJVSInRepo(t, repoPath, "lock", "release")

	// Verify all snapshots exist
	stdout, _, _ := runJVSInRepo(t, repoPath, "verify", "--all")
	if strings.Contains(stdout, "TAMPERED") {
		t.Error("snapshots should not be tampered")
	}
}

// Test 32: Worktree rename with active lock fails
func TestWorktree_RenameWithLockFails(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create worktree and acquire lock
	runJVSInRepo(t, repoPath, "worktree", "create", "feature")

	// Go to feature worktree and acquire lock
	featurePath := filepath.Join(repoPath, "worktrees", "feature")
	runJVS(t, featurePath, "lock", "acquire")

	// Try to rename from repo root
	_, _, code := runJVSInRepo(t, repoPath, "worktree", "rename", "feature", "new-feature")
	if code == 0 {
		t.Error("should not rename locked worktree")
	}
}

// Test 33: GC run with valid plan
func TestGC_RunWithPlan(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create snapshots
	runJVSInRepo(t, repoPath, "lock", "acquire")
	runJVSInRepo(t, repoPath, "snapshot", "v1")
	runJVSInRepo(t, repoPath, "snapshot", "v2")
	runJVSInRepo(t, repoPath, "lock", "release")

	// Create plan
	stdout, _, _ := runJVSInRepo(t, repoPath, "gc", "plan", "--json")
	planID := extractPlanID(stdout)
	if planID == "" {
		t.Fatal("could not get plan ID")
	}

	// Run GC (should succeed, though nothing to delete since all protected)
	_, _, code := runJVSInRepo(t, repoPath, "gc", "run", "--plan-id", planID)
	if code != 0 {
		t.Error("gc run should succeed")
	}
}

func extractPlanID(jsonOutput string) string {
	lines := strings.Split(jsonOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, `"plan_id"`) {
			parts := strings.Split(line, `"`)
			for i, p := range parts {
				if p == "plan_id" && i+2 < len(parts) {
					return parts[i+2]
				}
			}
		}
	}
	return ""
}
