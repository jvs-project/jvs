# Snapshot Scope & Lineage Spec (v6.7)

## Snapshot ID generation (MUST)

Format: `<timestamp_ms>-<random_hex8>`
- `<timestamp_ms>`: Unix epoch milliseconds at snapshot creation, zero-padded to 13 digits.
- `<random_hex8>`: 8 lowercase hex characters from cryptographic random source.
- Example: `1708300800000-a3f7c1b2`

Properties:
- Lexicographic sort approximates creation order.
- Collision probability is negligible (32-bit random within same millisecond).
- `shortid` (used in restore auto-naming) is the first 8 characters of the full ID.
- Snapshot IDs MUST be treated as opaque strings by consumers; ordering is advisory only.

## Scope (MUST)
A snapshot captures only the current worktree payload root:
- inside `repo/main/` -> source `repo/main/`
- inside `repo/worktrees/<name>/` -> source that worktree root

Payload roots contain pure user data (no control-plane artifacts), so no exclusion logic is required.

Snapshots MUST NOT include:
- `.jvs/` directory
- other worktree payload roots

## Storage and immutability (MUST)
Published snapshots live at:
`repo/.jvs/snapshots/<snapshot-id>/`

After READY publication:
- payload is immutable
- descriptor is immutable
- detected mutation marks snapshot `corrupt`

## Descriptor schema (MUST)
Path:
`repo/.jvs/descriptors/<snapshot-id>.json`

Required fields:
- `snapshot_id`
- `worktree_name`
- `parent_id` (or null)
- `created_at`
- `note` (optional)
- `tags` (optional array)
- `engine`
- `descriptor_checksum`
- `payload_root_hash`
- `integrity_state` (`verified|unverified|corrupt`)

## Descriptor checksum coverage (MUST)
`descriptor_checksum` is computed over all descriptor fields **except**:
- `descriptor_checksum` itself
- `integrity_state`

Computation:
1. Serialize covered fields as canonical JSON (sorted keys, no whitespace, UTF-8, no trailing zeros in numbers).
2. Compute SHA-256 of the serialized bytes.

## Lineage rules
- Lineage is per worktree via `parent_id` chain.
- Restoring an older snapshot into a new worktree starts a new lineage branch.
- merge/rebase remains out of scope for v0.x.

## Lineage integrity checks (MUST)
`jvs doctor --strict` and `jvs verify --all` MUST detect:
- missing parent descriptor
- parent cycles
- head pointer mismatch
- descriptor checksum mismatch
- payload hash mismatch

All findings MUST include machine-readable severity.
