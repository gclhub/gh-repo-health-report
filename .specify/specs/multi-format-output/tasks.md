# Tasks: Multi-Format Output

**Status**: migrated (all tasks completed)  
**Input**: Reverse-engineered from `internal/formatter/formatter.go` and `internal/formatter/formatter_test.go`  
**Implementation**: Existing feature — all tasks marked as completed

## Format: `[ID] [P?] [Category] Description`

- **[x]**: Task completed (all tasks pre-existing)
- **[P]**: Could have run in parallel (different files, no dependencies)
- **[Category]**: Which format this task belongs to

---

## Phase 1: Foundation & Helper Functions

**Purpose**: Core formatting infrastructure and symbol mapping

- [x] T001 [P] [Foundation] Create `internal/formatter/formatter.go` with package declaration
- [x] T002 [P] [Foundation] Define `Format(results []*checks.Result, format string, w io.Writer)` entry point
- [x] T003 [Foundation] Implement format routing switch (json/csv/md/default→table)
- [x] T004 [P] [Helpers] Implement `bool2check(v bool) string` helper (true→"✓", false→"✗")
- [x] T005 [P] [Helpers] Implement `tristate(ok, unknown bool) string` helper (unknown→"?", ok→"✓", else→"✗")
- [x] T006 [P] [Helpers] Implement `staleStr(v bool) string` helper (true→"YES", false→"NO")

**Checkpoint**: Foundation ready — format routing and symbol helpers available

---

## Phase 2: Table Format (Default Output)

**Purpose**: Human-readable terminal output with aligned columns

### User Story 1: Table Output for Terminal (P1) 🎯

- [x] T007 [Table] Define `tableHeader` constant with 28 tab-separated column names
- [x] T008 [Table] Implement `formatTable(results []*checks.Result, w io.Writer)` function
- [x] T009 [Table] Create `tabwriter.NewWriter` with 2-space minimum spacing
- [x] T010 [Table] Write header row to tabwriter
- [x] T011 [Table] Implement result iteration loop for data rows
- [x] T012 [Table] Format first 4 columns: REPO (FullName), STALE (staleStr), DESCRIPTION (bool2check), TOPICS (count)
- [x] T013 [Table] Format community file columns (6): README, LICENSE, CODE_CONDUCT, CODEOWNERS, SECURITY, CONTRIBUTING (bool2check)
- [x] T014 [Table] Format template columns (2): ISSUE_TMPL, PR_TMPL (bool2check)
- [x] T015 [Table] Format feature columns (3): ISSUES, WIKI, PROJECTS (bool2check)
- [x] T016 [Table] Format automation columns (2): DEPENDABOT, CI (bool2check)
- [x] T017 [Table] Format protection columns (2): BR_PROTECT, RULESETS (bool2check)
- [x] T018 [Table] Format security columns (3): VULN_ALERTS, SECRET_SCAN, PUSH_PROT (tristate)
- [x] T019 [Table] Format merge settings column (1): AUTO_DEL_BR (bool2check)
- [x] T020 [Table] Format count columns (5): BRANCHES, STALE_BR, TAGS, OPEN_ISSUES, SIZE_KB
- [x] T021 [Table] Call `tabwriter.Flush()` to output aligned table

**Checkpoint**: Table format working — default CLI output displays properly aligned

---

## Phase 3: JSON Format (Machine-Readable)

**Purpose**: Structured data output for automation and APIs

### User Story 2: JSON Output for Automation (P1) 🎯

- [x] T022 [P] [JSON] Define `jsonRow` struct with 28+ JSON-tagged fields
- [x] T023 [P] [JSON] Add snake_case json tags to all jsonRow fields (e.g., `json:"has_readme"`)
- [x] T024 [P] [JSON] Add tristate field pairs to jsonRow (e.g., `vulnerability_alerts_enabled`, `vulnerability_alerts_unknown`)
- [x] T025 [JSON] Implement `toRow(r *checks.Result) jsonRow` conversion function
- [x] T026 [JSON] Map all boolean fields from Result to jsonRow (20+ fields)
- [x] T027 [JSON] Map all count fields from Result to jsonRow (topics, branches, tags, etc.)
- [x] T028 [JSON] Map tristate security fields (6 fields: 3 pairs of enabled/unknown)
- [x] T029 [JSON] Implement `formatJSON(results []*checks.Result, w io.Writer)` function
- [x] T030 [JSON] Convert results slice to jsonRow slice using toRow()
- [x] T031 [JSON] Create `json.NewEncoder(w)` with 2-space indentation (`SetIndent("", "  ")`)
- [x] T032 [JSON] Encode jsonRow slice as JSON array

**Checkpoint**: JSON format working — output is valid JSON parseable by automation tools

---

## Phase 4: CSV Format (Spreadsheet Export)

**Purpose**: Tabular data for Excel/Google Sheets import

### User Story 3: CSV Output for Spreadsheets (P2)

