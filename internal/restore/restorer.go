package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jvs-project/jvs/internal/audit"
	"github.com/jvs-project/jvs/internal/engine"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/fsutil"
	"github.com/jvs-project/jvs/pkg/metrics"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/jvs-project/jvs/pkg/uuidutil"
)

// Restorer handles snapshot restore operations.
type Restorer struct {
	repoRoot    string
	engineType  model.EngineType
	engine      engine.Engine
	auditLogger *audit.FileAppender
	recordMetrics bool
}

// NewRestorer creates a new restorer.
func NewRestorer(repoRoot string, engineType model.EngineType) *Restorer {
	eng := engine.NewEngine(engineType)

	auditPath := filepath.Join(repoRoot, ".jvs", "audit", "audit.jsonl")
	return &Restorer{
		repoRoot:       repoRoot,
		engineType:     engineType,
		engine:         eng,
		auditLogger:    audit.NewFileAppender(auditPath),
		recordMetrics:  metrics.Enabled(),
	}
}

// Restore replaces the content of a worktree with a snapshot.
// This puts the worktree into a "detached" state (unless restoring to HEAD).
// The worktree is specified by name, not derived from the snapshot.
func (r *Restorer) Restore(worktreeName string, snapshotID model.SnapshotID) error {
	startTime := time.Now()
	err := r.restore(worktreeName, snapshotID)

	// Record metrics if enabled
	if r.recordMetrics {
		metrics.Default().RecordRestore(err == nil, time.Since(startTime))
	}

	return err
}

// restore performs the actual restore operation.
func (r *Restorer) restore(worktreeName string, snapshotID model.SnapshotID) error {
	// Load and verify snapshot
	_, err := snapshot.LoadDescriptor(r.repoRoot, snapshotID)
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}

	if err := snapshot.VerifySnapshot(r.repoRoot, snapshotID, false); err != nil {
		return fmt.Errorf("verify snapshot: %w", err)
	}

	// Get worktree info
	wtMgr := worktree.NewManager(r.repoRoot)
	cfg, err := wtMgr.Get(worktreeName)
	if err != nil {
		return fmt.Errorf("get worktree: %w", err)
	}

	payloadPath := wtMgr.Path(worktreeName)

	// Create backup directory for atomic swap
	backupPath := payloadPath + ".restore-backup-" + uuidutil.NewV4()[:8]
	snapshotDir := filepath.Join(r.repoRoot, ".jvs", "snapshots", string(snapshotID))
	tempPath := payloadPath + ".restore-tmp-" + uuidutil.NewV4()[:8]

	// Step 1: Clone snapshot to temp location
	if _, err := r.engine.Clone(snapshotDir, tempPath); err != nil {
		return fmt.Errorf("clone to temp: %w", err)
	}

	// Step 2: Atomic swap: rename current to backup, temp to payload
	if err := fsutil.RenameAndSync(payloadPath, backupPath); err != nil {
		os.RemoveAll(tempPath)
		return fmt.Errorf("backup current: %w", err)
	}

	if err := fsutil.RenameAndSync(tempPath, payloadPath); err != nil {
		// Try to rollback
		fsutil.RenameAndSync(backupPath, payloadPath)
		return fmt.Errorf("swap in restored: %w", err)
	}

	// Step 3: Cleanup backup synchronously with error logging
	if err := os.RemoveAll(backupPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to cleanup backup %s: %v\n", backupPath, err)
	}

	// Step 4: Update head (NOT latest - this puts worktree in detached state)
	if err := wtMgr.UpdateHead(worktreeName, snapshotID); err != nil {
		// Don't fail, head update is secondary
		fmt.Fprintf(os.Stderr, "warning: failed to update head: %v\n", err)
	}

	// Determine if we're now detached
	isDetached := snapshotID != cfg.LatestSnapshotID

	// Audit log
	r.auditLogger.Append(model.EventTypeRestore, worktreeName, snapshotID, map[string]any{
		"detached": isDetached,
	})

	return nil
}

// RestoreToLatest restores a worktree to its latest snapshot (exits detached state).
func (r *Restorer) RestoreToLatest(worktreeName string) error {
	wtMgr := worktree.NewManager(r.repoRoot)
	cfg, err := wtMgr.Get(worktreeName)
	if err != nil {
		return fmt.Errorf("get worktree: %w", err)
	}

	if cfg.LatestSnapshotID == "" {
		return fmt.Errorf("worktree has no snapshots")
	}

	return r.Restore(worktreeName, cfg.LatestSnapshotID)
}
