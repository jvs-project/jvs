# JVS Troubleshooting Guide

**Version:** v7.0
**Last Updated:** 2026-02-23

---

## Overview

This guide covers common issues when using JVS and how to resolve them. For issues not covered here, see [CONTRIBUTING.md](../CONTRIBUTING.md) for how to report bugs.

---

## Quick Diagnostics

Before diving into specific issues, run these commands to assess repository health:

```bash
# 1. Check repository health
jvs doctor --strict

# 2. Verify all snapshots
jvs verify --all

# 3. Check current worktree status
jvs status
```

---

## Engine Detection Issues

### Problem: "no suitable engine found"

**Symptom:**
```
Error: no suitable snapshot engine available
```

**Cause:** JVS cannot find a working snapshot engine. This happens when:
- JuiceFS is not mounted (for juicefs-clone engine)
- Filesystem doesn't support reflink (for reflink engine)
- Copy engine fallback is disabled

**Solutions:**

1. **Check if JuiceFS is mounted:**
   ```bash
   mount | grep juicefs
   ```

2. **Test engine explicitly:**
   ```bash
   jvs init test --engine copy
   ```

3. **Check engine availability:**
   ```bash
   jvs doctor --json | grep -A5 "engines"
   ```

4. **Force copy engine:**
   ```bash
   # In .jvs/config.yaml
   engine: copy
   ```

---

## Snapshot Creation Issues

### Problem: "failed to create snapshot: partial snapshot detected"

**Symptom:**
```
Error: partial snapshot detected - .READY file missing
```

**Cause:** A previous snapshot creation was interrupted (crash, power failure).

**Solutions:**

1. **Run doctor repair:**
   ```bash
   jvs doctor --strict --repair-runtime
   ```

2. **Manually clean up:**
   ```bash
   # List orphan intents
   ls -la .jvs/intents/

   # Remove orphan intents (be careful!)
   rm .jvs/intents/*.json
   ```

---

### Problem: "workspace is dirty"

**Symptom:**
```
Error: cannot create snapshot: workspace has uncommitted changes
```

**Note:** This is not a JVS error in v7.0. JVS creates snapshots of whatever state exists.

**Solution:** If you want to clean up before snapshotting:
```bash
# Reset to last snapshot first
jvs restore HEAD

# Then snapshot clean state
jvs snapshot "Clean state"
```

---

## Restore Issues

### Problem: "failed to restore: descriptor not found"

**Symptom:**
```
Error: descriptor not found: abc123
```

**Cause:** Snapshot ID doesn't exist or was garbage collected.

**Solutions:**

1. **List available snapshots:**
   ```bash
   jvs history
   ```

2. **Use fuzzy search:**
   ```bash
   jvs restore abc  # Will search for snapshots starting with "abc"
   ```

3. **Restore by tag:**
   ```bash
   jvs restore --latest-tag stable
   ```

---

### Problem: "worktree is in detached state"

**Symptom:**
```
Warning: worktree is in detached state
Cannot create snapshot while detached
```

**Cause:** You restored to a historical snapshot that is not the latest (HEAD).

**Solutions:**

1. **Return to latest state:**
   ```bash
   jvs restore HEAD
   ```

2. **Create a fork to continue work:**
   ```bash
   jvs worktree fork my-branch
   cd ../worktrees/my-branch
   ```

---

## Verification Issues

### Problem: "checksum verification failed"

**Symptom:**
```
Error: descriptor checksum mismatch for snapshot abc123
```

**Cause:** Descriptor file was modified or corrupted.

**Solutions:**

1. **Check if this is expected:**
   ```bash
   # Was this snapshot recently created?
   jvs history --format json | grep abc123
   ```

2. **Escalate if unexpected:**
   - Check audit log for suspicious activity
   - Preserve evidence for investigation

---

### Problem: "payload hash mismatch"

**Symptom:**
```
Error: payload root hash mismatch for snapshot abc123
```

**Cause:** Payload files were modified after snapshot was created.

**Solutions:**

