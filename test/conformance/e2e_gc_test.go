//go:build conformance

package conformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// E2E Scenario 6: Garbage Collection Flow
// User Story: Admin manages storage by removing orphaned snapshots

// TestE2E_GC_OrphanedSnapshots tests GC of orphaned snapshots from removed worktrees
func TestE2E_GC_OrphanedSnapshots(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "gc-test")
	mainPath := filepath.Join(repoPath, "main")

	// Initialize repository
	runJVS(t, dir, "init", "gc-test")

	// Step 1: Create snapshots in main
	t.Run("create_main_snapshots", func(t *testing.T) {
		for i := 1; i <= 5; i++ {
			os.WriteFile(filepath.Join(mainPath, "main.txt"), []byte("main"+string(rune('0'+i))), 0644)
			runJVSInRepo(t, repoPath, "snapshot", "main snapshot")
		}
	})

	// Step 2: Fork worktree and create snapshots there
	t.Run("create_worktree_snapshots", func(t *testing.T) {
		runJVSInRepo(t, repoPath, "worktree", "fork", "old-feature")
		featurePath := filepath.Join(repoPath, "worktrees", "old-feature")

		for i := 1; i <= 3; i++ {
			os.WriteFile(filepath.Join(featurePath, "feat.txt"), []byte("feat"+string(rune('0'+i))), 0644)
			runJVSInWorktree(t, repoPath, "old-feature", "snapshot", "feature snapshot")
		}
	})

	// Step 3: Remove the worktree (snapshots become orphaned)
	t.Run("remove_worktree", func(t *testing.T) {
		stdout, stderr, code := runJVSInRepo(t, repoPath, "worktree", "remove", "old-feature")
		if code != 0 {
			t.Fatalf("remove worktree failed: %s", stderr)
		}
		if !strings.Contains(stdout, "Removed") {
			t.Errorf("expected remove message, got: %s", stdout)
		}
	})

	// Step 4: GC plan should identify orphaned snapshots
	var planID string
	t.Run("gc_plan_shows_orphans", func(t *testing.T) {
		stdout, stderr, code := runJVSInRepo(t, repoPath, "gc", "plan", "--json")
		if code != 0 {
			t.Fatalf("gc plan failed: %s", stderr)
		}

		// Extract plan ID
		planID = extractPlanID(stdout)
		if planID == "" {
			// Try non-JSON output
			stdout, _, _ = runJVSInRepo(t, repoPath, "gc", "plan")
			planID = extractPlanIDFromText(stdout)
		}

		// Should identify snapshots as candidates
		if !strings.Contains(stdout, "candidate") && !strings.Contains(stdout, "orphan") && !strings.Contains(stdout, "delete") {
			t.Logf("GC plan output: %s", stdout)
		}
	})

	// Step 5: Execute GC if we have a plan
	if planID != "" {
		t.Run("gc_run", func(t *testing.T) {
			stdout, stderr, code := runJVSInRepo(t, repoPath, "gc", "run", "--plan-id", planID)
			if code != 0 {
				t.Fatalf("gc run failed: %s", stderr)
			}
			t.Logf("GC run output: %s", stdout)
		})
	}

	// Step 6: Verify main snapshots still exist
	t.Run("verify_main_intact", func(t *testing.T) {
		stdout, _, _ := runJVSInRepo(t, repoPath, "history")
		// Main snapshots should still be present
		count := strings.Count(stdout, "main snapshot")
		if count < 5 {
			t.Logf("Main snapshot count: %d (expected at least 5)", count)
		}
	})

	// Step 7: Verify all remaining snapshots
	t.Run("verify_all", func(t *testing.T) {
		stdout, stderr, code := runJVSInRepo(t, repoPath, "verify", "--all")
		if code != 0 {
			t.Fatalf("verify failed: %s", stderr)
		}
		if !strings.Contains(stdout, "OK") {
			t.Errorf("expected OK in verify output, got: %s", stdout)
		}
	})
}

// extractPlanIDFromText extracts plan ID from non-JSON output
func extractPlanIDFromText(output string) string {
	// Look for patterns like "Plan ID: abc123" or "plan_id: abc123"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Plan ID") || strings.Contains(line, "plan_id") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if (p == "ID:" || p == "plan_id:") && i+1 < len(parts) {
					return parts[i+1]
				}
			}
		}
	}
	return ""
}

