# Conformance Test Plan (v6.5)

## Purpose
Define mandatory spec tests that gate release quality.

## Profiles
- `dev`: core safety/integrity checks
- `release`: full matrix, no downgrade options

## Mandatory tests
1. lock conflict in exclusive mode -> `E_LOCK_CONFLICT`
2. renew failure halts writes
3. fencing mismatch blocks publish
4. quiesced snapshot succeeds with valid lock
5. best_effort snapshot is risk-labeled in history
6. in-place restore without valid lock fails even with `--force`
7. path traversal name rejected (`E_PATH_ESCAPE` or `E_NAME_INVALID`)
8. descriptor checksum tamper detected
9. payload tamper detected (`E_PAYLOAD_HASH_MISMATCH`)
10. crash tmp artifacts hidden and repairable via doctor
11. migration excludes runtime state and destination rebuild succeeds
12. gc run with mismatched plan id fails (`E_GC_PLAN_MISMATCH`)
13. gc fail/retry path preserves consistency
14. audit event exists for all mutating commands
15. audit hash chain validates end-to-end (`E_AUDIT_CHAIN_BROKEN` on tamper)
16. `format_version` mismatch rejected with `E_FORMAT_UNSUPPORTED`
17. worktree rename with active lock fails with `E_LOCK_CONFLICT`
18. snapshot ID format matches `<timestamp_ms>-<random_hex8>` pattern
19. payload root hash is deterministic for identical payload content
20. ref create/delete appends audit event and validates name safety
21. head orphan detected and `advance_head` repair offered by doctor
22. GC respects ref-targeted snapshots as non-deletable
23. lock steal uses atomic rename and concurrent stealer race resolves safely
24. descriptor checksum excludes checksum/integrity_state fields and matches canonical JSON computation
25. worktree discovery resolves correct `config.json` from nested CWD within payload root
26. payload roots contain zero control-plane artifacts after init and worktree create
27. `jvs init` creates `.jvs/worktrees/main/config.json` with valid schema
28. `jvs worktree remove` deletes both payload directory and `.jvs/worktrees/<name>/` metadata
29. migration sync including `.jvs/worktrees/` preserves worktree metadata at destination

## Acceptance
- release profile requires 100% pass
- any failed mandatory test blocks release
- failure output must include machine-readable error class
