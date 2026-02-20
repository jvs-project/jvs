package restore_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/internal/restore"
	"github.com/jvs-project/jvs/internal/snapshot"
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

func TestRestorer_SafeRestore(t *testing.T) {
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	// Modify main after snapshot
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("modified"), 0644)

	// Safe restore creates new worktree
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	cfg, err := restorer.SafeRestore(desc.SnapshotID, "restored-1", nil)
	require.NoError(t, err)
	assert.Equal(t, "restored-1", cfg.Name)
	assert.Equal(t, desc.SnapshotID, cfg.HeadSnapshotID)

	// Verify restored content
	restoredPath := filepath.Join(repoPath, "worktrees", "restored-1")
	content, err := os.ReadFile(filepath.Join(restoredPath, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "snapshot-content", string(content))

	// Original main should still have modified content
	content, err = os.ReadFile(filepath.Join(mainPath, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "modified", string(content))
}

func TestRestorer_SafeRestore_AutoName(t *testing.T) {
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	cfg, err := restorer.SafeRestore(desc.SnapshotID, "", nil) // auto-name
	require.NoError(t, err)
	assert.Contains(t, cfg.Name, "restore-")
}

func TestRestorer_InplaceRestore(t *testing.T) {
	repoPath := setupTestRepo(t)
	desc := createSnapshot(t, repoPath)

	// Modify main after snapshot
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("modified"), 0644)

	// Inplace restore
	restorer := restore.NewRestorer(repoPath, model.EngineCopy)
	err := restorer.InplaceRestore(desc.SnapshotID, "testing inplace restore")
	require.NoError(t, err)

	// Verify content is restored
	content, err := os.ReadFile(filepath.Join(mainPath, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "snapshot-content", string(content))
}
