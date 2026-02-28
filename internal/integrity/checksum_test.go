package integrity_test

import (
	"testing"
	"time"

	"github.com/jvs-project/jvs/internal/integrity"
	"github.com/jvs-project/jvs/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeDescriptorChecksum_Deterministic(t *testing.T) {
	desc := &model.Descriptor{
		SnapshotID:      "1708300800000-a3f7c1b2",
		WorktreeName:    "main",
		CreatedAt:       time.Date(2024, 2, 19, 0, 0, 0, 0, time.UTC),
		Engine:          model.EngineCopy,
		PayloadRootHash: "abc123",
		IntegrityState:  model.IntegrityVerified,
	}

	hash1, err := integrity.ComputeDescriptorChecksum(desc)
	require.NoError(t, err)
	hash2, err := integrity.ComputeDescriptorChecksum(desc)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2, "checksum must be deterministic")
}

func TestComputeDescriptorChecksum_ExcludesChecksumField(t *testing.T) {
	desc1 := &model.Descriptor{
		SnapshotID:         "1708300800000-a3f7c1b2",
		WorktreeName:       "main",
		DescriptorChecksum: "hash1",
	}
	desc2 := &model.Descriptor{
		SnapshotID:         "1708300800000-a3f7c1b2",
		WorktreeName:       "main",
		DescriptorChecksum: "hash2", // different
	}

	hash1, _ := integrity.ComputeDescriptorChecksum(desc1)
	hash2, _ := integrity.ComputeDescriptorChecksum(desc2)
	assert.Equal(t, hash1, hash2, "checksum must exclude descriptor_checksum field")
}

func TestComputeDescriptorChecksum_ExcludesIntegrityState(t *testing.T) {
	desc1 := &model.Descriptor{
		SnapshotID:     "1708300800000-a3f7c1b2",
		WorktreeName:   "main",
		IntegrityState: model.IntegrityVerified,
	}
	desc2 := &model.Descriptor{
		SnapshotID:     "1708300800000-a3f7c1b2",
		WorktreeName:   "main",
		IntegrityState: model.IntegrityTampered, // different
	}

	hash1, _ := integrity.ComputeDescriptorChecksum(desc1)
	hash2, _ := integrity.ComputeDescriptorChecksum(desc2)
	assert.Equal(t, hash1, hash2, "checksum must exclude integrity_state field")
}

func TestComputeDescriptorChecksum_DifferentContent(t *testing.T) {
	desc1 := &model.Descriptor{
		SnapshotID:   "1708300800000-a3f7c1b2",
		WorktreeName: "main",
	}
	desc2 := &model.Descriptor{
		SnapshotID:   "1708300800000-a3f7c1b2",
		WorktreeName: "feature", // different
	}

	hash1, _ := integrity.ComputeDescriptorChecksum(desc1)
	hash2, _ := integrity.ComputeDescriptorChecksum(desc2)
	assert.NotEqual(t, hash1, hash2, "different content must produce different checksum")
}

func TestComputeDescriptorChecksum_WithTags(t *testing.T) {
	desc1 := &model.Descriptor{
		SnapshotID:   "1708300800000-a3f7c1b2",
		WorktreeName: "main",
	}
	desc2 := &model.Descriptor{
		SnapshotID:   "1708300800000-a3f7c1b2",
		WorktreeName: "main",
		Tags:         []string{"v1.0", "release"},
	}

	hash1, _ := integrity.ComputeDescriptorChecksum(desc1)
	hash2, _ := integrity.ComputeDescriptorChecksum(desc2)
	assert.NotEqual(t, hash1, hash2, "different tags must produce different checksum")
}

func TestComputeDescriptorChecksum_WithParentID(t *testing.T) {
	desc := &model.Descriptor{
		SnapshotID:   "1708300800000-a3f7c1b2",
		WorktreeName: "main",
		ParentID:     func() *model.SnapshotID { id := model.SnapshotID("parent-id"); return &id }(),
	}

	hash, err := integrity.ComputeDescriptorChecksum(desc)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestComputeDescriptorChecksum_WithNote(t *testing.T) {
	desc := &model.Descriptor{
		SnapshotID:   "1708300800000-a3f7c1b2",
		WorktreeName: "main",
		Note:         "This is a test snapshot with important information",
	}

	hash, err := integrity.ComputeDescriptorChecksum(desc)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestComputeDescriptorChecksum_WithAllFields(t *testing.T) {
	// Test checksum with all fields populated
	parentID := model.SnapshotID("parent-snapshot-id")
	desc := &model.Descriptor{
		SnapshotID:      "1708300800000-a3f7c1b2",
		ParentID:        &parentID,
		WorktreeName:    "main",
		CreatedAt:       time.Date(2024, 2, 19, 12, 30, 45, 0, time.UTC),
		Note:            "Test snapshot with all fields",
		Tags:            []string{"v1.0", "release", "stable"},
		Engine:          model.EngineCopy,
		PayloadRootHash: "abc123def456",
	}

	hash, err := integrity.ComputeDescriptorChecksum(desc)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64) // SHA-256 hex encoded
}

func TestComputeDescriptorChecksum_EmptyFields(t *testing.T) {
	// Test checksum with minimal fields
	desc := &model.Descriptor{
		SnapshotID:   "1708300800000-a3f7c1b2",
		WorktreeName: "main",
		Note:         "",
		Tags:         []string{},
	}

	hash, err := integrity.ComputeDescriptorChecksum(desc)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestComputeDescriptorChecksum_NilParentID(t *testing.T) {
	// Test checksum with nil parent ID
	desc := &model.Descriptor{
		SnapshotID:   "1708300800000-a3f7c1b2",
		WorktreeName: "main",
		ParentID:     nil,
	}

	hash, err := integrity.ComputeDescriptorChecksum(desc)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestComputeDescriptorChecksum_DifferentEngineTypes(t *testing.T) {
	// Test that different engine types produce different checksums
	desc1 := &model.Descriptor{
		SnapshotID:   "1708300800000-a3f7c1b2",
		WorktreeName: "main",
		Engine:       model.EngineCopy,
	}
	desc2 := &model.Descriptor{
		SnapshotID:   "1708300800000-a3f7c1b2",
		WorktreeName: "main",
		Engine:       model.EngineReflinkCopy,
	}

	hash1, _ := integrity.ComputeDescriptorChecksum(desc1)
	hash2, _ := integrity.ComputeDescriptorChecksum(desc2)

	assert.NotEqual(t, hash1, hash2, "different engine types should produce different checksums")
}
