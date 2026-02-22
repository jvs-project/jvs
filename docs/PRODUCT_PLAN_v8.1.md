# JVS Product Plan v8.1 - The API Foundation Release

**Version:** v8.1
**Date:** 2026-02-23
**Status:** Revised
**Theme:** Stable JSON API for Tool Integration

---

## Executive Summary

### Scope Decision

**Original Thesis:** v8.1 = "Integration Release" (Python SDK, MLflow plugin, stable JSON API)

**Revised Thesis:** v8.1 = "API Foundation Release" (Stable JSON API only)

### Rationale

User research (see `PRODUCT_REPORT_v8.1.md`) identified **22 user cases** across AI/ML, Game Dev, and Data Engineering segments. While 73% of cases need programmatic access, **Phase 1 (Stable JSON API) is the foundation** that all future integrations depend on.

| Phase | Scope | Decision |
|-------|-------|----------|
| Phase 1: Stable JSON API | JSON schema, documentation, stability guarantees | ✅ v8.1 |
| Phase 2: Python SDK | `pip install jvs` package | ⏳ Deferred to v8.2 |
| Phase 3: MLflow Plugin | `mlflow-jvs` integration | ⏳ Deferred to v8.2 |
| Phase 4: Agent Examples | LangChain, AutoGen patterns | ⏳ Deferred to v8.2 |

### Why Phase 1 First

1. **Foundation for everything else** - SDK wraps CLI JSON output; unstable JSON = broken SDK
2. **Smallest scope, highest leverage** - 2 weeks of work enables all future integrations
3. **User feedback opportunity** - Ship stable JSON, measure what users build on top
4. **Risk reduction** - Validate API design before committing to SDK surface

---

## v8.1 Roadmap

### Phase 1: Stable JSON API Foundation

**Timeline:** Weeks 1-2

**Problem:** JVS has `--json` flag but no stable schema or documentation.

---

#### 1.1 JSON Schema Definition

Define stable schemas for all JSON outputs.

**Snapshot Schema:**
```json
{
  "$schema": "https://jvs.io/schema/v1/snapshot.json",
  "$id": "https://jvs.io/schema/v1/snapshot.json",
  "type": "object",
  "required": ["id", "created_at", "workspace_id"],
  "properties": {
    "id": {
      "type": "string",
      "description": "Unique snapshot identifier (e.g., 'abc123')"
    },
    "created_at": {
      "type": "string",
      "format": "date-time",
      "description": "ISO 8601 timestamp"
    },
    "workspace_id": {
      "type": "string",
      "description": "Workspace this snapshot belongs to"
    },
    "note": {
      "type": "string",
      "description": "User-provided note (may be empty)"
    },
    "tags": {
      "type": "array",
      "items": { "type": "string" },
      "description": "User-provided tags"
    },
    "parent_id": {
      "type": ["string", "null"],
      "description": "Parent snapshot ID (null for initial snapshot)"
    },
    "payload_hash": {
      "type": "string",
      "description": "SHA-256 hash of snapshot payload"
    },
    "descriptor_path": {
      "type": "string",
      "description": "Path to snapshot descriptor file"
    }
  }
}
```

**Snapshot List Schema:**
```json
{
  "$schema": "https://jvs.io/schema/v1/snapshot-list.json",
  "$id": "https://jvs.io/schema/v1/snapshot-list.json",
  "type": "object",
  "required": ["version", "snapshots"],
  "properties": {
    "version": {
      "type": "string",
      "const": "1.0",
      "description": "Schema version"
    },
    "snapshots": {
      "type": "array",
      "items": { "$ref": "https://jvs.io/schema/v1/snapshot.json" }
    },
    "truncated": {
      "type": "boolean",
      "description": "True if results were paginated"
    }
  }
}
```

**Error Schema:**
```json
{
  "$schema": "https://jvs.io/schema/v1/error.json",
  "$id": "https://jvs.io/schema/v1/error.json",
  "type": "object",
  "required": ["version", "error"],
  "properties": {
    "version": {
      "type": "string",
      "const": "1.0"
    },
    "error": {
      "type": "object",
      "required": ["code", "message"],
      "properties": {
        "code": {
          "type": "string",
          "description": "Machine-readable error code (e.g., 'E_LOCK_CONFLICT')"
        },
        "message": {
          "type": "string",
          "description": "Human-readable error message"
        },
        "details": {
          "type": "object",
          "description": "Additional error context"
        }
      }
    }
  }
}
```

---

#### 1.2 API Stability Guarantee

**Versioning Policy:**

1. **JSON schema versioned independently** from CLI version
2. **Version field in every response** (`"version": "1.0"`)
3. **Backward compatibility guarantee** for v1.x:
   - New fields may be added
   - Existing fields will not be removed or renamed
   - Existing field types will not change
4. **Deprecation process:**
   - Field marked deprecated in schema (min 1 release cycle)
   - Deprecation warning in `--json` output
   - Removal only in major version bump

**Compatibility Matrix:**

