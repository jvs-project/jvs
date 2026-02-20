package gc_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jvs-project/jvs/internal/gc"
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

func createTestSnapshot(t *testing.T, repoPath string) model.SnapshotID {
	mgr := lock.NewManager(repoPath, model.LockPolicy{DefaultLeaseTTL: time.Hour})
	rec, err := mgr.Acquire("main", "test")
	require.NoError(t, err)
	defer mgr.Release("main", rec.HolderNonce)

	// Add some content
	mainPath := filepath.Join(repoPath, "main")
	require.NoError(t, os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644))

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "test", model.ConsistencyQuiesced, rec.FencingToken)
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
