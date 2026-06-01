# Feature Specification: Branch and Tag Analysis

**Feature Branch**: N/A (existing feature)  
**Created**: 2026-06-01  
**Status**: migrated  
**Input**: Reverse-engineered from existing implementation in `internal/api/client.go` and `internal/checks/checks.go`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Branch Count Monitoring (Priority: P1) 🎯 MVP

As a repository administrator, I need to identify repositories with excessive branch counts, so I can enforce housekeeping policies and reduce repository overhead caused by abandoned branches.

**Why this priority**: Large branch counts indicate poor branch management hygiene; excessive branches slow down Git operations and create clutter. Configurable threshold allows organizations to set their own policies.

**Independent Test**: Run against repos with varying branch counts and verify threshold detection works correctly.

**Acceptance Scenarios**:

1. **Given** a repository with 60 branches and MaxBranches=50, **When** health check runs, **Then** CheckTooManyBranches appears in FailedChecks
2. **Given** a repository with 30 branches and MaxBranches=50, **When** health check runs, **Then** CheckTooManyBranches does NOT appear in FailedChecks
3. **Given** a repository with 60 branches and MaxBranches=0 (default), **When** health check runs, **Then** uses DefaultMaxBranches=50 threshold
4. **Given** a repository with 51 branches and default threshold, **When** health check runs, **Then** CheckTooManyBranches appears in FailedChecks
5. **Given** BranchCount populated via API, **When** health check runs, **Then** count includes all branches (not just non-default)

---

### User Story 2 - Stale Branch Detection (Priority: P1) 🎯 MVP

As a development team lead, I need to identify branches that haven't been updated in months, so I can prompt developers to merge or delete stale branches that are unlikely to be finished.

**Why this priority**: Stale branches indicate abandoned work or forgotten feature branches; identifying them enables cleanup and reduces confusion about active development.

**Independent Test**: Run against repos with branches at different staleness levels and verify detection logic.

**Acceptance Scenarios**:

1. **Given** a branch with no commits since cutoff date, **When** staleness check runs, **Then** StaleBranchCount increments
2. **Given** a branch with commits after cutoff date, **When** staleness check runs, **Then** branch is NOT counted as stale
3. **Given** the default branch, **When** staleness check runs, **Then** default branch is excluded from staleness analysis
4. **Given** a repository with StaleBranchCount > 0, **When** health check runs, **Then** CheckHasStaleBranches appears in FailedChecks
5. **Given** a branch with recent commits and cutoff date of 180 days ago, **When** staleness check runs, **Then** branch not counted as stale

---

### User Story 3 - Tag Count Monitoring (Priority: P2)

As a release engineer, I need to identify repositories with excessive tag counts, so I can identify repos that may have automated tagging creating clutter or repos that need tag pruning.

**Why this priority**: Nice-to-have for repository hygiene; excessive tags indicate aggressive automated tagging or lack of cleanup policies.

**Independent Test**: Run against repos with varying tag counts and verify threshold detection.

**Acceptance Scenarios**:

1. **Given** a repository with 150 tags and MaxTags=100, **When** health check runs, **Then** CheckTooManyTags appears in FailedChecks
2. **Given** a repository with 50 tags and MaxTags=100, **When** health check runs, **Then** CheckTooManyTags does NOT appear in FailedChecks
3. **Given** a repository with 110 tags and MaxTags=0 (default), **When** health check runs, **Then** uses DefaultMaxTags=100 threshold
4. **Given** TagCount populated via API, **When** health check runs, **Then** count includes all tags from all pages

---

### User Story 4 - Configurable Thresholds (Priority: P1) 🎯 MVP

As a repository auditor, I need to customize branch and tag thresholds based on organizational policies, so I can enforce different standards for different types of repositories.

**Why this priority**: Organizations have different policies; monorepos might need higher thresholds while small projects should keep counts low.

**Independent Test**: Run with different --max-branches and --max-tags values and verify custom thresholds apply.

**Acceptance Scenarios**:

1. **Given** MaxBranches=100 in Options, **When** checking 101 branches, **Then** uses 100 as threshold
2. **Given** MaxBranches=0 in Options, **When** checking branches, **Then** uses DefaultMaxBranches=50
3. **Given** MaxTags=200 in Options, **When** checking 201 tags, **Then** uses 200 as threshold
4. **Given** MaxTags=0 in Options, **When** checking tags, **Then** uses DefaultMaxTags=100
5. **Given** custom thresholds via CLI flags, **When** Options constructed, **Then** thresholds passed to PopulateBranchTagChecks

---

## Functional Requirements *(mandatory)*

### Core Requirements

#### R1: Branch Count Aggregation with Pagination (Priority: P1)

**Must have**: System must count all branches in a repository, handling pagination for repos with > 100 branches.

