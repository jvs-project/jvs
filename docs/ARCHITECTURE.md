# JVS Architecture

**Version:** v7.0
**Last Updated:** 2026-02-23
**Status:** Active

---

## Overview

JVS (Juicy Versioned Workspaces) is a **snapshot-first, filesystem-native versioning layer** built on JuiceFS. It provides O(1) workspace snapshots through Copy-on-Write (CoW) while maintaining a clear separation between control plane metadata and user payload data.

### Key Design Principles

1. **Control/Data Plane Separation** - Metadata lives in `.jvs/`; payload is pure data
2. **Filesystem as Source of Truth** - No virtualization, real directories
3. **Snapshot-First** - Complete workspace states, not diffs
4. **Local-First** - No remote protocol; JuiceFS handles transport

---

## System Components

```
┌─────────────────────────────────────────────────────────────────┐
│                         JVS CLI Layer                           │
├─────────────────────────────────────────────────────────────────┤
│  Commands: init, snapshot, restore, worktree, verify, doctor, gc │
└────────────────────┬────────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────────┐
│                      Internal Packages                          │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │   snapshot  │  │   restore   │  │  worktree   │             │
│  │   creator   │  │  restorer   │  │   manager   │             │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘             │
│         │                │                │                     │
│  ┌──────▼────────────────▼────────────────▼──────┐             │
│  │              repo (Repository)                 │             │
│  │         - Descriptor management                │             │
│  │         - Lineage tracking                     │             │
│  │         - Worktree registry                    │             │
│  └──────┬─────────────────────────────────────────┘             │
│         │                                                       │
│  ┌──────▼────────────┐  ┌──────────────────────────────────┐  │
│  │     engine        │  │          integrity               │  │
│  │  (abstraction)    │  │    - Checksum verification       │  │
│  │  - juicefs-clone  │  │    - Payload hash computation    │  │
│  │  - reflink        │  │    - Two-layer integrity         │  │
│  │  - copy           │  └──────────────────────────────────┘  │
│  └───────────────────┘                                       │
├─────────────────────────────────────────────────────────────────┤
│                    Supporting Packages                          │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐               │
│  │   doctor   │  │     gc     │  │   verify   │               │
│  │  - health  │  │  - collector│  │  - checker │               │
│  │  - repair  │  │  - planner │  │            │               │
│  └────────────┘  └────────────┘  └────────────┘               │
└─────────────────────────────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────────┐
│                    Storage Layer                                │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐         ┌─────────────────┐               │
│  │    Control      │         │     Data        │               │
│  │     Plane       │         │     Plane       │               │
│  │     .jvs/       │         │   main/         │               │
│  │                 │         │   worktrees/    │               │
│  │  - descriptors  │         │                 │               │
│  │  - snapshots    │         │   User Payload  │               │
│  │  - worktrees/   │         │                 │               │
│  │  - audit/       │         │                 │               │
│  │  - gc/          │         │                 │               │
│  └─────────────────┘         └─────────────────┘               │
└─────────────────────────────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────────┐
│                 Filesystem (JuiceFS or any FS)                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Component Responsibilities

### CLI Layer (`internal/cli/`)

**Purpose:** User-facing command interface

**Responsibilities:**
- Parse command-line arguments and flags
- Validate user input
- Format output (text, JSON)
- Return appropriate error classes
- Progress reporting for long operations

**Key Commands:**
- `jvs init` - Repository initialization
- `jvs snapshot` - Create workspace snapshot
- `jvs restore` - Restore to previous snapshot
- `jvs worktree` - Worktree management
- `jvs verify` - Integrity verification
- `jvs doctor` - Health checks and repair
- `jvs gc` - Garbage collection

---

### Snapshot Creator (`internal/snapshot/`)

**Purpose:** Create snapshots with proper integrity guarantees

**Responsibilities:**
- Execute 12-step atomic publish protocol
- Compute descriptor checksum
- Compute payload root hash
- Generate snapshot ID
- Append audit record
- Handle `.READY` file as publish gate

**Protocol Flow:**
1. Validate worktree state
2. Generate snapshot UUID
3. Create intent record
4. Trigger engine snapshot
5. Compute payload hash
6. Build descriptor
7. Write descriptor with checksum
8. Verify integrity
9. Write `.READY` file (atomic publish)
10. Update lineage (HEAD pointer)
11. Append audit record
12. Cleanup intent

---

### Restore Restorer (`internal/restore/`)

**Purpose:** Restore worktree to a previous snapshot state

**Responsibilities:**
- Validate snapshot exists and is verified
- Handle in-place restore (modifies worktree directly)
- Enter detached state when restoring to non-HEAD
- Support fuzzy snapshot lookup (ID, tag, note)
- Preserve worktree configuration

**Detached State Model:**
- `jvs restore <id>` puts worktree in detached state
- `jvs restore HEAD` returns to latest state
- `jvs worktree fork` creates new branch from current state

---

### Worktree Manager (`internal/worktree/`)

**Purpose:** Manage worktree lifecycle and discovery

**Responsibilities:**
- Create and delete worktrees
- Track worktree configuration in `.jvs/worktrees/<name>/config.json`
- Discover worktree from current directory
- Validate worktree invariants
- Handle worktree renaming

**Worktree Discovery:**
1. Walk up from CWD to find `.jvs/` (repo root)
2. Compute relative path within repo
3. Map to worktree name: `main/...` → `main`, `worktrees/<name>/...` → `<name>`
4. Load worktree configuration

---

### Repository Package (`internal/repo/`)

**Purpose:** Core repository operations and metadata management

**Responsibilities:**
- Descriptor CRUD operations
- Lineage tracking (parent-child relationships)
- Worktree registry management
- HEAD pointer management
- Format version validation

---

### Engine Abstraction (`internal/engine/`)

**Purpose:** Abstract snapshot engines for different filesystem capabilities

**Interface:**
```go
type Engine interface {
    Snapshot(src, dst string) error
    IsAvailable() bool
    Name() string
}
```

**Implementations:**

| Engine | Description | Performance | Requirement |
|--------|-------------|-------------|-------------|
| `juicefs-clone` | JuiceFS clone ioctl | O(1) | JuiceFS mounted |
| `reflink` | Copy-on-write via reflink | O(1) | CoW filesystem (btrfs, xfs) |
| `copy` | Recursive file copy | O(n) | Fallback |

---

### Integrity Package (`internal/integrity/`)

**Purpose:** Two-layer integrity verification

**Components:**

1. **Descriptor Checksum**
   - SHA-256 hash of entire descriptor JSON
   - Detects descriptor corruption/tampering
   - Stored in descriptor file

2. **Payload Root Hash**
   - SHA-256 hash of complete payload tree
   - Detects payload corruption/tampering
   - Independent of descriptor (cross-layer validation)

**Verification Modes:**
- `jvs verify` - Strong verification (checksum + payload hash)
- `jvs verify --all` - Verify all snapshots

---

### Doctor (`internal/doctor/`)

**Purpose:** Repository health checks and repair

**Checks:**
- Layout validation (format version, directory structure)
- Lineage integrity (parent-child consistency)
- Descriptor checksums
- Runtime state (orphan intents)
- Audit chain integrity

**Repair Actions:**
- `clean_tmp` - Remove orphan `.tmp` files
- `advance_head` - Advance HEAD to latest READY snapshot
- `rebuild_index` - Regenerate `index.sqlite`
- `audit_repair` - Recompute audit hash chain

---

### Garbage Collection (`internal/gc/`)

**Purpose:** Reclaim storage from unreferenced snapshots

**Process:**
1. `jvs gc plan` - Preview what would be deleted
2. `jvs gc run --plan-id <id>` - Execute plan

**Protection Rules:**
- HEAD snapshot always protected
- Tagged snapshots protected (default policy)
- Explicit pins override default protection
- Minimum retention period

**Two-Phase Protocol:**
- Phase 1: Plan generation (plan ID)
- Phase 2: Execution with confirmation

---

### Audit Package (`internal/audit/`)

**Purpose:** Tamper-evident operation history

**Audit Record Schema:**
```json
{
  "event_id": "uuid",
  "timestamp": "ISO-8601",
  "operation": "snapshot|restore|gc_run|...",
  "actor": "user@host",
  "target": "snapshot-id",
  "reason": "explanation",
  "prev_hash": "SHA-256(previous_record)",
  "record_hash": "SHA-256(this_record)"
}
```

**Hash Chain:** Each record hashes the previous, creating a tamper-evident chain.

---

## Data Flows

### Snapshot Creation Flow

```
User: jvs snapshot "fixed bug"
         │
         ▼
