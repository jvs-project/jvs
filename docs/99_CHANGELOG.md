# Changelog

## v6.1 (2026-02-19)
- Strengthened lock protocol: lease + renew + steal + fencing token.
- Added explicit snapshot consistency levels: `quiesced` and `best_effort`.
- Made descriptor checksum mandatory.
- Tightened dangerous restore semantics: `--force` no longer bypasses lock/fencing checks.
- Added intent/audit/gc control-plane directories in repo layout.
- Added `jvs verify`, lock commands, and GC planning/run commands to CLI spec.
- Added strict migration gates and backup restore drill requirements.
- Unified product wording around safe defaults and verifiable history.

## v6.0 (2026-02-20)
- Removed JVS remote/mirror/push/pull. Migration uses `juicefs sync`.
- Repo layout changed: `repo/main/` is the main worktree payload root.
- Removed backend config from `.jvs/` (JuiceFS prepared externally).
- Snapshot-first terminology: snapshot/history/restore/worktree.
- Snapshot storage fixed: `.jvs/snapshots/<id>/` full directory tree.
- Added snapshot engine selection with CoW-friendly fallback (reflink/copy).
