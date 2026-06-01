# Implementation Plan: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE] | **Spec**: [link]
**Input**: Feature specification from `/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.github/agents/speckit.plan.agent.md` for the execution workflow.

## Summary

[Extract from feature spec: primary requirement + technical approach from research]

## Technical Context

<!--
  Pre-filled with gh-repo-health-report project specifics.
  Update only if this feature requires different technology or constraints.
-->

**Language/Version**: Go 1.21+  
**Primary Dependencies**: github.com/cli/go-gh/v2 (GitHub CLI), github.com/spf13/cobra (CLI framework)  
**Storage**: N/A (stateless CLI tool)  
**Testing**: Go standard `testing` package — run with `go test ./...`  
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)  
**Project Type**: GitHub CLI extension  
**Performance Goals**: Sub-second response for small repos (<100), reasonable performance for org scans (hundreds of repos)  
**Constraints**: GitHub API rate limits (5000 req/hr authenticated), must respect `gh` CLI auth  
**Scale/Scope**: Designed for auditing 1-1000+ repositories per execution

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Review against `.specify/memory/constitution.md`:

- [ ] **CLI Extension First**: Does this maintain `gh` CLI integration standards?
- [ ] **CLI Interface Pattern**: Does this follow single-command-with-flags pattern?
- [ ] **Internal Package Architecture**: Does this respect package boundaries?
  - `cmd/` for CLI only, `internal/api` for GitHub API, `internal/checks` for evaluation, `internal/formatter` for output
- [ ] **Testing Requirements**: Are tests colocated, table-driven, and mock the GitHub API?
- [ ] **Dependency Management**: Are new dependencies necessary and minimal?
- [ ] **Build and Release**: Does this work with existing CI/CD workflows?

**Violations** (if any — justify in Complexity Tracking section):

[List any constitution violations and explain why they're necessary for this feature]

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

<!--
  Pre-filled with gh-repo-health-report structure.
  Update this section to show which files this feature will modify or add.
-->

```text
gh-repo-health-report/
├── cmd/gh-repo-health-report/
│   └── main.go                    # [Modify if adding CLI flags or commands]
│
├── internal/
│   ├── api/
│   │   ├── client.go              # [Modify if adding GitHub API calls]
│   │   └── client_test.go         # [Add tests for new API functions]
│   │
│   ├── checks/
│   │   ├── checks.go              # [Modify if adding new health checks]
│   │   └── checks_test.go         # [Add tests for new check logic]
│   │
│   └── formatter/
│       ├── formatter.go           # [Modify if changing output formats]
│       └── formatter_test.go      # [Add tests for formatting changes]
│
├── .github/workflows/
│   ├── ci.yml                     # [Modify only if CI process changes]
│   └── release.yml                # [Modify only if release process changes]
│
├── go.mod                         # [Modify if adding dependencies]
├── go.sum                         # [Auto-updated with go.mod]
└── README.md                      # [Update if adding user-facing features]
```

**Files This Feature Will Modify**:

- List specific files that will be changed
- Example: `internal/checks/checks.go` — add new check constants and evaluation functions
- Example: `cmd/gh-repo-health-report/main.go` — add `--new-flag` flag
- Example: `internal/api/client.go` — add function to fetch new data from GitHub API

**New Files This Feature Will Add** (if any):

- List new files being created
- Example: `internal/checks/advanced_checks.go` — if feature requires new package file

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
