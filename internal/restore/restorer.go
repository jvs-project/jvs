package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jvs-project/jvs/internal/audit"
	"github.com/jvs-project/jvs/internal/engine"
	"github.com/jvs-project/jvs/internal/lock"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/fsutil"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/jvs-project/jvs/pkg/uuidutil"
)

// Restorer handles snapshot restore operations.
type Restorer struct {
	repoRoot    string
	engineType  model.EngineType
	engine      engine.Engine
	auditLogger *audit.FileAppender
}

// NewRestorer creates a new restorer.
func NewRestorer(repoRoot string, engineType model.EngineType) *Restorer {
	var eng engine.Engine
	if engineType == model.EngineCopy {
		eng = engine.NewCopyEngine()
	} else {
		eng = engine.NewCopyEngine() // fallback
	}

	auditPath := filepath.Join(repoRoot, ".jvs", "audit", "audit.jsonl")
	return &Restorer{
		repoRoot:    repoRoot,
		engineType:  engineType,
		engine:      eng,
		auditLogger: audit.NewFileAppender(auditPath),
	}
}

// SafeRestore creates a new worktree from a snapshot.
// This is the default, safe restore operation.
func (r *Restorer) SafeRestore(snapshotID model.SnapshotID, name string, parentSnapshotID *model.SnapshotID) (*model.WorktreeConfig, error) {
	// Load and verify snapshot
	_, err := snapshot.LoadDescriptor(r.repoRoot, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("load snapshot: %w", err)
	}

	// Verify snapshot integrity
	if err := snapshot.VerifySnapshot(r.repoRoot, snapshotID, false); err != nil {
		return nil, fmt.Errorf("verify snapshot: %w", err)
	}

	// Generate name if not provided
	if name == "" {
		name = fmt.Sprintf("restore-%s", snapshotID.ShortID())
	}

	// Create worktree
	wtMgr := worktree.NewManager(r.repoRoot)
	cfg, err := wtMgr.Create(name, &snapshotID)
	if err != nil {
		return nil, fmt.Errorf("create worktree: %w", err)
	}

	// Clone snapshot to worktree
	snapshotDir := filepath.Join(r.repoRoot, ".jvs", "snapshots", string(snapshotID))
	payloadPath := wtMgr.Path(name)

	if _, err := r.engine.Clone(snapshotDir, payloadPath); err != nil {
		wtMgr.Remove(name)
		return nil, fmt.Errorf("clone snapshot: %w", err)
	}

	// Audit log
	r.auditLogger.Append(model.EventTypeRestore, name, snapshotID, map[string]any{
		"mode": "safe",
	})

	return cfg, nil
}

// InplaceRestore replaces the content of an existing worktree with a snapshot.
// This is dangerous and requires lock + fencing token + reason.
func (r *Restorer) InplaceRestore(worktreeName string, snapshotID model.SnapshotID, fencingToken int64, reason string) error {
	// Validate fencing
	lockMgr := lock.NewManager(r.repoRoot, model.LockPolicy{})
	if err := lockMgr.ValidateFencing(worktreeName, fencingToken); err != nil {
		return err
	}

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
	payloadPath := wtMgr.Path(worktreeName)

	// Create backup directory for atomic swap
	backupPath := payloadPath + ".restore-backup-" + uuidutil.NewV4()[:8]
	snapshotDir := filepath.Join(r.repoRoot, ".jvs", "snapshots", string(snapshotID))
	tempPath := payloadPath + ".restore-tmp-" + uuidutil.NewV4()[:8]

	// Step 1: Clone snapshot to temp location
	if _, err := r.engine.Clone(snapshotDir, tempPath); err != nil {
		return fmt.Errorf("clone to temp: %w", err)
	}

	// Step 2: Re-validate fencing before swap
	if err := lockMgr.ValidateFencing(worktreeName, fencingToken); err != nil {
		os.RemoveAll(tempPath)
		return err
	}

	// Step 3: Atomic swap: rename current to backup, temp to payload
	if err := fsutil.RenameAndSync(payloadPath, backupPath); err != nil {
		os.RemoveAll(tempPath)
		return fmt.Errorf("backup current: %w", err)
	}

	if err := fsutil.RenameAndSync(tempPath, payloadPath); err != nil {
		// Try to rollback
		fsutil.RenameAndSync(backupPath, payloadPath)
		return fmt.Errorf("swap in restored: %w", err)
	}

	// Step 4: Cleanup backup (async, non-blocking)
	go func() {
		time.Sleep(1 * time.Second)
		os.RemoveAll(backupPath)
	}()

	// Step 5: Update head
	if err := wtMgr.UpdateHead(worktreeName, snapshotID); err != nil {
		// Don't fail, head update is secondary
		fmt.Fprintf(os.Stderr, "warning: failed to update head: %v\n", err)
	}

	// Audit log
	r.auditLogger.Append(model.EventTypeRestore, worktreeName, snapshotID, map[string]any{
		"mode":   "inplace",
		"reason": reason,
	})

	return nil
}
