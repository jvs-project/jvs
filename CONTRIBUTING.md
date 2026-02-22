# Contributing to JVS

Thank you for your interest in contributing to JVS (Juicy Versioned Workspaces)!

## Quick Start

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/jvs.git`
3. Create a branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `make verify`
6. Commit: `git commit -m "Add some feature"`
7. Push: `git push origin feature/your-feature-name`
8. Open a Pull Request

## Development Environment

### Prerequisites

- **Go**: Version 1.25.6 or later
- **Operating System**: Linux, macOS, or Windows (with WSL2)
- **Storage**: JuiceFS mount (optional but recommended for O(1) snapshots)

### Building

```bash
# Build the jvs binary
make build

# The binary will be output to bin/jvs
./bin/jvs --help
```

### Running Tests

```bash
# Run unit tests
make test

# Run conformance tests (required before merging)
make conformance

# Run linters
make lint

# Run all verification (test + lint)
make verify
```

**All PRs must pass `make verify` before being merged.**

### Test Coverage Requirements

JVS maintains a target of **80%+ test coverage** for production readiness.

- Overall coverage: **77.6%** (as of v7.0)
- Critical paths must have higher coverage
- New features must include tests
- Use `go test -cover ./...` to check coverage

## Code Style Guidelines

### Go Conventions

JVS follows standard Go conventions:

1. **Effective Go**: Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
2. **gofmt**: All code must be formatted with `gofmt -s -w`
3. **golint**: Use `golangci-lint run` to catch issues
4. **Package names**: Short, lowercase, single words when possible
5. **Error handling**: Never ignore errors, use `errclass` for user-facing errors

### Error Class Usage

JVS uses stable error classes for user-facing errors:

```go
// Import the errclass package
import "github.com/jvs-project/jvs/pkg/errclass"

// Use predefined error classes
return errclass.ErrNameInvalid.WithMessage("worktree name cannot be empty")

// For internal errors, wrap with context
return fmt.Errorf("failed to read descriptor: %w", err)
```

**Available error classes** (from `pkg/errclass/errors.go`):
- `ErrNameInvalid` - Invalid name format
- `ErrPathEscape` - Path traversal attempt
- `ErrDescriptorCorrupt` - Descriptor checksum failed
- `ErrPayloadHashMismatch` - Payload hash verification failed
- `ErrLineageBroken` - Snapshot lineage inconsistency
- `ErrPartialSnapshot` - Incomplete snapshot detected
- `ErrGCPlanMismatch` - GC plan ID mismatch
- `ErrFormatUnsupported` - Format version not supported
- `ErrAuditChainBroken` - Audit hash chain validation failed

### Comment Guidelines

- **Public functions**: Must have godoc comments
- **Exported types**: Must have documentation
- **Complex logic**: Add explanatory comments
- **TODOs**: Use `// TODO:` for future work

## Commit Message Conventions

JVS follows a structured commit message format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test changes (adding/modifying tests)
- `refactor`: Code refactoring (no behavior change)
- `spec`: Specification document changes
- `chore`: Maintenance tasks
- `perf`: Performance improvements

### Examples

```
feat(snapshot): add --tag flag for snapshot tagging

Users can now attach tags during snapshot creation:
  jvs snapshot "initial setup" --tag v1.0 --tag stable

Tags are stored in the snapshot descriptor and can be used
for filtering in jvs history --tag <tag>.

Fixes #123
```

```
fix(restore): prevent snapshot creation in detached state

Previously, users could create snapshots while in detached state,
leading to unclear lineage. Now snapshot command returns an
error when worktree is detached.

Users must run `jvs restore HEAD` or `jvs worktree fork` first.

Closes #145
```

## Pull Request Process

### Before Opening a PR

1. **Search existing PRs** to avoid duplicates
2. **Discuss large changes** via issue first
3. **Update specs** if changing behavior (docs/*_SPEC.md)
4. **Add tests** for new functionality
5. **Update CHANGELOG** for user-visible changes
6. **Run `make verify`** and fix any issues

### PR Description Template

```markdown
## Summary
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Conformance tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added to complex code
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Specs updated if applicable
- [ ] CHANGELOG.md updated
```

### Review Process

1. **Automated checks**: CI runs `make verify`
2. **Maintainer review**: At least one maintainer must approve
3. **Conformance tests**: All 29 tests must pass
4. **Spec alignment**: Changes must align with `docs/CONSTITUTION.md`

## Project Structure

```
jvs/
├── cmd/jvs/           # Main CLI entry point
├── internal/          # Private implementation
│   ├── audit/         # Audit logging
│   ├── cli/           # CLI command handlers
│   ├── doctor/        # Repository health checks
│   ├── engine/        # Snapshot engine abstraction
│   ├── gc/            # Garbage collection
│   ├── integrity/     # Checksum and hash verification
│   ├── repo/          # Repository management
│   ├── restore/       # Restore operations
│   ├── snapshot/      # Snapshot creation
│   ├── verify/        # Verification commands
│   └── worktree/      # Worktree management
├── pkg/               # Public libraries
│   ├── config/        # Configuration
│   ├── errclass/      # Stable error classes
│   ├── fsutil/        # Filesystem utilities
│   ├── jsonutil/      # JSON handling
│   ├── logging/       # Logging utilities
│   ├── model/         # Data models
│   ├── pathutil/      # Path utilities
│   ├── progress/      # Progress reporting
│   └── uuidutil/      # UUID generation
├── test/conformance/  # Conformance tests (29 mandatory)
├── docs/              # Specification documents
└── Makefile           # Build automation
```

## Specification Documents

Before modifying behavior, review the relevant spec:

| Document | Purpose |
|----------|---------|
| `CONSTITUTION.md` | Core principles and design governance |
| `00_OVERVIEW.md` | Frozen design decisions |
| `01_REPO_LAYOUT_SPEC.md` | On-disk structure |
| `02_CLI_SPEC.md` | Command contract and error classes |
| `03_WORKTREE_SPEC.md` | Worktree lifecycle |
| `04_SNAPSHOT_SCOPE_AND_LINEAGE_SPEC.md` | Snapshot identity |
| `05_SNAPSHOT_ENGINE_SPEC.md` | Engine selection (juicefs-clone/reflink/copy) |
| `06_RESTORE_SPEC.md` | Restore and detached state |
| `11_CONFORMANCE_TEST_PLAN.md` | Mandatory test requirements |

## Questions?

- **GitHub Issues**: Use [Issues](https://github.com/jvs-project/jvs/issues) for bugs and feature requests
- **Discussions**: Use [Discussions](https://github.com/jvs-project/jvs/discussions) for questions and ideas
- **Team Charter**: See [docs/TEAM_CHARTER.md](docs/TEAM_CHARTER.md) for team structure

## License

By contributing to JVS, you agree that your contributions will be licensed under the [MIT License](LICENSE).

---

Thank you for contributing to JVS!
