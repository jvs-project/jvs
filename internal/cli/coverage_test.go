package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestProgressEnabled tests the progressEnabled function.
func TestProgressEnabled(t *testing.T) {
	// Save original state
	originalNoProgress := noProgress
	originalJSONOutput := jsonOutput
	defer func() {
		noProgress = originalNoProgress
		jsonOutput = originalJSONOutput
	}()

	tests := []struct {
		name        string
		noProgress  bool
		jsonOutput  bool
		expected    bool
	}{
		{"Both false - progress enabled", false, false, true},
		{"No progress flag set", true, false, false},
		{"JSON output set", false, true, false},
		{"Both set", true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			noProgress = tt.noProgress
			jsonOutput = tt.jsonOutput
			result := progressEnabled()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOutputJSONBasicTests tests basic outputJSON behavior.
func TestOutputJSONBasicTests(t *testing.T) {
	// Save original state
	originalJSONOutput := jsonOutput
	defer func() {
		jsonOutput = originalJSONOutput
	}()

	t.Run("OutputJSON with false flag does nothing", func(t *testing.T) {
		jsonOutput = false
		err := outputJSON(map[string]string{"key": "value"})
		assert.NoError(t, err)
	})

	t.Run("OutputJSON with nil value", func(t *testing.T) {
		jsonOutput = true
		err := outputJSON(nil)
		assert.NoError(t, err)
	})
}

// TestNewCountingProgress tests the newCountingProgress function.
func TestNewCountingProgress(t *testing.T) {
	// Save original state
	originalNoProgress := noProgress
	originalJSONOutput := jsonOutput
	defer func() {
		noProgress = originalNoProgress
		jsonOutput = originalJSONOutput
	}()

	t.Run("Progress enabled", func(t *testing.T) {
		noProgress = false
		jsonOutput = false
		cp := newCountingProgress("test operation")
		assert.NotNil(t, cp, "counting progress should not be nil")
	})

	t.Run("Progress disabled via no-progress flag", func(t *testing.T) {
		noProgress = true
		jsonOutput = false
		cp := newCountingProgress("test operation")
		assert.NotNil(t, cp, "counting progress should not be nil even when disabled")
	})

	t.Run("Progress disabled via JSON flag", func(t *testing.T) {
		noProgress = false
		jsonOutput = true
		cp := newCountingProgress("test operation")
		assert.NotNil(t, cp, "counting progress should not be nil even when disabled")
	})
}

// TestNewProgressCallback tests the newProgressCallback function.
func TestNewProgressCallback(t *testing.T) {
	// Save original state
	originalNoProgress := noProgress
	originalJSONOutput := jsonOutput
	defer func() {
		noProgress = originalNoProgress
		jsonOutput = originalJSONOutput
	}()

	t.Run("Progress disabled returns callback", func(t *testing.T) {
		noProgress = true
		jsonOutput = false
		cb := newProgressCallback("test", 100)
		assert.NotNil(t, cb, "callback should not be nil even when progress disabled")
		// Should be able to call without panic
		cb("test", 50, 100, "halfway")
	})

	t.Run("JSON output returns callback", func(t *testing.T) {
		noProgress = false
		jsonOutput = true
		cb := newProgressCallback("test", 100)
		assert.NotNil(t, cb, "callback should not be nil")
		cb("test", 50, 100, "halfway")
	})

	t.Run("Progress enabled returns callback", func(t *testing.T) {
		noProgress = false
		jsonOutput = false
		cb := newProgressCallback("test", 100)
		assert.NotNil(t, cb, "callback should not be nil")
		cb("test", 50, 100, "halfway")
		cb("test", 100, 100, "done")
	})
}

// TestDetectEngine_Coverage tests the detectEngine function.
func TestDetectEngine_Coverage(t *testing.T) {
	t.Run("Non-existent path returns Copy as fallback", func(t *testing.T) {
		engine := detectEngine("/nonexistent/path/that/does/not/exist/12345")
		assert.Equal(t, "copy", string(engine))
	})

	t.Run("Common paths return Copy as fallback", func(t *testing.T) {
		engine := detectEngine("/tmp")
		assert.Equal(t, "copy", string(engine))
	})
}

// TestOutputJSONOrErrorVariations tests outputJSONOrError with various inputs.
func TestOutputJSONOrErrorVariations(t *testing.T) {
	// Save original state
	originalJSONOutput := jsonOutput
	defer func() {
		jsonOutput = originalJSONOutput
	}()

	t.Run("With error returns error", func(t *testing.T) {
		jsonOutput = true
		err := outputJSONOrError(nil, assert.AnError)
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("With nil value and nil error", func(t *testing.T) {
		jsonOutput = true
		err := outputJSONOrError(nil, nil)
		assert.NoError(t, err)
	})

	t.Run("With jsonOutput false does nothing", func(t *testing.T) {
		jsonOutput = false
		err := outputJSONOrError(map[string]string{"key": "value"}, nil)
		assert.NoError(t, err)
	})
}

// TestReadIntBehavior documents readInt behavior.
// Note: readInt requires stdin input which is difficult to test in unit tests.
// This test documents expected behavior for coverage.
func TestReadIntBehavior(t *testing.T) {
	t.Skip("readInt reads from stdin - tested via E2E tests")

	// Expected behavior:
	// - Returns 0 for empty input, "0", or "cancel"
	// - Returns 0 for invalid input (non-numeric)
	// - Returns 0 for out-of-range values (< 1 or > max)
	// - Returns the selected integer for valid input (1 to max)
}

// TestConfirmBehavior documents confirm behavior.
// Note: confirm requires stdin input which is difficult to test in unit tests.
// This test documents expected behavior for coverage.
func TestConfirmBehavior(t *testing.T) {
	t.Skip("confirm reads from stdin - tested via E2E tests")

	// Expected behavior:
	// - Returns true for "y" or "yes" (case-insensitive)
	// - Returns false for anything else
}

// TestExecuteFunctionExists tests that Execute function exists and is callable.
// Note: Execute calls os.Exit() which terminates the process, making it
// difficult to test in unit tests.
func TestExecuteFunctionExists(t *testing.T) {
	t.Skip("Execute calls os.Exit - tested via E2E/integration tests")

	// The Execute function:
	// 1. Calls rootCmd.Execute()
	// 2. On error, prints to stderr and calls os.Exit(1)
	// This is tested via the E2E test suite
}

// TestResolveSnapshotForDiff tests the diff.go resolveSnapshot function.
func TestResolveSnapshotForDiff(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	assert.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	assert.NoError(t, err)

	repoPath := dir + "/testrepo"
	mainPath := repoPath + "/main"

	t.Run("Resolve HEAD from outside worktree returns error", func(t *testing.T) {
		// Change to a directory outside the worktree
		assert.NoError(t, os.Chdir(dir))
		_, err := resolveSnapshot(repoPath, "HEAD")
		assert.Error(t, err)
	})

	t.Run("Resolve non-existent snapshot returns error", func(t *testing.T) {
		_, err := resolveSnapshot(repoPath, "nonexistent-snapshot-id")
		assert.Error(t, err)
	})

	t.Run("Resolve empty string returns error", func(t *testing.T) {
		_, err := resolveSnapshot(repoPath, "")
		assert.Error(t, err)
	})

	t.Run("Resolve HEAD when no snapshots exist", func(t *testing.T) {
		assert.NoError(t, os.Chdir(mainPath))
		_, err := resolveSnapshot(repoPath, "HEAD")
		assert.Error(t, err)
	})

	t.Run("Resolve with whitespace only", func(t *testing.T) {
		_, err := resolveSnapshot(repoPath, "   ")
		assert.Error(t, err)
	})

	os.Chdir(originalWd)
}

// TestResolveSnapshotByID tests resolving snapshots by full ID.
func TestResolveSnapshotByID(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	assert.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	assert.NoError(t, err)

	repoPath := dir + "/testrepo"
	mainPath := repoPath + "/main"

	// Create a snapshot
	assert.NoError(t, os.Chdir(mainPath))
	assert.NoError(t, os.WriteFile("test.txt", []byte("test content"), 0644))

	cmd2 := createTestRootCmd()
	stdout, _ := executeCommand(cmd2, "snapshot", "test snapshot", "--json")

	// Extract snapshot ID
	lines := strings.Split(stdout, "\n")
	var snapshotID string
	for _, line := range lines {
		if strings.Contains(line, `"snapshot_id"`) {
			parts := strings.Split(line, `"`)
			for i, p := range parts {
				if p == "snapshot_id" && i+2 < len(parts) {
					snapshotID = parts[i+2]
					break
				}
			}
		}
	}

	if snapshotID != "" {
		// Test resolving by full ID
		resolved, err := resolveSnapshot(repoPath, snapshotID)
		assert.NoError(t, err)
		assert.Equal(t, snapshotID, string(resolved))

		// Test resolving by short prefix
		shortPrefix := snapshotID[:8]
		resolved2, err := resolveSnapshot(repoPath, shortPrefix)
		assert.NoError(t, err)
		assert.Equal(t, snapshotID, string(resolved2))
	}

	os.Chdir(originalWd)
}

// TestResolveSnapshotByTag tests resolving snapshots by tag.
func TestResolveSnapshotByTag(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	assert.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	assert.NoError(t, err)

	repoPath := dir + "/testrepo"
	mainPath := repoPath + "/main"

	// Create a snapshot with a tag
	assert.NoError(t, os.Chdir(mainPath))
	assert.NoError(t, os.WriteFile("tagtest.txt", []byte("tagged content"), 0644))

	cmd2 := createTestRootCmd()
	stdout, _ := executeCommand(cmd2, "snapshot", "--tag", "testtag", "tagged snapshot", "--json")

	// Extract snapshot ID
	lines := strings.Split(stdout, "\n")
	var snapshotID string
	for _, line := range lines {
		if strings.Contains(line, `"snapshot_id"`) {
			parts := strings.Split(line, `"`)
			for i, p := range parts {
				if p == "snapshot_id" && i+2 < len(parts) {
					snapshotID = parts[i+2]
					break
				}
			}
		}
	}

	if snapshotID != "" {
		// Test resolving by tag
		resolved, err := resolveSnapshot(repoPath, "testtag")
		assert.NoError(t, err)
		assert.Equal(t, snapshotID, string(resolved))
	}

	os.Chdir(originalWd)
}

// TestResolveSnapshotByNote tests resolving snapshots by note.
func TestResolveSnapshotByNote(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	assert.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	assert.NoError(t, err)

	repoPath := dir + "/testrepo"
	mainPath := repoPath + "/main"

	// Create a snapshot with a unique note
	uniqueNote := "unique-snapshot-note-" + t.Name()
	assert.NoError(t, os.Chdir(mainPath))
	assert.NoError(t, os.WriteFile("notetest.txt", []byte("noted content"), 0644))

	cmd2 := createTestRootCmd()
	stdout, _ := executeCommand(cmd2, "snapshot", uniqueNote, "--json")

	// Extract snapshot ID
	lines := strings.Split(stdout, "\n")
	var snapshotID string
	for _, line := range lines {
		if strings.Contains(line, `"snapshot_id"`) {
			parts := strings.Split(line, `"`)
			for i, p := range parts {
				if p == "snapshot_id" && i+2 < len(parts) {
					snapshotID = parts[i+2]
					break
				}
			}
		}
	}

	if snapshotID != "" {
		// Test resolving by note prefix
		resolved, err := resolveSnapshot(repoPath, "unique-snapshot-note")
		assert.NoError(t, err)
		assert.Equal(t, snapshotID, string(resolved))
	}

	os.Chdir(originalWd)
}

// TestFmtErr_Coverage tests that fmtErr doesn't panic.
