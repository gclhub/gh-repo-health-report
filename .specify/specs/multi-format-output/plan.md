# Implementation Plan: Multi-Format Output

**Status**: migrated (reverse-engineered from existing implementation)  
**Input**: Analysis of `internal/formatter/formatter.go` and `internal/formatter/formatter_test.go`

## Technical Context

### Technology Stack (Actual Implementation)

- **Language**: Go 1.x
- **Package**: `internal/formatter` — Output formatting layer
- **Dependencies**: 
  - `encoding/json` — JSON marshaling with indentation
  - `encoding/csv` — CSV writing with proper escaping
  - `text/tabwriter` — Aligned table output
  - `io` — Writer abstraction for output
  - `fmt` — String formatting
  - `strconv` — Integer/boolean conversion
- **Testing**: Format-specific tests with output verification

### Project Structure

```
internal/formatter/
├── formatter.go       # 255 lines - all format implementations
└── formatter_test.go  # 177 lines - format validation tests
```

### Output Format Breakdown

**Table Format** (55 lines):
- Tab-separated columns with tabwriter for alignment
- Header row with 28 column names
- ✓/✗/? symbols for visual scanning

**JSON Format** (35 lines):
- Array of objects with typed fields
- Pretty-printed with 2-space indentation
- Separate `jsonRow` struct for clean serialization

**CSV Format** (45 lines):
- Header row + data rows
- Mixed representation (booleans as strings, symbols for tristate)
- Automatic escaping via `csv.Writer`

**Markdown Format** (35 lines):
- GitHub-flavored Markdown table syntax
- Alignment row for right-aligned numeric columns
- Pipe-separated with proper spacing

## Implementation Approach (Actual)

### Phase 1: Format Router & Helper Functions

**Main entry point**:
```go
func Format(results []*checks.Result, format string, w io.Writer) error {
    switch format {
    case "json":
        return formatJSON(results, w)
    case "csv":
        return formatCSV(results, w)
    case "md":
        return formatMD(results, w)
    default:
        return formatTable(results, w)
    }
}
```

**Helper functions** (used across formats):
```go
// Boolean to check mark
func bool2check(v bool) string {
    if v { return "✓" }
    return "✗"
}

// Tristate: enabled/unknown/disabled → ✓/?/✗
func tristate(ok, unknown bool) string {
    if unknown { return "?" }
    return bool2check(ok)
}

// Stale status: true/false → YES/NO
func staleStr(v bool) string {
    if v { return "YES" }
    return "NO"
}
```

**Design Decision**: Single entry point with format routing. Helper functions provide consistent symbol mapping across formats.

### Phase 2: Table Format (User Story 1)

**Implementation**:
```go
func formatTable(results []*checks.Result, w io.Writer) error {
    tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
    fmt.Fprintln(tw, tableHeader)  // 28 tab-separated column names
    for _, r := range results {
        fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t%s\t...\n",
            r.Repository.FullName,
            staleStr(r.Stale),
            bool2check(r.HasDescription),
            r.TopicsCount,
            bool2check(r.HasReadme),
            bool2check(r.HasLicense),
            // ... 22 more columns
        )
    }
    return tw.Flush()
}
```

**Column Order** (28 columns):
1. REPO (full name)
2. STALE (YES/NO)
3. DESCRIPTION (✓/✗)
4. TOPICS (count)
5. README, LICENSE, CODE_CONDUCT, CODEOWNERS, SECURITY, CONTRIBUTING (✓/✗)
6. ISSUE_TMPL, PR_TMPL (✓/✗)
7. ISSUES, WIKI, PROJECTS (✓/✗)
8. DEPENDABOT, CI (✓/✗)
9. BR_PROTECT, RULESETS (✓/✗)
10. VULN_ALERTS, SECRET_SCAN, PUSH_PROT (✓/✗/?)
11. AUTO_DEL_BR (✓/✗)
12. BRANCHES, STALE_BR, TAGS (counts)
13. OPEN_ISSUES, SIZE_KB (counts)

**Design Decisions**:
- tabwriter handles alignment automatically (column width adapts)
- 2-space minimum spacing between columns
- No truncation of long repository names (may wrap on narrow terminals)
- Helper functions for consistent symbol usage

### Phase 3: JSON Format (User Story 2)

**Internal struct** (for clean JSON):
```go
type jsonRow struct {
    Repo                       string `json:"repo"`
    Stale                      bool   `json:"stale"`
    Description                bool   `json:"has_description"`
    Topics                     int    `json:"topics_count"`
    // ... 20+ more fields with json tags
    VulnerabilityAlertsEnabled bool   `json:"vulnerability_alerts_enabled"`
    VulnerabilityAlertsUnknown bool   `json:"vulnerability_alerts_unknown"`
    // ... tristate pairs for security settings
}
```

