<p align="center">
  <h1 align="center">JVS</h1>
  <p align="center">
    <strong>Instant workspace snapshots on JuiceFS</strong>
  </p>
  <p align="center">
    <a href="https://github.com/jvs-project/jvs/releases/latest"><img src="https://img.shields.io/github/v/release/jvs-project/jvs?style=flat-square" alt="Release"></a>
    <a href="https://github.com/jvs-project/jvs/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/jvs-project/jvs/ci.yml?branch=main&style=flat-square&label=CI" alt="CI"></a>
    <a href="https://goreportcard.com/report/github.com/jvs-project/jvs"><img src="https://goreportcard.com/badge/github.com/jvs-project/jvs?style=flat-square" alt="Go Report Card"></a>
    <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square" alt="License: MIT"></a>
  </p>
</p>

---

JVS (**Juicy Versioned Workspaces**) takes O(1) snapshots of entire workspace directories using JuiceFS Copy-on-Write. Think of it as `git init` + `git commit` for **any file type at any scale** — datasets, model weights, game assets, agent sandboxes — without staging areas, diffs, or blob graphs.

```bash
jvs init myproject          # create a versioned workspace
cd myproject/main
# ... work on files ...
jvs snapshot "baseline"     # O(1) snapshot via CoW — instant regardless of size
jvs snapshot "experiment-1" --tag exp
jvs restore baseline        # restore to any point
jvs worktree fork "branch"  # fork a parallel workspace
jvs verify --all            # SHA-256 integrity check
```

## Why JVS?

| Problem | JVS approach |
|---|---|
| Git chokes on large binary files | Snapshots the entire directory tree via filesystem CoW — size doesn't matter |
| Dataset versioning tools need a server | Local-first CLI, zero infrastructure — JuiceFS handles storage |
| Agent sandboxes need instant rollback | O(1) restore to any snapshot, fork to create parallel workspaces |
| Need tamper-evident audit trail | Hash-chained audit log + two-layer integrity (descriptor checksum + payload SHA-256) |

## Install

**Download a binary** from the [latest release](https://github.com/jvs-project/jvs/releases/latest) (Linux, macOS, Windows):

```bash
# Linux (amd64)
curl -L https://github.com/jvs-project/jvs/releases/latest/download/jvs-linux-amd64 -o jvs
chmod +x jvs && sudo mv jvs /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/jvs-project/jvs/releases/latest/download/jvs-darwin-arm64 -o jvs
chmod +x jvs && sudo mv jvs /usr/local/bin/
```

**Or build from source** (requires Go 1.25+):

```bash
git clone https://github.com/jvs-project/jvs.git
cd jvs && make build
# binary at bin/jvs
```

## Quick start

### 1. Set up a JuiceFS mount (recommended for O(1) snapshots)

```bash
juicefs format redis://127.0.0.1:6379/1 myvol
juicefs mount redis://127.0.0.1:6379/1 /mnt/jfs -d
```

> JVS also works on any POSIX filesystem — it auto-detects the best engine:
> **juicefs-clone** (O(1)) → **reflink** (O(1) on btrfs/XFS) → **copy** (fallback).

### 2. Create and use a workspace

```bash
cd /mnt/jfs
jvs init myproject
cd myproject/main        # this is your workspace root

echo "hello" > data.txt
jvs snapshot "first version"

echo "world" >> data.txt
jvs snapshot "second version" --tag release

jvs history              # see all snapshots
jvs diff                 # see what changed
jvs restore first        # go back to "first version"
```

### 3. Fork a parallel workspace

```bash
jvs worktree fork "experiment"
cd ../worktrees/experiment
# independent copy — changes here don't affect main
```

## Commands

| Command | What it does |
|---|---|
| `jvs init <name>` | Create a versioned workspace |
| `jvs snapshot [note] [--tag T]` | Snapshot the current state |
| `jvs history [--tag T] [--grep P]` | List snapshots |
| `jvs diff [from [to]]` | Compare two snapshots |
| `jvs restore <id\|note\|tag>` | Restore workspace to a snapshot |
| `jvs worktree fork [id] <name>` | Fork an independent workspace |
| `jvs worktree list\|remove` | Manage worktrees |
| `jvs verify [--all]` | Verify integrity (SHA-256) |
| `jvs doctor [--strict]` | Health check and auto-repair |
| `jvs gc plan` / `jvs gc run` | Two-phase garbage collection |

## How it works

```
myproject/
├── .jvs/                 # control plane (metadata)
│   ├── snapshots/        # snapshot payloads (CoW clones)
│   ├── descriptors/      # snapshot metadata (JSON)
│   ├── audit/            # hash-chained audit log
│   └── gc/               # GC plans and tombstones
├── main/                 # your workspace (data plane)
└── worktrees/            # forked workspaces
```

**Snapshot publish is atomic** — a 12-step protocol (intent → clone → hash → descriptor → READY → rename → head update) ensures snapshots are either fully committed or not visible at all. No partial states.

**Three snapshot engines**, auto-selected per filesystem:

| Engine | Mechanism | Performance | Filesystem |
|---|---|---|---|
| `juicefs-clone` | JuiceFS clone API | **O(1)** | JuiceFS |
| `reflink-copy` | FICLONE ioctl | **O(1)** | btrfs, XFS |
| `copy` | Recursive copy | O(n) | Any POSIX |

## Use cases

- **AI agent sandboxes** — snapshot before each tool call, rollback on failure ([guide](docs/agent_sandbox_quickstart.md))
- **ML experiment tracking** — snapshot datasets + code + configs together, fork for A/B experiments
- **Game development** — version large binary assets that Git can't handle ([guide](docs/game_dev_quickstart.md))
- **ETL pipelines** — checkpoint each pipeline stage, restore to reprocess ([guide](docs/etl_pipeline_quickstart.md))

## Integrity and security

Every snapshot is verified by two independent layers:

1. **Descriptor checksum** — detects metadata corruption
2. **Payload root hash** (SHA-256) — detects data tampering

The audit log is hash-chained (each entry includes the hash of the previous entry), making the history tamper-evident. Run `jvs verify --all` at any time to validate the entire repository.

## Development

```bash
make test           # unit tests
make test-race      # unit tests + race detector
make conformance    # end-to-end black-box tests
make lint           # golangci-lint
make fuzz           # fuzz testing (10s per target)
make release-gate   # full pre-release gate (all of the above)
```

## Documentation

| Document | Description |
|---|---|
| [Quick Start](docs/QUICKSTART.md) | 5-minute tutorial |
| [Architecture](docs/ARCHITECTURE.md) | System design and internals |
| [CLI Spec](docs/02_CLI_SPEC.md) | Complete command reference |
| [Constitution](docs/CONSTITUTION.md) | Core principles and non-goals |
| [Changelog](docs/99_CHANGELOG.md) | Release history |

## License

[MIT](LICENSE)
