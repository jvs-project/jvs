package integrity_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jvs-project/jvs/internal/integrity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputePayloadRootHash_Deterministic(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)

	hash1, err := integrity.ComputePayloadRootHash(dir)
	require.NoError(t, err)
	hash2, err := integrity.ComputePayloadRootHash(dir)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2, "hash must be deterministic")
}

func TestComputePayloadRootHash_DetectsContentChange(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	os.WriteFile(file, []byte("original"), 0644)

	hash1, _ := integrity.ComputePayloadRootHash(dir)
	os.WriteFile(file, []byte("modified"), 0644)
	hash2, _ := integrity.ComputePayloadRootHash(dir)

	assert.NotEqual(t, hash1, hash2, "content change must produce different hash")
}

func TestComputePayloadRootHash_DetectsPermissionChange(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	os.WriteFile(file, []byte("content"), 0644)

	hash1, _ := integrity.ComputePayloadRootHash(dir)
	os.Chmod(file, 0755)
	hash2, _ := integrity.ComputePayloadRootHash(dir)

	assert.NotEqual(t, hash1, hash2, "permission change must produce different hash")
}

func TestComputePayloadRootHash_IncludesDirectory(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "subdir")
	require.NoError(t, os.Mkdir(subdir, 0755))

	hash, err := integrity.ComputePayloadRootHash(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestComputePayloadRootHash_IncludesSymlink(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "target.txt"), []byte("target"), 0644)
	require.NoError(t, os.Symlink("target.txt", filepath.Join(dir, "link")))

	hash, err := integrity.ComputePayloadRootHash(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestComputePayloadRootHash_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	hash, err := integrity.ComputePayloadRootHash(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, hash, "empty dir should still produce a hash")
}
