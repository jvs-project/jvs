package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jvs-project/jvs/internal/audit"
	"github.com/jvs-project/jvs/internal/compression"
	"github.com/jvs-project/jvs/internal/engine"
	"github.com/jvs-project/jvs/internal/integrity"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/errclass"
	"github.com/jvs-project/jvs/pkg/fsutil"
	"github.com/jvs-project/jvs/pkg/model"
)

// Creator handles snapshot creation.
type Creator struct {
	repoRoot       string
	engineType     model.EngineType
	engine         engine.Engine
	auditLogger    *audit.FileAppender
	compression    *compression.Compressor
}

// NewCreator creates a new snapshot creator.
func NewCreator(repoRoot string, engineType model.EngineType) *Creator {
	return NewCreatorWithCompression(repoRoot, engineType, nil)
}

// NewCreatorWithCompression creates a new snapshot creator with compression.
func NewCreatorWithCompression(repoRoot string, engineType model.EngineType, comp *compression.Compressor) *Creator {
	eng := engine.NewEngine(engineType)

	auditPath := filepath.Join(repoRoot, ".jvs", "audit", "audit.jsonl")
	return &Creator{
		repoRoot:    repoRoot,
		engineType:  engineType,
		engine:      eng,
		auditLogger: audit.NewFileAppender(auditPath),
		compression: comp,
	}
}

// SetCompression sets the compression level for this creator.
func (c *Creator) SetCompression(level compression.CompressionLevel) {
	c.compression = compression.NewCompressor(level)
}

// Create performs a full snapshot of the worktree using the 12-step protocol.
func (c *Creator) Create(worktreeName, note string, tags []string) (*model.Descriptor, error) {
	return c.CreatePartial(worktreeName, note, tags, nil)
}

// CreatePartial performs a snapshot of specific paths within the worktree.
// If paths is nil or empty, performs a full snapshot.
func (c *Creator) CreatePartial(worktreeName, note string, tags []string, paths []string) (*model.Descriptor, error) {
	// Step 1: Validate worktree exists
	wtMgr := worktree.NewManager(c.repoRoot)
	cfg, err := wtMgr.Get(worktreeName)
	if err != nil {
		return nil, fmt.Errorf("get worktree: %w", err)
	}

	// Normalize and validate paths if provided
	var partialPaths []string
	if len(paths) > 0 {
		partialPaths, err = c.validateAndNormalizePaths(paths, worktreeName)
		if err != nil {
			return nil, err
		}
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

	// Step 4: Create snapshot .tmp directory (atomic publish pattern)
	snapshotTmpDir := filepath.Join(c.repoRoot, ".jvs", "snapshots", string(snapshotID)+".tmp")
	snapshotDir := filepath.Join(c.repoRoot, ".jvs", "snapshots", string(snapshotID))
	if err := os.MkdirAll(snapshotTmpDir, 0755); err != nil {
		return nil, fmt.Errorf("create snapshot tmp dir: %w", err)
	}

	// Cleanup helper for failure cases
	cleanupTmp := func() {
		os.RemoveAll(snapshotTmpDir)
	}

	// Step 5: Clone payload to snapshot .tmp directory
	payloadPath := wtMgr.Path(worktreeName)

	// For partial snapshots, only copy specified paths
	if len(partialPaths) > 0 {
		if err := c.clonePaths(payloadPath, snapshotTmpDir, partialPaths); err != nil {
			cleanupTmp()
			return nil, fmt.Errorf("clone partial paths: %w", err)
		}
	} else {
		if _, err := c.engine.Clone(payloadPath, snapshotTmpDir); err != nil {
			cleanupTmp()
			return nil, fmt.Errorf("clone payload: %w", err)
		}
	}

	// Step 6: Fsync the cloned tree for durability
	if err := fsutil.FsyncTree(snapshotTmpDir); err != nil {
		cleanupTmp()
		return nil, fmt.Errorf("fsync snapshot tree: %w", err)
	}

	// Step 7: Compute payload root hash
	payloadHash, err := integrity.ComputePayloadRootHash(snapshotTmpDir)
	if err != nil {
		cleanupTmp()
		return nil, fmt.Errorf("compute payload hash: %w", err)
	}

	// Step 8: Create descriptor
	var parentID *model.SnapshotID
	if cfg.HeadSnapshotID != "" {
		pid := cfg.HeadSnapshotID
		parentID = &pid
	}

	// Build descriptor with compression info if enabled
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
		PartialPaths:    partialPaths,
	}

	// Add compression info if compression is enabled
	if c.compression != nil && c.compression.IsEnabled() {
		desc.Compression = &model.CompressionInfo{
			Type:  string(c.compression.Type),
			Level: int(c.compression.Level),
		}
	}

	// Step 9: Compute descriptor checksum
	checksum, err := integrity.ComputeDescriptorChecksum(desc)
	if err != nil {
		cleanupTmp()
		return nil, fmt.Errorf("compute checksum: %w", err)
	}
	desc.DescriptorChecksum = checksum

	// Step 10: Write .READY marker in tmp
	readyMarker := &model.ReadyMarker{
		SnapshotID:         snapshotID,
		CompletedAt:        time.Now().UTC(),
		PayloadHash:        payloadHash,
		Engine:             c.engineType,
		DescriptorChecksum: checksum,
	}
	readyPath := filepath.Join(snapshotTmpDir, ".READY")
	if err := c.writeReadyMarker(readyPath, readyMarker); err != nil {
		cleanupTmp()
		return nil, fmt.Errorf("write ready marker: %w", err)
	}

	// Step 11: Atomic rename tmp -> final
	if err := fsutil.RenameAndSync(snapshotTmpDir, snapshotDir); err != nil {
		cleanupTmp()
		return nil, fmt.Errorf("atomic rename snapshot: %w", err)
	}

	// Step 11.5: Compress snapshot if enabled
	if c.compression != nil && c.compression.IsEnabled() {
		count, err := c.compression.CompressDir(snapshotDir)
		if err != nil {
			// Compression failure is non-fatal; snapshot is valid
			fmt.Fprintf(os.Stderr, "warning: compression failed: %v\n", err)
		} else if count > 0 {
			// Log compression success
			fmt.Fprintf(os.Stderr, "compressed %d files\n", count)
		}
	}

	// Step 12: Write descriptor atomically
	descriptorPath := filepath.Join(c.repoRoot, ".jvs", "descriptors", string(snapshotID)+".json")
	if err := c.writeDescriptor(descriptorPath, desc); err != nil {
		// Snapshot is already renamed, don't remove it
		return nil, fmt.Errorf("write descriptor: %w", err)
	}

	// Step 13: Update worktree head and latest
	if err := wtMgr.SetLatest(worktreeName, snapshotID); err != nil {
		// Don't remove snapshot, it's valid
		return nil, fmt.Errorf("update head: %w", err)
	}

	// Step 14: Write audit log
	auditData := map[string]any{
		"engine":   string(c.engineType),
		"note":     note,
		"checksum": string(checksum),
	}
	if len(partialPaths) > 0 {
		auditData["partial_paths"] = partialPaths
	}
	if err := c.auditLogger.Append(model.EventTypeSnapshotCreate, worktreeName, snapshotID, auditData); err != nil {
		// Non-fatal, just log
		fmt.Fprintf(os.Stderr, "warning: failed to write audit log: %v\n", err)
	}

	return desc, nil
}

