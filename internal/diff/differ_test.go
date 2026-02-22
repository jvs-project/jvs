package diff

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffer_Diff_NoChanges(t *testing.T) {
	tmpDir := t.TempDir()
	differ := NewDiffer(tmpDir)

	// Create identical snapshots
	snap1 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap1")
	snap2 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap2")
	require.NoError(t, os.MkdirAll(snap1, 0755))
	require.NoError(t, os.MkdirAll(snap2, 0755))

	// Add same file to both
	content := []byte("hello world")
	require.NoError(t, os.WriteFile(filepath.Join(snap1, "file.txt"), content, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(snap2, "file.txt"), content, 0644))

	result, err := differ.Diff("snap1", "snap2")
	require.NoError(t, err)

	assert.Equal(t, 0, result.TotalAdded)
	assert.Equal(t, 0, result.TotalRemoved)
	assert.Equal(t, 0, result.TotalModified)
}

func TestDiffer_Diff_AddedFile(t *testing.T) {
	tmpDir := t.TempDir()
	differ := NewDiffer(tmpDir)

	snap1 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap1")
	snap2 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap2")
	require.NoError(t, os.MkdirAll(snap1, 0755))
	require.NoError(t, os.MkdirAll(snap2, 0755))

	// Add file only to snap2
	require.NoError(t, os.WriteFile(filepath.Join(snap2, "newfile.txt"), []byte("new"), 0644))

	result, err := differ.Diff("snap1", "snap2")
	require.NoError(t, err)

	assert.Equal(t, 1, result.TotalAdded)
	assert.Equal(t, "newfile.txt", result.Added[0].Path)
	assert.Equal(t, 0, result.TotalRemoved)
	assert.Equal(t, 0, result.TotalModified)
}

func TestDiffer_Diff_RemovedFile(t *testing.T) {
	tmpDir := t.TempDir()
	differ := NewDiffer(tmpDir)

	snap1 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap1")
	snap2 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap2")
	require.NoError(t, os.MkdirAll(snap1, 0755))
	require.NoError(t, os.MkdirAll(snap2, 0755))

	// Add file only to snap1
	require.NoError(t, os.WriteFile(filepath.Join(snap1, "removed.txt"), []byte("gone"), 0644))

	result, err := differ.Diff("snap1", "snap2")
	require.NoError(t, err)

	assert.Equal(t, 0, result.TotalAdded)
	assert.Equal(t, 1, result.TotalRemoved)
	assert.Equal(t, "removed.txt", result.Removed[0].Path)
	assert.Equal(t, 0, result.TotalModified)
}

