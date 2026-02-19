# Security Model (v6.3)

## Scope
Defines trust, integrity, and operational security requirements for JVS repositories.

## Security objectives
- prevent stale-writer corruption via lock + fencing
- detect descriptor and payload tampering
- preserve auditable operation history

## Supported algorithms (MUST)

### Hash algorithms
- `sha256` (default): SHA-256 for `descriptor_checksum` and `payload_root_hash`.
- Future additions MUST be registered in this spec before use.

### Signature algorithms
- `ed25519` (default): Ed25519 (RFC 8032) for descriptor signing.
- Future additions MUST be registered in this spec before use.

Algorithm identifiers in descriptors MUST match values defined here exactly.

## Trust root
Trust state is stored in `.jvs/trust/`.

Required objects:
- `keyring.json` (trusted public keys)
- `policy.json` (verification policy)
- `revocations.json` (revoked key ids)

## Trust bootstrap (MUST)

### On `jvs init`
1. If no `--signing-key` is provided:
   - generate an Ed25519 keypair.
   - write public key to `.jvs/trust/keyring.json` with `trusted_since = now`.
   - write private key to `$JVS_SIGNING_KEY_PATH` (default: `~/.jvs/keys/<repo-id>.key`).
   - private keys MUST NOT be stored inside `.jvs/`.
2. If `--signing-key <path>` is provided:
   - import public key into keyring with `trusted_since = now`.
   - validate key format; fail on mismatch.
3. Write default `policy.json`:
   - `require_signature: true`
   - `require_trusted_key: true`
   - `allowed_algorithms: ["ed25519"]`
4. Write empty `revocations.json`: `{"revoked_keys": []}`.

### Keyring schema (MUST)
```json
{
  "keys": [
    {
      "key_id": "<hex fingerprint>",
      "algorithm": "ed25519",
      "public_key": "<base64>",
      "trusted_since": "<ISO 8601>",
      "label": "<optional human name>"
    }
  ]
}
```

### Signing key resolution
At snapshot time, JVS resolves the signing key:
1. `$JVS_SIGNING_KEY` environment variable (inline key or file path).
2. `$JVS_SIGNING_KEY_PATH` file path.
3. Default path `~/.jvs/keys/<repo-id>.key`.
4. If none found, fail with `E_SIGNING_KEY_MISSING`.

## Integrity model (MUST)
1. descriptor checksum layer
2. payload root hash layer
3. signature/trust layer

Snapshot trust requires all three to pass.

## Verification policy
- `jvs verify` defaults to strong verification (checksum + payload hash + signature/trust chain).
- `--allow-unsigned` is explicit downgrade for diagnostic use and MUST emit warning severity.
- release profile MUST NOT use downgrade mode.

## Key lifecycle
- key rotation MUST be documented in audit
- revocations are effective from declared timestamp
- signatures from revoked keys after effective timestamp are invalid

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

## Non-goals
- encryption-at-rest policy management
- in-JVS authn/authz framework
