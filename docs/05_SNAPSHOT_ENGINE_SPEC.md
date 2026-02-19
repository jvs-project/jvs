# Snapshot Engine Spec (v6.1)

JVS exposes one snapshot command while supporting multiple engines.

## Engines
### 1) `juicefs-clone` (preferred)
```bash
juicefs clone <SRC_WORKTREE> <DST_SNAPSHOT> [-p]
```
- O(1) metadata clone for directories.
- No exclude filter support, so source MUST be a clean payload root.

### 2) `reflink-copy` (fallback)
- Uses file-level reflink where supported.
- Recursive walk creates reflinked files and directories.

### 3) `copy` (fallback everywhere)
- Full recursive copy.

## Engine selection (MUST)
Default order:
1. JuiceFS mount + `juicefs` CLI available -> `juicefs-clone`
2. Reflink probe success -> `reflink-copy`
3. Else -> `copy`

Override:
- `JVS_SNAPSHOT_ENGINE=juicefs-clone|reflink-copy|copy`

## Metadata and file semantics (MUST)
For each engine, implementation MUST explicitly handle and report behavior for:
- symlinks
- hardlinks
- file permissions and timestamps
- xattrs
- ACLs

If an engine cannot preserve required metadata, it MUST either:
- fail with a clear error, or
- mark descriptor with explicit degraded preservation fields.
Silent degradation is forbidden.

## Snapshot atomic publish protocol (MUST)
1. Verify preconditions:
   - source worktree resolved
   - lock and fencing token valid when required
   - consistency level accepted (`quiesced` default)
2. Write intent record to `.jvs/intents/snapshot-<id>.json`.
3. Materialize payload into `.jvs/snapshots/<id>.tmp/`.
4. Build descriptor tmp `.jvs/descriptors/<id>.json.tmp` including integrity metadata.
5. Write `.READY` in tmp snapshot with descriptor checksum.
6. Atomically rename snapshot tmp -> `.jvs/snapshots/<id>/`.
7. Atomically rename descriptor tmp -> `.jvs/descriptors/<id>.json`.
8. Update worktree `head_snapshot` last.
9. Mark intent completed and append audit event.

Readers MUST never observe published descriptor without READY snapshot.

## Descriptor integrity (MUST)
Descriptor MUST include:
- `snapshot_id`
- `worktree_id`
- `parent_snapshot_id`
- `created_at`
- `note` (optional)
- `engine`
- `consistency_level`
- `fencing_token` (nullable only when lockless mode is valid)
- `descriptor_checksum` with `algo` and `value`

`descriptor_checksum` is mandatory in v6.1.

## READY marker
Path: `.jvs/snapshots/<id>/.READY`
Required contents:
- snapshot id
- created_at
- engine
- descriptor checksum (`algo:value`)

## Crash recovery
- Unfinished `*.tmp` or intent records are non-visible.
- `jvs doctor --strict` MUST detect:
  - orphan tmp snapshots
  - orphan tmp descriptors
  - incomplete intents
- Recovery actions MUST be explicit (`report`, `clean`, `rebuild-index`).
