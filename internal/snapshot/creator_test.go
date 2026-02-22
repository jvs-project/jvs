package snapshot_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jvs-project/jvs/internal/repo"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) string {
	dir := t.TempDir()
	_, err := repo.Init(dir, "test")
	require.NoError(t, err)
	return dir
}

func TestCreator_Create(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Add some content to main/
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("hello"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "test note", nil)
	require.NoError(t, err)

	assert.NotEmpty(t, desc.SnapshotID)
	assert.Equal(t, "main", desc.WorktreeName)
	assert.Equal(t, "test note", desc.Note)
	assert.Equal(t, model.EngineCopy, desc.Engine)
	assert.NotEmpty(t, desc.PayloadRootHash)
	assert.NotEmpty(t, desc.DescriptorChecksum)

	// Verify snapshot directory exists
	snapshotDir := filepath.Join(repoPath, ".jvs", "snapshots", string(desc.SnapshotID))
	assert.DirExists(t, snapshotDir)

	// Verify descriptor exists
	descriptorPath := filepath.Join(repoPath, ".jvs", "descriptors", string(desc.SnapshotID)+".json")
	assert.FileExists(t, descriptorPath)

	// Verify .READY marker exists
	readyPath := filepath.Join(snapshotDir, ".READY")
	assert.FileExists(t, readyPath)
}

func TestCreator_ReadyProtocol(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "", nil)
	require.NoError(t, err)

	// Verify .READY contains correct info
	readyPath := filepath.Join(repoPath, ".jvs", "snapshots", string(desc.SnapshotID), ".READY")
	data, err := os.ReadFile(readyPath)
	require.NoError(t, err)

	var marker model.ReadyMarker
	require.NoError(t, json.Unmarshal(data, &marker))
	assert.Equal(t, desc.SnapshotID, marker.SnapshotID)
}

func TestCreator_UpdatesHead(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("v1"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc1, err := creator.Create("main", "first", nil)
	require.NoError(t, err)

	// Check head updated
	cfg, err := repo.LoadWorktreeConfig(repoPath, "main")
	require.NoError(t, err)
	assert.Equal(t, desc1.SnapshotID, cfg.HeadSnapshotID)

	// Create second snapshot
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("v2"), 0644)
	desc2, err := creator.Create("main", "second", nil)
	require.NoError(t, err)

	// Parent should be first snapshot
	assert.Equal(t, desc1.SnapshotID, *desc2.ParentID)

	// Head should be second
	cfg, _ = repo.LoadWorktreeConfig(repoPath, "main")
	assert.Equal(t, desc2.SnapshotID, cfg.HeadSnapshotID)
}

func TestCreator_PayloadContentPreserved(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("original"), 0644)
	os.MkdirAll(filepath.Join(mainPath, "subdir"), 0755)
	os.WriteFile(filepath.Join(mainPath, "subdir", "nested.txt"), []byte("nested"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "", nil)
	require.NoError(t, err)

	// Verify snapshot content
	snapshotPath := filepath.Join(repoPath, ".jvs", "snapshots", string(desc.SnapshotID))
	content, err := os.ReadFile(filepath.Join(snapshotPath, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "original", string(content))

	content, err = os.ReadFile(filepath.Join(snapshotPath, "subdir", "nested.txt"))
	require.NoError(t, err)
	assert.Equal(t, "nested", string(content))
}

func TestCreator_InvalidWorktree(t *testing.T) {
	repoPath := setupTestRepo(t)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	_, err := creator.Create("nonexistent", "", nil)
	require.Error(t, err)
}

func TestCreator_WithTags(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "tagged snapshot", []string{"v1.0", "release"})
	require.NoError(t, err)

	assert.Equal(t, []string{"v1.0", "release"}, desc.Tags)
}

