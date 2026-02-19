# Restore Spec (v6.2)

## Default restore (SAFE)
`jvs restore <snapshot-id>` creates a new worktree.

Auto-generated worktree name format:
`restore-<shortid>-<YYYYMMDD>-<HHMMSS>-<rand4>`

## Safe restore flow (MUST)
1. Validate snapshot exists and has READY marker.
2. Validate descriptor checksum and signature according to trust policy.
3. Create destination worktree under `repo/worktrees/<name>/` using path-safe checks.
4. Materialize payload from snapshot.
5. Write `.jvs-worktree` metadata.
6. Append audit record and return created path.

## In-place restore (DANGEROUS)
`jvs restore <snapshot-id> --inplace --force --reason <text>`

### Hard requirements (MUST)
- In `exclusive`, caller must hold valid lock.
- Fencing token must match active token.
- Snapshot checksum and signature validation must pass.
- `--reason` must be non-empty for audit.

`--force` bypasses interactive confirmation only.
`--force` MUST NOT bypass lock, fencing, or integrity checks.

### Safety checks (MUST)
Before overwrite:
- record pre-restore head and integrity summary
- record `holder_id`, `fencing_token`, `decision_id`, `reason`

Failure behavior:
- operation must be atomic at worktree boundary, or
- if atomic boundary cannot be guaranteed, system must emit explicit failed state and recovery steps.

## Shared mode
- `restore --inplace` remains disabled by default in `shared`.
- explicit override (future) must emit high-risk warning and audit tag.
