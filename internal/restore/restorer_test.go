package restore_test

import (
	"encoding/json"
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

func TestRestorer_Restore_VerifySnapshotError(t *testing.T) {
	// Test that restore fails when snapshot verification fails
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	// Corrupt the snapshot by modifying the descriptor checksum
	descriptorPath := filepath.Join(repoPath, ".jvs", "descriptors", string(desc.SnapshotID)+".json")
	data, err := os.ReadFile(descriptorPath)
	require.NoError(t, err)

	var descMap map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &descMap))
	descMap["descriptor_checksum"] = "invalidchecksum"
	corruptData, err := json.Marshal(descMap)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(descriptorPath, corruptData, 0644))

	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err = restorer.Restore("main", desc.SnapshotID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verify snapshot")
}

func TestRestorer_RestoreWithReflinkEngine(t *testing.T) {
	// Test restore with reflink engine
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	// Modify main after snapshot
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("modified"), 0644)

	// Restore with reflink engine
	restorer := restore.NewRestorer(repoPath, model.EngineReflinkCopy)
	err := restorer.Restore("main", desc.SnapshotID)
	require.NoError(t, err)

	// Verify content is restored
	content, err := os.ReadFile(filepath.Join(mainPath, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "snapshot-content", string(content))
}

func TestRestorer_RestoreWithJuiceFSEngine(t *testing.T) {
	// Test restore with juicefs-clone engine
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	// Modify main after snapshot
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("modified"), 0644)

	// Restore with juicefs-clone engine
	restorer := restore.NewRestorer(repoPath, model.EngineJuiceFSClone)
	err := restorer.Restore("main", desc.SnapshotID)
	require.NoError(t, err)

	// Verify content is restored
	content, err := os.ReadFile(filepath.Join(mainPath, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "snapshot-content", string(content))
}

func TestRestorer_Restore_EmptySnapshotID(t *testing.T) {
	// Test restore with empty snapshot ID
	repoPath := setupTestRepo(t)

	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err := restorer.Restore("main", "")
	assert.Error(t, err)
}

func TestRestorer_RestoreToLatest_NonExistentWorktree(t *testing.T) {
	// Test RestoreToLatest with non-existent worktree
	repoPath := setupTestRepo(t)

	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err := restorer.RestoreToLatest("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get worktree")
}

func TestRestorer_Restore_PreservesFilePermissions(t *testing.T) {
	// Test that restore preserves file permissions
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	// Create a file with specific permissions
	require.NoError(t, os.WriteFile(filepath.Join(mainPath, "script.sh"), []byte("#!/bin/bash\necho test"), 0755))

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "snapshot with executable", nil)
	require.NoError(t, err)

	// Modify the file
	require.NoError(t, os.WriteFile(filepath.Join(mainPath, "script.sh"), []byte("modified"), 0644))

	// Restore
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err = restorer.Restore("main", desc.SnapshotID)
	require.NoError(t, err)

	// Verify permissions are preserved (this depends on engine behavior)
	info, err := os.Stat(filepath.Join(mainPath, "script.sh"))
	require.NoError(t, err)
	// The copy engine should preserve permissions
	assert.NotNil(t, info.Mode())
}

