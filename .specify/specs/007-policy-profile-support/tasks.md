# Tasks: Policy Profile Support

**Feature**: 007-policy-profile-support  
**Input**: Design documents from `.specify/specs/007-policy-profile-support/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli-contract.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

**gh-repo-health-report structure** (Go project):
- **CLI entry**: `cmd/gh-repo-health-report/main.go`
- **Internal packages**: `internal/api/`, `internal/checks/`, `internal/formatter/`
- **Tests colocated**: `*_test.go` files alongside source (e.g., `internal/checks/profile_test.go`)
- **No separate test directories** — Go convention is tests in same package
- **Build output**: `gh-repo-health-report` binary at repository root

**Go Testing Commands**:
- Run all tests: `go test ./...`
- Run specific package: `go test ./internal/checks`
- Run with coverage: `go test -cover ./...`
- Table-driven tests recommended for multiple scenarios

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and dependency management

- [ ] T001 Add `gopkg.in/yaml.v3` dependency to `go.mod` with `go get gopkg.in/yaml.v3`
- [ ] T002 [P] Verify build works: `go build ./...`
- [ ] T003 [P] Verify existing tests pass: `go test ./...`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 [P] Define `EnforcementLevel` enum type (Required=0, Recommended=1, Ignored=2) with String() method in `internal/checks/profile.go`
- [ ] T005 [P] Define `Profile` struct (Name, Description, Checks map[string]EnforcementLevel) in `internal/checks/profile.go`
- [ ] T006 [P] Define `SkippedCheck` struct (Name, Reason) in `internal/checks/checks.go`
- [ ] T007 [P] Add `Profile *Profile` field to `Options` struct in `internal/checks/checks.go`
- [ ] T008 [P] Add `SkippedChecks []SkippedCheck` field to `Result` struct in `internal/checks/checks.go`
- [ ] T009 [P] Define `Config` struct (DefaultProfile string) in `internal/checks/config.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Tailor Health Checks by Repository Type (Priority: P1) 🎯 MVP

**Goal**: Enable users to apply different health check expectations based on repository type (open-source, internal-service, application, archived, prototype) so health reports reflect realistic governance policies.

**Independent Test**: Run `gh repo-health-report --profile open-source --repo owner/library` and verify that checks marked as "ignored" in the open-source profile don't appear in the failed checks list, while "required" checks do.

### Implementation for User Story 1

- [ ] T010 [P] [US1] Define `ProfileOpenSource` variable with all existing checks mapped to enforcement levels in `internal/checks/profile.go`
- [ ] T011 [P] [US1] Define `ProfileInternalService` variable with all existing checks mapped to enforcement levels in `internal/checks/profile.go`
- [ ] T012 [P] [US1] Define `ProfileApplication` variable with all existing checks mapped to enforcement levels in `internal/checks/profile.go`
- [ ] T013 [P] [US1] Define `ProfileArchived` variable with all existing checks mapped to enforcement levels in `internal/checks/profile.go`
- [ ] T014 [P] [US1] Define `ProfilePrototype` variable with all existing checks mapped to enforcement levels in `internal/checks/profile.go`
- [ ] T015 [US1] Implement `GetProfile(name string) *Profile` function that returns predefined profiles or nil in `internal/checks/profile.go`
- [ ] T016 [P] [US1] Add table-driven test in `internal/checks/profile_test.go` to verify each predefined profile has all existing checks defined
- [ ] T017 [P] [US1] Add test in `internal/checks/profile_test.go` to verify `GetProfile()` returns correct profile for valid names
- [ ] T018 [P] [US1] Add test in `internal/checks/profile_test.go` to verify `GetProfile()` returns nil for invalid names
- [ ] T019 [US1] Update `Evaluate()` function in `internal/checks/checks.go` to handle nil profile (backward compatibility path)
- [ ] T020 [US1] Update `Evaluate()` function in `internal/checks/checks.go` to skip checks with EnforcementIgnored and add to SkippedChecks
- [ ] T021 [US1] Update `Evaluate()` function in `internal/checks/checks.go` to evaluate checks with EnforcementRequired or EnforcementRecommended
- [ ] T022 [P] [US1] Add test in `internal/checks/checks_test.go` to verify Evaluate() with nil profile maintains legacy behavior
- [ ] T023 [P] [US1] Add test in `internal/checks/checks_test.go` to verify Evaluate() with open-source profile correctly skips ignored checks
- [ ] T024 [P] [US1] Add test in `internal/checks/checks_test.go` to verify SkippedChecks field is populated correctly
- [ ] T025 [US1] Add `--profile` string flag to root command in `cmd/gh-repo-health-report/main.go`
- [ ] T026 [US1] Implement profile resolution logic in `cmd/gh-repo-health-report/main.go` to call GetProfile() with flag value
- [ ] T027 [US1] Pass resolved profile to checks.Options in Evaluate() calls in `cmd/gh-repo-health-report/main.go`
- [ ] T028 [US1] Update help text in `cmd/gh-repo-health-report/main.go` to document --profile flag and list five profile types
- [ ] T029 [US1] Test manually: `go build && ./gh-repo-health-report --profile open-source --repo owner/library`
- [ ] T030 [US1] Test manually: Verify without --profile flag, all 28 checks evaluated (backward compatibility)

