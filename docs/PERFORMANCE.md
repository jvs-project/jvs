# JVS Performance Tuning Guide

**Version:** v7.0
**Last Updated:** 2026-02-23

---

## Overview

JVS performance depends on several factors: storage engine, filesystem choices, hardware, and configuration. This guide helps you optimize JVS for your workload.

---

## Performance Characteristics

### Snapshot Performance by Engine

| Engine | Complexity | Use Case | Throughput |
|--------|------------|----------|------------|
| **juicefs-clone** | O(1) | JuiceFS mount | Unlimited |
| **reflink** | O(1) | CoW filesystems (btrfs, XFS) | High |
| **copy** | O(n) | Fallback | Limited by disk I/O |

### Verify Performance

| Operation | Complexity | Bottleneck |
|-----------|------------|------------|
| `jvs verify` | O(n) | Disk I/O, SHA-256 computation |
| `jvs gc plan` | O(m) | Descriptor reads (m = snapshots) |
| `jvs restore` | O(1)* | JuiceFS clone; O(n) for copy |

*With juicefs-clone engine

---

## Engine Selection

### Recommendation: Use juicefs-clone

**Best for:** Production workloads with large datasets

```bash
# Verify juicefs-clone is available
jvs doctor --json | grep -A5 engines

# Force juicefs-clone if detected
jvs init myrepo --engine juicefs-clone
```

**When juicefs-clone is not available:**
1. **Use reflink** on btrfs/XFS for O(1) without JuiceFS
2. **Use copy** as fallback for one-time operations

### reflink Engine

**Best for:** Local SSDs without JuiceFS

```bash
# Check if filesystem supports reflink
# On Linux: stat -c %i file1 stat -c %i file2 (same i-node = CoW)

jvs init myrepo --engine reflink
```

**Requirements:**
- btrfs with CoW enabled
- XFS with reflink enabled
- ZFS with clone support

### copy Engine

**Use for:**
- Testing
- Small workspaces (< 1GB)
- One-time migrations
- Fallback when other engines unavailable

---

## Filesystem Recommendations

### JuiceFS (Recommended)

**Optimization tips:**

1. **Use SSD cache** for JuiceFS metadata:
   ```bash
   juicefs mount redis://... /mnt/jfs \
     --cache-dir /var/lib/juicefs/cache \
     --free-space-ratio 0.1
   ```

2. **Adjust block size** for large files:
   ```bash
   juicefs format ... --block-size 4MiB
   ```

3. **Enable compression:**
   ```bash
   juicefs mount ... --compress lzo
   ```

### btrfs

**Optimization tips:**

1. **Enable CoW:** `chattr +C +T /path/to/dir`
2. **Disable cow** for databases: `chattr +C /var/lib/mysql`
3. **Schedule defrag:** `btrfs filesystem defrag -r /path`

### XFS

**Optimization tips:**

1. **Use reflink=1** mount option
2. **Large inode size:** `mkfs.xfs -i size=512`
3. **Noatime mount option** to reduce metadata

### ZFS

**Optimization tips:**

1. **Enable compression:** `compression=lz4`
2. **Atime:** `atime=off` or `relatime`
3. **Record size:** `recordsize=1M` for large files

---

## Hardware Recommendations

### Storage

**Minimum requirements:**
- **IOPS:** 1000+ for acceptable performance
- **Throughput:** 500 MB/s sequential
- **Latency:** < 10ms average

**Recommended:**
- **NVMe SSD** for best performance
- **SATA SSD** for good cost/performance
- **HDD** only for archival/backup

### Network (for JuiceFS)

**For local network:**
- **10 GbE** or better
- **Low latency** (< 1ms)

**For cloud:**
- **Same region** as Redis/object storage
- **VPC/peering** for connectivity

### RAM

**Minimum:** 4 GB
**Recommended:** 8-16 GB for concurrent operations

JVS itself uses minimal RAM, but Go runtime benefits from more memory.

---

## Configuration Tuning

### Engine Selection

**Auto-detection (default):**
```yaml
# .jvs/config.yaml
engine: auto  # Tries juicefs-clone, reflink, copy
```

**Force specific engine:**
```yaml
engine: juicefs-clone  # Always use JuiceFS
```

### Logging

**Reduce logging overhead:**
```yaml
logging:
  level: warn  # Only log warnings and errors
  format: text  # Faster than JSON
```

### Progress Reporting

**Disable for scripts:**
```bash
jvs snapshot "Automated snapshot" --quiet
```

---

## JuiceFS-Specific Tuning

### Client Configuration

**Optimal for JVS:**
```bash
juicefs mount redis://... /mnt/jfs \
  --cache-size 100GiB \
  --max-cached-inodes 10000000 \
  --buffer-size 10MiB \
  --upload-limit 10MiB
```

### Backend Configuration

