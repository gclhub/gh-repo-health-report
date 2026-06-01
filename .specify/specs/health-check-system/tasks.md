# Tasks: Health Check System

**Status**: migrated (all tasks completed)  
**Input**: Reverse-engineered from `internal/checks/checks.go` and `internal/checks/checks_test.go`  
**Implementation**: Existing feature — all tasks marked as completed

## Format: `[ID] [P?] [Category] Description`

- **[x]**: Task completed (all tasks pre-existing)
- **[P]**: Could have run in parallel (different files, no dependencies)
- **[Category]**: Which aspect this task belongs to

---

## Phase 1: Data Structure Foundation

**Purpose**: Define core types for health check evaluation

- [x] T001 [P] [Foundation] Define `Options` struct in `internal/checks/checks.go` with `Since`, `MaxBranches`, `MaxTags` fields
- [x] T002 [P] [Foundation] Define `Result` struct in `internal/checks/checks.go` with 28+ boolean check fields
- [x] T003 [P] [Foundation] Add security tristate fields to `Result`: `VulnerabilityAlertsEnabled/Unknown`, `SecretScanningEnabled/Unknown`, `PushProtectionEnabled/Unknown`
- [x] T004 [P] [Foundation] Add count fields to `Result`: `BranchCount`, `StaleBranchCount`, `TagCount`, `TopicsCount`, `OpenIssueCount`, `SizeKB`
- [x] T005 [P] [Foundation] Add `FailedChecks []string` field to `Result` for tracking failed check names

**Checkpoint**: Data structures complete — evaluation logic can now be implemented

---

## Phase 2: Check Name Constants

**Purpose**: Define string constants for all 25+ check types

- [x] T006 [P] [Metadata] Add constants for metadata checks: `CheckHasDescription`, `CheckHasHomepage`, `CheckStale`
- [x] T007 [P] [Community] Add constants for community file checks: `CheckMissingReadme`, `CheckMissingLicense`, `CheckMissingCodeOfConduct`, `CheckMissingCodeowners`, `CheckMissingSecurityMd`, `CheckMissingContributing`, `CheckMissingIssueTemplates`, `CheckMissingPRTemplate`
- [x] T008 [P] [Features] Add constants for feature checks: `CheckHasIssues`, `CheckHasProjects`, `CheckHasWiki`
- [x] T009 [P] [Security] Add constants for security/automation checks: `CheckMissingDependabot`, `CheckMissingCI`, `CheckNoBranchProtection`, `CheckNoRulesets`, `CheckNoVulnerabilityAlerts`, `CheckNoSecretScanning`, `CheckNoPushProtection`, `CheckNoDeleteBranchOnMerge`
- [x] T010 [P] [Branches] Add constants for branch/tag checks: `CheckTooManyBranches`, `CheckHasStaleBranches`, `CheckTooManyTags`
- [x] T011 [P] [Defaults] Add default threshold constants: `DefaultMaxBranches = 50`, `DefaultMaxTags = 100`

**Checkpoint**: All check names defined — evaluation function can reference them

---

## Phase 3: Core Evaluation Function

**Purpose**: Implement `Evaluate()` function with all check logic

### User Story 1: Community File Detection (P1) 🎯

- [x] T012 [Metadata] Implement `Evaluate(repo *api.Repository, opts Options) *Result` function signature in `internal/checks/checks.go`
- [x] T013 [Metadata] Initialize `Result` with repository reference and metadata fields (description, homepage, topics)
- [x] T014 [Community] Map community file boolean fields from `api.Repository` to `Result` (HasReadme, HasLicense, HasCodeOfConduct, HasCodeowners, HasSecurity, HasContributing, HasIssueTemplates, HasPRTemplate)
- [x] T015 [Community] Add failed check collection for missing community files (8 checks)

**Checkpoint**: Community file checks working — can test against repos with/without standard files

### User Story 2: Repository Metadata Assessment (P1) 🎯

- [x] T016 [Metadata] Implement staleness evaluation: check if `time.Since(repo.PushedAt) > opts.Since`, use 180 days default
- [x] T017 [Metadata] Add failed check collection for metadata: `has-description`, `has-homepage`, `stale`
- [x] T018 [Features] Map GitHub feature flags from `api.Repository` (HasIssuesEnabled, HasProjectsEnabled, HasWikiEnabled)
- [x] T019 [Features] Add failed check collection for disabled features (3 checks)

**Checkpoint**: Metadata and feature checks working — can test staleness and feature detection

### User Story 3: Security & Automation Checks (P2)

- [x] T020 [Security] Map extended check fields from `api.Repository`: `HasDependabot`, `HasCIWorkflows`, `DefaultBranchProtected`, `HasRulesets`, `DeleteBranchOnMerge`
- [x] T021 [Security] Map security tristate fields from `api.Repository`: vulnerability alerts, secret scanning, push protection (both Enabled and Unknown flags)
- [x] T022 [Security] Implement tristate failed check logic: only add to FailedChecks when `!Unknown && !Enabled`
- [x] T023 [Security] Add failed check collection for security/automation (8 checks with tristate support)

**Checkpoint**: Security checks working with graceful permission handling — shows "?" when access denied

### User Story 4: Branch & Tag Management (P2)

- [x] T024 [Branches] Map branch/tag counts from `api.Repository`: `BranchCount`, `StaleBranchCount`, `TagCount`
- [x] T025 [Branches] Implement threshold logic with defaults: `maxBranches := opts.MaxBranches; if maxBranches == 0 { maxBranches = DefaultMaxBranches }`
- [x] T026 [Branches] Add failed check collection for branch/tag thresholds: `too-many-branches`, `has-stale-branches`, `too-many-tags`

**Checkpoint**: Branch and tag checks working — can test with custom thresholds

