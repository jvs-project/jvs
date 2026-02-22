# JVS Target Users: Pain Points & Requirements

**Version:** v7.0
**Last Updated:** 2026-02-23

---

## Overview

This document analyzes three target user scenarios for JVS (Juicy Versioned Workspaces), identifying their core pain points and evaluating how JVS's current feature set addresses their needs. The goal is to ensure JVS solves real problems while maintaining the KISS principle.

---

## Target User 1: Game Asset Management

### User Persona

| Attribute | Description |
|-----------|-------------|
| **Company Type** | Game studios (indie to mid-size) |
| **Team Size** | 5-50 developers/artists |
| **Primary Role** | Technical Artists, Game Developers, Asset Pipeline Engineers |
| **Technical Level** | Medium-High (comfortable with CLI, build pipelines) |
| **Typical Workspace** | 100GB-2TB per project |

### Core Pain Points

#### 1. Large Binary File Versioning

**Problem:**
- Game assets (3D models, textures, audio, video) range from 100MB to 5GB+ per file
- Git struggles with large binaries (repository bloat, slow clones)
- Git LFS adds complexity and bandwidth costs
- Artists need to version assets without understanding Git

**Current Solutions & Limitations:**
| Solution | Pros | Cons |
|----------|------|------|
| **Git LFS** | Familiar to devs | Bandwidth costs, slow checkout, file size limits |
| **Perforce** | Designed for binaries | Expensive, complex setup, server maintenance |
| **Unity Version Control** | Integrated with Unity | Vendor lock-in, cloud-only pricing |
| **Manual file copying** | Simple | No version history, prone to errors |

**Real-world quotes from research:**
> "Git stores full copies for every change. Even small texture modifications cause massive repo bloat."
> "Artists shouldn't need to understand branching, merging, or conflict resolution."

#### 2. Unity/Unreal Metadata Synchronization

**Problem:**
- Unity generates `.meta` files that must stay in sync with assets
- Moving/renaming assets breaks GUID references if `.meta` files are lost
- `Library/` and `Temp/` directories should never be versioned

**Requirement:**
- Version `Assets/` and `ProjectSettings/` only
- Include `.meta` files automatically
- Exclude generated directories

#### 3. Team Collaboration Without Merge Conflicts

**Problem:**
- Binary assets cannot be merged (no 3-way merge for a `.psd` or `.fbx`)
- Multiple artists working on the same asset causes conflicts
- "Lock/edit" workflows slow down development

**Current Pattern:**
- File locking (Perforce-style) prevents concurrent work
- Or manual coordination ("who is working on character_model.fbx?")
- Neither solution scales well

### JVS Feature Match Analysis

| Pain Point | JVS Solution | Fit |
|------------|--------------|-----|
| Large binary file versioning | O(1) snapshots via juicefs-clone | ✅ Excellent |
| Git LFS complexity | Snapshot-first, no separate LFS | ✅ Excellent |
| Unity .meta file tracking | Snapshot entire workspace | ✅ Good |
| Binary merge conflicts | No merge, fork instead | ⚠️ Different paradigm |
| Asset locking | Not supported (removed in v6.7) | ❌ Gap |

**Key Insight:** JVS solves the storage problem but doesn't solve the collaboration coordination problem. This is by design - JVS is local-first and single-writer.

### Recommended JVS Workflow for Game Devs

```bash
# Project setup
cd /mnt/juicefs/game-projects
jvs init mygame
cd mygame/main

# Import Unity project
cp -r ~/UnityProjects/MyGame/* .
jvs snapshot "Initial Unity project import" --tag unity --tag baseline

# Daily workflow for artists
cd /mnt/juicefs/game-projects/mygame/main
jvs snapshot "Before modifying character" --tag prework
# ... work in Unity ...
jvs snapshot "Updated character model v2" --tag character --tag assets

# If work went wrong, rollback
jvs restore --latest-tag prework
```

### Feature Recommendations

**Keep (Essential):**
- O(1) snapshots - critical for large assets
- Partial snapshots - snapshot only `Assets/` directory
- Tags for organization (`unity`, `assets`, `v1.0`)

**Add (High Value, Low Complexity):**
- Unity-specific helper script (wrapper around JVS for Unity projects)
- Predefined `.jvsignore` template for Unity/Unreal

