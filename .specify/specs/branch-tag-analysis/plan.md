# Implementation Plan: Branch and Tag Analysis

**Feature**: Branch and Tag Analysis  
**Status**: migrated  
**Created**: 2026-06-01

## Technical Context

### Existing Architecture

This feature adds branch/tag counting and staleness analysis to the health check system. It requires pagination handling and per-branch API queries, making it one of the more API-intensive features.

**Key packages**:
- `internal/api/client.go`: GitHub API client with pagination logic
- `internal/checks/checks.go`: Threshold evaluation logic
- `cmd/gh-repo-health-report/main.go`: CLI flags for thresholds

**Dependencies**:
- `github.com/cli/go-gh/v2/pkg/api`: GitHub CLI REST client
- Standard library `time` package for date handling

### Integration Points

1. **Repository struct** (internal/api/client.go):
   - Added BranchCount, StaleBranchCount, TagCount fields (lines 73-75)
   - Populated by PopulateBranchTagChecks method

2. **Options struct** (internal/checks/checks.go):
   - Added MaxBranches and MaxTags threshold fields (lines 50-53)
   - Since duration for staleness cutoff (already existed)

3. **PopulateBranchTagChecks** (internal/api/client.go:365-429):
   - Pagination loop for branches (per_page=100)
   - Per-branch commit query for staleness detection
   - Pagination loop for tags (per_page=100)

4. **Check evaluation** (internal/checks/checks.go:209-227):
   - Threshold comparison with default fallbacks
   - FailedChecks population based on counts

### Technical Decisions

#### Decision 1: Pagination Strategy

**Chosen approach**: Loop until response length < 100, using page number increment.

**Rationale**:
- GitHub API returns max 100 items per page
- Simple loop condition: `if len(items) < 100 { break }`
- Page number increment is straightforward: `page++`
- Handles edge case of exactly 100*N items (final empty page breaks loop)

**Implementation**:
```go
page := 1
for {
    var branches []branchItem
    path := fmt.Sprintf("repos/%s/%s/branches?per_page=100&page=%d", owner, name, page)
    if err := c.rest.Get(path, &branches); err != nil {
        return err
    }
    repo.BranchCount += len(branches)
    // ... process branches ...
    if len(branches) < 100 {
        break  // Last page
    }
    page++
}
```

**Alternatives considered**:
- Link header parsing: More complex, requires header parsing logic
- Cursor-based pagination: Not supported by branches/tags endpoints
- Fetch all pages in parallel: Would complicate error handling and ordering

---

#### Decision 2: Default Branch Exclusion from Staleness

**Chosen approach**: Explicitly skip default branch in staleness loop with `if b.Name == repo.DefaultBranch { continue }`.

**Rationale**:
- Default branch represents "current" state even if no recent commits
- Organization might have long-lived stable branches without constant activity
- Avoids false positive when repo is intentionally in maintenance mode
- BranchCount includes default branch, but StaleBranchCount does not

**Implementation**:
```go
for _, b := range branches {
    if b.Name == repo.DefaultBranch {
        continue  // Exclude default branch from staleness analysis
    }
    // ... check for recent commits ...
}
```

**Alternatives considered**:
- Include default branch: Would flag maintenance-mode repos incorrectly
- Separate check for default branch staleness: Adds complexity without clear value
- Make it configurable: Over-engineering for simple exclusion rule

---

#### Decision 3: Per-Branch Commit Query with `since` Filter

**Chosen approach**: Query commits with `since={cutoff}` timestamp and check if array is empty.

**Rationale**:
- Efficient: Only fetches 1 commit per branch (`per_page=1`)
- Correct semantics: `since` parameter filters by author date
- Empty array means no commits after cutoff → stale
- Non-empty array means at least one commit after cutoff → active

**Implementation**:
```go
sinceStr := since.UTC().Format(time.RFC3339)
var commits []interface{}
cpath := fmt.Sprintf(
    "repos/%s/%s/commits?sha=%s&since=%s&per_page=1",
    owner, name, b.Name, sinceStr,
)
if err := c.rest.Get(cpath, &commits); err != nil {
    continue  // Skip on error
}
if len(commits) == 0 {
    repo.StaleBranchCount++
}
```

