# Operation Runbook (v6.3)

## Daily checks
1. run `jvs doctor --strict`
2. run `jvs verify --all`
3. inspect lock age and stale-holder alerts

## Incident: lock stuck
1. run `jvs lock status`
2. confirm holder liveness and nonce/session match
3. if expired, apply steal flow
4. confirm fencing token increment and audit event

## Incident: verification failure
1. freeze writes for affected repo
2. run `jvs verify --all --json`
3. classify failure: checksum, payload hash, signature, trust policy
4. escalate tamper events and preserve evidence

## Incident: partial snapshot artifacts
1. run `jvs doctor --strict --json`
2. apply recommended repair actions:
   - `clean_tmp`: remove orphan `.tmp` snapshot/descriptor files
   - `advance_head`: advance head to latest READY snapshot if head is stale
   - `rebuild_index`: regenerate `index.sqlite`
   - `audit_repair`: recompute audit hash chain
3. rerun verify

## Incident: audit chain broken
1. run `jvs doctor --strict --json`, look for `E_AUDIT_CHAIN_BROKEN`
2. apply `audit_repair` to recompute chain hashes from existing records
3. investigate cause (truncation, manual edit, migration error)
4. escalate if records are missing (indicates potential tampering)

## Migration runbook
1. freeze writers
2. doctor + verify pass on source
3. sync excluding `.jvs/locks/**` and `.jvs/intents/**`
4. run `jvs doctor --strict --repair-runtime` on destination, which:
   - `clean_locks`: removes stale lock files from source
   - `clean_intents`: removes abandoned intent files from source
   - `rebuild_index`: regenerates `index.sqlite`
5. run `jvs verify --all` and recovery drill

## GC runbook
1. run `jvs gc plan` and review `plan_id`
2. execute `jvs gc run --plan-id <id>`
3. if failure, inspect failed tombstones and retry safely
4. verify lineage/head integrity after gc batch
