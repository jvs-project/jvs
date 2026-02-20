package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/internal/worktree"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCatalogTestRepo(t *testing.T) string {
	dir := t.TempDir()
	_, err := repo.Init(dir, "test")
	require.NoError(t, err)
	return dir
}

func createCatalogSnapshot(t *testing.T, repoPath, note string, tags []string) *model.Descriptor {
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", note, tags)
	require.NoError(t, err)
	return desc
}

func TestListAll(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	// Empty repo
	all, err := snapshot.ListAll(repoPath)
	require.NoError(t, err)
	assert.Empty(t, all)

	// Create two snapshots
	desc1 := createCatalogSnapshot(t, repoPath, "first", nil)
	desc2 := createCatalogSnapshot(t, repoPath, "second", nil)

	all, err = snapshot.ListAll(repoPath)
	require.NoError(t, err)
	assert.Len(t, all, 2)

	// Newest first
	assert.Equal(t, desc2.SnapshotID, all[0].SnapshotID)
	assert.Equal(t, desc1.SnapshotID, all[1].SnapshotID)
}

func TestListAll_SortedByTime(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	// Create snapshots with time gap
	desc1 := createCatalogSnapshot(t, repoPath, "first", nil)
	time.Sleep(10 * time.Millisecond)
	desc2 := createCatalogSnapshot(t, repoPath, "second", nil)
	time.Sleep(10 * time.Millisecond)
	desc3 := createCatalogSnapshot(t, repoPath, "third", nil)

	all, err := snapshot.ListAll(repoPath)
	require.NoError(t, err)
	assert.Len(t, all, 3)

	// Should be newest first
	assert.Equal(t, desc3.SnapshotID, all[0].SnapshotID)
	assert.Equal(t, desc2.SnapshotID, all[1].SnapshotID)
	assert.Equal(t, desc1.SnapshotID, all[2].SnapshotID)
}

func TestFind_ByNote(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	createCatalogSnapshot(t, repoPath, "release v1", nil)
	createCatalogSnapshot(t, repoPath, "release v2", nil)
	createCatalogSnapshot(t, repoPath, "wip feature", nil)

	opts := snapshot.FilterOptions{NoteContains: "release"}
	matches, err := snapshot.Find(repoPath, opts)
	require.NoError(t, err)
	assert.Len(t, matches, 2)
}

func TestFind_ByTag(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	createCatalogSnapshot(t, repoPath, "first", []string{"v1.0"})
	createCatalogSnapshot(t, repoPath, "second", []string{"v1.1", "release"})
	createCatalogSnapshot(t, repoPath, "third", []string{"wip"})

	opts := snapshot.FilterOptions{HasTag: "release"}
	matches, err := snapshot.Find(repoPath, opts)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, "release", matches[0].Tags[1])
}