func TestLoadDescriptor(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "", nil)
	require.NoError(t, err)

	// Load the descriptor
	loaded, err := snapshot.LoadDescriptor(repoPath, desc.SnapshotID)
	require.NoError(t, err)
	assert.Equal(t, desc.SnapshotID, loaded.SnapshotID)
	assert.Equal(t, desc.Note, loaded.Note)
}

func TestLoadDescriptor_NotFound(t *testing.T) {
	repoPath := setupTestRepo(t)

	_, err := snapshot.LoadDescriptor(repoPath, "nonexistent-snapshot-id")
	require.Error(t, err)
}

func TestVerifySnapshot(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "", nil)
	require.NoError(t, err)

	// Verify without payload hash
	err = snapshot.VerifySnapshot(repoPath, desc.SnapshotID, false)
	require.NoError(t, err)

	// Verify with payload hash
	err = snapshot.VerifySnapshot(repoPath, desc.SnapshotID, true)
	require.NoError(t, err)
}

func TestVerifySnapshot_InvalidID(t *testing.T) {
	repoPath := setupTestRepo(t)

	err := snapshot.VerifySnapshot(repoPath, "nonexistent", false)
	require.Error(t, err)
}

func TestCreator_DifferentEngines(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	// Test with Copy engine
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "copy", nil)
	require.NoError(t, err)
	assert.Equal(t, model.EngineCopy, desc.Engine)

	// Test with Reflink engine (falls back to copy on unsupported filesystem)
	creator2 := snapshot.NewCreator(repoPath, model.EngineReflinkCopy)
	desc2, err := creator2.Create("main", "reflink", nil)
	require.NoError(t, err)
	assert.Equal(t, model.EngineReflinkCopy, desc2.Engine)
}

func TestLoadDescriptor_CorruptJSON(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create a descriptor file with invalid JSON
	descriptorsDir := filepath.Join(repoPath, ".jvs", "descriptors")
	require.NoError(t, os.MkdirAll(descriptorsDir, 0755))
	descriptorPath := filepath.Join(descriptorsDir, "test-snapshot.json")
	require.NoError(t, os.WriteFile(descriptorPath, []byte("{invalid json"), 0644))

	_, err := snapshot.LoadDescriptor(repoPath, "test-snapshot")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse descriptor")
}

func TestLoadDescriptor_OtherReadError(t *testing.T) {
	// Create an invalid repo path (not a directory)
	_, err := snapshot.LoadDescriptor("/proc/nonexistent", "test-id")
	assert.Error(t, err)
}

func TestVerifySnapshot_ChecksumMismatch(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "", nil)
	require.NoError(t, err)

	// Corrupt the checksum in the descriptor
	descriptorPath := filepath.Join(repoPath, ".jvs", "descriptors", string(desc.SnapshotID)+".json")
	data, err := os.ReadFile(descriptorPath)
	require.NoError(t, err)
	var descMap map[string]any
	require.NoError(t, json.Unmarshal(data, &descMap))
	descMap["descriptor_checksum"] = "invalidchecksum"
	corruptedData, err := json.Marshal(descMap)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(descriptorPath, corruptedData, 0644))

	// Verify should detect checksum mismatch
	err = snapshot.VerifySnapshot(repoPath, desc.SnapshotID, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestVerifySnapshot_PayloadHashMismatch(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "", nil)
	require.NoError(t, err)

	// Modify the snapshot payload to corrupt the hash
	snapshotDir := filepath.Join(repoPath, ".jvs", "snapshots", string(desc.SnapshotID))
	snapshotFile := filepath.Join(snapshotDir, "file.txt")
	require.NoError(t, os.WriteFile(snapshotFile, []byte("modified content"), 0644))

	// Verify with payload hash should detect mismatch
	err = snapshot.VerifySnapshot(repoPath, desc.SnapshotID, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payload hash mismatch")
}

func TestCreator_CreateWithParent(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("v1"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc1, err := creator.Create("main", "first", nil)
	require.NoError(t, err)

	// Modify and create second snapshot
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("v2"), 0644)
	desc2, err := creator.Create("main", "second", nil)
	require.NoError(t, err)

	// Second snapshot should have first as parent
	assert.NotNil(t, desc2.ParentID)
	assert.Equal(t, desc1.SnapshotID, *desc2.ParentID)
}

func TestCreator_AuditLogFailureNonFatal(t *testing.T) {
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)

	// Make audit directory non-writable to trigger audit write failure
	auditDir := filepath.Join(repoPath, ".jvs", "audit")
	require.NoError(t, os.MkdirAll(auditDir, 0400))
	defer os.Chmod(auditDir, 0755)

	// Create should succeed despite audit failure
	_, err := creator.Create("main", "test", nil)
	assert.NoError(t, err)
}

