# JVS Terraform Provider

The JVS Terraform provider enables managing JVS repositories, worktrees, and snapshots as Infrastructure as Code.

## Requirements

- Terraform >= 1.0
- Go >= 1.23 (for building the provider)

## Using the Provider

### Installation

#### Local Development Build

```bash
cd terraform-provider-jvs
go build -o terraform-provider-jvs
```

#### Terraform Configuration

```hcl
terraform {
  required_providers {
    jvs = {
      source = "github.com/jvs-project/jvs/terraform-provider-jvs"
      version = ">= 1.0.0"
    }
  }
}

provider "jvs" {
  repo_path = "/path/to/jvs/repositories" # Optional, defaults to current directory
}
```

## Resources

### jvs_repository

Creates a JVS repository.

```hcl
resource "jvs_repository" "example" {
  name   = "my-repo"
  path   = "/opt/jvs/my-repo"
  engine = "juicefs-clone"
}
```

**Arguments:**

- `name` - (Required) Name of the repository
- `path` - (Optional) Full path to the repository
- `engine` - (Optional) Default snapshot engine (copy, reflink-copy, juicefs-clone)

**Attributes:**

- `id` - Repository identifier (repo_id)
- `format_version` - Format version of the repository

### jvs_worktree

Creates a worktree in a JVS repository.

```hcl
resource "jvs_repository" "example" {
  name = "my-repo"
  path = "/opt/jvs/my-repo"
}

resource "jvs_worktree" "main" {
  repository = jvs_repository.example.path
  name       = "main"
  path       = "/opt/jvs/my-repo/main"
  engine     = "copy"
}

resource "jvs_worktree" "experiment" {
  repository = jvs_repository.example.path
  name       = "experiment-branch"
  engine     = "juicefs-clone"
}
```

**Arguments:**

- `repository` - (Required) Path to the JVS repository
- `name` - (Required) Name of the worktree
- `path` - (Optional) Path to the worktree payload directory
- `engine` - (Optional) Snapshot engine for this worktree

**Attributes:**

- `id` - Worktree name
- `head_snapshot_id` - Current head snapshot ID
- `latest_snapshot_id` - Latest snapshot ID

### jvs_snapshot

Creates a snapshot of a worktree.

```hcl
resource "jvs_snapshot" "initial" {
  repository = jvs_repository.example.path
  worktree   = jvs_worktree.main.name
  note       = "Initial snapshot"
  tags       = ["v1.0", "baseline"]
}
```

**Arguments:**

- `repository` - (Required) Path to the JVS repository
- `worktree` - (Required) Name of the worktree to snapshot
- `note` - (Optional) Descriptive note for the snapshot
- `tags` - (Optional) List of tags to attach

**Attributes:**

- `id` - Snapshot ID
- `snapshot_id` - The generated snapshot ID
- `created_at` - Timestamp when snapshot was created

**Note:** Snapshots are immutable and cannot be updated. To change state, create a new snapshot.

## Data Sources

### jvs_repository

Reads a JVS repository's information.

```hcl
data "jvs_repository" "example" {
  path = "/opt/jvs/my-repo"
}

output "repo_format" {
  value = data.jvs_repository.example.format_version
}
```

### jvs_worktree

Reads a worktree's information.

```hcl
data "jvs_worktree" "main" {
  repository = "/opt/jvs/my-repo"
  name       = "main"
}

output "worktree_path" {
  value = data.jvs_worktree.main.path
}
```

### jvs_snapshot

Reads a snapshot's information.

```hcl
data "jvs_snapshot" "latest" {
  repository = "/opt/jvs/my-repo"
  id         = "1234567890-abcd1234"  # snapshot ID
}

output "snapshot_note" {
  value = data.jvs_snapshot.latest.note
}
```

### jvs_snapshots

Lists all snapshots in a repository.

```hcl
data "jvs_snapshots" "all" {
  repository = "/opt/jvs/my-repo"
}

output "snapshot_count" {
  value = length(data.jvs_snapshots.all.snapshots)
}

output "snapshot_ids" {
  value = data.jvs_snapshots.all.snapshots[*].id
}
```

**Filters:**

- `worktree` - Filter by worktree name
- `tag` - Filter by tag

## Examples

### Complete Workspace Setup

```hcl
# Configure provider
terraform {
  required_providers {
    jvs = {
      source = "github.com/jvs-project/jvs"
    }
  }
}

provider "jvs" {
  repo_path = "/opt/workspaces"
}

# Create repository
resource "jvs_repository" "ml_project" {
  name   = "ml-experiments"
  path   = "/opt/workspaces/ml-experiments"
  engine = "juicefs-clone"
}

# Create main worktree
resource "jvs_worktree" "main" {
  repository = jvs_repository.ml_project.path
  name       = "main"
  engine     = "juicefs-clone"
}

# Create experiment worktree
resource "jvs_worktree" "exp1" {
  repository = jvs_repository.ml_project.path
  name       = "experiment-1"
  engine     = "copy"
}

# Create initial snapshot
resource "jvs_snapshot" "initial" {
  repository = jvs_repository.ml_project.path
  worktree   = jvs_worktree.main.name
  note       = "Initial ML project setup"
  tags       = ["baseline", "v0.1"]
}

# Output key information
output "repo_id" {
  value = jvs_repository.ml_project.id
}

output "main_worktree_path" {
  value = jvs_worktree.main.path
}

output "initial_snapshot_id" {
  value = jvs_snapshot.initial.snapshot_id
}
```

### Scheduled Snapshots with Timeouts

```hcl
resource "jvs_snapshot" "daily" {
  repository = var.repo_path
  worktree   = "main"
  note       = "Daily backup ${timestamp()}"
  tags       = ["backup", "daily"]

  # Trigger with external scheduler (cron, etc.)
  # Snapshots are immutable, so this creates a new snapshot each run
}
```

### Multi-Environment Setup

```hcl
variable "project_name" {
  type    = string
  default = "myproject"
}

variable "environments" {
  type    = list(string)
  default = ["dev", "staging", "production"]
}

# Single repository for all environments
resource "jvs_repository" "main" {
  name   = var.project_name
  path   = "/opt/jvs/${var.project_name}"
  engine = "juicefs-clone"
}

# Create worktree for each environment
resource "jvs_worktree" "environments" {
  for_each = toset(var.environments)

  repository = jvs_repository.main.path
  name       = "${var.project_name}-${each.value}"
  engine     = "copy"  # Use copy for non-production
}

# Create baseline snapshots
resource "jvs_snapshot" "baseline" {
  for_each = toset(var.environments)

  repository = jvs_repository.main.path
  worktree   = jvs_worktree.environments[each.key].name
  note       = "Baseline for ${each.value} environment"
  tags       = ["baseline", each.value]
}
```

## Development

### Building the Provider

```bash
cd terraform-provider-jvs
go build -o terraform-provider-jvs
```

### Local Testing with Terraform

1. Create a `.terraformrc` file:

```hcl
provider_installation {
  dev_overrides {
    "jvs-project/jvs" = "/path/to/terraform-provider-jvs"
  }
}
```

2. Run Terraform commands:

```bash
terraform init
terraform plan
terraform apply
```

### Running Acceptance Tests

```bash
cd terraform-provider-jvs
go test ./...
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](../CONTRIBUTING.md) for details.

## License

This provider is licensed under the MIT License, same as JVS itself.
