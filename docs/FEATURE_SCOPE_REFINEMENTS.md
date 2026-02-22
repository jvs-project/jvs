# JVS Feature Scope Refinements

**Version:** v7.0
**Last Updated:** 2026-02-23

---

## Overview

Based on target user research (see [TARGET_USERS.md](TARGET_USERS.md)), this document proposes refinements to JVS's feature scope. The goal is to **solve real problems while maintaining the KISS principle**.

---

## Executive Summary

### Key Findings

1. **JVS's core value is clear:** O(1) snapshots for large files that Git cannot handle
2. **Three target scenarios align well:** Game assets, Agent sandboxes, ETL pipelines
3. **Current feature set is solid:** Snapshot, restore, worktrees, tags, verify
4. **Risk of scope creep:** Many proposed features would violate KISS

### Strategic Positioning

> **JVS is a filesystem-level versioning tool, not a comprehensive platform.**
>
> We solve the storage problem, not the orchestration, collaboration, or integration problems.

---

## Current Feature Assessment

### Keep: Essential Features (Core Value Proposition)

| Feature | User Value | KISS Score | Verdict |
|---------|-----------|------------|---------|
| **O(1) snapshots** | Solves large file problem | ✅ Simple | Keep |
| **Snapshot restore** | Reproducibility | ✅ Simple | Keep |
| **Worktrees** | Isolation for parallel work | ✅ Simple | Keep |
| **Tags** | Organization & discovery | ✅ Simple | Keep |
| **Partial snapshots** | Selective versioning | ✅ Simple | Keep |
| **Verify integrity** | Trust & safety | ✅ Simple | Keep |
| **History/inspect** | Audit trail | ✅ Simple | Keep |

### Simplify: Good Features, Could Be Streamlined

| Feature | Current State | Simplification Proposal |
|---------|---------------|------------------------|
| **Config file support** | Full YAML config | Reduce to 3-4 key settings only |
| **GC retention policies** | Complex policy language | Simplify to `--keep-daily N` only |
| **Compression** | Multiple levels | Keep `fast`/`max` only, or remove entirely |
| **Progress bars** | Configurable | Make always-on (remove config) |

### Remove: Low Value, High Complexity

| Feature | Why Remove | Impact |
|---------|------------|--------|
| **Multiple output formats** | JSON useful for scripts, text for humans | Keep both, but remove table format |
| **Engine autodetection heuristics** | Adds complexity, users can specify | Require explicit `--engine` flag |
| **Snapshot templates** | Edge case, low demand | Remove |

---

## Proposed New Features (Prioritized)

### Tier 1: High Value, Low Complexity (Do These)

#### 1. Scenario-Specific Quick Start Guides

**Effort:** 2-3 days documentation work
**Value:** High - reduces onboarding friction for target users

**Deliverables:**
- `docs/game_dev_quickstart.md` - Unity/Unreal workflows
- `docs/agent_sandbox_quickstart.md` - Agent experiment workflows
- `docs/etl_pipeline_quickstart.md` - Data engineering workflows

**Rationale:** Documentation is cheaper than features. Show users how to solve their problems with existing JVS capabilities.

---

#### 2. `.jvsignore` Templates

**Effort:** 1 day to create templates
**Value:** Medium - reduces user confusion

**Deliverables:**
```
# Unity .jvsignore template
Library/
Temp/
obj/
*.userprefs
.vscode/
.idea/

# Unreal .jvsignore template
Binaries/
Build/
Intermediate/
Saved/
.vscode/
DerivedDataCache/

# Python/ML .jvsignore template
__pycache__/
*.py[cod]
*$py.class
.eggs/
dist/
*.egg-info/
.pytest_cache/
```

**Rationale:** Users shouldn't have to figure out what to exclude. Provide battle-tested templates.

---

#### 3. Script Examples Repository

**Effort:** 3-5 days to create examples
**Value:** High - demonstrates integration patterns

**Deliverables:**
- `examples/airflow_operator.py` - Airflow operator that calls JVS
- `examples/unity_build.sh` - Shell script for Unity build + snapshot
- `examples/agent_runner.py` - Python script for agent experiments
- `examples/etl_pipeline.sh` - ETL pipeline with JVS checkpoints

**Rationale:** Users want to see how to integrate JVS into their existing workflows. Examples are cheaper than built-in integrations.

---

### Tier 2: Medium Value, Medium Complexity (Consider for v8.0)

#### 4. Simplified Config

**Effort:** 3-5 days implementation
**Value:** Medium - reduces confusion

**Proposal:**
```yaml
# .jvs/config.yaml - Simplified to 3 settings
engine: juicefs-clone  # or auto
default_tags:
  - $USER             # Automatically tag with username
progress: true        # or false
```

**Remove:**
- `output_format` (use `--json` flag instead)
- Compression config (use `--compress` flag instead)
- Complex retention policies (use CLI flags instead)

**Rationale:** Config should be for user preferences, not comprehensive workflow configuration.

---

#### 5. Snapshot Notes Enhancement

**Effort:** 2-3 days implementation
**Value:** Medium - better organization

**Proposal:**
```bash
# Allow richer notes (multi-line, templates)
jvs snapshot <<EOF
ML Experiment: ResNet50 v2
Dataset: ImageNet (subset: 100k images)
Hyperparameters:
  - Learning rate: 0.001
  - Batch size: 256
  - Epochs: 100
Result: 92.3% accuracy
EOF

# Or use templates
jvs snapshot --note-template ml_experiment
```

**Rationale:** Better notes improve searchability without adding complexity.

---

### Tier 3: Low Value or High Complexity (Don't Do These Yet)

#### ❌ File Locking

**Why users want it:** Game devs need to prevent concurrent asset editing

