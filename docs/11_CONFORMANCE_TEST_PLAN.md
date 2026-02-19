# Conformance Test Plan (v6.2)

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

## Acceptance
- release profile requires 100% pass
- any failed mandatory test blocks release
- failure output must include machine-readable error class
