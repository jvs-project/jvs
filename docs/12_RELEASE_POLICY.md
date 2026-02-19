# Release Policy (v6.3)

## Versioning
- major: incompatible storage/model semantics
- minor: additive spec capabilities
- patch: clarifications and non-semantic corrections

## Release gates (MUST)
Before release tag:
1. `jvs doctor --strict` passes
2. `jvs verify --all` passes (default strong verification)
3. `jvs conformance run --profile release` passes
4. threat model residual risks reviewed
5. changelog complete and date-ordered

## Downgrade policy
- `--allow-unsigned` is forbidden in release profile.
- Any artifact verified with downgrade mode is non-release grade.

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
