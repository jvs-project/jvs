# Changelog

## v7.0 — 2026-02-22

### Redesigned restore command with detached state model
- **Single behavior restore**: `jvs restore <id>` now always operates inplace (no more --inplace/--force/--reason flags)
- **Detached state**: After restoring to a historical snapshot, the worktree enters "detached state" where snapshots cannot be created
- **Fork command**: `jvs worktree fork` creates a new worktree from current or specified snapshot (replaces old "safe restore" behavior)
- **Restore HEAD**: `jvs restore HEAD` returns the worktree to the latest snapshot (exits detached state)
- **History markers**: `jvs history` now shows [HEAD] marker and "you are here" indicator

### Worktree config changes
- Added `latest_snapshot_id` field to worktree config
- `IsDetached()` returns true when `head_snapshot_id != latest_snapshot_id`
- `CanSnapshot()` returns true only when not detached and has at least one snapshot

### Migration from v6.x
In v6.x, `restore` had two modes:
- Default: created new worktree (`SafeRestore`)
- `--inplace --force --reason`: overwrote current worktree

In v7.0:
- `restore` always does inplace restore
- Use `worktree fork` to create new worktree from snapshot
- No more `--inplace`, `--force`, `--reason` flags

### Affected files
- Modified: `internal/cli/restore.go`, `internal/cli/snapshot.go`, `internal/cli/worktree.go`, `internal/cli/history.go`, `internal/restore/restorer.go`, `internal/worktree/manager.go`, `pkg/model/worktree.go`
- Updated tests: `internal/restore/restorer_test.go`, `test/conformance/fencing_test.go`
- Updated docs: `06_RESTORE_SPEC.md`, `02_CLI_SPEC.md`, `11_CONFORMANCE_TEST_PLAN.md`, `CONSTITUTION.md`, `README.md`, `CLAUDE.md`
- Added: `docs/20_USER_SCENARIOS.md`

## v6.7 — 2026-02-20

### Removed lock mechanism entirely
- **Lock subsystem removed**: The entire lock/lease/fencing mechanism has been removed. For a local workspace versioning system, distributed locking was over-engineered and provided a false sense of security (it only protected snapshot/restore, not actual file operations).
- **Simplified snapshot**: `jvs snapshot` no longer requires acquiring a lock first. Just snapshot whenever you want.
- **Simplified restore**: `jvs restore --inplace` now only requires `--force --reason` (no fencing token).
- Removed `jvs lock acquire/release/renew/status/steal` commands.
- Removed `FencingToken` field from snapshot descriptor.
- Removed lock-related audit event types: `EventTypeLockAcquire`, `EventTypeLockRelease`, `EventTypeLockSteal`.
- Removed error classes: `E_LOCK_CONFLICT`, `E_LOCK_EXPIRED`, `E_LOCK_NOT_HELD`, `E_FENCING_MISMATCH`, `E_CLOCK_SKEW_EXCEEDED`, `E_CONSISTENCY_UNAVAILABLE`.
- Removed `.jvs/locks/` directory from repository layout.
- Removed `isolation` field from worktree config (was always `exclusive`).
- Updated conformance tests: removed lock-related tests, added in-place restore requirements tests.

### Rationale
The lock mechanism was designed for distributed systems but JVS is local-first. Users should coordinate their own workflow (when to snapshot, when to restore). The lock only checked at snapshot/restore time but didn't prevent concurrent file writes anyway.

### Affected files
- Deleted: `internal/lock/manager.go`, `internal/lock/manager_test.go`, `internal/cli/lock.go`, `pkg/model/lock.go`, `test/conformance/lock_test.go`, `docs/07_LOCKING_AND_CONSISTENCY_SPEC.md`
- Modified: `internal/cli/snapshot.go`, `internal/cli/restore.go`, `internal/snapshot/creator.go`, `internal/restore/restorer.go`, `internal/repo/repo.go`, `pkg/model/snapshot.go`, `pkg/model/audit.go`, `internal/cli/root.go`, `internal/cli/root_test.go`, `internal/doctor/doctor.go`, `internal/gc/collector.go`
- Updated docs: 00_OVERVIEW, 01_REPO_LAYOUT_SPEC, 02_CLI_SPEC, 03_WORKTREE_SPEC, 04_SNAPSHOT_SCOPE_AND_LINEAGE_SPEC, 05_SNAPSHOT_ENGINE_SPEC, 06_RESTORE_SPEC, 11_CONFORMANCE_TEST_PLAN, 14_TRACEABILITY_MATRIX, 18_MIGRATION_AND_BACKUP, README.md

