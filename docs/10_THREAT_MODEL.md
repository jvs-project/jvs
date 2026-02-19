# Threat Model (v6.3)

## Assets
- snapshot payloads
- descriptors and lineage metadata
- trust policy and signing keys
- audit trail

## Adversary assumptions
- can read and write files in repository path with compromised local account
- cannot break strong cryptography
- can race operations and attempt stale-lock writes

## Key threats and controls
1. Stale writer continues after lock steal
   Control: fencing token validation before commit.
2. Descriptor and checksum both rewritten
   Control: signature validation against trust root.
3. Path traversal on worktree operations
   Control: strict name validation + canonical path boundary checks.
4. Crash during snapshot publish
   Control: tmp+READY protocol + fsync durability sequence.
5. Runtime-state poisoning after migration
   Control: runtime-state exclusion and rebuild at destination.

## Residual risks
- compromised trusted signing key
- filesystem or kernel bugs bypassing expected durability semantics
- intentional use of `shared` mode in high-contention workloads

## Risk labeling
Commands and JSON output MUST label high-risk states:
- `shared` mode
- `best_effort` snapshots
- untrusted signature chain
