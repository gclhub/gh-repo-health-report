# Tasks: GitHub API Client

**Status**: migrated (all tasks completed)  
**Input**: Reverse-engineered from `internal/api/client.go` and `internal/api/client_test.go`  
**Implementation**: Existing feature — all tasks marked as completed

## Format: `[ID] [P?] [Category] Description`

- **[x]**: Task completed (all tasks pre-existing)
- **[P]**: Could have run in parallel (different files, no dependencies)
- **[Category]**: Which API category this task belongs to

---

## Phase 1: Client Foundation

**Purpose**: Initialize API client with authentication

### User Story 6: Client Initialization (P1) 🎯

- [x] T001 [P] [Foundation] Define `Client` struct in `internal/api/client.go` wrapping `*api.RESTClient`
- [x] T002 [P] [Foundation] Implement `NewClient()` function using `api.DefaultRESTClient()` from go-gh
- [x] T003 [P] [Foundation] Implement `NewClientFromREST(rest *api.RESTClient)` factory for testing
- [x] T004 [P] [Foundation] Add error handling for failed client initialization (auth issues)

**Checkpoint**: Client initialization working — can create authenticated API client

---

## Phase 2: Repository Data Model

**Purpose**: Define comprehensive repository structure

- [x] T005 [P] [Model] Define `Repository` struct with GitHub API standard fields (full_name, name, owner, description, homepage, topics, pushed_at)
- [x] T006 [P] [Model] Add feature flag fields to `Repository` (has_issues, has_projects, has_wiki)
- [x] T007 [P] [Model] Add metadata fields (open_issues_count, size, delete_branch_on_merge)
- [x] T008 [P] [Model] Add `SecurityAndAnalysis` nested struct with secret scanning and push protection fields
- [x] T009 [P] [Model] Add populated fields for file checks (HasReadme, HasLicense, HasCodeOfConduct, HasCodeowners, HasSecurity, HasContributing, HasIssueTemplates, HasPRTemplate)
- [x] T010 [P] [Model] Add populated fields for extended checks (HasDependabot, HasCIWorkflows, DefaultBranchProtected, HasRulesets)
- [x] T011 [P] [Model] Add tristate security fields (VulnerabilityAlertsEnabled/Unknown, SecretScanningEnabled/Unknown, PushProtectionEnabled/Unknown)
- [x] T012 [P] [Model] Add branch/tag count fields (BranchCount, StaleBranchCount, TagCount)
- [x] T013 [P] [Model] Define `branchItem` struct for branch list parsing
- [x] T014 [P] [Model] Define `securityFeatureStatus` struct for security settings parsing

**Checkpoint**: Repository data model complete — all API response fields and computed fields defined

---

## Phase 3: Repository Fetching

**Purpose**: Implement methods to fetch repository data from GitHub

### User Story 1: Repository Fetching (P1) 🎯

- [x] T015 [P] [Fetch] Implement `GetRepo(owner, name string) (*Repository, error)` method
- [x] T016 [Fetch] Implement `listRepos(basePath string, includeForks, includeArchived bool)` with pagination loop
- [x] T017 [Fetch] Add fork filtering logic in `listRepos()` (skip if `Fork && !includeForks`)
- [x] T018 [Fetch] Add archive filtering logic in `listRepos()` (skip if `Archived && !includeArchived`)
- [x] T019 [Fetch] Add pagination continuation logic (loop while `len(pageRepos) == 100`)
- [x] T020 [P] [Fetch] Implement `ListOrgRepos(org string, includeForks, includeArchived bool)` calling `listRepos()` with org path
- [x] T021 [P] [Fetch] Implement `ListUserRepos(user string, includeForks, includeArchived bool)` calling `listRepos()` with user path

**Checkpoint**: Repository fetching working — can fetch single repos, org repos, user repos with pagination

---

## Phase 4: File Existence Checks

**Purpose**: Detect community files in repositories

### User Story 2: Community File Detection (P1) 🎯

