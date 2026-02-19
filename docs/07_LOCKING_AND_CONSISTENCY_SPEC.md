# Locking & Consistency Spec (v6.1)

## Goals
- Enforce single-writer semantics for `exclusive` worktrees.
- Make write safety provable under crashes and stale holders.
- Expose snapshot consistency level explicitly to users and tooling.

## Lock scope and storage
- Locks are repository-local files under `repo/.jvs/locks/`.
- Each `exclusive` worktree has exactly one active writer lock.
- `shared` worktrees do not enforce SWMR and are marked high-risk.

## Lock record schema (MUST)
Each lock record MUST include:
- `lock_id` (uuid)
- `worktree_id`
- `holder_id` (`host:user:pid:start_time`)
- `created_at`
- `last_renewed_at`
- `lease_expires_at`
- `fencing_token` (monotonic integer)

## Protocol (MUST)
### Acquire
- Create lock atomically.
- If lock exists and not expired, return lock conflict.
- If expired, acquisition MUST use steal flow.

### Renew
- Only current holder can renew.
- Renew extends `lease_expires_at`.
- Renew failure MUST stop write operations for that holder.

### Steal
- Allowed only when current lock is expired.
- New lock MUST increment `fencing_token`.
- New holder MUST record steal metadata in audit log.

### Release
- Only current holder can release.
- Release by non-holder MUST fail.

## Fencing token rules (MUST)
- All mutating operations in `exclusive` mode (`snapshot`, `restore --inplace`) MUST validate current `fencing_token` before commit.
- If token is stale or mismatched, operation MUST fail with non-zero exit and no partial publish.

## Snapshot consistency levels
### `quiesced` (default, recommended)
- Snapshot source MUST be in a quiesced window.
- In `exclusive` mode this means holder has lock and no concurrent payload writer.
- Descriptor MUST store `consistency_level=quiesced`.

### `best_effort`
- Snapshot can run without guaranteed quiesced window.
- Descriptor MUST store `consistency_level=best_effort`.
- `history` and `info --json` MUST surface a warning flag for these snapshots.

## Operation requirements
- `snapshot`: requires valid lock + fencing token in `exclusive`; allowed without lock in `shared` but tagged by consistency level.
- `history`: lock-free and read-only.
- `restore` (safe mode): lock-free because it creates a new worktree.
- `restore --inplace`: requires valid lock + fencing token in `exclusive`; disabled by default in `shared`.

## READY semantics
- Only snapshots with `.READY` are visible to `history` and `restore`.
- Partial snapshots without `.READY` are ignored.
- `doctor --strict` MUST report and optionally clean orphan tmp/intents.

## Error classes (MUST)
Implementations MUST expose stable machine-readable error classes:
- `E_LOCK_CONFLICT`
- `E_LOCK_EXPIRED`
- `E_LOCK_NOT_HELD`
- `E_FENCING_MISMATCH`
- `E_CONSISTENCY_UNAVAILABLE`
- `E_PARTIAL_SNAPSHOT`
