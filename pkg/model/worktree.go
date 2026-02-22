package model

import "time"

// WorktreeConfig is stored at .jvs/worktrees/<name>/config.json
type WorktreeConfig struct {
	Name             string     `json:"name"`
	BaseSnapshotID   SnapshotID `json:"base_snapshot_id,omitempty"`   // Immutable snapshot worktree was created from
	HeadSnapshotID   SnapshotID `json:"head_snapshot_id,omitempty"`   // Current position (may differ from latest if detached)
	LatestSnapshotID SnapshotID `json:"latest_snapshot_id,omitempty"` // The most recent snapshot in this worktree's lineage
	CreatedAt        time.Time  `json:"created_at"`
}

// IsDetached returns true if the worktree is at a historical snapshot (not at HEAD).
// A worktree is in "detached" state when HeadSnapshotID differs from LatestSnapshotID.
func (c *WorktreeConfig) IsDetached() bool {
	if c.LatestSnapshotID == "" {
		// No snapshots yet, not detached
		return false
	}
	return c.HeadSnapshotID != c.LatestSnapshotID
}

// CanSnapshot returns true if the worktree can create new snapshots.
// Only worktrees at HEAD (not detached) can create snapshots.
func (c *WorktreeConfig) CanSnapshot() bool {
	return !c.IsDetached() && c.LatestSnapshotID != ""
}
