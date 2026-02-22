package worktree_test

import (
	"path/filepath"
	"testing"

	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/errclass"
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

func TestManager_Create(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	cfg, err := mgr.Create("feature", nil)
	require.NoError(t, err)
	assert.Equal(t, "feature", cfg.Name)

	// Config file exists
	assert.FileExists(t, filepath.Join(repoPath, ".jvs", "worktrees", "feature", "config.json"))

	// Payload directory exists
	assert.DirExists(t, filepath.Join(repoPath, "worktrees", "feature"))
}

func TestManager_Create_InvalidName(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	_, err := mgr.Create("../evil", nil)
	require.ErrorIs(t, err, errclass.ErrNameInvalid)
}

func TestManager_List(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	mgr.Create("feature1", nil)
	mgr.Create("feature2", nil)

	list, err := mgr.List()
	require.NoError(t, err)
	assert.Len(t, list, 3) // main + feature1 + feature2

	names := make(map[string]bool)
	for _, cfg := range list {
		names[cfg.Name] = true
	}
	assert.True(t, names["main"])
	assert.True(t, names["feature1"])
	assert.True(t, names["feature2"])
}

func TestManager_Path(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	mainPath := mgr.Path("main")
	assert.Equal(t, filepath.Join(repoPath, "main"), mainPath)

	featurePath := mgr.Path("feature")
	assert.Equal(t, filepath.Join(repoPath, "worktrees", "feature"), featurePath)
}

func TestManager_Rename(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	mgr.Create("old-name", nil)
	err := mgr.Rename("old-name", "new-name")
	require.NoError(t, err)

	// Old should not exist
	assert.NoDirExists(t, filepath.Join(repoPath, "worktrees", "old-name"))

	// New should exist
	assert.DirExists(t, filepath.Join(repoPath, "worktrees", "new-name"))
	cfg, err := mgr.Get("new-name")
	require.NoError(t, err)
	assert.Equal(t, "new-name", cfg.Name)
}

func TestManager_Remove(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	mgr.Create("to-delete", nil)
	err := mgr.Remove("to-delete")
	require.NoError(t, err)

	assert.NoDirExists(t, filepath.Join(repoPath, "worktrees", "to-delete"))
	assert.NoFileExists(t, filepath.Join(repoPath, ".jvs", "worktrees", "to-delete", "config.json"))
}

func TestManager_UpdateHead(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	err := mgr.UpdateHead("main", "1708300800000-abc12345")
	require.NoError(t, err)

	cfg, err := mgr.Get("main")
	require.NoError(t, err)
	assert.Equal(t, model.SnapshotID("1708300800000-abc12345"), cfg.HeadSnapshotID)
}

func TestManager_Get(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	cfg, err := mgr.Get("main")
	require.NoError(t, err)
	assert.Equal(t, "main", cfg.Name)
}

func TestManager_Get_NotFound(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	_, err := mgr.Get("nonexistent")
	assert.Error(t, err)
}

func TestManager_CannotRemoveMain(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	err := mgr.Remove("main")
	assert.Error(t, err)
}

func TestManager_Create_AlreadyExists(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	_, err := mgr.Create("feature", nil)
	require.NoError(t, err)

	_, err = mgr.Create("feature", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestManager_CreateFromSnapshot(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	cloneFunc := func(src, dst string) error {
		return nil // mock clone
	}

	cfg, err := mgr.CreateFromSnapshot("from-snap", "1708300800000-a3f7c1b2", cloneFunc)
	require.NoError(t, err)
	assert.Equal(t, "from-snap", cfg.Name)
	assert.Equal(t, model.SnapshotID("1708300800000-a3f7c1b2"), cfg.BaseSnapshotID)
}

func TestManager_CreateFromSnapshot_InvalidName(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	cloneFunc := func(src, dst string) error { return nil }

	_, err := mgr.CreateFromSnapshot("../evil", "1708300800000-a3f7c1b2", cloneFunc)
	require.ErrorIs(t, err, errclass.ErrNameInvalid)
}

func TestManager_CreateFromSnapshot_AlreadyExists(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	cloneFunc := func(src, dst string) error { return nil }

	_, err := mgr.CreateFromSnapshot("feature", "1708300800000-a3f7c1b2", cloneFunc)
	require.NoError(t, err)

	_, err = mgr.CreateFromSnapshot("feature", "1708300900000-b4d8e2c3", cloneFunc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestManager_CreateWithBaseSnapshot(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	snapID := model.SnapshotID("1708300800000-a3f7c1b2")
	cfg, err := mgr.Create("with-base", &snapID)
	require.NoError(t, err)
	assert.Equal(t, snapID, cfg.HeadSnapshotID)
}

func TestManager_Rename_InvalidName(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	mgr.Create("old-name", nil)
	err := mgr.Rename("old-name", "../evil")
	require.ErrorIs(t, err, errclass.ErrNameInvalid)
}

func TestManager_Rename_AlreadyExists(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	mgr.Create("name1", nil)
	mgr.Create("name2", nil)
	err := mgr.Rename("name1", "name2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestManager_SetLatest(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	err := mgr.SetLatest("main", "1708300800000-a3f7c1b2")
	require.NoError(t, err)

	cfg, err := mgr.Get("main")
	require.NoError(t, err)
	assert.Equal(t, model.SnapshotID("1708300800000-a3f7c1b2"), cfg.HeadSnapshotID)
	assert.Equal(t, model.SnapshotID("1708300800000-a3f7c1b2"), cfg.LatestSnapshotID)
}

func TestManager_Fork(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	cloneFunc := func(src, dst string) error {
		return nil // mock clone
	}

	cfg, err := mgr.Fork("1708300800000-a3f7c1b2", "forked", cloneFunc)
	require.NoError(t, err)
	assert.Equal(t, "forked", cfg.Name)
	assert.Equal(t, model.SnapshotID("1708300800000-a3f7c1b2"), cfg.HeadSnapshotID)
	assert.Equal(t, model.SnapshotID("1708300800000-a3f7c1b2"), cfg.LatestSnapshotID)
	assert.Equal(t, model.SnapshotID("1708300800000-a3f7c1b2"), cfg.BaseSnapshotID)
}

func TestManager_Fork_InvalidName(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	cloneFunc := func(src, dst string) error { return nil }

	_, err := mgr.Fork("1708300800000-a3f7c1b2", "../evil", cloneFunc)
	require.ErrorIs(t, err, errclass.ErrNameInvalid)
}

func TestManager_Fork_AlreadyExists(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	cloneFunc := func(src, dst string) error { return nil }

	_, err := mgr.Fork("1708300800000-a3f7c1b2", "feature", cloneFunc)
	require.NoError(t, err)

	_, err = mgr.Fork("1708300900000-b4d8e2c3", "feature", cloneFunc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestManager_Remove_NonExistent(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	// Removing non-existent worktree should not error (idempotent)
	err := mgr.Remove("nonexistent")
	assert.NoError(t, err)
}

func TestManager_UpdateHead_NonExistent(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	err := mgr.UpdateHead("nonexistent", "1708300800000-a3f7c1b2")
	assert.Error(t, err)
}

func TestManager_SetLatest_NonExistent(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	err := mgr.SetLatest("nonexistent", "1708300800000-a3f7c1b2")
	assert.Error(t, err)
}

func TestManager_List_EmptyAdditionalWorktrees(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := worktree.NewManager(repoPath)

	list, err := mgr.List()
	require.NoError(t, err)
	assert.Len(t, list, 1) // Only main
	assert.Equal(t, "main", list[0].Name)
}
