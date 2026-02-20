# Security Model (v6.5)

## Scope
Defines trust, integrity, and operational security requirements for JVS repositories.

## Security objectives
- prevent stale-writer corruption via lock + fencing
- detect descriptor and payload corruption or tampering via checksums and hashes
- preserve auditable operation history

## Supported algorithms (MUST)

### Hash algorithms
- `sha256` (default): SHA-256 for `descriptor_checksum` and `payload_root_hash`.
- Future additions MUST be registered in this spec before use.

Algorithm identifiers in descriptors MUST match values defined here exactly.

## Integrity model (MUST)
1. descriptor checksum layer
2. payload root hash layer

Snapshot integrity requires both layers to pass.

## Verification policy
- `jvs verify` defaults to strong verification (checksum + payload hash).

## Audit requirements
Every mutating operation MUST append audit record with:
- actor/holder
- operation type
- target snapshot/worktree
- fencing token
- holder nonce/session
- reason for dangerous operations

## Audit log format (MUST)

### Storage
Path: `.jvs/audit/audit.jsonl`

Format: JSON Lines (one JSON object per line, append-only).

### Record schema (MUST)
Each audit record MUST contain:
- `event_id`: unique event identifier (UUID v4)
- `timestamp`: ISO 8601 with timezone
- `operation`: operation type (`snapshot`, `restore`, `lock_acquire`, `lock_steal`, `lock_release`, `gc_run`, `ref_create`, `ref_delete`, `worktree_create`, `worktree_remove`, `worktree_rename`)
- `actor`: holder identity string
- `target`: affected snapshot/worktree ID
- `fencing_token`: current fencing token (nullable)
- `session_id`: caller session ID
- `reason`: mandatory for dangerous operations, nullable otherwise
- `prev_hash`: SHA-256 hash of the previous audit record (empty string for first record)
- `record_hash`: SHA-256 hash of this record (all fields except `record_hash` itself, serialized as canonical JSON)

Canonical JSON rules for `record_hash` computation:
- keys sorted lexicographically by Unicode code point
- no whitespace between tokens
- UTF-8 encoding
- strings escaped per RFC 8259
- numbers: no leading zeros, no trailing zeros in fractions, no positive sign
- null values serialized as `null`

### Integrity chain (MUST)
- Each record includes `prev_hash` linking to the prior record, forming a hash chain.
- `jvs doctor --strict` MUST validate the audit hash chain and report `E_AUDIT_CHAIN_BROKEN` on mismatch.
- `jvs verify --all` MUST include audit chain integrity in its checks.

### Rotation (SHOULD)
- When `audit.jsonl` exceeds 100 MB, rotate to `audit-<timestamp>.jsonl`.
- Rotated files are portable history state and included in migration.
- Rotation appends a final chain-closing record to the old file and a chain-opening record to the new file with `prev_hash` referencing the old file's last `record_hash`.

## v0.x accepted risks
- An attacker with filesystem write access can rewrite a descriptor and its checksum consistently without detection. Descriptor signing (planned for v1.x) will close this gap.
- This risk is acceptable for v0.x local single-user and agent workflows.

## Non-goals
- encryption-at-rest policy management
- in-JVS authn/authz framework
- Descriptor signing and trust policy (deferred to v1.x)