| JVS CLI Version | JSON Schema Version | Notes |
|-----------------|---------------------|-------|
| v8.1.x | 1.0 | Initial stable API |
| v8.2.x | 1.0, 1.1 | Backward compatible additions |
| v10.0.0 | 2.0 | Breaking changes allowed |

---

#### 1.3 JSON API Documentation

**Deliverables:**

| File | Purpose |
|------|---------|
| `docs/API_REFERENCE.md` | All commands with `--json` documented |
| `docs/schemas/v1/snapshot.json` | Official snapshot schema |
| `docs/schemas/v1/snapshot-list.json` | Official list schema |
| `docs/schemas/v1/error.json` | Official error schema |
| `docs/schemas/v1/workspace.json` | Workspace info schema |
| `docs/schemas/v1/worktree.json` | Worktree info schema |

**API Reference Structure:**

```markdown
# API Reference

## Version

All JSON responses include a `version` field. Current version: `1.0`

## Commands

### jvs init --json

**Response Schema:** workspace.json

**Example:**
```json
{
  "version": "1.0",
  "workspace": {
    "id": "ws-abc123",
    "path": "/mnt/juicefs/myrepo",
    "created_at": "2026-02-23T10:00:00Z"
  }
}
```

### jvs snapshot --json

**Response Schema:** snapshot.json

**Example:**
```json
{
  "version": "1.0",
  "snapshot": {
    "id": "snap-xyz789",
    "created_at": "2026-02-23T10:05:00Z",
    "note": "Before risky operation",
    "tags": ["baseline"],
    "parent_id": "snap-abc456",
    "payload_hash": "sha256:abc123..."
  }
}
```

...
```

---

#### 1.4 Command Coverage

All commands with `--json` flag must have documented schemas:

| Command | Schema | Status |
|---------|--------|--------|
| `jvs init --json` | workspace.json | Required |
| `jvs snapshot --json` | snapshot.json | Required |
| `jvs restore --json` | snapshot.json | Required |
| `jvs history --json` | snapshot-list.json | Required |
| `jvs worktree list --json` | worktree-list.json | Required |
| `jvs worktree fork --json` | worktree.json | Required |
| `jvs verify --json` | verify-result.json | Required |
| `jvs doctor --json` | doctor-result.json | Required |
| `jvs gc plan --json` | gc-plan.json | Required |

---

## Success Criteria

| Criterion | Measurement |
|-----------|-------------|
| Schema completeness | All 9 commands have JSON schemas |
| Schema validation | All `--json` outputs validate against schemas |
| Documentation | 100% API coverage in API_REFERENCE.md |
| Stability test | Schema version in output, no undocumented fields |
| Examples | 3+ examples in different languages (Python, Go, Bash) |

---

## What We're NOT Building (v8.1)

### Explicitly Out of Scope

| Feature | Why Not | v8.2+ Plan |
|---------|---------|------------|
| **Python SDK** | Foundation first | CLI wrapper in v8.2 |
| **MLflow plugin** | Depends on SDK | v8.2 or v9.2 |
| **Agent examples** | Depends on SDK | v8.2 or v9.2 |
| **CLI Polish** | Functional for humans | v8.2 |
| **Web UI** | Massive complexity | Future |
| **Environment capture** | Complex scope | Future |

---

## Positioning Statement

> **"JVS v8.1: Stable JSON API for workspace versioning. Parse our output with confidence, build integrations that don't break."**

### For Tool Builders
> "JVS provides versioned JSON schemas for all CLI commands. Your integration won't break on minor releases."

### For AI/ML Engineers
> "Build your own SDK wrapper with confidence. Stable JSON API guaranteed."

### For Game Developers
> "Script JVS in your build pipelines. Documented JSON output for CI/CD integration."

---

## Timeline

| Week | Deliverable |
|------|-------------|
| 1 | Schema definitions for all commands |
| 1 | Schema validation in CLI |
| 2 | API_REFERENCE.md documentation |
| 2 | Examples in Python, Go, Bash |
| 2 | Integration testing |

---

## Open Questions

### Q1: Should we include a minimal Python wrapper?

**Context:** Research shows 73% of user cases need programmatic access. A 50-line wrapper would provide immediate value.

**Options:**
| Option | Effort | Value |
|--------|--------|-------|
| No wrapper, JSON only | 0 | Users build their own |
| Minimal wrapper (50 lines) | 1 day | Basic `snapshot()` / `restore()` |
| Full SDK | 2 weeks | Deferred to v8.2 |

**Recommendation:** Consider minimal wrapper as stretch goal.

### Q2: How to handle schema evolution?

**Options:**
| Option | Pros | Cons |
|--------|------|------|
| Single version per JVS release | Simple | Breaks on downgrade |
| Version negotiation | Flexible | Complex |
| Strict backward compat | Predictable | Limits changes |

**Recommendation:** Strict backward compatibility for v1.x.

---

## References

- `PRODUCT_REPORT_v8.1.md` - User case research findings
- `02_CLI_SPEC.md` - CLI command specifications
- `04_SNAPSHOT_SCOPE_AND_LINEAGE_SPEC.md` - Snapshot data model

---

*End of Document*
