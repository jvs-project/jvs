# Overview

**Document set:** JVS v6.4 (JuiceFS-first, snapshot-first)
**Date:** 2026-02-20

## Core idea
JVS versions workspaces by full snapshots of a single worktree payload root.

## Frozen design decisions
1. No remote replication features in JVS; JuiceFS handles transport.
2. Main payload root is `repo/main/`.
3. Snapshot publish is READY-based and auditable.
4. Restore defaults to safe mode (new worktree).
5. `exclusive` is default; `shared` is high-risk and explicitly labeled.
6. Consistency level is explicit: `quiesced` or `best_effort`.
7. Verification default is strong: checksum + payload hash + signature/trust chain.
8. Runtime state (`locks`, active `intents`) is non-portable and rebuilt after migration.

## Product promise
- Safe-by-default restore
- Verifiable and tamper-evident history
- Filesystem-native scale on JuiceFS

## Non-goals
- Git parity and text merge semantics
- in-JVS authn/authz control plane
