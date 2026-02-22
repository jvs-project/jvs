# Traceability Matrix (v7.0)

This matrix maps product promises to normative specs and conformance tests.

## Promise 1: Detached state model
- Product statement:
  - `README.md` (Core guarantees)
  - `docs/00_OVERVIEW.md` (Product promise)
- Normative specs:
  - `docs/06_RESTORE_SPEC.md` (inplace restore, detached state, fork command)
  - `docs/02_CLI_SPEC.md` (`restore` and `worktree fork` contract)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` tests 21-24 (detached state, restore HEAD, fork)

## Promise 2: Verifiable tamper-evident history
- Product statement:
  - `README.md` (strong default verification)
  - `docs/00_OVERVIEW.md` (verification model)
- Normative specs:
  - `docs/04_SNAPSHOT_SCOPE_AND_LINEAGE_SPEC.md` (descriptor schema incl. payload hash)
  - `docs/05_SNAPSHOT_ENGINE_SPEC.md` (payload hash generation + READY/durability)
  - `docs/09_SECURITY_MODEL.md` (integrity model and audit)
  - `docs/02_CLI_SPEC.md` (`verify` default strong mode)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` tests 2, 3

## Promise 3: Safe migration semantics
- Product statement:
  - `README.md` (exclude runtime state)
  - `docs/00_OVERVIEW.md` (runtime-state non-portable)
- Normative specs:
  - `docs/18_MIGRATION_AND_BACKUP.md` (exclude `intents`, rebuild runtime)
  - `docs/01_REPO_LAYOUT_SPEC.md` (portability classes)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` tests 5, 20

## Promise 4: Safe retention and deletion
- Product statement:
  - `docs/00_OVERVIEW.md` (verifiable history, operational safety)
- Normative specs:
  - `docs/08_GC_SPEC.md` (plan/mark/commit protocol)
  - `docs/02_CLI_SPEC.md` (`gc plan`, `gc run --plan-id`)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` tests 6, 7

## Promise 5: Auditable operation history with tamper evidence
- Product statement:
  - `docs/00_OVERVIEW.md` (verifiable and tamper-evident history)
- Normative specs:
  - `docs/09_SECURITY_MODEL.md` (audit log format, hash chain, record schema)
  - `docs/02_CLI_SPEC.md` (`doctor` audit chain validation)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` tests 8, 9

## Promise 6: Deterministic snapshot identity and integrity
- Product statement:
  - `docs/00_OVERVIEW.md` (verifiable history)
- Normative specs:
  - `docs/04_SNAPSHOT_SCOPE_AND_LINEAGE_SPEC.md` (snapshot ID generation)
  - `docs/05_SNAPSHOT_ENGINE_SPEC.md` (payload root hash computation)
  - `docs/09_SECURITY_MODEL.md` (integrity hash algorithms)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` tests 11, 12

## Promise 7: Pure payload roots with centralized control plane
- Product statement:
  - `docs/CONSTITUTION.md` §2.3 (control-plane/data-plane separation)
  - `docs/CONSTITUTION.md` §4.2 (JuiceFS clone lacks exclude filters)
- Normative specs:
  - `docs/01_REPO_LAYOUT_SPEC.md` (layout invariants, worktree discovery)
  - `docs/03_WORKTREE_SPEC.md` (centralized metadata under `.jvs/worktrees/`)
  - `docs/04_SNAPSHOT_SCOPE_AND_LINEAGE_SPEC.md` (no exclusion logic required)
- Conformance tests:
  - `docs/11_CONFORMANCE_TEST_PLAN.md` tests 16, 17, 18, 19

## Release gating trace
- Normative release policy:
  - `docs/12_RELEASE_POLICY.md`
- Required operational checks:
  - `docs/13_OPERATION_RUNBOOK.md`
- Conformance execution:
  - `docs/11_CONFORMANCE_TEST_PLAN.md`
