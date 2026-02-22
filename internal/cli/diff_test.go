package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffCommand_TwoSnapshots(t *testing.T) {
	// This is an integration test that requires building the binary
	t.Skip("requires full build - manual testing only for now")
}

// TestDiff_SimpleIntegration is a manual integration test helper
func TestDiff_SimpleIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "testrepo")

	// Initialize repo
	cmd := exec.Command("/tmp/jvs", "init", "testrepo")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "init failed: %s", output)

	mainPath := filepath.Join(repoPath, "main")

	// Create initial file
	require.NoError(t, os.WriteFile(filepath.Join(mainPath, "file1.txt"), []byte("content1"), 0644))

	// Create first snapshot
	cmd = exec.Command("/tmp/jvs", "snapshot", "first snapshot")
	cmd.Dir = mainPath
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "first snapshot failed: %s", output)

	// Modify and add new file
	require.NoError(t, os.WriteFile(filepath.Join(mainPath, "file1.txt"), []byte("content1-modified"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(mainPath, "file2.txt"), []byte("new file"), 0644))

	// Create second snapshot
	cmd = exec.Command("/tmp/jvs", "snapshot", "second snapshot")
	cmd.Dir = mainPath
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "second snapshot failed: %s", output)

	// Get snapshot IDs
	cmd = exec.Command("/tmp/jvs", "history", "--json")
	cmd.Dir = mainPath
	output, err = cmd.Output()
	require.NoError(t, err, "history failed: %s", output)

	historyLines := strings.Split(strings.TrimSpace(string(output)), "\n")
	require.GreaterOrEqual(t, len(historyLines), 2)

	// Run diff command
	cmd = exec.Command("/tmp/jvs", "diff", "--stat")
	cmd.Dir = mainPath
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "diff failed: %s", output)

	diffOutput := string(output)
	// Should show added, removed, modified summary
	assert.Contains(t, diffOutput, "Added")
	assert.Contains(t, diffOutput, "Modified")
}