- [x] T022 [P] [Files] Implement `CheckFileExists(owner, repo, path string)` helper with 404 handling
- [x] T023 [P] [Files] Add `isNotFound(err error)` helper checking for HTTP 404 status
- [x] T024 [Files] Implement `PopulateFileChecks(repo *Repository)` function
- [x] T025 [Files] Add README check using `/readme` dedicated endpoint in `PopulateFileChecks()`
- [x] T026 [Files] Add LICENSE check using `/license` dedicated endpoint in `PopulateFileChecks()`
- [x] T027 [Files] Add CODE_OF_CONDUCT.md check in 3 locations (root, `.github/`, `docs/`) with short-circuit
- [x] T028 [Files] Add CODEOWNERS check in 3 locations (`.github/`, root, `docs/`) with short-circuit
- [x] T029 [Files] Add SECURITY.md check in 2 locations (root, `.github/`) with short-circuit
- [x] T030 [Files] Add CONTRIBUTING.md check in 2 locations (root, `.github/`) with short-circuit
- [x] T031 [Files] Add issue templates check in 3 locations (`.github/ISSUE_TEMPLATE/` dir, `.github/ISSUE_TEMPLATE.md`, `ISSUE_TEMPLATE.md`) with short-circuit
- [x] T032 [Files] Add PR template check in 4 locations (`.github/PULL_REQUEST_TEMPLATE.md`, `.github/PULL_REQUEST_TEMPLATE`, `PULL_REQUEST_TEMPLATE.md`, `docs/PULL_REQUEST_TEMPLATE.md`) with short-circuit

**Checkpoint**: File checks working — all community files detected in their standard locations

---

## Phase 5: Extended Checks (Security & Automation)

**Purpose**: Check for security settings and automation configuration

### User Story 3: Extended Security & Automation Checks (P2)

- [x] T033 [P] [Extended] Implement `PopulateExtendedChecks(repo *Repository)` function
- [x] T034 [P] [Extended] Add `isForbidden(err error)` helper checking for HTTP 403 status
- [x] T035 [Extended] Add Dependabot check for `.github/dependabot.yml` and `.github/dependabot.yaml` in `PopulateExtendedChecks()`
- [x] T036 [Extended] Add CI workflows check by listing `.github/workflows/` directory contents
- [x] T037 [Extended] Add default branch protection check via `/branches/{branch}/protection` endpoint (403/404 → unprotected)
- [x] T038 [Extended] Add rulesets check via `/rulesets` endpoint (non-empty array → HasRulesets)
- [x] T039 [Extended] Add vulnerability alerts check via `/vulnerability-alerts` endpoint (403 → Unknown, 404 → disabled, 200/204 → enabled)
- [x] T040 [Extended] Add secret scanning parsing from `SecurityAndAnalysis.SecretScanning.Status` field (empty → Unknown)
- [x] T041 [Extended] Add push protection parsing from `SecurityAndAnalysis.SecretScanningPushProtection.Status` field (empty → Unknown)

**Checkpoint**: Extended checks working — security and automation settings detected with permission-aware error handling

---

## Phase 6: Branch & Tag Checks

**Purpose**: Count branches/tags and detect stale branches

### User Story 4: Branch & Tag Analysis (P2)

