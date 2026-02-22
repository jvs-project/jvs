# JVS Development Plan v8.0

**Version:** v8.0-draft
**Created:** 2026-02-23
**Status:** Planning

---

## Executive Summary

This document outlines the development plan for JVS v8.0, addressing target user needs while maintaining the KISS principle.

### Key Questions Answered

| Question | Answer |
|----------|--------|
| **Is declarative merge implemented?** | ❌ No. Intentionally NOT implemented - conflicts with snapshot-first design. |
| **Is CLI UX optimal?** | ⚠️ Good, but can be improved with better colors, help text, and progress indicators. |
| **What features are missing?** | Documentation, integration examples, CLI polish - see below. |

---

## Current State Analysis

### What's Implemented (15 CLI Commands)

| Command | Status | User Value |
|---------|--------|------------|
| `jvs init` | ✅ Complete | Repository initialization |
| `jvs snapshot` | ✅ Complete | O(1) snapshots with tags |
| `jvs restore` | ✅ Complete | Inplace restore |
| `jvs worktree` | ✅ Complete | Worktree CRUD (fork, list, remove) |
| `jvs history` | ✅ Complete | Snapshot history |
| `jvs verify` | ✅ Complete | Integrity verification |
| `jvs doctor` | ✅ Complete | Health checks |
| `jvs gc` | ✅ Complete | Garbage collection |
| `jvs diff` | ✅ Complete | Snapshot diff |
| `jvs info` | ✅ Complete | Repository info |
| `jvs config` | ✅ Complete | Configuration management |
| `jvs conformance` | ✅ Complete | Conformance testing |
| `jvs completion` | ✅ Complete | Shell completion |

### What's NOT Implemented (And Why)

| Feature | Status | Reason |
|---------|--------|--------|
| **Merge support** | ❌ Rejected | Conflicts with snapshot-first design. Binary files cannot be meaningfully merged. Use worktree forks instead. |
| **File locking** | ❌ Rejected | Requires distributed coordination. Violates local-first design. |
| **Remote protocol** | ❌ Rejected | JuiceFS handles transport. Use `juicefs sync` instead. |
| **Built-in scheduling** | ❌ Rejected | Use Airflow/Prefect. JVS is a versioning tool, not an orchestrator. |
| **Web UI** | ❌ Rejected | CLI is sufficient for target users. Web UI adds massive complexity. |

---

## v8.0 Development Priorities

### Priority 1: Documentation (Week 1-2)

**Goal:** Make JVS easy to adopt for target users

#### 1.1 Scenario-Specific Quick Start Guides

Create targeted onboarding guides:

| Guide | Target User | Effort |
|-------|-------------|--------|
| `docs/game_dev_quickstart.md` | Game studios (Unity/Unreal) | 1 day |
| `docs/agent_sandbox_quickstart.md` | AI agent platforms | 1 day |
| `docs/etl_pipeline_quickstart.md` | Data engineering teams | 1 day |

#### 1.2 `.jvsignore` Templates

Provide battle-tested ignore patterns:

```
templates/
├── unity.jvsignore
├── unreal.jvsignore
├── python.jvsignore
├── ml-projects.jvsignore
└── etl-data.jvsignore
```

#### 1.3 Integration Examples

Create script examples for common integrations:

```
examples/
├── airflow/
│   └── jvs_operator.py       # Airflow DAG example
├── unity/
│   └── build_snapshot.sh     # Unity build + snapshot
├── agents/
│   └── experiment_runner.py  # Agent experiment loop
└── etl/
    └── pipeline_example.sh   # ETL with JVS checkpoints
```

---

### Priority 2: CLI UX Improvements (Week 3-4)

**Goal:** Make CLI more intuitive and visually appealing

#### 2.1 Colorized Output

Add terminal colors for better readability:

```go
// Success messages: green
// Warning messages: yellow
// Error messages: red
// Snapshot IDs: cyan
// Tags: blue
```

**Implementation:**
- Use `github.com/fatih/color` or similar
- Respect `NO_COLOR` environment variable
- Add `--no-color` flag

#### 2.2 Improved Help Text

Enhance command help with examples:

```bash
# Before
$ jvs snapshot --help
Create a snapshot of the current worktree.

# After
$ jvs snapshot --help
Create a snapshot of the current worktree.

Examples:
  # Basic snapshot with note
  jvs snapshot "Before refactoring"

  # Snapshot with tags
  jvs snapshot "v1.0 release" --tag v1.0 --tag release

  # Partial snapshot (specific directory)
  jvs snapshot "Assets only" --path Assets/

  # Snapshot for ML experiment
  jvs snapshot "Experiment 42: 92% accuracy" --tag exp-42 --tag ml
```

