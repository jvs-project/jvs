# JVS (Juicy Versioned Workspaces)

[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects?jvs-project/badge)](https://bestpractices.coreinfrastructure.org/projects?jvs-project)
[![Go Report Card](https://goreportcard.com/badge/github.com/jvs-project/jvs)](https://goreportcard.com/report/github.com/jvs-project/jvs)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Snapshot-native workspace versioning on top of JuiceFS.**

JVS versions full workspaces as snapshots and provides navigable, verifiable, tamper-evident history.

## Who this is for
- AI agents needing isolated, versioned sandboxes
- AI/Code/Data engineers running reproducible workflows
- Platform teams standardizing workspace lifecycle on JuiceFS

## Design boundaries
- No `remote`/`push`/`pull` in JVS
- No backend credential/storage config in JVS
- No diff/staging/merge object model in v0.x
- No distributed locking (local-first)
- Project Constitution: See /docs/CONSTITUTION.md before proposing new features.

## Core guarantees (v7.0)
- Detached state model: restore to historical snapshots, fork to create branches
- Two-layer integrity: descriptor checksum + payload hash (SHA-256)
- Simple workflow: snapshot, restore, and fork

## Installation

### From Source

```bash
git clone https://github.com/jvs-project/jvs.git
cd jvs
make build
# binary is at bin/jvs
```

## Quickstart

> **New to JVS?** See the [Quick Start Guide](docs/QUICKSTART.md) for a 5-minute tutorial.

**Scenario-specific guides:**
- [Game Development](docs/game_dev_quickstart.md) - Unity/Unreal asset versioning
- [Agent Sandboxes](docs/agent_sandbox_quickstart.md) - AI/ML experiment workflows
- [ETL Pipelines](docs/etl_pipeline_quickstart.md) - Data pipeline versioning

### 1) Prepare a JuiceFS mount (optional but recommended)
```bash
juicefs format redis://127.0.0.1:6379/1 myvol
juicefs mount redis://127.0.0.1:6379/1 /mnt/jfs -d
```

### 2) Create a JVS repository
```bash
cd /mnt/jfs  # or any directory
jvs init myrepo
cd myrepo/main
jvs snapshot "init" --tag v1.0
jvs history
```

## Commands

| Command | Description |
|---------|-------------|
| `jvs init <name>` | Initialize a new repository |
| `jvs snapshot [note] [--tag <tag>]` | Create a snapshot |
| `jvs history [--tag <tag>] [--grep <pattern>]` | Show snapshot history |
| `jvs diff [<from> [<to>]]` | Show differences between snapshots |
| `jvs restore <id>` | Restore worktree to snapshot (inplace) |
| `jvs restore HEAD` | Return to latest state |
| `jvs worktree fork [name]` | Fork from current position |
| `jvs worktree fork <id> <name>` | Fork from snapshot |
| `jvs worktree create/list/remove` | Manage worktrees |
| `jvs verify --all` | Verify all snapshots |
| `jvs doctor` | Check repository health |
| `jvs gc plan/run` | Garbage collection |

## Repository Layout

```
myrepo/
├── .jvs/
│   ├── format_version    # Format version (1)
│   ├── repo_id           # Unique repository ID
│   ├── worktrees/        # Worktree metadata
│   ├── snapshots/        # Snapshot payload directories
│   ├── descriptors/      # Snapshot descriptors (JSON)
│   ├── audit/            # Audit log (JSONL)
│   └── gc/               # GC plans and tombstones
├── main/                 # Main worktree (payload)
└── worktrees/            # Additional worktrees (payload)
```

## Architecture

- **Control plane**: `.jvs/` directory contains all metadata
- **Data plane**: `main/` and `worktrees/` contain pure payload
- **Engines**: juicefs-clone (O(1)), reflink-copy (O(1)), copy (fallback)
- **12-step atomic publish**: Intent → Clone → Hash → Descriptor → READY → Rename → Head

## Development

```bash
# Run unit tests
make test

# Run conformance tests
make conformance

# Build binary
make build

# Run all checks
make verify
```

## Production gate
```bash
jvs doctor --strict
jvs verify --all
```

## Migration / backup
Use `juicefs sync` and exclude runtime state (active `.jvs/intents`).
See `docs/18_MIGRATION_AND_BACKUP.md`.

**Spec version:** v8.1 (2026-02-28)

## Recent Changes
- **v8.1**: Removed Docker, Kubernetes operator, and Terraform provider infrastructure. JVS is a local CLI tool — container orchestration belongs in the consumer (agentsmith).
- **v8.0**: Production hardening — 7 critical bug fixes, 30+ new tests, release gate infrastructure.
- **v7.2**: KISS simplification — removed ~900 lines of unused code.

See [CHANGELOG.md](docs/99_CHANGELOG.md) for full history.
