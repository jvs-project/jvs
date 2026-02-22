# JVS Project Governance

**Version:** 1.0
**Last Updated:** 2026-02-23
**Status:** Active

---

## Overview

JVS (Juicy Versioned Workspaces) is an open-source project under the MIT License. This document describes how the project is governed, how decisions are made, and how community members can contribute and advance within the project.

---

## Project Roles

### 1. User

**Definition:** Anyone who uses JVS for workspace versioning.

**Responsibilities:** None specific. Users are encouraged to:
- Report bugs via GitHub Issues
- Request features via GitHub Discussions
- Ask questions via GitHub Discussions
- Provide feedback on user experience

**How to Become:** Simply use JVS! No formal process required.

---

### 2. Contributor

**Definition:** Anyone who contributes code, documentation, tests, or reviews to the JVS project.

**Responsibilities:**
- Follow the [CONTRIBUTING.md](CONTRIBUTING.md) guidelines
- Sign off commits using Developer Certificate of Origin (DCO)
- Write tests for new functionality
- Ensure `make verify` passes before submitting PRs
- Respond to review feedback in a timely manner

**How to Become:** Submit a pull request that is merged into the main branch.

**Rights:**
- Vote on non-technical community decisions (via GitHub reactions)
- Participate in GitHub Discussions and Issues
- Submit pull requests

---

### 3. Maintainer

**Definition:** A contributor with write access to the main repository who is responsible for the technical direction and health of the project.

**Current Maintainers:**
- @percy (Project Lead)

**Responsibilities:**
- Review and merge pull requests
- Ensure code quality and test coverage standards
- Release management (version tagging, changelog)
- Triage and prioritize issues
- Enforce the [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- Mentor contributors
- Make technical decisions within the scope of [CONSTITUTION.md](docs/CONSTITUTION.md)

**How to Become:**

New maintainers are added by existing maintainer consensus using the following process:

1. **Demonstrated Contribution:** Candidate has consistently contributed high-quality PRs over at least 3 months
2. **Review Activity:** Candidate has actively reviewed PRs from other contributors
3. **Endorsement:** An existing maintainer nominates the candidate in a GitHub Discussion
4. **Consensus:** All existing maintainers must approve; none object within 7 days
5. **Onboarding:** Candidate is invited, completes onboarding checklist

**Removing Maintainers:**

Maintainers may be removed by consensus of other maintainers for:
- Inactivity (no contributions for 6 months without notice)
- Repeated violations of the Code of Conduct
- Consistently blocking project progress

---

### 4. Emeritus Maintainer

**Definition:** Former maintainers who have stepped down but are recognized for their contributions.

**Rights:**
- Listed in project documentation
- May be consulted for historical context
- Can return to maintainer role through standard process

---

## Decision Making

### Technical Decisions

**Scope:** Code architecture, feature acceptance, API changes, bug triage

**Process:**
1. Proposals via GitHub Issues or PRs
2. Discussion period (minimum 3 days for significant changes)
3. Maintainer consensus required for merging
4. Project Lead makes final decision in case of disagreement

**Constraints:**
- All changes must comply with [CONSTITUTION.md](docs/CONSTITUTION.md)
- Breaking changes require spec document updates
- All code must pass `make verify`

---

### Non-Technical Decisions

**Scope:** Documentation, community processes, event participation

**Process:**
1. Proposal via GitHub Discussion
2. Community input via comments and reactions
3. Maintainer considers feedback and decides

---

### Changes to Governance

This document may be updated by:
1. Proposal via GitHub Discussion
2. Maintainer consensus (2/3 majority)
3. 7-day comment period for community
4. Approval by majority of maintainers

---

## Developer Certificate of Origin (DCO)

JVS requires the Developer Certificate of Origin (DCO) for all contributions.

**Process:**
1. Contributors must add a `Signed-off-by` line to each commit
2. The sign-off must match the author's email
3. Automated checks verify DCO compliance on all PRs

**Example:**
```
feat(snapshot): add support for snapshot tagging

Signed-off-by: Jane Doe <jane@example.com>
```

Git shortcut: `git commit -s` automatically adds the sign-off.

---

## Release Management

### Release Cadence

JVS follows **Semantic Versioning** (SemVer 2.0.0):

- **Major (X.0.0):** Breaking changes, CONSTITUTION amendments
- **Minor (x.Y.0):** New features, backward compatible
- **Patch (x.y.Z):** Bug fixes, backward compatible

### Release Process

1. **Pre-release:**
   - All tests must pass (`make verify`)
   - Conformance tests must pass (`make conformance`)
   - Changelog updated (`docs/99_CHANGELOG.md`)

2. **Release:**
   - Version tag created by maintainer
   - GitHub release published with release notes
   - Announcement in GitHub Discussions

3. **Post-release:**
   - Monitor for bug reports
   - Patch releases as needed

---

## Continuity Plan (Bus Factor)

### Risk Mitigation

JVS currently has a **bus factor of 1**. To ensure project continuity:

#### Access Redundancy

| Asset | Primary | Backup | Recovery Time |
|-------|---------|--------|---------------|
| GitHub Repository | @percy | To be established | TBD |
| Domain/Assets | N/A | N/A | N/A |
| Signing Keys | To be established | To be established | TBD |

#### Knowledge Transfer

**Documentation:**
- All architectural decisions documented in `docs/`
- Specs follow traceability matrix
- Code has comprehensive comments

**Escrow:**
- Maintainer credentials stored in secure location (TODO)
- Emergency contact list (TODO)

#### Succession Protocol

If the current maintainer becomes unavailable:

1. **Week 1-2:** Active contributors assess status
2. **Week 3:** Contributors nominate interim maintainers
3. **Week 4:** GitHub support contacted for repository transfer if needed

---

## Code of Conduct Enforcement

JVS follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).

**Enforcement:**
- Maintainers are responsible for enforcement
- Report violations to: conduct@jvs-project.org (to be configured)
- Or via [GitHub Security Advisory](https://github.com/jvs-project/jvs/security/advisories)

**Sanctions:**
- Warning for first offense
- Temporary suspension for repeated offenses
- Permanent ban for severe violations

---

## Communication Channels

| Channel | Purpose | Access |
|---------|---------|--------|
| GitHub Issues | Bug reports, feature requests | Public |
| GitHub Discussions | Questions, proposals, community | Public |
| GitHub PRs | Code review, contributions | Public |
| Security Advisories | Vulnerability reports | Private draft |

---

## Related Documents

- [CONSTITUTION.md](docs/CONSTITUTION.md) - Core principles and design governance
- [CONTRIBUTING.md](CONTRIBUTING.md) - How to contribute
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) - Community guidelines
- [SECURITY.md](SECURITY.md) - Security policy and vulnerability reporting
- [TEAM_CHARTER.md](docs/TEAM_CHARTER.md) - Team structure (internal)

---

## Acknowledgments

JVS governance is inspired by successful open-source projects including:
- Kubernetes (CNCF graduated)
- Prometheus (CNCF graduated)
- etcd (CNCF graduated)

---

*This governance document is a living document. Changes require maintainer consensus and community review period.*
