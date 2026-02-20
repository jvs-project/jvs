package ref

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jvs-project/jvs/internal/audit"
	"github.com/jvs-project/jvs/pkg/fsutil"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/jvs-project/jvs/pkg/pathutil"
)

// Manager handles named reference operations.
type Manager struct {
	repoRoot    string
	auditLogger *audit.FileAppender
}

// NewManager creates a new ref manager.
func NewManager(repoRoot string) *Manager {
	auditPath := filepath.Join(repoRoot, ".jvs", "audit", "audit.jsonl")
	return &Manager{
		repoRoot:    repoRoot,
		auditLogger: audit.NewFileAppender(auditPath),
	}
}

// Create creates a new named reference to a snapshot.
func (m *Manager) Create(name string, targetID model.SnapshotID, description string) (*model.RefRecord, error) {
	if err := pathutil.ValidateName(name); err != nil {
		return nil, err
	}

	refPath := m.refPath(name)
	if _, err := os.Stat(refPath); err == nil {
		return nil, fmt.Errorf("ref %s already exists", name)
	}

	rec := &model.RefRecord{
		Name:        name,
		TargetID:    targetID,
		CreatedAt:   time.Now().UTC(),
		Description: description,
	}

	if err := m.writeRef(refPath, rec); err != nil {
		return nil, fmt.Errorf("write ref: %w", err)
	}

	m.auditLogger.Append(model.EventTypeRefCreate, "", targetID, map[string]any{
		"ref_name": name,
	})

	return rec, nil
}

// List returns all named references.
func (m *Manager) List() ([]*model.RefRecord, error) {
	refsDir := filepath.Join(m.repoRoot, ".jvs", "refs")
	entries, err := os.ReadDir(refsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read refs directory: %w", err)
	}

	var refs []*model.RefRecord
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		name := entry.Name()[:len(entry.Name())-5] // remove .json
		rec, err := m.Get(name)
		if err != nil {
			continue // skip malformed
		}
		refs = append(refs, rec)
	}

	// Sort by name
	sort.Slice(refs, func(i, j int) bool {
		return refs[i].Name < refs[j].Name
	})

	return refs, nil
}

// Get returns a specific reference.
func (m *Manager) Get(name string) (*model.RefRecord, error) {
	refPath := m.refPath(name)
	data, err := os.ReadFile(refPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("ref %s not found", name)
		}
		return nil, fmt.Errorf("read ref: %w", err)
	}

	var rec model.RefRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("parse ref: %w", err)
	}

	return &rec, nil
}

// Delete removes a reference.
func (m *Manager) Delete(name string) error {
	rec, err := m.Get(name)
	if err != nil {
		return err
	}

	refPath := m.refPath(name)
	if err := os.Remove(refPath); err != nil {
		return fmt.Errorf("remove ref: %w", err)
	}

	m.auditLogger.Append(model.EventTypeRefDelete, "", rec.TargetID, map[string]any{
		"ref_name": name,
	})

	return nil
}

func (m *Manager) refPath(name string) string {
	return filepath.Join(m.repoRoot, ".jvs", "refs", name+".json")
}

func (m *Manager) writeRef(path string, rec *model.RefRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.AtomicWrite(path, data, 0644)
}
