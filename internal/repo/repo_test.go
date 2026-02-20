package repo_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit_CreatesDirectoryStructure(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "myrepo")

	r, err := repo.Init(repoPath, "myrepo")
	require.NoError(t, err)
	require.NotNil(t, r)

	// Verify .jvs/ structure
	assert.FileExists(t, filepath.Join(repoPath, ".jvs", "format_version"))
	assert.DirExists(t, filepath.Join(repoPath, ".jvs", "worktrees", "main"))
	assert.FileExists(t, filepath.Join(repoPath, ".jvs", "worktrees", "main", "config.json"))
	assert.DirExists(t, filepath.Join(repoPath, ".jvs", "snapshots"))
	assert.DirExists(t, filepath.Join(repoPath, ".jvs", "descriptors"))
	assert.DirExists(t, filepath.Join(repoPath, ".jvs", "intents"))
	assert.DirExists(t, filepath.Join(repoPath, ".jvs", "audit"))
	assert.DirExists(t, filepath.Join(repoPath, ".jvs", "gc"))

	// Verify main/ payload directory
	assert.DirExists(t, filepath.Join(repoPath, "main"))

	// Verify format_version content
	content, err := os.ReadFile(filepath.Join(repoPath, ".jvs", "format_version"))
	require.NoError(t, err)
	assert.Equal(t, "1\n", string(content))

	// Verify repo_id exists and is non-empty
	assert.FileExists(t, filepath.Join(repoPath, ".jvs", "repo_id"))
	repoIDContent, err := os.ReadFile(filepath.Join(repoPath, ".jvs", "repo_id"))
	require.NoError(t, err)
	assert.NotEmpty(t, string(repoIDContent))

	// Verify returned repo struct
	assert.Equal(t, repoPath, r.Root)
	assert.Equal(t, 1, r.FormatVersion)
	assert.NotEmpty(t, r.RepoID)
}

func TestInit_MainWorktreeConfig(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "testrepo")

	_, err := repo.Init(repoPath, "testrepo")
	require.NoError(t, err)

	cfg, err := repo.LoadWorktreeConfig(repoPath, "main")
	require.NoError(t, err)
	assert.Equal(t, "main", cfg.Name)
	assert.NotZero(t, cfg.CreatedAt)
}

func TestInit_InvalidName(t *testing.T) {
	dir := t.TempDir()

	_, err := repo.Init(dir, "../evil")
	assert.Error(t, err)

	_, err = repo.Init(dir, "name/with/slash")
	assert.Error(t, err)

	_, err = repo.Init(dir, "")
	assert.Error(t, err)
}

func TestInit_ExistingDirectory(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "existing")

	// Create directory first
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	// Init should still work
	_, err := repo.Init(repoPath, "existing")
	require.NoError(t, err)
}

func TestDiscover_FindsRepo(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "myrepo")
	_, err := repo.Init(repoPath, "myrepo")
	require.NoError(t, err)

	// Discover from repo root
	r, err := repo.Discover(repoPath)
	require.NoError(t, err)
	assert.Equal(t, repoPath, r.Root)

	// Discover from nested path
	nested := filepath.Join(repoPath, "main", "subdir")
	require.NoError(t, os.MkdirAll(nested, 0755))
	r, err = repo.Discover(nested)
	require.NoError(t, err)
	assert.Equal(t, repoPath, r.Root)
}

func TestDiscover_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := repo.Discover(dir)
	assert.Error(t, err)
}

func TestDiscoverWorktree_MainWorktree(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "myrepo")
	_, err := repo.Init(repoPath, "myrepo")
	require.NoError(t, err)

	// From main/ directory
	r, wtName, err := repo.DiscoverWorktree(filepath.Join(repoPath, "main"))
	require.NoError(t, err)
	assert.Equal(t, repoPath, r.Root)
	assert.Equal(t, "main", wtName)

	// From nested path in main/
	nested := filepath.Join(repoPath, "main", "deep", "path")
	require.NoError(t, os.MkdirAll(nested, 0755))
	r, wtName, err = repo.DiscoverWorktree(nested)
	require.NoError(t, err)
	assert.Equal(t, "main", wtName)
}

