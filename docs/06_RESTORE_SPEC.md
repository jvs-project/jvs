# Restore Spec (v6.7)

## Default restore (SAFE)
`jvs restore <snapshot-id>` creates a new worktree.

Auto-generated worktree name format:
`restore-<shortid>-<YYYYMMDD>-<HHMMSS>-<rand4>`

## Safe restore flow (MUST)
1. Validate snapshot exists and has READY marker.
2. Validate descriptor checksum.
3. Create destination worktree under `repo/worktrees/<name>/` using path-safe checks.
4. Materialize payload from snapshot.
5. Write worktree metadata to `.jvs/worktrees/<name>/config.json`.
6. Append audit record and return created path.

## In-place restore (DANGEROUS)
`jvs restore <snapshot-id> --inplace --force --reason <text>`

### Requirements (MUST)
- Snapshot checksum validation must pass.
- `--force` is mandatory to confirm the dangerous operation.
- `--reason` must be non-empty for audit trail.

`--force` confirms the user understands this is a destructive operation that will overwrite the current worktree.
`--force` MUST NOT bypass integrity checks.

### Safety checks (MUST)
Before overwrite:
- record pre-restore head and integrity summary
- record `reason` for audit trail

Failure behavior:
- operation must be atomic at worktree boundary, or
- if atomic boundary cannot be guaranteed, system must emit explicit failed state and recovery steps.
