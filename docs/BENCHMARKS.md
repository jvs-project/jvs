# JVS Performance Benchmarks

This document contains actual benchmark results for JVS operations, providing performance baselines and helping identify regressions.

## Test Environment

- **CPU**: Intel(R) Core(TM) Ultra 9 285H
- **Date**: 2026-02-23
- **OS**: Linux (Manjaro)
- **Filesystem**: tmpfs (for consistent benchmark results)
- **Go Version**: `go version`

## Running Benchmarks

Run all benchmarks:
```bash
go test -bench=. -benchmem ./...
```

Run specific package benchmarks:
```bash
go test -bench=. -benchmem ./internal/snapshot/
go test -bench=. -benchmem ./internal/restore/
```

Run a specific benchmark:
```bash
go test -bench=BenchmarkSnapshotCreation_CopyEngine_Small -benchmem ./internal/snapshot/
```

## Benchmark Categories

### Snapshot Operations (`internal/snapshot/bench_test.go`)

| Benchmark | ns/op | ops/sec | B/op | allocs/op | Status |
|-----------|--------|---------|------|-----------|--------|
| `BenchmarkSnapshotCreation_CopyEngine_Small` | 5,698,366 | 176 | 2,304,627 | 54,089 | ✅ |
| `BenchmarkSnapshotCreation_CopyEngine_Medium` | 2,599,263 | 385 | 802,500 | 18,048 | ✅ |
| `BenchmarkSnapshotCreation_ReflinkEngine_Small` | 6,867,611 | 146 | 2,744,647 | 64,154 | ✅ |
| `BenchmarkSnapshotCreation_ReflinkEngine_Medium` | 1,827,102 | 547 | 518,871 | 11,144 | ✅ |
| `BenchmarkSnapshotCreation_MultiFile` | 3,344,783 | 299 | 3,987,290 | 11,265 | ✅ |
| `BenchmarkSnapshotCreation_MultiFile_Large` | 29,390,816 | 34 | 36,983,815 | 46,684 | ✅ |
| `BenchmarkDescriptorSerialization` | 774.5 | 1,291,226 | 496 | 2 | ✅ |
| `BenchmarkDescriptorDeserialization` | 2,393 | 417,886 | 760 | 19 | ✅ |
| `BenchmarkLoadDescriptor` | 4,930 | 202,839 | 1,624 | 19 | ✅ |
| `BenchmarkVerifySnapshot_ChecksumOnly` | 11,949 | 83,694 | 4,703 | 78 | ✅ |
| `BenchmarkVerifySnapshot_WithPayloadHash` | 62,479 | 16,003 | 40,331 | 119 | ✅ |
| `BenchmarkComputeDescriptorChecksum` | 6,641 | 150,574 | 4,158 | 79 | ✅ |
| `BenchmarkListAll_Empty` | 1,429 | 699,790 | 264 | 5 | ✅ |
| `BenchmarkListAll_Single` | 7,357 | 135,929 | 2,027 | 29 | ✅ |
| `BenchmarkListAll_Many` | 261,998 | 3,817 | 91,220 | 1,169 | ✅ |
| `BenchmarkFind_ByTag` | 29,213 | 34,235 | 9,625 | 144 | ✅ |
| `BenchmarkFind_ByWorktree` | 55,170 | 18,126 | 18,825 | 250 | ✅ |
| `BenchmarkFindByTag` | 7,140 | 140,056 | 2,076 | 33 | ✅ |

### Restore Operations (`internal/restore/bench_test.go`)

| Benchmark | ns/op | ops/sec | B/op | allocs/op | Status |
|-----------|--------|---------|------|-----------|--------|
| `BenchmarkRestore_CopyEngine_Small` | 5,477,531 | 183 | 2,254,871 | 40,342 | ✅ |
| `BenchmarkRestore_CopyEngine_Medium` | 3,476,003 | 288 | 1,426,606 | 25,467 | ✅ |
| `BenchmarkRestore_ReflinkEngine_Small` | 6,216,151 | 161 | 2,527,065 | 45,224 | ✅ |
| `BenchmarkRestore_ReflinkEngine_Medium` | 3,771,654 | 265 | 1,512,263 | 26,996 | ✅ |
| `BenchmarkRestore_MultiFile` | 2,399,871 | 417 | 712,057 | 12,116 | ✅ |
| `BenchmarkRestore_MultiFile_Large` | 11,867,199 | 84 | 1,424,587 | 19,447 | ✅ |
| `BenchmarkRestoreToLatest` | 5,921,888 | 169 | 2,476,983 | 44,333 | ✅ |
| `BenchmarkRestore_DetachedState` | 7,014,758 | 143 | 2,888,837 | 51,747 | ✅ |
| `BenchmarkRestore_IntegrityVerification` | 6,580,486 | 152 | 2,805,400 | 50,225 | ✅ |
| `BenchmarkRestore_SnapshotToSnapshot` | 7,397,803 | 135 | 3,114,064 | 55,783 | ✅ |
| `BenchmarkRestore_EmptyWorktree` | 7,229,214 | 138 | 3,038,641 | 54,422 | ✅ |

### GC Operations (`internal/gc/bench_test.go`)