func TestFind_ByWorktree(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	// Create snapshots in main
	createCatalogSnapshot(t, repoPath, "main-snapshot", nil)

	// Create another worktree and snapshot
	wtMgr := worktree.NewManager(repoPath)
	cfg, err := wtMgr.Create("feature", nil)
	require.NoError(t, err)

	// Add content to feature worktree
	featurePath := wtMgr.Path("feature")
	os.WriteFile(filepath.Join(featurePath, "file.txt"), []byte("feature-content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	_, err = creator.Create("feature", "feature-snapshot", nil)
	require.NoError(t, err)
	_ = cfg

	// Filter by main worktree
	opts := snapshot.FilterOptions{WorktreeName: "main"}
	matches, err := snapshot.Find(repoPath, opts)
	require.NoError(t, err)
	assert.Len(t, matches, 1)
	assert.Equal(t, "main", matches[0].WorktreeName)
}

func TestFind_ByTimeRange(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	before := time.Now().UTC()
	time.Sleep(10 * time.Millisecond)
	createCatalogSnapshot(t, repoPath, "first", nil)
	middle := time.Now().UTC()
	time.Sleep(10 * time.Millisecond)
	createCatalogSnapshot(t, repoPath, "second", nil)
	_ = time.Now().UTC() // after

	// Find snapshots between before and middle
	opts := snapshot.FilterOptions{Since: before, Until: middle}
	matches, err := snapshot.Find(repoPath, opts)
	require.NoError(t, err)
	assert.Len(t, matches, 1)
	assert.Equal(t, "first", matches[0].Note)
}

func TestFind_CombinedFilters(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	createCatalogSnapshot(t, repoPath, "release v1", []string{"release"})
	createCatalogSnapshot(t, repoPath, "release v2", []string{"release", "stable"})
	createCatalogSnapshot(t, repoPath, "wip", []string{"wip"})

	// Filter by both note and tag
	opts := snapshot.FilterOptions{NoteContains: "release", HasTag: "stable"}
	matches, err := snapshot.Find(repoPath, opts)
	require.NoError(t, err)
	assert.Len(t, matches, 1)
	assert.Equal(t, "release v2", matches[0].Note)
}

func TestFind_EmptyResult(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	createCatalogSnapshot(t, repoPath, "test", nil)

	opts := snapshot.FilterOptions{HasTag: "nonexistent"}
	matches, err := snapshot.Find(repoPath, opts)
	require.NoError(t, err)
	assert.Empty(t, matches)
}

func TestFindOne_ByNotePrefix(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	createCatalogSnapshot(t, repoPath, "release v1.0", nil)
	createCatalogSnapshot(t, repoPath, "wip feature", nil)

	desc, err := snapshot.FindOne(repoPath, "release")
	require.NoError(t, err)
	assert.Equal(t, "release v1.0", desc.Note)
}

func TestFindOne_ByTag(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	createCatalogSnapshot(t, repoPath, "first", []string{"v1.0"})
	createCatalogSnapshot(t, repoPath, "second", []string{"wip"})

	desc, err := snapshot.FindOne(repoPath, "v1.0")
	require.NoError(t, err)
	assert.Equal(t, "first", desc.Note)
}

func TestFindOne_ByTagPrefix(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	createCatalogSnapshot(t, repoPath, "test", []string{"v1.0.0-alpha"})

	// Should find by tag prefix
	desc, err := snapshot.FindOne(repoPath, "v1.0")
	require.NoError(t, err)
	assert.Equal(t, "test", desc.Note)
}

func TestFindOne_BySnapshotIDPrefix(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	desc := createCatalogSnapshot(t, repoPath, "test", nil)

	// Use short ID (first 8 chars)
	shortID := desc.SnapshotID.ShortID()
	found, err := snapshot.FindOne(repoPath, shortID)
	require.NoError(t, err)
	assert.Equal(t, desc.SnapshotID, found.SnapshotID)

	// Use full ID
	found2, err := snapshot.FindOne(repoPath, string(desc.SnapshotID))
	require.NoError(t, err)
	assert.Equal(t, desc.SnapshotID, found2.SnapshotID)
}

func TestFindOne_Ambiguous(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	createCatalogSnapshot(t, repoPath, "release v1", nil)
	createCatalogSnapshot(t, repoPath, "release v2", nil)

	// Both notes start with "release"
	_, err := snapshot.FindOne(repoPath, "release")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous")
}

func TestFindOne_AmbiguousTags(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	createCatalogSnapshot(t, repoPath, "first", []string{"v1", "release"})
	createCatalogSnapshot(t, repoPath, "second", []string{"v2", "release"})

	// Both have "release" tag
	_, err := snapshot.FindOne(repoPath, "release")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous")
}

func TestFindOne_NotFound(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	_, err := snapshot.FindOne(repoPath, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no snapshot found")
}

func TestFindByTag(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	createCatalogSnapshot(t, repoPath, "old", []string{"release"})
	createCatalogSnapshot(t, repoPath, "new", []string{"release"})

	// FindByTag returns the latest (newest first in ListAll)
	desc, err := snapshot.FindByTag(repoPath, "release")
	require.NoError(t, err)
	assert.Equal(t, "new", desc.Note)
}

func TestFindByTag_NotFound(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	_, err := snapshot.FindByTag(repoPath, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no snapshot found")
}

func TestListAll_EmptySnapshotsDir(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	// No snapshots yet
	all, err := snapshot.ListAll(repoPath)
	require.NoError(t, err)
	assert.Empty(t, all)
}

func TestListAll_HandlesCorruptDescriptor(t *testing.T) {
	repoPath := setupCatalogTestRepo(t)

	// Create a valid snapshot
	desc := createCatalogSnapshot(t, repoPath, "valid", nil)

	// Create a corrupt snapshot directory (no descriptor)
	snapshotsDir := filepath.Join(repoPath, ".jvs", "snapshots")
	corruptDir := filepath.Join(snapshotsDir, "0000000000000-corrupt")
	require.NoError(t, os.Mkdir(corruptDir, 0755))

	// ListAll should skip the corrupt one
	all, err := snapshot.ListAll(repoPath)
	require.NoError(t, err)
	assert.Len(t, all, 1)
	assert.Equal(t, desc.SnapshotID, all[0].SnapshotID)
}