**Alternatives considered**:
- Fetch branch metadata (updated_at): Doesn't exist in branches API response
- Fetch single commit and compare timestamp: Requires parsing commit date, more complex
- Use GraphQL for batch query: Would require separate client setup

---

#### Decision 4: Graceful Error Handling in Staleness Loop

**Chosen approach**: `continue` on error instead of `return err`, accepting under-counted staleness.

**Rationale**:
- Transient errors (rate limits, deleted branches) shouldn't abort entire report
- BranchCount is still accurate (from initial branches query)
- StaleBranchCount might be slightly under-counted, but better than no data
- Prioritizes partial results over all-or-nothing failure

**Implementation**:
```go
if err := c.rest.Get(cpath, &commits); err != nil {
    // On transient errors (rate limit, deleted branch race, etc.)
    // skip the branch rather than aborting the whole report.
    continue
}
```

**Alternatives considered**:
- Return error: Would abort entire report on single branch failure
- Retry logic: Adds complexity; rate limit errors need backoff strategy
- Collect errors and report: Adds complexity without clear benefit

---

#### Decision 5: Zero Threshold Means Use Default

**Chosen approach**: Check `if opts.MaxBranches == 0` and substitute DefaultMaxBranches.

**Rationale**:
- Zero is natural "unset" value for int fields
- Allows CLI flags to be optional (default to 0)
- Explicit defaults (50 for branches, 100 for tags) are documented constants
- Prevents need for pointer types or sentinel values

**Implementation**:
```go
maxBranches := opts.MaxBranches
if maxBranches == 0 {
    maxBranches = DefaultMaxBranches  // 50
}
if r.BranchCount > maxBranches {
    r.FailedChecks = append(r.FailedChecks, CheckTooManyBranches)
}
```

**Alternatives considered**:
- Use `-1` as unset: Less intuitive, requires validation
- Always require thresholds: Reduces usability (most users want defaults)
- Use pointer types: Complicates JSON serialization and test fixtures

---

#### Decision 6: branchItem Struct for JSON Parsing

**Chosen approach**: Define minimal `branchItem` struct with only `Name` field needed.

**Rationale**:
- GitHub branches API returns many fields; we only need name
- Lightweight struct reduces memory footprint
- Clear intent: we're only extracting names for iteration
- Follows Go convention of defining minimal types for JSON unmarshaling

**Implementation**:
```go
// lines 10-13 in client.go
type branchItem struct {
    Name string `json:"name"`
}
```

**Alternatives considered**:
- Use map[string]interface{}: Loses type safety, requires casting
- Define full Branch struct: Wastes memory on unused fields
- Use []string with custom unmarshaling: Over-engineering for simple case

---

## Project Structure

### Modified Files

**internal/api/client.go** (~453 lines):
- Lines 10-13: branchItem struct definition
- Lines 73-75: BranchCount, StaleBranchCount, TagCount fields in Repository
- Lines 365-429: PopulateBranchTagChecks() implementation
  - Lines 374-411: Branch pagination and staleness detection
  - Lines 413-427: Tag pagination

**internal/checks/checks.go** (~230 lines):
- Lines 34-37: Branch/tag check name constants
- Lines 40-47: Default threshold constants (DefaultMaxBranches, DefaultMaxTags)
- Lines 50-53: Options struct with MaxBranches and MaxTags fields
- Lines 88-90: BranchCount, StaleBranchCount, TagCount fields in Result struct
- Lines 126-128: Populate result fields from repository
- Lines 209-227: Threshold evaluation logic

**internal/checks/checks_test.go** (~343 lines):
- Lines 38-40: Branch/tag fields in baseRepo() fixture
- Lines 260-272: TestEvaluate_TooManyBranches
- Lines 274-283: TestEvaluate_BranchCountWithinLimit
- Lines 285-297: TestEvaluate_StaleBranches
- Lines 299-311: TestEvaluate_TooManyTags
- Lines 313-322: TestEvaluate_TagCountWithinLimit
- Lines 324-334: TestEvaluate_DefaultBranchCountThresholds

**internal/api/client_test.go** (~254 lines):
- Lines 120-165: TestPopulateBranchTagChecks_Mock (mock HTTP test)

### File Organization

```
internal/
  api/
    client.go          # PopulateBranchTagChecks implementation
    client_test.go     # Mock tests for pagination and staleness
  checks/
    checks.go          # Threshold evaluation logic
    checks_test.go     # Threshold and staleness tests
```

