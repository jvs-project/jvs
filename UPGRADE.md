# Upgrade Guide

**Version:** 7.0
**Last Updated:** 2026-02-23

This guide explains how to upgrade JVS between versions, including backward compatibility considerations and migration paths.

---

## Quick Reference

| From Version | To Version | Compatibility | Action Required |
|--------------|-------------|----------------|-----------------|
| v6.x | v7.0 | **Breaking** | See [v6.x to v7.0](#v6x-to-v70) below |
| v6.7 | v7.0 | **Breaking** | Worktree config migration required |
| v6.6 or earlier | v7.0 | **Breaking** | Multi-step upgrade via v6.7 required |

---

## Version Support Policy

### Supported Versions

| Version | Support Status | EOL Date |
|---------|----------------|----------|
| v7.x | ✅ Active | TBD |
| v6.x | ❌ EOL | 2026-02-20 |
| v5.x and earlier | ❌ EOL | 2025-12-31 |

### Upgrade Recommendation

- **Always review the changelog** (`docs/99_CHANGELOG.md`) before upgrading
- **Test upgrades in a non-production environment first**
- **Take a backup** before upgrading (see [Backup Before Upgrade](#backup-before-upgrade))
- **Run `jvs doctor --strict` after upgrade** to validate repository health

---

## v6.x to v7.0

v7.0 introduced the **detached state model**, which significantly changes how `restore` and worktree management work.

### Breaking Changes

#### 1. Restore Command (Always In-Place)

**v6.x behavior:**
```bash
# v6.x: restore created new worktree by default
jvs restore <snapshot-id>  # Created new worktree

# v6.x: inplace restore required flags
jvs restore --inplace --force --reason "testing" <snapshot-id>
```

**v7.0 behavior:**
```bash
# v7.0: restore ALWAYS operates in-place
jvs restore <snapshot-id>

# v7.0: create new worktree from snapshot
jvs worktree fork <new-name> --from <snapshot-id>
```

#### 2. Detached State

After restoring to a historical snapshot in v7.0, the worktree enters **detached state**:
- Cannot create snapshots while detached
- Use `jvs restore HEAD` to return to latest state
- Use `jvs worktree fork` to create a new worktree from historical snapshot

#### 3. Removed CLI Flags

The following flags are **removed** in v7.0:
- `--inplace` (restore now always in-place)
- `--force` (no longer needed)
- `--reason` (no longer needed)

### Migration Steps

#### Step 1: Backup Your Repository

```bash
# Create a backup of the entire repository
cp -r /path/to/repo /path/to/repo.backup.v6
```

Or for JuiceFS-mounted repositories:
```bash
# Use juicefs sync to create a backup
juicefs sync /path/to/repo/ /backup/path/repo/ --threads 16
```

#### Step 2: Install JVS v7.0

```bash
# Download and install v7.0
# See README.md for installation instructions
jvs --version  # Should report v7.0
```

#### Step 3: Verify Repository Health

```bash
cd /path/to/repo/main
jvs doctor --strict
jvs verify --all
```

#### Step 4: Update Worktree Configs

v7.0 adds `latest_snapshot_id` to worktree configs. Run:

```bash
# Repair runtime state (including worktree configs)
jvs doctor --strict --repair-runtime
```

#### Step 5: Verify Migration

```bash
# Check that worktrees are properly configured
jvs worktree list

# Verify history is intact
jvs history --limit 10

# Verify all snapshots
jvs verify --all
```

### Rollback Procedure

If you need to rollback to v6.x:

```bash
# 1. Stop any operations
# 2. Restore from backup
rm -rf /path/to/repo
cp -r /path/to/repo.backup.v6 /path/to/repo

# 3. Install v6.x binary
# 4. Verify
cd /path/to/repo/main
jvs doctor --strict
```

---

## v6.6 to v6.7

v6.7 removed the entire **lock/lease/fencing subsystem**.

### Breaking Changes

Removed commands:
- `jvs lock acquire`
- `jvs lock release`
- `jvs lock renew`
- `jvs lock status`
- `jvs lock steal`

Removed files:
- `.jvs/locks/` directory

### Migration Steps

No data migration required. The lock system was only checked at snapshot/restore time and did not protect file operations.

After upgrading to v6.7:
1. Lock-related audit events remain in history for reference
2. No action needed for existing repositories
3. Snapshots work without acquiring locks

---

## v6.5 to v6.6

v6.6 removed the **refs subsystem** and added **tags to snapshots**.

### Breaking Changes

Removed commands:
- `jvs ref create`
- `jvs ref delete`
- `jvs ref list`

Removed files:
- `.jvs/refs/` directory

### Migration Steps

Existing refs are **not automatically converted** to tags. To preserve refs as tags:

```bash
# Before upgrading, list your refs
jvs ref list

# For each ref, create a snapshot with a corresponding tag
jvs snapshot "migrate ref" --tag <ref-name>
```

After upgrading, use `--tag` flag with `jvs snapshot`:

```bash
# v6.6: attach tags during snapshot creation
jvs snapshot "important checkpoint" --tag stable --tag v1.0
```

---

## Worktree Metadata Migration (v6.4)

v6.4 moved worktree metadata from `<worktree>/.jvs-worktree/` to `.jvs/worktrees/<name>/`.

### Impact

- Worktree payload roots now contain **zero control-plane artifacts**
- Pure payload structure required for `juicefs clone` compatibility

### Migration Steps

Existing repositories are **automatically migrated** on first run:

```bash
# First run of v6.4+ detects old layout
jvs doctor --strict

# If migration is needed, you'll see:
# "Migrating worktree metadata to new layout..."
```

To verify migration:
```bash
# Check that .jvs/worktrees/ exists
ls -la .jvs/worktrees/

# Check that old .jvs-worktree directories are gone
find . -name ".jvs-worktree" -type d
```

---

## Backup Before Upgrade

### Full Repository Backup

```bash
# Method 1: Direct copy
cp -r /path/to/repo /path/to/repo.backup.$(date +%Y%m%d)

# Method 2: For JuiceFS
juicefs sync /path/to/repo/ /backup/path/ \
  --exclude '.jvs/intents/**' \
  --threads 16
```

### Validate Backup

```bash
# Test the backup
cd /backup/path/main
jvs doctor --strict
jvs verify --all
```

---

## Post-Upgrade Verification

After any upgrade, run the following verification steps:

```bash
# 1. Check version
jvs --version

# 2. Verify repository health
jvs doctor --strict

# 3. Verify all snapshots
jvs verify --all

# 4. Check worktree status
jvs worktree list

# 5. Verify recent snapshots are accessible
jvs history --limit 5

# 6. Test snapshot creation
jvs snapshot "post-upgrade test"

# 7. Test restore (if not in production)
jvs worktree fork test-restore
cd ../test-restore
jvs restore HEAD
```

---

## Troubleshooting

### "Format version unsupported" Error

This occurs when downgrading to a version that doesn't support the repository format.

**Solution:** Re-install the newer version or restore from backup.

### Worktree Config Errors

After upgrade, if worktree configs are invalid:

```bash
# Rebuild runtime state
jvs doctor --strict --repair-runtime
```

### Snapshot Not Found

If snapshots are missing after upgrade:

1. Check the audit log: `cat .jvs/audit/audit.jsonl | grep snapshot`
2. Verify descriptor files exist: `ls .jvs/descriptors/`
3. Run `jvs verify --all` to find corrupted snapshots

### Performance Degradation

If performance is worse after upgrade:

1. Check which engine is being used: `jvs info`
2. Verify JuiceFS is mounted: `df -h | grep juicefs`
3. Run with verbose logging: `jvs --verbose snapshot "test"`

---

## Getting Help

If you encounter issues during upgrade:

1. **Check the changelog:** `docs/99_CHANGELOG.md`
2. **Check the runbook:** `docs/13_OPERATION_RUNBOOK.md`
3. **Search existing issues:** https://github.com/jvs-project/jvs/issues
4. **Open a new issue:** Include `jvs --version` and error output

---

## Related Documents

- [CHANGELOG.md](docs/99_CHANGELOG.md) - Detailed version history
- [MIGRATION_AND_BACKUP.md](docs/18_MIGRATION_AND_BACKUP.md) - Cross-host migration
- [RELEASE_POLICY.md](docs/12_RELEASE_POLICY.md) - Release process and gates
- [ROADMAP.md](ROADMAP.md) - Future version planning
