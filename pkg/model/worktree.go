package model

import "time"

// WorktreeConfig is stored at .jvs/worktrees/<name>/config.json
type WorktreeConfig struct {
	Name           string    `json:"name"`
	HeadSnapshotID SnapshotID `json:"head_snapshot_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	Isolation      string    `json:"isolation"` // v0.x always "exclusive"
}
