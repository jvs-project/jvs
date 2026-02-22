# Output examples for JVS Terraform Provider

# Output repository information
output "repository_info" {
  value = {
    id            = jvs_repository.ml_project.id
    path          = jvs_repository.ml_project.path
    format_version = jvs_repository.ml_project.format_version
  }
  description = "JVS repository information"
}

# Output worktree paths
output "worktree_paths" {
  value = {
    main       = jvs_worktree.main.path
    experiment1 = jvs_worktree.experiment1.path
    experiment2 = jvs_worktree.experiment2.path
  }
  description = "All worktree payload paths"
}

# Output snapshot details
output "snapshots" {
  value = {
    initial_id    = jvs_snapshot.initial.snapshot_id
    initial_time  = jvs_snapshot.initial.created_at
    checkpoint_id = jvs_snapshot.checkpoint.snapshot_id
  }
  description = "Snapshot IDs and metadata"
}

# Dynamic worktree listing
output "all_worktrees" {
  value = [
    jvs_worktree.main.name,
    jvs_worktree.experiment1.name,
    jvs_worktree.experiment2.name
  ]
  description = "All worktree names"
}
