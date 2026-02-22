# Variable definitions for JVS Terraform Provider

variable "repo_base_path" {
  type        = string
  default     = "/opt/jvs"
  description = "Base directory for JVS repositories"
}

variable "default_engine" {
  type        = string
  default     = "copy"
  description = "Default snapshot engine for repositories"

  validation {
    condition     = contains(["copy", "reflink-copy", "juicefs-clone"], var.default_engine)
    error_message = "Engine must be one of: copy, reflink-copy, juicefs-clone"
  }
}

variable "project_name" {
  type        = string
  default     = "myproject"
  description = "Name for the JVS repository"
}

variable "create_worktrees" {
  type        = bool
  default     = true
  description = "Whether to create additional worktrees"
}

variable "initial_tags" {
  type        = list(string)
  default     = ["baseline"]
  description = "Default tags for initial snapshot"
}
