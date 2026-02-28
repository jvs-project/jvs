# Release Policy (v7.0)

## Versioning
- major: incompatible storage/model semantics
- minor: additive spec capabilities
- patch: clarifications and non-semantic corrections

## Release gates (MUST)
Before release tag:
1. `make release-gate` passes (runs test-race, test-cover, lint, build, conformance, fuzz)
2. `jvs doctor --strict` passes on a representative repository
3. `jvs verify --all` passes (checksum + payload hash)
4. threat model residual risks reviewed
5. changelog complete and date-ordered

## Downgrade policy
- v0.x does not include signature verification. When signing is added in v1.x, `--allow-unsigned` downgrade will be forbidden in release profile.

## Breaking change process
- document rationale
- update affected specs and CLI contracts
- add/adjust conformance tests
- describe migration impact and recovery path

## Required release artifacts
- updated spec set
- conformance summary
- runbook references
- known limitations and risk labels
