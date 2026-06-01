# Feature Specification: Multi-Format Output

**Feature Branch**: N/A (existing feature)  
**Created**: 2026-06-01  
**Status**: migrated  
**Input**: Reverse-engineered from existing implementation in `internal/formatter/`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Table Output for Terminal (Priority: P1) 🎯 MVP

A user running health checks from the terminal wants human-readable table output with columns for all health checks, properly aligned and formatted with visual indicators (✓/✗/?) for quick scanning.

**Why this priority**: Default output format; most common use case for interactive terminal usage.

**Independent Test**: Run `gh repo-health-report --repo owner/name` (no format flag) and verify table output displays with headers and properly formatted columns.

**Acceptance Scenarios**:

1. **Given** health check results for repositories, **When** formatting with `--format table`, **Then** output displays as ASCII table with tab-separated columns
2. **Given** check passes, **When** rendering table, **Then** column shows "✓" for boolean checks
3. **Given** check fails, **When** rendering table, **Then** column shows "✗" for boolean checks
4. **Given** security setting is unknown (permission denied), **When** rendering table, **Then** column shows "?" for tristate checks
5. **Given** staleness check, **When** rendering table, **Then** displays "YES" if stale, "NO" if not stale
6. **Given** multiple repositories, **When** formatting table, **Then** all repos shown as rows with aligned columns
7. **Given** 28 columns of data, **When** rendering table, **Then** all columns fit in wide terminal with proper spacing

---

### User Story 2 - JSON Output for Automation (Priority: P1) 🎯 MVP

A CI/CD pipeline operator wants machine-readable JSON output with all health check data as boolean fields and counts, enabling automated parsing and reporting.

**Why this priority**: Essential for programmatic consumption; enables integration with dashboards and alerting systems.

**Independent Test**: Run `gh repo-health-report --repo owner/name --format json` and verify valid JSON array is output with all fields.

**Acceptance Scenarios**:

1. **Given** health check results, **When** formatting with `--format json`, **Then** output is valid JSON array
2. **Given** JSON output, **When** parsed, **Then** each repository is an object with boolean fields for all checks
3. **Given** security settings, **When** formatting JSON, **Then** includes both `*_enabled` and `*_unknown` fields (tristate representation)
4. **Given** counts (topics, branches, tags, etc.), **When** formatting JSON, **Then** includes integer fields for all counts
5. **Given** JSON output, **When** formatted, **Then** pretty-printed with 2-space indentation for readability
6. **Given** stale status, **When** formatting JSON, **Then** represented as boolean `stale: true/false`

---

### User Story 3 - CSV Output for Spreadsheets (Priority: P2)

A manager wants CSV-formatted output for importing into Excel or Google Sheets to generate reports, charts, and share with non-technical stakeholders.

**Why this priority**: Important for reporting but not needed for basic CLI usage; CSV is secondary to table/JSON.

**Independent Test**: Run `gh repo-health-report --org myorg --format csv --output report.csv` and verify CSV file can be opened in Excel with proper columns.

**Acceptance Scenarios**:

1. **Given** health check results, **When** formatting with `--format csv`, **Then** output includes CSV header row with all column names
2. **Given** CSV output, **When** imported to spreadsheet, **Then** boolean checks shown as "true"/"false" strings
3. **Given** CSV output, **When** imported to spreadsheet, **Then** stale status shown as "YES"/"NO" for readability
4. **Given** CSV output, **When** imported to spreadsheet, **Then** tristate security checks shown as "✓"/"✗"/"?" matching table format
5. **Given** `--output filename.csv` flag, **When** generating CSV, **Then** output written to file instead of stdout
6. **Given** CSV with special characters, **When** formatting, **Then** fields properly escaped according to CSV spec

---

### User Story 4 - Markdown Output for Documentation (Priority: P3)

A documentation maintainer wants Markdown-formatted tables for including repository health reports in README files, wikis, or documentation sites.

**Why this priority**: Nice-to-have for documentation workflows; less common than other formats.

**Independent Test**: Run `gh repo-health-report --org myorg --format md` and verify Markdown table syntax is correct with proper alignment.

**Acceptance Scenarios**:

1. **Given** health check results, **When** formatting with `--format md`, **Then** output is valid GitHub-flavored Markdown table
2. **Given** Markdown output, **When** rendered, **Then** includes header row with column names separated by pipes
3. **Given** Markdown output, **When** rendered, **Then** includes alignment row with colons for right-aligned numeric columns
4. **Given** Markdown output, **When** rendered, **Then** boolean checks shown as ✓/✗/? matching table format
5. **Given** Markdown output, **When** rendered in GitHub, **Then** table displays properly with aligned columns