- [x] T042 [Branches] Implement `PopulateBranchTagChecks(repo *Repository, since time.Time)` function
- [x] T043 [Branches] Add branch pagination loop with `per_page=100&page={n}` query parameters
- [x] T044 [Branches] Increment `BranchCount` for each branch returned
- [x] T045 [Branches] Add staleness check loop for non-default branches (skip default branch)
- [x] T046 [Branches] Query commits API for each branch: `/commits?sha={branch}&since={timestamp}&per_page=1`
- [x] T047 [Branches] Increment `StaleBranchCount` when commits array is empty (no recent commits)
- [x] T048 [Branches] Add error skip logic in branch staleness check (continue on error, don't abort)
- [x] T049 [P] [Branches] Add tag pagination loop with `per_page=100&page={n}` query parameters
- [x] T050 [P] [Branches] Increment `TagCount` for each tag returned

**Checkpoint**: Branch and tag checks working — counts accurate, stale branches detected (expensive but functional)

---

## Phase 7: Helper Functions & Error Handling

**Purpose**: Centralize error handling logic

### User Story 5: Error Handling & Permissions (P2)

- [x] T051 [P] [ErrorHandling] Implement `isNotFound(err error)` helper checking `api.HTTPError` for 404 status
- [x] T052 [P] [ErrorHandling] Implement `isForbidden(err error)` helper checking `api.HTTPError` for 403 status
- [x] T053 [ErrorHandling] Apply 404 handling to `CheckFileExists()` (return `false, nil` on 404)
- [x] T054 [ErrorHandling] Apply 403 handling to branch protection check (treat as unprotected)
- [x] T055 [ErrorHandling] Apply 403 handling to vulnerability alerts check (mark as Unknown)
- [x] T056 [ErrorHandling] Propagate other errors (non-404/403) to caller

**Checkpoint**: Error handling consistent — graceful degradation for permission issues

---

## Phase 8: Testing Infrastructure

**Purpose**: Create mock server tests for API interactions

- [x] T057 [P] [Tests] Create `mockServer(t *testing.T, handler http.Handler)` helper in `internal/api/client_test.go`
- [x] T058 [P] [Tests] Add `TestGetRepo_Mock` with JSON fixture for single repo fetch
- [x] T059 [P] [Tests] Add `TestListOrgRepos_Pagination_Mock` with multi-page fixture (page 1: 2 repos, page 2: 1 repo)
- [x] T060 [P] [Tests] Add `TestCheckFileExists_Mock` with 200/404 response handling
- [x] T061 [P] [Tests] Add test fixtures for repository JSON responses

**Checkpoint**: Test infrastructure working — can test API client with mock HTTP server

---

## Phase 9: Integration & CLI Wiring

**Purpose**: Connect API client to CLI and checks package

- [x] T062 [Integration] Verify `NewClient()` called from `cmd/gh-repo-health-report/main.go`
- [x] T063 [Integration] Verify `GetRepo()` called for `--repo` flag handling
- [x] T064 [Integration] Verify `ListOrgRepos()` called for `--org` flag handling
- [x] T065 [Integration] Verify `ListUserRepos()` called for `--owner` flag handling
- [x] T066 [Integration] Verify `PopulateFileChecks()` called for each repository before evaluation
- [x] T067 [Integration] Verify `PopulateExtendedChecks()` called for each repository before evaluation
- [x] T068 [Integration] Verify `PopulateBranchTagChecks()` called with `sinceTime` from `--since` flag parsing
- [x] T069 [Integration] Verify `Repository` passed to `checks.Evaluate()` after all Populate* methods complete

**Checkpoint**: API client fully integrated — CLI → API → Checks → Formatter pipeline working

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Client Foundation)**: No dependencies — must be first
- **Phase 2 (Data Model)**: Can run in parallel with Phase 1 (separate concerns)
- **Phase 3 (Fetching)**: Depends on Phase 1 + 2 (needs Client and Repository)
- **Phase 4 (File Checks)**: Depends on Phase 1 + 2 + 3 (needs fetch methods)
- **Phase 5 (Extended Checks)**: Depends on Phase 1 + 2 + 3 (needs fetch methods)
- **Phase 6 (Branch Checks)**: Depends on Phase 1 + 2 + 3 (needs fetch methods)
- **Phase 7 (Error Handling)**: Integrated throughout Phases 4-6 (helpers used by all)
- **Phase 8 (Testing)**: Depends on Phases 1-7 (tests implementation)
- **Phase 9 (Integration)**: Depends on all previous phases (final wiring)

### Within Phases

