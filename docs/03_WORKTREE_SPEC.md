# Worktree Spec (v6.1)

Mode C remains: default isolated (`exclusive`), optional `shared`.

## Worktree identity
A worktree is identified by directory path plus stable id metadata in `.jvs-worktree/`.

Required files:
```
<worktree>/.jvs-worktree/
├── id
├── base_snapshot
├── head_snapshot
└── config.json
```

## Isolation modes
### `exclusive` (default)
- Single writer enforced by lock + lease + fencing token.
- Required for deterministic `quiesced` snapshot behavior.

### `shared` (high-risk)
- Multiple writers allowed.
- No conflict resolution in v0.x.
- `restore --inplace` disabled by default.
- Snapshots from shared mode SHOULD be labeled `best_effort` unless explicit quiesce policy is active.

## Naming rules
- Worktree names MUST be filesystem-safe identifiers.
- Reserved auto-restore prefix: `restore-`.

## Lifecycle
create -> active -> snapshot(s) -> optional restore -> remove

Removing a worktree does not delete snapshots; snapshot retention is managed by GC policy.