**Do NOT Add (violates KISS):**
- File locking - use external coordination or Perforce for this
- Binary diff/merge - technically infeasible for most formats
- Unity Editor plugin - adds too much complexity

---

## Target User 2: AI Agent Sandbox Environments

### User Persona

| Attribute | Description |
|-----------|-------------|
| **Company Type** | AI Research Labs, AI Agent Platform Companies |
| **Team Size** | 2-20 engineers/researchers |
| **Primary Role** | ML Engineers, Agent Infrastructure Engineers |
| **Technical Level** | High (comfortable with Docker, orchestration) |
| **Typical Workspace** | 1-100GB per agent environment |

### Core Pain Points

#### 1. Rapid Environment Creation/Deletion

**Problem:**
- Each agent run needs a clean, deterministic starting state
- Containers are too heavy for rapid iteration
- VMs are too slow to spin up
- Need to run 100-1000s of experiments in parallel

**Current Solutions & Limitations:**
| Solution | Pros | Cons |
|----------|------|------|
| **Docker** | Fast startup, isolated | Shared kernel security concerns |
| **MicroVMs (gVisor, Firecracker)** | Strong isolation | Slower startup, complex setup |
| **E2B / AgentBox** | Purpose-built | Cloud-only, vendor lock-in |
| **Manual directory management** | Simple | No version tracking, hard to reproduce |

#### 2. Deterministic State Reset

**Problem:**
- Agent experiments must be reproducible
- File system state changes during execution
- Need to return to exact baseline state between runs

**Real-world quotes from research:**
> "We need to snapshot the environment state before each agent run so we can reproduce exact results."
> "Container startup takes 2-3 seconds, which adds up when running thousands of experiments."

#### 3. Parallel Experiment Execution

**Problem:**
- Multiple agents need independent environments
- Shared state causes interference
- Need to track which snapshot produced which result

### JVS Feature Match Analysis

| Pain Point | JVS Solution | Fit |
|------------|--------------|-----|
| O(1) environment snapshots | juicefs-clone engine | ✅ Perfect |
| Deterministic state reset | `jvs restore` to baseline | ✅ Perfect |
| Parallel execution | Worktrees (one per agent) | ✅ Good |
| Tracking results | Tags per experiment run | ✅ Good |
| Container-level isolation | Not provided (filesystem only) | ⚠️ Different layer |

**Key Insight:** JVS operates at the filesystem layer, not the container/VM layer. It complements rather than replaces containerization.

### Recommended JVS Workflow for Agent Sandboxes

```bash
# Baseline environment setup
cd /mnt/juicefs/agent-sandbox
jvs init agent-base
cd agent-base/main
cp -r /baseline/agent/* .
jvs snapshot "Agent baseline v1" --tag baseline --tag v1

# Agent execution loop
for RUN in {1..1000}; do
    # Create isolated worktree for this run
    jvs worktree fork run-$RUN

    # Restore to baseline
    cd worktrees/run-$RUN/main
    jvs restore baseline

    # Run agent
    python agent.py --seed $RUN --output results/$RUN.json

    # Snapshot result state
    RESULT=$(cat results/$RUN.json | jq -r '.outcome')
    jvs snapshot "Run $RUN: $RESULT" --tag "run-$RUN" --tag agent

    # Cleanup (optional)
    cd ../..
done
```

### Feature Recommendations

**Keep (Essential):**
- O(1) snapshots - critical for rapid iteration
- Worktrees - provides isolation for parallel runs
- `jvs restore` - enables deterministic reset

**Add (High Value, Low Complexity):**
- Script examples in documentation for agent workflows
- Template `jvs.yaml` for agent experiment tracking

**Do NOT Add (violates KISS):**
- Container orchestration - use Docker/Kubernetes for this
- Built-in Python agent framework - out of scope
- Scheduling system - use external tools

---

## Target User 3: Data ETL Pipelines

### User Persona

| Attribute | Description |
|-----------|-------------|
| **Company Type** | Data-Driven Companies (SaaS, Analytics, FinTech) |
| **Team Size** | 5-50 data engineers |
| **Primary Role** | Data Engineers, ML Engineers, Data Platform Engineers |
| **Technical Level** | High (comfortable with SQL, Python, orchestration) |
| **Typical Workspace** | 10GB-10TB per dataset |

### Core Pain Points

#### 1. Dataset Versioning for Reproducibility

