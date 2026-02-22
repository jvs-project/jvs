# Migration & Backup (v7.0)

**Note:** For upgrading between JVS versions, see [UPGRADE.md](../UPGRADE.md).

JVS does not provide remote replication. Use JuiceFS replication tools.

## Recommended method
Use `juicefs sync` for repository migration.

## Pre-migration gates (MUST)
1. freeze writers and stop agent jobs
2. ensure no active operations
3. run:
```bash
jvs doctor --strict
jvs verify --all
```
4. take final snapshots for critical worktrees

## Runtime-state policy (MUST)
Runtime state is non-portable and must not be migrated as authoritative state:
- active `.jvs/intents/`

Destination MUST rebuild runtime state:
```bash
jvs doctor --strict --repair-runtime
```

## Migration flow
1. mount source and destination volumes
2. sync repository excluding runtime state
```bash
juicefs sync /mnt/src/myrepo/ /mnt/dst/myrepo/ \
  --exclude '.jvs/intents/**' \
  --update --threads 16
```
3. validate destination
```bash
cd /mnt/dst/myrepo/main
jvs doctor --strict --repair-runtime
jvs verify --all
jvs history --limit 10
```

## What to sync
Portable history state:
- `.jvs/format_version`
- `.jvs/worktrees/`
- `.jvs/snapshots/`
- `.jvs/descriptors/`
- `.jvs/audit/`
- `.jvs/gc/`

Optional payload state:
- `main/`
- selected `worktrees/`

## Restore drill (SHOULD)
1. restore backup to fresh volume
2. run strict doctor + verify
3. restore at least one historical snapshot into new worktree
4. record drill result in operations log
