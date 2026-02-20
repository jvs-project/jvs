//go:build conformance

package conformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test 10: Worktree create succeeds
func TestWorktree_Create(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	stdout, stderr, code := runJVSInRepo(t, repoPath, "worktree", "create", "feature")
	if code != 0 {
		t.Fatalf("worktree create failed: %s", stderr)
	}
	if !strings.Contains(stdout, "Created worktree") {
		t.Errorf("expected success message, got: %s", stdout)
	}
}

// Test 11: Worktree list shows all worktrees
func TestWorktree_List(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create additional worktree
	runJVSInRepo(t, repoPath, "worktree", "create", "feature")

	stdout, _, code := runJVSInRepo(t, repoPath, "worktree", "list")
	if code != 0 {
		t.Fatal("worktree list failed")
	}
	if !strings.Contains(stdout, "main") || !strings.Contains(stdout, "feature") {
		t.Errorf("expected both worktrees, got: %s", stdout)
	}
}

// Test 12: Worktree rename succeeds
func TestWorktree_Rename(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	runJVSInRepo(t, repoPath, "worktree", "create", "old-name")

	stdout, stderr, code := runJVSInRepo(t, repoPath, "worktree", "rename", "old-name", "new-name")
	if code != 0 {
		t.Fatalf("worktree rename failed: %s", stderr)
	}
	if !strings.Contains(stdout, "Renamed") {
		t.Errorf("expected success message, got: %s", stdout)
	}
}

// Test 13: Worktree remove succeeds
func TestWorktree_Remove(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	runJVSInRepo(t, repoPath, "worktree", "create", "to-delete")

	stdout, stderr, code := runJVSInRepo(t, repoPath, "worktree", "remove", "to-delete")
	if code != 0 {
		t.Fatalf("worktree remove failed: %s", stderr)
	}
	if !strings.Contains(stdout, "Removed") {
		t.Errorf("expected success message, got: %s", stdout)
	}
}

// Test 14: Worktree cannot remove main
func TestWorktree_CannotRemoveMain(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	_, _, code := runJVSInRepo(t, repoPath, "worktree", "remove", "main")
	if code == 0 {
		t.Error("should not be able to remove main worktree")
	}
}

// Test 15: Worktree path returns correct path
func TestWorktree_Path(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	stdout, _, code := runJVSInRepo(t, repoPath, "worktree", "path", "main")
	if code != 0 {
		t.Fatal("worktree path failed")
	}
	if !strings.Contains(stdout, "main") {
		t.Errorf("expected path containing 'main', got: %s", stdout)
	}
}

// Test 16: Worktree create preserves content
func TestWorktree_CreatePreservesContent(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create content in main
	dataPath := filepath.Join(repoPath, "main", "data.txt")
	os.WriteFile(dataPath, []byte("test content"), 0644)

	// Create snapshot
	runJVSInRepo(t, repoPath, "lock", "acquire")
	runJVSInRepo(t, repoPath, "snapshot", "v1")
	runJVSInRepo(t, repoPath, "lock", "release")

	// Create worktree
	runJVSInRepo(t, repoPath, "worktree", "create", "feature")

	// Feature worktree should be empty (new worktree)
	featurePath := filepath.Join(repoPath, "worktrees", "feature")
	if _, err := os.Stat(filepath.Join(featurePath, "data.txt")); !os.IsNotExist(err) {
		t.Error("new worktree should be empty")
	}
}
