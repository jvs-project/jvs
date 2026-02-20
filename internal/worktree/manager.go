package worktree

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
			continue // skip malformed
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

	return nil
}

// UpdateHead atomically updates the head snapshot ID for a worktree.
func (m *Manager) UpdateHead(name string, snapshotID model.SnapshotID) error {
	cfg, err := repo.LoadWorktreeConfig(m.repoRoot, name)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.HeadSnapshotID = snapshotID
	return repo.WriteWorktreeConfig(m.repoRoot, name, cfg)
}