**Conversion function**:
```go
func toRow(r *checks.Result) jsonRow {
    return jsonRow{
        Repo:        r.Repository.FullName,
        Stale:       r.Stale,
        Description: r.HasDescription,
        // ... direct field mapping
    }
}
```

**Output function**:
```go
func formatJSON(results []*checks.Result, w io.Writer) error {
    rows := make([]jsonRow, len(results))
    for i, r := range results {
        rows[i] = toRow(r)
    }
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")  // 2-space indentation
    return enc.Encode(rows)
}
```

**Design Decisions**:
- Separate struct avoids exporting internal `checks.Result` structure
- Snake_case JSON field names (convention for APIs)
- Tristate represented as two boolean fields (explicit `*_unknown` flag)
- Pretty-printing for human readability (could be compact for production)

### Phase 4: CSV Format (User Story 3)

**Header definition**:
```go
var csvHeader = []string{
    "REPO", "STALE", "DESCRIPTION", "TOPICS", "README", "LICENSE",
    // ... 22 more column names (matches table header)
}
```

**Output function**:
```go
func formatCSV(results []*checks.Result, w io.Writer) error {
    cw := csv.NewWriter(w)
    if err := cw.Write(csvHeader); err != nil {
        return err
    }
    for _, r := range results {
        row := []string{
            r.Repository.FullName,
            staleStr(r.Stale),                          // YES/NO
            strconv.FormatBool(r.HasDescription),       // true/false
            strconv.Itoa(r.TopicsCount),                // integer
            strconv.FormatBool(r.HasReadme),            // true/false
            // ...
            tristate(r.VulnerabilityAlertsEnabled, r.VulnerabilityAlertsUnknown),  // ✓/✗/?
            // ...
        }
        if err := cw.Write(row); err != nil {
            return err
        }
    }
    cw.Flush()
    return cw.Error()
}
```

**Design Decisions**:
- Mixed representation strategy:
  - Most booleans: `true`/`false` strings (native CSV type)
  - Stale status: `YES`/`NO` (human-readable)
  - Tristate security: `✓`/`✗`/`?` (visual consistency with table)
- `csv.Writer` handles escaping automatically (commas, quotes, newlines)
- Header row uses uppercase names (matches table format)

### Phase 5: Markdown Format (User Story 4)

**Output function**:
```go
func formatMD(results []*checks.Result, w io.Writer) error {
    // Header row
    fmt.Fprintln(w, "| REPO | STALE | DESCRIPTION | ... |")
    
    // Alignment row (colons for right-align on numeric columns)
    fmt.Fprintln(w, "|------|-------|-------------|...|")
    
    // Data rows
    for _, r := range results {
        fmt.Fprintf(w, "| %s | %s | %s | %d | %s | ...\n",
            r.Repository.FullName,
            staleStr(r.Stale),
            bool2check(r.HasDescription),
            r.TopicsCount,
            bool2check(r.HasReadme),
            // ... 23 more columns
        )
    }
    return nil
}
```

**Alignment row**:
- Left-align (default): `|------|`
- Right-align (numeric): `|-----:|`
- Used for: TOPICS, BRANCHES, STALE_BR, TAGS, OPEN_ISSUES, SIZE_KB

**Design Decisions**:
- GitHub-flavored Markdown (most common)
- Right-align numeric columns for better readability
- Symbols (✓/✗/?) work in Markdown without escaping
- Spacing around pipes for visual consistency in source

## Architecture Decisions

### Decision 1: Format Routing with Switch
**Rationale**: Simple dispatch mechanism. Easy to add new formats.

**Trade-offs**:
- ✅ Clear entry point
- ✅ Easy to test each format independently
- ❌ No format interface (harder to extend outside package)
- ❌ Default case hides invalid format strings (could return error instead)

### Decision 2: Helper Functions for Symbol Mapping
**Rationale**: Consistent representation across formats. Centralized logic for ✓/✗/? mapping.

