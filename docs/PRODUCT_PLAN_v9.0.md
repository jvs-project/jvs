# JVS Product Plan v9.0 - The Integration Release

**Version:** v9.0
**Date:** 2026-02-23
**Status:** Approved
**Theme:** Enabling Tool Integration via Programmatic APIs

---

## Executive Summary

### Strategic Pivot (Post-Team Discussion)

**Original Thesis:** v9.0 = "Polish Release" (CLI UX, docs, examples)

**Revised Thesis:** v9.0 = "Integration Release" (Python SDK, MLflow plugin, stable JSON API)

### The Critical Insight

> "The problem isn't that users don't know about JVS - it's that their **tools** can't integrate with it."
>
> — pm-ai-agent strategic feedback

**Root Cause:** JVS has a CLI but no programmatic API. Production systems (agent platforms, CI/CD, ML pipelines) need APIs to call, not CLI commands to wrap with subprocess.

**Evidence from Market Research:**

| Tool | Why It Won | API Story |
|------|-----------|-----------|
| **E2B** | 40K→15M/month growth | `from e2b import Sandbox` — one line import |
| **MLflow** | De facto ML tracking | `mlflow.log_artifact()` — Python API |
| **DVC** | ML data versioning | Python SDK + Python bindings |

JVS has **zero** programmatic integration story.

---

## Strategic Decision: Three Options

| Option | Focus | Timeline | Trade-off |
|--------|-------|----------|-----------|
| **A** | Polish (CLI UX, docs, examples) | 12 weeks | Human users can adopt, tools cannot |
| **B** | Integration Layer (SDK, plugins) | 12-16 weeks | Enables tool integration, less polish |
| **C** | Both (extended timeline) | 18-20 weeks | Everything, slower to market |

### Recommendation: Option B - Integration Release

**Rationale:**

1. **Human users CAN use CLI today** - JVS is functional for humans
2. **Tools CANNOT use JVS** - No API means no integration
3. **Polish is visible, APIs are foundational** - Polish can wait, APIs unlock ecosystems

