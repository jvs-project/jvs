package engine_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jvs-project/jvs/internal/engine"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJuiceFSEngine_Name(t *testing.T) {
	eng := engine.NewJuiceFSEngine()
	assert.Equal(t, model.EngineJuiceFSClone, eng.Name())
}

func TestJuiceFSEngine_CloneWithFallback(t *testing.T) {
	// This test verifies the engine falls back to copy when juicefs clone is not available
	src := t.TempDir()
	dst := t.TempDir()
	dstPath := filepath.Join(dst, "cloned")

	// Create source content
	os.WriteFile(filepath.Join(src, "file.txt"), []byte("hello"), 0644)

	eng := engine.NewJuiceFSEngine()
	result, err := eng.Clone(src, dstPath)

	require.NoError(t, err)
	// Should be degraded since juicefs clone will fail or not be available
	// The degradation could be "juicefs-not-available", "not-on-juicefs", or "juicefs-clone-failed"
	assert.True(t, result.Degraded)
	// At least one degradation should be reported
	assert.NotEmpty(t, result.Degradations)

	// Verify content was still copied (via fallback)
	content, err := os.ReadFile(filepath.Join(dstPath, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "hello", string(content))
}

func TestJuiceFSEngine_CloneNestedStructure(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	dstPath := filepath.Join(dst, "cloned")

	// Create nested structure
	os.MkdirAll(filepath.Join(src, "a", "b"), 0755)
	os.WriteFile(filepath.Join(src, "a", "file1.txt"), []byte("file1"), 0644)
	os.WriteFile(filepath.Join(src, "a", "b", "file2.txt"), []byte("file2"), 0644)

	eng := engine.NewJuiceFSEngine()
	result, err := eng.Clone(src, dstPath)

	require.NoError(t, err)
	assert.True(t, result.Degraded) // Falls back to copy

	// Verify nested content
	content, err := os.ReadFile(filepath.Join(dstPath, "a", "b", "file2.txt"))
	require.NoError(t, err)
	assert.Equal(t, "file2", string(content))
}

func TestJuiceFSEngine_CloneWithSymlinks(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	dstPath := filepath.Join(dst, "cloned")

	// Create content with symlink
	os.WriteFile(filepath.Join(src, "target.txt"), []byte("target"), 0644)
	require.NoError(t, os.Symlink("target.txt", filepath.Join(src, "link")))

	eng := engine.NewJuiceFSEngine()
	result, err := eng.Clone(src, dstPath)

	require.NoError(t, err)
	assert.True(t, result.Degraded) // Falls back to copy

	// Verify symlink preserved
	target, err := os.Readlink(filepath.Join(dstPath, "link"))
	require.NoError(t, err)
	assert.Equal(t, "target.txt", target)
}
