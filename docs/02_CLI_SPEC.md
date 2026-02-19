# CLI Spec (v6.1)

This spec defines user-facing commands for snapshot-first, worktree-first JVS.

## Conventions
- Commands resolve repo/worktree from current path.
- Non-zero exit on failure.
- `--json` required for machine integration on key commands.
- JVS never changes caller CWD.

## Repository commands
### `jvs init <name>`
Creates `<name>/.jvs/`, `<name>/main/`, `<name>/worktrees/`.

### `jvs info [--json]`
Shows repo root, engine, lock health, and policy flags.

### `jvs doctor [--strict] [--json]`
Checks layout invariants, lock health, READY invariants, intents/tmp cleanup candidates.

### `jvs verify [--snapshot <id>|--all] [--json]`
Verifies descriptor checksums, lineage integrity, and head pointers.

## Worktree commands
### `jvs worktree create <name> [--from <snapshot-id>] [--isolation exclusive|shared]`
Create secondary worktree and metadata.

### `jvs worktree list [--json]`
List worktrees, isolation, head snapshot, lock status.

### `jvs worktree path <name>`
Print absolute worktree path.

### `jvs worktree rename <old> <new>`
Rename worktree and update metadata pointers.

### `jvs worktree remove <name> [--force]`
Remove worktree payload; snapshots remain.

## Lock commands
### `jvs lock status [--worktree <name>] [--json]`
Print lock record, lease expiry, fencing token.

### `jvs lock renew [--worktree <name>] [--json]`
Renew active lock for current holder.

## Snapshot commands
### `jvs snapshot [note] [--consistency quiesced|best_effort] [--json]`
Create snapshot of current worktree payload root.
- Default consistency is `quiesced`.
- In `exclusive`, snapshot requires valid lock and fencing token.

### `jvs history [--limit N] [--json]`
Show lineage from current worktree head.
- Must include flags for `best_effort` snapshots.

## Restore commands
### `jvs restore <snapshot-id> [--name <worktree>] [--json]`
Safe mode: create new worktree from snapshot.

### `jvs restore <snapshot-id> --inplace --force [--json]`
Dangerous mode: overwrite current payload.
- `--force` bypasses prompt only.
- MUST still require valid lock and fencing token.

## GC commands
### `jvs gc plan [--policy <name>] [--json]`
Compute deletable snapshots without mutating state.

### `jvs gc run [--policy <name>] [--json]`
Delete snapshots allowed by policy.
- MUST never delete live heads or lineage-pinned ancestors.

## Stable error classes
`E_LOCK_CONFLICT`, `E_LOCK_EXPIRED`, `E_LOCK_NOT_HELD`, `E_FENCING_MISMATCH`, `E_DESCRIPTOR_CORRUPT`, `E_LINEAGE_BROKEN`, `E_PARTIAL_SNAPSHOT`.

## Compatibility aliases
- `jvs commit` -> `jvs snapshot`
- `jvs log` -> `jvs history`
- `jvs checkout` -> `jvs restore` (safe mode)