**Details**:
- GET `/repos/{owner}/{repo}/branches?per_page=100&page={n}`
- Loop through pages until response length < 100
- Increment BranchCount for each branch found
- Include all branches (default branch + feature branches)
- Store final count in Repository.BranchCount field

**Why this priority**: Core metric for identifying repositories with excessive branches; requires pagination to handle large repos.

**Implementation**:
```go
// lines 374-411 in client.go
page := 1
for {
    var branches []branchItem
    path := fmt.Sprintf("repos/%s/%s/branches?per_page=100&page=%d", owner, name, page)
    if err := c.rest.Get(path, &branches); err != nil {
        return err
    }
    repo.BranchCount += len(branches)
    // ... staleness checks ...
    if len(branches) < 100 {
        break
    }
    page++
}
```

---

#### R2: Stale Branch Detection with Commit Query (Priority: P1)

**Must have**: System must determine which branches are stale by querying commits after a cutoff date.

**Details**:
- For each non-default branch, query: GET `/repos/{owner}/{repo}/commits?sha={branch}&since={cutoffDate}&per_page=1`
- If commits array is empty, branch has no commits after cutoff → stale
- If commits array has length > 0, branch has recent activity → not stale
- Exclude default branch from staleness analysis (always considered active)
- Increment StaleBranchCount for each stale branch
- Store final count in Repository.StaleBranchCount field

**Why this priority**: Provides actionable cleanup signal; requires per-branch API call (rate limit consideration).

**Implementation**:
```go
// lines 385-405 in client.go
sinceStr := since.UTC().Format(time.RFC3339)
for _, b := range branches {
    if b.Name == repo.DefaultBranch {
        continue  // Exclude default branch
    }
    var commits []interface{}
    cpath := fmt.Sprintf(
        "repos/%s/%s/commits?sha=%s&since=%s&per_page=1",
        owner, name, b.Name, sinceStr,
    )
    if err := c.rest.Get(cpath, &commits); err != nil {
        continue  // Skip on transient errors
    }
    if len(commits) == 0 {
        repo.StaleBranchCount++
    }
}
```

---

#### R3: Tag Count Aggregation with Pagination (Priority: P2)

**Must have**: System must count all tags in a repository, handling pagination for repos with > 100 tags.

**Details**:
- GET `/repos/{owner}/{repo}/tags?per_page=100&page={n}`
- Loop through pages until response length < 100
- Increment TagCount for each tag found
- Store final count in Repository.TagCount field

**Why this priority**: Nice-to-have metric for identifying aggressive tagging; less critical than branch monitoring.

**Implementation**:
```go
// lines 413-427 in client.go
page = 1
for {
    var tags []interface{}
    path := fmt.Sprintf("repos/%s/%s/tags?per_page=100&page=%d", owner, name, page)
    if err := c.rest.Get(path, &tags); err != nil {
        return err
    }
    repo.TagCount += len(tags)
    if len(tags) < 100 {
        break
    }
    page++
}
```

---

#### R4: Configurable Thresholds with Defaults (Priority: P1)

**Must have**: System must support custom branch/tag thresholds with sensible defaults when not specified.

**Details**:
- Options struct has MaxBranches and MaxTags fields (default 0 = use defaults)
- DefaultMaxBranches = 50 (line 40)
- DefaultMaxTags = 100 (line 47)
- Evaluate() function applies thresholds:
  - `if opts.MaxBranches == 0 { maxBranches = DefaultMaxBranches }`
  - Compare BranchCount > maxBranches
- CLI flags `--max-branches` and `--max-tags` map to Options fields

**Why this priority**: Allows organizational customization; defaults provide reasonable out-of-box behavior.

**Implementation**:
```go
// lines 209-216 in checks.go
maxBranches := opts.MaxBranches
if maxBranches == 0 {
    maxBranches = DefaultMaxBranches
}
if r.BranchCount > maxBranches {
    r.FailedChecks = append(r.FailedChecks, CheckTooManyBranches)
}
```

---

#### R5: Check Name Constants (Priority: P1)

**Must have**: System must define constant identifiers for branch/tag checks.

**Details**:
- `CheckTooManyBranches = "too-many-branches"` (line 35)
- `CheckHasStaleBranches = "has-stale-branches"` (line 36)
- `CheckTooManyTags = "too-many-tags"` (line 37)
- Used in FailedChecks slice for reporting
- Follow kebab-case convention for consistency

**Why this priority**: Required for check evaluation and reporting; prevents typos and ensures consistency.

---

#### R6: Staleness Calculation with Time-Based Cutoff (Priority: P1)

**Must have**: System must accept a time-based cutoff parameter to define staleness.

**Details**:
- PopulateBranchTagChecks accepts `since time.Time` parameter
- CLI provides `--since` flag (e.g., "180d")
- Cutoff converted to RFC3339 timestamp for GitHub API query
- Branches with no commits after this timestamp are stale

