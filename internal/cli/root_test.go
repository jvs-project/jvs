package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

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
	dir := setupTestDir(t)

	// Init repo
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
}

func TestWorktreeCommand_JSONList(t *testing.T) {
	dir := setupTestDir(t)

	// Init repo
	cmd := createTestRootCmd()
	_, err := executeCommand(cmd, "init", "testrepo")
	require.NoError(t, err)

	// Change into repo
	require.NoError(t, os.Chdir(filepath.Join(dir, "testrepo")))

	// List worktrees with JSON
	cmd2 := createTestRootCmd()
	stdout, err := executeCommand(cmd2, "--json", "worktree", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "main")
}

func TestInfoCommand_WithRepo(t *testing.T) {
	dir := setupTestDir(t)

	// Init repo
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
}

func TestInfoCommand_JSON(t *testing.T) {
	dir := setupTestDir(t)

	// Init repo
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
}

func TestDoctorCommand_Healthy(t *testing.T) {
	dir := setupTestDir(t)

	// Init repo
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
}

func TestHistoryCommand_Empty(t *testing.T) {
	dir := setupTestDir(t)

	// Init repo
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
	cmd.AddCommand(lockCmd)
	cmd.AddCommand(worktreeCmd)
	cmd.AddCommand(historyCmd)
	cmd.AddCommand(restoreCmd)
	cmd.AddCommand(refCmd)
	cmd.AddCommand(infoCmd)
	cmd.AddCommand(doctorCmd)
	cmd.AddCommand(verifyCmd)
	cmd.AddCommand(gcCmd)

	return cmd
}