**Problem:**
- ML models depend on exact data versions
- "Which dataset version produced this model?" is unanswerable
- ETL pipeline changes create implicit data drift
- Auditors require data lineage

**Current Solutions & Limitations:**
| Solution | Pros | Cons |
|----------|------|------|
| **Git + Git LFS** | Familiar | Doesn't scale to TB datasets |
| **DVC** | Designed for ML data | Complex setup, cache management overhead |
| **Databricks Delta / Iceberg** | Time travel queries | Vendor lock-in, cloud-only |
| **S3 versioning** | Simple bucket-level | No dataset-level semantics |

**Real-world quotes from research:**
> "Yesterday our model had 95% accuracy, today it's 83%. We don't know which data changed."
> "We need to prove to auditors that our Q4 report used the correct data."

#### 2. Pipeline Integration

**Problem:**
- Data versioning must integrate with Airflow/Prefect/Dagster
- Need to snapshot data after ETL stages complete
- Pipeline failures should not create invalid snapshots

#### 3. Incremental Processing Support

**Problem:**
- Daily data ingestion (TB/month)
- Cannot afford to snapshot everything every day
- Need to track which partitions have been processed

### JVS Feature Match Analysis

| Pain Point | JVS Solution | Fit |
|------------|--------------|-----|
| Dataset versioning | O(1) snapshots for large datasets | ✅ Excellent |
| Pipeline integration | CLI-based, works in scripts | ✅ Good |
| Incremental processing | Partial snapshots | ✅ Good |
| Data lineage | Snapshot notes + tags | ⚠️ Basic (sufficient for most) |
| Time travel queries | Not provided (use Iceberg/Delta) | ⚠️ Different layer |

**Key Insight:** JVS complements rather than replaces data lakehouse technologies. Use JVS for workspace-level snapshots, Iceberg/Delta for table-level time travel.

### Recommended JVS Workflow for ETL

```bash
# ETL pipeline with JVS snapshots

# Stage 1: Ingest raw data
ingest_raw() {
    cd /mnt/juicefs/etl-pipeline/main
    jvs restore baseline

    python ingest_raw.py --date $TODAY
    jvs snapshot "Raw ingestion $TODAY" --tag raw --tag $TODAY
}

# Stage 2: Clean and transform
transform_data() {
    python transform.py --input raw/ --output processed/
    jvs snapshot "Transformed $TODAY" --tag processed --tag $TODAY
}

# Stage 3: Feature engineering
build_features() {
    python build_features.py --input processed/ --output features/
    jvs snapshot "Features $TODAY" --tag features --tag $TODAY
}

# Stage 4: Train model
train_model() {
    python train.py --input features/ --output model.pkl
    jvs snapshot "Model trained on $TODAY" --tag model --tag $TODAY
}

# Full pipeline (orchestrated by Airflow)
ingest_raw && transform_data && build_features && train_model
```

### Feature Recommendations

**Keep (Essential):**
- O(1) snapshots - critical for large datasets
- Partial snapshots - snapshot only transformed directories
- Tags for daily organization

**Add (High Value, Low Complexity):**
- Airflow operator examples in documentation
- Pre-commit hook for pipeline validation

**Do NOT Add (violates KISS):**
- Built-in orchestrator - use Airflow/Prefect
- Data catalog integration - use external tools
- SQL-based querying - not a database

---

## Cross-Cutting Analysis: Common Patterns

### Pattern 1: Large Files Are the Norm

All three user scenarios deal with files that Git cannot handle efficiently:

| Scenario | Typical File Size | Count | Total Size |
|----------|------------------|-------|------------|
| Game Assets | 100MB - 5GB | 1,000s | 100GB - 2TB |
| Agent Sandboxes | 1MB - 10GB | 10s - 100s | 10GB - 100GB |
| ETL Pipelines | 1GB - 100GB | 10s | 100GB - 10TB |

**JVS Strength:** O(1) snapshots via juicefs-clone solve this uniformly well.

### Pattern 2: Version Semantics > Diff Semantics

None of these scenarios care about per-line diffs:

- Game assets: Binary files have no meaningful diff
- Agent sandboxes: Need complete state, not changes
- ETL pipelines: Dataset-level versions, not row-level changes

**JVS Strength:** Snapshot-first design aligns perfectly.

### Pattern 3: Reproducibility Is Critical

