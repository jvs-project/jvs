# JVS Kubernetes Operator

The JVS Kubernetes Operator provides a native Kubernetes way to manage JVS workspaces and snapshots.

## Overview

The operator extends Kubernetes with two Custom Resource Definitions (CRDs):

- **Workspace**: Represents a JVS workspace with its storage, retention policy, and auto-snapshot configuration
- **Snapshot**: Represents a snapshot of a workspace with notes, tags, and restore options

## Installation

### Prerequisites

- Kubernetes cluster (v1.25+)
- kubectl configured to talk to your cluster
- JVS installed on your cluster nodes or via init container

### Quick Install

```bash
# Install the CRDs
kubectl apply -f config/crd/bases/

# Install the operator
kubectl apply -f config/rbac/
kubectl apply -f config/manager/
```

### From Source

```bash
# Build the operator image
make -f Makefile.operator docker-build

# Push to your registry
make -f Makefile.operator docker-push IMG=your-registry/jvs-operator:v0.1.0

# Deploy
make -f Makefile.operator deploy-all
```

## Usage

### Creating a Workspace

```yaml
apiVersion: jvs.io/v1alpha1
kind: Workspace
metadata:
  name: ml-workspace
  namespace: default
spec:
  replicas: 1
  storage: 50Gi
  storageClassName: fast-ssd
  defaultEngine: copy
  juiceFsConfig:
    source: redis://redis.default.svc.cluster.local:6379/0
    cacheDir: /mnt/cache
  retentionPolicy:
    keepMinSnapshots: 20
    keepMinAge: 48h
    keepTags:
    - production
    - release
  autoSnapshot:
    enabled: true
    schedule: "0 */4 * * *"  # Every 4 hours
    template: checkpoint
    maxSnapshots: 50
```

Apply with:

```bash
kubectl apply -f config/samples/jvs.io_v1alpha1_workspace.yaml
```

### Creating Snapshots

#### Manual Snapshot

```yaml
apiVersion: jvs.io/v1alpha1
kind: Snapshot
metadata:
  name: pre-deployment-snapshot
spec:
  workspace: ml-workspace
  note: "Before deploying model v2.0"
  tags:
  - pre-deploy
  - model-v2
  compression:
    type: gzip
    level: 6
```

#### Snapshot with Template

```yaml
apiVersion: jvs.io/v1alpha1
kind: Snapshot
metadata:
  name: experiment-checkpoint
spec:
  workspace: ml-workspace
  template: pre-experiment
```

Available templates:
- `pre-experiment`: Tagged with "experiment", "checkpoint"
- `pre-deploy`: Tagged with "pre-deploy", "release"
- `checkpoint`: Tagged with "checkpoint"
- `work`: Tagged with "wip"
- `release`: Tagged with "release", "stable"
- `archive`: Maximum compression, tagged with "archive"

#### Snapshot with Restore

```yaml
apiVersion: jvs.io/v1alpha1
kind: Snapshot
metadata:
  name: test-restore-snapshot
spec:
  workspace: ml-workspace
  note: "Snapshot for testing"
  restoreOnCreate: true
  restoreWorktree: test-branch
```

### Checking Status

