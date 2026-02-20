# Overview

**Document set:** JVS v6.5 (JuiceFS-first, snapshot-first)
**Date:** 2026-02-20

## Core idea
JVS versions workspaces by full snapshots of a single worktree payload root.

## Frozen design decisions
1. No remote replication features in JVS; JuiceFS handles transport.
2. Main payload root is `repo/main/`.
3. Snapshot publish is READY-based and auditable.
4. Restore defaults to safe mode (new worktree).
5. v0.x supports `exclusive` mode only; `shared` mode is deferred to v1.x.
6. Consistency level is explicit: `quiesced` or `best_effort`.
7. Verification default is strong: checksum + payload hash. Signature/trust chain deferred to v1.x.
8. Runtime state (`locks`, active `intents`) is non-portable and rebuilt after migration.

## Product promise
- Safe-by-default restore
- Verifiable and tamper-evident history
- Filesystem-native scale on JuiceFS

## v0.x scope limitations
The following Constitution features are architecturally planned but deferred from v0.x implementation:
- **Shared isolation mode** (Constitution §7.3): v0.x enforces exclusive-only. Shared mode requires conflict-resolution semantics not yet designed.
- **Descriptor signing and trust policy** (Constitution §7.4 justification): v0.x integrity relies on descriptor checksum + payload root hash. Signing adds protection against coordinated checksum+descriptor rewrite by an attacker with filesystem write access; this threat is accepted as residual risk in v0.x.

Descriptor schema reserves optional fields for future signature support to ensure forward compatibility.

## Non-goals
- Git parity and text merge semantics
- in-JVS authn/authz control plane
