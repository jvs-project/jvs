# JVS Project Roadmap

**Version:** 1.0
**Last Updated:** 2026-02-23
**Current Release:** v7.0

---

## Vision

JVS aims to become the most reliable, feature-rich workspace version control system for:
- Data science and ML teams with large datasets
- Developer environments with reproducibility requirements
- Agent workflows requiring deterministic workspace states

---

## Current Status (v7.0)

**Released:** 2026-02-20

### Capabilities
- ✅ Snapshot-first workspace versioning
- ✅ O(1) snapshots via JuiceFS CoW (juicefs-clone engine)
- ✅ Fallback engines (reflink, copy)
- ✅ In-place restore with detached state
- ✅ Worktree forking
- ✅ Tag-based snapshot organization
- ✅ Strong verification (checksum + payload hash)
- ✅ Tamper-evident audit trail
- ✅ Garbage collection with plan-preview
- ✅ Health checks via `jvs doctor`
- ✅ 29 conformance tests (all passing)
- ✅ 77.6% test coverage

### Documentation
- ✅ 12 specification documents
- ✅ SECURITY.md, CONTRIBUTING.md, CODE_OF_CONDUCT.md
- ✅ Traceability matrix
- ✅ Operation runbook

---

## 12-Month Roadmap

### Q1 2026 (February - April): Stability & CNCF Readiness

**Focus:** Production hardening, CNCF Sandbox application

| Item | Status | Priority | Owner |
|------|--------|----------|-------|
| Increase test coverage to 80%+ | In Progress | P0 | Programmers |
| Add static analysis (gosec) to CI | Pending | P0 | Programmers |
| Add dynamic analysis (fuzzing) | Pending | P1 | Programmers |
| Implement signed releases | Pending | P0 | Programmers |
| Performance benchmarks | Pending | P1 | Programmer-1 |
| CNCF Sandbox application | Pending | P0 | Product Manager |
| GOVERNANCE.md | ✅ Done | P0 | Product Manager |
| ROADMAP.md | ✅ Done | P0 | Product Manager |
| Quick Start guide | Pending | P1 | Product Manager |

**Target Q1 Release:** v7.1 (bug fix release)

---

### Q2 2026 (May - July): Features & Usability

**Focus:** User experience improvements, v8.0 planning

#### v7.2 - Usability Release
| Feature | Description | Priority |
|---------|-------------|----------|
| Interactive restore | Fuzzy matching with confirmation prompts | P1 |
| Snapshot diff | Show what changed between snapshots | P1 |
| Progress bars | Visual progress for long operations | P2 |
| Config file | `.jvs/config.yaml` for user preferences | P2 |
| Shell completion | Bash/zsh completion scripts | P2 |

#### v8.0 Planning
- Gather user feedback from v7.x
- Define v8.0 feature candidates
- Update CONSTITUTION if needed

**Target Q2 Release:** v7.2

---

### Q3 2026 (August - October): Integration & Ecosystem

**Focus:** Tool integrations, ecosystem growth

#### Potential Integrations
| Tool | Integration | Complexity |
|------|-------------|------------|
| IDEs | VS Code extension for workspace switching | Medium |
| CI/CD | GitHub Action for workspace snapshots | Low |
| Monitoring | Prometheus metrics export | Medium |
| Storage | Alternative filesystem support (NFS, Ceph) | High |

#### Documentation Goals
- Video tutorials
- Case studies from early adopters
- API documentation for library usage

**Target Q3 Release:** v7.3 (or v8.0 if major features ready)

---

### Q4 2026 (November - January): Maturity & Graduation Prep

**Focus:** CNCF Incubator readiness (if Sandbox accepted)

#### Requirements for Incubator
- [ ] Multiple organizations using in production
- [ ] Bus factor ≥ 2
- [ ] 90%+ test coverage
- [ ] Security audit completed
- [ ] Vibrant contributor community
- [ ] Defined release cadence

**Target Q4 Release:** v8.0 (if ready) or v7.4

---

## v8.0 Feature Candidates

**Note:** These are candidates, not commitments. Priority based on user feedback.

### High Probability
| Feature | Rationale | Effort |
|---------|-----------|--------|
| Partial snapshot | Snapshot subset of workspace (e.g., `jvs snapshot models/`) | Medium |
| Remote sync | Helper script for syncing `.jvs/` between machines | Low |
| Snapshot templates | Pre-configured snapshot patterns (e.g., "pre-experiment") | Low |

### Medium Probability
| Feature | Rationale | Effort |
|---------|-----------|--------|
| Worktree sharing | Multiple readers per worktree (deferred from v0.x) | High |
| Compression | Compress snapshot descriptors/metadata | Medium |
| Encryption-at-rest | Integrate with filesystem encryption | Medium |

### Low Probability / Post-v1.0
| Feature | Rationale | Effort |
|---------|-----------|--------|
| Content-addressed storage | Optional blob store for deduplication | Very High |
| Signature verification | Ed25519 snapshot signing (deferred from v0.x) | Medium |
| Distributed consensus | Multi-site coordination | Very High |

---

## CNCF Sandbox Timeline

| Milestone | Target | Status |
|-----------|--------|--------|
| OpenSSF Best Practices - Passing | Q1 2026 | In Progress (80%) |
| OpenSSF Best Practices - Silver | Q2 2026 | Planned |
| Sandbox application submitted | Q2 2026 | Planned |
| Sandbox presentation | Q3 2026 | Planned |

---

## Maintenance Policy

### Supported Versions

| Version | Support Status | EOL Date |
|---------|----------------|----------|
| v7.x | ✅ Active | TBD |
| v6.x | ❌ EOL | 2026-02-20 |
| v5.x and earlier | ❌ EOL | 2025-12-31 |

### Patch Policy

- Critical security bugs: Patch release within 7 days
- High severity bugs: Patch release within 14 days
- Normal bugs: Next minor release

---

## Getting Involved

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

**Areas seeking help:**
- Bug fixes (always welcome)
- Test coverage improvements
- Documentation improvements
- Feature proposals via GitHub Discussions

---

## Related Documents

- [CHANGELOG.md](docs/99_CHANGELOG.md) - Detailed version history
- [UPGRADE.md](UPGRADE.md) - Upgrade guide and version compatibility
- [CONSTITUTION.md](docs/CONSTITUTION.md) - Core principles and non-goals
- [GOVERNANCE.md](GOVERNANCE.md) - Project governance and roles
- [CII_BADGE_ASSESSMENT.md](docs/CII_BADGE_ASSESSMENT.md) - Best practices badge progress

---

*This roadmap is a living document. Updates are made at maintainer discretion based on community feedback and project priorities.*
