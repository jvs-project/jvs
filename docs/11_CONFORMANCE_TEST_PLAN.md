# Conformance Test Plan (v6.4)

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
10. signature tamper detected (`E_SIGNATURE_INVALID`)
11. trust policy violation detected (`E_TRUST_POLICY_VIOLATION`)
12. crash tmp artifacts hidden and repairable via doctor
13. migration excludes runtime state and destination rebuild succeeds
14. gc run with mismatched plan id fails (`E_GC_PLAN_MISMATCH`)
15. gc fail/retry path preserves consistency
16. audit event exists for all mutating commands
17. audit hash chain validates end-to-end (`E_AUDIT_CHAIN_BROKEN` on tamper)
18. `format_version` mismatch rejected with `E_FORMAT_UNSUPPORTED`
19. worktree rename with active lock fails with `E_LOCK_CONFLICT`
20. snapshot ID format matches `<timestamp_ms>-<random_hex8>` pattern
21. trust bootstrap on `jvs init` produces valid keyring and policy
22. payload root hash is deterministic for identical payload content
23. ref create/delete appends audit event and validates name safety
24. head orphan detected and `advance_head` repair offered by doctor
25. GC respects ref-targeted snapshots as non-deletable
26. lock steal uses atomic rename and concurrent stealer race resolves safely
27. descriptor checksum excludes checksum/signature fields and matches canonical JSON computation
28. worktree discovery resolves correct `config.json` from nested CWD within payload root
29. payload roots contain zero control-plane artifacts after init and worktree create
30. `jvs init` creates `.jvs/worktrees/main/config.json` with valid schema
31. `jvs worktree remove` deletes both payload directory and `.jvs/worktrees/<name>/` metadata
32. migration sync including `.jvs/worktrees/` preserves worktree metadata at destination

## Acceptance
- release profile requires 100% pass
- any failed mandatory test blocks release
- failure output must include machine-readable error class