**Redis backend:**
```
maxmemory 16gb
maxmemory-policy allkeys-lru
```

**S3 backend:**
- Use same region as compute
- Enable S3 Transfer Acceleration

---

## Operating System Tuning

### Filesystem Cache

**Increase page cache (Linux):**
```bash
# Add to /etc/sysctl.conf
vm.vfs_cache_pressure=50
vm.dirty_ratio=15
vm.dirty_background_ratio=5
```

### File Descriptors

**Increase limits (Linux):**
```bash
# Add to /etc/security/limits.conf
* soft nofile 65536
* hard nofile 65536
```

### I/O Scheduler

**For SSDs:**
```bash
# Use deadline or noop scheduler
echo deadline | sudo tee /sys/block/sdX/queue/scheduler
```

**For HDDs:**
```bash
# Use cfq or deadline
echo cfq | sudo tee /sys/block/sdX/queue/scheduler
```

---

## Common Bottlenecks

### Issue: Slow first snapshot

**Cause:** Computing initial payload hashes for all files

**Solutions:**
- Use JuiceFS with metadata cache
- Pre-warm cache: `find . -type f -exec cat {} \; > /dev/null`
- Accept slower first snapshot for long-term benefit

### Issue: Slow verify

**Cause:** Hashing every file in workspace

**Solutions:**
- Verify specific snapshots instead of `--all`
- Use `--no-payload` to skip hash computation (not recommended for production)
- Run during off-peak hours

### Issue: Slow GC plan

**Cause:** Reading many descriptor files

**Solutions:**
- Keep descriptor count manageable (GC regularly)
- Use retention policies to limit snapshots

---

## Benchmarking

### Measuring Performance

**Snapshot performance:**
```bash
time jvs snapshot "Benchmark test"
```

**Restore performance:**
```bash
time jvs restore <snapshot-id>
```

**Verify performance:**
```bash
time jvs verify --all
```

**Identify bottlenecks:**
```bash
# Check what operations are slow
jvs doctor --json | jq '.timing'
```

### Example Benchmarks

| Workspace Size | Snapshot (juicefs-clone) | Snapshot (copy) | Verify |
|----------------|------------------------|---------------|--------|
| 1 GB | 0.1s | 2s | 3s |
| 10 GB | 0.1s | 25s | 35s |
| 100 GB | 0.1s | 250s | 380s |
| 1 TB | 0.1s | 2500s | 3800s |

*Benchmarks on NVMe SSD with JuiceFS backend*

---

## Optimization Checklist

### Before Deploying to Production

- [ ] Use JuiceFS with juicefs-clone engine
- [ ] Run on SSD or fast network storage
- [ ] Configure appropriate retention policies
- [ ] Test with actual workload size
- [ ] Run `jvs doctor --strict` to verify health
- [ ] Verify IOPS and throughput meet requirements
- [ ] Schedule GC during off-peak hours
- [ ] Enable JuiceFS client caching
- [ ] Configure appropriate logging level

### Monitoring

**Key metrics to monitor:**
- Snapshot creation time
- Restore time
- Verify time
- Disk usage (`.jvs/` and workspace)
- CPU usage during hash computation
- Network bandwidth (for JuiceFS)

---

## Scaling Considerations

### Snapshot Count

**Recommended:** < 10,000 snapshots per repository for best performance

**More snapshots?**
- Use tags to mark important snapshots
- More aggressive GC policies
- Consider splitting into multiple repositories

### Concurrent Operations

**JVS v7.0:** Single writer model

**For concurrent access:**
- Use external coordination (locks, queues)
- Separate repositories per team/member
- Single entry point for operations

### Workspace Size

**Recommended:** Up to 10 TB per workspace (practical limit)

**Larger workspaces:**
- Ensure sufficient IOPS
- Monitor hash computation time
- Consider partial snapshots (future feature)

---

## Troubleshooting Performance

### Issue: Snapshots are slow

**Diagnostics:**
```bash
# Check which engine is being used
jvs doctor --json | grep engine

# Check I/O
iostat -x 1 5

# Check JuiceFS cache
juicefs stats /mnt/jfs
```

**Solutions:**
- Switch to juicefs-clone engine
- Enable JuiceFS caching
- Check disk health

### Issue: Restores are slow

**Diagnostics:**
```bash
# Check if restore is using copy engine
jvs doctor --json | grep engine
```

**Solutions:**
- Ensure juicefs-clone is being used
- Check network bandwidth (for JuiceFS)

### Issue: Verify is slow

**This is expected** - verify reads and hashes every file. Optimization:

1. Verify fewer snapshots:
   ```bash
   jvs verify --since "2026-02-20"
   ```

2. Run during off-peak hours

---

## Related Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - System design
- [FAQ.md](FAQ.md) - Common questions
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Performance issues

---

*For specific performance issues, please open a GitHub Issue with diagnostic information.*
