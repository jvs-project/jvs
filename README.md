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

## Core guarantees (v6.2)
- Safe default restore: `jvs restore <id>` creates a new worktree
- Strong exclusive writer safety: lock + lease + fencing
- Strong default verification: descriptor checksum + payload hash + signature chain
- Explicit risk labels for `shared` and `best_effort`

## Quickstart
### 1) Prepare a JuiceFS mount
```bash
juicefs format redis://127.0.0.1:6379/1 myvol
juicefs mount redis://127.0.0.1:6379/1 /mnt/jfs -d
```

### 2) Create a JVS repository
```bash
cd /mnt/jfs
jvs init myrepo
cd myrepo/main
jvs lock acquire
jvs snapshot "init" --consistency quiesced
jvs history
```

## Production gate
```bash
jvs doctor --strict
jvs verify --all
jvs conformance run --profile release
```

## Migration / backup
Use `juicefs sync` and exclude runtime state (`.jvs/locks`, active `.jvs/intents`).
See `docs/18_MIGRATION_AND_BACKUP.md`.

**Spec version:** v6.2 (2026-02-19)
