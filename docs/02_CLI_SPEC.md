# CLI Spec (v6.2)

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
### `jvs init <name>`
Create repository skeleton.

### `jvs info [--json]`
Return engine, policy, lock defaults, and trust policy summary.

### `jvs doctor [--strict] [--repair-runtime] [--json]`
Validate layout, lineage, READY protocol, runtime-state hygiene, and repair candidates.

### `jvs verify [--snapshot <id>|--all] [--allow-unsigned] [--json]`
Default behavior is strong verification:
- descriptor checksum
- payload root hash
- signature chain and trust policy

`--allow-unsigned` is an explicit downgrade for diagnostic/development use and MUST emit warning state.

Required JSON fields:
- `checksum_valid`
- `payload_hash_valid`
- `signature_valid`
- `trust_chain_valid`
- `tamper_detected`
- `severity`

### `jvs conformance run [--profile dev|release] [--json]`
Execute conformance checks defined in `docs/11_CONFORMANCE_TEST_PLAN.md`.

## Worktree commands
### `jvs worktree create <name> [--from <snapshot-id>] [--isolation exclusive|shared]`
Create worktree with metadata.

### `jvs worktree list [--json]`
List worktrees with isolation, head, and lock state.

### `jvs worktree path <name>`
Print canonical absolute path.

### `jvs worktree rename <old> <new>`
Rename worktree with full path safety checks.

### `jvs worktree remove <name> [--force]`
Remove payload only; snapshots remain.

## Lock commands
### `jvs lock acquire [--worktree <name>] [--lease-ms <n>] [--json]`
Acquire exclusive writer lock.

### `jvs lock status [--worktree <name>] [--json]`
Show holder identity, lease window, fencing token, skew policy.

### `jvs lock renew [--worktree <name>] [--json]`
Renew lease for active holder.

### `jvs lock release [--worktree <name>] [--json]`
Release active lock if caller is holder.

## Snapshot commands
### `jvs snapshot [note] [--consistency quiesced|best_effort] [--json]`
Create snapshot from current payload root.
- default consistency: `quiesced`
- exclusive mode requires valid lock and fencing token

### `jvs history [--limit N] [--json]`
Show lineage from current head.
- must include risk labels for `best_effort`

## Restore commands
### `jvs restore <snapshot-id> [--name <worktree>] [--json]`
Safe mode: create a new worktree.

### `jvs restore <snapshot-id> --inplace --force --reason <text> [--json]`
Danger mode: overwrite current payload.
- `--force` bypasses prompt only
- valid lock + fencing token remain mandatory
- `--reason` is mandatory for audit

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
`E_NAME_INVALID`, `E_PATH_ESCAPE`, `E_LOCK_CONFLICT`, `E_LOCK_EXPIRED`, `E_LOCK_NOT_HELD`, `E_FENCING_MISMATCH`, `E_DESCRIPTOR_CORRUPT`, `E_PAYLOAD_HASH_MISMATCH`, `E_SIGNATURE_INVALID`, `E_TRUST_POLICY_VIOLATION`, `E_LINEAGE_BROKEN`, `E_PARTIAL_SNAPSHOT`, `E_GC_PLAN_MISMATCH`.
