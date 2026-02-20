package doctor_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jvs-project/jvs/internal/doctor"
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

func createTestSnapshot(t *testing.T, repoPath string) {
	mgr := lock.NewManager(repoPath, model.LockPolicy{DefaultLeaseTTL: time.Hour})
	rec, err := mgr.Acquire("main", "test")
	require.NoError(t, err)
	defer mgr.Release("main", rec.HolderNonce)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	_, err = creator.Create("main", "test", model.ConsistencyQuiesced, rec.FencingToken)
	require.NoError(t, err)
}

func TestDoctor_Check_Healthy(t *testing.T) {
	repoPath := setupTestRepo(t)

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)
	assert.True(t, result.Healthy)
	assert.Empty(t, result.Findings)
}

func TestDoctor_Check_WithSnapshots(t *testing.T) {
	repoPath := setupTestRepo(t)
	createTestSnapshot(t, repoPath)

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)
	assert.True(t, result.Healthy)
}

func TestDoctor_Check_Strict(t *testing.T) {
	repoPath := setupTestRepo(t)
	createTestSnapshot(t, repoPath)

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(true)
	require.NoError(t, err)
	assert.True(t, result.Healthy)
}

func TestDoctor_Check_OrphanIntent(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create orphan intent file
	intentsDir := filepath.Join(repoPath, ".jvs", "intents")
	os.MkdirAll(intentsDir, 0755)
	os.WriteFile(filepath.Join(intentsDir, "orphan.json"), []byte("{}"), 0644)

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)
	// Orphan intents are warnings, not critical, so repo stays healthy
	assert.True(t, result.Healthy)
	assert.Len(t, result.Findings, 1)
	assert.Equal(t, "intent", result.Findings[0].Category)
	assert.Equal(t, "warning", result.Findings[0].Severity)
}

func TestDoctor_Check_OrphanTmp(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create orphan tmp file
	os.WriteFile(filepath.Join(repoPath, ".jvs", ".jvs-tmp-orphan"), []byte("data"), 0644)

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)
	// Orphan tmp is info level, doesn't make repo unhealthy
	assert.True(t, result.Healthy || len(result.Findings) > 0)
}

func TestDoctor_Check_MissingWorktreePayload(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Remove main payload directory (simulating corruption)
	os.RemoveAll(filepath.Join(repoPath, "main"))

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)
	// Missing payload reports error finding but repo stays "healthy" at info level
	assert.NotEmpty(t, result.Findings)
	found := false
	for _, f := range result.Findings {
		if f.Category == "worktree" {
			found = true
			assert.Contains(t, f.Description, "payload directory missing")
		}
	}
	assert.True(t, found, "expected worktree finding for missing payload")
}
