# User Scenarios and Behavior Patterns

This document captures typical user scenarios and expected behaviors for JVS (Juicy Versioned Workspaces).

## Core Concepts

### Worktree States

| State | Description | Can Snapshot? |
|-------|-------------|---------------|
| **EMPTY** | Newly created worktree, no snapshots yet | No (nothing to snapshot) |
| **HEAD** | At the latest snapshot of the lineage | Yes |
| **DETACHED** | At a historical snapshot | No (must fork first) |

### State Transitions

```
EMPTY ──[snapshot]──► HEAD ◄──[restore HEAD]──► DETACHED
                          │                         │
                          │     [restore <id>]      │
                          └────────────────────────►│
                          │                         │
                          │    [worktree fork]      │
                          │◄────────────────────────┤
                          │                         │
                          │   [worktree fork]       │
                          └──────► NEW HEAD ◄───────┘
```

---

## Scenario 1: Basic Workspace Versioning

**User Goal**: Save checkpoints while working on a project.

```bash
# Initialize repository
$ cd /projects
$ jvs init myproject
$ cd myproject/main

# Work on project...
$ echo "version 1" > file.txt

# Create first snapshot
$ jvs snapshot "initial version"
Created snapshot 1771589366482-abc12345

# Continue working...
$ echo "version 2" > file.txt

# Create another snapshot
$ jvs snapshot "updated content"
Created snapshot 1771589366483-def78901

# View history
$ jvs history
1771589  2026-02-21 10:30  updated content     [HEAD]
1771588  2026-02-21 10:25  initial version
◄── you are here (HEAD)
```

**Key Behaviors**:
- Each snapshot automatically becomes the new HEAD
- User can always create new snapshots (in HEAD state)
- Files in worktree are always "live" - what you see is what you have

---

## Scenario 2: Exploring History (Time Travel)

**User Goal**: Look at how the project looked at a previous point in time.

```bash
# Current state: at HEAD
$ jvs history
1771589  2026-02-21 10:30  release v2     [HEAD]
1771588  2026-02-21 10:25  release v1
1771587  2026-02-21 10:20  initial
◄── you are here (HEAD)

# Restore to historical snapshot
$ jvs restore 1771587
Restored to snapshot 1771587-xyz78901
Worktree is now in DETACHED state.

$ cat file.txt
initial content  # Files now show historical state

# History shows we're at a historical point
$ jvs history
1771589  2026-02-21 10:30  release v2     [HEAD]
1771588  2026-02-21 10:25  release v1
1771587  2026-02-21 10:20  initial
◄── you are here (detached)

# Just looking around, now want to go back to latest
$ jvs restore HEAD
Restored to latest snapshot 1771589
Worktree is back at HEAD state.
```

**Key Behaviors**:
- `restore <id>` always does inplace restore (no separate "safe restore")
- After restore, worktree is in DETACHED state
- `restore HEAD` brings back to the latest state
- No data loss - all snapshots in the lineage are preserved

---

## Scenario 3: Creating a Branch from History

**User Goal**: Found a bug introduced after a certain snapshot, want to create a fix branch from that point.

```bash
# Restore to the known-good point
$ jvs restore 1771587
Restored to snapshot 1771587
Worktree is now in DETACHED state.

# Verify this is the right starting point
$ cat file.txt
known good content

# Try to create snapshot - NOT ALLOWED in detached state
$ jvs snapshot "bugfix attempt"
Error: cannot create snapshot in detached state

You are currently at snapshot '1771587' (historical).
To continue working from this point:

    jvs worktree fork bugfix-branch

Or return to the latest state:

    jvs restore HEAD

# Create a new worktree from current position
$ jvs worktree fork bugfix-branch
Created worktree 'bugfix-branch' from snapshot 1771587
Worktree is at HEAD state - you can now create snapshots.

# Switch to the new branch
$ cd ../worktrees/bugfix-branch

# Now can make changes and snapshot
$ echo "bugfix applied" > file.txt
$ jvs snapshot "fixed the bug"
Created snapshot 1771590-aaa11111
```

**Key Behaviors**:
- Cannot create snapshots in detached state (prevents history corruption)
- Must use `worktree fork` to create a new branch
- Fork from current position by omitting snapshot ID

---

## Scenario 4: Fork from Any Snapshot

**User Goal**: Create an experimental branch from any historical point.

```bash
# Fork from specific snapshot (even while at HEAD)
$ jvs worktree fork 1771588 experiment-v1
Created worktree 'experiment-v1' from snapshot 1771588

# Or fork from current position
$ jvs restore 1771587
$ jvs worktree fork experiment-v2
Created worktree 'experiment-v2' from snapshot 1771587

# List all worktrees
$ jvs worktree list
main              /repo/main              HEAD at 1771589
experiment-v1     /repo/worktrees/exp-1   HEAD at 1771588
experiment-v2     /repo/worktrees/exp-2   HEAD at 1771587
```

