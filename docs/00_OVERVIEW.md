# Overview

**Document set:** JVS v6.1 (JuiceFS-first, snapshot-first)
**Date:** 2026-02-19

## Core idea
JVS is workspace versioning with full snapshots as the version unit.
- Version unit is a snapshot of one worktree payload root.
- A repo has one main worktree plus zero or more secondary worktrees.
- Worktree selection is filesystem-native (`cd`), no virtual remapping.

## Frozen design decisions
1. No remote/mirror/push/pull in JVS; replication belongs to JuiceFS.
2. Main worktree path is `repo/main/` to preserve clean clone source.
3. Snapshots are stored at `.jvs/snapshots/<snapshot-id>/` and published via READY protocol.
4. Restore defaults to safe mode (new worktree).
5. `exclusive` isolation is default; `shared` is explicitly high-risk.
6. Snapshot consistency level is explicit: `quiesced` (default) or `best_effort`.
7. Integrity is mandatory: descriptor checksum + verification workflow.

## Product promise
- Safe-by-default restore behavior
- Verifiable immutable history
- Filesystem-native scale on JuiceFS

## Non-goals
- Git parity
- merge/rebase engine in v0.x
- central server/auth in v0.x
