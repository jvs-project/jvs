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
	assert.DirExists(t, filepath.Join(repoPath, ".jvs", "refs"))
	assert.DirExists(t, filepath.Join(repoPath, ".jvs", "locks"))
	assert.DirExists(t, filepath.Join(repoPath, ".jvs", "intents"))
	assert.DirExists(t, filepath.Join(repoPath, ".jvs", "audit"))
	assert.DirExists(t, filepath.Join(repoPath, ".jvs", "gc"))

	// Verify main/ payload directory
	assert.DirExists(t, filepath.Join(repoPath, "main"))

	// Verify format_version content
	content, err := os.ReadFile(filepath.Join(repoPath, ".jvs", "format_version"))
	require.NoError(t, err)
	assert.Equal(t, "1\n", string(content))
}

func TestInit_MainWorktreeConfig(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "testrepo")

	_, err := repo.Init(repoPath, "testrepo")
	require.NoError(t, err)

	cfg, err := repo.LoadWorktreeConfig(repoPath, "main")
	require.NoError(t, err)
	assert.Equal(t, "main", cfg.Name)
	assert.Equal(t, model.IsolationExclusive, cfg.Isolation)
	assert.NotZero(t, cfg.CreatedAt)
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