**Key Behaviors**:
- `worktree fork <id> <name>` - fork from specific snapshot
- `worktree fork <name>` - fork from current position (convenient shorthand)
- New worktree is always at HEAD state (can snapshot immediately)

---

## Scenario 5: Parallel Development

**User Goal**: Work on multiple features in parallel without interference.

```bash
# Create feature branches from main
$ jvs worktree fork feature-auth
Created worktree 'feature-auth'

$ jvs worktree fork feature-ui
Created worktree 'feature-ui'

# Work on auth feature
$ cd /repo/worktrees/feature-auth
$ echo "auth implementation" > auth.py
$ jvs snapshot "auth module complete"
Created snapshot 1771590-aaa11111

# Work on UI feature (independent)
$ cd /repo/worktrees/feature-ui
$ echo "ui implementation" > ui.py
$ jvs snapshot "ui module complete"
Created snapshot 1771591-bbb22222

# Both features have independent lineages
# main worktree unchanged
$ cd /repo/main
$ jvs history
# Only shows main's history, not feature branches
```

**Key Behaviors**:
- Each worktree has its own independent snapshot lineage
- No "merging" needed - worktrees are isolated
- JuiceFS handles storage efficiency (CoW)

---

## Scenario 6: Recovering from Mistakes

**User Goal**: Made a mistake, want to go back to a known-good state.

```bash
# Current state with unwanted changes
$ cat file.txt
terrible mistake

# View history to find good state
$ jvs history
1771589  2026-02-21 10:30  bad changes       [HEAD]
1771588  2026-02-21 10:25  good state
◄── you are here (HEAD)

# Restore to good state
$ jvs restore 1771588
Restored to snapshot 1771588
Worktree is now in DETACHED state.

$ cat file.txt
good content here  # Back to good state

# Option A: Discard the bad snapshot, continue from here
$ jvs worktree fork main-v2
# ... continue in new worktree ...

# Option B: Go back to HEAD and try again
$ jvs restore HEAD
# Back at bad state, but can fix and create new snapshot
```

**Key Behaviors**:
- Restoring doesn't delete any snapshots
- User can always explore and return to any state
- "Bad" snapshots can be cleaned up later via GC

---

## Scenario 7: Using Tags for Releases

**User Goal**: Mark important snapshots with tags for easy reference.

```bash
# Create snapshot with tags
$ jvs snapshot "release 1.0" --tag v1.0 --tag release --tag stable
Created snapshot 1771589-abc12345

# Create more snapshots
$ jvs snapshot "release 1.1" --tag v1.1 --tag release
Created snapshot 1771590-def78901

# Find by tag
$ jvs history --tag release
1771590  2026-02-21 10:30  release 1.1  [v1.1, release]
1771589  2026-02-21 10:25  release 1.0  [v1.0, release, stable]

# Restore by tag (using fuzzy match)
$ jvs restore v1.0
Restored to snapshot 1771589 (v1.0)
Worktree is now in DETACHED state.
```

**Key Behaviors**:
- Tags are metadata on snapshots
- Multiple tags per snapshot allowed
- Fuzzy match by tag or note prefix

---

## Command Reference Summary

| Command | Description | State Change |
|---------|-------------|--------------|
| `jvs snapshot [note]` | Create snapshot | HEAD → HEAD (new head) |
| `jvs restore <id>` | Restore to snapshot | Any → DETACHED |
| `jvs restore HEAD` | Restore to latest | DETACHED → HEAD |
| `jvs worktree fork [name]` | Fork from current | (creates new HEAD) |
| `jvs worktree fork <id> [name]` | Fork from snapshot | (creates new HEAD) |
| `jvs history` | Show snapshot history | (no change) |

---

## Error Messages and Guidance

### Snapshot in Detached State

```
$ jvs snapshot "my changes"
Error: cannot create snapshot in detached state

You are currently at snapshot '1771587' (historical).
To continue working from this point:

    jvs worktree fork <name>        # Create new worktree from here
    jvs restore HEAD                # Return to latest state
```

### Restore Non-existent Snapshot

```
$ jvs restore nonexistent
Error: snapshot not found: nonexistent

Use 'jvs history' to see available snapshots.
```

### Fork with Existing Name

```
$ jvs worktree fork existing-name
Error: worktree 'existing-name' already exists

Use 'jvs worktree list' to see existing worktrees.
```

---

## Design Principles

1. **One Command, One Action**: Each command does exactly one thing. No mode flags.

2. **Safe by Default**: `restore` doesn't destroy data - it just moves a pointer.

3. **Explicit Over Implicit**: User must explicitly `fork` to create branches.

4. **Clear State Indication**: `history` always shows current position.

5. **No Surprise Data Loss**: All snapshots are preserved until explicit GC.