┌─────────────────┐
│  CLI (snapshot) │
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│ Snapshot Creator    │◄─────┐
└────────┬────────────┘      │
         │                   │
         ├──► Validate       │
         │                   │
         ├──► Generate UUID  │
         │                   │
         ├──► Create Intent  │
         │                   │
         ▼                   │
┌─────────────────┐          │
│     Engine      │          │
│  (juicefs-clone)│          │
└────────┬────────┘          │
         │                   │
         ▼                   │
┌─────────────────┐          │
│  Compute Hash   │          │
└────────┬────────┘          │
         │                   │
         ▼                   │
┌─────────────────┐          │
│ Build Descriptor│          │
└────────┬────────┘          │
         │                   │
         ▼                   │
┌─────────────────┐          │
│   Write +       │          │
│  Verify Checksum│          │
└────────┬────────┘          │
         │                   │
         ▼                   │
┌─────────────────┐          │
│ Write .READY    │          │
│  (Atomic Publish)│          │
└────────┬────────┘          │
         │                   │
         ├──► Update Lineage │
         │                   │
         ▼                   │
┌─────────────────┐          │
│  Append Audit   │──────────┘
└─────────────────┘
```

### Restore Flow

```
User: jvs restore abc123
         │
         ▼
┌─────────────────┐
│  CLI (restore)  │
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│ Restore Restorer    │
└────────┬────────────┘
         │
         ├──► Lookup snapshot (fuzzy match)
         │
         ├──► Verify integrity
         │
         ├──► Clear worktree (preserving .jvs/)
         │
         ├──► Engine: copy snapshot to worktree
         │
         ├──► Update worktree state
         │
         └──► Enter detached state (if not HEAD)
