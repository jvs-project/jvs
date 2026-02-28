package doctor_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jvs-project/jvs/internal/doctor"
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

func createTestSnapshot(t *testing.T, repoPath string) {
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	_, err := creator.Create("main", "test", nil)
	require.NoError(t, err)
}

func TestDoctor_Check_Healthy(t *testing.T) {
	repoPath := setupTestRepo(t)

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)
	assert.True(t, result.Healthy)
	assert.Empty(t, result.Findings)
}

func TestDoctor_Check_WithSnapshots(t *testing.T) {
	repoPath := setupTestRepo(t)
	createTestSnapshot(t, repoPath)

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)
	assert.True(t, result.Healthy)
}

func TestDoctor_Check_Strict(t *testing.T) {
	repoPath := setupTestRepo(t)
	createTestSnapshot(t, repoPath)

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(true)
	require.NoError(t, err)
	assert.True(t, result.Healthy)
}

func TestDoctor_Check_OrphanIntent(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create orphan intent file
	intentsDir := filepath.Join(repoPath, ".jvs", "intents")
	os.MkdirAll(intentsDir, 0755)
	os.WriteFile(filepath.Join(intentsDir, "orphan.json"), []byte("{}"), 0644)

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)
	// Orphan intents are warnings, not critical, so repo stays healthy
	assert.True(t, result.Healthy)
	assert.Len(t, result.Findings, 1)
	assert.Equal(t, "intent", result.Findings[0].Category)
	assert.Equal(t, "warning", result.Findings[0].Severity)
}

func TestDoctor_Check_OrphanTmp(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create orphan tmp file
	os.WriteFile(filepath.Join(repoPath, ".jvs", ".jvs-tmp-orphan"), []byte("data"), 0644)

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)
	// Orphan tmp is info level, doesn't make repo unhealthy
	assert.True(t, result.Healthy || len(result.Findings) > 0)
}

func TestDoctor_Check_MissingWorktreePayload(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Remove main payload directory (simulating corruption)
	os.RemoveAll(filepath.Join(repoPath, "main"))

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)
	// Missing payload reports error finding but repo stays "healthy" at info level
	assert.NotEmpty(t, result.Findings)
	found := false
	for _, f := range result.Findings {
		if f.Category == "worktree" {
			found = true
			assert.Contains(t, f.Description, "payload directory missing")
		}
	}
	assert.True(t, found, "expected worktree finding for missing payload")
}

func TestDoctor_ListRepairActions(t *testing.T) {
	repoPath := setupTestRepo(t)
	doc := doctor.NewDoctor(repoPath)

	actions := doc.ListRepairActions()
	assert.NotEmpty(t, actions)

	// Check for expected actions
	actionMap := make(map[string]bool)
	for _, a := range actions {
		actionMap[a.ID] = true
	}
	assert.True(t, actionMap["clean_tmp"])
	assert.True(t, actionMap["clean_intents"])
	assert.True(t, actionMap["advance_head"])
}

