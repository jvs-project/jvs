# Snapshot Engine Spec (v6.2)

JVS provides one snapshot command with pluggable engines.

## Engines
### `juicefs-clone` (preferred)
```bash
juicefs clone <SRC_WORKTREE> <DST_SNAPSHOT> [-p]
```

### `reflink-copy` (fallback)
Recursive file walk with reflink where supported.

### `copy` (fallback everywhere)
Recursive deep copy.

## Engine selection (MUST)
1. JuiceFS mount + `juicefs` CLI -> `juicefs-clone`
2. reflink probe success -> `reflink-copy`
3. fallback -> `copy`

Override: `JVS_SNAPSHOT_ENGINE=juicefs-clone|reflink-copy|copy`

## Metadata behavior declaration (MUST)
Implementation MUST define behavior for:
- symlinks
- hardlinks
- mode/owner/timestamps
- xattrs
- ACLs

If preservation is degraded, command MUST fail or write explicit degraded fields. Silent downgrade is forbidden.

## Atomic publish and durability protocol (MUST)
1. Verify preconditions (source, lock/fencing, consistency policy).
2. Create intent `.jvs/intents/snapshot-<id>.json`; fsync intent file and parent dir.
3. Materialize payload into `.jvs/snapshots/<id>.tmp/`.
4. Compute `payload_root_hash` over the materialized tmp payload.
5. Fsync all new files and directories in snapshot tmp tree.
6. Build descriptor tmp `.jvs/descriptors/<id>.json.tmp` with:
   - `descriptor_checksum`
   - `payload_root_hash`
   - signature metadata
7. Fsync descriptor tmp file.
8. Write `.READY` in snapshot tmp with descriptor checksum and signing key id; fsync.
9. Rename snapshot tmp -> `.jvs/snapshots/<id>/`; fsync snapshots parent dir.
10. Rename descriptor tmp -> `.jvs/descriptors/<id>.json`; fsync descriptors parent dir.
11. Update `head_snapshot` last; fsync `.jvs-worktree/` metadata dir.
12. Mark intent completed; append audit event.

Success return is allowed only after steps 1-12 complete.

## Integrity and verification model (MUST)
Descriptor MUST include:
- `descriptor_checksum`
- `payload_root_hash`
- `signature`
- `signing_key_id`
- `signed_at`

`jvs verify` defaults to checksum + payload hash + signature/trust validation.
`--allow-unsigned` is explicit downgrade mode and MUST be warning-labeled.

## READY marker
Path: `.jvs/snapshots/<id>/.READY`
Required contents:
- snapshot id
- created_at
- engine
- descriptor checksum
- payload root hash
- signing key id

## Crash recovery
- orphan `*.tmp` and incomplete intents are non-visible.
- `jvs doctor --strict` MUST classify repair actions:
  - `clean_tmp`
  - `rebuild_index`
  - `audit_repair`
