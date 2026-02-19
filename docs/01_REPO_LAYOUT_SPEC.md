# Repository Layout Spec (v6.1)

## Definitions
- Volume: mounted filesystem (JuiceFS preferred).
- Repository: directory with `.jvs/` and standard JVS layout.
- Worktree: directory containing workspace payload.

## Standard on-disk layout (MUST)
```
repo/
├── .jvs/
│   ├── format_version
│   ├── snapshots/
│   ├── descriptors/
│   ├── refs/
│   ├── locks/
│   ├── intents/        # in-flight snapshot/restore intents
│   ├── audit/          # append-only audit events
│   ├── gc/             # gc plans, pins, execution metadata
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

## Invariants
- `.jvs/` MUST NOT be inside any worktree payload root.
- A worktree payload root MUST NOT contain `.jvs/` or `worktrees/`.
- Worktrees are real directories selected by `cd`.

## Why `repo/main/` exists
JuiceFS `clone` is 1:1 directory clone with no exclude filter.
`repo/main/` keeps a clean payload source so snapshots never include control-plane directories.