- [x] T033 [P] [CSV] Define `csvHeader` slice with 28 column names (matching table format)
- [x] T034 [CSV] Implement `formatCSV(results []*checks.Result, w io.Writer)` function
- [x] T035 [CSV] Create `csv.NewWriter(w)` for automatic escaping
- [x] T036 [CSV] Write header row using csvHeader slice
- [x] T037 [CSV] Implement result iteration loop for data rows
- [x] T038 [CSV] Build row slice with FullName and staleStr for first 2 columns
- [x] T039 [CSV] Add boolean columns using `strconv.FormatBool` (true/false strings)
- [x] T040 [CSV] Add count columns using `strconv.Itoa` (integer strings)
- [x] T041 [CSV] Add tristate security columns using `tristate()` helper (✓/✗/? strings)
- [x] T042 [CSV] Write each row with `csv.Write(row)`
- [x] T043 [CSV] Call `csv.Flush()` and return `csv.Error()` for error checking

**Checkpoint**: CSV format working — output imports cleanly into spreadsheet software

---

## Phase 5: Markdown Format (Documentation)

**Purpose**: Markdown tables for GitHub README files and wikis

### User Story 4: Markdown Output for Documentation (P3)

- [x] T044 [Markdown] Implement `formatMD(results []*checks.Result, w io.Writer)` function
- [x] T045 [Markdown] Write header row with 28 pipe-separated column names
- [x] T046 [Markdown] Write alignment row with dashes and colons (right-align numeric columns)
- [x] T047 [Markdown] Implement result iteration loop for data rows
- [x] T048 [Markdown] Format each row with pipe separators and column values
- [x] T049 [Markdown] Use bool2check for boolean columns (✓/✗)
- [x] T050 [Markdown] Use tristate for security columns (✓/✗/?)
- [x] T051 [Markdown] Use staleStr for stale column (YES/NO)
- [x] T052 [Markdown] Use raw integers for count columns
- [x] T053 [Markdown] Return nil (no flush needed, writes directly)

**Checkpoint**: Markdown format working — tables render correctly in GitHub

---

## Phase 6: Testing Infrastructure

**Purpose**: Validate all format implementations

- [x] T054 [P] [Tests] Create `internal/formatter/formatter_test.go` with package declaration
- [x] T055 [P] [Tests] Implement `sampleResults()` helper returning test data with mixed pass/fail checks
- [x] T056 [P] [Tests] Include unknown security settings in sampleResults for tristate testing

### Format-Specific Tests

- [x] T057 [P] [Tests] Add `TestFormatTable` verifying header presence (REPO, CODE_CONDUCT, RULESETS, etc.)
- [x] T058 [P] [Tests] Verify table contains symbols (✓, ✗, ?)
- [x] T059 [P] [Tests] Verify table contains repository name
- [x] T060 [P] [Tests] Add `TestFormatJSON` verifying valid JSON output
- [x] T061 [P] [Tests] Parse JSON with `json.Unmarshal` to validate structure
- [x] T062 [P] [Tests] Verify JSON contains expected field names (snake_case)
- [x] T063 [P] [Tests] Add `TestFormatCSV` verifying header row present
- [x] T064 [P] [Tests] Parse CSV output with `csv.Reader` to validate format
- [x] T065 [P] [Tests] Verify CSV contains true/false strings for booleans
- [x] T066 [P] [Tests] Add `TestFormatMD` verifying pipe-separated structure
- [x] T067 [P] [Tests] Verify markdown contains alignment row with dashes
- [x] T068 [P] [Tests] Verify markdown contains symbols (✓, ✗, ?)

**Checkpoint**: All formats have test coverage — output correctness validated

---

## Phase 7: CLI Integration

**Purpose**: Connect formatter to CLI flags and file output

- [x] T069 [Integration] Verify `--format` flag defined in `cmd/gh-repo-health-report/main.go` with default "table"
- [x] T070 [Integration] Verify `--output` flag defined for file output (default stdout)
- [x] T071 [Integration] Verify CLI creates file writer when `--output` specified
- [x] T072 [Integration] Verify CLI defaults to stdout when `--output` not specified
- [x] T073 [Integration] Verify `formatter.Format()` called with results, format string, and writer
- [x] T074 [Integration] Verify errors from `formatter.Format()` propagated to user

**Checkpoint**: Formatter fully integrated — CLI flags control output format and destination

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Foundation)**: No dependencies — must be first
- **Phase 2 (Table)**: Depends on Phase 1 (uses helpers)
- **Phase 3 (JSON)**: Depends on Phase 1 (format routing only; no helpers needed)
- **Phase 4 (CSV)**: Depends on Phase 1 (uses helpers and routing)
- **Phase 5 (Markdown)**: Depends on Phase 1 (uses helpers and routing)
- **Phase 6 (Tests)**: Depends on Phases 2-5 (tests all formats)
- **Phase 7 (Integration)**: Depends on Phases 1-5 (CLI wiring)

### Within Phases

**Phase 1 (Foundation)**:
- T001-T003 sequential (file, function, switch)
- T004-T006 parallel (independent helpers)

