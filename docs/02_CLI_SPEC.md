# CLI Spec (v7.0)

This spec defines the JVS command contract.

## Conventions
- Commands resolve repository and worktree from current path.
- Non-zero exit on error.
- `--json` is required for machine integration.
- JVS does not mutate caller CWD.

## Path and name safety (MUST)
For all commands accepting `<name>` or path-like values:
- reject empty names
- reject `/`, `\\`, and `..`
- reject absolute paths
- normalize Unicode to NFC before validation
- resolve canonical target and enforce it remains under repo root
- reject symlink escape outside repo root

## Repository commands
### `jvs init <name> [--json]`
Create repository skeleton.
- Creates `repo/.jvs/` control plane with all required subdirectories.
- Creates `repo/main/` payload directory and `.jvs/worktrees/main/config.json` (main worktree metadata).

### `jvs info [--json]`
Return engine, policy, and trust policy summary.

Required JSON fields:
- `format_version`
- `snapshot_engine`
- `total_snapshots`
- `total_worktrees`

### `jvs doctor [--strict] [--repair-runtime] [--json]`
Validate layout, lineage, READY protocol, runtime-state hygiene, and repair candidates.

### `jvs verify [--snapshot <id>|--all] [--json]`
Default behavior is strong verification:
- descriptor checksum
- payload root hash

Required JSON fields:
- `checksum_valid`
- `payload_hash_valid`
- `tamper_detected`
- `severity`

### `jvs conformance run [--profile dev|full|ci] [--json]`
Execute conformance checks defined in `docs/11_CONFORMANCE_TEST_PLAN.md`.

Profiles:
- `dev`: Development profile, runs with `-short` flag (default)
- `full`: Full test suite including slow tests
- `ci`: CI profile with JSON output formatting

## Worktree commands
### `jvs worktree create <name> [--from <snapshot-id>]`
Create worktree with metadata.

### `jvs worktree list [--json]`
List worktrees with head snapshot.

### `jvs worktree path <name>`
Print canonical absolute path.

### `jvs worktree rename <old> <new>`
Rename worktree with full path safety checks.

### `jvs worktree remove <name> [--force]`
Remove payload only; snapshots remain.

## Snapshot commands
### `jvs snapshot [note] [--tag <tag>]... [--json]`
Create snapshot from current payload root.
- Captures the current state of the worktree at a point in time.
- `--tag` may be repeated to attach multiple tags.
- Tag format: `[a-zA-Z0-9._-]+`

### `jvs history [--limit N] [--grep <pattern>] [--tag <tag>] [--all] [--json]`
Show snapshot history.
- `--limit N` limits output to N entries
- `--grep <pattern>` filters by note substring
- `--tag <tag>` filters by tag
- `--all` shows all snapshots (not just current worktree lineage)

### `jvs diff [<from> [<to>]] [--stat] [--json]`
Show differences between two snapshots.
- With no arguments: compares the two most recent snapshots
- With one argument: compares that snapshot with itself (full output)
- With two arguments: compares from-snapshot to to-snapshot
- `--stat` shows summary statistics only
- Snapshot references can be: full ID, short ID prefix, tag name, or `HEAD`

Required JSON fields:
- `from_snapshot_id`
- `to_snapshot_id`
- `from_time`
- `to_time`
- `added` - array of added file paths with metadata
- `removed` - array of removed file paths with metadata
- `modified` - array of modified file paths with old/new sizes
- `total_added`, `total_removed`, `total_modified`

## Restore commands
### `jvs restore <snapshot-id> [-i | --interactive] [--json]`
Inplace restore: restore current worktree to the specified snapshot.
- `<snapshot-id>` can be a full ID, short ID prefix, tag name, or note prefix (fuzzy match)
- After restore, worktree enters **detached state** (unless restoring to HEAD)
- In detached state, cannot create new snapshots
- `--interactive` (`-i`): Shows fuzzy-matched snapshots with confirmation prompt

### `jvs restore HEAD [--json]`
Return to latest state: restore worktree to its latest snapshot.
- Exits detached state
- Worktree returns to HEAD state where snapshots can be created

## Fork commands
### `jvs worktree fork <name> [--json]`
Fork from current position: create a new worktree from the current snapshot.
- Uses current worktree's `head_snapshot_id` as the base

### `jvs worktree fork <snapshot-id> <name> [--json]`
Fork from snapshot: create a new worktree from a specific snapshot.
- `<snapshot-id>` can be a full ID, short ID prefix, tag name, or note prefix (fuzzy match)
- New worktree starts at HEAD state (can create snapshots)

## GC commands
### `jvs gc plan [--policy <name>] [--json]`
Compute deletion candidates only.

Required JSON fields:
- `plan_id`
- `candidate_count`
- `protected_by_pin`
- `protected_by_lineage`
- `deletable_bytes_estimate`

### `jvs gc run --plan-id <id> [--json]`
Execute two-phase deletion for an accepted plan.

## Stable error classes
`E_NAME_INVALID`, `E_PATH_ESCAPE`, `E_DESCRIPTOR_CORRUPT`, `E_PAYLOAD_HASH_MISMATCH`, `E_LINEAGE_BROKEN`, `E_PARTIAL_SNAPSHOT`, `E_GC_PLAN_MISMATCH`, `E_FORMAT_UNSUPPORTED`, `E_AUDIT_CHAIN_BROKEN`.
