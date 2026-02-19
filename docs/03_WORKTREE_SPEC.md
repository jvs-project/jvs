# Worktree Spec (v6.2)

Default mode is isolated (`exclusive`). `shared` is opt-in and high-risk.

## Worktree identity
Each worktree contains:
```
<worktree>/.jvs-worktree/
├── id
├── base_snapshot
├── head_snapshot
└── config.json
```

## Isolation modes
### `exclusive` (default)
- single writer enforced by lock/lease/fencing
- required for deterministic `quiesced` snapshot operation

### `shared` (high-risk)
- multiple writers allowed
- no conflict-resolution semantics in v0.x
- `restore --inplace` disabled by default
- snapshots SHOULD be tagged `best_effort` unless an explicit quiesce policy is active

## Naming and path rules (MUST)
- Name charset: `[a-zA-Z0-9._-]+`
- Name MUST NOT contain separators, `..`, control chars, or empty segments
- Name MUST normalize to NFC before validation
- Canonical resolved path MUST remain under `repo/worktrees/`
- Operations MUST fail on symlink escape detection

## Lifecycle
create -> active -> snapshot -> restore(optional) -> remove

Removing worktree payload does not remove snapshots. Retention is controlled by GC policy.
