# Implementation Plan: Policy Profile Support

**Branch**: `007-policy-profile-support` | **Date**: 2026-06-01 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `.specify/specs/007-policy-profile-support/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.github/agents/speckit.plan.agent.md` for the execution workflow.

## Summary

Add policy profile support to gh-repo-health-report, enabling users to apply different health check expectations based on repository type (open-source, internal-service, application, archived, prototype). This feature allows organizations to tailor governance policies, eliminating false positives and making health scores contextually relevant. Implementation includes profile definitions with enforcement levels (required/recommended/ignored), CLI flags (--profile, --profile-config), auto-detection heuristics, profile-aware scoring, and output format changes to indicate skipped checks.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: github.com/cli/go-gh/v2 (GitHub CLI), github.com/spf13/cobra (CLI framework)  
**Storage**: N/A (stateless CLI tool), config file at `.gh-repo-health-report.yml` or `~/.gh-repo-health-report.yml`  
**Testing**: Go standard `testing` package — run with `go test ./...`  
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)  
**Project Type**: GitHub CLI extension  
**Performance Goals**: Profile evaluation adds negligible overhead (<1ms per repo), maintains sub-second response for small repos  
**Constraints**: GitHub API rate limits (5000 req/hr authenticated), must respect `gh` CLI auth; profile logic is client-side filtering with no additional API calls  
**Scale/Scope**: Designed for auditing 1-1000+ repositories per execution with profile-aware filtering  
**Config File Format**: YAML (primary) or JSON for profile configuration, optional, supports default_profile setting

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Review against `.specify/memory/constitution.md`:

- [x] **CLI Extension First**: Maintains `gh` CLI integration standards — no changes to extension model
- [x] **CLI Interface Pattern**: Follows single-command-with-flags pattern — adds `--profile` and `--profile-config` flags to existing root command
- [x] **Internal Package Architecture**: Respects package boundaries:
  - New profile logic in `internal/checks/profile.go` (profile definitions, enforcement evaluation)
  - Config loading in `internal/checks/config.go` (reads YAML/JSON config files)
  - CLI flag parsing in `cmd/gh-repo-health-report/main.go`
  - Output changes in `internal/formatter/formatter.go` (add skipped check indicators)
  - No changes to `internal/api` (no new API calls required)
- [x] **Testing Requirements**: Tests will be colocated (`profile_test.go`, `config_test.go`), table-driven, and mock-free (pure data structure tests)
- [x] **Dependency Management**: May require YAML parsing library (`gopkg.in/yaml.v3`), minimal and standard in Go ecosystem
- [x] **Build and Release**: Works with existing CI/CD workflows — no build process changes

**Violations**: None. Feature adds new internal package files and flags but maintains all constitutional boundaries.

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

```text
gh-repo-health-report/
├── cmd/gh-repo-health-report/
│   └── main.go                    # [MODIFY] Add --profile and --profile-config flags, integrate profile evaluation
│
├── internal/
│   ├── api/
│   │   ├── client.go              # [NO CHANGE] No new API calls needed
│   │   └── client_test.go         # [NO CHANGE]
│   │
│   ├── checks/
│   │   ├── checks.go              # [MODIFY] Update Evaluate() to accept profile, filter checks by enforcement level
│   │   ├── checks_test.go         # [MODIFY] Add tests for profile-aware evaluation
│   │   ├── profile.go             # [NEW] Profile type, enforcement levels, predefined profiles, auto-detection logic
│   │   ├── profile_test.go        # [NEW] Tests for profile definitions and auto-detection
│   │   ├── config.go              # [NEW] Config file loading (YAML/JSON), default profile resolution
│   │   └── config_test.go         # [NEW] Tests for config loading and precedence
│   │
│   └── formatter/
│       ├── formatter.go           # [MODIFY] Add skipped check indicators to all output formats
│       └── formatter_test.go      # [MODIFY] Add tests for skipped check display
│
├── .github/workflows/
│   ├── ci.yml                     # [NO CHANGE] Existing tests will validate backward compatibility
│   └── release.yml                # [NO CHANGE]
│
├── go.mod                         # [MODIFY] Add gopkg.in/yaml.v3 dependency
├── go.sum                         # [AUTO-UPDATE]
└── README.md                      # [MODIFY] Document --profile flag, profile types, config file
```