**Checkpoint**: User Story 1 complete - profiles can be specified via CLI and checks are filtered correctly

---

## Phase 4: User Story 2 - Set Organization-Wide Profile Defaults (Priority: P2)

**Goal**: Allow platform engineering managers to define a default profile in a configuration file that applies to all repositories without requiring users to specify the profile flag each time.

**Independent Test**: Create config file with `default_profile: internal-service`, run `gh repo-health-report --org myorg` without profile flag, and verify all repos are evaluated against the internal-service profile.

### Implementation for User Story 2

- [ ] T031 [P] [US2] Implement `LoadConfig(path string) (*Config, error)` function for explicit path in `internal/checks/config.go`
- [ ] T032 [P] [US2] Add YAML parsing support using `gopkg.in/yaml.v3` in `internal/checks/config.go`
- [ ] T033 [P] [US2] Add JSON parsing support using `encoding/json` in `internal/checks/config.go`
- [ ] T034 [US2] Implement `DiscoverConfig() (*Config, error)` function in `internal/checks/config.go` to search for config files
- [ ] T035 [US2] Add search for `./.gh-repo-health-report.yml` in current directory in `internal/checks/config.go`
- [ ] T036 [US2] Add search for `./.gh-repo-health-report.json` in current directory in `internal/checks/config.go`
- [ ] T037 [US2] Add search for `~/.gh-repo-health-report.yml` in home directory in `internal/checks/config.go`
- [ ] T038 [US2] Add search for `~/.gh-repo-health-report.json` in home directory in `internal/checks/config.go`
- [ ] T039 [P] [US2] Add test in `internal/checks/config_test.go` for loading valid YAML config file
- [ ] T040 [P] [US2] Add test in `internal/checks/config_test.go` for loading valid JSON config file
- [ ] T041 [P] [US2] Add test in `internal/checks/config_test.go` for config file discovery order (current dir before home)
- [ ] T042 [P] [US2] Add test in `internal/checks/config_test.go` for error handling on invalid YAML syntax
- [ ] T043 [P] [US2] Add test in `internal/checks/config_test.go` for error handling on invalid JSON syntax
- [ ] T044 [P] [US2] Add test in `internal/checks/config_test.go` for missing config file returns nil without error
- [ ] T045 [US2] Add `--profile-config` string flag to root command in `cmd/gh-repo-health-report/main.go`
- [ ] T046 [US2] Update profile resolution logic in `cmd/gh-repo-health-report/main.go` to call LoadConfig() when --profile-config specified
- [ ] T047 [US2] Update profile resolution logic in `cmd/gh-repo-health-report/main.go` to call DiscoverConfig() when no --profile-config
- [ ] T048 [US2] Implement precedence logic: --profile flag > config default_profile > no profile in `cmd/gh-repo-health-report/main.go`
- [ ] T049 [US2] Add validation for invalid profile names in config file in `cmd/gh-repo-health-report/main.go`
- [ ] T050 [US2] Update help text in `cmd/gh-repo-health-report/main.go` to document --profile-config flag and config file format
- [ ] T051 [US2] Test manually: Create `.gh-repo-health-report.yml` with `default_profile: internal-service` and run without --profile flag
- [ ] T052 [US2] Test manually: Verify --profile flag overrides config file default

**Checkpoint**: User Story 2 complete - default profiles can be set in config files

---

## Phase 5: User Story 3 - Understand Why Checks Were Skipped (Priority: P2)

**Goal**: Provide transparency by showing explanations for why certain checks were skipped or ignored, helping users understand profile policy decisions.

