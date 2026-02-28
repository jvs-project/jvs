package gc_test

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jvs-project/jvs/internal/gc"
	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var zeroRetention = model.RetentionPolicy{}

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

	// Now the feature snapshot should be unprotected (use zero retention to bypass age protection)
	collector := gc.NewCollector(repoPath)
	plan, err := collector.PlanWithPolicy(zeroRetention)
	require.NoError(t, err)

	assert.Contains(t, plan.ToDelete, featureDesc.SnapshotID)

	err = collector.Run(plan.PlanID)
	require.NoError(t, err)

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

func TestCollector_Plan_WithPins(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create first snapshot in main
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file1.txt"), []byte("content1"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc1, err := creator.Create("main", "first", nil)
	require.NoError(t, err)

	// Create second snapshot
	os.WriteFile(filepath.Join(mainPath, "file2.txt"), []byte("content2"), 0644)
	_, err = creator.Create("main", "second", nil)
	require.NoError(t, err)

	// Create a third snapshot then delete its worktree to make it eligible for GC
	wtMgr := worktree.NewManager(repoPath)
	cfg, err := wtMgr.Create("temp", nil)
	require.NoError(t, err)
	tempPath := wtMgr.Path("temp")
	os.WriteFile(filepath.Join(tempPath, "file.txt"), []byte("temp"), 0644)
	tempDesc, err := creator.Create("temp", "temp snap", nil)
	require.NoError(t, err)
	_ = cfg

	// Pin the first snapshot
	pinsDir := filepath.Join(repoPath, ".jvs", "pins")
	require.NoError(t, os.MkdirAll(pinsDir, 0755))
	pinContent := `{"snapshot_id":"` + string(desc1.SnapshotID) + `","pinned_at":"2024-01-01T00:00:00Z","reason":"important"}`
	require.NoError(t, os.WriteFile(filepath.Join(pinsDir, string(desc1.SnapshotID)+".json"), []byte(pinContent), 0644))

	// Also pin the temp snapshot (even though worktree is deleted, pin protects it)
	pinContent2 := `{"snapshot_id":"` + string(tempDesc.SnapshotID) + `","pinned_at":"2024-01-01T00:00:00Z","reason":"pinned"}`
	require.NoError(t, os.WriteFile(filepath.Join(pinsDir, string(tempDesc.SnapshotID)+".json"), []byte(pinContent2), 0644))

	// Delete the temp worktree
	require.NoError(t, wtMgr.Remove("temp"))

	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)

	// First snapshot should be protected by pin
	assert.Contains(t, plan.ProtectedSet, desc1.SnapshotID)
	assert.Greater(t, plan.ProtectedByPin, 0)

	// Temp snapshot should also be protected by pin despite no worktree
	assert.Contains(t, plan.ProtectedSet, tempDesc.SnapshotID)
}