**Files This Feature Will Modify**:

- `cmd/gh-repo-health-report/main.go` — Add `--profile` and `--profile-config` flags, load config, pass profile to evaluation
- `internal/checks/checks.go` — Update `Evaluate()` to accept `Profile`, filter `FailedChecks` by enforcement level, add `SkippedChecks` field to `Result`
- `internal/checks/checks_test.go` — Add table-driven tests for profile-aware evaluation
- `internal/formatter/formatter.go` — Add `SkippedChecks` to output (table: `[IGNORED]` notation, JSON: `"skipped_checks"` field, CSV: column, markdown: similar to table)
- `internal/formatter/formatter_test.go` — Add tests for skipped check formatting
- `go.mod` — Add `gopkg.in/yaml.v3` dependency for config file parsing
- `README.md` — Document profile feature, CLI flags, config file format, profile definitions

**New Files This Feature Will Add**:

- `internal/checks/profile.go` — Profile type, EnforcementLevel enum (Required/Recommended/Ignored), predefined profiles (OpenSource, InternalService, Application, Archived, Prototype), auto-detection heuristics
- `internal/checks/profile_test.go` — Table-driven tests for profile definitions, auto-detection logic
- `internal/checks/config.go` — Config file loading (supports YAML/JSON), searches current directory and home directory for `.gh-repo-health-report.yml` or `.gh-repo-health-report.json`, returns default profile
- `internal/checks/config_test.go` — Tests for config loading, file precedence, default profile resolution

## Complexity Tracking

No constitution violations. This feature cleanly extends the existing architecture:

- Profile logic naturally belongs in `internal/checks` package (governance evaluation)
- Config loading is a pure utility (file I/O, no coupling to other packages)
- CLI changes are minimal flag additions (maintains single-command pattern)
- Output format changes are additive (preserve existing data, add skipped check indicators)
- No new dependencies beyond standard YAML parsing (common in Go ecosystem)
- Backward compatibility maintained (no profile = legacy behavior)

---

## Implementation Phases

### Phase 0: Research & Discovery ✅

**Status**: Complete  
**Artifact**: `research.md`

**Key Decisions**:
- Profile data structure: `map[string]EnforcementLevel` for O(1) lookup
- Config format: YAML (primary) with JSON fallback using `gopkg.in/yaml.v3`
- Auto-detection: Prioritize archived status > topics > visibility > fallback
- Scoring: Exclude ignored checks from pass/fail counts
- Backward compatibility: nil profile = legacy behavior (all checks required)
- Output indicators: Format-specific markers (`[SKIP]`, JSON fields, CSV columns)
- --fail-on interaction: "any" respects profile, specific checks override
- Profile maintenance: Package-level variables with clear documentation

---

### Phase 1: Design & Contracts ✅

**Status**: Complete  
**Artifacts**: `data-model.md`, `contracts/cli-contract.md`, `quickstart.md`

**Data Model Highlights**:
- `EnforcementLevel` enum: Required (0), Recommended (1), Ignored (2)
- `Profile` struct: Name, Description, Checks map
- Five predefined profiles: open-source, internal-service, application, archived, prototype
- Auto-detection heuristics with priority-based rules
- Extended `Result` struct with `SkippedChecks []SkippedCheck`
- Extended `Options` struct with `Profile *Profile`
- Config file: `default_profile` setting in YAML/JSON

**CLI Contract Highlights**:
- `--profile [name]` flag: Specifies profile to apply
- `--profile-config [path]` flag: Explicit config file path
- Modified `--fail-on` behavior: Respects profile for "any", overrides for specific checks
- Output format changes: All formats show skipped check indicators when profile active
- Help text updates: Document profiles, config file format, examples
- Backward compatibility: No profile = existing behavior

**Quickstart Highlights**:
- Implementation checklist for 7 phases
- Key functions and usage patterns
- Development tips and common pitfalls
- Estimated implementation time: 12-19 hours
- Success criteria and validation tests

