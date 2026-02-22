# CNCF Sandbox Application Package

**Project:** JVS (Juicy Versioned Workspaces)
**Prepared:** 2026-02-23
**Target:** CNCF Sandbox Application

---

## Executive Summary

JVS is a **snapshot-first, filesystem-native workspace versioning system** built on JuiceFS. We are applying for CNCF Sandbox to join the cloud-native ecosystem and contribute our unique approach to workspace versioning for data-intensive workloads.

**Why CNCF?**
- JVS solves workspace versioning for data science, ML, and agent workflows
- Aligns with CNCF's cloud-native storage philosophy
- Complements existing CNCF projects (JuiceFS, Prometheus, etcd)
- Ready for community collaboration and feedback

---

## Project Pitch

### What is JVS?

JVS (Juicy Versioned Workspaces) is a **workspace-native versioning layer** that provides:

- **O(1) snapshots** via JuiceFS Copy-on-Write
- **Two-layer integrity** (checksum + payload hash)
- **Detached state model** for safe history navigation
- **Filesystem-native UX** (no virtualization)
- **Local-first design** (no remote protocol complexity)

### Why JVS Matters

**Problem:** Traditional version control (Git) doesn't work well for:
- Large datasets (ML models, scientific data)
- Binary-heavy workspaces
- Agent workflows requiring full reproducibility
- Teams using JuiceFS for storage

**Solution:** JVS treats the entire workspace as the version unit, not individual files.

### Use Cases

1. **Data Science / ML Teams** - Version experiment environments with datasets
2. **CI/CD Pipelines** - Reproducible build environments
3. **Agent Workflows** - Deterministic sandbox states for AI agents
4. **Platform Engineering** - Standardized workspace lifecycle on JuiceFS

---

## CNCF Readiness Checklist

### ✅ Required Documents

| Document | Status | Location |
|----------|--------|----------|
| LICENSE | ✅ MIT | `/LICENSE` |
| SECURITY.md | ✅ Complete | `/SECURITY.md` |
| CONTRIBUTING.md | ✅ Complete | `/CONTRIBUTING.md` |
| CODE_OF_CONDUCT.md | ✅ Complete | `/CODE_OF_CONDUCT.md` |
| GOVERNANCE.md | ✅ Complete | `/GOVERNANCE.md` |
| README.md | ✅ Complete | `/README.md` |

### ✅ Best Practices Badge

| Level | Status | Score |
|-------|--------|-------|
| OpenSSF Passing | ✅ **Achieved** | 95% |
| OpenSSF Silver | ⏳ In Progress | 85% |

### ✅ Quality Metrics

| Metric | Value | Target |
|--------|-------|--------|
| Test Coverage | 83.7% | 80% ✅ |
| Conformance Tests | 54+ passing | - |
| Static Analysis | ✅ gosec + staticcheck | - |
| CI/CD | ✅ GitHub Actions | - |

---

## Value to CNCF Community

### Technical Contributions

1. **New Versioning Paradigm** - Snapshot-first approach for data-intensive workloads
2. **Filesystem-Native Design** - Lessons on CoW utilization for versioning
3. **Two-Layer Integrity Model** - Checksum + payload hash pattern
4. **Go Implementation** - Idiomatic Go codebase for reference

### Integration Opportunities

| CNCF Project | Integration Potential |
|--------------|---------------------|
| **JuiceFS** | Primary storage backend |
| **Prometheus** | Metrics export for monitoring |
| **etcd** | Distributed metadata (future) |
| **BuildKit** | Workspace-aware builds |
| **containerd** | Snapshot-aware container images |

### Ecosystem Fit

JVS fills a gap in the CNCF landscape:
- **Before:** Git (code), Docker (containers), Helm (packages)
- **With JVS:** Complete workspace versioning for data-intensive workflows

---

## TOC Sponsor Outreach Template

### Subject: CNCF Sandbox Application - JVS (Juicy Versioned Workspaces)

Dear [TOC Member Name],

I hope this message finds you well. I'm writing to introduce JVS (Juicy Versioned Workspaces), a new open-source project applying for CNCF Sandbox, and to seek your feedback and potential sponsorship.

**What is JVS?**

JVS is a snapshot-first workspace versioning system built on JuiceFS. We provide O(1) snapshots with strong integrity verification for data-intensive workspaces (ML, data science, agent workflows).

**Why CNCF?**

