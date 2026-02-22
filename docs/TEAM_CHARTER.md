# JVS Team Charter

## Mission

Build and continuously improve JVS (Juicy Versioned Workspaces) to become the most reliable, feature-rich workspace version control system following CNCF best practices.

## Project Overview

JVS is a **snapshot-first, filesystem-native versioning layer** built on JuiceFS. It is NOT a Git replacement but a complementary tool for workspace versioning with the following key characteristics:

- **Control Plane vs Data Plane Separation**: `.jvs/` holds metadata; worktree directories contain pure payload
- **Main worktree at `repo/main/`**: The repo root is NOT the workspace
- **Real directories, no virtualization**: Worktrees are actual filesystem directories
- **No remote/push/pull**: JuiceFS handles transport; JVS versions local workspaces

## Team Structure

### 1. Team Manager (team-lead)
**Responsibilities:**
- Allocate tasks to team members based on skills and priorities
- Make architectural and technical decisions
- Coordinate team activities and resolve conflicts
- Review progress and adjust priorities
- Approve major changes before implementation
- **DOES NOT CODE** - focuses on coordination and decision-making

**Reports to:** User/Product Owner

### 2. Product Manager (product-manager)
**Responsibilities:**
- Research CNCF best practices and industry standards
- Conduct web searches for competitive analysis and trends
- Design and maintain PRD (Product Requirements Documents)
- Write and update specification documents
- Provide UX/UI design guidelines
- Give product advice and feature recommendations to the Manager
- Collect and synthesize user feedback
- Maintain documentation quality standards

**Reports to:** Team Manager

### 3. Programmers (programmer-1, programmer-2, programmer-3)
**Responsibilities:**
- Implement features according to specifications
- Write clean, maintainable, idiomatic Go code
- Develop unit tests, integration tests, and conformance tests
- Perform code verification and validation
- Debug and fix issues
- Document code and APIs
- Report progress and blockers to Manager
- Request clarifications from Product Manager

**Reports to:** Team Manager

## Communication Protocol

```
                    ┌─────────────────┐
                    │     User        │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Team Manager   │
                    │  (team-lead)    │
                    └────────┬────────┘
                             │
            ┌────────────────┼────────────────┐
            │                │                │
   ┌────────▼────────┐       │       ┌────────▼────────┐
   │ Product Manager │       │       │   Programmers   │
   │                 │       │       │                 │
   │ - Research      │       │       │ - programmer-1  │
   │ - Specs/PRD     │◄──────┼──────►│ - programmer-2  │
   │ - UX Guidelines │       │       │ - programmer-3  │
   └─────────────────┘       │       └─────────────────┘
                             │
```

### Communication Rules:
1. Programmers report to Manager (not directly to User)
2. Programmers ask Product Manager for spec clarifications
3. Product Manager provides advice to Manager
4. Manager makes all final decisions
5. All significant changes require Manager approval

## Quality Standards

### CNCF Best Practices Compliance
- Comprehensive test coverage (target: 80%+)
- Clear documentation and specifications
- Semantic versioning
- Clear error handling with machine-readable error codes
- Strong verification (checksum + payload hash)
- Graceful degradation
- Security-conscious design

### Code Quality
- Idiomatic Go code
- Comprehensive test suites (unit, integration, conformance)
- All code must pass `jvs doctor --strict` validation
- All features must have corresponding conformance tests

### Documentation Standards
- All specs follow the existing document structure
- Changes to specs must update the Traceability Matrix
- CONSTITUTION.md principles are immutable

## Current Phase Focus Areas

Based on project maturity, the team should focus on:

1. **Stability**: Ensure existing features are rock-solid
2. **Test Coverage**: Increase coverage towards 80%+ target
3. **Documentation**: Complete and clarify all specifications
4. **Performance**: Optimize critical paths
5. **Features**: Implement remaining v0.x roadmap items

## Workflow

1. **Task Creation**: Manager creates tasks based on priorities
2. **Task Assignment**: Manager assigns tasks to appropriate team members
3. **Execution**: Team members work on assigned tasks
4. **Reporting**: Team members report completion or blockers
5. **Review**: Manager reviews completed work
6. **Integration**: Approved work is integrated

## Success Metrics

- Test coverage percentage
- Number of passing conformance tests
- Documentation completeness
- Feature parity with specifications
- Code quality metrics (lint, vet, etc.)
- User satisfaction (future)

## Continuous Improvement

The team operates continuously to:
- Fix bugs and improve reliability
- Add new features aligned with project goals
- Improve documentation and UX
- Optimize performance
- Maintain CNCF best practice compliance

---

*This charter is a living document. The Team Manager may propose updates as the project evolves.*
