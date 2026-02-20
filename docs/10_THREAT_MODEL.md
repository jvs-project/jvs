# Threat Model (v6.3)

## Assets
- snapshot payloads
- descriptors and lineage metadata
- audit trail
- descriptor checksums and payload hashes

## Adversary assumptions
- can read and write files in repository path with compromised local account
- cannot break strong cryptography
- can race operations and attempt stale-lock writes

## Key threats and controls
1. Stale writer continues after lock steal
   Control: fencing token validation before commit.
2. Descriptor and checksum both rewritten
   Control: descriptor checksum + payload root hash detect independent tampering. Coordinated rewrite is a v0.x accepted risk (see 09_SECURITY_MODEL.md).
3. Path traversal on worktree operations
   Control: strict name validation + canonical path boundary checks.
4. Crash during snapshot publish
   Control: tmp+READY protocol + fsync durability sequence.
5. Runtime-state poisoning after migration
   Control: runtime-state exclusion and rebuild at destination.

## Residual risks
- filesystem or kernel bugs bypassing expected durability semantics
- coordinated descriptor + checksum rewrite by attacker with filesystem write access (mitigated by signing in v1.x)

## Risk labeling
Commands and JSON output MUST label high-risk states:
- `best_effort` snapshots
