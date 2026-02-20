package repo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jvs-project/jvs/pkg/errclass"
	"github.com/jvs-project/jvs/pkg/fsutil"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/jvs-project/jvs/pkg/pathutil"
	"github.com/jvs-project/jvs/pkg/uuidutil"
)

const (
	FormatVersion    = 1
	JVSDirName       = ".jvs"
	FormatVersionFile = "format_version"
	RepoIDFile       = "repo_id"
)

// Repo represents an initialized JVS repository.
type Repo struct {
	Root         string
	FormatVersion int
	RepoID       string
}

// Init creates a new JVS repository at the specified path.
func Init(path string, name string) (*Repo, error) {
	if err := pathutil.ValidateName(name); err != nil {
		return nil, err
	}

	// Create directory structure
	jvsDir := filepath.Join(path, JVSDirName)
	dirs := []string{
		jvsDir,
		filepath.Join(jvsDir, "worktrees", "main"),
		filepath.Join(jvsDir, "snapshots"),
		filepath.Join(jvsDir, "descriptors"),
		filepath.Join(jvsDir, "refs"),
		filepath.Join(jvsDir, "locks"),
		filepath.Join(jvsDir, "intents"),
		filepath.Join(jvsDir, "audit"),
		filepath.Join(jvsDir, "gc"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	// Write format_version
	if err := os.WriteFile(filepath.Join(jvsDir, FormatVersionFile), []byte("1\n"), 0644); err != nil {
		return nil, fmt.Errorf("write format_version: %w", err)
	}

	// Write repo_id
	repoID := uuidutil.NewV4()
	if err := os.WriteFile(filepath.Join(jvsDir, RepoIDFile), []byte(repoID+"\n"), 0644); err != nil {
		return nil, fmt.Errorf("write repo_id: %w", err)
	}

	// Create main/ payload directory
	mainDir := filepath.Join(path, "main")
	if err := os.MkdirAll(mainDir, 0755); err != nil {
		return nil, fmt.Errorf("create main directory: %w", err)
	}

	// Create worktrees/ payload directory
	worktreesPayload := filepath.Join(path, "worktrees")
	if err := os.MkdirAll(worktreesPayload, 0755); err != nil {
		return nil, fmt.Errorf("create worktrees directory: %w", err)
	}

	// Write main worktree config
	cfg := &model.WorktreeConfig{
		Name:      "main",
		CreatedAt: time.Now().UTC(),
		Isolation: model.IsolationExclusive,
	}
	if err := WriteWorktreeConfig(path, "main", cfg); err != nil {
		return nil, fmt.Errorf("write main config: %w", err)
	}

	// Fsync parent to ensure durability
	if err := fsutil.FsyncDir(path); err != nil {
		return nil, fmt.Errorf("fsync repo root: %w", err)
	}

	return &Repo{
		Root:          path,
		FormatVersion: FormatVersion,
		RepoID:        repoID,
	}, nil
}

// Discover walks up from cwd to find the repo root (directory containing .jvs/).
func Discover(cwd string) (*Repo, error) {
	path := cwd
	for {
		jvsDir := filepath.Join(path, JVSDirName)
		if info, err := os.Stat(jvsDir); err == nil && info.IsDir() {
			// Found .jvs/, read format_version
			version, err := readFormatVersion(jvsDir)
			if err != nil {
				return nil, err
			}
			if version > FormatVersion {
				return nil, errclass.ErrFormatUnsupported.WithMessagef(
					"format version %d > supported %d", version, FormatVersion)
			}
			repoID, _ := readRepoID(jvsDir)
			return &Repo{
				Root:          path,
				FormatVersion: version,
				RepoID:        repoID,
			}, nil
		}

		parent := filepath.Dir(path)
		if parent == path {
			// Reached root without finding .jvs/
			return nil, fmt.Errorf("no JVS repository found (no .jvs/ in parent directories)")
		}
		path = parent
	}
}

// DiscoverWorktree discovers the repo and maps cwd to a worktree name.
func DiscoverWorktree(cwd string) (*Repo, string, error) {
	r, err := Discover(cwd)
	if err != nil {
		return nil, "", err
	}

	// Get relative path from repo root
	rel, err := filepath.Rel(r.Root, cwd)
	if err != nil {
		return nil, "", fmt.Errorf("compute relative path: %w", err)
	}

	// Map to worktree name
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) == 0 {
		return r, "", nil
	}

	switch parts[0] {
	case "main":
		return r, "main", nil
	case "worktrees":
		if len(parts) >= 2 {
			return r, parts[1], nil
		}
		return r, "", nil
	case JVSDirName:
		// Inside .jvs/, not a worktree
		return r, "", nil
	default:
		return r, "", nil
	}
}

// WorktreeConfigPath returns the path to a worktree's config.json.
func WorktreeConfigPath(repoRoot, name string) string {
	return filepath.Join(repoRoot, JVSDirName, "worktrees", name, "config.json")
}

// WriteWorktreeConfig atomically writes a worktree config.
func WriteWorktreeConfig(repoRoot, name string, cfg *model.WorktreeConfig) error {
	path := WorktreeConfigPath(repoRoot, name)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal worktree config: %w", err)
	}
	return fsutil.AtomicWrite(path, data, 0644)
}

// LoadWorktreeConfig loads a worktree config.
func LoadWorktreeConfig(repoRoot, name string) (*model.WorktreeConfig, error) {
	path := WorktreeConfigPath(repoRoot, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read worktree config: %w", err)
	}
	var cfg model.WorktreeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse worktree config: %w", err)
	}
	return &cfg, nil
}

// WorktreePayloadPath returns the payload directory for a worktree.
func WorktreePayloadPath(repoRoot, name string) string {
	if name == "main" {
		return filepath.Join(repoRoot, "main")
	}
	return filepath.Join(repoRoot, "worktrees", name)
}

func readFormatVersion(jvsDir string) (int, error) {
	data, err := os.ReadFile(filepath.Join(jvsDir, FormatVersionFile))
	if err != nil {
		return 0, fmt.Errorf("read format_version: %w", err)
	}
	var version int
	if _, err := fmt.Sscanf(string(data), "%d", &version); err != nil {
		return 0, fmt.Errorf("parse format_version: %w", err)
	}
	return version, nil
}

func readRepoID(jvsDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(jvsDir, RepoIDFile))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
