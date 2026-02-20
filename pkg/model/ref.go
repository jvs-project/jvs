package model

import "time"

// RefRecord is stored at .jvs/refs/<name>.json
type RefRecord struct {
	Name        string     `json:"name"`
	TargetID    SnapshotID `json:"target_id"`
	CreatedAt   time.Time  `json:"created_at"`
	Description string     `json:"description,omitempty"`
}