**Counter-argument:** What if game studios (Option A's primary audience) don't need APIs because they integrate via build scripts?

**Decision point pending:** pm-game-etl input on game dev integration patterns.

---

## v9.0 Roadmap (Option B: Integration Release)

### Phase 1: Stable JSON API Foundation (Weeks 1-2)

**Problem:** JVS has `--json` flag but no stable schema or documentation.

**Deliverables:**

1. **Stable JSON Schema**
   ```json
   {
     "$schema": "https://jvs.io/schema/v1",
     "type": "object",
     "properties": {
       "version": "1.0",
       "snapshots": {
         "type": "array",
         "items": { "$ref": "#/definitions/snapshot" }
       }
     }
   }
   ```

2. **API Stability Guarantee**
   - JSON format versioned separately from CLI
   - Backward compatibility承诺 for v1.x
   - Deprecation warnings before breaking changes

3. **JSON API Documentation**
   - `docs/API_REFERENCE.md` - All JSON outputs documented
   - `docs/SCHEMA_v1.json` - Official schema file
   - Examples in multiple languages (Python, Go, JS)

**Success Criteria:**
- Third party can parse JVS JSON without breaking on minor releases
- All commands with `--json` documented with response schema

---

### Phase 2: Python SDK (Weeks 3-7)

**Problem:** Agent platforms and ML teams need Python APIs.

**Deliverables:**

1. **`pip install jvs` Package**

```python
# jvs/__init__.py
from jvs import Workspace, Snapshot

# Core API
class Workspace:
    def __init__(self, path: str):
        """Initialize JVS workspace at path."""

    def snapshot(self, note: str, tags: List[str] = None, path: str = None) -> Snapshot:
        """Create snapshot. Returns Snapshot object."""

    def restore(self, snapshot_id: str) -> None:
        """Restore workspace to snapshot."""

    def history(self, filters: dict = None) -> List[Snapshot]:
        """Get snapshot history."""

    @property
    def current_snapshot(self) -> Snapshot:
        """Get current HEAD snapshot."""

class Snapshot:
    def __init__(self, snapshot_id: str):
        """Snapshot reference."""

    @property
    def id(self) -> str:
        """Snapshot ID."""

    @property
    def note(self) -> str:
        """Snapshot note."""

    @property
    def tags(self) -> List[str]:
        """Snapshot tags."""

    @property
    def created_at(self) -> datetime:
        """Snapshot timestamp."""

# Context manager support
from jvs import transaction

with transaction(workspace, note="Before risky operation"):
    # ... do work ...
    # Automatically snapshots on success, rolls back on exception
    pass
```

2. **Implementation Approach**

**Option 1:** CLI wrapper via subprocess (faster, v9.0 feasible)
**Option 2:** Native Go with Python bindings (better performance, more complex)

**Recommendation:** Start with Option 1, plan Option 2 for v9.1

3. **Python SDK Documentation**

```python
# examples/sdk_basic.py
from jvs import Workspace

ws = Workspace("/mnt/juicefs/agent-repo")

# Before running agent
baseline = ws.snapshot("Before agent run")

try:
    result = agent.run()
    ws.snapshot(f"Agent completed: {result}")
except Exception as e:
    ws.restore(baseline.id)
    raise
```

**Success Criteria:**
- `pip install jvs` works on Python 3.9+
- Coverage for all core JVS operations (init, snapshot, restore, history)
- Context manager for atomic operations

---

### Phase 3: MLflow Plugin (Weeks 8-10)

**Problem:** ML teams use MLflow for experiment tracking. JVS needs to integrate.

**Deliverables:**

1. **`mlflow-jvs` Plugin**

```python
# mlflow_jvs/__init__.py
import mlflow
from jvs import Workspace

def log_jvs_snapshot(workspace_path: str, note: str = None, tags: list = None):
    """
    Log current JVS workspace state as MLflow artifact.

    Usage:
        mlflow.log_jvs_snapshot("/mnt/juicefs/experiment", "v1.0")
    """
    ws = Workspace(workspace_path)
    snapshot = ws.snapshot(note, tags)

    # Log snapshot metadata to MLflow
    mlflow.log_param("jvs_snapshot_id", snapshot.id)
    mlflow.log_text(snapshot.to_dict(), "jvs_snapshot.json")

    return snapshot.id

def load_jvs_snapshot(snapshot_id: str, workspace_path: str):
    """
    Restore JVS workspace from snapshot ID logged in MLflow run.

    Usage:
        mlflow.load_jvs_snapshot("abc123", "/mnt/juicefs/experiment")
    """
    ws = Workspace(workspace_path)
    ws.restore(snapshot_id)
```

2. **MLflow UI Integration**

```
MLflow Run UI shows:
┌─────────────────────────────────────┐
│ Parameters:                          │
│   learning_rate: 0.001               │
│   batch_size: 256                    │
│   jvs_snapshot_id: abc12345          │  ← New!
│                                      │
│ Artifacts:                           │
│   model.pkl                          │
│   jvs_snapshot.json                  │  ← New!
│   predictions.csv                    │
└─────────────────────────────────────┘
```

**Success Criteria:**
- `pip install mlflow-jvs` works
- One-line `mlflow.log_jvs_snapshot()` API
- Snapshot metadata visible in MLflow UI

---

### Phase 4: Agent Platform Examples (Weeks 11-12)

**Problem:** Agent platforms need integration patterns.

**Deliverables:**

1. **LangChain Integration**

```python
# examples/langchain_jvs.py
from langchain.agents import AgentExecutor
from jvs import transaction

class JVSToolCallback:
    def on_agent_start(self, agent_input: dict):
        self.baseline = self.ws.snapshot("Before agent")

    def on_agent_end(self, agent_output: dict):
        self.ws.snapshot(f"After agent: {agent_output}")

# Usage
executor = AgentExecutor(
    agent=agent,
    tools=tools,
    callbacks=[JVSToolCallback(workspace="/mnt/juicefs/agent")]
)
```

2. **AutoGen Integration**

```python
# examples/autogen_jvs.py
from autogen import AssistantAgent
from jvs import Workspace

ws = Workspace("/mnt/juicefs/autogen-exp")

with transaction(ws, note="AutoGen conversation"):
    assistant = AssistantAgent(
        name="assistant",
        workspace=ws  # AutoGen uses JVS for state
    )
    assistant.run(task)
```

3. **Parallel Agent Runner**

```python
# examples/parallel_agents.py
from jvs import Workspace
from concurrent.futures import ThreadPoolExecutor

ws = Workspace("/mnt/juicefs/agent-baseline")

def run_agent_experiment(seed: int):
    # Fork worktree for this run
    exp_ws = ws.fork(f"experiment-{seed}")

    # Restore to baseline
    exp_ws.restore("baseline")

    # Run agent
    result = agent.run(seed=seed)

    # Snapshot result
    exp_ws.snapshot(f"Seed {seed}: {result}")

    return result

with ThreadPoolExecutor(max_workers=100) as executor:
    results = executor.map(run_agent_experiment, range(100))
```

**Success Criteria:**
- 3+ agent framework integration examples
- Copy-pasteable patterns for common workflows

---

## Implementation Details

### Python SDK Design

**Architecture Decision: CLI Wrapper vs Native Bindings**

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| **CLI wrapper** | Fast to implement, stable | Slower, subprocess overhead | ✅ v9.0 |
| **Go bindings** | Fast, direct API access | Complex build, cgo overhead | ⏳ v9.1 |
| **gRPC server** | Language agnostic | Separate process, complexity | ❌ Reject |

**v9.0 Implementation (CLI Wrapper):**

```python
# jvs/_cli.py
import subprocess
import json
from typing import List, Optional, Dict, Any

class JVSError(Exception):
    """JVS CLI error."""

class Workspace:
    def __init__(self, path: str):
        self.path = path
        self._check_workspace()

    def snapshot(self, note: str, tags: List[str] = None,
                 path: str = None) -> "Snapshot":
        cmd = ["jvs", "snapshot", note, "--json"]
        if tags:
            cmd.extend([f"--tag={t}" for t in tags])
        if path:
            cmd.extend(["--path", path])

        result = self._run(cmd, cwd=self.path)
        return Snapshot.from_dict(result)

    def _run(self, cmd: List[str], cwd: str = None) -> Dict[str, Any]:
        """Run JVS CLI and return parsed JSON output."""
        result = subprocess.run(
            cmd,
            cwd=cwd or self.path,
            capture_output=True,
            text=True
        )

        if result.returncode != 0:
            raise JVSError(result.stderr)

        return json.loads(result.stdout)
```

### MLflow Plugin Design

```python
# mlflow_jvs/plugin.py
import mlflow
from mlflow import pyfunc
from jvs import Workspace

class JVSArtifact(mlflow.pyfunc.PythonModel):
    """MLflow Python model wrapper for JVS snapshots."""

    def __init__(self, workspace_path: str, snapshot_id: str):
        self.workspace_path = workspace_path
        self.snapshot_id = snapshot_id

    def load_context(self, context):
        """Restore JVS snapshot when loading model."""
        ws = Workspace(self.workspace_path)
        ws.restore(self.snapshot_id)

@staticmethod
def log_jvs_snapshot(workspace_path: str, note: str = None,
                     artifact_path: str = "jvs_snapshot"):
    """
    Log JVS snapshot as MLflow artifact.

    This enables:
    1. Reproducible model loading (snapshot restored with model)
    2. Experiment tracking (workspace state captured)
    3. Model versioning (data + code + environment)
    """
    ws = Workspace(workspace_path)
    snapshot = ws.snapshot(note or "MLflow snapshot")

    # Log snapshot metadata
    mlflow.log_params({
        "jvs_snapshot_id": snapshot.id,
        "jvs_snapshot_note": snapshot.note,
        "jvs_snapshot_tags": ",".join(snapshot.tags or [])
    })

    # Create artifact wrapper
    artifact = JVSArtifact(workspace_path, snapshot.id)
    mlflow.pyfunc.log_model(artifact_path, python_model=artifact)

    return snapshot.id
```

---

## Updated Success Metrics

### Integration Metrics (New)

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Python SDK installs | 500+ | PyPI downloads |
| MLflow plugin installs | 100+ | PyPI downloads |
| External integrations | 3+ | GitHub stars, forks |
| LangChain example usage | TBD | GitHub traffic |
| Agent platform adoption | 2+ | Direct outreach |

### Developer Metrics

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Time to first integration | <10 min | SDK onboarding |
| API documentation coverage | 100% | Docs review |
| SDK test coverage | 80%+ | Code coverage |

---

## What We're NOT Building (v9.0)

### Explicitly Out of Scope

| Feature | Why Not | Alternative |
|---------|---------|-------------|
| **CLI Polish** (colors, rich errors) | Deferring to v9.1 | CLI is functional for humans |
| **Web UI** | Massive complexity | SDK enables others to build |
| **Container orchestration** | K8s/Docker handle this | Provide examples |
| **File locking** | Distributed coordination | External coordination |
| **Merge support** | Binary files can't merge | Worktree forks |
| **Remote protocol** | JuiceFS handles transport | `juicefs sync` |

---

## Open Questions

### Q1: Game Dev Integration Pattern

**Question:** Do game studios need programmatic JVS APIs?

**Context:**
- Game engines have build pipelines (Unity Editor, Unreal Build Tool)
- Integration could be CLI calls in build scripts
- OR could be native C#/C++ APIs

**Waiting on:** pm-game-etl analysis of game dev integration patterns

### Q2: SDK Performance

**Question:** Is CLI wrapper performance acceptable?

**Considerations:**
- Subprocess overhead ~10-50ms per call
- Snapshot operations are O(1) but still take time
- For 1000s of parallel agents, overhead matters

**Decision point:** Benchmark CLI wrapper vs native bindings in Phase 2

### Q3: API Stability Timeline

**Question:** How long to guarantee JSON API stability?

**Options:**
- v1.x stable guarantee (12+ months)
- v9.0-v9.5 stable (6 months)
- Per-version stability with deprecation warnings

**Recommendation:** v1.x stable guarantee to enable ecosystem confidence

---

## Revised Positioning Statements

### For AI/ML Engineers (Primary Audience)
> **"JVS Python SDK: O(1) workspace snapshots in three lines of code. Integrate with MLflow, LangChain, and your agent framework."**

### For Game Developers (Secondary)
> **"JVS: Instant snapshots of 100GB+ assets. CLI for build pipelines, no server needed."**

### For Data Engineers (Tertiary)
> **"JVS + MLflow: Dataset versioning without DVC complexity. Log snapshots alongside your experiments."**

---

## Conclusion

### The v9.0 Promise (Revised)

> **"JVS v9.0 enables programmatic workspace versioning. Import the SDK, call snapshot(), integrate with your tools."**

### Strategic Rationale

**Polish is visible, APIs are foundational.**

- A polished CLI makes humans happy
- A stable API unlocks entire ecosystems

JVS's CLI is functional. The gap is integration. v9.0 closes that gap.

---

*End of Document*
