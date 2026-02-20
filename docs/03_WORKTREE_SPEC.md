# Worktree Spec (v6.7)

## Worktree identity
Worktree metadata is stored centrally under the control plane:
```
repo/.jvs/worktrees/<name>/
└── config.json       # sole authoritative source for worktree metadata
```

`config.json` is the **only** authoritative source for worktree identity and state.
No separate `id`, `base_snapshot`, or `head_snapshot` files exist.

Worktree payload directories (`repo/main/`, `repo/worktrees/<name>/`) contain **pure user data only** — no control-plane artifacts. This ensures `juicefs clone` captures a clean payload without exclusion logic (see `01_REPO_LAYOUT_SPEC.md` §Worktree discovery).

## `config.json` schema (MUST)
Path: `repo/.jvs/worktrees/<name>/config.json`

Required fields:
- `name`: worktree name (matches directory name)
- `created_at`: ISO 8601 timestamp
- `base_snapshot_id`: snapshot ID used to create this worktree (nullable for `main`)
- `head_snapshot_id`: current head snapshot (nullable before first snapshot)

Optional fields:
- `label`: human-readable description

## Naming and path rules (MUST)
- Name charset: `[a-zA-Z0-9._-]+`
- Name MUST NOT contain separators, `..`, control chars, or empty segments
- Name MUST normalize to NFC before validation
- Canonical resolved path MUST remain under `repo/worktrees/` or be `repo/main/`
- Operations MUST fail on symlink escape detection

## Lifecycle
create -> active -> snapshot -> restore(optional) -> remove

### Remove semantics (MUST)
`jvs worktree remove` MUST:
1. Delete the payload directory (`repo/worktrees/<name>/`).
2. Delete the worktree metadata directory (`.jvs/worktrees/<name>/`).
3. Append audit event recording the removal.

Removing a worktree does not remove its snapshots. Retention is controlled by GC policy.
After removal, the worktree's `head_snapshot_id` no longer exists; its snapshots lose "current head" GC protection and become deletion candidates unless pinned.