---

## Phase 4: Test Suite

**Purpose**: Comprehensive test coverage for all check scenarios

### Test Infrastructure

- [x] T027 [P] [Tests] Create `baseRepo()` helper in `internal/checks/checks_test.go` returning healthy repository
- [x] T028 [P] [Tests] Create `contains()` helper for checking FailedChecks slice contents

### Test Cases (Table-Driven and Individual)

- [x] T029 [P] [Tests] Add `TestEvaluate_Healthy` — verify no failures for fully healthy repo
- [x] T030 [P] [Tests] Add `TestEvaluate_Stale` — verify staleness detection for old repos
- [x] T031 [P] [Tests] Add `TestEvaluate_NotStale` — verify recent repos pass staleness check
- [x] T032 [P] [Tests] Add `TestEvaluate_MissingFiles` — verify all community file checks can fail
- [x] T033 [P] [Tests] Add test for disabled GitHub features (issues, projects, wiki)
- [x] T034 [P] [Tests] Add test for security/automation checks (dependabot, CI, branch protection, rulesets)
- [x] T035 [P] [Tests] Add test for security tristate behavior (unknown status handling)
- [x] T036 [P] [Tests] Add test for branch/tag threshold logic with custom values
- [x] T037 [P] [Tests] Add test for threshold disabled behavior (MaxBranches = 0, MaxTags = 0)
- [x] T038 [P] [Tests] Add test for stale branch detection

**Checkpoint**: Full test coverage achieved — all check scenarios verified

---

## Phase 5: Integration & Documentation

**Purpose**: Ensure checks integrate with rest of system

- [x] T039 [Integration] Verify `Evaluate()` called from `cmd/gh-repo-health-report/main.go` with Options from CLI flags
- [x] T040 [Integration] Verify `Result` passed to `formatter.Format()` for output generation
- [x] T041 [Integration] Verify `FailedChecks` used in `shouldFail()` logic for `--fail-on` flag
- [x] T042 [P] [Documentation] Document check names and evaluation logic in package comments
- [x] T043 [P] [Documentation] Add examples of check behavior in README.md check names table

**Checkpoint**: Health check system fully integrated and documented

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Data Structures)**: No dependencies - foundation for everything
- **Phase 2 (Constants)**: Can run in parallel with Phase 1
- **Phase 3 (Evaluation)**: Depends on Phase 1 + 2 completion
- **Phase 4 (Tests)**: Depends on Phase 3 completion
- **Phase 5 (Integration)**: Depends on Phase 3 completion (can run in parallel with Phase 4)

### Within Phase 3 (User Stories)

All user stories implemented in single `Evaluate()` function, but logical grouping:
1. Foundation (T012-T013) — Function structure
2. Community files (T014-T015) — P1
3. Metadata (T016-T019) — P1
4. Security (T020-T023) — P2
5. Branches (T024-T026) — P2

### Parallel Opportunities

- Phase 1: All tasks marked [P] (T001-T005) — different struct fields
- Phase 2: All tasks marked [P] (T006-T011) — different constant groups
- Phase 4: All test tasks marked [P] (T029-T038) — independent test cases
- Phase 5: Documentation tasks marked [P] (T042-T043) — different files

---

## Implementation Notes

### Actual Development Pattern

**Real-world implementation** likely followed this order:
1. Basic structure: `Evaluate()` function with `Result` initialization
2. Iterative addition of check types as requirements emerged
3. Tests added alongside each check category
4. Threshold configuration added when users requested customization
5. Tristate security logic added when permission issues discovered

### Key Design Decisions Made

**Decision 1**: Single evaluation function instead of separate check functions
- **Rationale**: Simpler, all data available, no coordination overhead
- **Trade-off**: Large function but straightforward logic

**Decision 2**: String-based check names instead of enums
- **Rationale**: CLI `--fail-on` flag matching, human-readable output
- **Trade-off**: No compile-time checking but more flexible

**Decision 3**: Pre-populated repository data from API client
- **Rationale**: Separation of concerns, evaluation is pure logic
- **Trade-off**: Tight coupling to `api.Repository` structure

### Testing Philosophy

**Table-driven tests** for related scenarios, **individual tests** for specific edge cases:
- Healthy baseline ensures all checks can pass
- Individual failure tests ensure each check can fail independently
- Threshold tests verify configurable behavior
- Tristate tests verify graceful permission handling

### Gaps Identified During Migration

1. ⚠️ **No test for zero-time PushedAt** (empty repositories) — edge case not explicitly covered
2. ⚠️ **No benchmark tests** for evaluation performance — scalability not measured
3. ⚠️ **Limited documentation** on tristate logic — could be clearer for future maintainers

---

## Total Implementation Effort

**Completed Tasks**: 43 tasks across 5 phases  
**Files Modified**: 2 files (`checks.go`, `checks_test.go`)  
**Lines of Code**: 573 total (230 implementation + 343 tests)  
**Test Coverage**: Comprehensive (healthy baseline + individual check failures + edge cases)  
**Check Types**: 25+ distinct health checks implemented

---

## Maintenance Guidance

**Adding New Check**:
1. Add constant (Phase 2 pattern)
2. Add field to `Result` struct (Phase 1 pattern)
3. Add evaluation logic in `Evaluate()` (Phase 3 pattern)
4. Add failure collection (Phase 3 pattern)
5. Add test case (Phase 4 pattern)
6. Update formatter to display new check
7. Document in README check names table

**Modifying Threshold**:
- Update `DefaultMaxBranches` or `DefaultMaxTags` constant
- No other code changes needed (defaults auto-apply)

**Changing Tristate Behavior**:
- Coordinate with formatter package (affects display)
- Update failure collection logic
- Add/update tests for new behavior