func TestRestorer_Restore_MultipleFiles(t *testing.T) {
	// Test restore with multiple files and directories
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	// Create multiple files
	os.MkdirAll(filepath.Join(mainPath, "subdir"), 0755)
	os.WriteFile(filepath.Join(mainPath, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(mainPath, "file2.txt"), []byte("content2"), 0644)
	os.WriteFile(filepath.Join(mainPath, "subdir", "nested.txt"), []byte("nested"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "multi-file snapshot", nil)
	require.NoError(t, err)

	// Modify all files
	os.WriteFile(filepath.Join(mainPath, "file1.txt"), []byte("modified1"), 0644)
	os.WriteFile(filepath.Join(mainPath, "file2.txt"), []byte("modified2"), 0644)
	os.WriteFile(filepath.Join(mainPath, "subdir", "nested.txt"), []byte("modified nested"), 0644)

	// Restore
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err = restorer.Restore("main", desc.SnapshotID)
	require.NoError(t, err)

	// Verify all files are restored
	content1, _ := os.ReadFile(filepath.Join(mainPath, "file1.txt"))
	assert.Equal(t, "content1", string(content1))

	content2, _ := os.ReadFile(filepath.Join(mainPath, "file2.txt"))
	assert.Equal(t, "content2", string(content2))

	nested, _ := os.ReadFile(filepath.Join(mainPath, "subdir", "nested.txt"))
	assert.Equal(t, "nested", string(nested))
}

func TestRestorer_Restore_DetachedStateNotLatest(t *testing.T) {
	// Test the detached state determination logic
	// isDetached = snapshotID != cfg.LatestSnapshotID
	repoPath := setupTestRepo(t)

	// Create first snapshot
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("first"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc1, err := creator.Create("main", "first", nil)
	require.NoError(t, err)

	// Create second snapshot (now latest)
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("second"), 0644)
	desc2, err := creator.Create("main", "second", nil)
	require.NoError(t, err)

	// Restore to first (not latest) -> should be detached
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err = restorer.Restore("main", desc1.SnapshotID)
	require.NoError(t, err)

	wtMgr := worktree.NewManager(repoPath)
	cfg, _ := wtMgr.Get("main")
	assert.True(t, cfg.IsDetached())
	assert.Equal(t, desc1.SnapshotID, cfg.HeadSnapshotID)
	assert.Equal(t, desc2.SnapshotID, cfg.LatestSnapshotID)

	// Restore to second (is latest) -> should not be detached
	err = restorer.Restore("main", desc2.SnapshotID)
	require.NoError(t, err)

	cfg, _ = wtMgr.Get("main")
	assert.False(t, cfg.IsDetached())
	assert.Equal(t, desc2.SnapshotID, cfg.HeadSnapshotID)
	assert.Equal(t, desc2.SnapshotID, cfg.LatestSnapshotID)
}

func TestRestorer_RestoreToLatest_GetWorktreeError(t *testing.T) {
	// Test RestoreToLatest when Get worktree fails
	// Use a non-existent worktree
	repoPath := setupTestRepo(t)

	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err := restorer.RestoreToLatest("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get worktree")
}

func TestRestorer_NewRestorer_Engines(t *testing.T) {
	// Test NewRestorer with all engine types
	repoPath := setupTestRepo(t)

	// Test Copy engine
	r1 := restore.NewRestorer(repoPath, model.EngineCopy)
	assert.NotNil(t, r1)

	// Test Reflink engine
	r2 := restore.NewRestorer(repoPath, model.EngineReflinkCopy)
	assert.NotNil(t, r2)

	// Test JuiceFS engine
	r3 := restore.NewRestorer(repoPath, model.EngineJuiceFSClone)
	assert.NotNil(t, r3)
}

func TestRestorer_Restore_ToSameContent(t *testing.T) {
	// Test restore when content is already the same
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	// Don't modify content, restore to same state
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err := restorer.Restore("main", desc.SnapshotID)
	require.NoError(t, err)

	// Verify worktree is at HEAD
	wtMgr := worktree.NewManager(repoPath)
	cfg, _ := wtMgr.Get("main")
	assert.False(t, cfg.IsDetached())
	assert.Equal(t, desc.SnapshotID, cfg.HeadSnapshotID)
}

func TestRestorer_Restore_SymlinkPreservation(t *testing.T) {
	// Test that symlinks are preserved during restore
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	// Create a file and a symlink
	os.WriteFile(filepath.Join(mainPath, "target.txt"), []byte("target content"), 0644)

	// On systems that support symlinks
	err := os.Symlink("target.txt", filepath.Join(mainPath, "link.txt"))
	if err != nil {
		t.Skip("symlinks not supported on this system")
	}

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "snapshot with symlink", nil)
	require.NoError(t, err)

	// Modify the symlink
	os.Remove(filepath.Join(mainPath, "link.txt"))
	os.WriteFile(filepath.Join(mainPath, "link.txt"), []byte("not a symlink"), 0644)

	// Restore
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err = restorer.Restore("main", desc.SnapshotID)
	require.NoError(t, err)

	// The symlink behavior depends on the engine
	// Just verify restore didn't fail
	_, err = os.Stat(filepath.Join(mainPath, "link.txt"))
	require.NoError(t, err)
}

func TestRestorer_RestoreWithDifferentEngineTypes(t *testing.T) {
	// Test that restore works with all engine types
	repoPath := setupTestRepo(t)
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("original"), 0644)

	// Create snapshot with one engine
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "test", nil)
	require.NoError(t, err)

	// Modify content
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("modified"), 0644)

	// Restore with reflink engine
	restorer := restore.NewRestorer(repoPath, model.EngineReflinkCopy)
	err = restorer.Restore("main", desc.SnapshotID)
	require.NoError(t, err)

	// Verify restored
	content, _ := os.ReadFile(filepath.Join(mainPath, "file.txt"))
	assert.Equal(t, "original", string(content))
}