## Implementation Phases

### Phase 1: Repository Struct Extensions ✅ COMPLETED

**Implemented**:
- Added BranchCount, StaleBranchCount, TagCount int fields
- Added branchItem struct for JSON parsing
- Fields populated by PopulateBranchTagChecks

**Lines of code**: ~10 lines (struct definitions)

**Test coverage**: Verified via check evaluation tests

---

### Phase 2: Pagination Logic for Branches ✅ COMPLETED

**Implemented**:
- Branch listing with `per_page=100` pagination
- Loop until response length < 100
- Increment BranchCount for each page
- Extract branch names for staleness checking

**Lines of code**: ~40 lines (pagination loop + staleness setup)

**Key functions**:
- `PopulateBranchTagChecks(repo *Repository, since time.Time) error`
- Uses existing `rest.Get()` from go-gh client
- Handles pagination state with page counter

**Error handling**:
- HTTP errors on branches query return error (abort check)
- Malformed responses caught by unmarshaling errors

---

### Phase 3: Stale Branch Detection ✅ COMPLETED

**Implemented**:
- Per-branch commit query with `since` filter
- Default branch exclusion logic
- Empty commits array → increment StaleBranchCount
- Graceful error handling (continue on failure)

**Lines of code**: ~25 lines (staleness loop)

**Key logic**:
```go
if b.Name == repo.DefaultBranch {
    continue
}
var commits []interface{}
// ... query with since filter ...
if len(commits) == 0 {
    repo.StaleBranchCount++
}
```

**Design principle**: Partial results better than complete failure

---

### Phase 4: Pagination Logic for Tags ✅ COMPLETED

