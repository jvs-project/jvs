package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jvs-project/jvs/internal/audit"
	"github.com/jvs-project/jvs/internal/engine"
	"github.com/jvs-project/jvs/internal/integrity"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/errclass"
	"github.com/jvs-project/jvs/pkg/fsutil"
	"github.com/jvs-project/jvs/pkg/model"
)

// Creator handles snapshot creation.
type Creator struct {
	repoRoot    string
	engineType  model.EngineType
	engine      engine.Engine
	auditLogger *audit.FileAppender
}

// NewCreator creates a new snapshot creator.
func NewCreator(repoRoot string, engineType model.EngineType) *Creator {
	eng := engine.NewEngine(engineType)

	auditPath := filepath.Join(repoRoot, ".jvs", "audit", "audit.jsonl")
	return &Creator{
		repoRoot:    repoRoot,
		engineType:  engineType,
		engine:      eng,
		auditLogger: audit.NewFileAppender(auditPath),
	}
}

// Create performs a full snapshot of the worktree using the 12-step protocol.
func (c *Creator) Create(worktreeName, note string, tags []string) (*model.Descriptor, error) {
	// Step 1: Validate worktree exists
	wtMgr := worktree.NewManager(c.repoRoot)
	cfg, err := wtMgr.Get(worktreeName)
	if err != nil {
		return nil, fmt.Errorf("get worktree: %w", err)
	}

	// Step 2: Generate snapshot ID
	snapshotID := model.NewSnapshotID()

	// Step 3: Create intent record (for crash recovery)
	intentPath := filepath.Join(c.repoRoot, ".jvs", "intents", string(snapshotID)+".json")
	intent := &model.IntentRecord{
		SnapshotID:   snapshotID,
		WorktreeName: worktreeName,
		StartedAt:    time.Now().UTC(),
		Engine:       c.engineType,
	}
	if err := c.writeIntent(intentPath, intent); err != nil {
		return nil, fmt.Errorf("write intent: %w", err)
	}
	defer os.Remove(intentPath) // cleanup on success

	// Step 4: Create snapshot directory
	snapshotDir := filepath.Join(c.repoRoot, ".jvs", "snapshots", string(snapshotID))
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return nil, fmt.Errorf("create snapshot dir: %w", err)
	}

	// Step 5: Clone payload to snapshot directory
	payloadPath := wtMgr.Path(worktreeName)
	if _, err := c.engine.Clone(payloadPath, snapshotDir); err != nil {
		os.RemoveAll(snapshotDir)
		return nil, fmt.Errorf("clone payload: %w", err)
	}

	// Step 6: Compute payload root hash
	payloadHash, err := integrity.ComputePayloadRootHash(snapshotDir)
	if err != nil {
		os.RemoveAll(snapshotDir)
		return nil, fmt.Errorf("compute payload hash: %w", err)
	}

	// Step 7: Create descriptor
	var parentID *model.SnapshotID
	if cfg.HeadSnapshotID != "" {
		pid := cfg.HeadSnapshotID
		parentID = &pid
	}

	desc := &model.Descriptor{
		SnapshotID:      snapshotID,
		ParentID:        parentID,
		WorktreeName:    worktreeName,
		CreatedAt:       time.Now().UTC(),
		Note:            note,
		Tags:            tags,
		Engine:          c.engineType,
		PayloadRootHash: payloadHash,
		IntegrityState:  model.IntegrityVerified,
	}

	// Step 8: Compute descriptor checksum
	checksum, err := integrity.ComputeDescriptorChecksum(desc)
	if err != nil {
		os.RemoveAll(snapshotDir)
		return nil, fmt.Errorf("compute checksum: %w", err)
	}
	desc.DescriptorChecksum = checksum

	// Step 9: Write .READY marker
	readyMarker := &model.ReadyMarker{
		SnapshotID:         snapshotID,
		CompletedAt:        time.Now().UTC(),
		PayloadHash:        payloadHash,
		Engine:             c.engineType,
		DescriptorChecksum: checksum,
	}
	readyPath := filepath.Join(snapshotDir, ".READY")
	if err := c.writeReadyMarker(readyPath, readyMarker); err != nil {
		os.RemoveAll(snapshotDir)
		return nil, fmt.Errorf("write ready marker: %w", err)
	}

	// Step 10: Write descriptor atomically
	descriptorPath := filepath.Join(c.repoRoot, ".jvs", "descriptors", string(snapshotID)+".json")
	if err := c.writeDescriptor(descriptorPath, desc); err != nil {
		os.RemoveAll(snapshotDir)
		return nil, fmt.Errorf("write descriptor: %w", err)
	}

	// Step 11: Update worktree head and latest
	if err := wtMgr.SetLatest(worktreeName, snapshotID); err != nil {
		// Don't remove snapshot, it's valid
		return nil, fmt.Errorf("update head: %w", err)
	}

	// Step 12: Write audit log
	if err := c.auditLogger.Append(model.EventTypeSnapshotCreate, worktreeName, snapshotID, map[string]any{
		"engine":   string(c.engineType),
		"note":     note,
		"checksum": string(checksum),
	}); err != nil {
		// Non-fatal, just log
		fmt.Fprintf(os.Stderr, "warning: failed to write audit log: %v\n", err)
	}

	return desc, nil
}

func (c *Creator) writeIntent(path string, intent *model.IntentRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(intent)
	if err != nil {
		return err
	}
	return fsutil.AtomicWrite(path, data, 0644)
}

func (c *Creator) writeReadyMarker(path string, marker *model.ReadyMarker) error {
	data, err := json.Marshal(marker)
	if err != nil {
		return err
	}
	return fsutil.AtomicWrite(path, data, 0644)
}

func (c *Creator) writeDescriptor(path string, desc *model.Descriptor) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.AtomicWrite(path, data, 0644)
}

// LoadDescriptor loads a descriptor from disk.
func LoadDescriptor(repoRoot string, snapshotID model.SnapshotID) (*model.Descriptor, error) {
	path := filepath.Join(repoRoot, ".jvs", "descriptors", string(snapshotID)+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errclass.ErrDescriptorCorrupt.WithMessagef("descriptor not found: %s", snapshotID)
		}
		return nil, fmt.Errorf("read descriptor: %w", err)
	}
	var desc model.Descriptor
	if err := json.Unmarshal(data, &desc); err != nil {
		return nil, errclass.ErrDescriptorCorrupt.WithMessagef("parse descriptor: %v", err)
	}
	return &desc, nil
}

// VerifySnapshot verifies a snapshot's integrity.
func VerifySnapshot(repoRoot string, snapshotID model.SnapshotID, verifyPayloadHash bool) error {
	desc, err := LoadDescriptor(repoRoot, snapshotID)
	if err != nil {
		return err
	}

	// Verify checksum
	computedChecksum, err := integrity.ComputeDescriptorChecksum(desc)
	if err != nil {
		return fmt.Errorf("compute checksum: %w", err)
	}
	if computedChecksum != desc.DescriptorChecksum {
		return errclass.ErrDescriptorCorrupt.WithMessage("checksum mismatch")
	}

	if verifyPayloadHash {
		snapshotDir := filepath.Join(repoRoot, ".jvs", "snapshots", string(snapshotID))
		computedHash, err := integrity.ComputePayloadRootHash(snapshotDir)
		if err != nil {
			return fmt.Errorf("compute payload hash: %w", err)
		}
		if computedHash != desc.PayloadRootHash {
			return errclass.ErrPayloadHashMismatch.WithMessage("payload hash mismatch")
		}
	}

	return nil
}