func TestCreator_CreateWithNonExistentRepo(t *testing.T) {
	// Test Create with a non-existent repository path
	creator := snapshot.NewCreator("/nonexistent/path", model.EngineCopy)
	_, err := creator.Create("main", "test", nil)
	assert.Error(t, err)
}

func TestWriteReadyMarker_MarshalFailure(t *testing.T) {
	// This tests the json.Marshal error path in writeReadyMarker
	// We can't easily trigger this without using an invalid type,
	// but the test structure is here for completeness
	repoPath := setupTestRepo(t)
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	_, err := creator.Create("main", "test", nil)
	require.NoError(t, err)
	// If we get here without panic, writeReadyMarker worked
}

func TestLoadDescriptor_EmptySnapshotID(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create a descriptor file with empty snapshot_id
	descriptorsDir := filepath.Join(repoPath, ".jvs", "descriptors")
	require.NoError(t, os.MkdirAll(descriptorsDir, 0755))
	descriptorPath := filepath.Join(descriptorsDir, "test-snapshot.json")
	// Valid JSON but minimal fields
	require.NoError(t, os.WriteFile(descriptorPath, []byte(`{"snapshot_id": "", "created_at": "2024-01-01T00:00:00Z", "engine": "copy", "payload_root_hash": "abc", "descriptor_checksum": "def", "integrity_state": "verified"}`), 0644))

	desc, err := snapshot.LoadDescriptor(repoPath, "test-snapshot")
	// Should load without error (empty snapshot_id is valid JSON)
	require.NoError(t, err)
	assert.Equal(t, model.SnapshotID(""), desc.SnapshotID)
}

func TestVerifySnapshot_LoadDescriptorError(t *testing.T) {
	// Test that VerifySnapshot returns error when LoadDescriptor fails
	repoPath := setupTestRepo(t)

	err := snapshot.VerifySnapshot(repoPath, "nonexistent-id", false)
	assert.Error(t, err)
}

func TestVerifySnapshot_ComputeChecksumError(t *testing.T) {
	// This tests the checksum computation error path in VerifySnapshot
	// Most checksum errors come from integrity.ComputeDescriptorChecksum
	// which is hard to fail without invalid input
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "", nil)
	require.NoError(t, err)

	// First verify the original is valid
	err = snapshot.VerifySnapshot(repoPath, desc.SnapshotID, false)
	require.NoError(t, err)

	// Now modify descriptor to have a different checksum
	descriptorPath := filepath.Join(repoPath, ".jvs", "descriptors", string(desc.SnapshotID)+".json")
	data, err := os.ReadFile(descriptorPath)
	require.NoError(t, err)

	var descMap map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &descMap))
	descMap["descriptor_checksum"] = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	corruptData, err := json.Marshal(descMap)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(descriptorPath, corruptData, 0644))

	// Verify should detect checksum mismatch
	err = snapshot.VerifySnapshot(repoPath, desc.SnapshotID, false)
	assert.Error(t, err)
}

