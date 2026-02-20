# Repository Layout Spec (v6.5)

## Definitions
- Volume: mounted filesystem (JuiceFS preferred)
- Repository: directory containing `.jvs/` and standard JVS layout
- Worktree: pure payload directory registered in `.jvs/worktrees/<name>/`

## Standard on-disk layout (MUST)
```
repo/
├── .jvs/
│   ├── format_version
│   ├── worktrees/      # worktree metadata (centralized)
│   │   ├── main/
│   │   │   └── config.json
│   │   └── <name>/
│   │       └── config.json
│   ├── snapshots/
│   ├── descriptors/
│   ├── refs/
│   ├── locks/          # runtime state; not migrated as-is
│   ├── intents/        # in-flight operations; not migrated as-is
│   ├── audit/          # append-only audit events
│   ├── gc/             # retention policy, pin sets, gc plans/results
│   └── index.sqlite    # optional, rebuildable
│
├── main/               # pure payload — zero control-plane artifacts
│   └── <workspace payload...>
│
└── worktrees/
    └── <name>/         # pure payload — zero control-plane artifacts
        └── <workspace payload...>
```

## `format_version` (MUST)
Path: `.jvs/format_version`

Contents: single line with integer format version.
- `jvs init` writes `1`.
- JVS MUST read `format_version` before any operation.
- If `format_version` > supported version, fail with `E_FORMAT_UNSUPPORTED`.
- If `format_version` < current version and migration is available, `jvs doctor --strict` SHOULD report upgrade recommendation.
- Format version increments only on incompatible on-disk layout changes.

## `refs/` — named snapshot references (MUST)
Path: `.jvs/refs/<name>.json`

Refs provide stable, human-readable names for snapshots (e.g., tags, release markers).

### Schema (MUST)
- `ref_name`: matches `[a-zA-Z0-9._-]+`
- `snapshot_id`: target snapshot ID
- `created_at`: ISO 8601 timestamp
- `created_by`: actor identity
- `note`: optional description

### Rules (MUST)
- Ref names follow the same safety rules as worktree names (no separators, `..`, control chars; NFC normalized).
- Refs are immutable once created. To retarget, delete and recreate.
- Deletion appends audit event.
- `jvs verify --all` MUST validate that ref targets exist and are READY.
- Refs are portable history state and MUST be included in migration sync.

### CLI
- `jvs ref create <name> <snapshot-id>` — create named reference.
- `jvs ref list [--json]` — list all refs.
- `jvs ref delete <name>` — remove ref with audit.

## Invariants (MUST)
- `.jvs/` MUST NOT exist under any payload root.
- Payload roots MUST contain zero control-plane artifacts (no hidden metadata directories).
- Payload roots MUST NOT contain `worktrees/`.
- Worktree roots MUST resolve to canonical paths under repo root.
- All control-plane paths MUST reject symlink traversal outside repo root.
- Every worktree payload directory MUST have a corresponding entry in `.jvs/worktrees/<name>/config.json`.

## Portability classes
- Portable history state: `format_version`, `worktrees/`, `snapshots/`, `descriptors/`, `refs/`, `audit/`, `gc/`.
- Rebuildable cache state: `index.sqlite`.
- Runtime state (non-portable): `locks/`, active `intents/`.

## Why `repo/main/` exists
JuiceFS clone performs 1:1 directory clone without excludes.
Separating `main/` from `.jvs/` guarantees clean payload snapshot scope.
Worktree metadata is stored under `.jvs/worktrees/` (not inside payload roots) for the same reason — clone cannot exclude subdirectories, so payload roots must contain zero control-plane artifacts.

## Worktree discovery
JVS locates worktree metadata by:
1. Walking up from CWD to find the repo root (directory containing `.jvs/`).
2. Computing the relative path of CWD within the repo.
3. Mapping to the worktree name: `main/...` → `main`; `worktrees/<name>/...` → `<name>`.
4. Loading `.jvs/worktrees/<name>/config.json`.
