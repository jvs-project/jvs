package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWorkspacePhase(t *testing.T) {
	tests := []struct {
		name  string
		phase WorkspacePhase
		valid bool
	}{
		{"pending phase", WorkspacePhasePending, true},
		{"creating phase", WorkspacePhaseCreating, true},
		{"ready phase", WorkspacePhaseReady, true},
		{"updating phase", WorkspacePhaseUpdating, true},
		{"failed phase", WorkspacePhaseFailed, true},
		{"terminating phase", WorkspacePhaseTerminating, true},
		{"invalid phase", WorkspacePhase("Invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validPhases := map[WorkspacePhase]bool{
				WorkspacePhasePending:    true,
				WorkspacePhaseCreating:   true,
				WorkspacePhaseReady:      true,
				WorkspacePhaseUpdating:   true,
				WorkspacePhaseFailed:     true,
				WorkspacePhaseTerminating: true,
			}
			assert.Equal(t, tt.valid, validPhases[tt.phase])
		})
	}
}

func TestWorkspaceConditions(t *testing.T) {
	ws := &Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "default",
		},
	}

	now := metav1.Now()
	readyCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: now,
		Reason:             "WorkspaceReady",
		Message:            "Workspace is ready",
	}

	ws.SetConditions(readyCondition)

	assert.Equal(t, 1, len(ws.Status.Conditions))
	assert.Equal(t, "Ready", ws.Status.Conditions[0].Type)

	// Test GetCondition
	cond := ws.GetCondition("Ready")
	assert.NotNil(t, cond)
	assert.Equal(t, "Ready", cond.Type)

	// Test non-existent condition
	cond = ws.GetCondition("NonExistent")
	assert.Nil(t, cond)
}

func TestWorkspaceDefaults(t *testing.T) {
	ws := &Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "default",
		},
		Spec: WorkspaceSpec{
			Replicas:         1,
			Storage:          "10Gi",
			DefaultEngine:    "copy",
			StorageClassName: "standard",
		},
	}

	assert.Equal(t, int32(1), ws.Spec.Replicas)
	assert.Equal(t, "10Gi", ws.Spec.Storage)
	assert.Equal(t, "copy", ws.Spec.DefaultEngine)
}

func TestWorkspaceRetentionPolicy(t *testing.T) {
	ws := &Workspace{
		Spec: WorkspaceSpec{
			RetentionPolicy: &RetentionPolicy{
				KeepMinSnapshots: 20,
				KeepMinAge:       "48h",
				KeepTags:         []string{"production", "release"},
			},
		},
	}

	assert.Equal(t, int32(20), ws.Spec.RetentionPolicy.KeepMinSnapshots)
	assert.Equal(t, "48h", ws.Spec.RetentionPolicy.KeepMinAge)
	assert.Equal(t, []string{"production", "release"}, ws.Spec.RetentionPolicy.KeepTags)
}

func TestWorkspaceAutoSnapshot(t *testing.T) {
	ws := &Workspace{
		Spec: WorkspaceSpec{
			AutoSnapshot: &AutoSnapshotConfig{
				Enabled:     true,
				Schedule:    "0 */6 * * *",
				Template:    "checkpoint",
				MaxSnapshots: 50,
			},
		},
	}

	assert.True(t, ws.Spec.AutoSnapshot.Enabled)
	assert.Equal(t, "0 */6 * * *", ws.Spec.AutoSnapshot.Schedule)
	assert.Equal(t, "checkpoint", ws.Spec.AutoSnapshot.Template)
	assert.Equal(t, int32(50), ws.Spec.AutoSnapshot.MaxSnapshots)
}
