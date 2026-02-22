package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSnapshotPhase(t *testing.T) {
	tests := []struct {
		name  string
		phase SnapshotPhase
		valid bool
	}{
		{"pending phase", SnapshotPhasePending, true},
		{"in progress phase", SnapshotPhaseInProgress, true},
		{"ready phase", SnapshotPhaseReady, true},
		{"failed phase", SnapshotPhaseFailed, true},
		{"expiring phase", SnapshotPhaseExpiring, true},
		{"invalid phase", SnapshotPhase("Invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validPhases := map[SnapshotPhase]bool{
				SnapshotPhasePending:    true,
				SnapshotPhaseInProgress: true,
				SnapshotPhaseReady:      true,
				SnapshotPhaseFailed:     true,
				SnapshotPhaseExpiring:   true,
			}
			assert.Equal(t, tt.valid, validPhases[tt.phase])
		})
	}
}

func TestRestorePhase(t *testing.T) {
	tests := []struct {
		name  string
		phase RestorePhase
		valid bool
	}{
		{"pending phase", RestorePhasePending, true},
		{"in progress phase", RestorePhaseInProgress, true},
		{"completed phase", RestorePhaseCompleted, true},
		{"failed phase", RestorePhaseFailed, true},
		{"invalid phase", RestorePhase("Invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validPhases := map[RestorePhase]bool{
				RestorePhasePending:    true,
				RestorePhaseInProgress: true,
				RestorePhaseCompleted:  true,
				RestorePhaseFailed:     true,
			}
			assert.Equal(t, tt.valid, validPhases[tt.phase])
		})
	}
}

func TestSnapshotConditions(t *testing.T) {
	snap := &Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-snapshot",
			Namespace: "default",
		},
	}

	now := metav1.Now()
	readyCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: now,
		Reason:             "SnapshotCreated",
		Message:            "Snapshot created successfully",
	}

	snap.SetConditions(readyCondition)

	assert.Equal(t, 1, len(snap.Status.Conditions))
	assert.Equal(t, "Ready", snap.Status.Conditions[0].Type)

	// Test GetCondition
	cond := snap.GetCondition("Ready")
	assert.NotNil(t, cond)
	assert.Equal(t, "Ready", cond.Type)

	// Test non-existent condition
	cond = snap.GetCondition("NonExistent")
	assert.Nil(t, cond)
}

func TestSnapshotSpec(t *testing.T) {
	snap := &Snapshot{
		Spec: SnapshotSpec{
			Workspace:       "test-workspace",
			Note:            "Test snapshot",
			Tags:            []string{"test", "checkpoint"},
			PartialPaths:    []string{"src/", "data/"},
			Engine:          "copy",
			Template:        "checkpoint",
			RestoreOnCreate: false,
		},
	}

	assert.Equal(t, "test-workspace", snap.Spec.Workspace)
	assert.Equal(t, "Test snapshot", snap.Spec.Note)
	assert.Equal(t, []string{"test", "checkpoint"}, snap.Spec.Tags)
	assert.Equal(t, []string{"src/", "data/"}, snap.Spec.PartialPaths)
	assert.Equal(t, "copy", snap.Spec.Engine)
	assert.Equal(t, "checkpoint", snap.Spec.Template)
	assert.False(t, snap.Spec.RestoreOnCreate)
}

func TestSnapshotCompression(t *testing.T) {
	snap := &Snapshot{
		Spec: SnapshotSpec{
			Compression: &CompressionSpec{
				Type:  "gzip",
				Level: 6,
			},
		},
	}

	assert.NotNil(t, snap.Spec.Compression)
	assert.Equal(t, "gzip", snap.Spec.Compression.Type)
	assert.Equal(t, int32(6), snap.Spec.Compression.Level)
}

func TestSnapshotRestoreStatus(t *testing.T) {
	now := metav1.Now()
	snap := &Snapshot{
		Status: SnapshotStatus{
			RestoreStatus: &RestoreStatus{
				WorktreeName: "test-branch",
				Phase:        RestorePhaseCompleted,
				StartedAt:    &now,
				CompletedAt:  &now,
			},
		},
	}

	assert.NotNil(t, snap.Status.RestoreStatus)
	assert.Equal(t, "test-branch", snap.Status.RestoreStatus.WorktreeName)
	assert.Equal(t, RestorePhaseCompleted, snap.Status.RestoreStatus.Phase)
	assert.NotNil(t, snap.Status.RestoreStatus.StartedAt)
	assert.NotNil(t, snap.Status.RestoreStatus.CompletedAt)
}

func TestSnapshotWithRestore(t *testing.T) {
	snap := &Snapshot{
		Spec: SnapshotSpec{
			Workspace:       "test-workspace",
			RestoreOnCreate: true,
			RestoreWorktree: "feature-branch",
		},
	}

	assert.True(t, snap.Spec.RestoreOnCreate)
	assert.Equal(t, "feature-branch", snap.Spec.RestoreWorktree)
}
