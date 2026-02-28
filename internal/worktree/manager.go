// Package worktree provides worktree management operations.
package worktree

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jvs-project/jvs/internal/audit"
	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/jvs-project/jvs/pkg/pathutil"
)

// Manager handles worktree CRUD operations.
type Manager struct {
	repoRoot string
}

// NewManager creates a new worktree manager.
func NewManager(repoRoot string) *Manager {
	return &Manager{repoRoot: repoRoot}
}

// Create creates a new worktree with the given name.
func (m *Manager) Create(name string, baseSnapshotID *model.SnapshotID) (*model.WorktreeConfig, error) {
	if err := pathutil.ValidateName(name); err != nil {
		return nil, err
	}

	// Check if already exists
	configPath := repo.WorktreeConfigPath(m.repoRoot, name)
	if _, err := os.Stat(configPath); err == nil {
		return nil, fmt.Errorf("worktree %s already exists", name)
	}

	// Create payload directory
	payloadPath := repo.WorktreePayloadPath(m.repoRoot, name)
	if err := os.MkdirAll(payloadPath, 0755); err != nil {
		return nil, fmt.Errorf("create payload directory: %w", err)
	}

	// Create config directory
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("create config directory: %w", err)
	}

	// Create config
	cfg := &model.WorktreeConfig{
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}
	if baseSnapshotID != nil {
		cfg.HeadSnapshotID = *baseSnapshotID
	}

	if err := repo.WriteWorktreeConfig(m.repoRoot, name, cfg); err != nil {
		os.RemoveAll(payloadPath)
		return nil, fmt.Errorf("write config: %w", err)
	}

	return cfg, nil
}

// CreateFromSnapshot creates a new worktree with content cloned from a snapshot.
// This is similar to Fork but uses "create" semantics (for the --from flag).
func (m *Manager) CreateFromSnapshot(name string, snapshotID model.SnapshotID, cloneFunc func(src, dst string) error) (*model.WorktreeConfig, error) {
	if err := pathutil.ValidateName(name); err != nil {
		return nil, err
	}

	// Check if already exists
	configPath := repo.WorktreeConfigPath(m.repoRoot, name)
	if _, err := os.Stat(configPath); err == nil {
		return nil, fmt.Errorf("worktree %s already exists", name)
	}

	// Create payload directory
	payloadPath := repo.WorktreePayloadPath(m.repoRoot, name)
	if err := os.MkdirAll(payloadPath, 0755); err != nil {
		return nil, fmt.Errorf("create payload directory: %w", err)
	}

	// Clone snapshot content to worktree
	snapshotDir := filepath.Join(m.repoRoot, ".jvs", "snapshots", string(snapshotID))
	if err := cloneFunc(snapshotDir, payloadPath); err != nil {
		os.RemoveAll(payloadPath)
		return nil, fmt.Errorf("clone snapshot content: %w", err)
	}

	// Create config directory
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		os.RemoveAll(payloadPath)
		return nil, fmt.Errorf("create config directory: %w", err)
	}

	// Create config
	cfg := &model.WorktreeConfig{
		Name:           name,
		CreatedAt:      time.Now().UTC(),
		BaseSnapshotID: snapshotID,
	}

	if err := repo.WriteWorktreeConfig(m.repoRoot, name, cfg); err != nil {
		os.RemoveAll(payloadPath)
		return nil, fmt.Errorf("write config: %w", err)
	}

	return cfg, nil
}

// List returns all worktrees.
func (m *Manager) List() ([]*model.WorktreeConfig, error) {
	worktreesDir := filepath.Join(m.repoRoot, ".jvs", "worktrees")
	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		return nil, fmt.Errorf("read worktrees directory: %w", err)
	}

	var configs []*model.WorktreeConfig
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		cfg, err := repo.LoadWorktreeConfig(m.repoRoot, entry.Name())
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping malformed worktree %s: %v\n", entry.Name(), err)
			continue
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

// Get returns the config for a specific worktree.
func (m *Manager) Get(name string) (*model.WorktreeConfig, error) {
	return repo.LoadWorktreeConfig(m.repoRoot, name)
}

// Path returns the payload path for a worktree.
func (m *Manager) Path(name string) string {
	return repo.WorktreePayloadPath(m.repoRoot, name)
}

