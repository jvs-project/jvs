# Changelog

## v6.4 (2026-02-20)
- **Layout migration**: worktree metadata moved from `<worktree>/.jvs-worktree/` to `.jvs/worktrees/<name>/`. Payload roots now contain zero control-plane artifacts. Rationale: `juicefs clone` cannot exclude subdirectories, so payload purity must be structural, not filter-based.
- Added worktree discovery algorithm (walk-up to repo root, path-based name resolution) in 01_REPO_LAYOUT_SPEC.
- `.jvs/worktrees/` added to portable history state.
- Constitution v1.0 → v1.2: added §7.4 Integrity and Verifiability; acknowledged `shared` mode exception in §7.3; strengthened §2.3 control-plane separation; clarified §4.2 rationale; clarified O(1)/O(n) in §1; simplified §5.1 exclusion list (only `.jvs/`).
- Simplified 04_SNAPSHOT_SCOPE: no exclusion logic needed (payloads are pure data).
- Simplified 05_SNAPSHOT_ENGINE: removed `.jvs-worktree/` exclusion invariant; updated head pointer path to `.jvs/worktrees/<name>/config.json`; added O(1)/O(n) performance characteristics per engine.
- Updated 03_WORKTREE_SPEC: centralized metadata under `.jvs/worktrees/`; documented pure-payload guarantee.
- Updated 06_RESTORE_SPEC: metadata write targets `.jvs/worktrees/<name>/config.json`.
- Fixed 18_MIGRATION: added `.jvs/worktrees/` and `.jvs/format_version` to portable sync list.
- Defined worktree remove semantics: deletes both payload directory and `.jvs/worktrees/<name>/` metadata.
- Specified `jvs init` must create `.jvs/worktrees/main/config.json` alongside `repo/main/`.
- Corrected `reflink-copy` performance: O(n) file-count walk with O(1) per-file reflink (not O(1) overall).
- Added conformance tests 28-32 (worktree discovery, payload purity, init metadata, remove cleanup, migration).
- Added traceability Promise 9 (pure payload roots with centralized control plane).

## v6.3 (2026-02-20)
- Consolidated worktree metadata: `config.json` is sole authoritative source; removed redundant `id`/`base_snapshot`/`head_snapshot` files.
- Defined `descriptor_checksum` coverage scope: all fields except checksum, signature, and mutable state fields.
- Added ref-targeted snapshots to GC protection rules.
- Defined lock steal atomic file replacement: write-to-tmp + `rename()` + concurrent stealer resolution.
- Added `--signing-key` and `--json` parameters to `jvs init` in CLI spec.
- Added `E_CONSISTENCY_UNAVAILABLE` to stable error class list.
- Specified canonical JSON rules for deterministic hashing (audit records and descriptor checksums).
- Clarified `audit_repair` scope: recomputes hash chain over present records only; missing records require escalation.
- Added `format_version` to portable history state enumeration.
- Documented worktree remove → GC eligibility implication.
- Added conformance tests 25-27 (ref GC protection, steal atomicity, descriptor checksum computation).
- Defined `refs/` as named snapshot references with immutable-create semantics and CLI (`ref create/list/delete`).
- Defined audit log format: JSON Lines with hash chain for tamper evidence (`E_AUDIT_CHAIN_BROKEN`).
- Defined trust bootstrap flow: `jvs init` generates Ed25519 keypair, writes keyring/policy, private key stored outside `.jvs/`.
- Defined snapshot ID format: `<timestamp_ms>-<random_hex8>` for deterministic ordering and collision avoidance.
- Defined `format_version` file semantics and `E_FORMAT_UNSUPPORTED` error for version mismatch.
- Specified lock atomic creation mechanism (`O_CREAT|O_EXCL` with fsync) and JuiceFS metadata engine guarantee.
- Defined worktree `config.json` schema with required and optional fields.
- Defined `jvs info` required JSON output fields.
- Added worktree rename + active lock conflict rule (`E_LOCK_CONFLICT`).
- Added head-pointer orphan crash recovery scenario and `advance_head` repair action.
- Specified supported algorithm set: SHA-256 for hashing, Ed25519 for signatures.
- Defined payload root hash computation: deterministic sorted-path Merkle walk with SHA-256.
- Specified clock skew detection logic and `E_CLOCK_SKEW_EXCEEDED` trigger conditions.
- Expanded `doctor --repair-runtime` repair action catalogue: `clean_tmp`, `rebuild_index`, `audit_repair`, `advance_head`, `clean_locks`, `clean_intents`.
- Added new error classes: `E_CLOCK_SKEW_EXCEEDED`, `E_SIGNING_KEY_MISSING`, `E_FORMAT_UNSUPPORTED`, `E_AUDIT_CHAIN_BROKEN`.
- Added conformance tests 17-24 covering new spec requirements.
- Added traceability promises 7-8 (auditable history, deterministic identity).
- Document numbering 15-17 intentionally reserved for future specs.

## v6.2 (2026-02-19)
- Closed integrity gap by adding signed `payload_root_hash` to descriptor requirements.
- Changed `jvs verify` default to strong verification; `--allow-unsigned` is explicit downgrade only.
- Replaced GC atomic-delete wording with a two-phase `plan/mark/commit` protocol using `plan_id`.
- Strengthened lock identity with `holder_nonce`, `session_id`, and `acquire_seq`.
- Unified migration policy: runtime state (`locks`, active `intents`) is excluded and rebuilt.
- Aligned README/overview/release/runbook language to avoid over-claims.

## v6.1 (2026-02-18)
- Strengthened lock protocol: lease + renew + steal + fencing token.
- Added explicit snapshot consistency levels: `quiesced` and `best_effort`.
- Made descriptor checksum mandatory.
- Tightened dangerous restore semantics: `--force` no longer bypasses lock/fencing checks.
- Added intent/audit/gc control-plane directories in repo layout.
- Added `jvs verify`, lock commands, and GC planning/run commands to CLI spec.
- Added strict migration gates and backup restore drill requirements.
- Unified product wording around safe defaults and verifiable history.

## v6.0 (2026-02-17)
- Removed JVS remote/mirror/push/pull. Migration uses `juicefs sync`.
- Repo layout changed: `repo/main/` is the main worktree payload root.
- Removed backend config from `.jvs/` (JuiceFS prepared externally).
- Snapshot-first terminology: snapshot/history/restore/worktree.
- Snapshot storage fixed: `.jvs/snapshots/<id>/` full directory tree.
- Added snapshot engine selection with CoW-friendly fallback (reflink/copy).