// validateAndNormalizePaths validates and normalizes the partial snapshot paths.
func (c *Creator) validateAndNormalizePaths(paths []string, worktreeName string) ([]string, error) {
	wtMgr := worktree.NewManager(c.repoRoot)
	wtPath := wtMgr.Path(worktreeName)

	var normalized []string
	for _, p := range paths {
		// Clean the path
		p = filepath.Clean(p)

		// Ensure it's relative
		if filepath.IsAbs(p) {
			return nil, fmt.Errorf("path must be relative: %s", p)
		}

		// Check for path traversal attempts
		if strings.Contains(p, "..") {
			return nil, fmt.Errorf("path cannot contain '..': %s", p)
		}

		// Build full path and verify it exists within worktree
		fullPath := filepath.Join(wtPath, p)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("path does not exist: %s", p)
		}

		// Verify it's within worktree
		absWtPath, err := filepath.Abs(wtPath)
		if err != nil {
			return nil, fmt.Errorf("resolve worktree path: %w", err)
		}
		absFullPath, err := filepath.Abs(fullPath)
		if err != nil {
			return nil, fmt.Errorf("resolve full path: %w", err)
		}

		rel, err := filepath.Rel(absWtPath, absFullPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return nil, fmt.Errorf("path is outside worktree: %s", p)
		}

		normalized = append(normalized, p)
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var unique []string
	for _, p := range normalized {
		if !seen[p] {
			seen[p] = true
			unique = append(unique, p)
		}
	}

	return unique, nil
}

// clonePaths clones only the specified paths from source to destination.
func (c *Creator) clonePaths(src, dst string, paths []string) error {
	for _, p := range paths {
		srcPath := filepath.Join(src, p)
		dstPath := filepath.Join(dst, p)

		// Get source info
		info, err := os.Stat(srcPath)
		if err != nil {
			return fmt.Errorf("stat %s: %w", p, err)
		}

		if info.IsDir() {
			// Clone directory tree
			if _, err := c.engine.Clone(srcPath, dstPath); err != nil {
				return fmt.Errorf("clone directory %s: %w", p, err)
			}
		} else {
			// Clone single file - ensure parent dir exists
			if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
				return fmt.Errorf("create parent dir for %s: %w", p, err)
			}
			if _, err := c.engine.Clone(srcPath, dstPath); err != nil {
				return fmt.Errorf("clone file %s: %w", p, err)
			}
		}
	}
	return nil
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
