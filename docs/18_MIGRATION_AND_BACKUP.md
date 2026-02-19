# Migration & Backup (v6.1)

JVS does not implement remote/push/pull. Replication is done with JuiceFS tooling.

## Recommended method: `juicefs sync`
`juicefs sync` supports include/exclude rules and incremental transfer.

## Pre-migration gates (MUST)
1. Freeze writers or stop agents.
2. Ensure no active valid writer locks for target repo.
3. Run:
```bash
jvs doctor --strict
jvs verify --all
```
4. Optionally create final quiesced snapshots for critical worktrees.

## Migration flow
1. Mount source and destination volumes.
```bash
juicefs mount <SRC_META> /mnt/src -d
juicefs mount <DST_META> /mnt/dst -d
```
2. Sync a repository directory.
```bash
juicefs sync /mnt/src/myrepo/ /mnt/dst/myrepo/ --update --threads 16
```
3. Post-migration validation.
```bash
cd /mnt/dst/myrepo/main
jvs doctor --strict
jvs verify --all
jvs history --limit 10
```

## What to sync
Mandatory for full historical recovery:
- `.jvs/` (snapshots, descriptors, locks, intents, audit, gc)

Optional:
- `main/` payload
- selected `worktrees/` payloads (often ephemeral)

Example excluding ephemeral worktrees:
```bash
juicefs sync /mnt/src/myrepo/ /mnt/dst/myrepo/ --exclude 'worktrees/**' --update
```

## Backup restore drill (SHOULD)
At regular intervals, perform a drill:
1. Restore backup into a new volume.
2. Run strict doctor + full verify.
3. Confirm lineage continuity and recover at least one historical snapshot into a new worktree.