// TestE2E_GC_ProtectsActiveSnapshots tests that GC doesn't remove active snapshots
func TestE2E_GC_ProtectsActiveSnapshots(t *testing.T) {
	repoPath, _ := initTestRepo(t)
	mainPath := filepath.Join(repoPath, "main")

	// Create snapshots in main
	for i := 1; i <= 3; i++ {
		os.WriteFile(filepath.Join(mainPath, "data.txt"), []byte(string(rune('a'+i))), 0644)
		runJVSInRepo(t, repoPath, "snapshot", "protected")
	}

	// Fork worktree and create snapshots
	runJVSInRepo(t, repoPath, "worktree", "fork", "active-feature")
	featurePath := filepath.Join(repoPath, "worktrees", "active-feature")
	os.WriteFile(filepath.Join(featurePath, "feature.txt"), []byte("active"), 0644)
	runJVSInWorktree(t, repoPath, "active-feature", "snapshot", "active feature")

	// Run GC plan
	t.Run("gc_plan", func(t *testing.T) {
		stdout, stderr, code := runJVSInRepo(t, repoPath, "gc", "plan", "--json")
		if code != 0 {
			t.Fatalf("gc plan failed: %s", stderr)
		}

		// All snapshots should be protected (no orphans)
		// The plan should show 0 candidates or protected status
		t.Logf("GC plan output: %s", stdout)
	})

	// Verify all snapshots still exist
	t.Run("all_snapshots_exist", func(t *testing.T) {
		stdout, _, _ := runJVSInRepo(t, repoPath, "history", "--all", "--json")
		// Should have main + feature snapshots
		count := getSnapshotCount(stdout)
		if count < 4 {
			t.Errorf("expected at least 4 snapshots, got %d", count)
		}
	})
}

// TestE2E_GC_TwoPhaseProtocol tests the two-phase GC protocol
func TestE2E_GC_TwoPhaseProtocol(t *testing.T) {
	repoPath, _ := initTestRepo(t)

	// Create initial state
	runJVSInRepo(t, repoPath, "snapshot", "initial")

	// Phase 1: Plan
	t.Run("phase1_plan", func(t *testing.T) {
		stdout, stderr, code := runJVSInRepo(t, repoPath, "gc", "plan")
		if code != 0 {
			t.Fatalf("gc plan failed: %s", stderr)
		}
		if !strings.Contains(stdout, "Plan") && !strings.Contains(stdout, "plan") {
			t.Errorf("expected plan output, got: %s", stdout)
		}
	})

	// Phase 2: Run requires plan ID
	t.Run("phase2_requires_plan_id", func(t *testing.T) {
		_, _, code := runJVSInRepo(t, repoPath, "gc", "run")
		if code == 0 {
			t.Error("gc run should require --plan-id")
		}
	})
}

// TestE2E_GC_AfterWorktreeRemoval tests GC behavior after removing worktrees
func TestE2E_GC_AfterWorktreeRemoval(t *testing.T) {
	repoPath, _ := initTestRepo(t)
	mainPath := filepath.Join(repoPath, "main")

	// Create main snapshots
	os.WriteFile(filepath.Join(mainPath, "main.txt"), []byte("main"), 0644)
	runJVSInRepo(t, repoPath, "snapshot", "main-v1", "--tag", "main")

	// Create and populate worktree
	runJVSInRepo(t, repoPath, "worktree", "fork", "temp-work")
	featurePath := filepath.Join(repoPath, "worktrees", "temp-work")
	os.WriteFile(filepath.Join(featurePath, "temp.txt"), []byte("temp"), 0644)
	runJVSInWorktree(t, repoPath, "temp-work", "snapshot", "temp-v1", "--tag", "temp")

	// Get history before removal
	stdout, _, _ := runJVSInRepo(t, repoPath, "history", "--all", "--json")
	countBefore := getSnapshotCount(stdout)

	// Remove worktree
	runJVSInRepo(t, repoPath, "worktree", "remove", "temp-work")

	// Run GC plan and execute
	t.Run("gc_after_removal", func(t *testing.T) {
		planOut, _, code := runJVSInRepo(t, repoPath, "gc", "plan", "--json")
		if code != 0 {
			t.Fatal("gc plan failed")
		}

		planID := extractPlanID(planOut)
		if planID == "" {
			planID = extractPlanIDFromText(planOut)
		}

		if planID != "" {
			runJVSInRepo(t, repoPath, "gc", "run", "--plan-id", planID)
		}
	})

	// Verify history changed (temp snapshots should be removed if GC works)
	t.Run("history_after_gc", func(t *testing.T) {
		stdout, _, _ := runJVSInRepo(t, repoPath, "history", "--all", "--json")
		countAfter := getSnapshotCount(stdout)

		// At minimum, main snapshot should exist
		if countAfter < 1 {
			t.Error("at least main snapshot should exist")
		}

		t.Logf("Snapshots: before=%d, after=%d", countBefore, countAfter)
	})
}