## v6.6 — 2026-02-20

### Simplified snapshot references: tags replace refs
- **Removed refs subsystem**: The separate `.jvs/refs/` directory and `jvs ref` commands are removed. For local workspace versioning, a separate refs feature was over-engineered.
- **Added tags to snapshots**: Tags are now embedded directly in snapshot descriptors as a `tags` array field. This provides a simpler UX - users just work with snapshots and optional tags.
- `jvs snapshot` now accepts `--tag <tag>` (repeatable) to attach tags during snapshot creation.
- `jvs history` now supports `--grep <pattern>` (filter by note), `--tag <tag>` (filter by tag), and `--all` (show all snapshots, not just current worktree lineage).
- `jvs restore` now supports fuzzy snapshot lookup: snapshot-id can be a full ID, short ID prefix, tag name, or note prefix.
- `jvs restore --latest-tag <tag>` restores the most recent snapshot with the given tag.
- Removed `EventTypeRefCreate` and `EventTypeRefDelete` audit event types.
- Removed refs from GC protection rules (tags are embedded in descriptors, so tagged snapshots are protected by lineage).
- Updated conformance tests: removed ref-related tests, added tag validation test.

### Affected files
- Deleted: `internal/ref/manager.go`, `internal/ref/manager_test.go`, `internal/cli/ref.go`, `pkg/model/ref.go`, `test/conformance/ref_test.go`
- Modified: `internal/gc/collector.go`, `internal/repo/repo.go`, `internal/repo/repo_test.go`, `pkg/model/audit.go`, `pkg/model/snapshot.go`, `internal/snapshot/creator.go`, `internal/cli/snapshot.go`, `internal/cli/history.go`, `internal/cli/restore.go`, `internal/cli/root_test.go`
- Added: `internal/snapshot/catalog.go`
- Updated docs: 01_REPO_LAYOUT_SPEC, 02_CLI_SPEC, 03_WORKTREE_SPEC, 08_GC_SPEC, 11_CONFORMANCE_TEST_PLAN, 14_TRACEABILITY_MATRIX, 18_MIGRATION_AND_BACKUP, README.md

## v6.5 — 2026-02-20

### Scope simplifications for v0.x implementation
- **Shared mode deferred to v1.x**: v0.x supports `exclusive` isolation only. Shared mode requires conflict-resolution semantics not yet designed. Worktree creation no longer accepts `--isolation shared`.
- **Signing system deferred to v1.x**: v0.x integrity relies on descriptor checksum + payload root hash (2-layer model). Descriptor signing, trust policy, keyring, and key lifecycle are architecturally planned but removed from v0.x scope. Coordinated descriptor+checksum rewrite by an attacker with filesystem access is accepted as v0.x residual risk.
- Removed `.jvs/trust/` directory from on-disk layout.
- Removed `--signing-key` from `jvs init`, `--allow-unsigned` from `jvs verify`.
- Removed error classes: `E_SIGNATURE_INVALID`, `E_SIGNING_KEY_MISSING`, `E_TRUST_POLICY_VIOLATION`.
- Descriptor schema removes: `signature`, `signing_key_id`, `signed_at`, `tamper_evidence_state`.
- Conformance tests reduced from 32 to 29 (removed signature/trust tests 10, 11, 21).
- Pin CLI (`jvs gc pin/unpin`) noted as v1.x planned feature.

### Affected specs
- 00_OVERVIEW.md, 01_REPO_LAYOUT_SPEC.md, 02_CLI_SPEC.md, 03_WORKTREE_SPEC.md
- 04_SNAPSHOT_SCOPE_AND_LINEAGE_SPEC.md, 05_SNAPSHOT_ENGINE_SPEC.md
- 06_RESTORE_SPEC.md, 09_SECURITY_MODEL.md, 10_THREAT_MODEL.md
- 11_CONFORMANCE_TEST_PLAN.md, 14_TRACEABILITY_MATRIX.md, 18_MIGRATION_AND_BACKUP.md

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