func TestDiscoverWorktree_NamedWorktree(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "myrepo")
	_, err := repo.Init(repoPath, "myrepo")
	require.NoError(t, err)

	// Create a named worktree
	wtPath := filepath.Join(repoPath, "worktrees", "feature")
	require.NoError(t, os.MkdirAll(wtPath, 0755))

	// Create config for worktree
	cfgDir := filepath.Join(repoPath, ".jvs", "worktrees", "feature")
	require.NoError(t, os.MkdirAll(cfgDir, 0755))

	// Discover from named worktree
	r, wtName, err := repo.DiscoverWorktree(wtPath)
	require.NoError(t, err)
	assert.Equal(t, repoPath, r.Root)
	assert.Equal(t, "feature", wtName)
}

func TestDiscoverWorktree_FromJvsDir(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "myrepo")
	_, err := repo.Init(repoPath, "myrepo")
	require.NoError(t, err)

	// From .jvs/ directory - should map to "main" as default
	r, wtName, err := repo.DiscoverWorktree(filepath.Join(repoPath, ".jvs"))
	require.NoError(t, err)
	assert.Equal(t, repoPath, r.Root)
	// .jvs is not a worktree, should return empty or error
	assert.Equal(t, "", wtName)
}

func TestWorktreeConfigPath(t *testing.T) {
	path := repo.WorktreeConfigPath("/repo", "main")
	assert.Equal(t, "/repo/.jvs/worktrees/main/config.json", path)

	path = repo.WorktreeConfigPath("/repo", "feature")
	assert.Equal(t, "/repo/.jvs/worktrees/feature/config.json", path)
}

func TestWorktreePayloadPath(t *testing.T) {
	path := repo.WorktreePayloadPath("/repo", "main")
	assert.Equal(t, "/repo/main", path)

	path = repo.WorktreePayloadPath("/repo", "feature")
	assert.Equal(t, "/repo/worktrees/feature", path)
}

func TestWriteAndLoadWorktreeConfig(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "testrepo")
	_, err := repo.Init(repoPath, "testrepo")
	require.NoError(t, err)

	// Load existing config
	cfg, err := repo.LoadWorktreeConfig(repoPath, "main")
	require.NoError(t, err)
	assert.Equal(t, "main", cfg.Name)

	// Modify and write
	cfg.HeadSnapshotID = "1708300800000-abc12345"
	err = repo.WriteWorktreeConfig(repoPath, "main", cfg)
	require.NoError(t, err)

	// Load again
	cfg2, err := repo.LoadWorktreeConfig(repoPath, "main")
	require.NoError(t, err)
	assert.Equal(t, model.SnapshotID("1708300800000-abc12345"), cfg2.HeadSnapshotID)
}

func TestLoadWorktreeConfig_NotFound(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "testrepo")
	_, err := repo.Init(repoPath, "testrepo")
	require.NoError(t, err)

	_, err = repo.LoadWorktreeConfig(repoPath, "nonexistent")
	assert.Error(t, err)
}

func TestDiscover_WrongFormatVersion(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "myrepo")
	_, err := repo.Init(repoPath, "myrepo")
	require.NoError(t, err)

	// Overwrite format_version with higher version
	formatFile := filepath.Join(repoPath, ".jvs", "format_version")
	err = os.WriteFile(formatFile, []byte("999\n"), 0644)
	require.NoError(t, err)

	_, err = repo.Discover(repoPath)
	assert.Error(t, err)
}

func TestDiscover_MissingFormatVersion(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "myrepo")
	_, err := repo.Init(repoPath, "myrepo")
	require.NoError(t, err)

	// Remove format_version
	formatFile := filepath.Join(repoPath, ".jvs", "format_version")
	os.Remove(formatFile)

	_, err = repo.Discover(repoPath)
	assert.Error(t, err)
}
