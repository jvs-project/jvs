package snapshot_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jvs-project/jvs/internal/lock"
	"github.com/jvs-project/jvs/internal/repo"
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

func acquireLock(t *testing.T, repoPath string) *model.LockRecord {
	mgr := lock.NewManager(repoPath, model.LockPolicy{DefaultLeaseTTL: time.Hour})
	rec, err := mgr.Acquire("main", "test")
	require.NoError(t, err)
	return rec
}

func TestCreator_Create(t *testing.T) {
	repoPath := setupTestRepo(t)
	lockRec := acquireLock(t, repoPath)

	// Add some content to main/
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("hello"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "test note", model.ConsistencyQuiesced, lockRec.FencingToken)
	require.NoError(t, err)

	assert.NotEmpty(t, desc.SnapshotID)
	assert.Equal(t, "main", desc.WorktreeName)
	assert.Equal(t, "test note", desc.Note)
	assert.Equal(t, model.EngineCopy, desc.Engine)
	assert.Equal(t, model.ConsistencyQuiesced, desc.ConsistencyLevel)
	assert.NotEmpty(t, desc.PayloadRootHash)
	assert.NotEmpty(t, desc.DescriptorChecksum)
	assert.Equal(t, lockRec.FencingToken, desc.FencingToken)

	// Verify snapshot directory exists
	snapshotDir := filepath.Join(repoPath, ".jvs", "snapshots", string(desc.SnapshotID))
	assert.DirExists(t, snapshotDir)

	// Verify descriptor exists
	descriptorPath := filepath.Join(repoPath, ".jvs", "descriptors", string(desc.SnapshotID)+".json")
	assert.FileExists(t, descriptorPath)

	// Verify .READY marker exists
	readyPath := filepath.Join(snapshotDir, ".READY")
	assert.FileExists(t, readyPath)
}

func TestCreator_ReadyProtocol(t *testing.T) {
	repoPath := setupTestRepo(t)
	lockRec := acquireLock(t, repoPath)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "", model.ConsistencyQuiesced, lockRec.FencingToken)
	require.NoError(t, err)

	// Verify .READY contains correct info
	readyPath := filepath.Join(repoPath, ".jvs", "snapshots", string(desc.SnapshotID), ".READY")
	data, err := os.ReadFile(readyPath)
	require.NoError(t, err)

	var marker model.ReadyMarker
	require.NoError(t, json.Unmarshal(data, &marker))
	assert.Equal(t, desc.SnapshotID, marker.SnapshotID)
}

func TestCreator_UpdatesHead(t *testing.T) {
	repoPath := setupTestRepo(t)
	lockRec := acquireLock(t, repoPath)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("v1"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc1, err := creator.Create("main", "first", model.ConsistencyQuiesced, lockRec.FencingToken)
	require.NoError(t, err)

	// Check head updated
	cfg, err := repo.LoadWorktreeConfig(repoPath, "main")
	require.NoError(t, err)
	assert.Equal(t, desc1.SnapshotID, cfg.HeadSnapshotID)

	// Create second snapshot
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("v2"), 0644)
	desc2, err := creator.Create("main", "second", model.ConsistencyQuiesced, lockRec.FencingToken)
	require.NoError(t, err)

	// Parent should be first snapshot
	assert.Equal(t, desc1.SnapshotID, *desc2.ParentID)

	// Head should be second
	cfg, _ = repo.LoadWorktreeConfig(repoPath, "main")
	assert.Equal(t, desc2.SnapshotID, cfg.HeadSnapshotID)
}

func TestCreator_PayloadContentPreserved(t *testing.T) {
	repoPath := setupTestRepo(t)
	lockRec := acquireLock(t, repoPath)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("original"), 0644)
	os.MkdirAll(filepath.Join(mainPath, "subdir"), 0755)
	os.WriteFile(filepath.Join(mainPath, "subdir", "nested.txt"), []byte("nested"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "", model.ConsistencyQuiesced, lockRec.FencingToken)
	require.NoError(t, err)

	// Verify snapshot content
	snapshotPath := filepath.Join(repoPath, ".jvs", "snapshots", string(desc.SnapshotID))
	content, err := os.ReadFile(filepath.Join(snapshotPath, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "original", string(content))

	content, err = os.ReadFile(filepath.Join(snapshotPath, "subdir", "nested.txt"))
	require.NoError(t, err)
	assert.Equal(t, "nested", string(content))
}

func TestCreator_FencingValidation(t *testing.T) {
	repoPath := setupTestRepo(t)
	lockRec := acquireLock(t, repoPath)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	// Wrong fencing token should fail
	_, err := creator.Create("main", "", model.ConsistencyQuiesced, lockRec.FencingToken+1)
	assert.Error(t, err)
}