1. **Identify changed files:**
   ```bash
   # Find files with modification time after snapshot
   find . -newer .jvs/snapshots/abc123 -ls
   ```

2. **Recompute hash to verify:**
   ```bash
   jvs verify abc123 --recompute
   ```

---

## Garbage Collection Issues

### Problem: "GC plan ID not found"

**Symptom:**
```
Error: GC plan not found: plan-123
```

**Cause:** Plan expired or was already executed.

**Solution:**
```bash
# Create a new plan
jvs gc plan --keep-daily 7
```

---

### Problem: "cannot delete protected snapshot"

**Symptom:**
```
Error: cannot delete snapshot: snapshot is protected
```

**Cause:** Snapshot is protected by:
- Being the HEAD snapshot
- Having a protection pin
- Matching retention policy

**Solutions:**

1. **Check protection status:**
   ```bash
   jvs history --format json | jq '.protection'
   ```

2. **Override protection (use carefully):**
   ```bash
   jvs gc plan --keep-daily 7 --allow-protected
   ```

---

## Worktree Issues

### Problem: "worktree not found"

**Symptom:**
```
Error: worktree not found: my-worktree
```

**Cause:** Worktree config is missing or worktree was removed.

**Solutions:**

1. **List available worktrees:**
   ```bash
   jvs worktree list
   ```

2. **Recreate worktree:**
   ```bash
   jvs worktree fork my-worktree --from <snapshot-id>
   ```

---

### Problem: "cannot remove current worktree"

**Symptom:**
```
Error: cannot remove worktree: currently in this worktree
```

**Cause:** You're trying to remove the worktree you're currently in.

**Solution:**
```bash
# Switch to a different worktree first
cd ../main  # or any other worktree

# Then remove
jvs worktree remove my-worktree
```

---

## Permission Issues

### Problem: "permission denied when writing to .jvs/"

**Symptom:**
```
Error: permission denied: .jvs/descriptors/
```

**Cause:** Insufficient permissions on repository directory.

**Solutions:**

1. **Check ownership:**
   ```bash
   ls -la .jvs/
   ```

2. **Fix permissions:**
   ```bash
   # Ensure you own the .jvs/ directory
   sudo chown -R $USER:$USER .jvs/
   chmod 700 .jvs/
   ```

---

### Problem: "cannot write to worktree"

**Symptom:**
```
Error: permission denied when restoring
```

**Cause:** Insufficient permissions on worktree directory.

**Solution:**
```bash
# Ensure you have write access to worktree
chmod u+w /path/to/worktree
```

---

## Integrity Issues

### Problem: "audit chain broken"

**Symptom:**
```
Error: audit chain broken at record xyz
```

**Cause:** Audit record was modified or deleted, breaking the hash chain.

**Solutions:**

1. **Run doctor repair:**
   ```bash
   jvs doctor --strict --repair-runtime
   ```

2. **Investigate cause:**
   ```bash
   # Check audit log
   cat .jvs/audit/audit.jsonl | tail -10
   ```

3. **Escalate if records are missing** - May indicate security incident

---

## Storage Issues

### Problem: "out of space"

**Symptom:**
```
Error: no space left on device
```

**Cause:** Insufficient disk space for snapshots.

**Solutions:**

1. **Check space usage:**
   ```bash
   du -sh .jvs/snapshots/
   du -sh .jvs/descriptors/
   ```

2. **Run garbage collection:**
   ```bash
   jvs gc plan --keep-daily 7
   jvs gc run --plan-id <plan-id>
   ```

3. **Clean up large files:**
   ```bash
   # Find large files in workspace
   find . -type f -size +100M -ls
   ```

---

### Problem: "JuiceFS mount issues"

**Symptom:**
```
Error: juicefs-clone failed: operation not permitted
```

**Cause:** JuiceFS is not mounted or has issues.

**Solutions:**

1. **Check mount status:**
   ```bash
   mount | grep juicefs
   ```

2. **Test JuiceFS directly:**
   ```bash
   # Try a simple clone operation
   juicefs clone src dst
   ```