**Phase 2 (Table)**:
- T007-T009 sequential (constants, function, tabwriter)
- T010-T020 sequential (build table row incrementally)
- T021 final (flush output)

**Phase 3 (JSON)**:
- T022-T024 parallel (struct definition, independent fields)
- T025-T028 sequential (toRow implementation, field mapping)
- T029-T032 sequential (formatJSON implementation)

**Phase 4 (CSV)**:
- T033-T035 sequential (header, function, writer)
- T036-T042 sequential (build CSV row incrementally)
- T043 final (flush and error check)

**Phase 5 (Markdown)**:
- T044-T046 sequential (function, header, alignment)
- T047-T052 sequential (row formatting)
- T053 final (return)

**Phase 6 (Tests)**:
- T054-T056 setup (parallel if multiple test files)
- T057-T068 all parallel (independent test cases)

**Phase 7 (Integration)**:
- T069-T074 verification tasks (all parallel, different concerns)

### Parallel Opportunities

**Maximum parallelism scenario**:
1. Phase 1: T001-T003 sequential, then T004-T006 parallel
2. After Phase 1: Phases 2-5 can all start in parallel (independent formats)
3. Phase 6: All test tasks parallel (T057-T068)
4. Phase 7: All integration verification parallel (T069-T074)

**Real development probably was**:
1. Table first (MVP default output)
2. JSON second (automation use case)
3. CSV third (spreadsheet export)
4. Markdown last (nice-to-have)
5. Tests added incrementally alongside each format

---

## Implementation Notes

### Actual Development Pattern

**Real-world implementation** likely evolved as:
1. Table format (MVP) — human-readable terminal output
2. Helper functions added when implementing table (bool2check, staleStr)
3. JSON format added for CI/CD use cases
4. jsonRow struct created to clean up JSON output
5. CSV format added for manager/stakeholder reporting
6. Markdown format added as final polish for documentation
7. Tests added per format to prevent regressions

### Key Design Decisions Made

**Decision 1**: Single entry point with format routing
- **Rationale**: Simple dispatch, easy to extend
- **Trade-off**: No format interface; harder to extend outside package

**Decision 2**: Helper functions for symbol mapping
- **Rationale**: DRY principle, consistent representation
- **Trade-off**: Small function call overhead (negligible)

**Decision 3**: Separate jsonRow struct
- **Rationale**: Clean JSON API, explicit field control
- **Trade-off**: Field duplication, maintenance burden

**Decision 4**: Mixed CSV representation (booleans vs. symbols)
- **Rationale**: Balance CSV best practices with human readability
- **Trade-off**: Inconsistent representation (confusing?)

### Testing Philosophy

**Format validation approach**:
- Each format test uses same sample data (consistency)
- Tests verify structure (headers, separators) not exact content
- Symbol presence tested (✓/✗/?) to ensure helper usage
- JSON/CSV parsing validation ensures valid output

**No integration tests**:
- Formatter tested in isolation (unit tests)
- CLI integration manually verified (no E2E tests)

### Gaps Identified During Migration

1. ⚠️ **No format validation** — Invalid format string defaults to table silently
2. ⚠️ **No column width control** — Long repo names cause wide tables
3. ⚠️ **Hardcoded JSON indentation** — No compact mode for automation
4. ⚠️ **No ASCII fallback** — Symbols (✓/✗/?) assume UTF-8 support
5. ⚠️ **Limited CSV test coverage** — No tests for special character escaping

---

## Total Implementation Effort

**Completed Tasks**: 74 tasks across 7 phases  
**Files Modified**: 2 files (`formatter.go`, `formatter_test.go`)  
**Lines of Code**: 432 total (255 implementation + 177 tests)  
**Formats Implemented**: 4 (table, JSON, CSV, markdown)  
**Helper Functions**: 3 (bool2check, tristate, staleStr)

---

## Maintenance Guidance

**Adding New Column**:
1. Update tableHeader constant (Phase 2)
2. Add fprintf parameter to formatTable (Phase 2)
3. Add field to jsonRow struct with json tag (Phase 3)
4. Update toRow() mapping (Phase 3)
5. Add to csvHeader slice (Phase 4)
6. Add to formatCSV row slice (Phase 4)
7. Add to formatMD header and data rows (Phase 5)
8. Update tests to check for new column

**Adding New Format**:
1. Add case to Format() switch (Phase 1)
2. Implement formatXXX() function (new phase)
3. Add TestFormatXXX test (Phase 6)
4. Update CLI flag help text
5. Update README with format examples

**Changing Symbol Representation**:
- Update helper functions (bool2check, tristate, staleStr)
- Changes propagate to all formats automatically
- Consider backward compatibility for scripts

**Performance Optimization**:
- Pre-allocate jsonRow slice: `rows := make([]jsonRow, len(results))`
- Use strings.Builder for CSV rows (micro-optimization)
- Buffer writes for file output (already done by os.File)
