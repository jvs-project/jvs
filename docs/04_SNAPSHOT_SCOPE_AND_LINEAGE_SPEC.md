# Snapshot Scope & Lineage Spec (v6.2)

## Scope (MUST)
A snapshot captures only the current worktree payload root:
- inside `repo/main/` -> source `repo/main/`
- inside `repo/worktrees/<name>/` -> source that worktree root

Snapshots MUST NOT include `.jvs/` or other worktree payload roots.

## Storage and immutability (MUST)
Published snapshots live at:
`repo/.jvs/snapshots/<snapshot-id>/`

After READY publication:
- payload is immutable
- descriptor is immutable
- signature metadata is immutable
- detected mutation marks snapshot `corrupt`

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
- `fencing_token` (nullable only when lockless mode is valid)
- `descriptor_checksum` (`algo`, `value`)
- `payload_root_hash` (`algo`, `value`)
- `signature` (`algo`, `value`)
- `signing_key_id`
- `signed_at`
- `tamper_evidence_state` (`trusted|untrusted|tampered`)
- `integrity_state` (`verified|unverified|corrupt`)

## Signature coverage (MUST)
Descriptor signature MUST cover at least:
- `descriptor_checksum`
- `payload_root_hash`
- `snapshot_id`
- `parent_snapshot_id`
- `created_at`

## Lineage rules
- Lineage is per worktree via `parent_snapshot_id` chain.
- Restoring an older snapshot into a new worktree starts a new lineage branch.
- merge/rebase remains out of scope for v0.x.

## Lineage integrity checks (MUST)
`jvs doctor --strict` and `jvs verify --all` MUST detect:
- missing parent descriptor
- parent cycles
- head pointer mismatch
- descriptor checksum mismatch
- payload hash mismatch
- signature invalid
- trust policy violation

All findings MUST include machine-readable severity.