**Why this priority**: Allows flexible staleness definitions (30d for fast-moving projects, 365d for maintenance-mode repos).

**Implementation**:
```go
// line 385 in client.go
sinceStr := since.UTC().Format(time.RFC3339)
```

---

#### R7: Graceful Error Handling for Branch Staleness (Priority: P1)

**Must have**: System must continue processing when individual branch commit queries fail.

**Details**:
- Transient errors (rate limits, deleted branch races) skip the branch
- BranchCount still reflects total (from initial branches query)
- StaleBranchCount may be under-counted in error cases
- Does not abort entire report on single branch failure

**Why this priority**: Prevents one problematic branch from breaking entire audit; prioritizes partial results over failure.

**Implementation**:
```go
// lines 395-401 in client.go
if err := c.rest.Get(cpath, &commits); err != nil {
    // On transient errors, skip the branch rather than aborting
    continue
}
```

---

#### R8: Result Struct Integration (Priority: P1)

**Must have**: System must store branch/tag metrics in Result struct for evaluation.

**Details**:
- Result.BranchCount (int) — total branches
- Result.StaleBranchCount (int) — stale branches
- Result.TagCount (int) — total tags
- Populated in Evaluate() from Repository fields (lines 126-128)
- Used in check evaluation logic (lines 209-227)

**Why this priority**: Required for check evaluation and formatter output.

---

## Success Criteria *(mandatory)*

### Test Coverage

- ✅ Branch count correctly aggregated across pagination
- ✅ Stale branch detection works with different cutoff dates
- ✅ Default branch excluded from staleness analysis
- ✅ Tag count correctly aggregated across pagination
- ✅ Custom thresholds override defaults (MaxBranches, MaxTags)
- ✅ Zero thresholds use default values (DefaultMaxBranches, DefaultMaxTags)
- ✅ FailedChecks populated correctly based on thresholds
- ✅ Transient errors don't abort entire check

### Integration Points

- ✅ PopulateBranchTagChecks() called from main command flow
- ✅ Cutoff date derived from --since CLI flag
- ✅ Thresholds derived from --max-branches and --max-tags CLI flags
- ✅ Check results included in Result struct
- ✅ FailedChecks correctly populated based on thresholds
- ✅ Formatter displays branch/tag counts in all output formats

### Performance Characteristics

- ✅ Pagination handles repos with hundreds of branches/tags
- ✅ Per-branch commit queries accept rate limit cost (documented tradeoff)
- ✅ Error handling prevents cascade failures on problematic branches
- ✅ Branch/tag queries run after other checks (allows early exit on errors)

---

## Assumptions & Constraints *(mandatory)*

### Assumptions

1. **Pagination behavior**: Assumes GitHub API returns max 100 items per page; loop exits when response < 100
2. **Default branch always active**: Assumes default branch should never be flagged as stale (intentional exclusion)
3. **Commit query semantics**: Assumes `since` parameter correctly filters commits by author date
4. **Staleness definition**: Assumes "no commits after cutoff" is sufficient signal (doesn't check last push date or PR activity)
5. **Branch existence stability**: Assumes branch won't be deleted between listing and commit query (handles race with continue)
6. **Tag immutability**: Assumes tags are append-only (no cleanup happening during scan)

### Constraints

1. **API rate limits**: Each non-default branch requires a commit query; repos with 100s of branches consume significant rate limit budget
2. **Pagination overhead**: Large repos (100+ branches/tags) require multiple API calls with network latency
3. **No caching**: Each run fetches fresh data; repeated scans of same repo re-fetch everything
4. **Staleness granularity**: Cutoff is date-based, not considering PR/review activity (branch might be stale but PR active)
5. **No branch metadata**: Doesn't capture branch author, creation date, or associated PRs
6. **No tag metadata**: Doesn't capture tag type (annotated vs lightweight), author, or associated releases

### Performance Impact

**Branch staleness checking is expensive**:
- Repository with 50 branches → 49 extra API calls (excluding default)
- Repository with 200 branches → 199 extra API calls
- Organization with 100 repos averaging 30 branches → ~3000 API calls

**Mitigation**: Documented as known constraint; users can disable check by not calling PopulateBranchTagChecks or filtering repos before scanning.

### Migration Gaps Identified

- ⚠️ No test coverage for pagination boundary (exactly 100 branches/tags)
- ⚠️ No test coverage for transient error handling in commit query
- ⚠️ No performance benchmarks for large repos (100+ branches)
- ⚠️ No documentation of rate limit impact in user-facing docs
- ⚠️ No option to disable staleness check while keeping branch count check

**Recommendation**: Add optional flag `--skip-stale-branches` to reduce API calls when staleness data not needed. Document rate limit impact in README.md.
