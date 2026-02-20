package verify_test

import (
	"testing"
	"time"

	"github.com/jvs-project/jvs/internal/lock"
	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/internal/verify"
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

func TestVerifier_VerifySnapshot(t *testing.T) {
	repoPath := setupTestRepo(t)
	snapshotID := createTestSnapshot(t, repoPath)

	v := verify.NewVerifier(repoPath)
	result, err := v.VerifySnapshot(snapshotID, true)
	require.NoError(t, err)
	t.Logf("Result: %+v", result)
	assert.True(t, result.ChecksumValid, "checksum should be valid")
	assert.True(t, result.PayloadHashValid, "payload hash should be valid")
	assert.False(t, result.TamperDetected, "no tamper should be detected")
}

func TestVerifier_VerifyAll(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create two snapshots with the same lock
	mgr := lock.NewManager(repoPath, model.LockPolicy{DefaultLeaseTTL: time.Hour})
	rec, err := mgr.Acquire("main", "test")
	require.NoError(t, err)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	_, err = creator.Create("main", "first", model.ConsistencyQuiesced, rec.FencingToken)
	require.NoError(t, err)
	_, err = creator.Create("main", "second", model.ConsistencyQuiesced, rec.FencingToken)
	require.NoError(t, err)

	mgr.Release("main", rec.HolderNonce)

	v := verify.NewVerifier(repoPath)
	results, err := v.VerifyAll(false)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	for _, r := range results {
		assert.True(t, r.ChecksumValid)
	}
}

func TestVerifier_VerifySnapshot_Nonexistent(t *testing.T) {
	repoPath := setupTestRepo(t)

	v := verify.NewVerifier(repoPath)
	result, err := v.VerifySnapshot("nonexistent-snapshot-id", false)
	require.NoError(t, err)
	assert.True(t, result.TamperDetected)
	assert.Equal(t, "critical", result.Severity)
	assert.NotEmpty(t, result.Error)
}

func TestVerifier_VerifyAll_Empty(t *testing.T) {
	repoPath := setupTestRepo(t)

	v := verify.NewVerifier(repoPath)
	results, err := v.VerifyAll(false)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestVerifier_VerifySnapshot_NoPayloadHash(t *testing.T) {
	repoPath := setupTestRepo(t)
	snapshotID := createTestSnapshot(t, repoPath)

	v := verify.NewVerifier(repoPath)
	result, err := v.VerifySnapshot(snapshotID, false)
	require.NoError(t, err)
	assert.True(t, result.ChecksumValid)
	assert.False(t, result.PayloadHashValid) // Not verified when verifyPayloadHash=false
	assert.False(t, result.TamperDetected)
}