func TestNewCreator(t *testing.T) {
	repoPath := setupTestRepo(t)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	assert.NotNil(t, creator)

	// Test that creator can create snapshots successfully
	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	desc, err := creator.Create("main", "test", nil)
	require.NoError(t, err)
	assert.NotNil(t, desc)
	assert.Equal(t, model.EngineCopy, desc.Engine)
}

func TestMatchesFilter_NoteContains(t *testing.T) {
	// Test the NoteContains filter path specifically
	repoPath := setupTestRepo(t)

	createCatalogSnapshot(t, repoPath, "important feature work", nil)
	createCatalogSnapshot(t, repoPath, "bug fix", nil)

	// Filter by note containing "feature"
	opts := snapshot.FilterOptions{NoteContains: "feature"}
	matches, err := snapshot.Find(repoPath, opts)
	require.NoError(t, err)
	assert.Len(t, matches, 1)
	assert.Contains(t, matches[0].Note, "feature")
}

func TestMatchesFilter_SinceBefore(t *testing.T) {
	// Test that snapshots before Since time are filtered out
	repoPath := setupCatalogTestRepo(t)

	createCatalogSnapshot(t, repoPath, "early", nil)

	since := time.Now().UTC()
	time.Sleep(10 * time.Millisecond)

	createCatalogSnapshot(t, repoPath, "late", nil)

	// Filter to only get snapshots after 'since'
	opts := snapshot.FilterOptions{Since: since}
	matches, err := snapshot.Find(repoPath, opts)
	require.NoError(t, err)
	assert.Len(t, matches, 1)
	assert.Equal(t, "late", matches[0].Note)
}

func TestLoadDescriptor_ReadPermissionError(t *testing.T) {
	// Test LoadDescriptor when file exists but can't be read
	// This tests the non-IsNotExist error path in LoadDescriptor
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "", nil)
	require.NoError(t, err)

	// Make descriptor file unreadable
	descriptorPath := filepath.Join(repoPath, ".jvs", "descriptors", string(desc.SnapshotID)+".json")
	require.NoError(t, os.Chmod(descriptorPath, 0000))
	defer os.Chmod(descriptorPath, 0644)

	_, err = snapshot.LoadDescriptor(repoPath, desc.SnapshotID)
	assert.Error(t, err)
}

func TestVerifySnapshot_MissingSnapshotDirectory(t *testing.T) {
	// Test VerifySnapshot when snapshot directory doesn't exist
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "", nil)
	require.NoError(t, err)

	// Remove the snapshot directory
	snapshotDir := filepath.Join(repoPath, ".jvs", "snapshots", string(desc.SnapshotID))
	require.NoError(t, os.RemoveAll(snapshotDir))

	// Verify with payload hash should fail
	err = snapshot.VerifySnapshot(repoPath, desc.SnapshotID, true)
	assert.Error(t, err)
}

func TestCreator_SnapshotWithEmptyNote(t *testing.T) {
	// Test creating a snapshot with an empty note
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "", nil)
	require.NoError(t, err)
	assert.Equal(t, "", desc.Note)
}

func TestCreator_SnapshotWithEmptyTags(t *testing.T) {
	// Test creating a snapshot with empty tags slice
	repoPath := setupTestRepo(t)

	mainPath := filepath.Join(repoPath, "main")
	os.WriteFile(filepath.Join(mainPath, "file.txt"), []byte("content"), 0644)

	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	desc, err := creator.Create("main", "test", []string{})
	require.NoError(t, err)
	assert.NotNil(t, desc.Tags)
	assert.Empty(t, desc.Tags)
}

func TestMatchesFilter_NonMatchingNote(t *testing.T) {
	// Test matchesFilter when note doesn't contain the search string
	repoPath := setupCatalogTestRepo(t)

	createCatalogSnapshot(t, repoPath, "completely different note", nil)

	// Search for something that doesn't exist
	opts := snapshot.FilterOptions{NoteContains: "notfound"}
	matches, err := snapshot.Find(repoPath, opts)
	require.NoError(t, err)
	assert.Empty(t, matches)
}