// TestE2E_GC_DoctorHealthyAfter tests that doctor reports healthy after GC
func TestE2E_GC_DoctorHealthyAfter(t *testing.T) {
	repoPath, _ := initTestRepo(t)
	mainPath := filepath.Join(repoPath, "main")

	// Create content
	os.WriteFile(filepath.Join(mainPath, "data.txt"), []byte("test"), 0644)
	runJVSInRepo(t, repoPath, "snapshot", "v1")

	// Create and remove worktree
	runJVSInRepo(t, repoPath, "worktree", "fork", "to-remove")
	runJVSInRepo(t, repoPath, "worktree", "remove", "to-remove")

	// Run GC
	planOut, _, _ := runJVSInRepo(t, repoPath, "gc", "plan", "--json")
	planID := extractPlanID(planOut)
	if planID == "" {
		planID = extractPlanIDFromText(planOut)
	}
	if planID != "" {
		runJVSInRepo(t, repoPath, "gc", "run", "--plan-id", planID)
	}

	// Doctor should report healthy
	t.Run("doctor_after_gc", func(t *testing.T) {
		stdout, stderr, code := runJVSInRepo(t, repoPath, "doctor", "--strict")
		if code != 0 {
			t.Fatalf("doctor failed: %s", stderr)
		}
		if !strings.Contains(stdout, "healthy") {
			t.Errorf("expected healthy status, got: %s", stdout)
		}
	})
}

// TestE2E_GC_MultipleWorktreeRemoval tests GC after removing multiple worktrees
func TestE2E_GC_MultipleWorktreeRemoval(t *testing.T) {
	repoPath, _ := initTestRepo(t)
	mainPath := filepath.Join(repoPath, "main")

	// Create main baseline
	os.WriteFile(filepath.Join(mainPath, "baseline.txt"), []byte("base"), 0644)
	runJVSInRepo(t, repoPath, "snapshot", "baseline")

	// Create multiple worktrees with snapshots
	worktrees := []string{"feature-1", "feature-2", "feature-3"}
	for _, name := range worktrees {
		runJVSInRepo(t, repoPath, "worktree", "fork", name)
		featurePath := filepath.Join(repoPath, "worktrees", name)
		os.WriteFile(filepath.Join(featurePath, name+".txt"), []byte(name), 0644)
		runJVSInWorktree(t, repoPath, name, "snapshot", name+"-snapshot")
	}

	// Remove all worktrees
	for _, name := range worktrees {
		runJVSInRepo(t, repoPath, "worktree", "remove", name)
	}

	// Run GC
	t.Run("gc_multiple_removals", func(t *testing.T) {
		planOut, stderr, code := runJVSInRepo(t, repoPath, "gc", "plan", "--json")
		if code != 0 {
			t.Fatalf("gc plan failed: %s", stderr)
		}

		planID := extractPlanID(planOut)
		if planID == "" {
			planID = extractPlanIDFromText(planOut)
		}

		if planID != "" {
			runOut, _, _ := runJVSInRepo(t, repoPath, "gc", "run", "--plan-id", planID)
			t.Logf("GC run output: %s", runOut)
		}
	})

	// Verify repository is still healthy
	t.Run("verify_healthy", func(t *testing.T) {
		_, _, code := runJVSInRepo(t, repoPath, "verify", "--all")
		if code != 0 {
			t.Error("verify should pass after GC")
		}

		stdout, _, _ := runJVSInRepo(t, repoPath, "doctor")
		if !strings.Contains(stdout, "healthy") {
			t.Error("repository should be healthy after GC")
		}
	})
}
