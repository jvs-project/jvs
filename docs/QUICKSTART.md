# JVS Quick Start Guide

**Get started with JVS in 5 minutes**

---

## Prerequisites

### Required
- **Go 1.25+** - For building from source
- **JuiceFS** (recommended) - For O(1) snapshot performance

### Optional
- A CoW-capable filesystem (btrfs, XFS) for reflink engine
- Any POSIX filesystem (fallback to copy engine)

---

## Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/jvs-project/jvs.git
cd jvs

# Build the binary
make build

# (Optional) Install to your PATH
sudo cp bin/jvs /usr/local/bin/
```

### Option 2: Using Go Install

```bash
go install github.com/jvs-project/jvs@latest
```

### Verify Installation

```bash
jvs --version
# Output: jvs version 7.0.0
```

---

## 5-Minute Tutorial

### Step 1: Initialize a Repository

Create a new repository with a main workspace:

```bash
# Navigate to where you want your repository
cd /path/to/workspace

# Initialize JVS repository
jvs init myproject

# JVS creates:
# myproject/
# ├── .jvs/          # Metadata (control plane)
# └── main/          # Your workspace (data plane)
```

### Step 2: Enter Your Workspace

```bash
cd myproject/main
```

**Important:** The repository root (`myproject/`) is NOT your workspace. `main/` is your actual working directory.

### Step 3: Create Your First Snapshot

```bash
# Add some files to your workspace
echo "Hello JVS" > README.md
echo "print('hello')" > script.py

# Create a snapshot
jvs snapshot "Initial setup"

# JVS responds with snapshot ID:
# ✓ Snapshot created: 01abcd...
```

### Step 4: Make Changes and Snapshot Again

```bash
# Modify files
echo "Updated content" >> README.md
echo "print('world')" >> script.py

# Create another snapshot
jvs snapshot "Added more content"
# ✓ Snapshot created: 01efgh...
```

### Step 5: View History

```bash
jvs history

# Output:
# SNAPSHOT ID   TIMESTAMP         NOTE                TAGS
# 01efgh...     2026-02-23 12:05  Added more content
# 01abcd...     2026-02-23 12:00  Initial setup
```

### Step 6: Restore to a Previous Snapshot

```bash
# Restore to initial state
jvs restore 01abcd

# JVS modifies main/ in-place to match snapshot
# ✓ Restored to 01abcd... (Initial setup)
# ℹ Worktree is now in detached state
```

**Detached State:** After restoring to an old snapshot, your worktree is "detached" from HEAD. New snapshots will create a new lineage.

### Step 7: Return to Latest State

```bash
jvs restore HEAD
# ✓ Restored to latest snapshot
```

---

## Common Workflows

### Create a Branch

```bash
# Create a new worktree (branch) from current state
jvs worktree fork experiment

# New worktree created at:
# myproject/worktrees/experiment/

# Navigate to your new worktree
cd ../worktrees/experiment
```

### Use Tags for Organization

```bash
# Create snapshot with tags
jvs snapshot "Stable point" --tag stable --tag v1.0

# Restore by tag
jvs restore --latest-tag stable
```

### Verify Integrity

```bash
# Verify all snapshots
jvs verify --all

# ✓ All snapshots verified
```

### Check Repository Health

```bash
jvs doctor --strict

# ✓ Repository is healthy
```

---

## Common Commands Reference

| Command | Description | Example |
|---------|-------------|---------|
| `jvs init <name>` | Create new repository | `jvs init myproject` |
| `jvs snapshot [note]` | Create snapshot | `jvs snapshot "Fixed bug"` |
| `jvs restore <id>` | Restore to snapshot | `jvs restore HEAD` |
| `jvs worktree fork <name>` | Create branch | `jvs worktree fork feature-x` |
| `jvs history` | Show snapshots | `jvs history --tag v1.0` |
| `jvs verify` | Verify integrity | `jvs verify --all` |
| `jvs doctor` | Health check | `jvs doctor --strict` |
| `jvs gc plan` | Preview GC | `jvs gc plan --keep-daily 7` |

---

## Tips and Gotchas

### ✅ Do

- Work in `main/` or `worktrees/<name>/`, not the repository root
- Use descriptive snapshot notes
- Run `jvs doctor --strict` if something seems wrong
- Use tags to mark important snapshots (releases, milestones)

### ❌ Don't

- Don't manually edit `.jvs/` contents
- Don't expect Git-like merge behavior (JVS doesn't merge)
- Don't ignore detached state warnings
- Don't commit `.jvs/` to Git (it's metadata, not payload)

---

## What's Next?

- **Full Documentation:** See [docs/](docs/) for detailed specifications
- **Contributing:** See [CONTRIBUTING.md](CONTRIBUTING.md) to contribute
- **Reporting Issues:** Use [GitHub Issues](https://github.com/jvs-project/jvs/issues)
- **Architecture:** See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for design details

---

## Getting Help

| Resource | Link |
|----------|------|
| Documentation | https://github.com/jvs-project/jvs/tree/main/docs |
| GitHub Issues | https://github.com/jvs-project/jvs/issues |
| GitHub Discussions | https://github.com/jvs-project/jvs/discussions |
| Security Reporting | See [SECURITY.md](SECURITY.md) |

---

*This guide covers JVS basics. For advanced usage, see the full specification documents.*
