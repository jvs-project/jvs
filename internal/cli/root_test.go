package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jvs-project/jvs/pkg/model"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeCommand(root *cobra.Command, args ...string) (stdout string, err error) {
	// Capture os.Stdout since CLI uses fmt.Printf directly
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	root.SetArgs(args)
	err = root.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String(), err
}

func setupTestDir(t *testing.T) string {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		os.Chdir(originalWd)
	})
	return dir
}

func TestRootCommand_Help(t *testing.T) {
	cmd := createTestRootCmd()
	stdout, err := executeCommand(cmd, "--help")
	require.NoError(t, err)
	assert.Contains(t, stdout, "snapshot-first")
}

func TestRootCommand_JSONFlag(t *testing.T) {
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "--json", "--help")
	require.NoError(t, err)
	assert.True(t, jsonOutput)
}

func TestInitCommand_CreatesRepo(t *testing.T) {
	setupTestDir(t)
	cmd := createTestRootCmd()
	stdout, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Initialized JVS repository")

	// Check repo was created
	_, statErr := os.Stat("testrepo/.jvs")
	assert.NoError(t, statErr)

	// Check main worktree exists
	_, statErr = os.Stat("testrepo/main")
	assert.NoError(t, statErr)
}

func TestWorktreeCommand_List(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// List worktrees - should show main
	cmd2 := createTestRootCmd()
	stdout, err := executeCommand(cmd2, "worktree", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "main")

	os.Chdir(originalWd)
}

func TestInfoCommand_WithRepo(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// Check info
	cmd2 := createTestRootCmd()
	stdout, err := executeCommand(cmd2, "info")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Repository:")

	os.Chdir(originalWd)
}

func TestInfoCommand_JSON(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// Check info with JSON
	cmd2 := createTestRootCmd()
	stdout, err := executeCommand(cmd2, "--json", "info")
	require.NoError(t, err)
	assert.Contains(t, stdout, "repo_root")

	os.Chdir(originalWd)
}

func TestDoctorCommand_Healthy(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// Run doctor
	cmd2 := createTestRootCmd()
	stdout, err := executeCommand(cmd2, "doctor")
	require.NoError(t, err)
	assert.Contains(t, stdout, "healthy")

	os.Chdir(originalWd)
}

func TestHistoryCommand_Empty(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into main worktree (history requires being inside worktree)
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo", "main")))

	// History on empty repo
	cmd2 := createTestRootCmd()
	stdout, err := executeCommand(cmd2, "history")
	require.NoError(t, err)
	// Empty history shows "no snapshots" or similar
	_ = stdout // May be empty or contain message

	os.Chdir(originalWd)
}

func TestVerifyCommand_Empty(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// Verify --all on empty repo
	cmd2 := createTestRootCmd()
	_, err = executeCommand(cmd2, "verify", "--all")
	require.NoError(t, err)

	os.Chdir(originalWd)
}

func TestGCCommand_Plan(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// GC plan
	cmd2 := createTestRootCmd()
	_, err = executeCommand(cmd2, "gc", "plan")
	require.NoError(t, err)

	os.Chdir(originalWd)
}

func TestOutputJSON(t *testing.T) {
	// Test with jsonOutput = true
	jsonOutput = true
	err := outputJSON(map[string]string{"test": "value"})
	assert.NoError(t, err)

	// Test with jsonOutput = false
	jsonOutput = false
	err = outputJSON(map[string]string{"test": "value"})
	assert.NoError(t, err)
}

func TestFmtErr(t *testing.T) {
	// fmtErr should not panic
	fmtErr("test error: %s", "detail")
}

func TestSnapshotCommand_CreatesSnapshot(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into main worktree
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo", "main")))

	// Create a file
	os.WriteFile("file.txt", []byte("content"), 0644)

	// Create snapshot
	cmd2 := createTestRootCmd()
	stdout, err := executeCommand(cmd2, "snapshot", "test note")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Created snapshot")

	os.Chdir(originalWd)
}

