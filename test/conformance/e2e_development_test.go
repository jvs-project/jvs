//go:build conformance

package conformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// E2E Scenario 2: Daily Development Cycle
// User Story: Developer creates snapshots, restores to previous state, returns to HEAD

// TestE2E_Development_DailyCycle tests a typical development day workflow
func TestE2E_Development_DailyCycle(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "devproject")
	mainPath := filepath.Join(repoPath, "main")
	versionPath := filepath.Join(mainPath, "version.txt")

	// Initialize repository
	runJVS(t, dir, "init", "devproject")

	// Step 1: Create morning baseline with tag
	t.Run("morning_baseline", func(t *testing.T) {
		os.WriteFile(versionPath, []byte("v1"), 0644)
		stdout, stderr, code := runJVSInRepo(t, repoPath, "snapshot", "morning baseline", "--tag", "daily")
		if code != 0 {
			t.Fatalf("snapshot failed: %s", stderr)
		}
		if !strings.Contains(stdout, "Created snapshot") {
			t.Errorf("expected success message, got: %s", stdout)
		}
	})

	// Step 2: Create second snapshot (added feature)
	t.Run("added_feature", func(t *testing.T) {
		os.WriteFile(versionPath, []byte("v2"), 0644)
		runJVSInRepo(t, repoPath, "snapshot", "added feature")
	})

	// Step 3: Create third snapshot (version 3)
	t.Run("version_3", func(t *testing.T) {
		os.WriteFile(versionPath, []byte("v3"), 0644)
		runJVSInRepo(t, repoPath, "snapshot", "version 3")
	})

	// Get snapshot IDs from history
	var snapshots []string
	t.Run("get_snapshot_ids", func(t *testing.T) {
		stdout, _, _ := runJVSInRepo(t, repoPath, "history", "--json")
		snapshots = extractAllSnapshotIDs(stdout)
		if len(snapshots) < 3 {
			t.Fatalf("expected at least 3 snapshots, got %d", len(snapshots))
		}
	})

	// Step 4: Restore to v2 snapshot (detached state)
	v2Snapshot := snapshots[len(snapshots)-2] // second oldest
	t.Run("restore_to_v2", func(t *testing.T) {
		stdout, stderr, code := runJVSInRepo(t, repoPath, "restore", v2Snapshot)
		if code != 0 {
			t.Fatalf("restore failed: %s", stderr)
		}

		// Verify file content is v2
		content := readFile(t, mainPath, "version.txt")
		if content != "v2" {
			t.Errorf("expected 'v2', got '%s'", content)
		}

		// Verify detached state message
		if !strings.Contains(stdout, "DETACHED") && !strings.Contains(stderr, "DETACHED") {
			t.Error("expected DETACHED state message")
		}
	})

	// Step 5: Try to create snapshot in detached state - must fail
	t.Run("snapshot_fails_in_detached", func(t *testing.T) {
		_, _, code := runJVSInRepo(t, repoPath, "snapshot", "should fail")
		if code == 0 {
			t.Error("snapshot should fail in detached state")
		}
	})

	// Step 6: Restore HEAD to return to latest
	t.Run("restore_head", func(t *testing.T) {
		stdout, stderr, code := runJVSInRepo(t, repoPath, "restore", "HEAD")
		if code != 0 {
			t.Fatalf("restore HEAD failed: %s", stderr)
		}

		// Verify file content is back to v3
		content := readFile(t, mainPath, "version.txt")
		if content != "v3" {
			t.Errorf("expected 'v3', got '%s'", content)
		}

		// Verify HEAD state message
		if !strings.Contains(stdout, "HEAD") {
			t.Errorf("expected HEAD state message, got: %s", stdout)
		}
	})

	// Step 7: Can create snapshots after restore HEAD
	t.Run("continue_working", func(t *testing.T) {
		os.WriteFile(versionPath, []byte("v4"), 0644)
		stdout, stderr, code := runJVSInRepo(t, repoPath, "snapshot", "continue working")
		if code != 0 {
			t.Fatalf("snapshot failed: %s", stderr)
		}
		if !strings.Contains(stdout, "Created snapshot") {
			t.Errorf("expected success message, got: %s", stdout)
		}
	})
}

// TestE2E_Development_LineageChain tests that snapshots form a proper lineage
func TestE2E_Development_LineageChain(t *testing.T) {
	repoPath, _ := initTestRepo(t)
	mainPath := filepath.Join(repoPath, "main")

	// Create a series of snapshots
	for i := 1; i <= 5; i++ {
		os.WriteFile(filepath.Join(mainPath, "data.txt"), []byte(string(rune('a'+i-1))), 0644)
		runJVSInRepo(t, repoPath, "snapshot", "step")
	}

	// Get history and verify lineage
	stdout, _, _ := runJVSInRepo(t, repoPath, "history", "--json")
	snapshots := extractAllSnapshotIDs(stdout)

	if len(snapshots) != 5 {
		t.Errorf("expected 5 snapshots, got %d", len(snapshots))
	}

	// Verify we can restore to each snapshot in the chain
	for i, snapID := range snapshots {
		_, _, code := runJVSInRepo(t, repoPath, "restore", snapID)
		if code != 0 {
			t.Errorf("failed to restore snapshot %d (%s)", i, snapID)
		}
	}
}

// TestE2E_Development_WorkflowWithTags tests daily workflow with tag-based navigation
func TestE2E_Development_WorkflowWithTags(t *testing.T) {
	repoPath, _ := initTestRepo(t)
	mainPath := filepath.Join(repoPath, "main")

	// Create snapshots with meaningful tags
	os.WriteFile(filepath.Join(mainPath, "status.txt"), []byte("started"), 0644)
	runJVSInRepo(t, repoPath, "snapshot", "start of day", "--tag", "morning")

	os.WriteFile(filepath.Join(mainPath, "status.txt"), []byte("feature-done"), 0644)
	runJVSInRepo(t, repoPath, "snapshot", "feature complete", "--tag", "feature")

	os.WriteFile(filepath.Join(mainPath, "status.txt"), []byte("end"), 0644)
	runJVSInRepo(t, repoPath, "snapshot", "end of day", "--tag", "eod")

	// Restore by tag
	t.Run("restore_by_tag", func(t *testing.T) {
		_, stderr, code := runJVSInRepo(t, repoPath, "restore", "feature")
		if code != 0 {
			t.Fatalf("restore by tag failed: %s", stderr)
		}

		content := readFile(t, mainPath, "status.txt")
		if content != "feature-done" {
			t.Errorf("expected 'feature-done', got '%s'", content)
		}
	})

	// Return to HEAD
	t.Run("return_to_head", func(t *testing.T) {
		runJVSInRepo(t, repoPath, "restore", "HEAD")
		content := readFile(t, mainPath, "status.txt")
		if content != "end" {
			t.Errorf("expected 'end', got '%s'", content)
		}
	})
}