JVS aligns with CNCF's cloud-native philosophy:
- Filesystem-native design (no virtualization)
- Integrates with JuiceFS (CNCF sandbox project)
- Solves real pain points for data science and ML teams
- 83.7% test coverage, OpenSSF Passing badge achieved

**Key Metrics:**
- MIT licensed
- 54+ conformance tests (all passing)
- Go implementation
- Active development with clear governance

**What We're Looking For:**

1. Feedback on our approach to workspace versioning
2. Guidance on CNCF community engagement
3. Potential TOC sponsorship for Sandbox application

**Resources:**
- GitHub: https://github.com/jvs-project/jvs
- Documentation: https://github.com/jvs-project/jvs/tree/main/docs
- CII Assessment: https://bestpractices.coreinfrastructure.org/projects/XXX

Would you be available for a 30-minute call to discuss JVS and get your feedback? We're particularly interested in your perspective on workspace versioning in the cloud-native ecosystem.

Thank you for your time and consideration.

Best regards,
[Your Name]
JVS Project Lead

---

## Presentation Outline (10 minutes)

### Slide 1: Title
- JVS: Workspace Versioning for Data-Intensive Workloads
- Tagline: "Snapshot-first, filesystem-native versioning on JuiceFS"

### Slide 2: The Problem
- Git doesn't work well for large datasets
- ML experiments need full workspace reproducibility
- Agents need deterministic sandbox states
- Current solutions are complex or incomplete

### Slide 3: What is JVS?
```
jvs init myproject
cd myproject/main
jvs snapshot "experiment 1"
jvs restore abc123  # O(1) rollback!
```
- Snapshot-first (not diff-first)
- Filesystem-native (no virtualization)
- O(1) via JuiceFS CoW

### Slide 4: Architecture
```
┌─────────────────┐
│   JVS CLI       │
├─────────────────┤
│   .jvs/         │ ← Control plane
│   metadata      │
├─────────────────┤
│   main/         │ ← Data plane
│   workspace     │
└─────────────────┘
         ↓
    JuiceFS
```

### Slide 5: Integrity Model
- Two-layer verification
- Descriptor checksum (SHA-256)
- Payload root hash (SHA-256)
- Tamper-evident audit trail

### Slide 6: Current Status
- v7.0 released (2026-02-20)
- 83.7% test coverage
- 54+ conformance tests
- OpenSSF Passing badge (95%)
- MIT licensed

### Slide 7: Use Cases
- Data science / ML experiment tracking
- CI/CD environment versioning
- Agent workflow sandboxes
- Platform engineering workspace lifecycle

### Slide 8: Why CNCF?
- Aligns with JuiceFS (CNCF Sandbox)
- Fills gap in versioning landscape
- Ready for community collaboration
- OpenSSF best practices compliant

### Slide 9: Roadmap
- Q2 2026: Silver badge, features
- Q3 2026: Ecosystem integrations
- Q4 2026: Incubator readiness

### Slide 10: Call to Action
- Try JVS: `go install github.com/jvs-project/jvs@latest`
- Read docs: github.com/jvs-project/jvs
- Give feedback!
- **Sponsor our Sandbox application?**

---

## Case Study Template

### Case Study: [Company/Team Name]

**Background:**
- Industry / Domain
- Team size
- Primary challenge

**Problem:**
- What wasn't working before JVS
- Why existing solutions were insufficient

**Solution:**
- How JVS was deployed
- Integration with existing infrastructure
- Migration process

**Results:**
- Quantitative metrics (time saved, storage efficiency, etc.)
- Qualitative improvements (developer experience, reproducibility)

**Quotes:**
- "[Quote from team lead about JVS impact]"

**Technical Details:**
- Storage backend (JuiceFS configuration)
- Workspace size and snapshot frequency
- CI/CD integration approach

---

## Application Timeline

| Milestone | Target Date | Status |
|-----------|-------------|--------|
| Document package complete | 2026-02-23 | ✅ |
| TOC outreach (3-5 sponsors) | Week of 2026-02-24 | ⏳ |
| Feedback incorporation | 2026-03-15 | Planned |
| Application submitted | 2026-04-01 | Planned |
| Sandbox presentation | 2026-05/06 | Planned |

---

## Contact Information

**Project:** JVS (Juicy Versioned Workspaces)
**Website:** https://github.com/jvs-project/jvs
**Email:** [to be configured]
**Lead:** @percy

---

*This document is a living package. Updates will be made as we progress through the application process.*
