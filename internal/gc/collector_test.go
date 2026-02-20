package gc_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jvs-project/jvs/internal/gc"
	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) string {
	dir := t.TempDir()
	_, err := repo.Init(dir, "test")
	require.NoError(t, err)
	return dir
}

func createTestSnapshot(t *testing.T, repoPath string) model.SnapshotID {
	// Add some content
	mainPath := filepath.Join(repoPath, "main")
	require.NoError(t, os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644))

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "test", nil)
	require.NoError(t, err)
	return desc.SnapshotID
}

func TestCollector_Plan(t *testing.T) {
	repoPath := setupTestRepo(t)

	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)
	assert.NotEmpty(t, plan.PlanID)
	// Fresh repo has no snapshots, so protected set may be empty
}

func TestCollector_Plan_WithSnapshots(t *testing.T) {
	repoPath := setupTestRepo(t)
	snapshotID := createTestSnapshot(t, repoPath)

	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)
	assert.NotEmpty(t, plan.ProtectedSet)
	// Protected set should contain the snapshot ID we just created
	assert.Contains(t, plan.ProtectedSet, snapshotID)
}

func TestCollector_Run(t *testing.T) {
	repoPath := setupTestRepo(t)
	createTestSnapshot(t, repoPath)

	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)

	err = collector.Run(plan.PlanID)
	require.NoError(t, err)
}

func TestCollector_Run_InvalidPlanID(t *testing.T) {
	repoPath := setupTestRepo(t)

	collector := gc.NewCollector(repoPath)
	err := collector.Run("nonexistent-plan-id")
	assert.Error(t, err)
}

func TestCollector_Plan_WithLineage(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create first snapshot
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file1.txt"), []byte("content1"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc1, err := creator.Create("main", "first", nil)
	require.NoError(t, err)

	// Create second snapshot (parent = first)
	os.WriteFile(filepath.Join(mainPath, "file2.txt"), []byte("content2"), 0644)
	desc2, err := creator.Create("main", "second", nil)
	require.NoError(t, err)

	// Verify lineage
	assert.Equal(t, desc1.SnapshotID, *desc2.ParentID)

	// GC plan should protect both (lineage traversal)
	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)

	assert.Contains(t, plan.ProtectedSet, desc1.SnapshotID)
	assert.Contains(t, plan.ProtectedSet, desc2.SnapshotID)
}

func TestCollector_Run_WithDeletions(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create snapshot in main worktree
	createTestSnapshot(t, repoPath)

	// Create another worktree with its own snapshot
	wtMgr := worktree.NewManager(repoPath)
	cfg, err := wtMgr.Create("feature", nil)
	require.NoError(t, err)

	featurePath := wtMgr.Path("feature")
	os.WriteFile(filepath.Join(featurePath, "file.txt"), []byte("feature content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	featureDesc, err := creator.Create("feature", "feature snapshot", nil)
	require.NoError(t, err)
	_ = cfg

	// Delete the feature worktree
	require.NoError(t, wtMgr.Remove("feature"))

	// Now the feature snapshot should be unprotected
	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)

	// Feature snapshot should be in toDelete since worktree was deleted
	assert.Contains(t, plan.ToDelete, featureDesc.SnapshotID)

	// Run GC to delete the unprotected snapshot
	err = collector.Run(plan.PlanID)
	require.NoError(t, err)

	// Verify snapshot was deleted
	snapshotsDir := filepath.Join(repoPath, ".jvs", "snapshots")
	entries, _ := os.ReadDir(snapshotsDir)
	for _, e := range entries {
		assert.NotEqual(t, string(featureDesc.SnapshotID), e.Name())
	}
}

func TestCollector_Plan_EmptyRepo(t *testing.T) {
	repoPath := setupTestRepo(t)

	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)
	assert.Empty(t, plan.ProtectedSet)
	assert.Empty(t, plan.ToDelete)
}