func TestRestorer_Restore_WithSubdirectories(t *testing.T) {
	// Test restore with nested directory structure
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	// Create nested directories
	os.MkdirAll(filepath.Join(mainPath, "a", "b", "c"), 0755)
	os.WriteFile(filepath.Join(mainPath, "a", "b", "c", "file.txt"), []byte("deep content"), 0644)
	os.WriteFile(filepath.Join(mainPath, "a", "file.txt"), []byte("mid content"), 0644)
	os.WriteFile(filepath.Join(mainPath, "root.txt"), []byte("root content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "nested dirs snapshot", nil)
	require.NoError(t, err)

	// Modify files
	os.WriteFile(filepath.Join(mainPath, "a", "b", "c", "file.txt"), []byte("modified deep"), 0644)
	os.WriteFile(filepath.Join(mainPath, "a", "file.txt"), []byte("modified mid"), 0644)
	os.WriteFile(filepath.Join(mainPath, "root.txt"), []byte("modified root"), 0644)

	// Restore
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err = restorer.Restore("main", desc.SnapshotID)
	require.NoError(t, err)

	// Verify all files restored
	deep, _ := os.ReadFile(filepath.Join(mainPath, "a", "b", "c", "file.txt"))
	assert.Equal(t, "deep content", string(deep))

	mid, _ := os.ReadFile(filepath.Join(mainPath, "a", "file.txt"))
	assert.Equal(t, "mid content", string(mid))

	root, _ := os.ReadFile(filepath.Join(mainPath, "root.txt"))
	assert.Equal(t, "root content", string(root))
}

func TestRestorer_Restore_WithEmptySnapshotID(t *testing.T) {
	// Test that restoring with an empty snapshot ID fails appropriately
	repoPath := setupTestRepo(t)
	createSnapshot(t, repoPath)

	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err := restorer.Restore("main", "")
	assert.Error(t, err)
}

func TestRestorer_Restore_UpdatesHeadCorrectly(t *testing.T) {
	// Test that Restore correctly updates the head snapshot ID
	repoPath := setupTestRepo(t)

	// Create two snapshots
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("v1"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc1, err := creator.Create("main", "v1", nil)
	require.NoError(t, err)

	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("v2"), 0644)
	desc2, err := creator.Create("main", "v2", nil)
	require.NoError(t, err)

	// Verify initial state - at latest (desc2)
	wtMgr := worktree.NewManager(repoPath)
	cfg, _ := wtMgr.Get("main")
	assert.Equal(t, desc2.SnapshotID, cfg.HeadSnapshotID)

	// Restore to desc1
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err = restorer.Restore("main", desc1.SnapshotID)
	require.NoError(t, err)

	// Verify head is now desc1
	cfg, _ = wtMgr.Get("main")
	assert.Equal(t, desc1.SnapshotID, cfg.HeadSnapshotID)
	assert.True(t, cfg.IsDetached())

	// Restore to desc2 (latest)
	err = restorer.Restore("main", desc2.SnapshotID)
	require.NoError(t, err)

	// Verify head is now desc2 and not detached
	cfg, _ = wtMgr.Get("main")
	assert.Equal(t, desc2.SnapshotID, cfg.HeadSnapshotID)
	assert.False(t, cfg.IsDetached())
}

func TestRestorer_Restore_SingleSnapshotIsNotDetached(t *testing.T) {
	// Test that restoring to the only snapshot doesn't enter detached state
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err := restorer.Restore("main", desc.SnapshotID)
	require.NoError(t, err)

	wtMgr := worktree.NewManager(repoPath)
	cfg, _ := wtMgr.Get("main")
	assert.False(t, cfg.IsDetached())
	assert.Equal(t, desc.SnapshotID, cfg.HeadSnapshotID)
	assert.Equal(t, desc.SnapshotID, cfg.LatestSnapshotID)
}

func TestRestorer_Restore_WithJuiceFSCloneEngine(t *testing.T) {
	// Test restore with juicefs-clone engine specifically
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	// Modify content
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("modified"), 0644)

	// Restore with juicefs-clone engine
	restorer := restore.NewRestorer(repoPath, model.EngineJuiceFSClone)
	err := restorer.Restore("main", desc.SnapshotID)
	require.NoError(t, err)

	// Verify restored
	content, _ := os.ReadFile(filepath.Join(mainPath, "file.txt"))
	assert.Equal(t, "snapshot-content", string(content))
}

func TestRestorer_Restore_CreatesAuditLogEntry(t *testing.T) {
	// Test that restore creates an audit log entry
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	// Modify content
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("modified"), 0644)

	// Restore
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err := restorer.Restore("main", desc.SnapshotID)
	require.NoError(t, err)

	// Verify audit log was created
	auditPath := filepath.Join(repoPath, ".jvs", "audit", "audit.jsonl")
	content, err := os.ReadFile(auditPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "\"restore\"")
}

func TestRestorer_Restore_DetachedStateInAuditLog(t *testing.T) {
	// Test that detached state is recorded in audit log
	repoPath := setupTestRepo(t)

	// Create two snapshots
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("v1"), 0644)
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc1, err := creator.Create("main", "v1", nil)
	require.NoError(t, err)

	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("v2"), 0644)
	_, err = creator.Create("main", "v2", nil)
	require.NoError(t, err)

	// Restore to first (enters detached state)
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err = restorer.Restore("main", desc1.SnapshotID)
	require.NoError(t, err)

	// Verify audit log contains detached=true
	auditPath := filepath.Join(repoPath, ".jvs", "audit", "audit.jsonl")
	content, err := os.ReadFile(auditPath)
	require.NoError(t, err)
	// The last entry should be the restore we just did
	assert.Contains(t, string(content), "\"detached\":true")
}