| Benchmark | ns/op | ops/sec | B/op | allocs/op | Status |
|-----------|--------|---------|------|-----------|--------|
| `BenchmarkGCPlan_Small` | 107,075 | 9,339 | 31,946 | 388 | ✅ |
| `BenchmarkGCPlan_Medium` | 645,163 | 1,550 | 221,602 | 2,480 | ✅ |
| `BenchmarkGCPlan_Large` | 6,899,356 | 145 | 2,185,981 | 23,211 | ✅ |
| `BenchmarkGCPlan_WithDeletable` | 375,042 | 2,666 | 131,304 | 1,431 | ✅ |
| `BenchmarkGCRun_DeleteSingle` | 7,740,680 | 129 | 3,265,488 | 66,443 | ✅ |
| `BenchmarkGCRun_DeleteMultiple` | 61,918,604 | 16 | 26,045,173 | 546,843 | ✅ |
| `BenchmarkGCLineageTraversal` | 636,721 | 1,571 | 221,617 | 2,480 | ✅ |
| `BenchmarkGCWithPins` | 414,074 | 2,415 | 144,892 | 1,631 | ✅ |
| `BenchmarkGCEmptyRepo` | 25,318 | 39,497 | 5,304 | 64 | ✅ |
| `BenchmarkGCWithIntents` | 399,813 | 2,501 | 125,590 | 1,353 | ✅ |

## Performance Expectations

### Snapshot Creation
- **Small files (<100KB)**: Completes in ~5-7ms (includes setup overhead)
- **Medium files (~1MB)**: Completes in ~1.8-2.6ms
- **Multi-file (100+ files)**: Scales linearly with file count (~3.3ms for 100 files)
- **Multi-file large (1000 files)**: ~29ms with ~37MB allocations
- **Reflink vs Copy**: Performance varies by workload; reflink is faster for medium files, copy is slightly faster for small files on tmpfs

### Restore Operations
- **Small files (<100KB)**: Completes in ~5.5-6.2ms
- **Medium files (~1MB)**: Completes in ~3.5-3.8ms
- **Multi-file large (1000 files)**: ~12ms with ~1.4MB allocations
- **Detached state restore**: ~7ms (includes integrity verification)
- **Empty worktree restore**: ~7.2ms (similar to content replace due to verification overhead)

### Catalog Operations
- **ListAll_Empty**: ~1.4μs (fast path)
- **ListAll_Single**: ~7.4μs
- **ListAll_Many**: ~262μs for 50 snapshots (~5μs per snapshot)
- **Find by tag**: ~7.1μs (optimized index lookup)
- **Find by worktree**: ~55μs (requires filtering)

### Integrity Verification
- **Checksum only**: ~12μs (SHA-256 of descriptor)
- **With payload hash**: ~62μs for small payloads (SHA-256 tree hash)
- **ComputeDescriptorChecksum**: ~6.6μs

## Memory Allocations

Key allocations to track:

- **Large snapshots**: 1000 files ~46K allocations (~37MB)
- **Large restores**: 1000 files ~19K allocations (~1.4MB)
- **ListAll(50)**: ~1.2K allocations (~91KB)

## Performance Regression Detection

When making changes to critical paths:

1. Run benchmarks before and after
2. Compare using `benchstat`:
   ```bash
   go test -bench=. -benchmem ./internal/snapshot/ > old.txt
   # make changes
   go test -bench=. -benchmem ./internal/snapshot/ > new.txt
   benchstat old.txt new.txt
   ```
3. Flag any regressions >10% for review

## Known Performance Characteristics

1. **Engine selection impact** (on tmpfs):
   - For small files: Copy is slightly faster than Reflink (~5.5ms vs ~6.2ms)
   - For medium files: Reflink is faster than Copy (~1.8ms vs ~2.6ms for snapshot)
   - For large operations: Copy is generally faster on tmpfs (no CoW overhead)
   - Real filesystems (ext4/xfs) will favor reflink due to block-level cloning
   - Copy engine is always available as fallback

2. **Memory allocation patterns**:
   - Snapshot creation: ~50K allocations for small files (setup overhead dominates)
   - Restore operations: ~40K-50K allocations per operation
   - Large snapshots (1000 files): ~47K allocations, ~37MB
   - Large restores (1000 files): ~19K allocations, ~1.4MB
   - GC delete operations: Scale heavily with snapshot count (547K allocations for 100 snapshots)

3. **Filesystem matters**: Benchmarks run on tmpfs; results will vary on:
   - ext4/xfs: Reflink performance improves significantly
   - JuiceFS: Network latency dominates; engine selection less critical
   - btrfs/zfs: CoW performance similar to reflink results

4. **Concurrency**: Current implementation is single-threaded; parallel snapshot/restore is a future optimization

5. **GC Performance**:
   - Planning scales linearly with snapshot count (~107μs for 10 snapshots)
   - Deleting snapshots is the most expensive GC operation (~7.7ms for single, ~62ms for 100)
   - Lineage traversal is efficient (~637μs for complex histories)
   - Empty repo operations are very fast (~25μs)

## Additional Benchmark Opportunities

Future benchmark areas to consider:
- [ ] Worktree forking performance
- [ ] Lock acquisition overhead
- [ ] Concurrent operations
- [ ] Large-scale scenarios (10K+ snapshots, 100K+ files)
- [ ] Cross-engine performance comparison under various filesystems
