package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WorkspaceSpec defines the desired state of Workspace
type WorkspaceSpec struct {
	// Replicas is the number of pods that will mount this workspace
	// +optional
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`

	// StorageClassName is the name of the StorageClass for the PV
	// +optional
	StorageClassName string `json:"storageClassName,omitempty"`

	// Storage is the storage request for the workspace
	// +optional
	// +kubebuilder:default="10Gi"
	Storage string `json:"storage,omitempty"`

	// JuiceFSConfig holds JuiceFS configuration for the workspace
	// +optional
	JuiceFSConfig *JuiceFSConfig `json:"juicefsConfig,omitempty"`

	// DefaultEngine is the snapshot engine to use (juicefs-clone, reflink-copy, copy)
	// +optional
	// +kubebuilder:default="copy"
	DefaultEngine string `json:"defaultEngine,omitempty"`

	// RetentionPolicy defines how snapshots should be retained
	// +optional
	RetentionPolicy *RetentionPolicy `json:"retentionPolicy,omitempty"`

	// AutoSnapshot enables automatic snapshots on schedule
	// +optional
	AutoSnapshot *AutoSnapshotConfig `json:"autoSnapshot,omitempty"`

	// MountOptions are additional options passed to the mount command
	// +optional
	MountOptions []string `json:"mountOptions,omitempty"`

	// InitialSnapshot is the snapshot to restore when creating the workspace
	// +optional
	InitialSnapshot string `json:"initialSnapshot,omitempty"`
}

// JuiceFSConfig contains JuiceFS specific configuration
type JuiceFSConfig struct {
	// Source is the JuiceFS source URL (e.g., redis://localhost:6379/0)
	Source string `json:"source"`

	// MetaDir is the metadata directory for JuiceFS
	// +optional
	MetaDir string `json:"metaDir,omitempty"`

	// CacheDir is the cache directory for JuiceFS
	// +optional
	CacheDir string `json:"cacheDir,omitempty"`

	// SecretsRef references secrets containing JuiceFS credentials
	// +optional
	SecretsRef *JuiceFSSecretsRef `json:"secretsRef,omitempty"`
}

// JuiceFSSecretsRef references secrets for JuiceFS authentication
type JuiceFSSecretsRef struct {
	// Name is the name of the secret
	Name string `json:"name"`

	// Keys in the secret containing authentication data
	// +optional
	Keys map[string]string `json:"keys,omitempty"`
}

// RetentionPolicy defines snapshot retention rules
type RetentionPolicy struct {
	// KeepMinSnapshots is the minimum number of snapshots to keep
	// +optional
	// +kubebuilder:default=10
	KeepMinSnapshots int32 `json:"keepMinSnapshots,omitempty"`

	// KeepMinAge is the minimum age before a snapshot can be pruned
	// +optional
	// +kubebuilder:default="24h"
	KeepMinAge string `json:"keepMinAge,omitempty"`

	// KeepTags defines retention based on tags (snapshots with these tags are kept)
	// +optional
	KeepTags []string `json:"keepTags,omitempty"`
}

// AutoSnapshotConfig defines automatic snapshot creation
type AutoSnapshotConfig struct {
	// Enabled enables automatic snapshot creation
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// Schedule is a cron expression for snapshot creation
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// Template is the snapshot template to use
	// +optional
	Template string `json:"template,omitempty"`

	// MaxSnapshots is the maximum number of auto-snapshots to keep
	// +optional
	// +kubebuilder:default=30
	MaxSnapshots int32 `json:"maxSnapshots,omitempty"`
}

// WorkspaceStatus defines the observed state of Workspace
type WorkspaceStatus struct {
	// Phase is the current phase of the workspace
	// +optional
	Phase WorkspacePhase `json:"phase,omitempty"`

	// Message provides human-readable status information
	// +optional
	Message string `json:"message,omitempty"`

	// CurrentSnapshot is the ID of the current snapshot
	// +optional
	CurrentSnapshot string `json:"currentSnapshot,omitempty"`

	// SnapshotCount is the number of snapshots in this workspace
	// +optional
	SnapshotCount int32 `json:"snapshotCount,omitempty"`

	// ReadyReplicas is the number of pods that are ready
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastSnapshotTime is the timestamp of the last snapshot
	// +optional
	LastSnapshotTime *metav1.Time `json:"lastSnapshotTime,omitempty"`

	// NextSnapshotTime is the timestamp when the next auto-snapshot is due
	// +optional
	NextSnapshotTime *metav1.Time `json:"nextSnapshotTime,omitempty"`
}

// WorkspacePhase represents the lifecycle phase of a workspace
// +kubebuilder:validation:Enum=Pending;Creating;Ready;Updating;Failed;Terminating
type WorkspacePhase string

const (
	// WorkspacePhasePending means the workspace is pending creation
	WorkspacePhasePending WorkspacePhase = "Pending"

	// WorkspacePhaseCreating means the workspace is being created
	WorkspacePhaseCreating WorkspacePhase = "Creating"

	// WorkspacePhaseReady means the workspace is ready for use
	WorkspacePhaseReady WorkspacePhase = "Ready"

	// WorkspacePhaseUpdating means the workspace is being updated
	WorkspacePhaseUpdating WorkspacePhase = "Updating"

	// WorkspacePhaseFailed means the workspace has failed
	WorkspacePhaseFailed WorkspacePhase = "Failed"

	// WorkspacePhaseTerminating means the workspace is being terminated
	WorkspacePhaseTerminating WorkspacePhase = "Terminating"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=jvs
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Snapshots",type=integer,JSONPath=`.status.snapshotCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +genclient

// Workspace is the Schema for the workspaces API
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceSpec   `json:"spec,omitempty"`
	Status WorkspaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceList contains a list of Workspace
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}

// SetConditions sets the conditions on the workspace status
func (w *Workspace) SetConditions(conditions ...metav1.Condition) {
	w.Status.Conditions = conditions
}

// GetCondition returns the condition with the given type
func (w *Workspace) GetCondition(conditionType string) *metav1.Condition {
	for i := range w.Status.Conditions {
		if w.Status.Conditions[i].Type == conditionType {
			return &w.Status.Conditions[i]
		}
	}
	return nil
}
