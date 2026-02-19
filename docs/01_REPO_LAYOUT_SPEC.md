# Repository Layout Spec (v6.2)

## Definitions
- Volume: mounted filesystem (JuiceFS preferred)
- Repository: directory containing `.jvs/` and standard JVS layout
- Worktree: payload directory with `.jvs-worktree/` marker

## Standard on-disk layout (MUST)
```
repo/
├── .jvs/
│   ├── format_version
│   ├── snapshots/
│   ├── descriptors/
│   ├── refs/
│   ├── locks/          # runtime state; not migrated as-is
│   ├── intents/        # in-flight operations; not migrated as-is
│   ├── audit/          # append-only audit events
│   ├── trust/          # keyring, trust policy, signature metadata
│   ├── gc/             # retention policy, pin sets, gc plans/results
│   └── index.sqlite    # optional, rebuildable
│
├── main/
│   ├── .jvs-worktree/
│   └── <workspace payload...>
│
└── worktrees/
    └── <name>/
        ├── .jvs-worktree/
        └── <workspace payload...>
```

## Invariants (MUST)
- `.jvs/` MUST NOT exist under any payload root.
- Payload roots MUST NOT contain `worktrees/`.
- Worktree roots MUST resolve to canonical paths under repo root.
- All control-plane paths MUST reject symlink traversal outside repo root.

## Portability classes
- Portable history state: `snapshots/`, `descriptors/`, `refs/`, `audit/`, `trust/`, `gc/`.
- Rebuildable cache state: `index.sqlite`.
- Runtime state (non-portable): `locks/`, active `intents/`.

## Why `repo/main/` exists
JuiceFS clone performs 1:1 directory clone without excludes.
Separating `main/` from `.jvs/` guarantees clean payload snapshot scope.