**Independent Test**: Run `gh repo-health-report --profile archived --format table --repo owner/old-repo` and verify output includes indicators showing which checks were skipped (e.g., `[SKIP]` in table or explanatory fields in JSON).

### Implementation for User Story 3

- [ ] T053 [P] [US3] Update `formatTable()` function in `internal/formatter/formatter.go` to mark skipped checks with `[SKIP]` notation
- [ ] T054 [P] [US3] Add footer legend to table output in `internal/formatter/formatter.go` when profile is active
- [ ] T055 [P] [US3] Update `formatJSON()` function in `internal/formatter/formatter.go` to add `profile` field with profile name
- [ ] T056 [P] [US3] Update `formatJSON()` function in `internal/formatter/formatter.go` to add `skipped_checks` array with check objects
- [ ] T057 [P] [US3] Update `formatCSV()` function in `internal/formatter/formatter.go` to add `profile` column
- [ ] T058 [P] [US3] Update `formatCSV()` function in `internal/formatter/formatter.go` to add `skipped_checks` column with comma-separated names
- [ ] T059 [P] [US3] Update `formatMD()` function in `internal/formatter/formatter.go` to mark skipped checks with `[SKIP]` in table
- [ ] T060 [P] [US3] Add profile name and skipped checks summary below markdown table in `internal/formatter/formatter.go`
- [ ] T061 [P] [US3] Add test in `internal/formatter/formatter_test.go` for table format with profile includes `[SKIP]` markers
- [ ] T062 [P] [US3] Add test in `internal/formatter/formatter_test.go` for table format without profile has no `[SKIP]` markers
- [ ] T063 [P] [US3] Add test in `internal/formatter/formatter_test.go` for JSON format includes `skipped_checks` field when profile active
- [ ] T064 [P] [US3] Add test in `internal/formatter/formatter_test.go` for CSV format includes `skipped_checks` column
- [ ] T065 [P] [US3] Add test in `internal/formatter/formatter_test.go` for markdown format includes profile summary
- [ ] T066 [US3] Test manually: Run with `--profile archived --format table` and verify `[SKIP]` markers appear
- [ ] T067 [US3] Test manually: Run with `--profile open-source --format json` and verify `skipped_checks` array is present
- [ ] T068 [US3] Test manually: Run with `--profile internal-service --format csv` and verify `skipped_checks` column is populated

**Checkpoint**: User Story 3 complete - all output formats show skipped check indicators

---

## Phase 6: User Story 4 - Define Profile-Aware Fail Thresholds (Priority: P3)

**Goal**: Make the `--fail-on` flag work correctly with profiles so exit codes reflect only the checks that matter for the repository type.

**Independent Test**: Run `gh repo-health-report --profile prototype --fail-on any --repo owner/experiment` where a prototype-ignored check fails, verify exit code 0. Then test with a required check failing and verify exit code 1.

### Implementation for User Story 4

- [ ] T069 [US4] Update `--fail-on` logic in `cmd/gh-repo-health-report/main.go` to respect profile when value is "any"
- [ ] T070 [US4] Update `--fail-on` logic in `cmd/gh-repo-health-report/main.go` to override profile when specific check name specified
- [ ] T071 [US4] Implement exit code 1 only when non-ignored checks fail with `--fail-on any` in `cmd/gh-repo-health-report/main.go`
- [ ] T072 [US4] Implement exit code 1 when specific check fails even if profile ignores it in `cmd/gh-repo-health-report/main.go`
- [ ] T073 [US4] Add error handling for invalid profile names with helpful error message in `cmd/gh-repo-health-report/main.go`
- [ ] T074 [US4] Test manually: Run with `--profile archived --fail-on stale` where archived profile ignores stale, verify exit 0
- [ ] T075 [US4] Test manually: Run with `--profile internal-service --fail-on any` with required check failing, verify exit 1
- [ ] T076 [US4] Test manually: Run with `--profile internal-service --fail-on any` with only ignored checks failing, verify exit 0
- [ ] T077 [US4] Test manually: Run with `--profile open-source --fail-on missing-codeowners` (ignored by profile), verify exit 1 (explicit override)
- [ ] T078 [US4] Test manually: Run without profile and `--fail-on any`, verify backward compatibility (all failures exit 1)

**Checkpoint**: User Story 4 complete - `--fail-on` flag respects profile policies

---

## Phase 7: User Story 5 - Auto-Detect Profile from Repository Metadata (Priority: P3)