**Implemented**:
- Tag listing with `per_page=100` pagination
- Loop until response length < 100
- Increment TagCount for each page
- Uses generic interface{} slice (don't need tag details)

**Lines of code**: ~15 lines (pagination loop)

**Key implementation**:
```go
page = 1
for {
    var tags []interface{}
    // ... fetch and count ...
    repo.TagCount += len(tags)
    if len(tags) < 100 {
        break
    }
    page++
}
```

---

### Phase 5: Options and Default Thresholds ✅ COMPLETED

**Implemented**:
- DefaultMaxBranches = 50 constant
- DefaultMaxTags = 100 constant
- MaxBranches and MaxTags fields in Options struct
- Zero-value means use default

**Lines of code**: ~10 lines (constants + struct fields)

**Usage**:
- CLI flags `--max-branches` and `--max-tags` populate Options
- Evaluate() function checks `if opts.MaxBranches == 0`

---

### Phase 6: Check Evaluation Logic ✅ COMPLETED

**Implemented**:
- Check name constants (3 new checks)
- Result struct fields populated from Repository
- Threshold comparison with default fallback
- FailedChecks population

**Lines of code**: ~20 lines (check logic)

**Key logic**:
```go
maxBranches := opts.MaxBranches
if maxBranches == 0 {
    maxBranches = DefaultMaxBranches
}
if r.BranchCount > maxBranches {
    r.FailedChecks = append(r.FailedChecks, CheckTooManyBranches)
}
```

---

### Phase 7: Test Coverage ✅ COMPLETED

**Implemented**:
- TestEvaluate_TooManyBranches: Threshold exceeded
- TestEvaluate_BranchCountWithinLimit: Within threshold
- TestEvaluate_StaleBranches: Stale branch detection
- TestEvaluate_TooManyTags: Tag threshold exceeded
- TestEvaluate_TagCountWithinLimit: Tag within threshold
- TestEvaluate_DefaultBranchCountThresholds: Default threshold usage
- TestPopulateBranchTagChecks_Mock: Mock API pagination test

**Lines of code**: ~90 lines (test code)

**Test strategy**:
- Unit tests with baseRepo() fixture
- Table-driven tests for different thresholds
- Mock HTTP tests for pagination behavior

---

## Validation Approach

### Automated Testing

**Unit tests** (internal/checks/checks_test.go):
- ✅ Too many branches detection
- ✅ Branch count within limit (no failure)
- ✅ Stale branch detection
- ✅ Too many tags detection
- ✅ Tag count within limit (no failure)
- ✅ Default thresholds used when MaxBranches=0

**Mock tests** (internal/api/client_test.go):
- ✅ Branch pagination (2 branches returned)
- ✅ Stale branch detection (empty commits array)
- ✅ Tag pagination (2 tags returned)

### Manual Testing

**Test commands**:
```bash
# Test default thresholds
go build && ./gh-repo-health-report --repo owner/repo

# Test custom thresholds
go build && ./gh-repo-health-report --repo owner/repo --max-branches 30 --max-tags 50

# Test staleness cutoff
go build && ./gh-repo-health-report --repo owner/repo --since 90d

# Test large repo with many branches
go build && ./gh-repo-health-report --repo kubernetes/kubernetes
```

**Expected outcomes**:
- Repos with > 50 branches → CheckTooManyBranches appears
- Repos with > 100 tags → CheckTooManyTags appears
- Stale branches counted correctly (default branch excluded)
- Custom thresholds override defaults

### Quality Gates

All checks passed:
- ✅ `go build ./...` — builds successfully
- ✅ `go test ./...` — all tests pass
- ✅ `go vet ./...` — no warnings
- ✅ Integration with existing formatter (displays counts and staleness)

---

## Complexity Assessment

### Code Metrics

| Component | Files | Lines | Functions | Test Lines |
|-----------|-------|-------|-----------|------------|
| API Integration | 1 | ~65 | 1 main | ~45 (mocks) |
| Check Evaluation | 1 | ~30 | 1 main | ~90 |
| Struct Definitions | 2 | ~20 | 0 | 0 |
| **Total** | **4** | **~115** | **2** | **~135** |

### Complexity Factors

**Low complexity**:
- Pagination is simple loop-until-short-page pattern
- Threshold comparison is straightforward boolean logic
- Default fallback is simple zero-check

**Medium complexity**:
- Per-branch commit query requires nested loop
- Staleness determination requires date formatting and API query
- Error handling in staleness loop (continue vs. return)

**High complexity**:
- Performance impact of N+1 queries (branches list + N commit queries)
- Rate limit implications for large repos
- Pagination boundary cases (exactly 100*N items)

**Overall**: Medium complexity feature with performance trade-offs documented as constraints.

---

## Performance Characteristics

### API Call Analysis

**Branch/tag counting**:
- Repos with ≤100 branches → 1 API call
- Repos with 101-200 branches → 2 API calls
- Repos with 201-300 branches → 3 API calls
- Pattern: ⌈BranchCount / 100⌉ API calls

**Staleness detection**:
- Additional API call for each non-default branch
- Repo with 50 branches → 49 extra calls
- Repo with 200 branches → 199 extra calls

**Total cost example** (repo with 75 branches, default branch):
- 1 call for branches list
- 74 calls for staleness detection
- **75 API calls total**

### Rate Limit Impact

GitHub API rate limits:
- Authenticated: 5000 requests/hour
- Organization scan with 50 repos × 30 branches each = ~1500 API calls
- Consumes ~30% of hourly rate limit

**Mitigation**: Feature is intentionally expensive; documented as known constraint.

---

## Risk Mitigation

### Identified Risks

1. **Rate limit exhaustion**: Mitigated by documentation; consider adding `--skip-stale-branches` flag
2. **Pagination errors**: Handled by returning error (aborts check but not entire report)
3. **Branch deletion race**: Handled by continuing on commit query errors
4. **Large repos slow down scan**: Accepted constraint; staleness check is optional feature

### Testing Gaps

- ⚠️ No test for pagination boundary (exactly 100 items)
- ⚠️ No test for transient error handling in commit query
- ⚠️ No performance benchmark for large repos

**Mitigation**: Manual testing with real repos validates behavior; consider adding integration tests.

---

## Future Enhancements

Potential improvements not in current implementation:

1. **Parallel commit queries**: Use goroutines + waitgroup to query branches concurrently
2. **Optional staleness check**: Add flag to skip per-branch queries when not needed
3. **Branch metadata caching**: Cache commit dates to avoid repeated queries
4. **GraphQL batch query**: Use GraphQL to fetch all branch commit dates in one query
5. **Progressive results**: Stream results as branches are checked (long-running feedback)
6. **Branch age histogram**: Report distribution of branch ages (not just count)

These were intentionally excluded to maintain simplicity and avoid premature optimization.
