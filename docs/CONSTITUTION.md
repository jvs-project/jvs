# JVS Constitution
## Juicy Versioned Workspaces — Core Principles, Philosophy, and Scope

Version: 1.2
Status: Foundational  
Scope: Architecture, Product Philosophy, and Design Governance  

---

# 1. Core Mission（核心使命）

JVS is a **workspace-native, snapshot-first versioning system** built on top of a mounted filesystem (preferably JuiceFS), designed to provide:

- Safe, versioned workspaces for humans and agents
- O(1) large-scale snapshots via CoW (graceful O(n) fallback when CoW is unavailable)
- Predictable, filesystem-aligned UX
- Zero coupling to backend storage services

JVS is **not** a Git replacement.  
JVS is a **Workspace Versioning Layer**.

---

# 2. Foundational Philosophy（基础哲学）

## 2.1 Snapshot First, Not Diff First
JVS treats the **entire workspace state** as the primary version unit.

> A snapshot represents a complete, reproducible filesystem state, not a textual delta.

Implications:
- No staging area
- No patch/diff object store
- No blob graph complexity
- No content-addressed DAG requirement (v0.x)

---

## 2.2 Filesystem as Source of Truth（文件系统即权威）
The filesystem is the authoritative state container.

JVS:
- Does not virtualize the workspace
- Does not remap directories dynamically
- Does not maintain shadow working trees

Instead:
> Real directories = Real workspaces = Real state

This ensures:
- Agent determinism
- Toolchain compatibility
- POSIX predictability

---

## 2.3 Control Plane vs Data Plane Separation（控制面与数据面分离）
JVS strictly separates:

| Layer | Responsibility |
|-------|---------------|
Control Plane (`.jvs/`) | All metadata: snapshots, descriptors, worktree config, audit trail |
Data Plane (worktrees) | User workspace payload files only |

Critical rules:
> `.jvs/` MUST NEVER be part of snapshot payload.
> Worktree payload roots MUST contain zero control-plane artifacts.

---

## 2.4 JuiceFS as Infrastructure, Not Coupling
JVS assumes:
- Storage backend is already mounted (JuiceFS or any filesystem)
- Backend lifecycle is managed externally

JVS:
- DOES NOT manage credentials
- DOES NOT configure object storage
- DOES NOT implement remote replication

This enforces:
> Single responsibility: workspace versioning only.

---

# 3. System Scope Definition（系统范围定义）

## 3.1 What JVS IS
JVS is:
- A workspace snapshot manager
- A history lineage tracker
- A reproducibility tool for agents and engineers
- A filesystem-native versioning system
- A large-file friendly alternative to Git-like workflows

---

## 3.2 What JVS IS NOT (Hard Non-Goals)
JVS explicitly rejects:

- Git compatibility layer
- Text merge engine
- Remote/push/pull/mirror protocols
- Centralized server orchestration (v0.x)
- Object storage reimplementation
- Diff-first architecture

These are considered **out-of-scope by design**, not missing features.

---

# 4. Architectural Core Model（核心架构模型）

## 4.1 Volume ≠ Repository
A mounted filesystem (e.g., JuiceFS volume) may contain multiple repositories.

```
Volume (mounted FS)
└── repo/
    ├── .jvs/
    ├── main/
    └── worktrees/
```

---

## 4.2 Main Worktree Design (Critical Rule)
The repository root is NOT the main workspace payload.

Instead:
```
repo/main/ = Main Worktree (payload root)
repo/.jvs/ = Control plane
```

Rationale:
- JuiceFS clone lacks exclude filters — payload roots must be free of control-plane artifacts
- Worktree metadata lives under `.jvs/worktrees/<name>/`, not inside payload
- Enables clean snapshot source with zero exclusion logic

---

## 4.3 Worktree = Real Directory (No Virtual Switching)
JVS does not implement virtual workspace switching.

Users and agents select workspaces via:
```
cd repo/main
cd repo/worktrees/<name>
```

This guarantees:
- Absolute path stability
- No hidden state
- Agent-safe execution environments

---

# 5. Snapshot Model Constitution（快照模型宪章）

## 5.1 Snapshot Scope (Immutable Rule)
A snapshot MUST capture:
> Only the current worktree payload root.

Never:
- Entire repo
- Other worktrees
- `.jvs/` directory (which includes all control-plane state)

