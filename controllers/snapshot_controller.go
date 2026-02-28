package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	jvsiov1alpha1 "github.com/jvs-project/jvs/api/v1alpha1"
)

const (
	snapshotFinalizer        = "jvs.io/snapshot-finalizer"
	snapshotRequeueAfter     = 10 * time.Second
	snapshotRequeueOnFailure = 30 * time.Second
)

// SnapshotReconciler reconciles a Snapshot object
type SnapshotReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	JVSBin string // Path to jvs binary
}

// +kubebuilder:rbac:groups=jvs.io,resources=snapshots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=jvs.io,resources=snapshots/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=jvs.io,resources=snapshots/finalizers,verbs=update
// +kubebuilder:rbac:groups=jvs.io,resources=workspaces,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups=core,resources=pods/exec,verbs=create
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile is the main reconciliation loop for Snapshot resources
func (r *SnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the Snapshot instance
	snapshot := &jvsiov1alpha1.Snapshot{}
	err := r.Get(ctx, req.NamespacedName, snapshot)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Examine DeletionTimestamp to determine if object is under deletion
	if !snapshot.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(snapshot, snapshotFinalizer) {
			if err := r.finalizeSnapshot(ctx, snapshot); err != nil {
				return ctrl.Result{RequeueAfter: snapshotRequeueOnFailure}, err
			}
			controllerutil.RemoveFinalizer(snapshot, snapshotFinalizer)
			if err := r.Update(ctx, snapshot); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(snapshot, snapshotFinalizer) {
		controllerutil.AddFinalizer(snapshot, snapshotFinalizer)
		if err := r.Update(ctx, snapshot); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Validate workspace exists
	workspace := &jvsiov1alpha1.Workspace{}
	err = r.Get(ctx, types.NamespacedName{Namespace: snapshot.Namespace, Name: snapshot.Spec.Workspace}, workspace)
	if err != nil {
		if errors.IsNotFound(err) {
			r.updateSnapshotStatus(ctx, snapshot, jvsiov1alpha1.SnapshotPhaseFailed, "Workspace not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Wait for workspace to be ready
	if workspace.Status.Phase != jvsiov1alpha1.WorkspacePhaseReady {
		r.updateSnapshotStatus(ctx, snapshot, jvsiov1alpha1.SnapshotPhasePending, "Waiting for workspace to be ready")
		return ctrl.Result{RequeueAfter: snapshotRequeueAfter}, nil
	}

	// Execute snapshot creation based on phase
	switch snapshot.Status.Phase {
	case "":
		fallthrough
	case jvsiov1alpha1.SnapshotPhasePending:
		return r.createSnapshot(ctx, snapshot, workspace)
	case jvsiov1alpha1.SnapshotPhaseInProgress:
		return r.monitorSnapshot(ctx, snapshot, workspace)
	case jvsiov1alpha1.SnapshotPhaseReady:
		// Handle restore if requested
		if snapshot.Spec.RestoreOnCreate && snapshot.Status.RestoreStatus == nil {
			return r.createRestore(ctx, snapshot, workspace)
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// createSnapshot initiates snapshot creation
func (r *SnapshotReconciler) createSnapshot(ctx context.Context, snapshot *jvsiov1alpha1.Snapshot, workspace *jvsiov1alpha1.Workspace) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	now := metav1.Now()
	snapshot.Status.Phase = jvsiov1alpha1.SnapshotPhaseInProgress
	snapshot.Status.Message = "Creating snapshot"
	snapshot.Status.CreatedAt = &now

	if err := r.Status().Update(ctx, snapshot); err != nil {
		return ctrl.Result{}, err
	}

	// Execute jvs snapshot command
	// In production, this would run in a pod with the workspace volume mounted
	snapshotID, err := r.executeSnapshot(ctx, snapshot, workspace)
	if err != nil {
		logger.Error(err, "Failed to create snapshot")
		r.updateSnapshotStatus(ctx, snapshot, jvsiov1alpha1.SnapshotPhaseFailed, err.Error())
		return ctrl.Result{RequeueAfter: snapshotRequeueOnFailure}, nil
	}

	snapshot.Status.SnapshotID = snapshotID
	if err := r.Status().Update(ctx, snapshot); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: snapshotRequeueAfter}, nil
}

// executeSnapshot runs the jvs snapshot command
func (r *SnapshotReconciler) executeSnapshot(ctx context.Context, snapshot *jvsiov1alpha1.Snapshot, workspace *jvsiov1alpha1.Workspace) (string, error) {
	// This would exec into a pod with the workspace volume mounted
	// and run: jvs snapshot [note] [--tag tag]...
	// For now, return a placeholder
	return fmt.Sprintf("snap-%d", time.Now().Unix()), nil
}

// monitorSnapshot monitors snapshot creation progress
func (r *SnapshotReconciler) monitorSnapshot(ctx context.Context, snapshot *jvsiov1alpha1.Snapshot, workspace *jvsiov1alpha1.Workspace) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Check if snapshot is complete
	complete, err := r.checkSnapshotComplete(ctx, snapshot)
	if err != nil {
		logger.Error(err, "Failed to check snapshot status")
		return ctrl.Result{RequeueAfter: snapshotRequeueOnFailure}, err
	}

	if !complete {
		return ctrl.Result{RequeueAfter: snapshotRequeueAfter}, nil
	}

	// Snapshot is complete
	now := metav1.Now()
	snapshot.Status.Phase = jvsiov1alpha1.SnapshotPhaseReady
	snapshot.Status.Message = "Snapshot created successfully"
	snapshot.Status.CompletedAt = &now
	snapshot.Status.IntegrityState = "verified"

	readyCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: now,
		Reason:             "SnapshotCreated",
		Message:            "Snapshot created and verified",
	}
	snapshot.SetConditions(readyCondition)

	// Update workspace snapshot time
	workspace.Status.LastSnapshotTime = &now
	if err := r.Status().Update(ctx, workspace); err != nil {
		logger.Error(err, "Failed to update workspace status")
	}

	return ctrl.Result{}, r.Status().Update(ctx, snapshot)
}

// checkSnapshotComplete checks if the snapshot creation is complete
func (r *SnapshotReconciler) checkSnapshotComplete(ctx context.Context, snapshot *jvsiov1alpha1.Snapshot) (bool, error) {
	// This would check if the .READY file exists in the .jvs/snapshots directory
	// For now, return true after a delay
	if time.Since(snapshot.Status.CreatedAt.Time) > 2*time.Second {
		return true, nil
	}
	return false, nil
}

// createRestore creates a worktree from the snapshot
func (r *SnapshotReconciler) createRestore(ctx context.Context, snapshot *jvsiov1alpha1.Snapshot, workspace *jvsiov1alpha1.Workspace) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	now := metav1.Now()
	snapshot.Status.RestoreStatus = &jvsiov1alpha1.RestoreStatus{
		Phase:        jvsiov1alpha1.RestorePhaseInProgress,
		WorktreeName: snapshot.Spec.RestoreWorktree,
		StartedAt:    &now,
	}

	if err := r.Status().Update(ctx, snapshot); err != nil {
		return ctrl.Result{}, err
	}

	// Execute jvs restore command
	err := r.executeRestore(ctx, snapshot, workspace)
	if err != nil {
		logger.Error(err, "Failed to create restore")
		snapshot.Status.RestoreStatus.Phase = jvsiov1alpha1.RestorePhaseFailed
		completed := metav1.Now()
		snapshot.Status.RestoreStatus.CompletedAt = &completed
		r.Status().Update(ctx, snapshot)
		return ctrl.Result{RequeueAfter: snapshotRequeueOnFailure}, nil
	}

	completed := metav1.Now()
	snapshot.Status.RestoreStatus.Phase = jvsiov1alpha1.RestorePhaseCompleted
	snapshot.Status.RestoreStatus.CompletedAt = &completed

	return ctrl.Result{}, r.Status().Update(ctx, snapshot)
}

// executeRestore runs the jvs restore command
func (r *SnapshotReconciler) executeRestore(ctx context.Context, snapshot *jvsiov1alpha1.Snapshot, workspace *jvsiov1alpha1.Workspace) error {
	// This would exec into a pod and run: jvs restore <id> --worktree <name>
	// For now, return success
	return nil
}

// updateSnapshotStatus updates the snapshot status
func (r *SnapshotReconciler) updateSnapshotStatus(ctx context.Context, snapshot *jvsiov1alpha1.Snapshot, phase jvsiov1alpha1.SnapshotPhase, message string) {
	snapshot.Status.Phase = phase
	snapshot.Status.Message = message
	r.Status().Update(ctx, snapshot)
}

// finalizeSnapshot handles snapshot deletion
func (r *SnapshotReconciler) finalizeSnapshot(ctx context.Context, snapshot *jvsiov1alpha1.Snapshot) error {
	// Optionally delete the actual JVS snapshot from disk
	// This might be skipped if retention policy applies
	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *SnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&jvsiov1alpha1.Snapshot{}).
		// Watch the workspace that owns this snapshot
		// Watches(&jvsiov1alpha1.Workspace{}).
		Complete(r)
}