#### 2.3 Progress Indicators

Enhance progress feedback:

- Show snapshot progress with file count and size
- Show restore progress with remaining files
- Show GC progress with space reclaimed

#### 2.4 Better Error Messages

Improve error clarity:

```bash
# Before
Error: snapshot not found

# After
Error: snapshot 'abc12345' not found

Run 'jvs history' to see available snapshots.
Did you mean: abc12346?
```

---

### Priority 3: Test Coverage (Week 5-6)

**Goal:** Achieve 80%+ test coverage (CNCF best practice)

Current coverage: ~70%
Target coverage: 80%+

| Package | Current | Target | Priority |
|---------|---------|--------|----------|
| `internal/engine` | 78.7% | 85% | High |
| `internal/snapshot` | ~70% | 80% | High |
| `internal/restore` | ~70% | 80% | High |
| `internal/worktree` | ~70% | 80% | Medium |
| `pkg/*` | 90%+ | 95% | Low |

---

### Priority 4: Minor Enhancements (Week 7-8)

#### 4.1 Multi-line Snapshot Notes

Allow rich notes in snapshots:

```bash
jvs snapshot <<EOF
ML Experiment: ResNet50 v2
Dataset: ImageNet (subset: 100k images)
Hyperparameters:
  - Learning rate: 0.001
  - Batch size: 256
Result: 92.3% accuracy
EOF
```

#### 4.2 Snapshot Search

Add search capability:

```bash
# Search by note content
jvs history --search "ResNet"

# Search by tag
jvs history --tag ml

# Search by date range
jvs history --since 2026-01-01 --until 2026-02-01
```

#### 4.3 Shell Completion Enhancement

Improve shell completion for:
- Snapshot IDs (from history)
- Tags (from existing snapshots)
- Worktree names

---

## Implementation Timeline

| Week | Focus | Deliverables |
|------|-------|--------------|
| 1-2 | Documentation | Quick start guides, templates, examples |
| 3-4 | CLI UX | Colors, help text, progress, errors |
| 5-6 | Test Coverage | 80%+ coverage |
| 7-8 | Enhancements | Multi-line notes, search, completion |

---

## KISS Principle Guardrails

Before implementing any feature, verify:

1. ✅ **Solves real problem for target users?** (game devs, agents, ETL)
2. ✅ **Cannot be done with shell script?** (if yes, create example instead)
3. ✅ **Complexity is proportional to value?**
4. ✅ **Maintains local-first, snapshot-first design?**
5. ✅ **No breaking changes to existing workflows?**

---

## Success Criteria

### Adoption Metrics

| Metric | Target |
|--------|--------|
| Game studio adoption | 5+ studios by end of 2026 |
| Agent platform adoption | 3+ platforms by end of 2026 |
| ETL pipeline adoption | 10+ data teams by end of 2026 |

### Quality Metrics

| Metric | Target |
|--------|--------|
| Test coverage | 80%+ |
| Documentation coverage | 100% of features have examples |
| CLI help clarity | < 5 seconds to understand any command |
| GitHub issues closed | < 10 open bugs |

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Feature creep | Strict KISS review process |
| Breaking changes | Maintain backward compatibility |
| Documentation lag | Doc updates required with every PR |
| Test coverage drop | CI fails if coverage < 80% |

---

## Appendix: Why No Merge?

### User Question: "Is declarative merge implemented?"

**Answer: No.**

### Why Merge is Not in JVS

1. **Binary files cannot be merged:**
   - A 3D model (.fbx) has no meaningful diff
   - A texture file (.psd) cannot be auto-merged
   - ML datasets are often binary (parquet, npy)

2. **Snapshot-first design:**
   - JVS stores complete workspace states, not diffs
   - There is no "patch store" to merge from
   - Merging would require a diff-first architecture

3. **Correct mental model:**
   - **Git:** branch → edit → merge → commit
   - **JVS:** edit → snapshot → fork if needed

4. **Worktree forks are the answer:**
   ```bash
   # Instead of "merging branches", fork a new worktree
   jvs worktree fork abc12345 feature-x
   cd worktrees/feature-x/main
   # ... work independently ...
   jvs snapshot "Feature X complete"
   ```

### Alternative Solutions

| Need | Solution |
|------|----------|
| Parallel development | Use worktree forks |
| Track experiments | Use tags + notes |
| Share work | Use JuiceFS sync or rsync |
| Coordinate team | Use external tools (Perforce, Slack) |

---

*This plan maintains focus on JVS's core value: O(1) snapshots for large files that Git cannot handle.*