---

### Phase 2: Task Planning (Next Step)

**Command**: `/speckit.tasks`  
**Expected Artifacts**: `tasks.md` with atomic implementation tasks

**This plan ends here.** The `/speckit.plan` command stops after Phase 1 (design). Use `/speckit.tasks` to generate implementation tasks based on this plan.

---

## Summary

**Feature**: Policy Profile Support for gh-repo-health-report

**What**: Add five predefined policy profiles (open-source, internal-service, application, archived, prototype) that define which of the 28 existing health checks are required, recommended, or ignored based on repository type.

**Why**: Eliminate false positives by tailoring governance expectations to repository context. A prototype shouldn't be penalized for lacking CODEOWNERS; an archived repo shouldn't fail for staleness.

**How**: 
1. **Profile System** (`internal/checks/profile.go`): Define `Profile` struct with `map[string]EnforcementLevel`, create five predefined profiles, implement auto-detection heuristics
2. **Config Support** (`internal/checks/config.go`): Load default profile from YAML/JSON config files (`.gh-repo-health-report.yml`)
3. **Evaluation** (`internal/checks/checks.go`): Update `Evaluate()` to filter checks by profile, add `SkippedChecks` to `Result`
4. **Output** (`internal/formatter/formatter.go`): Add skipped check indicators to all formats (table: `[SKIP]`, JSON: `skipped_checks` field, CSV: column, markdown: summary)
5. **CLI** (`cmd/gh-repo-health-report/main.go`): Add `--profile` and `--profile-config` flags, resolve profile precedence, update `--fail-on` logic

**Backward Compatibility**: No profile = legacy behavior (all 28 checks evaluated as required). Existing users see zero behavior change.

**Key Files Modified**:
- `cmd/gh-repo-health-report/main.go` — CLI flags, profile resolution
- `internal/checks/checks.go` — Profile-aware evaluation
- `internal/formatter/formatter.go` — Skipped check indicators
- `go.mod` — Add `gopkg.in/yaml.v3` dependency

**Key Files Added**:
- `internal/checks/profile.go` — Profile definitions, auto-detection
- `internal/checks/config.go` — Config file loading
- `internal/checks/profile_test.go` — Profile tests
- `internal/checks/config_test.go` — Config tests

**Testing Strategy**:
- Table-driven tests for profile definitions (validate all 28 checks present)
- Auto-detection tests (archived status, topics, visibility, fallback)
- Config loading tests (YAML/JSON, discovery order, error handling)
- Evaluation tests (nil profile, each profile type, skipped checks recorded)
- Output format tests (all 4 formats show skipped checks correctly)
- Backward compatibility tests (no profile = existing behavior)
- Integration tests (CLI flags, --fail-on interaction)

**Migration Path**:
1. Users without profiles: No action needed (backward compatible)
2. Users adopting profiles: Add `--profile` flag to invocations or create config file
3. Organizations: Create `.gh-repo-health-report.yml` with `default_profile: internal-service`
4. Diverse portfolios: Use `--profile auto` for automatic detection

**Estimated Effort**: 12-19 hours (medium complexity)

**Success Metrics**:
- All tests pass (`go test ./...`)
- Manual validation with 100+ repositories across profiles
- Zero breaking changes for existing users
- Profile-aware scoring correctly excludes ignored checks
- All output formats display skipped check indicators
- Auto-detection correctly infers profiles 85%+ of the time

---

## Branch and Artifacts

**Branch**: `007-policy-profile-support`  
**Spec**: `.specify/specs/007-policy-profile-support/spec.md`  
**Plan**: `.specify/specs/007-policy-profile-support/plan.md` (this file)  
**Research**: `.specify/specs/007-policy-profile-support/research.md`  
**Data Model**: `.specify/specs/007-policy-profile-support/data-model.md`  
**Contracts**: `.specify/specs/007-policy-profile-support/contracts/cli-contract.md`  
**Quickstart**: `.specify/specs/007-policy-profile-support/quickstart.md`

**Next Command**: `/speckit.tasks` to generate atomic implementation tasks
