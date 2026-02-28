// Package gc provides garbage collection for snapshots.
package gc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jvs-project/jvs/internal/audit"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/fsutil"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/jvs-project/jvs/pkg/uuidutil"
)

// Collector handles garbage collection of unused snapshots.
type Collector struct {
	repoRoot         string
	auditLogger      *audit.FileAppender
	progressCallback func(string, int, int, string)
}

// NewCollector creates a new GC collector.
func NewCollector(repoRoot string) *Collector {
	auditPath := filepath.Join(repoRoot, ".jvs", "audit", "audit.jsonl")
	return &Collector{
		repoRoot:    repoRoot,
		auditLogger: audit.NewFileAppender(auditPath),
	}
}

// SetProgressCallback sets a callback for progress updates.
func (c *Collector) SetProgressCallback(cb func(string, int, int, string)) {
	c.progressCallback = cb
}

// Plan creates a GC plan.
func (c *Collector) Plan() (*model.GCPlan, error) {
	return c.PlanWithPolicy(model.DefaultRetentionPolicy())
}

// PlanWithPolicy creates a GC plan using the given retention policy.
func (c *Collector) PlanWithPolicy(policy model.RetentionPolicy) (*model.GCPlan, error) {
	protectedSet, protectedByLineage, protectedByPin, err := c.computeProtectedSet()
	if err != nil {
		return nil, fmt.Errorf("compute protected set: %w", err)
	}

	// Find all snapshots with descriptors for retention analysis
	allSnapshots, err := c.listAllSnapshots()
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}

	protectedMap := make(map[model.SnapshotID]bool)
	for _, id := range protectedSet {
		protectedMap[id] = true
	}

	// Apply retention policy: protect by age
	protectedByRetention := 0
	now := time.Now()
	if policy.KeepMinAge > 0 {
		for _, id := range allSnapshots {
			if protectedMap[id] {
				continue
			}
			desc, err := snapshot.LoadDescriptor(c.repoRoot, id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: gc: skipping descriptor %s: %v\n", id, err)
				continue
			}
			if now.Sub(desc.CreatedAt) < policy.KeepMinAge {
				protectedMap[id] = true
				protectedByRetention++
			}
		}
	}

	// Apply retention policy: protect by count (keep most recent N)
	if policy.KeepMinSnapshots > 0 {
		allDescs, err := snapshot.ListAll(c.repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: gc: failed to list all descriptors for retention-by-count: %v\n", err)
		}
		if err == nil {
			kept := 0
			for _, desc := range allDescs {
				if kept >= policy.KeepMinSnapshots {
					break
				}
				if !protectedMap[desc.SnapshotID] {
					protectedMap[desc.SnapshotID] = true
					protectedByRetention++
				}
				kept++
			}
		}
	}

	// Rebuild protected set from map
	protectedSet = protectedSet[:0]
	for id := range protectedMap {
		protectedSet = append(protectedSet, id)
	}

	var toDelete []model.SnapshotID
	for _, id := range allSnapshots {
		if !protectedMap[id] {
			toDelete = append(toDelete, id)
		}
	}

	deletableBytes := int64(len(toDelete)) * 1024 * 1024

	plan := &model.GCPlan{
		PlanID:                 uuidutil.NewV4(),
		CreatedAt:              time.Now().UTC(),
		ProtectedSet:           protectedSet,
		ProtectedByPin:         protectedByPin,
		ProtectedByLineage:     protectedByLineage,
		ProtectedByRetention:   protectedByRetention,
		CandidateCount:         len(toDelete),
		ToDelete:               toDelete,
		DeletableBytesEstimate: deletableBytes,
		RetentionPolicy:        policy,
	}

	if err := c.writePlan(plan); err != nil {
		return nil, fmt.Errorf("write plan: %w", err)
	}

	return plan, nil
}

