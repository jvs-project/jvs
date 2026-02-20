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
		SnapshotID:       "1708300800000-a3f7c1b2",
		WorktreeName:     "main",
		CreatedAt:        time.Date(2024, 2, 19, 0, 0, 0, 0, time.UTC),
		Engine:           model.EngineCopy,
		ConsistencyLevel: model.ConsistencyQuiesced,
		PayloadRootHash:  "abc123",
		FencingToken:     1,
		IntegrityState:   model.IntegrityVerified,
	}

	hash1, err := integrity.ComputeDescriptorChecksum(desc)
	require.NoError(t, err)
	hash2, err := integrity.ComputeDescriptorChecksum(desc)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2, "checksum must be deterministic")
}

func TestComputeDescriptorChecksum_ExcludesChecksumField(t *testing.T) {
	desc1 := &model.Descriptor{
		SnapshotID:        "1708300800000-a3f7c1b2",
		WorktreeName:      "main",
		DescriptorChecksum: "hash1",
	}
	desc2 := &model.Descriptor{
		SnapshotID:        "1708300800000-a3f7c1b2",
		WorktreeName:      "main",
		DescriptorChecksum: "hash2", // different
	}

	hash1, _ := integrity.ComputeDescriptorChecksum(desc1)
	hash2, _ := integrity.ComputeDescriptorChecksum(desc2)
	assert.Equal(t, hash1, hash2, "checksum must exclude descriptor_checksum field")
}

func TestComputeDescriptorChecksum_ExcludesIntegrityState(t *testing.T) {
	desc1 := &model.Descriptor{
		SnapshotID:      "1708300800000-a3f7c1b2",
		WorktreeName:    "main",
		IntegrityState:  model.IntegrityVerified,
	}
	desc2 := &model.Descriptor{
		SnapshotID:      "1708300800000-a3f7c1b2",
		WorktreeName:    "main",
		IntegrityState:  model.IntegrityTampered, // different
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
