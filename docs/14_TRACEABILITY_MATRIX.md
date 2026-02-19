# Traceability Matrix (v6.4)

This matrix maps product promises to normative specs and conformance tests.

## Promise 1: Safe-by-default restore
- Product statement:
  - `README.md` (Core guarantees)
  - `docs/00_OVERVIEW.md` (Product promise)
- Normative specs:
  - `docs/06_RESTORE_SPEC.md` (default safe restore, danger-mode constraints)
  - `docs/02_CLI_SPEC.md` (`restore` contract and required flags)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` test 6 (in-place restore lock requirement)

## Promise 2: Strong exclusive writer safety
- Product statement:
  - `README.md` (lock + lease + fencing)
  - `docs/00_OVERVIEW.md` (exclusive default)
- Normative specs:
  - `docs/07_LOCKING_AND_CONSISTENCY_SPEC.md` (lock schema, acquire/renew/steal/release)
  - `docs/02_CLI_SPEC.md` (`lock` commands)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` tests 1, 2, 3

## Promise 3: Verifiable tamper-evident history
- Product statement:
  - `README.md` (strong default verification)
  - `docs/00_OVERVIEW.md` (verification model)
- Normative specs:
  - `docs/04_SNAPSHOT_SCOPE_AND_LINEAGE_SPEC.md` (descriptor schema incl. payload hash + signature)
  - `docs/05_SNAPSHOT_ENGINE_SPEC.md` (payload hash generation + READY/durability)
  - `docs/09_SECURITY_MODEL.md` (trust policy and key lifecycle)
  - `docs/02_CLI_SPEC.md` (`verify` default strong mode)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` tests 8, 9, 10, 11

## Promise 4: Explicit risk labeling for degraded modes
- Product statement:
  - `README.md` (shared/best_effort risk labels)
  - `docs/00_OVERVIEW.md` (risk-explicit modes)
- Normative specs:
  - `docs/07_LOCKING_AND_CONSISTENCY_SPEC.md` (`best_effort` risk exposure)
  - `docs/02_CLI_SPEC.md` (`history` risk labels)
  - `docs/03_WORKTREE_SPEC.md` (`shared` high-risk constraints)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` test 5

## Promise 5: Safe migration semantics
- Product statement:
  - `README.md` (exclude runtime state)
  - `docs/00_OVERVIEW.md` (runtime-state non-portable)
- Normative specs:
  - `docs/18_MIGRATION_AND_BACKUP.md` (exclude `locks/intents`, rebuild runtime)
  - `docs/01_REPO_LAYOUT_SPEC.md` (portability classes)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` test 13

## Promise 6: Safe retention and deletion
- Product statement:
  - `docs/00_OVERVIEW.md` (verifiable history, operational safety)
- Normative specs:
  - `docs/08_GC_SPEC.md` (plan/mark/commit protocol, ref protection)
  - `docs/02_CLI_SPEC.md` (`gc plan`, `gc run --plan-id`)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` tests 14, 15, 25

## Promise 7: Auditable operation history with tamper evidence
- Product statement:
  - `docs/00_OVERVIEW.md` (verifiable and tamper-evident history)
- Normative specs:
  - `docs/09_SECURITY_MODEL.md` (audit log format, hash chain, record schema)
  - `docs/02_CLI_SPEC.md` (`doctor` audit chain validation)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` tests 16, 17

## Promise 8: Deterministic snapshot identity and integrity
- Product statement:
  - `docs/00_OVERVIEW.md` (verifiable history)
- Normative specs:
  - `docs/04_SNAPSHOT_SCOPE_AND_LINEAGE_SPEC.md` (snapshot ID generation)
  - `docs/05_SNAPSHOT_ENGINE_SPEC.md` (payload root hash computation)
  - `docs/09_SECURITY_MODEL.md` (supported algorithms, trust bootstrap)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` tests 20, 21, 22

## Promise 9: Pure payload roots with centralized control plane
- Product statement:
  - `docs/CONSTITUTION.md` §2.3 (control-plane/data-plane separation)
  - `docs/CONSTITUTION.md` §4.2 (JuiceFS clone lacks exclude filters)
- Normative specs:
  - `docs/01_REPO_LAYOUT_SPEC.md` (layout invariants, worktree discovery)
  - `docs/03_WORKTREE_SPEC.md` (centralized metadata under `.jvs/worktrees/`)
  - `docs/04_SNAPSHOT_SCOPE_AND_LINEAGE_SPEC.md` (no exclusion logic required)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` tests 28, 29, 30, 31, 32

## Release gating trace
- Normative release policy:
  - `docs/12_RELEASE_POLICY.md`
- Required operational checks:
  - `docs/13_OPERATION_RUNBOOK.md`
- Conformance execution:
  - `docs/11_CONFORMANCE_TEST_PLAN.md`
