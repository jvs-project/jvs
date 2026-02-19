# Security Model (v6.2)

## Scope
Defines trust, integrity, and operational security requirements for JVS repositories.

## Security objectives
- prevent stale-writer corruption via lock + fencing
- detect descriptor and payload tampering
- preserve auditable operation history

## Trust root
Trust state is stored in `.jvs/trust/`.

Required objects:
- `keyring.json` (trusted public keys)
- `policy.json` (verification policy)
- `revocations.json` (revoked key ids)

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

## Non-goals
- encryption-at-rest policy management
- in-JVS authn/authz framework
