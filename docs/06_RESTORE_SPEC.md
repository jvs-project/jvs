# Restore Spec (v7.0)

## Overview

The `restore` command has a single behavior: **inplace restore** of a worktree to a specific snapshot.

After restore, the worktree enters **detached state** (unless restoring to HEAD).

## Command

```
jvs restore <snapshot-id>
jvs restore HEAD
```

### Arguments

- `<snapshot-id>`: The snapshot to restore to. Can be:
  - Full snapshot ID
  - Short ID prefix
  - Tag name
  - Note prefix (fuzzy match)
- `HEAD`: Special keyword to restore to the latest snapshot (exit detached state)

## Behavior

### Default Restore (inplace)

1. Validate snapshot exists and passes integrity check.
2. Atomically replace worktree content with snapshot content.
3. Update worktree's `head_snapshot_id` to the restored snapshot.
4. Worktree is now in **detached state** (unless this is the latest snapshot).

### Restore HEAD

1. Look up the worktree's `latest_snapshot_id`.
2. Perform restore to that snapshot.
3. Worktree is now at **HEAD state**.

## Detached State

A worktree is in **detached state** when `head_snapshot_id != latest_snapshot_id`.

In detached state:
- **Cannot create snapshots** - must use `worktree fork` first
- Can still modify files (but cannot save changes via snapshot)
- Can navigate to other snapshots via `restore`
- Can return to HEAD via `restore HEAD`

## Safety

Restore is **safe by default**:
- No data is lost - all snapshots are preserved
- The lineage chain remains intact
- GC will not delete snapshots in the lineage

## Examples

```bash
# Restore to specific snapshot
jvs restore 1771589366482-abc12345

# Restore by tag
jvs restore v1.0

# Return to latest state
jvs restore HEAD

# After restore, create branch if you want to continue working
jvs restore v1.0              # Now in detached state
jvs worktree fork hotfix-123  # Create new worktree from here
```

## Error Handling

| Error | Cause | Resolution |
|-------|-------|------------|
| Snapshot not found | Invalid ID or tag | Use `history` to find valid IDs |
| Cannot snapshot in detached state | Attempted `snapshot` while detached | Use `worktree fork` or `restore HEAD` |

## Migration from v6.x

In v6.x, `restore` had two modes:
- Default: created new worktree (`SafeRestore`)
- `--inplace --force --reason`: overwrote current worktree

In v7.0:
- `restore` always does inplace
- Use `worktree fork` to create new worktree from snapshot
- No more `--inplace`, `--force`, `--reason` flags