**Why to reject:**
- Requires distributed coordination (locks, leases, heartbeats)
- Violates local-first design
- Adds significant complexity
- Perforce already solves this problem

**Alternative:** Document how to use JVS + external coordination (e.g., "use Perforce for collaboration, JVS for local versioning")

---

#### ❌ Merge Support

**Why users want it:** Git-trained users expect merge functionality

**Why to reject:**
- Binary files cannot be meaningfully merged
- Fundamentally conflicts with snapshot-first model
- Would require diff/patch storage
- Massive complexity increase

**Alternative:** Forking worktrees is the correct model for JVS

---

#### ❌ Built-in Scheduling

**Why users want it:** ETL users want automated snapshots

**Why to reject:**
- Airflow, Prefect, Dagster already exist
- Reinventing the wheel
- Out of scope for a versioning tool

**Alternative:** Provide Airflow operator examples

---

#### ❌ Remote Protocol

**Why users want it:** Git-like push/pull

**Why to reject:**
- JuiceFS handles data transport
- Would require authentication, authorization, conflict resolution
- Violates local-first design

**Alternative:** Document `juicefs sync` and `rsync` for repository sync

---

#### ❌ Container Integration

**Why users want it:** Agent users want containerized JVS

**Why to reject:**
- Docker Compose can mount JVS workspaces
- Kubernetes can use PVCs backed by JuiceFS
- Integration should be external, not built-in

**Alternative:** Provide Docker Compose examples in documentation

---

#### ❌ Web UI

**Why users want it:** GUI for non-technical users

**Why to reject:**
- Massive implementation effort
- Requires authentication, authorization, session management
- CLI is sufficient for target users (technical users)

**Alternative:** Improve CLI UX (better help text, clearer error messages)

---

## Proposed Feature Removals (Simplification)

### Remove to Reduce Complexity

| Feature | Remove? | Replacement |
|---------|---------|-------------|
| `--output-format table` | ✅ Yes | Use default text output |
| Engine autodetection | ✅ Yes | Require explicit `--engine` or `engine: auto` in config |
| `--compress max` | ✅ Maybe | Keep only `--compress` (single level) or remove entirely |
| Snapshot templates | ✅ Yes | Use shell scripts instead |
| Multiple GC policy types | ✅ Yes | Keep only `--keep-daily N` |

### Rationale for Removals

1. **Table output format:** Adds maintenance burden, text output is sufficient
2. **Engine autodetection:** Heuristics are fragile; explicit is better than implicit
3. **Complex GC policies:** Most users only need daily retention
4. **Snapshot templates:** Edge case, shell scripts are more flexible
5. **Multiple compression levels:** Users rarely need more than fast/max

---

## Revised Feature Scope for v7.2

### Core Features (Maintain)

- `jvs init` - Initialize repository
- `jvs snapshot` - Create snapshot (with tags, partial paths)
- `jvs restore` - Restore workspace to snapshot
- `jvs worktree fork` - Create new worktree
- `jvs history` - Show snapshot history
- `jvs verify` - Verify integrity
- `jvs doctor` - Health check

### Simplify

- Config: Reduce to 3 settings (engine, default_tags, progress)
- GC: Single `--keep-daily N` policy
- Compression: Single `--compress` flag or remove

### Add

- Quick start guides for each target scenario
- `.jvsignore` templates for Unity, Unreal, Python
- Script examples repository
- Enhanced snapshot notes (multi-line)

### Remove

- Table output format
- Engine autodetection heuristics
- Snapshot templates
- Complex GC policy language

---

## Success Metrics

### Adoption Metrics

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Game studio adoption | 5+ studios by end of 2026 | GitHub issues, case studies |
| Agent platform adoption | 3+ platforms by end of 2026 | GitHub issues, integrations |
| ETL pipeline adoption | 10+ data teams by end of 2026 | GitHub issues, blog posts |

### Quality Metrics

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Documentation coverage | 100% of features have examples | Docs review |
| CLI help clarity | < 5 seconds to understand command | User testing |
| Config file complexity | < 5 settings | Code review |

---

## Implementation Phasing

### Phase 1: Documentation (Immediate - Week 1)

1. Create quick start guides for 3 target scenarios
2. Create `.jvsignore` templates
3. Create script examples repository

### Phase 2: Simplification (v7.2 - Month 1-2)

1. Simplify config file schema
2. Simplify GC policy language
3. Remove unused output formats
4. Update documentation to reflect changes

### Phase 3: Enhanced Notes (v7.3 - Month 3-4)

1. Multi-line snapshot notes support
2. Note templates (if still needed after Phase 1)
3. Improved search/filtering

### Phase 4: Evaluation (v8.0 Planning - Month 6)

1. Review adoption metrics
2. Gather user feedback
3. Evaluate Tier 2 features based on demand

---

## Guardrails: KISS Principle Checklist

Before adding any feature, ask:

1. **Does this solve a real problem for our target users?**
   - If no, reject.

2. **Is there an existing tool that solves this problem?**
   - If yes, document integration instead of building.

3. **Can this be done with a shell script instead?**
   - If yes, create example script instead of feature.

4. **Does this add significant complexity?**
   - If yes, reject unless value is proven.

5. **Would this require breaking changes?**
   - If yes, defer to major version.

---

## Conclusion

JVS is at its best when it focuses on its core strength: **O(1) snapshots for large files**.

The proposed refinements:
- **Keep** the essential features that provide value
- **Simplify** features that have become overly complex
- **Add** only high-value, low-complexity features
- **Remove** features that violate KISS

By maintaining this discipline, JVS will remain a focused, reliable tool that solves real problems for its target users.

---

*Remember: "Perfection is achieved not when there is nothing more to add, but when there is nothing left to take away." — Antoine de Saint-Exupéry*
