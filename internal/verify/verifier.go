package verify

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jvs-project/jvs/internal/integrity"
	"github.com/jvs-project/jvs/internal/snapshot"
	"github.com/jvs-project/jvs/pkg/model"
)

// Result contains verification results for a single snapshot.
type Result struct {
	SnapshotID      model.SnapshotID `json:"snapshot_id"`
	ChecksumValid   bool             `json:"checksum_valid"`
	PayloadHashValid bool            `json:"payload_hash_valid"`
	TamperDetected  bool             `json:"tamper_detected"`
	Severity        string           `json:"severity,omitempty"`
	Error           string           `json:"error,omitempty"`
}

// Verifier performs integrity verification on snapshots.
type Verifier struct {
	repoRoot string
}

// NewVerifier creates a new verifier.
func NewVerifier(repoRoot string) *Verifier {
	return &Verifier{repoRoot: repoRoot}
}

// VerifySnapshot verifies a single snapshot's integrity.
func (v *Verifier) VerifySnapshot(snapshotID model.SnapshotID, verifyPayloadHash bool) (*Result, error) {
	result := &Result{
		SnapshotID: snapshotID,
	}

	desc, err := snapshot.LoadDescriptor(v.repoRoot, snapshotID)
	if err != nil {
		result.Error = err.Error()
		result.TamperDetected = true
		result.Severity = "critical"
		return result, nil
	}

	// Verify descriptor checksum
	computedChecksum, err := integrity.ComputeDescriptorChecksum(desc)
	if err != nil {
		result.Error = fmt.Sprintf("compute checksum: %v", err)
		result.Severity = "error"
		return result, nil
	}

	result.ChecksumValid = computedChecksum == desc.DescriptorChecksum
	if !result.ChecksumValid {
		result.TamperDetected = true
		result.Severity = "critical"
		result.Error = "descriptor checksum mismatch"
		return result, nil
	}

	// Optionally verify payload hash (expensive)
	if verifyPayloadHash {
		snapshotDir := filepath.Join(v.repoRoot, ".jvs", "snapshots", string(snapshotID))
		computedHash, err := integrity.ComputePayloadRootHash(snapshotDir)
		if err != nil {
			result.Error = fmt.Sprintf("compute payload hash: %v", err)
			result.Severity = "error"
			return result, nil
		}

		result.PayloadHashValid = computedHash == desc.PayloadRootHash
		if !result.PayloadHashValid {
			result.TamperDetected = true
			result.Severity = "critical"
			result.Error = "payload hash mismatch"
		}
	}

	return result, nil
}

// VerifyAll verifies all snapshots in the repository.
func (v *Verifier) VerifyAll(verifyPayloadHash bool) ([]*Result, error) {
	snapshotsDir := filepath.Join(v.repoRoot, ".jvs", "snapshots")
	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read snapshots directory: %w", err)
	}

	var results []*Result
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		snapshotID := model.SnapshotID(entry.Name())
		result, err := v.VerifySnapshot(snapshotID, verifyPayloadHash)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}
