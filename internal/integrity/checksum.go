// Package integrity provides checksum and payload hash computation for snapshots.
package integrity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/jvs-project/jvs/pkg/jsonutil"
	"github.com/jvs-project/jvs/pkg/model"
)

// ComputeDescriptorChecksum computes SHA-256 checksum of the descriptor.
// Excludes: descriptor_checksum, integrity_state (per spec 04)
func ComputeDescriptorChecksum(desc *model.Descriptor) (model.HashValue, error) {
	// Create copy with excluded fields zeroed
	checksumDesc := &model.Descriptor{
		SnapshotID:      desc.SnapshotID,
		ParentID:        desc.ParentID,
		WorktreeName:    desc.WorktreeName,
		CreatedAt:       desc.CreatedAt,
		Note:            desc.Note,
		Tags:            desc.Tags,
		Engine:          desc.Engine,
		PayloadRootHash: desc.PayloadRootHash,
		// DescriptorChecksum: excluded
		// IntegrityState: excluded
	}

	data, err := jsonutil.CanonicalMarshal(checksumDesc)
	if err != nil {
		return "", fmt.Errorf("canonical marshal descriptor: %w", err)
	}

	hash := sha256.Sum256(data)
	return model.HashValue(hex.EncodeToString(hash[:])), nil
}
