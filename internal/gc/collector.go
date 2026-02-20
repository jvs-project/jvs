package gc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jvs-project/jvs/internal/audit"
	"github.com/jvs-project/jvs/internal/ref"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/fsutil"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/jvs-project/jvs/pkg/uuidutil"
)

// Collector handles garbage collection.
type Collector struct {
	repoRoot    string
	auditLogger *audit.FileAppender
}

// NewCollector creates a new GC collector.
func NewCollector(repoRoot string) *Collector {
	auditPath := filepath.Join(repoRoot, ".jvs", "audit", "audit.jsonl")
	return &Collector{
		repoRoot:    repoRoot,
		auditLogger: audit.NewFileAppender(auditPath),
	}
}

// Plan creates a GC plan.
func (c *Collector) Plan() (*model.GCPlan, error) {
	protectedSet, err := c.computeProtectedSet()
	if err != nil {
		return nil, fmt.Errorf("compute protected set: %w", err)
	}

	// Find all snapshots
	allSnapshots, err := c.listAllSnapshots()
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}

	// Determine what to delete
	protectedMap := make(map[model.SnapshotID]bool)
	for _, id := range protectedSet {
		protectedMap[id] = true
	}

	var toDelete []model.SnapshotID
	for _, id := range allSnapshots {
		if !protectedMap[id] {
			toDelete = append(toDelete, id)
		}
	}

	plan := &model.GCPlan{
		PlanID:       uuidutil.NewV4(),
		CreatedAt:    time.Now().UTC(),
		ProtectedSet: protectedSet,
		ToDelete:     toDelete,
		RetentionPolicy: model.RetentionPolicy{
			KeepMinSnapshots: 10,
			KeepMinAge:       24 * time.Hour,
		},
	}

	// Estimate bytes (rough)
	plan.EstimatedBytes = int64(len(toDelete)) * 1024 * 1024 // assume 1MB each

	// Write plan
	if err := c.writePlan(plan); err != nil {
		return nil, fmt.Errorf("write plan: %w", err)
	}

	return plan, nil
}

// Run executes a GC plan.
func (c *Collector) Run(planID string) error {
	plan, err := c.loadPlan(planID)
	if err != nil {
		return fmt.Errorf("load plan: %w", err)
	}

	// Revalidate protected set
	currentProtected, err := c.computeProtectedSet()
	if err != nil {
		return fmt.Errorf("revalidate protected set: %w", err)
	}

	protectedMap := make(map[model.SnapshotID]bool)
	for _, id := range currentProtected {
		protectedMap[id] = true
	}

	// Check for plan mismatch
	for _, id := range plan.ToDelete {
		if protectedMap[id] {
			return fmt.Errorf("plan mismatch: %s is now protected", id)
		}
	}

	// Delete snapshots
	var deleted []model.SnapshotID
	for _, snapshotID := range plan.ToDelete {
		if err := c.deleteSnapshot(snapshotID); err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "warning: failed to delete %s: %v\n", snapshotID, err)
			continue
		}
		deleted = append(deleted, snapshotID)
	}

	// Write tombstones
	for _, snapshotID := range deleted {
		tombstone := &model.Tombstone{
			SnapshotID:  snapshotID,
			DeletedAt:   time.Now().UTC(),
			Reclaimable: true,
		}
		c.writeTombstone(tombstone)
	}

	// Cleanup plan
	c.deletePlan(planID)

	// Audit
	c.auditLogger.Append(model.EventTypeGCRun, "", "", map[string]any{
		"plan_id":       planID,
		"deleted_count": len(deleted),
	})

	return nil
}

func (c *Collector) computeProtectedSet() ([]model.SnapshotID, error) {
	protected := make(map[model.SnapshotID]bool)

	// 1. All worktree heads
	wtMgr := worktree.NewManager(c.repoRoot)
	wtList, err := wtMgr.List()
	if err != nil {
		return nil, err
	}
	for _, cfg := range wtList {
		if cfg.HeadSnapshotID != "" {
			protected[cfg.HeadSnapshotID] = true
		}
	}

	// 2. All refs
	refMgr := ref.NewManager(c.repoRoot)
	refs, err := refMgr.List()
	if err == nil {
		for _, r := range refs {
			protected[r.TargetID] = true
		}
	}

	// 3. Lineage traversal (keep parent chains)
	for id := range protected {
		c.walkLineage(id, protected)
	}

	// 4. All intents (in-progress operations)
	intentsDir := filepath.Join(c.repoRoot, ".jvs", "intents")
	entries, _ := os.ReadDir(intentsDir)
	for _, entry := range entries {
		protected[model.SnapshotID(entry.Name()[:len(entry.Name())-5])] = true
	}

	var result []model.SnapshotID
	for id := range protected {
		result = append(result, id)
	}
	return result, nil
}

func (c *Collector) walkLineage(snapshotID model.SnapshotID, protected map[model.SnapshotID]bool) {
	desc, err := snapshot.LoadDescriptor(c.repoRoot, snapshotID)
	if err != nil {
		return
	}
	if desc.ParentID != nil && !protected[*desc.ParentID] {
		protected[*desc.ParentID] = true
		c.walkLineage(*desc.ParentID, protected)
	}
}

func (c *Collector) listAllSnapshots() ([]model.SnapshotID, error) {
	snapshotsDir := filepath.Join(c.repoRoot, ".jvs", "snapshots")
	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var ids []model.SnapshotID
	for _, entry := range entries {
		if entry.IsDir() {
			ids = append(ids, model.SnapshotID(entry.Name()))
		}
	}
	return ids, nil
}

func (c *Collector) deleteSnapshot(snapshotID model.SnapshotID) error {
	// Delete snapshot directory
	snapshotDir := filepath.Join(c.repoRoot, ".jvs", "snapshots", string(snapshotID))
	if err := os.RemoveAll(snapshotDir); err != nil {
		return err
	}

	// Delete descriptor
	descriptorPath := filepath.Join(c.repoRoot, ".jvs", "descriptors", string(snapshotID)+".json")
	os.Remove(descriptorPath)

	return nil
}

func (c *Collector) writePlan(plan *model.GCPlan) error {
	gcDir := filepath.Join(c.repoRoot, ".jvs", "gc")
	os.MkdirAll(gcDir, 0755)
	path := filepath.Join(gcDir, plan.PlanID+".json")
	data, _ := json.MarshalIndent(plan, "", "  ")
	return fsutil.AtomicWrite(path, data, 0644)
}

func (c *Collector) loadPlan(planID string) (*model.GCPlan, error) {
	path := filepath.Join(c.repoRoot, ".jvs", "gc", planID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var plan model.GCPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

func (c *Collector) deletePlan(planID string) {
	path := filepath.Join(c.repoRoot, ".jvs", "gc", planID+".json")
	os.Remove(path)
}

func (c *Collector) writeTombstone(tombstone *model.Tombstone) {
	gcDir := filepath.Join(c.repoRoot, ".jvs", "gc")
	os.MkdirAll(gcDir, 0755)
	path := filepath.Join(gcDir, "tombstones", string(tombstone.SnapshotID)+".json")
	data, _ := json.MarshalIndent(tombstone, "", "  ")
	fsutil.AtomicWrite(path, data, 0644)
}
