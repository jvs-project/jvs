# Worktree Spec (v6.4)

Default mode is isolated (`exclusive`). `shared` is opt-in and high-risk.

## Worktree identity
Worktree metadata is stored centrally under the control plane:
```
repo/.jvs/worktrees/<name>/
└── config.json       # sole authoritative source for worktree metadata
```

`config.json` is the **only** authoritative source for worktree identity and state.
No separate `id`, `base_snapshot`, or `head_snapshot` files exist.

Worktree payload directories (`repo/main/`, `repo/worktrees/<name>/`) contain **pure user data only** — no control-plane artifacts. This ensures `juicefs clone` captures a clean payload without exclusion logic (see `01_REPO_LAYOUT_SPEC.md` §Worktree discovery).

## Isolation modes
### `exclusive` (default)
- single writer enforced by lock/lease/fencing
- required for deterministic `quiesced` snapshot operation

### `shared` (high-risk)
- multiple writers allowed
- no conflict-resolution semantics in v0.x
- `restore --inplace` disabled by default
- snapshots SHOULD be tagged `best_effort` unless an explicit quiesce policy is active

## `config.json` schema (MUST)
Path: `repo/.jvs/worktrees/<name>/config.json`

Required fields:
- `worktree_id`: unique worktree identifier (matches directory name)
- `isolation`: `exclusive` or `shared`
- `created_at`: ISO 8601 timestamp
- `base_snapshot_id`: snapshot ID used to create this worktree (nullable for `main`)
- `head_snapshot_id`: current head snapshot (nullable before first snapshot)

Optional fields:
- `label`: human-readable description
- `snapshot_defaults`: object with default `consistency_level` override

## Naming and path rules (MUST)
- Name charset: `[a-zA-Z0-9._-]+`
- Name MUST NOT contain separators, `..`, control chars, or empty segments
- Name MUST normalize to NFC before validation
- Canonical resolved path MUST remain under `repo/worktrees/`
- Operations MUST fail on symlink escape detection

## Rename and lock interaction (MUST)
- `jvs worktree rename` MUST fail with `E_LOCK_CONFLICT` if the source worktree has an active (non-expired) lock.
- Rationale: renaming changes the worktree identity; an active lock holder references the old identity, making the lock semantically invalid.
- Operator must release or wait for lock expiry before renaming.

## Lifecycle
create -> active -> snapshot -> restore(optional) -> remove

### Remove semantics (MUST)
`jvs worktree remove` MUST:
1. Delete the payload directory (`repo/worktrees/<name>/`).
2. Delete the worktree metadata directory (`.jvs/worktrees/<name>/`).
3. Append audit event recording the removal.

Removing a worktree does not remove its snapshots. Retention is controlled by GC policy.
After removal, the worktree's `head_snapshot_id` no longer exists; its snapshots lose "current head" GC protection and become deletion candidates unless pinned or referenced by a ref.
