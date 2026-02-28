# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Nature

This is a **specification + implementation repository** for JVS (Juicy Versioned Workspaces). It contains design specifications (docs/), an implementation design plan (docs/plans/), and a Go implementation of a workspace versioning system built on JuiceFS.

## Core Architecture

JVS is a **snapshot-first, filesystem-native versioning layer** (not a Git replacement):

- **Control Plane vs Data Plane Separation**: `.jvs/` holds all metadata (snapshots, descriptors, locks, worktree config); worktree directories contain pure payload
- **Main worktree at `repo/main/`**: The repo root is NOT the workspace - `repo/main/` is the primary payload root
- **Real directories, no virtualization**: Worktrees are actual filesystem directories; users switch via `cd`, not commands
- **No remote/push/pull**: JuiceFS handles transport; JVS only versions local workspaces

## Document Structure

| Document | Purpose |
|----------|---------|
| `CONSTITUTION.md` | **Read first for any feature proposals** - defines core philosophy, non-goals, and governance rules |
| `00_OVERVIEW.md` | Frozen design decisions and product promises |
| `01_REPO_LAYOUT_SPEC.md` | On-disk structure, worktree discovery, portability classes |
| `02_CLI_SPEC.md` | Command contract, error classes, JSON output requirements |
| `03_WORKTREE_SPEC.md` | Worktree lifecycle, exclusive isolation (shared deferred to v1.x) |
| `04_SNAPSHOT_SCOPE_AND_LINEAGE_SPEC.md` | Snapshot identity, descriptor schema, lineage chain |
| `05_SNAPSHOT_ENGINE_SPEC.md` | Engine selection (juicefs-clone/reflink-copy/copy), READY protocol, payload hash |
| `06_RESTORE_SPEC.md` | Restore command with detached state model |
| `07_LOCKING_AND_CONSISTENCY_SPEC.md` | SWMR, fencing tokens, clock skew handling |
| `14_TRACEABILITY_MATRIX.md` | Maps product promises to normative specs to conformance tests |

## Key Design Principles

From CONSTITUTION.md - these are **immutable** without a major version RFC:

1. **Snapshot First, Not Diff First**: No staging area, no patch/diff store, no blob graph
2. **Filesystem as Source of Truth**: No virtualization, no shadow working trees
3. **`.jvs/` MUST NEVER be in snapshot payload**: Payload roots must contain zero control-plane artifacts
4. **Detached state model**: `jvs restore <id>` restores inplace and enters detached state; use `jvs worktree fork` to create branches
5. **Exclusive mode only (v0.x)**: `shared` mode deferred to v1.x
6. **Strong verification by default**: Checksum + payload hash (signing deferred to v1.x)

## Hard Non-Goals

Do not propose features for:
- Git compatibility or text merge semantics
- Remote/push/pull/mirror protocols
- Centralized server orchestration (v0.x)
- Object storage management or credential handling
- Diff-first architecture

## Specification Conventions

- **MUST**: Required behavior; conformance tests validate
- **MUST NOT**: Prohibited behavior
- **SHOULD**: Recommended; valid reasons may exist to deviate
- Error classes (e.g., `E_LOCK_CONFLICT`, `E_FENCING_MISMATCH`) are stable and machine-readable

## When Modifying Specifications

1. Check CONSTITUTION.md for alignment with core principles
2. Update 14_TRACEABILITY_MATRIX.md if adding/changing product promises
3. Ensure conformance tests in 11_CONFORMANCE_TEST_PLAN.md cover new requirements
4. Increment version number in affected specs and 00_OVERVIEW.md
5. Update 99_CHANGELOG.md with changes

## Build & Test

| Command | Purpose |
|---------|---------|
| `make build` | Build `bin/jvs` binary |
| `make test` | Unit tests (`internal/` + `pkg/`) |
| `make test-race` | Unit tests with Go race detector |
| `make test-cover` | Unit tests with coverage report; fails if total < 60% |
| `make lint` | golangci-lint |
| `make conformance` | E2E conformance tests (requires built binary) |
| `make fuzz` | Fuzzing tests (10s per target) |
| `make integration` | Build + conformance |
| `make release-gate` | **Full pre-release gate**: test-race, test-cover, lint, build, conformance, fuzz |
| `make clean` | Remove build artifacts and reports |

### Testing Conventions

- **Unit tests** live next to source files (`*_test.go`). Use `testify/require` for setup, `testify/assert` for assertions.
- **Conformance tests** are in `test/conformance/` with build tag `conformance`. They invoke the `bin/jvs` binary as a black-box.
- **Regression tests** are in `test/regression/`. Each test documents the bug, fix date, and scenario.
- **Fuzz tests** are in `test/fuzz/`.
- **Stress tests** are in `test/stress/`.
- All tests MUST pass with `-race` flag. Use `atomic` or `sync` primitives for shared state.
- Error paths MUST be tested, not just happy paths.
- Coverage target: 60% minimum (enforced by `make test-cover`). Current baseline: ~80%.

### Pre-Merge Gate

Before merging any change, run `make release-gate`. This satisfies the release gates defined in `docs/12_RELEASE_POLICY.md`.

## CLI Commands (from 02_CLI_SPEC.md)

Key commands for reference:
- `jvs init <name>` - Create repository with `.jvs/` and `main/`
- `jvs snapshot [note] [--tag <tag>]` - Create snapshot
- `jvs restore <id>` - Restore worktree to snapshot (inplace, may enter detached state)
- `jvs restore HEAD` - Return to latest state
- `jvs worktree fork <name>` - Fork from current position
- `jvs worktree fork <id> <name>` - Fork from snapshot
- `jvs verify [--all]` - Strong verification (checksum + payload hash)
- `jvs doctor --strict` - Validate layout, lineage, runtime state
- `jvs gc plan` / `jvs gc run --plan-id <id>` - Two-phase garbage collection
