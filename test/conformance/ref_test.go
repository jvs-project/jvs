//go:build conformance

package conformance

import (
	"strings"
	"testing"
)

// Test 17: Ref create succeeds
func TestRef_Create(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create snapshot first
	runJVSInRepo(t, repoPath, "lock", "acquire")
	runJVSInRepo(t, repoPath, "snapshot", "v1")

	// Get snapshot ID
	stdout, _, _ := runJVSInRepo(t, repoPath, "history", "--json")
	snapshotID := extractSnapshotID(stdout)
	if snapshotID == "" {
		t.Fatal("could not get snapshot ID")
	}

	runJVSInRepo(t, repoPath, "lock", "release")

	// Create ref
	stdout, stderr, code := runJVSInRepo(t, repoPath, "ref", "create", "v1.0", snapshotID)
	if code != 0 {
		t.Fatalf("ref create failed: %s", stderr)
	}
	if !strings.Contains(stdout, "Created ref") {
		t.Errorf("expected success message, got: %s", stdout)
	}
}

// Test 18: Ref list shows refs
func TestRef_List(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create snapshot and ref
	runJVSInRepo(t, repoPath, "lock", "acquire")
	runJVSInRepo(t, repoPath, "snapshot", "v1")
	stdout, _, _ := runJVSInRepo(t, repoPath, "history", "--json")
	snapshotID := extractSnapshotID(stdout)
	runJVSInRepo(t, repoPath, "lock", "release")

	runJVSInRepo(t, repoPath, "ref", "create", "v1.0", snapshotID)

	// List refs
	stdout, _, code := runJVSInRepo(t, repoPath, "ref", "list")
	if code != 0 {
		t.Fatal("ref list failed")
	}
	if !strings.Contains(stdout, "v1.0") {
		t.Errorf("expected ref in list, got: %s", stdout)
	}
}

// Test 19: Ref delete succeeds
func TestRef_Delete(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create snapshot and ref
	runJVSInRepo(t, repoPath, "lock", "acquire")
	runJVSInRepo(t, repoPath, "snapshot", "v1")
	stdout, _, _ := runJVSInRepo(t, repoPath, "history", "--json")
	snapshotID := extractSnapshotID(stdout)
	runJVSInRepo(t, repoPath, "lock", "release")

	runJVSInRepo(t, repoPath, "ref", "create", "v1.0", snapshotID)

	// Delete ref
	_, stderr, code := runJVSInRepo(t, repoPath, "ref", "delete", "v1.0")
	if code != 0 {
		t.Fatalf("ref delete failed: %s", stderr)
	}

	// Verify ref is gone
	stdout, _, _ = runJVSInRepo(t, repoPath, "ref", "list")
	if strings.Contains(stdout, "v1.0") {
		t.Error("ref should be deleted")
	}
}
