# Operation Runbook (v6.2)

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
2. apply recommended repair actions
3. rerun verify

## Migration runbook
1. freeze writers
2. doctor + verify pass on source
3. sync excluding `.jvs/locks/**` and `.jvs/intents/**`
4. run `jvs doctor --strict --repair-runtime` on destination
5. run `jvs verify --all` and recovery drill

## GC runbook
1. run `jvs gc plan` and review `plan_id`
2. execute `jvs gc run --plan-id <id>`
3. if failure, inspect failed tombstones and retry safely
4. verify lineage/head integrity after gc batch