3. **Remount JuiceFS:**
   ```bash
   juicefs umount /mnt/jfs
   juicefs mount ... /mnt/jfs
   ```

---

## Performance Issues

### Problem: "slow snapshot creation"

**Symptom:** Snapshots take much longer than expected.

**Possible Causes:**

1. **Wrong engine:**
   ```bash
   # Check which engine is being used
   jvs doctor --json | grep engine
   ```

2. **Large number of small files:**
   ```bash
   # Count files
   find . -type f | wc -l
   ```

3. **Disk I/O bottleneck:**
   ```bash
   # Check disk activity
   iostat -x 1 5
   ```

**Solutions:**
- Use JuiceFS with juicefs-clone engine for O(1) snapshots
- Consider fewer, larger snapshots instead of many small ones
- Run during off-peak hours

---

### Problem: "slow verify"

**Symptom:** `jvs verify --all` takes very long.

**Possible Cause:** Computing payload hashes is I/O intensive.

**Solutions:**

1. **Verify specific snapshots:**
   ```bash
   jvs verify abc123 --no-payload  # Skip hash computation
   ```

2. **Run during off-peak hours**

3. **Use verify for recent snapshots only:**
   ```bash
   jvs verify --since "2026-02-20"
   ```

---

## Doctor Issues

### Problem: "doctor reports E_RUNTIME_STATE issues"

**Symptom:**
```
E_RUNTIME_STATE: orphan intent files detected
```

**Cause:** Crash during snapshot creation left temporary files.

**Solution:**
```bash
jvs doctor --strict --repair-runtime
```

---

### Problem: "doctor reports E_INDEX_MISSING"

**Symptom:**
```
E_INDEX_MISSING: index.sqlite not found or corrupted
```

**Cause:** Index is rebuildable but missing.

**Solution:**
```bash
jvs doctor --strict --repair-runtime
```

---

## Getting Help

### If issues persist:

1. **Gather diagnostic information:**
   ```bash
   jvs doctor --strict --json > diagnostics.json
   jvs verify --all --json > verification.json
   ```

2. **Check known issues:**
   - [GitHub Issues](https://github.com/jvs-project/jvs/issues)
   - [FAQ](README.md#faq)

3. **Report a bug:**
   - See [CONTRIBUTING.md](../CONTRIBUTING.md)
   - Include: JVS version, OS, steps to reproduce, diagnostic output

4. **Security issues:**
   - See [SECURITY.md](../SECURITY.md) for vulnerability reporting

---

## Error Codes Reference

| Error Code | Description | Common Fix |
|------------|-------------|-------------|
| `E_NAME_INVALID` | Invalid worktree/snapshot name | Use valid characters `[a-zA-Z0-9._-]+` |
| `E_PATH_ESCAPE` | Path traversal attempt | Use valid path within repository |
| `E_DESCRIPTOR_CORRUPT` | Descriptor checksum failed | Investigate, preserve evidence |
| `E_PAYLOAD_HASH_MISMATCH` | Payload hash mismatch | Identify changed files |
| `E_LINEAGE_BROKEN` | Parent snapshot missing | Check history, rebuild index |
| `E_PARTIAL_SNAPSHOT` | Incomplete snapshot | Run `jvs doctor --repair-runtime` |
| `E_GC_PLAN_MISMATCH` | GC plan ID mismatch | Create new plan |
| `E_FORMAT_UNSUPPORTED` | Format version too old/new | Upgrade JVS |
| `E_AUDIT_CHAIN_BROKEN` | Audit hash chain broken | Run `jvs doctor --repair-runtime` |

---

## Related Documentation

- [README.md](../README.md) - Overview and quickstart
- [13_OPERATION_RUNBOOK.md](13_OPERATION_RUNBOOK.md) - Operations guide
- [UPGRADE.md](../UPGRADE.md) - Upgrade guide
- [SECURITY.md](../SECURITY.md) - Security policy

---

*For issues not covered here, please open a GitHub Issue or contact the maintainers.*