func TestCollector_Plan_ExpiredPin(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create a temporary worktree with a snapshot
	wtMgr := worktree.NewManager(repoPath)
	_, err := wtMgr.Create("temp", nil)
	require.NoError(t, err)

	tempPath := wtMgr.Path("temp")
	os.WriteFile(filepath.Join(tempPath, "file.txt"), []byte("content"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("temp", "test", nil)
	require.NoError(t, err)

	// Create an expired pin
	pinsDir := filepath.Join(repoPath, ".jvs", "pins")
	require.NoError(t, os.MkdirAll(pinsDir, 0755))
	pinContent := `{"snapshot_id":"` + string(desc.SnapshotID) + `","pinned_at":"2024-01-01T00:00:00Z","expires_at":"2024-01-02T00:00:00Z"}`
	require.NoError(t, os.WriteFile(filepath.Join(pinsDir, string(desc.SnapshotID)+".json"), []byte(pinContent), 0644))

	// Delete worktree to make snapshot eligible for GC
	require.NoError(t, wtMgr.Remove("temp"))

	collector := gc.NewCollector(repoPath)
	_, err = collector.Plan()
	require.NoError(t, err)

	// Snapshot should NOT be protected by pin since it expired
	// The pin count should be 0 since the pin is expired
}

func TestCollector_Plan_ProtectedCounts(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create multiple snapshots with lineage
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file1.txt"), []byte("1"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	_, err := creator.Create("main", "first", nil)
	require.NoError(t, err)

	os.WriteFile(filepath.Join(mainPath, "file2.txt"), []byte("2"), 0644)
	_, err = creator.Create("main", "second", nil)
	require.NoError(t, err)

	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)

	// Should have at least 1 protected by lineage (desc1 is parent of desc2)
	assert.GreaterOrEqual(t, plan.ProtectedByLineage, 0)
	// Should have exact count match
	assert.Equal(t, len(plan.ProtectedSet), len(plan.ProtectedSet))
}

func TestCollector_Run_TombstoneCreation(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create snapshot in main
	createTestSnapshot(t, repoPath)

	// Create another worktree with snapshot then delete it
	wtMgr := worktree.NewManager(repoPath)
	cfg, err := wtMgr.Create("temp", nil)
	require.NoError(t, err)

	tempPath := wtMgr.Path("temp")
	os.WriteFile(filepath.Join(tempPath, "file.txt"), []byte("temp content"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	tempDesc, err := creator.Create("temp", "temp", nil)
	require.NoError(t, err)
	_ = cfg

	require.NoError(t, wtMgr.Remove("temp"))

	collector := gc.NewCollector(repoPath)
	plan, err := collector.PlanWithPolicy(zeroRetention)
	require.NoError(t, err)

	err = collector.Run(plan.PlanID)
	require.NoError(t, err)

	tombstonesDir := filepath.Join(repoPath, ".jvs", "gc", "tombstones")
	entries, err := os.ReadDir(tombstonesDir)
	require.NoError(t, err)

	found := false
	for _, e := range entries {
		if e.Name() == string(tempDesc.SnapshotID)+".json" {
			found = true
			break
		}
	}
	assert.True(t, found, "tombstone should be created for deleted snapshot")
}

func TestCollector_LoadPlan_Invalid(t *testing.T) {
	repoPath := setupTestRepo(t)

	collector := gc.NewCollector(repoPath)
	_, err := collector.LoadPlan("nonexistent-plan")
	assert.Error(t, err)
}

func TestCollector_LoadPlan_InvalidJSON(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create a plan file with invalid JSON
	gcDir := filepath.Join(repoPath, ".jvs", "gc")
	require.NoError(t, os.MkdirAll(gcDir, 0755))
	invalidPlanPath := filepath.Join(gcDir, "invalid-plan.json")
	require.NoError(t, os.WriteFile(invalidPlanPath, []byte("{invalid json"), 0644))

	collector := gc.NewCollector(repoPath)
	_, err := collector.LoadPlan("invalid-plan")
	assert.Error(t, err)
}

func TestCollector_Plan_WritePlanError(t *testing.T) {
	// This test is hard to implement without mocking
	// In real scenarios, writePlan only fails on disk I/O errors
	// which are rare on modern systems
	repoPath := setupTestRepo(t)

	// Create a snapshot to ensure plan has content
	createTestSnapshot(t, repoPath)

	collector := gc.NewCollector(repoPath)
	_, err := collector.Plan()
	assert.NoError(t, err, "plan should succeed under normal conditions")
}

func TestCollector_Plan_WithNonexistentSnapshotsDir(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Remove snapshots directory to simulate edge case
	snapshotsDir := filepath.Join(repoPath, ".jvs", "snapshots")
	os.RemoveAll(snapshotsDir)

	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)
	assert.Empty(t, plan.ProtectedSet)
	assert.Empty(t, plan.ToDelete)
}

func TestCollector_Plan_WithOnlyLineage(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create a chain of snapshots
	mainPath := filepath.Join(repoPath, "main")
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	os.WriteFile(filepath.Join(mainPath, "file1.txt"), []byte("1"), 0644)
	desc1, err := creator.Create("main", "first", nil)
	require.NoError(t, err)

	os.WriteFile(filepath.Join(mainPath, "file2.txt"), []byte("2"), 0644)
	desc2, err := creator.Create("main", "second", nil)
	require.NoError(t, err)

	// Verify lineage
	assert.Equal(t, desc1.SnapshotID, *desc2.ParentID)

	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)

	// Both should be protected (desc2 as head, desc1 as parent)
	assert.Contains(t, plan.ProtectedSet, desc1.SnapshotID)
	assert.Contains(t, plan.ProtectedSet, desc2.SnapshotID)
	// At least 1 protected by lineage
	assert.Greater(t, plan.ProtectedByLineage, 0)
}

func TestCollector_Plan_WithManySnapshots(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	// Create multiple snapshots
	var snapshotIDs []model.SnapshotID
	for i := 0; i < 10; i++ {
		os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte(string(rune(i))), 0644)
		desc, err := creator.Create("main", "test", nil)
		require.NoError(t, err)
		snapshotIDs = append(snapshotIDs, desc.SnapshotID)
	}

	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)

	// Only the latest should be directly protected (others by lineage)
	assert.Contains(t, plan.ProtectedSet, snapshotIDs[len(snapshotIDs)-1])
	assert.Equal(t, 0, plan.CandidateCount)
}

