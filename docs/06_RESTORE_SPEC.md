# Restore Spec (v6.1)

## Default restore (SAFE)
`jvs restore <snapshot-id>` creates a new worktree.

If `--name` is omitted, generated name format MUST be:
`restore-<shortid>-<YYYYMMDD>-<HHMMSS>-<rand4>`

## Safe restore flow (MUST)
1. Validate snapshot existence and READY marker.
2. Validate descriptor checksum.
3. Create destination worktree under `repo/worktrees/<name>/`.
4. Materialize payload from snapshot using selected restore engine.
5. Write `.jvs-worktree` metadata (`id`, `base_snapshot`, `head_snapshot`, config).
6. Append audit event and print path (or JSON output).

## In-place restore (DANGEROUS)
`jvs restore <snapshot-id> --inplace --force`

### Hard requirements (MUST)
- In `exclusive` mode, caller MUST hold valid writer lock.
- Fencing token MUST match current lock token.
- Snapshot READY and descriptor checksum MUST pass.

`--force` only means non-interactive confirmation bypass.
`--force` MUST NOT bypass lock or fencing checks.

### Additional protections (MUST)
- Before overwrite, record pre-restore state in audit:
  - target worktree id
  - previous head snapshot
  - caller holder_id and fencing_token
- On failure, command MUST leave current payload either unchanged or explicitly marked failed with recovery instructions.

## Shared mode
- `restore --inplace` is disabled by default in `shared` worktrees.
- If explicitly enabled by future policy, tool MUST emit high-risk warning and audit tag.

## Rename
Users can rename restored worktree:
```bash
jvs worktree rename <old> <new>
```