**Goal**: Automatically detect the appropriate profile based on repository metadata (archived status, topics, naming conventions) to reduce manual profile specification.

**Independent Test**: Run `gh repo-health-report --profile auto --repo owner/archived-repo` (where repository is archived) and verify it automatically uses the archived profile. Test with a repo containing topic `prototype` and verify it selects the prototype profile.

### Implementation for User Story 5

- [ ] T079 [P] [US5] Implement `DetectProfile(repo *api.Repository) *Profile` function in `internal/checks/profile.go`
- [ ] T080 [US5] Add priority 1 check for archived status → archived profile in `internal/checks/profile.go`
- [ ] T081 [US5] Add priority 2 topic matching: prototype/experimental/poc/spike → prototype profile in `internal/checks/profile.go`
- [ ] T082 [US5] Add priority 2 topic matching: library/package/npm-package/gem/pypi → open-source profile in `internal/checks/profile.go`
- [ ] T083 [US5] Add priority 2 topic matching: service/api/microservice → internal-service profile in `internal/checks/profile.go`
- [ ] T084 [US5] Add priority 2 topic matching: app/webapp/mobile-app/desktop → application profile in `internal/checks/profile.go`
- [ ] T085 [US5] Add priority 3 visibility check: public repositories → open-source profile in `internal/checks/profile.go`
- [ ] T086 [US5] Add priority 4 fallback: private repositories → internal-service profile in `internal/checks/profile.go`
- [ ] T087 [P] [US5] Add test in `internal/checks/profile_test.go` for auto-detection with archived repositories
- [ ] T088 [P] [US5] Add test in `internal/checks/profile_test.go` for auto-detection with prototype topic
- [ ] T089 [P] [US5] Add test in `internal/checks/profile_test.go` for auto-detection with library topic
- [ ] T090 [P] [US5] Add test in `internal/checks/profile_test.go` for auto-detection with service topic
- [ ] T091 [P] [US5] Add test in `internal/checks/profile_test.go` for auto-detection with app topic
- [ ] T092 [P] [US5] Add test in `internal/checks/profile_test.go` for auto-detection with public visibility
- [ ] T093 [P] [US5] Add test in `internal/checks/profile_test.go` for auto-detection fallback with private repo no topics
- [ ] T094 [P] [US5] Add test in `internal/checks/profile_test.go` for topic conflict resolution (first matching wins)
- [ ] T095 [US5] Update profile resolution logic in `cmd/gh-repo-health-report/main.go` to call DetectProfile() when --profile is "auto"
- [ ] T096 [US5] Update help text in `cmd/gh-repo-health-report/main.go` to document "auto" profile option and detection heuristics
- [ ] T097 [US5] Test manually: Run with `--profile auto` on archived repository, verify archived profile selected
- [ ] T098 [US5] Test manually: Run with `--profile auto` on repository with `prototype` topic, verify prototype profile selected
- [ ] T099 [US5] Test manually: Run with `--profile auto` on public repository without topics, verify open-source profile selected
- [ ] T100 [US5] Test manually: Run with `--profile auto --org myorg` and verify different repos get different profiles

**Checkpoint**: User Story 5 complete - auto-detection infers profiles from repository metadata

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T101 [P] Update README.md to add "Policy Profiles" section explaining the feature
- [ ] T102 [P] Document five predefined profiles with check enforcement lists in README.md
- [ ] T103 [P] Document `--profile` and `--profile-config` flags with usage examples in README.md
- [ ] T104 [P] Add config file format examples (YAML and JSON) to README.md
- [ ] T105 [P] Document auto-detection behavior and heuristics in README.md
- [ ] T106 [P] Explain `--fail-on` interaction with profiles in README.md
- [ ] T107 [P] Add migration guide for existing users in README.md
- [ ] T108 Run `go fmt ./...` to format all code
- [ ] T109 Run `go vet ./...` to check for code issues
- [ ] T110 Run full test suite with coverage: `go test -cover ./...`
- [ ] T111 Build binary: `go build -o gh-repo-health-report ./cmd/gh-repo-health-report`
- [ ] T112 Test GitHub CLI integration: `gh repo-health-report --help` (verify it works as extension)
- [ ] T113 Test profile validation on 100+ repositories with different profiles
- [ ] T114 Verify backward compatibility: Test without profiles on existing repositories
- [ ] T115 Performance validation: Profile evaluation adds <1ms overhead per repository
- [ ] T116 Final integration test: All output formats (table, JSON, CSV, markdown) with all five profiles

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - User Story 1 (P1): Core profile system - no dependencies on other stories
  - User Story 2 (P2): Config file support - can run parallel with US1 after Foundational
  - User Story 3 (P2): Output formats - depends on US1 (needs Profile struct and evaluation logic)
  - User Story 4 (P3): Fail thresholds - depends on US1 (needs profile-aware evaluation)
  - User Story 5 (P3): Auto-detection - depends on US1 (needs Profile struct), can run parallel with US2
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Independent of US1 but integrates with it
- **User Story 3 (P2)**: Can start after US1 complete (needs profile evaluation working) - Independent of US2
- **User Story 4 (P3)**: Can start after US1 complete (needs profile evaluation working) - Independent of US2, US3
- **User Story 5 (P3)**: Can start after US1 complete (needs Profile struct) - Can run parallel with US2

