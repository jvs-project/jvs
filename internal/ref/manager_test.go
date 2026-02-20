package ref_test

import (
	"testing"
	"time"

	"github.com/jvs-project/jvs/internal/lock"
	"github.com/jvs-project/jvs/internal/ref"
	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/internal/snapshot"
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

func createTestSnapshot(t *testing.T, repoPath string) model.SnapshotID {
	mgr := lock.NewManager(repoPath, model.LockPolicy{DefaultLeaseTTL: time.Hour})
	rec, err := mgr.Acquire("main", "test")
	require.NoError(t, err)
	defer mgr.Release("main", rec.HolderNonce)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "test", model.ConsistencyQuiesced, rec.FencingToken)
	require.NoError(t, err)
	return desc.SnapshotID
}

func TestManager_Create(t *testing.T) {
	repoPath := setupTestRepo(t)
	snapshotID := createTestSnapshot(t, repoPath)

	mgr := ref.NewManager(repoPath)
	rec, err := mgr.Create("v1.0", snapshotID, "First release")
	require.NoError(t, err)
	assert.Equal(t, "v1.0", rec.Name)
	assert.Equal(t, snapshotID, rec.TargetID)
	assert.Equal(t, "First release", rec.Description)
}

func TestManager_Create_InvalidName(t *testing.T) {
	repoPath := setupTestRepo(t)
	snapshotID := createTestSnapshot(t, repoPath)

	mgr := ref.NewManager(repoPath)
	_, err := mgr.Create("../evil", snapshotID, "")
	require.ErrorIs(t, err, errclass.ErrNameInvalid)
}

func TestManager_Create_Duplicate(t *testing.T) {
	repoPath := setupTestRepo(t)
	snapshotID := createTestSnapshot(t, repoPath)

	mgr := ref.NewManager(repoPath)
	mgr.Create("v1.0", snapshotID, "")
	_, err := mgr.Create("v1.0", snapshotID, "")
	assert.Error(t, err)
}

func TestManager_List(t *testing.T) {
	repoPath := setupTestRepo(t)
	snapshotID := createTestSnapshot(t, repoPath)

	mgr := ref.NewManager(repoPath)
	mgr.Create("v1.0", snapshotID, "")
	mgr.Create("v1.1", snapshotID, "")

	list, err := mgr.List()
	require.NoError(t, err)
	assert.Len(t, list, 2)

	names := make(map[string]bool)
	for _, r := range list {
		names[r.Name] = true
	}
	assert.True(t, names["v1.0"])
	assert.True(t, names["v1.1"])
}

func TestManager_Get(t *testing.T) {
	repoPath := setupTestRepo(t)
	snapshotID := createTestSnapshot(t, repoPath)

	mgr := ref.NewManager(repoPath)
	mgr.Create("v1.0", snapshotID, "test")

	rec, err := mgr.Get("v1.0")
	require.NoError(t, err)
	assert.Equal(t, snapshotID, rec.TargetID)
}

func TestManager_Get_NotFound(t *testing.T) {
	repoPath := setupTestRepo(t)
	mgr := ref.NewManager(repoPath)

	_, err := mgr.Get("nonexistent")
	assert.Error(t, err)
}

func TestManager_Delete(t *testing.T) {
	repoPath := setupTestRepo(t)
	snapshotID := createTestSnapshot(t, repoPath)

	mgr := ref.NewManager(repoPath)
	mgr.Create("v1.0", snapshotID, "")

	err := mgr.Delete("v1.0")
	require.NoError(t, err)

	_, err = mgr.Get("v1.0")
	assert.Error(t, err)
}
