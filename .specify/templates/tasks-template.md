---

description: "Task list template for feature implementation"
---

# Tasks: [FEATURE NAME]

**Input**: Design documents from `/specs/[###-feature-name]/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: The examples below include test tasks. Tests are OPTIONAL - only include them if explicitly requested in the feature specification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

**gh-repo-health-report structure** (Go project):
- **CLI entry**: `cmd/gh-repo-health-report/main.go`
- **Internal packages**: `internal/api/`, `internal/checks/`, `internal/formatter/`
- **Tests colocated**: `*_test.go` files alongside source (e.g., `internal/checks/checks_test.go`)
- **No separate test directories** — Go convention is tests in same package
- **Build output**: `gh-repo-health-report` binary at repository root

**Go Testing Commands**:
- Run all tests: `go test ./...`
- Run specific package: `go test ./internal/checks`
- Run with coverage: `go test -cover ./...`
- Table-driven tests recommended for multiple scenarios

<!-- 
  ============================================================================
  IMPORTANT: The tasks below are SAMPLE TASKS for illustration purposes only.
  
  The /speckit.tasks command MUST replace these with actual tasks based on:
  - User stories from spec.md (with their priorities P1, P2, P3...)
  - Feature requirements from plan.md
  - Entities from data-model.md
  - Endpoints from contracts/
  
  Tasks MUST be organized by user story so each story can be:
  - Implemented independently
  - Tested independently
  - Delivered as an MVP increment
  
  DO NOT keep these sample tasks in the generated tasks.md file.
  ============================================================================
-->

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure (Go projects typically need minimal setup)

- [ ] T001 Update `go.mod` if adding new dependencies (run `go get <package>`)
- [ ] T002 [P] Verify build works: `go build ./...`
- [ ] T003 [P] Verify tests pass: `go test ./...`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

Examples of foundational tasks for gh-repo-health-report features:

- [ ] T004 Add new constants to `internal/checks/checks.go` (if adding new check types)
- [ ] T005 [P] Add new field to `api.Repository` struct in `internal/api/client.go` (if API data model changes)
- [ ] T006 [P] Extend `formatter.Format()` signature in `internal/formatter/formatter.go` (if output structure changes)
- [ ] T007 Add new flag to root command in `cmd/gh-repo-health-report/main.go` (if CLI interface changes)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - [Title] (Priority: P1) 🎯 MVP

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 1 (OPTIONAL - only if tests requested) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T010 [P] [US1] Add test table to `internal/checks/checks_test.go` for new check logic
- [ ] T011 [P] [US1] Add test cases to `internal/api/client_test.go` for new API functions (use fixtures, not real API)

### Implementation for User Story 1

- [ ] T012 [P] [US1] Add check constant `Check[Name]` in `internal/checks/checks.go`
- [ ] T013 [P] [US1] Add API function `Fetch[Data]()` in `internal/api/client.go` (if API call needed)
- [ ] T014 [US1] Implement check evaluation logic in `internal/checks/checks.go` (depends on T012)
- [ ] T015 [US1] Update formatter to display new check result in `internal/formatter/formatter.go`
- [ ] T016 [US1] Add CLI flag for feature in `cmd/gh-repo-health-report/main.go` (if user-configurable)
- [ ] T017 [US1] Test manually: `go build && ./gh-repo-health-report --org test-org`

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - [Title] (Priority: P2)

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 2 (OPTIONAL - only if tests requested) ⚠️

- [ ] T018 [P] [US2] Add test table to appropriate `*_test.go` file for new functionality
- [ ] T019 [P] [US2] Add mocked API test cases in `internal/api/client_test.go` (if applicable)

### Implementation for User Story 2

- [ ] T020 [P] [US2] Add new types/structs to appropriate `internal/` package
- [ ] T021 [US2] Implement core logic in relevant `internal/` package
- [ ] T022 [US2] Wire up to CLI or API layer as needed
- [ ] T023 [US2] Test manually and verify both US1 and US2 work independently

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - [Title] (Priority: P3)

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 3 (OPTIONAL - only if tests requested) ⚠️

- [ ] T024 [P] [US3] Add test table to appropriate `*_test.go` file
- [ ] T025 [P] [US3] Add edge case tests for new functionality

### Implementation for User Story 3

- [ ] T026 [P] [US3] Add new types/constants to appropriate package
- [ ] T027 [US3] Implement core logic
- [ ] T028 [US3] Integrate with existing components

**Checkpoint**: All user stories should now be independently functional

---

[Add more user story phases as needed, following the same pattern]

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] TXXX [P] Update README.md with new feature documentation
- [ ] TXXX Code cleanup and refactoring (run `gofmt` or `goimports`)
- [ ] TXXX Run `go vet ./...` to check for issues
- [ ] TXXX Run full test suite: `go test -cover ./...`
- [ ] TXXX Build and test manually: `go build && ./gh-repo-health-report [flags]`
- [ ] TXXX Performance optimization (if needed for large org scans)
- [ ] TXXX [P] Additional table-driven tests for edge cases
- [ ] TXXX Verify GitHub CLI integration still works: `gh repo-health-report --help`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May integrate with US1 but should be independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - May integrate with US1/US2 but should be independently testable

### Within Each User Story

- Tests (if included) MUST be written and FAIL before implementation
- Models before services
- Services before endpoints
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Models within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all test updates for User Story 1 together (if tests requested):
Task: "Add test table to internal/checks/checks_test.go"
Task: "Add test cases to internal/api/client_test.go"

# Launch all constant/type additions for User Story 1 together:
Task: "Add check constant CheckNewFeature in internal/checks/checks.go"
Task: "Add new field to Repository struct in internal/api/client.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
