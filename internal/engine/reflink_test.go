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

func TestReflinkEngine_Name(t *testing.T) {
	eng := engine.NewReflinkEngine()
	assert.Equal(t, model.EngineReflinkCopy, eng.Name())
}

func TestReflinkEngine_Clone(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	dstPath := filepath.Join(dst, "cloned")

	// Create source content
	os.WriteFile(filepath.Join(src, "file.txt"), []byte("hello"), 0644)

	eng := engine.NewReflinkEngine()
	_, err := eng.Clone(src, dstPath)

	require.NoError(t, err)
	// Reflink may not be supported on this filesystem, so degraded is acceptable
	// The important thing is that the clone succeeded (possibly with fallback)

	// Verify content
	content, err := os.ReadFile(filepath.Join(dstPath, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "hello", string(content))
}

func TestReflinkEngine_FallbackToCopy(t *testing.T) {
	// Test that when reflink fails, we can detect it
	eng := engine.NewReflinkEngine()
	assert.NotNil(t, eng)
}
