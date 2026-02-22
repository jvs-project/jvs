# Terraform configuration for JVS Terraform Provider

terraform {
  required_providers {
    jvs = {
      # For local development, use dev override
      source  = "github.com/jvs-project/jvs"
      version = ">= 1.0.0"
    }
  }
}

# Provider configuration
provider "jvs" {
  # Base path for all repositories
  repo_path = "/opt/jvs"
}

# Example: Create a simple repository
resource "jvs_repository" "simple" {
  name   = "simple-example"
  engine = "copy"
}

# Example: Create a repository with explicit path
resource "jvs_repository" "ml_project" {
  name   = "ml-experiments"
  path   = "/opt/jvs/ml-experiments"
  engine = "juicefs-clone"
}

# Example: Create main worktree
resource "jvs_worktree" "main" {
  repository = jvs_repository.ml_project.path
  name       = "main"
  engine     = "juicefs-clone"
}

# Example: Create additional worktrees
resource "jvs_worktree" "experiment1" {
  repository = jvs_repository.ml_project.path
  name       = "experiment-1"
  engine     = "copy"
}

resource "jvs_worktree" "experiment2" {
  repository = jvs_repository.ml_project.path
  name       = "experiment-2"
  engine     = "copy"
}

# Example: Create initial snapshot
resource "jvs_snapshot" "initial" {
  repository = jvs_repository.ml_project.path
  worktree   = jvs_worktree.main.name
  note       = "Initial ML project setup with model architecture"
  tags       = ["baseline", "v0.1", "initial"]
}

# Example: Create checkpoint snapshots
resource "jvs_snapshot" "checkpoint" {
  repository = jvs_repository.ml_project.path
  worktree   = jvs_worktree.experiment1.name
  note       = "Checkpoint before hyperparameter tuning"
  tags       = ["checkpoint", "experiment-1"]
}

# Outputs
output "simple_repo_id" {
  value = jvs_repository.simple.id
}

output "ml_repo_path" {
  value = jvs_repository.ml_project.path
}

output "main_worktree_path" {
  value = jvs_worktree.main.path
}

output "initial_snapshot_id" {
  value = jvs_snapshot.initial.snapshot_id
}

output "all_snapshot_ids" {
  value = [
    jvs_snapshot.initial.snapshot_id,
    jvs_snapshot.checkpoint.snapshot_id
  ]
}
