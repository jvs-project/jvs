# Conformance Test Plan (v6.7)

## Purpose
Define mandatory spec tests that gate release quality.

## Profiles
- `dev`: core safety/integrity checks
- `release`: full matrix, no downgrade options

## Mandatory tests
1. path traversal name rejected (`E_PATH_ESCAPE` or `E_NAME_INVALID`)
2. descriptor checksum tamper detected
3. payload tamper detected (`E_PAYLOAD_HASH_MISMATCH`)
4. crash tmp artifacts hidden and repairable via doctor
5. migration excludes runtime state and destination rebuild succeeds
6. gc run with mismatched plan id fails (`E_GC_PLAN_MISMATCH`)
7. gc fail/retry path preserves consistency
8. audit event exists for all mutating commands
9. audit hash chain validates end-to-end (`E_AUDIT_CHAIN_BROKEN` on tamper)
10. `format_version` mismatch rejected with `E_FORMAT_UNSUPPORTED`
11. snapshot ID format matches `<timestamp_ms>-<random_hex8>` pattern
12. payload root hash is deterministic for identical payload content
13. tags are stored in descriptor and validated against `[a-zA-Z0-9._-]+`
14. head orphan detected and `advance_head` repair offered by doctor
15. descriptor checksum excludes checksum/integrity_state fields and matches canonical JSON computation
16. worktree discovery resolves correct `config.json` from nested CWD within payload root
17. payload roots contain zero control-plane artifacts after init and worktree create
18. `jvs init` creates `.jvs/worktrees/main/config.json` with valid schema
19. `jvs worktree remove` deletes both payload directory and `.jvs/worktrees/<name>/` metadata
20. migration sync including `.jvs/worktrees/` preserves worktree metadata at destination
21. in-place restore requires `--force` flag
22. in-place restore requires `--reason` flag

## Acceptance
- release profile requires 100% pass
- any failed mandatory test blocks release
- failure output must include machine-readable error class
