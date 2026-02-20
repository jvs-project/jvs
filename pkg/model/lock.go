package model

import "time"

// LockRecord is stored at .jvs/worktrees/<name>/lock.json
type LockRecord struct {
	WorktreeName string    `json:"worktree_name"`
	HolderNonce  string    `json:"holder_nonce"`
	SessionID    string    `json:"session_id"`
	AcquiredAt   time.Time `json:"acquired_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	FencingToken int64     `json:"fencing_token"`
	Purpose      string    `json:"purpose,omitempty"`
}

// IsExpired returns true if the lock has expired.
func (l *LockRecord) IsExpired(now time.Time) bool {
	return now.After(l.ExpiresAt)
}

// LockSession is persisted to .jvs/worktrees/<name>/.session for cross-CLI continuity.
type LockSession struct {
	SessionID   string `json:"session_id"`
	HolderNonce string `json:"holder_nonce"`
}

// LockPolicy configures lock timing parameters.
type LockPolicy struct {
	DefaultLeaseTTL time.Duration `json:"default_lease_ttl"`
	MaxLeaseTTL     time.Duration `json:"max_lease_ttl"`
	ClockSkewTolerance time.Duration `json:"clock_skew_tolerance"`
}