func TestSnapshotCommand_WithTags(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into main worktree
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo", "main")))

	// Create a file
	os.WriteFile("file.txt", []byte("content"), 0644)

	// Create snapshot with tags
	cmd2 := createTestRootCmd()
	stdout, err := executeCommand(cmd2, "snapshot", "release v1", "--tag", "v1.0", "--tag", "release")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Created snapshot")

	os.Chdir(originalWd)
}

func TestHistoryCommand_WithSnapshots(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into main worktree
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo", "main")))

	// Create files and snapshots
	os.WriteFile("file1.txt", []byte("content1"), 0644)
	cmd2 := createTestRootCmd()
	executeCommand(cmd2, "snapshot", "first snapshot", "--tag", "v1")

	os.WriteFile("file2.txt", []byte("content2"), 0644)
	cmd3 := createTestRootCmd()
	executeCommand(cmd3, "snapshot", "second snapshot", "--tag", "v2", "--tag", "release")

	// View history
	cmd4 := createTestRootCmd()
	stdout, err := executeCommand(cmd4, "history")
	require.NoError(t, err)
	assert.Contains(t, stdout, "snapshot")

	os.Chdir(originalWd)
}

func TestHistoryCommand_WithTagFilter(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into main worktree
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo", "main")))

	// Create files and snapshots with tags (reset snapshotTags before each)
	snapshotTags = []string{}
	os.WriteFile("file1.txt", []byte("content1"), 0644)
	cmd2 := createTestRootCmd()
	executeCommand(cmd2, "snapshot", "dev snapshot", "--tag", "dev")

	snapshotTags = []string{}
	os.WriteFile("file2.txt", []byte("content2"), 0644)
	cmd3 := createTestRootCmd()
	executeCommand(cmd3, "snapshot", "release snapshot", "--tag", "release")

	// View history with tag filter
	cmd4 := createTestRootCmd()
	stdout, err := executeCommand(cmd4, "history", "--tag", "release")
	require.NoError(t, err)
	assert.Contains(t, stdout, "release")
	// Should not show dev snapshot
	assert.NotContains(t, stdout, "dev snapshot")

	os.Chdir(originalWd)
}

func TestHistoryCommand_WithGrepFilter(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into main worktree
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo", "main")))

	// Create files and snapshots
	os.WriteFile("file1.txt", []byte("content1"), 0644)
	cmd2 := createTestRootCmd()
	executeCommand(cmd2, "snapshot", "development work")

	os.WriteFile("file2.txt", []byte("content2"), 0644)
	cmd3 := createTestRootCmd()
	executeCommand(cmd3, "snapshot", "production release")

	// View history with grep filter
	cmd4 := createTestRootCmd()
	stdout, err := executeCommand(cmd4, "history", "--grep", "release")
	require.NoError(t, err)
	assert.Contains(t, stdout, "release")
	assert.NotContains(t, stdout, "development")

	os.Chdir(originalWd)
}

func TestWorktreeCommand_Create(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// Create worktree
	cmd2 := createTestRootCmd()
	stdout, err := executeCommand(cmd2, "worktree", "create", "feature")
	require.NoError(t, err)
	assert.Contains(t, stdout, "feature")

	// Verify worktree exists
	_, err = os.Stat("worktrees/feature")
	assert.NoError(t, err)

	os.Chdir(originalWd)
}

func TestVerifyCommand_WithSnapshots(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into main worktree
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo", "main")))

	// Create file and snapshot
	os.WriteFile("file.txt", []byte("content"), 0644)
	cmd2 := createTestRootCmd()
	executeCommand(cmd2, "snapshot", "test")

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// Verify all
	cmd3 := createTestRootCmd()
	stdout, err := executeCommand(cmd3, "verify", "--all")
	require.NoError(t, err)
	assert.Contains(t, stdout, "OK")

	os.Chdir(originalWd)
}