func TestDiffer_Diff_ModifiedFile(t *testing.T) {
	tmpDir := t.TempDir()
	differ := NewDiffer(tmpDir)

	snap1 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap1")
	snap2 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap2")
	require.NoError(t, os.MkdirAll(snap1, 0755))
	require.NoError(t, os.MkdirAll(snap2, 0755))

	// Add same path with different content
	require.NoError(t, os.WriteFile(filepath.Join(snap1, "file.txt"), []byte("old"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(snap2, "file.txt"), []byte("new"), 0644))

	result, err := differ.Diff("snap1", "snap2")
	require.NoError(t, err)

	assert.Equal(t, 0, result.TotalAdded)
	assert.Equal(t, 0, result.TotalRemoved)
	assert.Equal(t, 1, result.TotalModified)
	assert.Equal(t, "file.txt", result.Modified[0].Path)
	assert.Equal(t, int64(3), result.Modified[0].OldSize)
	assert.Equal(t, int64(3), result.Modified[0].Size)
}

func TestDiffer_Diff_Symlink(t *testing.T) {
	tmpDir := t.TempDir()
	differ := NewDiffer(tmpDir)

	snap1 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap1")
	snap2 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap2")
	require.NoError(t, os.MkdirAll(snap1, 0755))
	require.NoError(t, os.MkdirAll(snap2, 0755))

	// Add symlink to snap1
	require.NoError(t, os.Symlink("target.txt", filepath.Join(snap1, "link")))

	// Add symlink to snap2 with different target
	require.NoError(t, os.Symlink("othertarget.txt", filepath.Join(snap2, "link")))

	result, err := differ.Diff("snap1", "snap2")
	require.NoError(t, err)

	assert.Equal(t, 0, result.TotalAdded)
	assert.Equal(t, 0, result.TotalRemoved)
	assert.Equal(t, 1, result.TotalModified)
	assert.True(t, result.Modified[0].IsSymlink)
}

func TestDiffer_Diff_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	differ := NewDiffer(tmpDir)

	snap1 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap1")
	snap2 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap2")
	require.NoError(t, os.MkdirAll(snap1, 0755))
	require.NoError(t, os.MkdirAll(snap2, 0755))

	// Add nested files
	require.NoError(t, os.MkdirAll(filepath.Join(snap1, "a", "b"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(snap1, "a", "b", "file.txt"), []byte("nested"), 0644))

	require.NoError(t, os.MkdirAll(filepath.Join(snap2, "a", "b"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(snap2, "a", "b", "file.txt"), []byte("modified"), 0644))

	result, err := differ.Diff("snap1", "snap2")
	require.NoError(t, err)

	assert.Equal(t, 1, result.TotalModified)
	assert.Equal(t, filepath.Join("a", "b", "file.txt"), result.Modified[0].Path)
}

func TestDiffer_Diff_EmptyFrom(t *testing.T) {
	tmpDir := t.TempDir()
	differ := NewDiffer(tmpDir)

	snap2 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap2")
	require.NoError(t, os.MkdirAll(snap2, 0755))

	// Add file to snap2
	require.NoError(t, os.WriteFile(filepath.Join(snap2, "file.txt"), []byte("content"), 0644))

	// Diff from empty (no fromID)
	result, err := differ.Diff("", "snap2")
	require.NoError(t, err)

	assert.Equal(t, 1, result.TotalAdded)
	assert.Equal(t, "file.txt", result.Added[0].Path)
	assert.Equal(t, 0, result.TotalRemoved)
	assert.Equal(t, 0, result.TotalModified)
}

func TestDiffer_Diff_SkipsReadyMarker(t *testing.T) {
	tmpDir := t.TempDir()
	differ := NewDiffer(tmpDir)

	snap1 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap1")
	snap2 := filepath.Join(tmpDir, ".jvs", "snapshots", "snap2")
	require.NoError(t, os.MkdirAll(snap1, 0755))
	require.NoError(t, os.MkdirAll(snap2, 0755))

	// Add .READY marker to snap2 (should be ignored)
	require.NoError(t, os.WriteFile(filepath.Join(snap2, ".READY"), []byte("{}"), 0644))

	result, err := differ.Diff("", "snap2")
	require.NoError(t, err)

	assert.Equal(t, 0, result.TotalAdded)
}

func TestDiffResult_FormatHuman(t *testing.T) {
	result := &DiffResult{
		FromSnapshotID: "snap1",
		ToSnapshotID:   "snap2",
		Added: []*Change{
			{Path: "newfile.txt", Type: ChangeAdded},
		},
		Removed: []*Change{
			{Path: "oldfile.txt", Type: ChangeRemoved},
		},
		Modified: []*Change{
			{Path: "changed.txt", Type: ChangeModified, OldSize: 100, Size: 200},
		},
		TotalAdded:    1,
		TotalRemoved:  1,
		TotalModified: 1,
	}

	output := result.FormatHuman()
	assert.Contains(t, output, "Added (1):")
	assert.Contains(t, output, "+ newfile.txt")
	assert.Contains(t, output, "Removed (1):")
	assert.Contains(t, output, "- oldfile.txt")
	assert.Contains(t, output, "Modified (1):")
	assert.Contains(t, output, "~ changed.txt")
	assert.Contains(t, output, "(100 -> 200 bytes)")
}

func TestDiffResult_FormatHuman_NoChanges(t *testing.T) {
	result := &DiffResult{
		FromSnapshotID: "snap1",
		ToSnapshotID:   "snap2",
		TotalAdded:     0,
		TotalRemoved:   0,
		TotalModified:  0,
	}

	output := result.FormatHuman()
	assert.Contains(t, output, "No changes.")
}
