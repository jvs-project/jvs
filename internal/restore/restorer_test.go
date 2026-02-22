package restore_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/internal/restore"
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

func createSnapshot(t *testing.T, repoPath string) *model.Descriptor {
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("snapshot-content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "test snapshot", nil)
	require.NoError(t, err)

	return desc
}

func TestRestorer_Restore(t *testing.T) {
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	// Modify main after snapshot
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("modified"), 0644)

	// Restore (now always inplace)
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err := restorer.Restore("main", desc.SnapshotID)
	require.NoError(t, err)

	// Verify content is restored
	content, err := os.ReadFile(filepath.Join(mainPath, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "snapshot-content", string(content))

	// Verify worktree state (since this is the only snapshot, we're at HEAD, not detached)
	wtMgr := worktree.NewManager(repoPath)
	cfg, err := wtMgr.Get("main")
	require.NoError(t, err)
	// After restoring to the only snapshot, we're at HEAD (not detached)
	assert.False(t, cfg.IsDetached())
	assert.Equal(t, desc.SnapshotID, cfg.HeadSnapshotID)
	assert.Equal(t, desc.SnapshotID, cfg.LatestSnapshotID)
}

func TestRestorer_RestoreToLatest(t *testing.T) {
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	// Modify and create second snapshot
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("second"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc2, err := creator.Create("main", "second snapshot", nil)
	require.NoError(t, err)

	// Restore to first snapshot (detached)
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err = restorer.Restore("main", desc.SnapshotID)
	require.NoError(t, err)

	wtMgr := worktree.NewManager(repoPath)
	cfg, _ := wtMgr.Get("main")
	assert.True(t, cfg.IsDetached())

	// Restore to latest
	err = restorer.RestoreToLatest("main")
	require.NoError(t, err)

	// Verify content is from second snapshot
	content, err := os.ReadFile(filepath.Join(mainPath, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "second", string(content))

	// Verify worktree is back at HEAD
	cfg, _ = wtMgr.Get("main")
	assert.False(t, cfg.IsDetached())
	assert.Equal(t, desc2.SnapshotID, cfg.HeadSnapshotID)
	assert.Equal(t, desc2.SnapshotID, cfg.LatestSnapshotID)
}

func TestRestorer_Restore_SetsDetachedState(t *testing.T) {
	repoPath := setupTestRepo(t)
	desc1 := createSnapshot(t, repoPath)

	// Create second snapshot
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("second"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc2, err := creator.Create("main", "second snapshot", nil)
	require.NoError(t, err)

	// Verify we're at HEAD
	wtMgr := worktree.NewManager(repoPath)
	cfg, _ := wtMgr.Get("main")
	assert.False(t, cfg.IsDetached())
	assert.Equal(t, desc2.SnapshotID, cfg.LatestSnapshotID)

	// Restore to first snapshot
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err = restorer.Restore("main", desc1.SnapshotID)
	require.NoError(t, err)

	// Verify detached state
	cfg, _ = wtMgr.Get("main")
	assert.True(t, cfg.IsDetached())
	assert.Equal(t, desc1.SnapshotID, cfg.HeadSnapshotID)
	assert.Equal(t, desc2.SnapshotID, cfg.LatestSnapshotID) // Latest unchanged
}

func TestWorktree_Fork(t *testing.T) {
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	// Fork from snapshot
	wtMgr := worktree.NewManager(repoPath)
	eng := &mockEngine{content: "snapshot-content"}
	cfg, err := wtMgr.Fork(desc.SnapshotID, "feature", eng.clone)
	require.NoError(t, err)
	assert.Equal(t, "feature", cfg.Name)
	assert.Equal(t, desc.SnapshotID, cfg.HeadSnapshotID)
	assert.Equal(t, desc.SnapshotID, cfg.LatestSnapshotID)
	assert.False(t, cfg.IsDetached())

	// Verify forked content
	forkPath := filepath.Join(repoPath, "worktrees", "feature")
	content, err := os.ReadFile(filepath.Join(forkPath, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "snapshot-content", string(content))
}

// mockEngine for testing
type mockEngine struct {
	content string
}

func (m *mockEngine) clone(src, dst string) error {
	// Copy test content
	return os.WriteFile(filepath.Join(dst, "file.txt"), []byte(m.content), 0644)
}

func TestRestorer_Restore_NonExistentSnapshot(t *testing.T) {
	repoPath := setupTestRepo(t)

	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err := restorer.Restore("main", "nonexistent-snapshot-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load snapshot")
}

func TestRestorer_Restore_NonExistentWorktree(t *testing.T) {
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err := restorer.Restore("nonexistent", desc.SnapshotID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get worktree")
}

func TestRestorer_RestoreToLatest_NoSnapshots(t *testing.T) {
	repoPath := setupTestRepo(t)

	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err := restorer.RestoreToLatest("main")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no snapshots")
}

func TestRestorer_Restore_SameSnapshot(t *testing.T) {
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	// Restore to same snapshot (no-op effectively)
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err := restorer.Restore("main", desc.SnapshotID)
	require.NoError(t, err)

	// Verify worktree is at HEAD (not detached)
	wtMgr := worktree.NewManager(repoPath)
	cfg, _ := wtMgr.Get("main")
	assert.False(t, cfg.IsDetached())
}

func TestRestorer_Restore_MultipleTimes(t *testing.T) {
	repoPath := setupTestRepo(t)
	desc1 := createSnapshot(t, repoPath)

	// Create second snapshot
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("second"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc2, err := creator.Create("main", "second snapshot", nil)
	require.NoError(t, err)

	// Create third snapshot
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("third"), 0644)
	_, err = creator.Create("main", "third snapshot", nil)
	require.NoError(t, err)

	restorer := restore.NewRestorer(repoPath, model.EngineCopy)

	// Restore to first
	err = restorer.Restore("main", desc1.SnapshotID)
	require.NoError(t, err)

	// Restore to second
	err = restorer.Restore("main", desc2.SnapshotID)
	require.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(filepath.Join(mainPath, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "second", string(content))
}

func TestRestorer_NewRestorer(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Test with different engine types
	r1 := restore.NewRestorer(repoPath, model.EngineCopy)
	assert.NotNil(t, r1)

	r2 := restore.NewRestorer(repoPath, model.EngineJuiceFSClone)
	assert.NotNil(t, r2)

	r3 := restore.NewRestorer(repoPath, model.EngineReflinkCopy)
	assert.NotNil(t, r3)
}

func TestRestorer_RestoreToLatest_FromDetached(t *testing.T) {
	repoPath := setupTestRepo(t)
	desc1 := createSnapshot(t, repoPath)

	// Create second snapshot
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("second"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc2, err := creator.Create("main", "second snapshot", nil)
	require.NoError(t, err)

	// Restore to first (detached)
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err = restorer.Restore("main", desc1.SnapshotID)
	require.NoError(t, err)

	wtMgr := worktree.NewManager(repoPath)
	cfg, _ := wtMgr.Get("main")
	assert.True(t, cfg.IsDetached())

	// Restore to latest (exit detached)
	err = restorer.RestoreToLatest("main")
	require.NoError(t, err)

	cfg, _ = wtMgr.Get("main")
	assert.False(t, cfg.IsDetached())
	assert.Equal(t, desc2.SnapshotID, cfg.HeadSnapshotID)
}
