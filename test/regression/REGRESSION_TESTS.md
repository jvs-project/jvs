# Regression Tests

This directory contains regression tests for bugs that have been fixed in JVS. These tests ensure that fixed bugs don't reappear in future versions.

## Purpose

Regression tests serve as a permanent record of bugs that were found and fixed. They:

1. **Prevent recurrence**: Ensure bugs stay fixed
2. **Document issues**: Provide a historical record of problems encountered
3. **Improve confidence**: Make refactoring safer by validating past fixes
4. **Support CII Silver Badge**: Demonstrate commitment to quality (required for CNCF/CII Silver badge)

## Adding a Regression Test

When you fix a bug, add a regression test following these steps:

### 1. Create the Test Function

```go
// TestRegression_GarbageCollectionLeak tests that GC properly cleans up
// orphaned snapshots when their parent is deleted.
//
// Bug: GC was not cleaning up orphaned snapshots
// Fixed: 2024-02-20, PR #456
// Issue: #123
func TestRegression_GarbageCollectionLeak(t *testing.T) {
    // Test code here
}
```

### 2. Follow the Naming Convention

Use the format: `TestRegression_<BriefDescription>`

- Use a descriptive name that explains what was broken
- Use CamelCase with underscores
- Include the issue number in the comment block, not the function name

### 3. Document the Bug

Each test MUST include a comment block with:

```go
// TestRegression_GarbageCollectionLeak tests that [what it tests].
//
// Bug: [Brief description of what was broken]
// Fixed: [YYYY-MM-DD], PR #[number]
// Issue: #[number]
```

### 4. Update This Document

Add an entry to the "Regression Test Catalog" section below:

```markdown
### Garbage Collection Leak
- **Test**: `TestRegression_GarbageCollectionLeak`
- **Fixed**: 2024-02-20, PR #456
- **Description**: GC was not cleaning up orphaned snapshots when parent was deleted
```

### 5. Link to the Original Issue

If there's a GitHub issue or PR, include it in the commit message:

```
test(regression): add test for GC orphaned snapshot leak

Fixes #123

This test ensures that GC properly cleans up orphaned snapshots
when their parent snapshot is deleted.
```

## Running Regression Tests

Regression tests use the `conformance` build tag:

```bash
# Run all regression tests
go test -tags conformance -v ./test/regression/...

# Run a specific regression test
go test -tags conformance -v ./test/regression/... -run TestRegression_GarbageCollectionLeak

# Run with coverage
go test -tags conformance -cover ./test/regression/...

# Run as part of full conformance suite
make conformance
```

## Regression Test Catalog

| Test Name | Fixed | Description |
|-----------|-------|-------------|
| `TestRegression_TemplateExample` | Template | Example template for new regression tests |
| `TestRegression_RestoreNonExistentSnapshot` | 2024-02-20 | Restore fails gracefully for non-existent snapshot (no panic) |
| `TestRegression_SnapshotEmptyNote` | 2024-02-20 | Snapshot accepts empty note without error |
| `TestRegression_HistoryWithTags` | 2024-02-20 | History command properly filters by tag |
| `TestRegression_MultipleTags` | 2024-02-20 | Multiple --tag flags are all saved to snapshot |
| `TestRegression_RestoreHead` | 2024-02-20 | Restore HEAD returns to latest snapshot correctly |
| `TestRegression_WorktreeFork` | 2024-02-20 | Worktree fork properly sets up new worktree state |
| `TestRegression_GCWithEmptySnapshot` | 2024-02-20 | GC handles snapshots with empty payloads without panicking |
| `TestRegression_DoctorRuntimeRepair` | 2024-02-20, PR #7d0db0c | Doctor --repair-runtime executes repairs correctly |
| `TestRegression_InfoCommand` | 2024-02-20, PR #7d0db0c | Info command displays all required repository fields |

## Test Categories

### Snapshot Operations
- `TestRegression_SnapshotEmptyNote` - Empty note handling
- `TestRegression_MultipleTags` - Multiple tag preservation
- `TestRegression_GCWithEmptySnapshot` - Empty payload handling

### Restore Operations
- `TestRegression_RestoreNonExistentSnapshot` - Invalid snapshot ID handling
- `TestRegression_RestoreHead` - HEAD reference resolution

### History & Display
- `TestRegression_HistoryWithTags` - Tag filtering

### Worktree Management
- `TestRegression_WorktreeFork` - Fork state initialization

### Garbage Collection
- `TestRegression_GCWithEmptySnapshot` - Empty snapshot handling

### Doctor & Repair
- `TestRegression_DoctorRuntimeRepair` - Runtime repair execution

### Info & Display
- `TestRegression_InfoCommand` - Repository info display

## When NOT to Add a Regression Test

Regression tests are NOT for:

1. **New features**: Use the standard test suite
2. **Refactoring**: Tests should already exist; improve them instead
3. **Performance issues**: Use `internal/*/bench_test.go` instead
4. **Documentation bugs**: Fix the docs directly
5. **Trivial fixes**: Typos, obvious errors that tests wouldn't catch

## Best Practices

### DO:
- Make tests independent and isolated
- Use clear, descriptive test names
- Include the exact scenario that caused the bug
- Verify the fix prevents the bug
- Keep tests fast and focused
- Use existing helper functions (`runJVS`, `runJVSInRepo`, `createFiles`)

### DON'T:
- Don't mock at a level that hides the bug
- Don't test implementation details (test behavior)
- Don't make tests dependent on each other
- Don't add tests for unfixed bugs
- Don't remove regression tests without team discussion

## Review Process

Regression tests require:

1. **Issue/PR reference**: Must link to the original bug report
2. **Reviewer approval**: At least one maintainer must review
3. **Clear documentation**: Comment block must explain the bug
4. **Verification**: Test must fail on the old code and pass on the fix

## CII Silver Badge Compliance

This regression test suite helps JVS meet CNCF/CII Silver badge requirements:

- **Static Analysis**: Regression tests catch issues that static analysis might miss
- **Quality**: Demonstrates commitment to preventing bug recurrence
- **Documentation**: Each test documents a historical bug and its fix
- **Testing**: Expands test coverage with real-world scenarios

For more information on CII badge requirements, see:
https://github.com/coreinfrastructure/best-practices-badge/

## Template for New Tests

Copy this template when adding a new regression test:

```go
// TestRegression_YourBugDescription tests that [what it fixes].
//
// Bug: [Brief description of what was broken]
// Fixed: [YYYY-MM-DD], PR #[number]
// Issue: #[number]
func TestRegression_YourBugDescription(t *testing.T) {
    repoPath := initTestRepo(t)
    mainPath := filepath.Join(repoPath, "main")

    // Setup: Create the scenario that triggered the bug
    createFiles(t, mainPath, map[string]string{
        "file.txt": "content",
    })

    // Action: Perform the operation that was buggy
    stdout, stderr, code := runJVSInRepo(t, repoPath, "command", "args")

    // Assertion: Verify the fix works correctly
    assert.Equal(t, 0, code, "command should succeed")
    assert.NotContains(t, stderr, "error", "should not show errors")
}
```
