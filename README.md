# JVS (Juicy Versioned Workspaces)

**Snapshot-native workspace versioning on top of JuiceFS.**

JVS tracks full workspaces as snapshots and exposes navigable, verifiable history.

## Who this is for
- AI agents that need isolated, versioned sandboxes
- AI/Code/Data engineers working with large files and reproducible runs
- Platform teams using JuiceFS and needing workspace version semantics

## Design boundaries
- No `remote`/`push`/`pull` in JVS
- No backend credential/storage config in JVS
- No diff/staging/merge object model in v0.x

## Core guarantees (v6.1)
- Safe default restore: `jvs restore <id>` creates a new worktree
- Verifiable history: snapshots require READY and descriptor checksums
- Strong writer safety in `exclusive` mode: lease lock + fencing token

## Quickstart
### 1) Prepare a JuiceFS mount
```bash
juicefs format redis://127.0.0.1:6379/1 myvol
juicefs mount redis://127.0.0.1:6379/1 /mnt/jfs -d
```

### 2) Create a JVS repo
```bash
cd /mnt/jfs
jvs init myrepo
cd myrepo/main
echo hello > a.txt
jvs snapshot "init"
jvs history
```

### 3) Create an isolated worktree
```bash
cd /mnt/jfs/myrepo
jvs worktree create run-001
cd worktrees/run-001
jvs snapshot "after agent run"
```

## Production checks
Before migration/backup or critical recovery:
```bash
jvs doctor --strict
jvs verify --all
```

## Migration / backup
Use `juicefs sync` for cross-environment migration:
```bash
juicefs mount redis://SRC /mnt/src -d
juicefs mount redis://DST /mnt/dst -d
juicefs sync /mnt/src/myrepo/ /mnt/dst/myrepo/ --update --threads 16
```

See `docs/18_MIGRATION_AND_BACKUP.md`.

**Spec version:** v6.1 (2026-02-19)
