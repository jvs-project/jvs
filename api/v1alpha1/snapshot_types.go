package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SnapshotSpec defines the desired state of Snapshot
type SnapshotSpec struct {
	// Workspace is the name of the workspace this snapshot belongs to
	Workspace string `json:"workspace"`

	// SourceSnapshot is the parent snapshot ID (empty for base snapshots)
	// +optional
	SourceSnapshot string `json:"sourceSnapshot,omitempty"`

	// Note is a human-readable note for the snapshot
	// +optional
	Note string `json:"note,omitempty"`

	// Tags are labels for categorizing snapshots
	// +optional
	Tags []string `json:"tags,omitempty"`

	// PartialPaths are paths to include (empty = full snapshot)
	// +optional
	PartialPaths []string `json:"partialPaths,omitempty"`

	// Engine is the snapshot engine to use
	// +optional
	// +kubebuilder:default="copy"
	Engine string `json:"engine,omitempty"`

	// Template is the snapshot template to use
	// +optional
	Template string `json:"template,omitempty"`

	// Compression settings for the snapshot
	// +optional
	Compression *CompressionSpec `json:"compression,omitempty"`

	// RestoreOnCreate creates a new worktree from this snapshot immediately
	// +optional
	RestoreOnCreate bool `json:"restoreOnCreate,omitempty"`

	// RestoreWorktree is the name of the worktree to create on restore
	// +optional
	RestoreWorktree string `json:"restoreWorktree,omitempty"`
}

// CompressionSpec defines compression settings
type CompressionSpec struct {
	// Type is the compression type (gzip, zstd)
	// +kubebuilder:validation:Enum=gzip;zstd;none
	Type string `json:"type,omitempty"`

	// Level is the compression level (0-9)
	// +kubebuilder:minimum=0
	// +kubebuilder:maximum=9
	// +optional
	Level int32 `json:"level,omitempty"`
}

// SnapshotStatus defines the observed state of Snapshot
type SnapshotStatus struct {
	// Phase is the current phase of the snapshot
	// +optional
	Phase SnapshotPhase `json:"phase,omitempty"`

	// Message provides human-readable status information
	// +optional
	Message string `json:"message,omitempty"`

	// SnapshotID is the unique identifier of the snapshot
	// +optional
	SnapshotID string `json:"snapshotID,omitempty"`

	// CreatedAt is when the snapshot was created
	// +optional
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// CompletedAt is when the snapshot completed
	// +optional
	CompletedAt *metav1.Time `json:"completedAt,omitempty"`

	// Size is the size of the snapshot in bytes
	// +optional
	Size int64 `json:"size,omitempty"`

	// IntegrityState is the verification status
	// +optional
	IntegrityState string `json:"integrityState,omitempty"`

	// DescriptorChecksum is the checksum of the descriptor
	// +optional
	DescriptorChecksum string `json:"descriptorChecksum,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// RestoreStatus tracks the status of restore operations
	// +optional
	RestoreStatus *RestoreStatus `json:"restoreStatus,omitempty"`
}

// RestoreStatus tracks restore operation status
type RestoreStatus struct {
	// WorktreeName is the name of the restored worktree
	// +optional
	WorktreeName string `json:"worktreeName,omitempty"`

	// Phase is the phase of the restore operation
	// +optional
	Phase RestorePhase `json:"phase,omitempty"`

	// StartedAt is when the restore started
	// +optional
	StartedAt *metav1.Time `json:"startedAt,omitempty"`

	// CompletedAt is when the restore completed
	// +optional
	CompletedAt *metav1.Time `json:"completedAt,omitempty"`
}

// RestorePhase represents the phase of a restore operation
// +kubebuilder:validation:Enum=Pending;InProgress;Completed;Failed
type RestorePhase string

const (
	// RestorePhasePending means the restore is pending
	RestorePhasePending RestorePhase = "Pending"

	// RestorePhaseInProgress means the restore is in progress
	RestorePhaseInProgress RestorePhase = "InProgress"

	// RestorePhaseCompleted means the restore completed successfully
	RestorePhaseCompleted RestorePhase = "Completed"

	// RestorePhaseFailed means the restore failed
	RestorePhaseFailed RestorePhase = "Failed"
)

// SnapshotPhase represents the lifecycle phase of a snapshot
// +kubebuilder:validation:Enum=Pending;InProgress;Ready;Failed;Expiring
type SnapshotPhase string

const (
	// SnapshotPhasePending means the snapshot is pending creation
	SnapshotPhasePending SnapshotPhase = "Pending"

	// SnapshotPhaseInProgress means the snapshot is being created
	SnapshotPhaseInProgress SnapshotPhase = "InProgress"

	// SnapshotPhaseReady means the snapshot is ready
	SnapshotPhaseReady SnapshotPhase = "Ready"

	// SnapshotPhaseFailed means the snapshot failed
	SnapshotPhaseFailed SnapshotPhase = "Failed"

	// SnapshotPhaseExpiring means the snapshot is being expired
	SnapshotPhaseExpiring SnapshotPhase = "Expiring"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=jvssnap
// +kubebuilder:printcolumn:name="Workspace",type=string,JSONPath=`.spec.workspace`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="SnapshotID",type=string,JSONPath=`.status.snapshotID`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +genclient

// Snapshot is the Schema for the snapshots API
type Snapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotSpec   `json:"spec,omitempty"`
	Status SnapshotStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SnapshotList contains a list of Snapshot
type SnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Snapshot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Snapshot{}, &SnapshotList{})
}

// SetConditions sets the conditions on the snapshot status
func (s *Snapshot) SetConditions(conditions ...metav1.Condition) {
	s.Status.Conditions = conditions
}

// GetCondition returns the condition with the given type
func (s *Snapshot) GetCondition(conditionType string) *metav1.Condition {
	for i := range s.Status.Conditions {
		if s.Status.Conditions[i].Type == conditionType {
			return &s.Status.Conditions[i]
		}
	}
	return nil
}
