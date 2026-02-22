# JVS Kubernetes Operator

This directory contains the Kubernetes Operator for JVS (Juicy Versioned Workspaces).

## Overview

The JVS Operator extends Kubernetes with Custom Resource Definitions (CRDs) for managing JVS workspaces and snapshots natively within Kubernetes clusters.

## Components

### Custom Resources

1. **Workspace** (`jvs.io/v1alpha1/Workspace`)
   - Represents a JVS workspace with storage, retention policies, and auto-snapshot configuration
   - Shortname: `jvs`

2. **Snapshot** (`jvs.io/v1alpha1/Snapshot`)
   - Represents a point-in-time snapshot of a workspace
   - Shortname: `jvssnap`

### Controllers

1. **Workspace Controller**
   - Manages workspace lifecycle (creation, updates, deletion)
   - Handles PVC provisioning
   - Schedules automatic snapshots
   - Enforces retention policies

2. **Snapshot Controller**
   - Creates snapshots using the JVS CLI
   - Monitors snapshot completion
   - Handles restore operations

## Quick Start

### Install CRDs and Operator

```bash
# Install CRDs
kubectl apply -f config/crd/bases/

# Install RBAC
kubectl apply -f config/rbac/

# Deploy operator
kubectl apply -f config/manager/
```

### Create a Workspace

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
  retentionPolicy:
    keepMinSnapshots: 20
    keepMinAge: 48h
  autoSnapshot:
    enabled: true
    schedule: "0 */4 * * *"
    template: checkpoint
    maxSnapshots: 50
```

### Create a Snapshot

```yaml
apiVersion: jvs.io/v1alpha1
kind: Snapshot
metadata:
  name: pre-deployment
  namespace: default
spec:
  workspace: ml-workspace
  note: "Before deploying model v2.0"
  tags:
  - pre-deploy
  - release
  compression:
    type: gzip
    level: 6
```

## Development

### Building

```bash
# Build the operator binary
go build -o bin/jvs-operator ./cmd/jvs-operator

# Build Docker image
make -f Makefile.operator docker-build
```

### Running Locally

```bash
# Install CRDs
make -f Makefile.operator install

# Run operator locally (requires kubeconfig)
make -f Makefile.operator run
```

### Testing

```bash
# Run tests
make -f Makefile.operator test
```

## Architecture

```
┌─────────────────────────────────────────────────┐
│              Kubernetes API Server              │
│  Workspace CRD        Snapshot CRD              │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│           JVS Operator Pod                      │
│  ┌──────────────┐     ┌──────────────┐         │
│  │  Workspace   │     │  Snapshot    │         │
│  │  Controller  │     │  Controller  │         │
│  └──────────────┘     └──────────────┘         │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│            Workspace Pods                        │
│  ┌────────────┐   ┌────────────┐               │
│  │   Pod 1    │   │   Pod 2    │               │
│  │ (main/)    │   │ (main/)    │               │
│  └────────────┘   └────────────┘               │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│           Persistent Volume                      │
│  .jvs/ (metadata)     main/ (payload)           │
└─────────────────────────────────────────────────┘
```

## Documentation

See [OPERATOR.md](../../docs/OPERATOR.md) for complete documentation.

## License

Copyright JVS Project Contributors
