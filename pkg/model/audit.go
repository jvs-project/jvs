package model

import "time"

// AuditEventType identifies the type of auditable event.
type AuditEventType string

const (
	EventTypeSnapshotCreate AuditEventType = "snapshot_create"
	EventTypeSnapshotDelete AuditEventType = "snapshot_delete"
	EventTypeRestore        AuditEventType = "restore"
	EventTypeWorktreeCreate AuditEventType = "worktree_create"
	EventTypeWorktreeRename AuditEventType = "worktree_rename"
	EventTypeWorktreeRemove AuditEventType = "worktree_remove"
	EventTypeGCPlan         AuditEventType = "gc_plan"
	EventTypeGCRun          AuditEventType = "gc_run"
)

// AuditRecord is a single line in the audit log (JSONL format).
type AuditRecord struct {
	Timestamp    time.Time      `json:"timestamp"`
	EventType    AuditEventType `json:"event_type"`
	SnapshotID   SnapshotID     `json:"snapshot_id,omitempty"`
	WorktreeName string         `json:"worktree_name,omitempty"`
	Details      map[string]any `json:"details,omitempty"`
	PrevHash     HashValue      `json:"prev_hash"`
	RecordHash   HashValue      `json:"record_hash"`
}
