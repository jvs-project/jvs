# CLI Spec (v6.7)

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

### `jvs conformance run [--profile dev|release] [--json]`
Execute conformance checks defined in `docs/11_CONFORMANCE_TEST_PLAN.md`.

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

## Restore commands
### `jvs restore <snapshot-id> [--name <worktree>] [--json]`
Safe mode: create a new worktree.
- `<snapshot-id>` can be a full ID, short ID prefix, tag name, or note prefix (fuzzy match)

### `jvs restore --latest-tag <tag> [--name <worktree>] [--json]`
Safe mode: restore the most recent snapshot with the given tag.

### `jvs restore <snapshot-id> --inplace --force --reason <text> [--json]`
Danger mode: overwrite current payload.
- `--force` is mandatory to confirm the dangerous operation.
- `--reason` is mandatory for audit trail.

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