**Phase 2 (Model)**: All tasks marked [P] — independent struct field definitions
**Phase 3 (Fetch)**: T015-T019 sequential (shared logic), T020-T021 parallel (different methods)
**Phase 4 (Files)**: T022-T024 sequential (helper → main function), T025-T032 could be parallel (independent checks)
**Phase 5 (Extended)**: T033-T034 first (setup), T035-T041 could be parallel (independent checks)
**Phase 6 (Branches)**: T042-T048 sequential (branch logic), T049-T050 parallel (tag logic separate)
**Phase 7 (Errors)**: T051-T052 parallel (independent helpers), T053-T056 sequential (applications)
**Phase 8 (Tests)**: All marked [P] — independent test cases

### Parallel Opportunities

- Phase 1 + Phase 2 can run simultaneously
- Within Phase 2: All 10 tasks parallel (T005-T014)
- Within Phase 4: File check implementations parallel after helpers done
- Within Phase 5: Individual check implementations parallel after setup
- Within Phase 8: All test tasks parallel (T057-T061)

---

## Implementation Notes

### Actual Development Pattern

**Real-world implementation** likely evolved as:
1. Basic fetch methods (single repo, org list) — MVP
2. File checks added incrementally as user requests came in
3. Extended checks added later (security, automation)
4. Branch/tag checks added last (expensive, lower priority)
5. Error handling refined as permission issues discovered
6. Tests added alongside each major feature

### Key Design Decisions Made

**Decision 1**: Populate pattern instead of fetch-with-all-data
- **Rationale**: Separation of concerns, caller controls which checks run
- **Trade-off**: More API calls (incremental) vs. GraphQL (batch)

**Decision 2**: Client-side filtering for forks/archives
- **Rationale**: GitHub API doesn't support server-side filtering
- **Trade-off**: Wasted API calls vs. accurate filtering

**Decision 3**: Skip errors in branch staleness checks
- **Rationale**: Partial data better than no data
- **Trade-off**: May under-count stale branches vs. accurate count

**Decision 4**: Tristate security fields with inconsistent 403 handling
- **Rationale**: Evolved incrementally; branch protection predates tristate
- **Trade-off**: Confusing inconsistency vs. major refactor

### Testing Philosophy

**Mock HTTP server approach**:
- Avoids real API calls (no auth required, fast)
- Tests pagination and error handling logic
- Limited fixture maintenance (simple JSON responses)
- No integration tests against real GitHub API (relies on go-gh library correctness)

### Gaps Identified During Migration

1. ⚠️ **No rate limit handling** — Branch checks can consume 100+ API calls; may hit 5000/hour limit
2. ⚠️ **No retry logic** — Transient errors abort entire operation
3. ⚠️ **Inconsistent 403 handling** — Branch protection vs. vulnerability alerts treated differently
4. ⚠️ **Limited test coverage** — Extended checks and branch staleness not tested (fixtures too complex)
5. ⚠️ **No caching** — Repeated audits re-fetch same data

---

## Total Implementation Effort

**Completed Tasks**: 69 tasks across 9 phases  
**Files Modified**: 2 files (`client.go`, `client_test.go`)  
**Lines of Code**: 707 total (453 implementation + 254 tests)  
**API Endpoints**: 15+ distinct endpoints used  
**Check Methods**: 3 main Populate* methods covering 20+ individual checks

---

## Maintenance Guidance

**Adding New File Check**:
1. Add boolean field to `Repository` struct (Phase 2 pattern)
2. Add check location loop in `PopulateFileChecks()` (Phase 4 pattern)
3. Use `CheckFileExists()` helper with short-circuit on first match
4. Add test fixture in `client_test.go`

**Adding New Extended Check**:
1. Add field(s) to `Repository` struct (consider tristate if permission-gated)
2. Add API call in `PopulateExtendedChecks()` (Phase 5 pattern)
3. Handle 404/403 appropriately based on check semantics
4. Add test with mock response

**Improving Performance**:
- Consider GraphQL for batching (major refactor)
- Add caching layer for repeated audits
- Skip branch staleness for repos with <10 branches
- Add rate limit detection and throttling
