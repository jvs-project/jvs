package lock_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jvs-project/jvs/internal/lock"
	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/pkg/errclass"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func shortPolicy() model.LockPolicy {
	return model.LockPolicy{
		DefaultLeaseTTL:    100 * time.Millisecond,
		MaxLeaseTTL:        500 * time.Millisecond,
		ClockSkewTolerance: 50 * time.Millisecond,
	}
}

func setupRepo(t *testing.T) string {
	dir := t.TempDir()
	_, err := repo.Init(dir, "test")
	require.NoError(t, err)
	return dir
}

func TestManager_Acquire(t *testing.T) {
	repoPath := setupRepo(t)
	mgr := lock.NewManager(repoPath, shortPolicy())

	rec, err := mgr.Acquire("main", "test-purpose")
	require.NoError(t, err)
	assert.NotEmpty(t, rec.HolderNonce)
	assert.NotEmpty(t, rec.SessionID)
	assert.Equal(t, "main", rec.WorktreeName)
	assert.Equal(t, int64(1), rec.FencingToken)
}

func TestManager_Acquire_Conflict(t *testing.T) {
	repoPath := setupRepo(t)
	mgr := lock.NewManager(repoPath, shortPolicy())

	_, err := mgr.Acquire("main", "first")
	require.NoError(t, err)

	_, err = mgr.Acquire("main", "second")
	require.ErrorIs(t, err, errclass.ErrLockConflict)
}

func TestManager_Renew(t *testing.T) {
	repoPath := setupRepo(t)
	mgr := lock.NewManager(repoPath, shortPolicy())

	rec, _ := mgr.Acquire("main", "test")
	time.Sleep(20 * time.Millisecond) // let some time pass

	renewed, err := mgr.Renew("main", rec.HolderNonce)
	require.NoError(t, err)
	assert.True(t, renewed.ExpiresAt.After(rec.ExpiresAt))
}

func TestManager_Renew_Expired(t *testing.T) {
	repoPath := setupRepo(t)
	policy := shortPolicy()
	policy.DefaultLeaseTTL = 50 * time.Millisecond
	mgr := lock.NewManager(repoPath, policy)

	rec, _ := mgr.Acquire("main", "test")
	time.Sleep(100 * time.Millisecond) // wait for expiry

	_, err := mgr.Renew("main", rec.HolderNonce)
	require.ErrorIs(t, err, errclass.ErrLockExpired)
}

func TestManager_Renew_WrongNonce(t *testing.T) {
	repoPath := setupRepo(t)
	mgr := lock.NewManager(repoPath, shortPolicy())

	mgr.Acquire("main", "test")

	_, err := mgr.Renew("main", "wrong-nonce")
	require.ErrorIs(t, err, errclass.ErrLockNotHeld)
}

func TestManager_Release(t *testing.T) {
	repoPath := setupRepo(t)
	mgr := lock.NewManager(repoPath, shortPolicy())

	rec, _ := mgr.Acquire("main", "test")
	err := mgr.Release("main", rec.HolderNonce)
	require.NoError(t, err)

	// Should be able to acquire again
	_, err = mgr.Acquire("main", "second")
	require.NoError(t, err)
}

func TestManager_Steal(t *testing.T) {
	repoPath := setupRepo(t)
	policy := shortPolicy()
	policy.DefaultLeaseTTL = 50 * time.Millisecond
	mgr := lock.NewManager(repoPath, policy)

	rec1, _ := mgr.Acquire("main", "first")
	assert.Equal(t, int64(1), rec1.FencingToken)

	time.Sleep(100 * time.Millisecond) // wait for expiry

	rec2, err := mgr.Steal("main", "second")
	require.NoError(t, err)
	assert.Equal(t, int64(2), rec2.FencingToken) // fencing token incremented
	assert.NotEqual(t, rec1.HolderNonce, rec2.HolderNonce)
}

func TestManager_Steal_NotExpired(t *testing.T) {
	repoPath := setupRepo(t)
	mgr := lock.NewManager(repoPath, shortPolicy())

	mgr.Acquire("main", "first")

	_, err := mgr.Steal("main", "second")
	require.ErrorIs(t, err, errclass.ErrLockConflict)
}

func TestManager_ValidateFencing(t *testing.T) {
	repoPath := setupRepo(t)
	mgr := lock.NewManager(repoPath, shortPolicy())

	rec, _ := mgr.Acquire("main", "test")

	err := mgr.ValidateFencing("main", rec.FencingToken)
	require.NoError(t, err)
}

func TestManager_ValidateFencing_Mismatch(t *testing.T) {
	repoPath := setupRepo(t)
	mgr := lock.NewManager(repoPath, shortPolicy())

	rec, _ := mgr.Acquire("main", "test")

	err := mgr.ValidateFencing("main", rec.FencingToken+1)
	require.ErrorIs(t, err, errclass.ErrFencingMismatch)
}

func TestManager_Status(t *testing.T) {
	repoPath := setupRepo(t)
	mgr := lock.NewManager(repoPath, shortPolicy())

	// No lock
	state, rec, err := mgr.Status("main")
	require.NoError(t, err)
	assert.Equal(t, model.LockStateFree, state)

	// With lock
	acquired, _ := mgr.Acquire("main", "test")
	state, rec, err = mgr.Status("main")
	require.NoError(t, err)
	assert.Equal(t, model.LockStateHeld, state)
	assert.Equal(t, acquired.HolderNonce, rec.HolderNonce)
}

func TestManager_SessionFile(t *testing.T) {
	repoPath := setupRepo(t)
	mgr := lock.NewManager(repoPath, shortPolicy())

	rec, _ := mgr.Acquire("main", "test")

	// Session file should exist
	sessionPath := filepath.Join(repoPath, ".jvs", "worktrees", "main", ".session")
	_, err := os.Stat(sessionPath)
	require.NoError(t, err)

	// Load session
	sess, err := mgr.LoadSession("main")
	require.NoError(t, err)
	assert.Equal(t, rec.SessionID, sess.SessionID)
	assert.Equal(t, rec.HolderNonce, sess.HolderNonce)
}
