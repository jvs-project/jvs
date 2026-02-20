package fsutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jvs-project/jvs/pkg/fsutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAtomicWrite_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	data := []byte(`{"key": "value"}`)

	err := fsutil.AtomicWrite(path, data, 0644)
	require.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, data, content)
}

func TestAtomicWrite_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	os.WriteFile(path, []byte("old"), 0644)

	err := fsutil.AtomicWrite(path, []byte("new"), 0644)
	require.NoError(t, err)

	content, _ := os.ReadFile(path)
	assert.Equal(t, "new", string(content))
}

func TestAtomicWrite_NoTmpLeftOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	fsutil.AtomicWrite(path, []byte("data"), 0644)

	entries, _ := os.ReadDir(dir)
	assert.Len(t, entries, 1, "only the target file should exist")
}

func TestRenameAndSync(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	os.WriteFile(src, []byte("data"), 0644)

	err := fsutil.RenameAndSync(src, dst)
	require.NoError(t, err)

	assert.NoFileExists(t, src)
	content, _ := os.ReadFile(dst)
	assert.Equal(t, "data", string(content))
}

func TestFsyncDir(t *testing.T) {
	dir := t.TempDir()
	err := fsutil.FsyncDir(dir)
	assert.NoError(t, err)
}