---

### Edge Cases

- What happens when repository name contains special CSV characters (commas, quotes)? → Proper CSV escaping applied by `csv.Writer`
- How are Unicode characters (✓, ✗) handled in different terminal encodings? → Relies on terminal UTF-8 support; no fallback to ASCII
- What happens with very long repository names in table format? → No truncation; table columns expand; may wrap on narrow terminals
- How is output written to files vs. stdout? → Same formatting logic; `--output` flag determines writer (file or stdout)
- What happens when formatter receives empty results slice? → Headers printed, no data rows (valid but empty output)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support 4 output formats: `table`, `json`, `csv`, `md`
- **FR-002**: System MUST default to `table` format when `--format` flag not specified
- **FR-003**: Table format MUST use tab-separated columns with `tabwriter` for alignment
- **FR-004**: Table format MUST include header row with all 28 column names
- **FR-005**: Table format MUST use "✓" for passing boolean checks, "✗" for failing checks
- **FR-006**: Table format MUST use "?" for unknown security settings (tristate handling)
- **FR-007**: Table format MUST use "YES"/"NO" for stale status (not ✓/✗)
- **FR-008**: JSON format MUST output array of objects with one object per repository
- **FR-009**: JSON format MUST use boolean fields for all checks (not strings)
- **FR-010**: JSON format MUST include both `*_enabled` and `*_unknown` fields for security settings
- **FR-011**: JSON format MUST use 2-space indentation for pretty-printing
- **FR-012**: CSV format MUST include header row with column names matching table headers
- **FR-013**: CSV format MUST use `true`/`false` strings for most boolean fields
- **FR-014**: CSV format MUST use "YES"/"NO" strings for stale status (human-readable)
- **FR-015**: CSV format MUST use "✓"/"✗"/"?" for tristate security settings
- **FR-016**: CSV format MUST properly escape special characters (commas, quotes, newlines)
- **FR-017**: Markdown format MUST generate GitHub-flavored Markdown table syntax
- **FR-018**: Markdown format MUST include alignment row with colons for numeric columns
- **FR-019**: Markdown format MUST use ✓/✗/? for checks (matching table format)
- **FR-020**: All formats MUST display all 28 data columns for each repository
- **FR-021**: Formatter MUST accept `io.Writer` for output destination (supports stdout and files)
- **FR-022**: Formatter MUST return error if formatting fails (write errors, invalid data)

### Key Entities *(include if feature involves data)*

- **Result**: Input from checks package (one per repository)
  - Boolean fields for all checks
  - Counts (branches, tags, topics, open issues, size)
  - Tristate security fields (Enabled + Unknown flags)
  - Repository reference for name/metadata

- **jsonRow**: Internal struct for JSON serialization
  - Maps Result fields to JSON-tagged fields
  - Separate struct for clean JSON output (no unexported fields)

### CLI Interface Requirements *(for CLI features)*

- **CLI-001**: `--format` flag accepts: `table`, `json`, `csv`, `md`
- **CLI-002**: Default format is `table` when flag not specified
- **CLI-003**: `--output` flag specifies file path for output (optional)
- **CLI-004**: Output defaults to stdout when `--output` not specified
- **CLI-005**: Error messages for invalid format values (unknown format string)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Table output renders correctly in terminals with width ≥120 columns
- **SC-002**: JSON output is valid and parseable by `jq`, `python json.loads()`, and `JSON.parse()`
- **SC-003**: CSV output imports correctly into Excel and Google Sheets with proper column types
- **SC-004**: Markdown output renders properly in GitHub README files and wikis
- **SC-005**: All formats include all 28 data columns without loss of information
- **SC-006**: Tristate security settings correctly represented in all formats (?, unknown flag, etc.)

## Assumptions

- Terminal supports UTF-8 encoding for ✓/✗/? characters
- Users have terminals wide enough for 28-column table (≥120 columns recommended)
- CSV consumers understand both boolean (`true`/`false`) and symbolic (✓/✗/?) representations
- Markdown renderers support GitHub-flavored Markdown table syntax
- File output destination is writable (no permission checks in formatter)
- go standard library `encoding/json` and `encoding/csv` packages provide sufficient functionality
- tabwriter package handles complex alignment scenarios correctly