```bash
# List all workspaces
kubectl get workspaces --all-namespaces

# List all snapshots
kubectl get snapshots --all-namespaces

# Describe a specific workspace
kubectl describe workspace ml-workspace -n default

# Describe a specific snapshot
kubectl describe snapshot pre-deployment-snapshot -n default

# View operator logs
kubectl logs -f deployment/jvs-operator-controller-manager -n jvs-system
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                       │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │              JVS Operator Pod                           │ │
│  │  ┌─────────────────┐  ┌─────────────────┐              │ │
│  │  │ Workspace       │  │ Snapshot        │              │ │
│  │  │ Controller      │  │ Controller      │              │ │
│  │  └─────────────────┘  └─────────────────┘              │ │
│  └────────────────────────────────────────────────────────┘ │
│                           │                                  │
│                           ▼                                  │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                    API Server                           │ │
│  │  Workspace CRD          Snapshot CRD                    │ │
│  └────────────────────────────────────────────────────────┘ │
│                           │                                  │
│                           ▼                                  │
│  ┌────────────────────────────────────────────────────────┐ │
│  │              Workspace Pods                             │ │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐       │ │
│  │  │ Pod 1      │  │ Pod 2      │  │ Pod 3      │       │ │
│  │  │ (main/)    │  │ (main/)    │  │ (main/)    │       │ │
│  │  └────────────┘  └────────────┘  └────────────┘       │ │
│  └────────────────────────────────────────────────────────┘ │
│                           │                                  │
│                           ▼                                  │
│  ┌────────────────────────────────────────────────────────┐ │
│  │              Persistent Volume                          │ │
│  │  .jvs/ (metadata)     main/ (payload)                   │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Reconciliation Flow

### Workspace Reconciliation

1. **Creation**: Operator creates PVC and initializes JVS workspace
2. **Ready**: Workspace marked ready when `.jvs/` directory exists
3. **Auto-snapshot**: Scheduled snapshots created based on cron schedule
4. **Deletion**: Finalizer cleans up snapshots before removing workspace

### Snapshot Reconciliation

1. **Pending**: Waits for workspace to be ready
2. **InProgress**: Executes `jvs snapshot` command
3. **Ready**: Snapshot verified and marked ready
4. **Restore**: Optional worktree creation if `restoreOnCreate: true`

## Configuration

### JuiceFS Storage

The operator supports JuiceFS for storage backends:

```yaml
spec:
  juiceFsConfig:
    source: redis://localhost:6379/0  # Or other metadata engines
    metaDir: /var/lib/juicefs/meta    # Optional
    cacheDir: /var/lib/juicefs/cache  # Optional
    secretsRef:
      name: juicefs-credentials
      keys:
        source: "REDIS_URL"
```

### Auto-Snapshot Scheduling

Auto-snapshots use cron syntax:

```yaml
autoSnapshot:
  enabled: true
  schedule: "0 */6 * * *"  # Every 6 hours
  template: checkpoint
  maxSnapshots: 30
```

### Retention Policy

```yaml
retentionPolicy:
  keepMinSnapshots: 10    # Always keep at least 10 snapshots
  keepMinAge: 24h         # Don't delete snapshots younger than 24h
  keepTags:               # Always keep snapshots with these tags
  - production
  - release
```

## Troubleshooting

### Operator Not Starting

```bash
# Check logs
kubectl logs -f deployment/jvs-operator-controller-manager -n jvs-system

# Check RBAC
kubectl get clusterrole jvs-operator-manager-role
kubectl get clusterrolebinding jvs-operator-manager-rolebinding
```

### Workspace Stuck in Pending

```bash
# Check PVC status
kubectl get pvc -n <namespace>

# Describe workspace for events
kubectl describe workspace <name> -n <namespace>
```

### Snapshot Failing

```bash
# Check snapshot status
kubectl get snapshot <name> -n <namespace> -o yaml

# Check operator logs
kubectl logs -f deployment/jvs-operator-controller-manager -n jvs-system
```

## Development

### Running Locally

```bash
# Install CRDs
make -f Makefile.operator install

# Run operator locally (requires kubeconfig)
make -f Makefile.operator run
```

### Building

```bash
# Build binary
make -f Makefile.operator build

# Build Docker image
make -f Makefile.operator docker-build
```

### Testing

```bash
# Run tests
make -f Makefile.operator test
```

## RBAC Permissions

The operator requires the following permissions:

- **jvs.io/workspaces**: Full CRUD on workspaces and finalizers
- **jvs.io/snapshots**: Full CRUD on snapshots and finalizers
- **persistentvolumeclaims**: Create, delete, get, list, watch
- **pods**: Create, delete, get, list, watch
- **pods/exec**: Create (for running jvs commands)
- **secrets**: Get, list, watch
- **events**: Create, patch

## References

- JVS Documentation: [docs/](../docs/)
- JVS CLI Reference: [02_CLI_SPEC.md](02_CLI_SPEC.md)
- Kubernetes Operators: https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
