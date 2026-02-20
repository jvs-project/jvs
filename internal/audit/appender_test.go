package audit_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/jvs-project/jvs/internal/audit"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileAppender_AppendCreatesJSONL(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.jsonl")

	appender := audit.NewFileAppender(logPath)
	err := appender.Append(model.EventTypeSnapshotCreate, "main", "1708300800000-a3f7c1b2", nil)
	require.NoError(t, err)

	// Verify file exists and has content
	file, err := os.Open(logPath)
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	require.True(t, scanner.Scan())
	line := scanner.Text()

	var record model.AuditRecord
	require.NoError(t, json.Unmarshal([]byte(line), &record))
	assert.Equal(t, model.EventTypeSnapshotCreate, record.EventType)
	assert.Equal(t, "main", record.WorktreeName)
}

func TestFileAppender_HashChain(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.jsonl")

	appender := audit.NewFileAppender(logPath)

	// First record
	err := appender.Append(model.EventTypeSnapshotCreate, "main", "id1", nil)
	require.NoError(t, err)

	// Second record
	err = appender.Append(model.EventTypeWorktreeCreate, "feature", "", map[string]any{"base": "main"})
	require.NoError(t, err)

	// Read both records
	file, err := os.Open(logPath)
	require.NoError(t, err)
	defer file.Close()

	var records []model.AuditRecord
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var r model.AuditRecord
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &r))
		records = append(records, r)
	}

	require.Len(t, records, 2)

	// First record has empty prev_hash
	assert.Equal(t, model.HashValue(""), records[0].PrevHash)

	// Second record's prev_hash equals first record's record_hash
	assert.Equal(t, records[0].RecordHash, records[1].PrevHash)

	// Both records have non-empty record_hash
	assert.NotEmpty(t, records[0].RecordHash)
	assert.NotEmpty(t, records[1].RecordHash)
}

func TestFileAppender_ConcurrentAppends(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.jsonl")

	appender := audit.NewFileAppender(logPath)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			appender.Append(model.EventTypeSnapshotCreate, "main", "id", map[string]any{"idx": idx})
		}(i)
	}
	wg.Wait()

	// Verify all records are present
	file, err := os.Open(logPath)
	require.NoError(t, err)
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count++
	}
	assert.Equal(t, 10, count)
}

func TestFileAppender_GetLastRecordHash(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.jsonl")

	appender := audit.NewFileAppender(logPath)

	// Empty file returns empty hash
	hash, err := appender.GetLastRecordHash()
	require.NoError(t, err)
	assert.Equal(t, model.HashValue(""), hash)

	// After append, returns the record hash
	err = appender.Append(model.EventTypeSnapshotCreate, "main", "id1", nil)
	require.NoError(t, err)

	hash, err = appender.GetLastRecordHash()
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}
