package model

import "time"

// Pin protects a snapshot from garbage collection.
type Pin struct {
	SnapshotID SnapshotID `json:"snapshot_id"`
	PinnedAt   time.Time  `json:"pinned_at"`
	Reason     string     `json:"reason,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// GCPlan is the output of gc plan phase.
type GCPlan struct {
	PlanID                 string          `json:"plan_id"`
	CreatedAt              time.Time       `json:"created_at"`
	ProtectedSet           []SnapshotID    `json:"protected_set"`
	ProtectedByPin         int             `json:"protected_by_pin"`
	ProtectedByLineage     int             `json:"protected_by_lineage"`
	CandidateCount         int             `json:"candidate_count"`
	ToDelete               []SnapshotID    `json:"to_delete"`
	DeletableBytesEstimate int64           `json:"deletable_bytes_estimate"`
	EstimatedBytes         int64           `json:"estimated_bytes"` // Legacy, same as deletable_bytes_estimate
	RetentionPolicy        RetentionPolicy `json:"retention_policy"`
}

// Tombstone marks a snapshot as deleted but not yet reclaimed.
type Tombstone struct {
	SnapshotID  SnapshotID `json:"snapshot_id"`
	DeletedAt   time.Time  `json:"deleted_at"`
	Reclaimable bool       `json:"reclaimable"`
}

// RetentionPolicy configures which snapshots to keep.
type RetentionPolicy struct {
	KeepMinSnapshots int           `json:"keep_min_snapshots"`
	KeepMinAge       time.Duration `json:"keep_min_age"`
	KeepAllWithin    time.Duration `json:"keep_all_within"`
}