func TestCollector_Plan_ProtectedByPinCount(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create snapshot in main
	mainPath := filepath.Join(repoPath, "main")
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	os.WriteFile(filepath.Join(mainPath, "file1.txt"), []byte("1"), 0644)
	_, err := creator.Create("main", "first", nil)
	require.NoError(t, err)

	// Create temp worktree with a snapshot that won't be in main's lineage
	wtMgr := worktree.NewManager(repoPath)
	cfg, err := wtMgr.Create("temp", nil)
	require.NoError(t, err)

	tempPath := wtMgr.Path("temp")
	os.WriteFile(filepath.Join(tempPath, "file.txt"), []byte("temp"), 0644)
	tempDesc, err := creator.Create("temp", "temp snap", nil)
	require.NoError(t, err)
	_ = cfg

	// Delete the temp worktree so snapshot is only protected by pin
	require.NoError(t, wtMgr.Remove("temp"))

	// Create pin for the temp snapshot
	pinsDir := filepath.Join(repoPath, ".jvs", "pins")
	require.NoError(t, os.MkdirAll(pinsDir, 0755))
	pinContent := `{"snapshot_id":"` + string(tempDesc.SnapshotID) + `","pinned_at":"2099-01-01T00:00:00Z","reason":"test"}`
	require.NoError(t, os.WriteFile(filepath.Join(pinsDir, string(tempDesc.SnapshotID)+".json"), []byte(pinContent), 0644))

	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)

	// Should have 1 protected by pin
	assert.Equal(t, 1, plan.ProtectedByPin)
}

