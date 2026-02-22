package gc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jvs-project/jvs/internal/audit"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/config"
	"github.com/jvs-project/jvs/pkg/fsutil"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/jvs-project/jvs/pkg/uuidutil"
)

// Collector handles garbage collection.
type Collector struct {
	repoRoot         string
	auditLogger      *audit.FileAppender
	progressCallback func(string, int, int, string)
	retentionPolicy  model.RetentionPolicy
}

// NewCollector creates a new GC collector.
func NewCollector(repoRoot string) *Collector {
	auditPath := filepath.Join(repoRoot, ".jvs", "audit", "audit.jsonl")
	return &Collector{
		repoRoot:        repoRoot,
		auditLogger:     audit.NewFileAppender(auditPath),
		retentionPolicy: loadRetentionPolicy(repoRoot),
	}
}

// loadRetentionPolicy loads the retention policy from config.
func loadRetentionPolicy(repoRoot string) model.RetentionPolicy {
	cfg, err := config.Load(repoRoot)
	if err != nil {
		return model.DefaultRetentionPolicy()
	}
	return cfg.GetRetentionPolicy()
}

// SetProgressCallback sets a callback for progress updates.
func (c *Collector) SetProgressCallback(cb func(string, int, int, string)) {
	c.progressCallback = cb
}

// Plan creates a GC plan based on the current retention policy.
func (c *Collector) Plan() (*model.GCPlan, error) {
	allSnapshots, err := c.listAllSnapshots()
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}

	protectedMap, protectedByLineage, protectedByPin, protectedByRetention, err := c.computeProtectedSet(allSnapshots)
	if err != nil {
		return nil, fmt.Errorf("compute protected set: %w", err)
	}

	var protectedSet []model.SnapshotID
	for id := range protectedMap {
		protectedSet = append(protectedSet, id)
	}

	var toDelete []model.SnapshotID
	for _, id := range allSnapshots {
		if !protectedMap[id] {
			toDelete = append(toDelete, id)
		}
	}

	// Estimate bytes (rough)
	deletableBytes := int64(len(toDelete)) * 1024 * 1024 // assume 1MB each

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
		RetentionPolicy:        c.retentionPolicy,
	}

	// Write plan
	if err := c.writePlan(plan); err != nil {
		return nil, fmt.Errorf("write plan: %w", err)
	}

	return plan, nil
}

// Run executes a GC plan.
func (c *Collector) Run(planID string) error {
	plan, err := c.LoadPlan(planID)
	if err != nil {
		return fmt.Errorf("load plan: %w", err)
	}

	// Revalidate protected set
	allSnapshots, err := c.listAllSnapshots()
	if err != nil {
		return fmt.Errorf("list snapshots: %w", err)
	}

	protectedMap, _, _, _, err := c.computeProtectedSet(allSnapshots)
	if err != nil {
		return fmt.Errorf("revalidate protected set: %w", err)
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

// snapshotWithTime is used for sorting snapshots by creation time.
type snapshotWithTime struct {
	ID        model.SnapshotID
	CreatedAt time.Time
}

// computeProtectedSet determines which snapshots should be protected from GC.
// Returns: (protected map, protected by lineage count, protected by pin count, protected by retention count, error)
func (c *Collector) computeProtectedSet(allSnapshots []model.SnapshotID) (map[model.SnapshotID]bool, int, int, int, error) {
	protected := make(map[model.SnapshotID]bool)
	lineageCount := 0
	pinCount := 0
	retentionCount := 0

	// 1. All worktree heads
	wtMgr := worktree.NewManager(c.repoRoot)
	wtList, err := wtMgr.List()
	if err != nil {
		return nil, 0, 0, 0, err
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

	// 5. Retention policy: keep N most recent snapshots
	snapshotsByTime := c.getSnapshotsByTime(allSnapshots)
	keepCount := c.retentionPolicy.KeepMinSnapshots
	if keepCount > 0 {
		// Keep the N most recent snapshots (that aren't already protected)
		for i := len(snapshotsByTime) - 1; i >= 0 && keepCount > 0; i-- {
			snap := snapshotsByTime[i]
			if !protected[snap.ID] {
				protected[snap.ID] = true
				retentionCount++
				keepCount--
			}
		}
	}

	// 6. Retention policy: keep snapshots within the minimum age
	minAge := c.retentionPolicy.KeepMinAge
	if minAge > 0 {
		cutoff := time.Now().Add(-minAge)
		for _, snap := range snapshotsByTime {
			if !protected[snap.ID] && snap.CreatedAt.After(cutoff) {
				protected[snap.ID] = true
				retentionCount++
			}
		}
	}

	return protected, lineageCount, pinCount, retentionCount, nil
}

// getSnapshotsByTime returns all snapshots sorted by creation time.
func (c *Collector) getSnapshotsByTime(allSnapshots []model.SnapshotID) []snapshotWithTime {
	var snapshots []snapshotWithTime
	for _, id := range allSnapshots {
		desc, err := snapshot.LoadDescriptor(c.repoRoot, id)
		if err != nil {
			continue // Skip snapshots we can't read
		}
		snapshots = append(snapshots, snapshotWithTime{
			ID:        id,
			CreatedAt: desc.CreatedAt,
		})
	}
	// Sort by creation time (oldest first)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].CreatedAt.Before(snapshots[j].CreatedAt)
	})
	return snapshots
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