### Within Each User Story

- Tests can run in parallel (all marked [P])
- Implementation tasks for different files can run in parallel (all marked [P])
- Tasks modifying the same file must run sequentially
- Manual testing comes after implementation complete

### Parallel Opportunities

- All Setup tasks (T001-T003) can run in parallel
- All Foundational tasks (T004-T009) can run in parallel
- Within US1: T010-T014 (define profiles) can run in parallel, T016-T018 (tests) can run in parallel, T022-T024 (tests) can run in parallel
- Within US2: T031-T033 (config functions) can run in parallel, T039-T044 (tests) can run in parallel
- Within US3: T053-T060 (formatter updates) can run in parallel, T061-T065 (tests) can run in parallel
- Within US5: T080-T086 (detection rules) can run in parallel, T087-T094 (tests) can run in parallel
- Polish tasks T101-T107 (documentation) can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all profile definitions together:
Task T010: "Define ProfileOpenSource in internal/checks/profile.go"
Task T011: "Define ProfileInternalService in internal/checks/profile.go"
Task T012: "Define ProfileApplication in internal/checks/profile.go"
Task T013: "Define ProfileArchived in internal/checks/profile.go"
Task T014: "Define ProfilePrototype in internal/checks/profile.go"

# Launch all profile validation tests together:
Task T016: "Test each profile has all 28 checks"
Task T017: "Test GetProfile returns correct profile"
Task T018: "Test GetProfile returns nil for invalid names"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (add YAML dependency)
2. Complete Phase 2: Foundational (define core types)
3. Complete Phase 3: User Story 1 (profiles selectable via CLI)
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

**Estimated Time for MVP**: 8-10 hours

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!) - Core profile selection
3. Add User Story 2 → Test independently → Deploy/Demo - Config file defaults
4. Add User Story 3 → Test independently → Deploy/Demo - Output transparency
5. Add User Story 4 → Test independently → Deploy/Demo - CI/CD integration
6. Add User Story 5 → Test independently → Deploy/Demo - Auto-detection convenience
7. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (core profiles)
   - Developer B: User Story 2 (config files) - can start in parallel
   - Developer C: User Story 5 (auto-detection) - can start in parallel
3. After US1 completes:
   - Developer D: User Story 3 (output formats)
   - Developer E: User Story 4 (fail thresholds)
4. Stories complete and integrate independently

---

## Summary

**Total Tasks**: 116  
**MVP Tasks (Setup + Foundational + US1)**: 30  
**Critical Path**: Setup → Foundational → US1 → US3 (output transparency depends on profile evaluation)

**User Story Breakdown**:
- US1 (P1 - Core profiles): 21 tasks (T010-T030)
- US2 (P2 - Config files): 22 tasks (T031-T052)
- US3 (P2 - Output transparency): 16 tasks (T053-T068)
- US4 (P3 - Fail thresholds): 10 tasks (T069-T078)
- US5 (P3 - Auto-detection): 22 tasks (T079-T100)
- Polish: 16 tasks (T101-T116)

**Parallel Opportunities**: 45 tasks marked [P] can run in parallel within their phase

**Estimated Total Time**: 12-19 hours
- Setup + Foundational: 1-2 hours
- User Story 1 (MVP): 6-8 hours
- User Story 2: 2-3 hours
- User Story 3: 2-3 hours
- User Story 4: 1-2 hours
- User Story 5: 2-3 hours
- Polish: 2-3 hours

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label (US1, US2, US3, US4, US5) maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Tests are included for all components (profile definitions, config loading, evaluation logic, output formats)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Backward compatibility maintained throughout (nil profile = legacy behavior)
