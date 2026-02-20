# JVS (Juicy Versioned Workspaces)

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
- Project Constitution: See /docs/CONSTITUTION.md before proposing new features.

## Core guarantees (v6.5)
- Safe default restore: `jvs restore <id>` creates a new worktree
- Strong exclusive writer safety: lock + lease + fencing
- Two-layer integrity: descriptor checksum + payload hash (SHA-256)
- Exclusive mode only in v0.x (shared mode deferred to v1.x)

## Installation

```bash
git clone https://github.com/jvs-project/jvs.git
cd jvs
make build
```

## Quickstart

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
jvs lock acquire
jvs snapshot "init"
jvs history
jvs lock release
```

## Commands

| Command | Description |
|---------|-------------|
| `jvs init <name>` | Initialize a new repository |
| `jvs snapshot [note]` | Create a snapshot (requires lock) |
| `jvs history` | Show snapshot history |
| `jvs restore <id>` | Restore to new worktree (safe) |
| `jvs restore <id> --inplace --force --reason <text>` | Overwrite current worktree |
| `jvs lock acquire` | Acquire exclusive lock |
| `jvs lock release` | Release lock |
| `jvs lock status` | Show lock status |
| `jvs worktree create/list/remove` | Manage worktrees |
| `jvs ref create/list/delete` | Manage named references |
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
│   ├── refs/             # Named references
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
Use `juicefs sync` and exclude runtime state (`.jvs/locks`, active `.jvs/intents`).
See `docs/18_MIGRATION_AND_BACKUP.md`.

**Spec version:** v6.5 (2026-02-20)