All three scenarios need to answer "which exact state produced this result?"

- Game assets: Which asset version was in the build?
- Agent runs: Which environment state produced this outcome?
- ETL: Which dataset version trained this model?

**JVS Strength:** Snapshot IDs + tags provide clear lineage.

### Pattern 4: Collaboration Needs Vary

| Scenario | Collaboration Model | JVS Fit |
|----------|-------------------|---------|
| Game Assets | Multi-user, needs coordination | ⚠️ Requires external coordination |
| Agent Sandboxes | Single-user or batch jobs | ✅ Excellent |
| ETL Pipelines | Single writer, orchestrator-driven | ✅ Good |

**Key Insight:** JVS's single-writer model fits 2/3 scenarios perfectly. For game dev, external coordination is needed anyway (due to binary conflicts).

---

## KISS Principle Assessment

### Current JVS Features (Essential, Keep)

| Feature | Scenario Use | Complexity | Verdict |
|---------|--------------|------------|---------|
| O(1) snapshots | All | Low | ✅ Keep |
| Snapshot restore | All | Low | ✅ Keep |
| Worktrees | Agents, Game Dev | Low | ✅ Keep |
| Tags | All | Low | ✅ Keep |
| Partial snapshots | Game Dev, ETL | Low | ✅ Keep |
| Verify integrity | All | Low | ✅ Keep |

### Current JVS Features (Low Value, Could Remove)

| Feature | Use Case | Complexity | Verdict |
|---------|----------|------------|---------|
| Config files | Automation | Medium | ⚠️ Keep but simplify |
| GC retention policies | Ops | Medium | ⚠️ Keep but simplify |
| Compression | ETL (maybe) | Low | ⚠️ Edge case, maybe remove |

### Proposed Features (Do NOT Add - Violates KISS)

| Proposed Feature | Why It's Tempting | Why to Reject |
|------------------|-------------------|---------------|
| File locking | Game devs want it | Requires distributed coordination, breaks local-first |
| Merge support | Users expect Git-like behavior | Fundamentally conflicts with snapshot-first model |
| Built-in scheduling | ETL users want automation | Use Airflow/Prefect instead |
| Remote protocol | Users want push/pull | Use JuiceFS sync or rsync instead |
| Container integration | Agent users want isolation | Use Docker/K8s instead |
| Database backend | ETL users want SQL | Not a database, use Iceberg/Delta instead |
| Web UI | All users want GUI | Adds massive complexity, CLI is sufficient |

---

## Recommended Documentation Additions

### 1. Scenario-Specific Quick Start Guides

Create targeted guides for each user type:
- `docs/game_dev_quickstart.md` - Unity/Unreal workflows
- `docs/agent_sandbox_quickstart.md` - Agent experiment workflows
- `docs/etl_pipeline_quickstart.md` - Data engineering workflows

### 2. Integration Recipes

Document common integrations without building them into JVS:
- Airflow operator example (Python script that calls JVS CLI)
- Unity build script example (shell script that snapshots before builds)
- Docker Compose example (volume mount JVS workspace)

### 3. `.jvsignore` Templates

Provide templates for common scenarios:
```bash
# Unity .jvsignore
Library/
Temp/
obj/
*.userprefs

# Unreal .jvsignore
Binaries/
Build/
Intermediate/
Saved/
.vscode/
```

---

## Conclusion

### JVS Is Well-Suited For:

1. **AI Agent Sandboxes** - Perfect fit. O(1) snapshots enable rapid iteration; worktrees provide isolation.

2. **Data ETL Pipelines** - Excellent fit for workspace-level versioning. Complements data lakehouse tools.

3. **Game Asset Storage** - Good fit for versioning large binaries. Collaboration requires external coordination (acceptable trade-off).

### Core Value Proposition:

> **JVS provides O(1) workspace snapshots for large files that Git cannot handle.**

This single value proposition addresses the primary pain point across all three scenarios.

### Product Strategy (KISS):

1. **Focus on the core:** O(1) snapshots for large files
2. **Document patterns:** Show users how to integrate with their existing tools
3. **Avoid scope creep:** Don't build features that existing tools already provide
4. **Accept limitations:** Single-writer model is a feature, not a bug

---

*Remember: The goal is to make JVS indispensable for a specific set of problems, not to solve every version control problem for everyone.*