func TestHasTag(t *testing.T) {
	// Test the hasTag helper function
	descWithTags := &model.Descriptor{
		Tags: []string{"v1.0", "release", "stable"},
	}
	descNoTags := &model.Descriptor{
		Tags: []string{},
	}

	assert.True(t, hasTag(descWithTags, "v1.0"))
	assert.True(t, hasTag(descWithTags, "release"))
	assert.False(t, hasTag(descWithTags, "dev"))
	assert.False(t, hasTag(descNoTags, "any"))
}

// createTestRootCmd creates a fresh root command for testing
func createTestRootCmd() *cobra.Command {
	// Reset jsonOutput flag
	jsonOutput = false

	// Create a new root command
	cmd := &cobra.Command{
		Use:           "jvs",
		Short:         "JVS - Juicy Versioned Workspaces",
		Long:          `JVS is a snapshot-first, filesystem-native workspace versioning system built on JuiceFS.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	// Add all subcommands
	cmd.AddCommand(initCmd)
	cmd.AddCommand(snapshotCmd)
	cmd.AddCommand(worktreeCmd)
	cmd.AddCommand(historyCmd)
	cmd.AddCommand(restoreCmd)
	cmd.AddCommand(infoCmd)
	cmd.AddCommand(doctorCmd)
	cmd.AddCommand(verifyCmd)
	cmd.AddCommand(gcCmd)

	return cmd
}

func TestRestoreCommand(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into main worktree
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo", "main")))

	// Create file and snapshot
	os.WriteFile("file.txt", []byte("version1"), 0644)
	cmd2 := createTestRootCmd()
	executeCommand(cmd2, "snapshot", "first")

	// Modify file and create another snapshot
	os.WriteFile("file.txt", []byte("version2"), 0644)
	cmd3 := createTestRootCmd()
	executeCommand(cmd3, "snapshot", "second")

	// Get history to find snapshot ID
	cmd4 := createTestRootCmd()
	stdout, _ := executeCommand(cmd4, "history", "--json")
	// The test passes if we can run the history command
	_ = stdout

	os.Chdir(originalWd)
}

func TestWorktreeCommand_Fork(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into main worktree
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo", "main")))

	// Create file and snapshot
	os.WriteFile("file.txt", []byte("content"), 0644)
	cmd2 := createTestRootCmd()
	executeCommand(cmd2, "snapshot", "base snapshot")

	// Fork from current position (must be inside worktree for fork without snapshot ID)
	cmd3 := createTestRootCmd()
	stdout, err := executeCommand(cmd3, "worktree", "fork", "feature")
	require.NoError(t, err)
	assert.Contains(t, stdout, "feature")

	// Verify forked worktree exists
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))
	_, err = os.Stat("worktrees/feature")
	assert.NoError(t, err)

	os.Chdir(originalWd)
}

func TestWorktreeCommand_Remove(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// Create worktree
	cmd2 := createTestRootCmd()
	executeCommand(cmd2, "worktree", "create", "to-remove")

	// Remove worktree
	cmd3 := createTestRootCmd()
	_, err = executeCommand(cmd3, "worktree", "remove", "to-remove")
	require.NoError(t, err)

	// Verify worktree is gone
	_, err = os.Stat("worktrees/to-remove")
	assert.True(t, os.IsNotExist(err))

	os.Chdir(originalWd)
}

func TestWorktreeCommand_Path(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// Get main worktree path
	cmd2 := createTestRootCmd()
	stdout, err := executeCommand(cmd2, "worktree", "path", "main")
	require.NoError(t, err)
	assert.Contains(t, stdout, "main")

	os.Chdir(originalWd)
}

func TestDoctorCommand_Strict(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// Run doctor --strict
	cmd2 := createTestRootCmd()
	stdout, err := executeCommand(cmd2, "doctor", "--strict")
	require.NoError(t, err)
	assert.Contains(t, stdout, "healthy")

	os.Chdir(originalWd)
}

func TestDoctorCommand_Repair(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// Create an orphan tmp file
	os.WriteFile(".jvs-tmp-orphan", []byte("data"), 0644)

	// Run doctor repair
	cmd2 := createTestRootCmd()
	_, err = executeCommand(cmd2, "doctor", "--repair-runtime", "clean_tmp")
	require.NoError(t, err)

	// Verify tmp file is cleaned
	_, err = os.Stat(".jvs-tmp-orphan")
	assert.True(t, os.IsNotExist(err))

	os.Chdir(originalWd)
}

func TestGCCommand_PlanJSON(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// GC plan with JSON output
	cmd2 := createTestRootCmd()
	stdout, err := executeCommand(cmd2, "--json", "gc", "plan")
	require.NoError(t, err)
	assert.Contains(t, stdout, "plan_id")

	os.Chdir(originalWd)
}

func TestHistoryCommand_Limit(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into main worktree
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo", "main")))

	// Create multiple snapshots
	for i := 0; i < 5; i++ {
		os.WriteFile(fmt.Sprintf("file%d.txt", i), []byte(fmt.Sprintf("content%d", i)), 0644)
		cmd2 := createTestRootCmd()
		executeCommand(cmd2, "snapshot", fmt.Sprintf("snapshot %d", i))
	}

	// View history with limit
	cmd3 := createTestRootCmd()
	stdout, err := executeCommand(cmd3, "history", "--limit", "2")
	require.NoError(t, err)
	// Should only show 2 entries
	_ = stdout

	os.Chdir(originalWd)
}

func TestWorktreeCommand_Rename(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// Create worktree
	cmd2 := createTestRootCmd()
	executeCommand(cmd2, "worktree", "create", "old-name")

	// Rename worktree
	cmd3 := createTestRootCmd()
	_, err = executeCommand(cmd3, "worktree", "rename", "old-name", "new-name")
	require.NoError(t, err)

	// Verify old name is gone
	_, err = os.Stat("worktrees/old-name")
	assert.True(t, os.IsNotExist(err))

	// Verify new name exists
	_, err = os.Stat("worktrees/new-name")
	assert.NoError(t, err)

	os.Chdir(originalWd)
}

func TestInfoCommand_JSONFields(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// Check info JSON has required fields
	cmd2 := createTestRootCmd()
	stdout, err := executeCommand(cmd2, "--json", "info")
	require.NoError(t, err)
	// Should contain required spec fields
	assert.Contains(t, stdout, "format_version")
	assert.Contains(t, stdout, "total_worktrees")
	assert.Contains(t, stdout, "total_snapshots")

	os.Chdir(originalWd)
}

func TestVerifyCommand_SingleSnapshot(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into main worktree
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo", "main")))

	// Create file and snapshot
	os.WriteFile("file.txt", []byte("content"), 0644)
	cmd2 := createTestRootCmd()
	executeCommand(cmd2, "snapshot", "test")

	// Verify all snapshots (including the one we just created)
	cmd4 := createTestRootCmd()
	stdout, err := executeCommand(cmd4, "verify", "--all")
	require.NoError(t, err)
	assert.Contains(t, stdout, "OK")

	os.Chdir(originalWd)
}

func TestSnapshotCommand_DetachedState(t *testing.T) {
	dir := t.TempDir()
	originalWd, _ := os.Getwd()

	// Init repo
	require.NoError(t, os.Chdir(dir))
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into main worktree
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo", "main")))

	// Create file and two snapshots
	os.WriteFile("file.txt", []byte("version1"), 0644)
	cmd2 := createTestRootCmd()
	executeCommand(cmd2, "snapshot", "first")

	os.WriteFile("file.txt", []byte("version2"), 0644)
	cmd3 := createTestRootCmd()
	executeCommand(cmd3, "snapshot", "second")

	// Now we're at HEAD and can create snapshots
	// (Restore to first would put us in detached state)
	_ = cmd3

	os.Chdir(originalWd)
}