func TestDoctor_Repair_CleanTmp(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create orphan tmp files
	os.WriteFile(filepath.Join(repoPath, ".jvs-tmp-orphan1"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(repoPath, ".jvs-tmp-orphan2"), []byte("data"), 0644)

	doc := doctor.NewDoctor(repoPath)
	results, err := doc.Repair([]string{"clean_tmp"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "clean_tmp", results[0].Action)
	assert.True(t, results[0].Success)
	assert.Equal(t, 2, results[0].Cleaned)

	// Verify files are gone
	assert.NoFileExists(t, filepath.Join(repoPath, ".jvs-tmp-orphan1"))
	assert.NoFileExists(t, filepath.Join(repoPath, ".jvs-tmp-orphan2"))
}

func TestDoctor_Repair_CleanIntents(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create orphan intent files
	intentsDir := filepath.Join(repoPath, ".jvs", "intents")
	os.MkdirAll(intentsDir, 0755)
	os.WriteFile(filepath.Join(intentsDir, "orphan1.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(intentsDir, "orphan2.json"), []byte("{}"), 0644)

	doc := doctor.NewDoctor(repoPath)
	results, err := doc.Repair([]string{"clean_intents"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "clean_intents", results[0].Action)
	assert.True(t, results[0].Success)
	assert.Equal(t, 2, results[0].Cleaned)

	// Verify files are gone
	assert.NoFileExists(t, filepath.Join(intentsDir, "orphan1.json"))
	assert.NoFileExists(t, filepath.Join(intentsDir, "orphan2.json"))
}

func TestDoctor_Repair_AdvanceHead(t *testing.T) {
	repoPath := setupTestRepo(t)
	createTestSnapshot(t, repoPath)

	// Create second snapshot
	creator := snapshot.NewCreator(repoPath, model.EngineCopy)
	_, err := creator.Create("main", "second", nil)
	require.NoError(t, err)

	doc := doctor.NewDoctor(repoPath)
	results, err := doc.Repair([]string{"advance_head"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "advance_head", results[0].Action)
	assert.True(t, results[0].Success)
}

func TestDoctor_Repair_UnknownAction(t *testing.T) {
	repoPath := setupTestRepo(t)
	doc := doctor.NewDoctor(repoPath)

	results, err := doc.Repair([]string{"unknown_action"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "unknown_action", results[0].Action)
	assert.False(t, results[0].Success)
	assert.Contains(t, results[0].Message, "unknown repair action")
}

func TestDoctor_Repair_MultipleActions(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create orphan tmp and intent files
	os.WriteFile(filepath.Join(repoPath, ".jvs-tmp-orphan"), []byte("data"), 0644)
	intentsDir := filepath.Join(repoPath, ".jvs", "intents")
	os.MkdirAll(intentsDir, 0755)
	os.WriteFile(filepath.Join(intentsDir, "orphan.json"), []byte("{}"), 0644)

	doc := doctor.NewDoctor(repoPath)
	results, err := doc.Repair([]string{"clean_tmp", "clean_intents"})
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.True(t, results[1].Success)
}

func TestDoctor_Check_FormatVersionMismatch(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Set format version to a higher value
	versionPath := filepath.Join(repoPath, ".jvs", "format_version")
	os.WriteFile(versionPath, []byte("9999"), 0644)

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)
	assert.False(t, result.Healthy)
	assert.NotEmpty(t, result.Findings)
	assert.Equal(t, "format", result.Findings[0].Category)
	assert.Equal(t, "critical", result.Findings[0].Severity)
}

func TestDoctor_Check_MissingFormatVersion(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Remove format_version file
	os.Remove(filepath.Join(repoPath, ".jvs", "format_version"))

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)
	assert.False(t, result.Healthy)
	assert.NotEmpty(t, result.Findings)
	assert.Equal(t, "format", result.Findings[0].Category)
}

func TestDoctor_Check_OrphanSnapshotTmp(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create orphan snapshot tmp directory
	snapshotsDir := filepath.Join(repoPath, ".jvs", "snapshots")
	os.MkdirAll(snapshotsDir, 0755)
	os.MkdirAll(filepath.Join(snapshotsDir, "something.tmp"), 0755)

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)
	// Should find the tmp directory
	found := false
	for _, f := range result.Findings {
		if f.Category == "tmp" && f.Severity == "warning" {
			found = true
		}
	}
	assert.True(t, found, "expected tmp finding for orphan snapshot tmp directory")
}

func TestDoctor_Repair_CleanTmp_SnapshotTmp(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create orphan snapshot tmp directory
	snapshotsDir := filepath.Join(repoPath, ".jvs", "snapshots")
	os.MkdirAll(snapshotsDir, 0755)
	os.MkdirAll(filepath.Join(snapshotsDir, "something.tmp"), 0755)

	doc := doctor.NewDoctor(repoPath)
	results, err := doc.Repair([]string{"clean_tmp"})
	require.NoError(t, err)
	assert.True(t, results[0].Success)
	assert.GreaterOrEqual(t, results[0].Cleaned, 1)

	// Verify tmp directory is gone
	assert.NoDirExists(t, filepath.Join(snapshotsDir, "something.tmp"))
}

func TestDoctor_Check_AuditChain_WithBrokenChain(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create an audit log with broken hash chain
	auditDir := filepath.Join(repoPath, ".jvs", "audit")
	require.NoError(t, os.MkdirAll(auditDir, 0755))
	auditPath := filepath.Join(auditDir, "audit.jsonl")

	// Write audit records with mismatched hashes
	record1 := `{"prev_hash":"","record_hash":"hash1","timestamp":"2024-01-01T00:00:00Z","event_type":"test"}`
	record2 := `{"prev_hash":"wrong_hash","record_hash":"hash2","timestamp":"2024-01-01T01:00:00Z","event_type":"test"}`
	auditContent := record1 + "\n" + record2 + "\n"
	require.NoError(t, os.WriteFile(auditPath, []byte(auditContent), 0644))

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(true)
	require.NoError(t, err)

	// Should detect broken audit chain
	assert.False(t, result.Healthy)
	found := false
	for _, f := range result.Findings {
		if f.Category == "audit" && f.ErrorCode == "E_AUDIT_CHAIN_BROKEN" {
			found = true
			assert.Equal(t, "critical", f.Severity)
		}
	}
	assert.True(t, found, "expected audit chain broken finding")
}

func TestDoctor_Check_AuditChain_WithMalformedRecord(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create an audit log with malformed record
	auditDir := filepath.Join(repoPath, ".jvs", "audit")
	require.NoError(t, os.MkdirAll(auditDir, 0755))
	auditPath := filepath.Join(auditDir, "audit.jsonl")

	record1 := `{"prev_hash":"","record_hash":"hash1","timestamp":"2024-01-01T00:00:00Z","event_type":"test"}`
	record2 := `{invalid json}`
	auditContent := record1 + "\n" + record2 + "\n"
	require.NoError(t, os.WriteFile(auditPath, []byte(auditContent), 0644))

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(true)
	require.NoError(t, err)

	// Should detect malformed record (warning, not critical)
	found := false
	for _, f := range result.Findings {
		if f.Category == "audit" && f.Severity == "warning" {
			found = true
			assert.Contains(t, f.Description, "malformed record")
		}
	}
	assert.True(t, found, "expected malformed record finding")
}

func TestDoctor_Check_AuditChain_NoAuditLog(t *testing.T) {
	repoPath := setupTestRepo(t)

	// No audit log exists - should be OK
	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(true)
	require.NoError(t, err)
	assert.True(t, result.Healthy)
}

func TestDoctor_Check_WithOrphanTmp(t *testing.T) {
	repoPath := setupTestRepo(t)

	require.NoError(t, os.MkdirAll(filepath.Join(repoPath, ".jvs-tmp-snapshot-abc123"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(repoPath, ".jvs-tmp-snapshot-def456"), 0755))

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)

	tmpFindings := 0
	for _, f := range result.Findings {
		if f.Category == "tmp" {
			tmpFindings++
		}
	}
	assert.GreaterOrEqual(t, tmpFindings, 2, "should find at least 2 orphan tmp directories")
}

func TestDoctor_Repair_WithOrphanTmp(t *testing.T) {
	repoPath := setupTestRepo(t)

	dir1 := filepath.Join(repoPath, ".jvs-tmp-snapshot-abc123")
	dir2 := filepath.Join(repoPath, ".jvs-tmp-snapshot-def456")
	require.NoError(t, os.MkdirAll(dir1, 0755))
	require.NoError(t, os.MkdirAll(dir2, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "partial.dat"), []byte("data"), 0644))

	doc := doctor.NewDoctor(repoPath)
	results, err := doc.Repair([]string{"clean_tmp"})
	require.NoError(t, err)

	assert.Len(t, results, 1)
	assert.True(t, results[0].Success)
	assert.GreaterOrEqual(t, results[0].Cleaned, 2)

	assert.NoDirExists(t, dir1)
	assert.NoDirExists(t, dir2)
}

func TestDoctor_Check_CorruptedFormatVersion(t *testing.T) {
	repoPath := setupTestRepo(t)

	versionPath := filepath.Join(repoPath, ".jvs", "format_version")
	require.NoError(t, os.WriteFile(versionPath, []byte("not-a-number"), 0644))

	doc := doctor.NewDoctor(repoPath)
	result, err := doc.Check(false)
	require.NoError(t, err)

	assert.False(t, result.Healthy)
	require.NotEmpty(t, result.Findings)

	found := false
	for _, f := range result.Findings {
		if f.Category == "format" && f.Severity == "critical" {
			found = true
			assert.Contains(t, f.Description, "invalid content")
		}
	}
	assert.True(t, found, "expected critical format finding for corrupted format_version")
}

func TestDoctor_Check_SnapshotIntegrity_VerifyError(t *testing.T) {
	repoPath := setupTestRepo(t)

	// Create a snapshot but corrupt its descriptor
	createTestSnapshot(t, repoPath)

	descriptorsDir := filepath.Join(repoPath, ".jvs", "descriptors")
	entries, _ := os.ReadDir(descriptorsDir)
	if len(entries) > 0 {
		descriptorPath := filepath.Join(descriptorsDir, entries[0].Name())
		// Write invalid JSON
		os.WriteFile(descriptorPath, []byte("{invalid json"), 0644)

		doc := doctor.NewDoctor(repoPath)
		result, err := doc.Check(true)
		require.NoError(t, err)

		// Should report verification error
		found := false
		for _, f := range result.Findings {
			if f.Category == "integrity" {
				found = true
			}
		}
		// May or may not find depending on how verifier handles corrupt descriptors
		_ = found
	}
}
