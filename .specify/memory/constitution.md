# gh-repo-health-report Constitution

## Core Principles

### I. GitHub CLI Extension First

This project is a **GitHub CLI extension** built with Go:
- Must integrate with `gh` CLI using `github.com/cli/go-gh/v2`
- Must follow GitHub CLI extension naming: `gh repo-health-report` (executable: `gh-repo-health-report`)
- Must respect GitHub CLI authentication and API access patterns
- Extension should feel native to the `gh` CLI experience

### II. CLI Interface Pattern

Command-line interface follows these rules:
- **Single command with flags** — all functionality via root command flags, not subcommands
- **Flexible input** — support `--org`, `--owner`, or `--repo` for different audit scopes
- **Multiple output formats** — table (default), JSON, CSV, markdown via `--format`
- **Actionable exit codes** — exit 1 on `--fail-on` check failures for CI integration
- **Configurable thresholds** — `--since`, `--max-branches`, `--max-tags` allow customization

### III. Internal Package Architecture

Code organization follows Go best practices:
- **`cmd/gh-repo-health-report/`** — CLI entry point, cobra setup, flag parsing
- **`internal/api/`** — GitHub API client, repository fetching, GraphQL queries
- **`internal/checks/`** — Health check logic, check name constants, evaluation functions
- **`internal/formatter/`** — Output formatting (table, JSON, CSV, markdown)
- **No external imports** — `internal/` packages must not be imported outside this module

### IV. Testing Requirements

Testing follows Go conventions:
- **Test files colocated** — `*_test.go` files live alongside source files
- **Table-driven tests** — use test tables for multiple input scenarios
- **Standard testing** — use `testing` package, run with `go test ./...`
- **Package-level tests** — test files in same package (`package api`, `package checks`, etc.)
- **Mock GitHub API** — tests must not hit real GitHub API; use fixtures or mocks

### V. Dependency Management

Dependencies are minimal and purposeful:
- **Core dependencies only**:
  - `github.com/cli/go-gh/v2` — GitHub CLI integration
  - `github.com/spf13/cobra` — CLI framework
- **Go modules** — use `go.mod`/`go.sum` for dependency tracking
- **Go 1.21+** — minimum Go version requirement
- **No external dependencies** for core logic — avoid unnecessary third-party packages

### VI. Build and Release

GitHub Actions workflows handle CI/CD:
- **`.github/workflows/ci.yml`** — runs tests, linting, builds on push/PR
- **`.github/workflows/release.yml`** — builds and publishes releases
- **No Makefile** — use `go build`, `go test`, `go run` directly
- **Binary output** — builds to `gh-repo-health-report` executable at repository root

## Code Boundaries

### Directory Structure

```
gh-repo-health-report/
├── cmd/gh-repo-health-report/    # CLI entry point
│   └── main.go                    # Root command, flag parsing, execution flow
├── internal/
│   ├── api/                       # GitHub API client
│   │   ├── client.go              # Repository fetching, GraphQL queries
│   │   └── client_test.go
│   ├── checks/                    # Health check evaluation
│   │   ├── checks.go              # Check constants, evaluation logic
│   │   └── checks_test.go
│   └── formatter/                 # Output formatting
│       ├── formatter.go           # Table, JSON, CSV, markdown formatters
│       └── formatter_test.go
├── .github/workflows/             # CI/CD automation
├── go.mod                         # Go module definition
└── gh-repo-health-report          # Compiled binary (gitignored)
```

### Module Responsibilities

- **`cmd/gh-repo-health-report`** — CLI interface only; no business logic
- **`internal/api`** — GitHub API interactions only; no check evaluation
- **`internal/checks`** — Check evaluation only; no API calls or formatting
- **`internal/formatter`** — Output formatting only; no check logic

### Allowed Dependencies

- `cmd/gh-repo-health-report` → `internal/api`, `internal/checks`, `internal/formatter`
- `internal/*` packages → **no cross-package dependencies** within `internal/`
- All packages → standard library, `cobra`, `go-gh`

## Naming Conventions

### File Naming

- **Package files** — lowercase, descriptive: `client.go`, `checks.go`, `formatter.go`
- **Test files** — `*_test.go` suffix, same base name: `client_test.go`
- **No underscores** — prefer short package-level files over `internal_checks_something.go`

### Variable Naming

- **Go conventions** — camelCase for unexported, PascalCase for exported
- **Short names** — `repo` not `repository`, `org` not `organization` in local scope
- **Descriptive names** — exported functions use full words: `NewClient`, `EvaluateChecks`

### Constant Naming

- **Check name constants** — `Check` prefix: `CheckMissingReadme`, `CheckStale`
- **All caps for exported constants** — `DefaultMaxBranches = 50`
- **String values kebab-case** — `"missing-readme"`, `"has-stale-branches"`

### Git Branch Naming

- **Feature branches** — `feature/description` or `###-feature-name` (spec-kit convention)
- **Bug fixes** — `fix/description`
- **Default branch** — `main`

## Quality Gates

### Before Commit

- **Build succeeds** — `go build ./...` must complete without errors
- **Tests pass** — `go test ./...` must pass with no failures
- **No warnings** — `go vet ./...` must report no issues
- **Formatted** — code must be formatted with `gofmt` or `goimports`

### CI Checks

All PRs must pass:
- Go build for multiple platforms
- Test suite execution
- Linting and formatting checks
- No security vulnerabilities in dependencies

### Code Quality

- **Error handling** — all errors must be checked and handled appropriately
- **Package documentation** — exported functions, types, constants must have doc comments
- **Minimal complexity** — avoid deeply nested logic; extract functions when needed

## Feature Development Workflow

### 1. Specification Phase

- Create feature spec in `specs/###-feature-name/spec.md`
- Define user stories with acceptance criteria
- Identify which internal packages are affected

### 2. Design Phase

- Create implementation plan in `specs/###-feature-name/plan.md`
- Map feature to existing package structure
- Identify new types, functions, constants needed
- Define test strategy

### 3. Implementation Phase

- Create feature branch: `###-feature-name`
- Write tests first (TDD recommended but not mandatory)
- Implement in appropriate package(s):
  - New checks → `internal/checks/checks.go`
  - New API calls → `internal/api/client.go`
  - New output formats → `internal/formatter/formatter.go`
  - New CLI flags → `cmd/gh-repo-health-report/main.go`
- Run tests: `go test ./...`
- Build and manual test: `go build && ./gh-repo-health-report --help`

### 4. Validation Phase

- Verify all quality gates pass
- Test CLI manually with different flag combinations
- Verify backward compatibility (existing flags still work)
- Update README.md if adding user-facing features

## Governance

### Constitution Authority

- This constitution supersedes informal practices
- All PRs must comply with these principles
- Feature specs must reference relevant constitution sections
- Complexity that violates these rules must be explicitly justified in plan.md

### Amendment Process

- Constitution changes require explicit documentation
- Changes must be approved before implementation
- Migration plan required if changes affect existing code

### Development Guidance

- For new features, start with `/speckit.specify` to create spec
- Use `/speckit.plan` to create implementation plan
- Reference this constitution in all planning documents
- When in doubt, ask: "Does this align with the constitution?"

**Version**: 1.0.0 | **Ratified**: 2026-06-01 | **Last Amended**: 2026-06-01