---

## 5.2 Snapshot Storage Strategy
All snapshots MUST be stored as full directory trees:
```
.jvs/snapshots/<snapshot-id>/
```

Design Choice:
- Full tree clone (JuiceFS CoW or FS CoW)
- Not object graph
- Not delta storage

Optimized for:
- Large datasets
- Deterministic restore
- Simplicity

---

## 5.3 Immutable Snapshot Principle
Once marked READY:
- Snapshot content MUST be immutable
- Mutation is considered repository corruption

---

# 6. Engine Abstraction Principle（引擎抽象原则）

JVS supports adaptive snapshot engines:

| Engine | Condition |
|--------|----------|
| juicefs-clone | Preferred — O(1) CoW metadata operation |
| reflink-copy | CoW filesystems (ZFS, Btrfs, XFS) — no data duplication |
| copy | Generic fallback — O(n) deep copy |

Key Principle:
> Same UX, different engines.  
> Capability adapts to filesystem, not user commands.

---

# 7. Safety and Determinism Principles（安全与确定性）

## 7.1 Detached State Model
Restore always operates inplace:
- `jvs restore <id>` restores worktree to the specified snapshot
- After restore, worktree enters **detached state** if not at HEAD
- In detached state, cannot create new snapshots (must fork first)

Return to HEAD state:
```
jvs restore HEAD
```

Create branch from historical point:
```
jvs worktree fork <snapshot-id> <new-worktree-name>
```

---

## 7.2 Atomic Visibility (READY Protocol)
A snapshot is only visible when:
```
snapshot/.READY exists
```

Before READY:
- Snapshot is treated as incomplete
- Ignored by history and restore

---

## 7.3 SWMR (Single Writer Multiple Reader)
Default isolation:
- Exclusive worktree write lock
- Unlimited readers
- Agent-safe concurrency model

Exception:
- `shared` mode exists as an explicit, opt-in downgrade with no SWMR guarantee.
- `shared` MUST be risk-labeled and documented as high-risk.

---

## 7.4 Integrity and Verifiability
Snapshot history MUST be verifiable and tamper-evident.

Required properties:
- Each snapshot carries a cryptographic integrity proof (checksum + payload hash)
- History lineage is auditable and append-only
- Tampering is detectable, not merely preventable

This justifies:
- Descriptor signing and trust policy
- Audit trail with integrity chain
- Strong-by-default verification

---

# 8. UX & Mental Model Principles（用户心智模型）

## 8.1 Git-like Familiarity, Not Git Complexity
JVS borrows:
- History concept
- Version lineage

But rejects:
- Index
- Rebase
- Merge conflicts
- Detached HEAD semantics

---

## 8.2 Directory-Oriented Interaction
Primary workflow:
```
cd workspace
modify files
jvs snapshot
```

Not:
```
stage → commit → push
```

---

## 8.3 Agent-First Design
JVS is optimized for:
- Cloud agents (Codex, Claude Code, etc.)
- Sandboxed execution
- Reproducible workspace states
- Absolute path stability

---

# 9. Migration & Portability Principle（迁移原则）

JVS does not implement replication protocols.

Official migration method:
```
juicefs sync <repo> <target>
```

Rationale:
- Reuse mature storage tooling
- Avoid protocol duplication
- Preserve single-dependency architecture

---

# 10. Governance & Evolution Rules（演进治理规则）

Future features MUST NOT:
- Introduce hidden workspace states
- Break filesystem transparency
- Couple to specific storage vendors
- Replace snapshot-first with diff-first architecture
- Introduce mandatory server components

Any violation requires:
> Constitution Amendment (major version RFC)

---

# 11. Target Users（目标用户）

Primary:
- ACD (AI / Code / Data engineers)
- Cloud agent users
- Platform & infra teams
- Large dataset workflows

Secondary:
- JuiceFS users
- CoW filesystem users (ZFS/Btrfs/XFS)
- Reproducible research pipelines

---

# 12. Design Style Declaration（设计风格声明）

JVS follows:
- Minimalism over feature richness
- Determinism over abstraction
- Filesystem realism over virtual layers
- Infrastructure reuse over reinvention
- Explicit behavior over hidden magic

---

# 13. Final Architectural Motto

> Real directories.  
> Real snapshots.  
> Zero magic.  
> Filesystem-native versioning for the AI era.