// Rename renames a worktree.
func (m *Manager) Rename(oldName, newName string) error {
	if err := pathutil.ValidateName(newName); err != nil {
		return err
	}

	// Check if new name exists
	newConfigPath := repo.WorktreeConfigPath(m.repoRoot, newName)
	if _, err := os.Stat(newConfigPath); err == nil {
		return fmt.Errorf("worktree %s already exists", newName)
	}

	// Rename payload directory (if not main)
	if oldName != "main" {
		oldPayload := repo.WorktreePayloadPath(m.repoRoot, oldName)
		newPayload := repo.WorktreePayloadPath(m.repoRoot, newName)
		if err := os.Rename(oldPayload, newPayload); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("rename payload: %w", err)
		}
	}

	// Rename config directory
	oldConfigDir := filepath.Join(m.repoRoot, ".jvs", "worktrees", oldName)
	newConfigDir := filepath.Join(m.repoRoot, ".jvs", "worktrees", newName)
	if err := os.Rename(oldConfigDir, newConfigDir); err != nil {
		return fmt.Errorf("rename config directory: %w", err)
	}

	// Update config with new name
	cfg, err := repo.LoadWorktreeConfig(m.repoRoot, newName)
	if err != nil {
		return fmt.Errorf("load config after rename: %w", err)
	}
	cfg.Name = newName
	return repo.WriteWorktreeConfig(m.repoRoot, newName, cfg)
}

// Remove deletes a worktree. Fails if the worktree is main.
func (m *Manager) Remove(name string) error {
	if name == "main" {
		return errors.New("cannot remove main worktree")
	}

	// Get config before removal for audit logging
	cfg, _ := repo.LoadWorktreeConfig(m.repoRoot, name)

	// Remove payload directory
	payloadPath := repo.WorktreePayloadPath(m.repoRoot, name)
	if err := os.RemoveAll(payloadPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove payload: %w", err)
	}

	// Remove config directory
	configDir := filepath.Join(m.repoRoot, ".jvs", "worktrees", name)
	if err := os.RemoveAll(configDir); err != nil {
		return fmt.Errorf("remove config: %w", err)
	}

	// Audit log the removal
	if cfg != nil {
		auditPath := filepath.Join(m.repoRoot, ".jvs", "audit", "audit.jsonl")
		auditLogger := audit.NewFileAppender(auditPath)
		auditLogger.Append(model.EventTypeWorktreeRemove, name, "", map[string]any{
			"head_snapshot_id": string(cfg.HeadSnapshotID),
		})
	}

	return nil
}

// UpdateHead atomically updates the head snapshot ID for a worktree.
// This is used by restore to move to a different point in history.
func (m *Manager) UpdateHead(name string, snapshotID model.SnapshotID) error {
	cfg, err := repo.LoadWorktreeConfig(m.repoRoot, name)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.HeadSnapshotID = snapshotID
	return repo.WriteWorktreeConfig(m.repoRoot, name, cfg)
}

// SetLatest updates both head and latest snapshot IDs for a worktree.
// This is used by snapshot creation to mark a new latest state.
func (m *Manager) SetLatest(name string, snapshotID model.SnapshotID) error {
	cfg, err := repo.LoadWorktreeConfig(m.repoRoot, name)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.HeadSnapshotID = snapshotID
	cfg.LatestSnapshotID = snapshotID
	return repo.WriteWorktreeConfig(m.repoRoot, name, cfg)
}

// Fork creates a new worktree from a snapshot with content cloned.
// The new worktree will be at HEAD state (can create snapshots immediately).
func (m *Manager) Fork(snapshotID model.SnapshotID, name string, cloneFunc func(src, dst string) error) (*model.WorktreeConfig, error) {
	if err := pathutil.ValidateName(name); err != nil {
		return nil, err
	}

	// Check if already exists
	configPath := repo.WorktreeConfigPath(m.repoRoot, name)
	if _, err := os.Stat(configPath); err == nil {
		return nil, fmt.Errorf("worktree %s already exists", name)
	}

	// Create payload directory
	payloadPath := repo.WorktreePayloadPath(m.repoRoot, name)
	if err := os.MkdirAll(payloadPath, 0755); err != nil {
		return nil, fmt.Errorf("create payload directory: %w", err)
	}

	// Clone snapshot content to worktree
	snapshotDir := filepath.Join(m.repoRoot, ".jvs", "snapshots", string(snapshotID))
	if err := cloneFunc(snapshotDir, payloadPath); err != nil {
		os.RemoveAll(payloadPath)
		return nil, fmt.Errorf("clone snapshot content: %w", err)
	}

	// Create config directory
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		os.RemoveAll(payloadPath)
		return nil, fmt.Errorf("create config directory: %w", err)
	}

	// Create config with both head and latest set (HEAD state)
	cfg := &model.WorktreeConfig{
		Name:             name,
		CreatedAt:        time.Now().UTC(),
		BaseSnapshotID:   snapshotID,
		HeadSnapshotID:   snapshotID,
		LatestSnapshotID: snapshotID,
	}

	if err := repo.WriteWorktreeConfig(m.repoRoot, name, cfg); err != nil {
		os.RemoveAll(payloadPath)
		return nil, fmt.Errorf("write config: %w", err)
	}

	return cfg, nil
}
