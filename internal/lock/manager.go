package lock

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jvs-project/jvs/pkg/errclass"
	"github.com/jvs-project/jvs/pkg/fsutil"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/jvs-project/jvs/pkg/uuidutil"
)

// Manager handles SWMR lock operations.
type Manager struct {
	repoRoot string
	policy   model.LockPolicy
	mu       sync.Mutex
}

// NewManager creates a new lock manager.
func NewManager(repoRoot string, policy model.LockPolicy) *Manager {
	return &Manager{
		repoRoot: repoRoot,
		policy:   policy,
	}
}

// Acquire attempts to acquire an exclusive lock on the worktree.
func (m *Manager) Acquire(worktreeName, purpose string) (*model.LockRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	lockPath := m.lockPath(worktreeName)
	dir := filepath.Dir(lockPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create lock dir: %w", err)
	}

	// Try O_CREAT|O_EXCL for atomic acquire
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			// Lock exists, check if expired
			rec, readErr := m.readLock(lockPath)
			if readErr != nil {
				return nil, fmt.Errorf("read existing lock: %w", readErr)
			}
			if rec.IsExpired(time.Now()) {
				// Expired, but O_EXCL failed - use steal path
				return nil, errclass.ErrLockConflict.WithMessage("lock exists but expired, use steal")
			}
			return nil, errclass.ErrLockConflict.WithMessagef("worktree %s is locked", worktreeName)
		}
		return nil, fmt.Errorf("create lock: %w", err)
	}
	defer file.Close()

	now := time.Now().UTC()
	rec := &model.LockRecord{
		WorktreeName: worktreeName,
		HolderNonce:  uuidutil.NewV4(),
		SessionID:    uuidutil.NewV4(),
		AcquiredAt:   now,
		ExpiresAt:    now.Add(m.policy.DefaultLeaseTTL),
		FencingToken: 1,
		Purpose:      purpose,
	}

	if err := m.writeLock(file, rec); err != nil {
		os.Remove(lockPath)
		return nil, err
	}

	// Write session file for cross-CLI continuity
	if err := m.writeSession(worktreeName, rec); err != nil {
		os.Remove(lockPath)
		return nil, fmt.Errorf("write session: %w", err)
	}

	return rec, nil
}

// Renew extends the lease on an existing lock.
func (m *Manager) Renew(worktreeName, holderNonce string) (*model.LockRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	lockPath := m.lockPath(worktreeName)
	rec, err := m.readLock(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errclass.ErrLockNotHeld.WithMessage("no lock held")
		}
		return nil, fmt.Errorf("read lock: %w", err)
	}

	// Check if already expired
	if rec.IsExpired(time.Now()) {
		return nil, errclass.ErrLockExpired.WithMessage("lock has expired")
	}

	// Verify nonce matches
	if rec.HolderNonce != holderNonce {
		return nil, errclass.ErrLockNotHeld.WithMessage("nonce mismatch")
	}

	// Extend lease
	rec.ExpiresAt = time.Now().UTC().Add(m.policy.DefaultLeaseTTL)

	// Atomic write via temp file
	if err := m.updateLock(lockPath, rec); err != nil {
		return nil, fmt.Errorf("update lock: %w", err)
	}

	return rec, nil
}

// Steal acquires the lock after the previous holder's lease expired.
func (m *Manager) Steal(worktreeName, purpose string) (*model.LockRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	lockPath := m.lockPath(worktreeName)

	rec, err := m.readLock(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No lock exists, use regular acquire
			m.mu.Unlock()
			return m.Acquire(worktreeName, purpose)
		}
		return nil, fmt.Errorf("read lock: %w", err)
	}

	// Must be expired
	if !rec.IsExpired(time.Now()) {
		return nil, errclass.ErrLockConflict.WithMessage("lock not expired yet")
	}

	// Increment fencing token and take over
	now := time.Now().UTC()
	newRec := &model.LockRecord{
		WorktreeName: worktreeName,
		HolderNonce:  uuidutil.NewV4(),
		SessionID:    uuidutil.NewV4(),
		AcquiredAt:   now,
		ExpiresAt:    now.Add(m.policy.DefaultLeaseTTL),
		FencingToken: rec.FencingToken + 1,
		Purpose:      purpose,
	}

	if err := m.updateLock(lockPath, newRec); err != nil {
		return nil, fmt.Errorf("steal lock: %w", err)
	}

	if err := m.writeSession(worktreeName, newRec); err != nil {
		return nil, fmt.Errorf("write session: %w", err)
	}

	return newRec, nil
}

// Release frees the lock.
func (m *Manager) Release(worktreeName, holderNonce string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	lockPath := m.lockPath(worktreeName)
	rec, err := m.readLock(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // already released
		}
		return fmt.Errorf("read lock: %w", err)
	}

	if rec.HolderNonce != holderNonce {
		return errclass.ErrLockNotHeld.WithMessage("cannot release: nonce mismatch")
	}

	if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove lock: %w", err)
	}

	// Remove session file
	sessionPath := m.sessionPath(worktreeName)
	os.Remove(sessionPath)

	return nil
}

// ValidateFencing checks if the provided fencing token matches the current lock.
func (m *Manager) ValidateFencing(worktreeName string, token int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	lockPath := m.lockPath(worktreeName)
	rec, err := m.readLock(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errclass.ErrLockNotHeld.WithMessage("no lock held")
		}
		return fmt.Errorf("read lock: %w", err)
	}

	if rec.FencingToken != token {
		return errclass.ErrFencingMismatch.WithMessagef(
			"expected token %d, got %d", rec.FencingToken, token)
	}

	return nil
}

// Status returns the current lock state.
func (m *Manager) Status(worktreeName string) (model.LockState, *model.LockRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	lockPath := m.lockPath(worktreeName)
	rec, err := m.readLock(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return model.LockStateFree, nil, nil
		}
		return model.LockStateFree, nil, fmt.Errorf("read lock: %w", err)
	}

	if rec.IsExpired(time.Now()) {
		return model.LockStateExpired, rec, nil
	}
	return model.LockStateHeld, rec, nil
}

// LoadSession loads the session file for cross-CLI continuity.
func (m *Manager) LoadSession(worktreeName string) (*model.LockSession, error) {
	sessionPath := m.sessionPath(worktreeName)
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, fmt.Errorf("read session: %w", err)
	}
	var sess model.LockSession
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("parse session: %w", err)
	}
	return &sess, nil
}

func (m *Manager) lockPath(worktreeName string) string {
	return filepath.Join(m.repoRoot, ".jvs", "worktrees", worktreeName, "lock.json")
}

func (m *Manager) sessionPath(worktreeName string) string {
	return filepath.Join(m.repoRoot, ".jvs", "worktrees", worktreeName, ".session")
}

func (m *Manager) readLock(path string) (*model.LockRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rec model.LockRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("parse lock: %w", err)
	}
	return &rec, nil
}

func (m *Manager) writeLock(file *os.File, rec *model.LockRecord) error {
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal lock: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write lock: %w", err)
	}
	return file.Sync()
}

func (m *Manager) updateLock(path string, rec *model.LockRecord) error {
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal lock: %w", err)
	}
	return fsutil.AtomicWrite(path, data, 0644)
}

func (m *Manager) writeSession(worktreeName string, rec *model.LockRecord) error {
	sessionPath := m.sessionPath(worktreeName)
	sess := &model.LockSession{
		SessionID:   rec.SessionID,
		HolderNonce: rec.HolderNonce,
	}
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	return fsutil.AtomicWrite(sessionPath, data, 0644)
}