```

---

## Trust Boundaries

```
┌─────────────────────────────────────────────────────────────┐
│                      TRUSTED BOUNDARY                       │
│                    (User's Machine)                         │
│                                                             │
│  ┌──────────────┐         ┌──────────────┐                 │
│  │    JVS CLI   │────────▶│  .jvs/ Dir   │                 │
│  └──────────────┘         │  (Metadata)  │                 │
│                           └──────────────┘                 │
│                                                                 │
│  ┌──────────────────────────────────────────────┐             │
│  │            Payload Directories               │             │
│  │      (User-modified, untrusted content)      │             │
│  │  main/  worktrees/<name>/                   │             │
│  └──────────────────────────────────────────────┘             │
└─────────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   FILESYSTEM BOUNDARY                        │
│              (JuiceFS, NFS, local FS, etc.)                  │
└─────────────────────────────────────────────────────────────┘
```

**Trust Assumptions:**
- User's machine is trusted for JVS operations
- Payload content is user-controlled and may be untrusted
- Filesystem permissions provide access control
- No in-JVS authentication/authorization (delegated to OS/JuiceFS)

---

## Extension Points

### Adding a New Snapshot Engine

1. Implement `Engine` interface in `internal/engine/`
2. Register in engine factory
3. Add auto-detection logic
4. Update `05_SNAPSHOT_ENGINE_SPEC.md`

### Adding a New Command

1. Create command handler in `internal/cli/`
2. Define error classes in `pkg/errclass/`
3. Add conformance test in `test/conformance/`
4. Update `02_CLI_SPEC.md`

### Adding Audit Event Types

1. Define new event type in `pkg/model/audit.go`
2. Append record in operation
3. Update `09_SECURITY_MODEL.md`
4. Add conformance test for audit trail

---

## Performance Characteristics

| Operation | Complexity (juicefs-clone) | Complexity (copy) |
|-----------|---------------------------|-------------------|
| Snapshot  | O(1)                      | O(n)              |
| Restore   | O(1)                      | O(n)              |
| Verify    | O(n)                      | O(n)              |
| GC Plan   | O(m) where m = snapshots  | O(m)              |

Where `n` is payload size and `m` is number of snapshots.

---

## Related Documents

- [CONSTITUTION.md](CONSTITUTION.md) - Core principles and non-goals
- [00_OVERVIEW.md](00_OVERVIEW.md) - Frozen design decisions
- [01_REPO_LAYOUT_SPEC.md](01_REPO_LAYOUT_SPEC.md) - On-disk structure
- [02_CLI_SPEC.md](02_CLI_SPEC.md) - Command contract
- [05_SNAPSHOT_ENGINE_SPEC.md](05_SNAPSHOT_ENGINE_SPEC.md) - Engine details
- [09_SECURITY_MODEL.md](09_SECURITY_MODEL.md) - Integrity and audit
- [10_THREAT_MODEL.md](10_THREAT_MODEL.md) - Threat analysis

---

*This architecture document covers the high-level design of JVS. For implementation details, see the Go package documentation and code comments.*
