# CLI Spec (v6.4)

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
### `jvs init <name> [--signing-key <path>] [--json]`
Create repository skeleton, initialize trust policy, and generate or import signing key.
- Creates `repo/.jvs/` control plane with all required subdirectories.
- Creates `repo/main/` payload directory and `.jvs/worktrees/main/config.json` (main worktree metadata).
- Without `--signing-key`: auto-generate Ed25519 keypair (see `docs/09_SECURITY_MODEL.md` trust bootstrap).
- With `--signing-key <path>`: import existing public key into keyring.

### `jvs info [--json]`
Return engine, policy, lock defaults, and trust policy summary.

Required JSON fields:
- `format_version`
- `snapshot_engine`
- `default_isolation`
- `default_consistency`
- `lease_duration_ms`
- `renew_interval_ms`
- `max_clock_skew_ms`
- `trust_policy_summary` (object: `require_signature`, `require_trusted_key`, `allowed_algorithms`)
- `total_snapshots`
- `total_worktrees`

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

## Ref commands
### `jvs ref create <name> <snapshot-id> [--json]`
Create named snapshot reference. Name validation follows worktree name rules.

### `jvs ref list [--json]`
List all refs with target snapshot ID and creation time.

### `jvs ref delete <name> [--json]`
Delete named reference. Appends audit event.

## Stable error classes
`E_NAME_INVALID`, `E_PATH_ESCAPE`, `E_LOCK_CONFLICT`, `E_LOCK_EXPIRED`, `E_LOCK_NOT_HELD`, `E_FENCING_MISMATCH`, `E_CLOCK_SKEW_EXCEEDED`, `E_CONSISTENCY_UNAVAILABLE`, `E_DESCRIPTOR_CORRUPT`, `E_PAYLOAD_HASH_MISMATCH`, `E_SIGNATURE_INVALID`, `E_SIGNING_KEY_MISSING`, `E_TRUST_POLICY_VIOLATION`, `E_LINEAGE_BROKEN`, `E_PARTIAL_SNAPSHOT`, `E_GC_PLAN_MISMATCH`, `E_FORMAT_UNSUPPORTED`, `E_AUDIT_CHAIN_BROKEN`.
