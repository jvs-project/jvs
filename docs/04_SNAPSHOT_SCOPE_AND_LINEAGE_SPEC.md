# Snapshot Scope & Lineage Spec (v6.1)

## Scope (MUST)
A snapshot captures only the current worktree payload root.
- Inside `repo/main/` -> source is `repo/main/`
- Inside `repo/worktrees/<name>/` -> source is that worktree root

Snapshots MUST NOT include `.jvs/` or other worktree directories.

## Storage and immutability (MUST)
Published snapshots are stored at:
`repo/.jvs/snapshots/<snapshot-id>/`

After READY publish:
- snapshot payload is immutable
- descriptor is immutable
- any mutation detection marks repository state as corrupted

## Descriptor schema (MUST)
Path:
`repo/.jvs/descriptors/<snapshot-id>.json`

Required fields:
- `snapshot_id`
- `worktree_id`
- `parent_snapshot_id` (or null)
- `created_at`
- `note` (optional)
- `engine`
- `consistency_level` (`quiesced|best_effort`)
- `fencing_token` (nullable only when lockless is valid)
- `descriptor_checksum` (`algo`, `value`)
- `integrity_state` (`verified|unverified|corrupt`)

## Lineage rules
- Lineage is per worktree via `parent_snapshot_id` chain.
- Restoring older snapshot into new worktree creates new lineage root at restore point.
- Merge/rebase semantics are out of scope for v0.x.

## Lineage integrity checks (MUST)
`jvs doctor --strict` and `jvs verify --all` MUST detect:
- missing parent descriptor
- parent cycle
- head pointer not matching existing descriptor
- checksum mismatch

Detected issues MUST be machine-readable and severity-tagged.