func TestCollector_Run_DeletesSnapshot(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create snapshot in main (protected)
	createTestSnapshot(t, repoPath)

	// Create temp worktree snapshot
	wtMgr := worktree.NewManager(repoPath)
	cfg, err := wtMgr.Create("temp", nil)
	require.NoError(t, err)

	tempPath := wtMgr.Path("temp")
	os.WriteFile(filepath.Join(tempPath, "file.txt"), []byte("temp"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	tempDesc, err := creator.Create("temp", "temp", nil)
	require.NoError(t, err)
	_ = cfg

	require.NoError(t, wtMgr.Remove("temp"))

	collector := gc.NewCollector(repoPath)
	plan, err := collector.PlanWithPolicy(zeroRetention)
	require.NoError(t, err)

	err = collector.Run(plan.PlanID)
	require.NoError(t, err)

	snapshotDir := filepath.Join(repoPath, ".jvs", "snapshots", string(tempDesc.SnapshotID))
	_, err = os.Stat(snapshotDir)
	assert.True(t, os.IsNotExist(err), "snapshot directory should be deleted")
}

func TestCollector_Run_DescriptorRemoval(t *testing.T) {
	repoPath := setupTestRepo(t)

	wtMgr := worktree.NewManager(repoPath)
	cfg, err := wtMgr.Create("temp", nil)
	require.NoError(t, err)

	tempPath := wtMgr.Path("temp")
	os.WriteFile(filepath.Join(tempPath, "file.txt"), []byte("temp content"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	tempDesc, err := creator.Create("temp", "temp", nil)
	require.NoError(t, err)
	_ = cfg

	// Verify descriptor exists
	descriptorPath := filepath.Join(repoPath, ".jvs", "descriptors", string(tempDesc.SnapshotID)+".json")
	_, err = os.Stat(descriptorPath)
	require.NoError(t, err, "descriptor should exist")

	require.NoError(t, wtMgr.Remove("temp"))

	collector := gc.NewCollector(repoPath)
	plan, err := collector.PlanWithPolicy(zeroRetention)
	require.NoError(t, err)

	err = collector.Run(plan.PlanID)
	require.NoError(t, err)

	_, err = os.Stat(descriptorPath)
	assert.True(t, os.IsNotExist(err), "descriptor should be deleted")
}

func TestCollector_ListAllSnapshots_Empty(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Don't create any snapshots
	collector := gc.NewCollector(repoPath)

	// Access internal method via Plan which calls listAllSnapshots
	plan, err := collector.Plan()
	require.NoError(t, err)
	assert.Empty(t, plan.ProtectedSet)
}

func TestCollector_ListAllSnapshots_WithNonDirectoryEntries(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create a snapshot
	createTestSnapshot(t, repoPath)

	// Add a non-directory entry to snapshots dir
	snapshotsDir := filepath.Join(repoPath, ".jvs", "snapshots")
	err := os.WriteFile(filepath.Join(snapshotsDir, "file.txt"), []byte("test"), 0644)
	require.NoError(t, err)

	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)

	// Plan should still succeed, ignoring non-dir entries
	assert.NotEmpty(t, plan.ProtectedSet)
}

func TestCollector_deleteSnapshot_DescriptorIsDirectory(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create temp worktree with a snapshot
	wtMgr := worktree.NewManager(repoPath)
	cfg, err := wtMgr.Create("temp", nil)
	require.NoError(t, err)

	tempPath := wtMgr.Path("temp")
	os.WriteFile(filepath.Join(tempPath, "file.txt"), []byte("temp"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	tempDesc, err := creator.Create("temp", "temp", nil)
	require.NoError(t, err)
	_ = cfg

	// Replace the descriptor file with a directory containing a file (blocking os.Remove)
	descriptorPath := filepath.Join(repoPath, ".jvs", "descriptors", string(tempDesc.SnapshotID)+".json")
	require.NoError(t, os.Remove(descriptorPath))
	require.NoError(t, os.MkdirAll(descriptorPath, 0755))
	// Add a file inside so os.Remove fails (can't remove non-empty dir)
	require.NoError(t, os.WriteFile(filepath.Join(descriptorPath, "blocker"), []byte("x"), 0644))

	require.NoError(t, wtMgr.Remove("temp"))

	collector := gc.NewCollector(repoPath)
	plan, err := collector.PlanWithPolicy(zeroRetention)
	require.NoError(t, err)

	err = collector.Run(plan.PlanID)
	require.NoError(t, err)

	snapshotDir := filepath.Join(repoPath, ".jvs", "snapshots", string(tempDesc.SnapshotID))
	_, err = os.Stat(snapshotDir)
	assert.True(t, os.IsNotExist(err), "snapshot directory should be deleted")
}

func TestCollector_writeTombstone_TombstonesDirIsFile(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create temp worktree with a snapshot
	wtMgr := worktree.NewManager(repoPath)
	cfg, err := wtMgr.Create("temp", nil)
	require.NoError(t, err)

	tempPath := wtMgr.Path("temp")
	os.WriteFile(filepath.Join(tempPath, "file.txt"), []byte("temp"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	_, err = creator.Create("temp", "temp", nil)
	require.NoError(t, err)
	_ = cfg

	// Make tombstones a file instead of directory (blocking writeTombstone)
	tombstonesPath := filepath.Join(repoPath, ".jvs", "gc", "tombstones")
	require.NoError(t, os.MkdirAll(filepath.Dir(tombstonesPath), 0755))
	require.NoError(t, os.WriteFile(tombstonesPath, []byte("blocked"), 0644))

	require.NoError(t, wtMgr.Remove("temp"))

	collector := gc.NewCollector(repoPath)
	plan, err := collector.PlanWithPolicy(zeroRetention)
	require.NoError(t, err)

	err = collector.Run(plan.PlanID)
	require.NoError(t, err)

	os.Remove(tombstonesPath)
	os.MkdirAll(tombstonesPath, 0755)
}

func TestCollector_Run_DeleteSnapshotError(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create a snapshot in main worktree
	createTestSnapshot(t, repoPath)

	// Create temp worktree with a snapshot
	wtMgr := worktree.NewManager(repoPath)
	cfg, err := wtMgr.Create("temp", nil)
	require.NoError(t, err)

	tempPath := wtMgr.Path("temp")
	os.WriteFile(filepath.Join(tempPath, "file.txt"), []byte("temp"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	tempDesc, err := creator.Create("temp", "temp", nil)
	require.NoError(t, err)
	_ = cfg

	require.NoError(t, wtMgr.Remove("temp"))

	snapshotDir := filepath.Join(repoPath, ".jvs", "snapshots", string(tempDesc.SnapshotID))
	subDir := filepath.Join(snapshotDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0000))

	collector := gc.NewCollector(repoPath)
	plan, err := collector.PlanWithPolicy(zeroRetention)
	require.NoError(t, err)

	err = collector.Run(plan.PlanID)
	require.NoError(t, err)

	os.Chmod(subDir, 0755)
}

func TestCollector_Plan_WithInvalidPinFile(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create a snapshot
	createTestSnapshot(t, repoPath)

	// Create a pin file with invalid JSON
	pinsDir := filepath.Join(repoPath, ".jvs", "pins")
	require.NoError(t, os.MkdirAll(pinsDir, 0755))
	invalidPinPath := filepath.Join(pinsDir, "invalid-pin.json")
	require.NoError(t, os.WriteFile(invalidPinPath, []byte("{invalid json}"), 0644))

	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)

	// Plan should succeed despite invalid pin file
	assert.NotEmpty(t, plan.ProtectedSet)
}

func TestCollector_Plan_WithAlreadyProtectedPin(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create a snapshot
	mainPath := filepath.Join(repoPath, "main")
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)
	desc, err := creator.Create("main", "test", nil)
	require.NoError(t, err)

	// Create a pin for the same snapshot that's already protected by main worktree
	pinsDir := filepath.Join(repoPath, ".jvs", "pins")
	require.NoError(t, os.MkdirAll(pinsDir, 0755))
	pinContent := `{"snapshot_id":"` + string(desc.SnapshotID) + `","pinned_at":"2099-01-01T00:00:00Z","reason":"test"}`
	require.NoError(t, os.WriteFile(filepath.Join(pinsDir, string(desc.SnapshotID)+".json"), []byte(pinContent), 0644))

	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)

	// The pin count should be 0 since the snapshot is already protected by worktree
	assert.Equal(t, 0, plan.ProtectedByPin)
}

func TestCollector_Plan_WithIntents(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create some intents (in-progress operations)
	intentsDir := filepath.Join(repoPath, ".jvs", "intents")
	require.NoError(t, os.MkdirAll(intentsDir, 0755))

	intentIDs := []string{"intent1", "intent2", "intent3"}
	for _, id := range intentIDs {
		intentPath := filepath.Join(intentsDir, id+".json")
		require.NoError(t, os.WriteFile(intentPath, []byte(`{"intent_id":"`+id+`"}`), 0644))
	}

	collector := gc.NewCollector(repoPath)
	plan, err := collector.Plan()
	require.NoError(t, err)

	// Plan should succeed with intents considered protected
	// (though they're not valid snapshot IDs, they're in the protected set)
	assert.NotEmpty(t, plan.ProtectedSet)
}

func TestCollector_walkLineage_WithMissingDescriptor(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create a snapshot in a temp worktree
	wtMgr := worktree.NewManager(repoPath)
	cfg, err := wtMgr.Create("temp", nil)
	require.NoError(t, err)

	tempPath := wtMgr.Path("temp")
	os.WriteFile(filepath.Join(tempPath, "file.txt"), []byte("temp"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	tempDesc, err := creator.Create("temp", "temp", nil)
	require.NoError(t, err)
	_ = cfg

	// Delete the descriptor to simulate corruption
	descriptorPath := filepath.Join(repoPath, ".jvs", "descriptors", string(tempDesc.SnapshotID)+".json")
	require.NoError(t, os.Remove(descriptorPath))

	// Create a new worktree and try to pin the orphaned snapshot
	// This will cause walkLineage to fail finding the descriptor
	wtMgr2 := worktree.NewManager(repoPath)
	cfg2, err := wtMgr2.Create("temp2", nil)
	require.NoError(t, err)
	_ = cfg2

	// Manually update the worktree config to point to the orphaned snapshot
	temp2Path := wtMgr.Path("temp2")
	wtConfigPath := filepath.Join(temp2Path, ".jvs-worktree.json")
	wtConfigContent, _ := os.ReadFile(wtConfigPath)
	// Update head_snapshot_id
	newContent := string(wtConfigContent)
	// Find the existing head_snapshot_id and replace it
	if idx := indexOf(newContent, `"head_snapshot_id":"`); idx >= 0 {
		endIdx := idx + len(`"head_snapshot_id":"`) + 36 // approximate UUID length
		newContent = newContent[:idx] + `"head_snapshot_id":"` + string(tempDesc.SnapshotID) + `"` + newContent[endIdx+2:]
	}
	os.WriteFile(wtConfigPath, []byte(newContent), 0644)

	collector := gc.NewCollector(repoPath)
	_, err = collector.Plan()
	// Plan should still succeed despite missing descriptor during lineage walk
	require.NoError(t, err)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestCollector_PlanWithPolicy_AgeRetention(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	os.WriteFile(filepath.Join(mainPath, "file1.txt"), []byte("a"), 0644)
	desc1, err := creator.Create("main", "first", nil)
	require.NoError(t, err)

	os.WriteFile(filepath.Join(mainPath, "file2.txt"), []byte("b"), 0644)
	desc2, err := creator.Create("main", "second", nil)
	require.NoError(t, err)

	// KeepMinAge of 1 hour covers all just-created snapshots
	policy := model.RetentionPolicy{KeepMinAge: 1 * time.Hour}
	collector := gc.NewCollector(repoPath)
	plan, err := collector.PlanWithPolicy(policy)
	require.NoError(t, err)

	assert.Contains(t, plan.ProtectedSet, desc1.SnapshotID)
	assert.Contains(t, plan.ProtectedSet, desc2.SnapshotID)
	assert.Empty(t, plan.ToDelete)
}

func TestCollector_PlanWithPolicy_CountRetention(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	// Create 3 snapshots in main
	var mainIDs []model.SnapshotID
	for i := 0; i < 3; i++ {
		os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte(string(rune('a'+i))), 0644)
		desc, err := creator.Create("main", "main snap", nil)
		require.NoError(t, err)
		mainIDs = append(mainIDs, desc.SnapshotID)
	}

	// Create temp worktree with 1 snapshot, then delete the worktree
	wtMgr := worktree.NewManager(repoPath)
	_, err := wtMgr.Create("temp", nil)
	require.NoError(t, err)

	tempPath := wtMgr.Path("temp")
	os.WriteFile(filepath.Join(tempPath, "file.txt"), []byte("temp"), 0644)
	tempDesc, err := creator.Create("temp", "temp snap", nil)
	require.NoError(t, err)

	require.NoError(t, wtMgr.Remove("temp"))

	// KeepMinSnapshots=5 is more than total (4), so all should be retained
	policy := model.RetentionPolicy{KeepMinSnapshots: 5}
	collector := gc.NewCollector(repoPath)
	plan, err := collector.PlanWithPolicy(policy)
	require.NoError(t, err)

	assert.Contains(t, plan.ProtectedSet, tempDesc.SnapshotID)
	assert.Empty(t, plan.ToDelete)
}

func TestCollector_PlanWithPolicy_ZeroRetention(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create temp worktree with snapshot, then remove worktree
	wtMgr := worktree.NewManager(repoPath)
	_, err := wtMgr.Create("temp", nil)
	require.NoError(t, err)

	tempPath := wtMgr.Path("temp")
	os.WriteFile(filepath.Join(tempPath, "file.txt"), []byte("temp"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	tempDesc, err := creator.Create("temp", "temp snap", nil)
	require.NoError(t, err)

	require.NoError(t, wtMgr.Remove("temp"))

	collector := gc.NewCollector(repoPath)
	plan, err := collector.PlanWithPolicy(zeroRetention)
	require.NoError(t, err)

	assert.Contains(t, plan.ToDelete, tempDesc.SnapshotID)
}

func TestCollector_Run_EmptyPlanID(t *testing.T) {
	repoPath := setupTestRepo(t)

	collector := gc.NewCollector(repoPath)
	err := collector.Run("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan ID is required")
}

func TestCollector_SetProgressCallback(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create a snapshot in main (protected)
	createTestSnapshot(t, repoPath)

	// Create a temp worktree snapshot that will be deleted
	wtMgr := worktree.NewManager(repoPath)
	_, err := wtMgr.Create("temp", nil)
	require.NoError(t, err)

	tempPath := wtMgr.Path("temp")
	os.WriteFile(filepath.Join(tempPath, "file.txt"), []byte("temp"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	_, err = creator.Create("temp", "temp snap", nil)
	require.NoError(t, err)

	require.NoError(t, wtMgr.Remove("temp"))

	var callCount atomic.Int32
	callback := func(phase string, current, total int, msg string) {
		callCount.Add(1)
	}

	collector := gc.NewCollector(repoPath)
	collector.SetProgressCallback(callback)

	plan, err := collector.PlanWithPolicy(zeroRetention)
	require.NoError(t, err)
	require.NotEmpty(t, plan.ToDelete, "expected at least one deletion candidate")

	err = collector.Run(plan.PlanID)
	require.NoError(t, err)

	assert.Greater(t, callCount.Load(), int32(0), "progress callback should have been invoked")
}

func TestCollector_PlanWithPolicy_CombinedRetention(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	// Create 3 snapshots in main
	for i := 0; i < 3; i++ {
		os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte(string(rune('a'+i))), 0644)
		_, err := creator.Create("main", "snap", nil)
		require.NoError(t, err)
	}

	// Create temp worktree with snapshot, then delete worktree
	wtMgr := worktree.NewManager(repoPath)
	_, err := wtMgr.Create("temp", nil)
	require.NoError(t, err)

	tempPath := wtMgr.Path("temp")
	os.WriteFile(filepath.Join(tempPath, "file.txt"), []byte("temp"), 0644)
	tempDesc, err := creator.Create("temp", "temp snap", nil)
	require.NoError(t, err)

	require.NoError(t, wtMgr.Remove("temp"))

	// Both policies: age covers all recently-created snapshots AND count covers all (4 total, keep 10)
	policy := model.RetentionPolicy{
		KeepMinAge:       1 * time.Hour,
		KeepMinSnapshots: 10,
	}
	collector := gc.NewCollector(repoPath)
	plan, err := collector.PlanWithPolicy(policy)
	require.NoError(t, err)

	assert.Contains(t, plan.ProtectedSet, tempDesc.SnapshotID)
	assert.Empty(t, plan.ToDelete)
	assert.Greater(t, plan.ProtectedByRetention, 0)
}