// Run executes a GC plan.
func (c *Collector) Run(planID string) error {
	if planID == "" {
		return fmt.Errorf("plan ID is required")
	}

	plan, err := c.LoadPlan(planID)
	if err != nil {
		return fmt.Errorf("load plan: %w", err)
	}

	// Revalidate protected set
	currentProtected, _, _, err := c.computeProtectedSet()
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

	totalToDelete := len(plan.ToDelete)

	// Delete snapshots
	var deleted []model.SnapshotID
	for i, snapshotID := range plan.ToDelete {
		// Report progress
		if c.progressCallback != nil {
			c.progressCallback("gc", i+1, totalToDelete, fmt.Sprintf("deleting %s", snapshotID.ShortID()))
		}

		if err := c.deleteSnapshot(snapshotID); err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "warning: failed to delete %s: %v\n", snapshotID, err)
			continue
		}
		deleted = append(deleted, snapshotID)
	}

	// Report completion
	if c.progressCallback != nil && totalToDelete > 0 {
		c.progressCallback("gc", totalToDelete, totalToDelete, fmt.Sprintf("deleted %d snapshots", len(deleted)))
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

func (c *Collector) computeProtectedSet() ([]model.SnapshotID, int, int, error) {
	protected := make(map[model.SnapshotID]bool)
	lineageCount := 0
	pinCount := 0

	// 1. All worktree heads
	wtMgr := worktree.NewManager(c.repoRoot)
	wtList, err := wtMgr.List()
	if err != nil {
		return nil, 0, 0, err
	}
	for _, cfg := range wtList {
		if cfg.HeadSnapshotID != "" {
			protected[cfg.HeadSnapshotID] = true
		}
	}

	// 2. Lineage traversal (keep parent chains)
	for id := range protected {
		lineageCount += c.walkLineage(id, protected)
	}

	// 3. All intents (in-progress operations)
	intentsDir := filepath.Join(c.repoRoot, ".jvs", "intents")
	entries, _ := os.ReadDir(intentsDir)
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			protected[model.SnapshotID(strings.TrimSuffix(name, ".json"))] = true
		}
	}

	// 4. All pins
	pinsDir := filepath.Join(c.repoRoot, ".jvs", "pins")
	pinEntries, err := os.ReadDir(pinsDir)
	if err == nil {
		for _, entry := range pinEntries {
			name := entry.Name()
			if strings.HasSuffix(name, ".json") {
				pinPath := filepath.Join(pinsDir, name)
				data, err := os.ReadFile(pinPath)
				if err != nil {
					continue
				}
				var pin model.Pin
				if err := json.Unmarshal(data, &pin); err != nil {
					continue
				}
				// Check if pin has expired
				if pin.ExpiresAt != nil && pin.ExpiresAt.Before(time.Now()) {
					continue // Skip expired pins
				}
				if !protected[pin.SnapshotID] {
					protected[pin.SnapshotID] = true
					pinCount++
				}
			}
		}
	}

	var result []model.SnapshotID
	for id := range protected {
		result = append(result, id)
	}
	return result, lineageCount, pinCount, nil
}

func (c *Collector) walkLineage(snapshotID model.SnapshotID, protected map[model.SnapshotID]bool) int {
	count := 0
	desc, err := snapshot.LoadDescriptor(c.repoRoot, snapshotID)
	if err != nil {
		return count
	}
	if desc.ParentID != nil && !protected[*desc.ParentID] {
		protected[*desc.ParentID] = true
		count = 1 + c.walkLineage(*desc.ParentID, protected)
	}
	return count
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
		return fmt.Errorf("remove snapshot dir: %w", err)
	}

	// Delete descriptor - log warning if fails but don't fail the operation
	descriptorPath := filepath.Join(c.repoRoot, ".jvs", "descriptors", string(snapshotID)+".json")
	if err := os.Remove(descriptorPath); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "warning: failed to remove descriptor %s: %v\n", snapshotID, err)
	}

	return nil
}

func (c *Collector) writePlan(plan *model.GCPlan) error {
	gcDir := filepath.Join(c.repoRoot, ".jvs", "gc")
	if err := os.MkdirAll(gcDir, 0755); err != nil {
		return fmt.Errorf("create gc dir: %w", err)
	}
	path := filepath.Join(gcDir, plan.PlanID+".json")
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal plan: %w", err)
	}
	return fsutil.AtomicWrite(path, data, 0644)
}

// LoadPlan loads a GC plan by ID.
func (c *Collector) LoadPlan(planID string) (*model.GCPlan, error) {
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
	gcDir := filepath.Join(c.repoRoot, ".jvs", "gc", "tombstones")
	if err := os.MkdirAll(gcDir, 0755); err != nil {
		return // Best effort - log in production
	}
	path := filepath.Join(gcDir, string(tombstone.SnapshotID)+".json")
	data, err := json.MarshalIndent(tombstone, "", "  ")
	if err != nil {
		return
	}
	fsutil.AtomicWrite(path, data, 0644)
}