**Trade-offs**:
- ✅ DRY (Don't Repeat Yourself)
- ✅ Easy to change symbol representation (single location)
- ✅ Self-documenting (function names describe intent)
- ❌ Small overhead for function calls (negligible in practice)

### Decision 3: Separate jsonRow Struct
**Rationale**: Clean JSON output without exposing internal structure. Explicit control over field names and tags.

**Trade-offs**:
- ✅ API stability (internal `Result` struct can change)
- ✅ Snake_case JSON convention
- ❌ Duplication (mapping fields manually)
- ❌ Maintenance burden (add field in two places)

### Decision 4: Mixed CSV Representation
**Rationale**: Balance between CSV best practices (booleans as true/false) and human readability (YES/NO, symbols).

**Trade-offs**:
- ✅ Spreadsheet software understands true/false
- ✅ YES/NO more readable than true/false for stale
- ✅ Symbols (✓/✗/?) visually consistent with table
- ❌ Inconsistent representation (confusing?)
- ❌ Symbols may not display in all CSV viewers

## Complexity Assessment

**Cyclomatic Complexity**: Low
- Simple iteration over results
- No complex logic (mostly formatting)
- Error handling is straightforward

**Lines of Code**: 255 lines (implementation) + 177 lines (tests) = 432 total

**Dependencies**: Standard library only
- No external formatting libraries
- Leverages built-in `tabwriter`, `json.Encoder`, `csv.Writer`

**Test Coverage**: Comprehensive
- Tests for all 4 formats
- Verifies headers present
- Checks symbol usage (✓/✗/?)
- JSON parse validation

## Known Gaps & Limitations

### Gap 1: No Format Validation
**Issue**: Invalid format string falls through to default (table). No error returned.

**Impact**: Typo in `--format` flag silently ignored (user expects JSON, gets table).

**Potential Fix**: Return error for unknown format strings instead of defaulting to table.

### Gap 2: No Column Width Control
**Issue**: Table format relies on tabwriter auto-sizing. Very long repo names cause wide columns.

**Impact**: Tables may exceed terminal width and wrap awkwardly.

**Potential Fix**: Add optional truncation or max width configuration.

### Gap 3: Hardcoded JSON Indentation
**Issue**: Always uses 2-space indentation. No compact mode.

**Impact**: Larger file sizes when compact JSON preferred (e.g., piping to `jq`).

**Potential Fix**: Add `--compact` flag to disable indentation.

### Gap 4: No Locale Support
**Issue**: Symbols (✓/✗/?) assume UTF-8 encoding. No ASCII fallback.

**Impact**: Broken display on terminals without UTF-8 support (rare in 2024).

**Potential Fix**: Add `--ascii` flag to use `[Y]/[N]/[?]` instead of Unicode symbols.

### Gap 5: Markdown Alignment Hardcoded
**Issue**: Numeric column alignment hardcoded in formatMD. No flexibility.

**Impact**: Minor; alignment is cosmetic.

**Not a real issue**: Markdown rendering handles it fine.

## Testing Strategy (Actual Implementation)

### Test Structure

**Sample Data Helper**:
```go
func sampleResults() []*checks.Result {
    // Returns one result with mixed pass/fail checks
    // Includes unknown security settings for tristate testing
}
```

### Test Cases

**Table Format** (`TestFormatTable`):
- Verifies output contains all headers (REPO, CODE_CONDUCT, RULESETS, etc.)
- Checks for symbols (✓, ✗, ?)
- Validates repository name present

**JSON Format** (`TestFormatJSON`):
- Verifies valid JSON array
- Parses output with `json.Unmarshal`
- Checks field presence and types

**CSV Format** (`TestFormatCSV`):
- Verifies header row present
- Checks for true/false strings
- Validates CSV parsing via `csv.Reader`

**Markdown Format** (`TestFormatMD`):
- Verifies pipe-separated structure
- Checks for alignment row
- Validates symbols present

**Pattern**: Each test uses `sampleResults()` → format to buffer → validate output structure.

## Integration Points

### Input: Checks Package
- Receives `[]*checks.Result` slice from `checks.Evaluate()`
- All check evaluation complete before formatting

### Output: CLI / Files
- Writes to `io.Writer` (stdout or file)
- CLI handles `--output` flag and creates file writer
- Formatter agnostic to output destination

### CLI Integration
- CLI parses `--format` flag, defaults to "table"
- CLI calls `formatter.Format(results, format, writer)`
- CLI handles errors (write failures, etc.)

## Performance Considerations

**Time Complexity**: O(n × c) for n repos and c columns (28)
- Single pass through results
- No sorting or complex transformations

**Space Complexity**: O(n) for n repositories
- JSON: Allocates jsonRow slice (temporary)
- CSV: Builds row strings (temporary)
- Table/Markdown: Writes directly to output (no buffering)

**Bottleneck**: IO writing, not formatting logic
- Formatting is fast (string concatenation)
- Writing to files or network may be slow

**Scalability**: Can format 1000+ repositories efficiently
- Memory usage: ~1KB per result
- No algorithmic bottlenecks

## Maintenance Notes

**Adding New Column**:
1. Update table header constant (add column name)
2. Add field to formatTable fprintf (add value)
3. Add field to jsonRow struct (json tag)
4. Update toRow() mapping
5. Add to csvHeader slice
6. Add to formatCSV row slice
7. Add to formatMD header and alignment rows
8. Update tests to verify new column

**Adding New Format**:
1. Add case to Format() switch
2. Implement formatXXX() function
3. Add test TestFormatXXX
4. Update CLI flag documentation
5. Update help text

**Changing Symbol Representation**:
- Update helper functions (bool2check, tristate, staleStr)
- All formats use helpers; change propagates automatically
- Consider impact on existing scripts parsing output

**Performance Optimization**:
- Pre-allocate result slices (jsonRow array)
- Use strings.Builder for CSV row construction
- Buffer writes for file output (already done by default)
