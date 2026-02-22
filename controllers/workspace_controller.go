package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	jvsiov1alpha1 "github.com/jvs-project/jvs/api/v1alpha1"
)

const (
workspaceFinalizer       = "jvs.io/workspace-finalizer"
workspaceRequeueAfter     = 30 * time.Second
workspaceRequeueOnFailure = 1 * time.Minute
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	JVSBin string // Path to jvs binary
}

// +kubebuilder:rbac:groups=jvs.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=jvs.io,resources=workspaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=jvs.io,resources=workspaces/finalizers,verbs=update
// +kubebuilder:rbac:groups=jvs.io,resources=snapshots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile is the main reconciliation loop for Workspace resources
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Workspace instance
	workspace := &jvsiov1alpha1.Workspace{}
	err := r.Get(ctx, req.NamespacedName, workspace)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// Examine DeletionTimestamp to determine if object is under deletion
	if workspace.ObjectMeta.DeletionTimestamp.IsZero() {
		// Object not being deleted, add finalizer if needed
		if !controllerutil.ContainsFinalizer(workspace, workspaceFinalizer) {
			controllerutil.AddFinalizer(workspace, workspaceFinalizer)
			if err := r.Update(ctx, workspace); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// Object being deleted
		if controllerutil.ContainsFinalizer(workspace, workspaceFinalizer) {
			// Run finalization logic
			if err := r.finalizeWorkspace(ctx, workspace); err != nil {
				// If finalization fails, return with error so requeue
				return ctrl.Result{RequeueAfter: workspaceRequeueOnFailure}, err
			}

			// Remove finalizer once successfully completed
			controllerutil.RemoveFinalizer(workspace, workspaceFinalizer)
			if err := r.Update(ctx, workspace); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Get or create the PV/PVC for this workspace
	pvcReady, err := r.ensurePVC(ctx, workspace)
	if err != nil {
		log.Error(err, "Failed to ensure PVC")
		r.updateStatus(ctx, workspace, jvsiov1alpha1.WorkspacePhaseFailed, err.Error())
		return ctrl.Result{RequeueAfter: workspaceRequeueOnFailure}, err
	}

	if !pvcReady {
		r.updateStatus(ctx, workspace, jvsiov1alpha1.WorkspacePhasePending, "Waiting for PVC to be bound")
		return ctrl.Result{RequeueAfter: workspaceRequeueAfter}, nil
	}

	// Initialize JVS workspace if needed
	if workspace.Status.Phase == "" || workspace.Status.Phase == jvsiov1alpha1.WorkspacePhasePending {
		if err := r.initWorkspace(ctx, workspace); err != nil {
			log.Error(err, "Failed to initialize workspace")
			r.updateStatus(ctx, workspace, jvsiov1alpha1.WorkspacePhaseFailed, err.Error())
			return ctrl.Result{RequeueAfter: workspaceRequeueOnFailure}, err
		}

		r.updateStatus(ctx, workspace, jvsiov1alpha1.WorkspacePhaseCreating, "Workspace initialized")
	}

	// Check if workspace is ready
	ready, err := r.checkWorkspaceReady(ctx, workspace)
	if err != nil {
		log.Error(err, "Failed to check workspace readiness")
		return ctrl.Result{RequeueAfter: workspaceRequeueOnFailure}, err
	}

	if ready && workspace.Status.Phase != jvsiov1alpha1.WorkspacePhaseReady {
		r.updateStatus(ctx, workspace, jvsiov1alpha1.WorkspacePhaseReady, "Workspace ready")
	}

	// Handle auto-snapshot scheduling
	if workspace.Spec.AutoSnapshot != nil && workspace.Spec.AutoSnapshot.Enabled {
		nextSnapshot, err := r.scheduleAutoSnapshot(ctx, workspace)
		if err != nil {
			log.Error(err, "Failed to schedule auto-snapshot")
		} else if nextSnapshot != nil {
			return ctrl.Result{RequeueAfter: *nextSnapshot}, nil
		}
	}

	// Update snapshot count
	snapshotCount, err := r.updateSnapshotCount(ctx, workspace)
	if err != nil {
		log.Error(err, "Failed to update snapshot count")
	}

	workspace.Status.SnapshotCount = snapshotCount
	if err := r.Status().Update(ctx, workspace); err != nil {
		return ctrl.Result{}, err
	}

	// Requeue periodically to check status
	return ctrl.Result{RequeueAfter: workspaceRequeueAfter}, nil
}

// ensurePVC creates or gets the PVC for the workspace
func (r *WorkspaceReconciler) ensurePVC(ctx context.Context, workspace *jvsiov1alpha1.Workspace) (bool, error) {
	// PVC management logic would go here
	// For now, return true as PVC creation is typically handled separately
	return true, nil
}

// initWorkspace initializes the JVS workspace
func (r *WorkspaceReconciler) initWorkspace(ctx context.Context, workspace *jvsiov1alpha1.Workspace) error {
	// This would execute: jvs init <name> in the mounted volume
	// Implementation would involve running the jvs binary in a pod
	return nil
}

// checkWorkspaceReady checks if the workspace is ready
func (r *WorkspaceReconciler) checkWorkspaceReady(ctx context.Context, workspace *jvsiov1alpha1.Workspace) (bool, error) {
	// Check if the .jvs directory exists and is properly initialized
	// This would involve exec-ing into the workspace pod
	return true, nil
}

// scheduleAutoSnapshot schedules automatic snapshots based on the cron schedule
func (r *WorkspaceReconciler) scheduleAutoSnapshot(ctx context.Context, workspace *jvsiov1alpha1.Workspace) (*time.Duration, error) {
	// Parse cron schedule and calculate next snapshot time
	// Create Snapshot CR at the appropriate time
	return nil, nil
}

// updateSnapshotCount updates the snapshot count in the status
func (r *WorkspaceReconciler) updateSnapshotCount(ctx context.Context, workspace *jvsiov1alpha1.Workspace) (int32, error) {
	// List snapshots for this workspace
	snapshots := &jvsiov1alpha1.SnapshotList{}
	if err := r.List(ctx, snapshots, client.InNamespace(workspace.Namespace), client.MatchingFields{"spec.workspace": workspace.Name}); err != nil {
		return 0, err
	}
	return int32(len(snapshots.Items)), nil
}

// updateStatus updates the workspace status
func (r *WorkspaceReconciler) updateStatus(ctx context.Context, workspace *jvsiov1alpha1.Workspace, phase jvsiov1alpha1.WorkspacePhase, message string) {
	workspace.Status.Phase = phase
	workspace.Status.Message = message

	now := metav1.Now()
	if phase == jvsiov1alpha1.WorkspacePhaseReady {
		readyCondition := metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: now,
			Reason:             "WorkspaceReady",
			Message:            message,
		}
		workspace.SetConditions(readyCondition)
	}
}

// finalizeWorkspace handles workspace deletion
func (r *WorkspaceReconciler) finalizeWorkspace(ctx context.Context, workspace *jvsiov1alpha1.Workspace) error {
	// Clean up any snapshots
	snapshots := &jvsiov1alpha1.SnapshotList{}
	if err := r.List(ctx, snapshots, client.InNamespace(workspace.Namespace), client.MatchingFields{"spec.workspace": workspace.Name}); err != nil {
		return err
	}

	for _, snap := range snapshots.Items {
		if err := r.Delete(ctx, &snap); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete snapshot %s: %w", snap.Name, err)
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&jvsiov1alpha1.Workspace{}).
		// Watch Snapshot resources owned by this workspace
		// Owns(&jvsiov1alpha1.Snapshot{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=jvs.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=jvs.io,resources=workspaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=jvs.io,resources=workspaces/finalizers,verbs=update
