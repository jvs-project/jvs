package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// SnapshotID is the unique identifier for a snapshot: <unix_ms>-<rand8hex>
type SnapshotID string

// NewSnapshotID generates a new unique snapshot ID.
func NewSnapshotID() SnapshotID {
	ts := time.Now().UnixMilli()
	var randBytes [4]byte
	if _, err := rand.Read(randBytes[:]); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return SnapshotID(fmt.Sprintf("%013d-%s", ts, hex.EncodeToString(randBytes[:])))
}

// ShortID returns the first 8 characters for display.
func (id SnapshotID) ShortID() string {
	s := string(id)
	if len(s) >= 8 {
		return s[:8]
	}
	return s
}

// String returns the full snapshot ID as string.
func (id SnapshotID) String() string {
	return string(id)
}

// Descriptor is the on-disk snapshot metadata (11 fields for v0.x).
type Descriptor struct {
	SnapshotID        SnapshotID       `json:"snapshot_id"`
	ParentID          *SnapshotID      `json:"parent_id,omitempty"`
	WorktreeName      string           `json:"worktree_name"`
	CreatedAt         time.Time        `json:"created_at"`
	Note              string           `json:"note,omitempty"`
	Engine            EngineType       `json:"engine"`
	ConsistencyLevel  ConsistencyLevel `json:"consistency_level"`
	PayloadRootHash   HashValue        `json:"payload_root_hash"`
	DescriptorChecksum HashValue       `json:"descriptor_checksum"`
	FencingToken      int64            `json:"fencing_token"`
	IntegrityState    IntegrityState   `json:"integrity_state"`
}

// ReadyMarker is the .READY file content indicating complete snapshot.
type ReadyMarker struct {
	SnapshotID   SnapshotID `json:"snapshot_id"`
	CompletedAt  time.Time  `json:"completed_at"`
	PayloadHash  HashValue  `json:"payload_root_hash"`
}

// IntentRecord tracks in-progress snapshot creation for crash recovery.
type IntentRecord struct {
	SnapshotID   SnapshotID `json:"snapshot_id"`
	WorktreeName string     `json:"worktree_name"`
	StartedAt    time.Time  `json:"started_at"`
	Engine       EngineType `json:"engine"`
}
