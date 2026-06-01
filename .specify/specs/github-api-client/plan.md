# Implementation Plan: GitHub API Client

**Status**: migrated (reverse-engineered from existing implementation)  
**Input**: Analysis of `internal/api/client.go` and `internal/api/client_test.go`

## Technical Context

### Technology Stack (Actual Implementation)

- **Language**: Go 1.x
- **Package**: `internal/api` — GitHub API interaction layer
- **Dependencies**: 
  - `github.com/cli/go-gh/v2/pkg/api` — Official GitHub CLI Go library
  - `time` package for timestamp handling
  - `fmt` for error formatting
- **Testing**: HTTP mock server tests with fixtures in `client_test.go`

### Project Structure

```
internal/api/
├── client.go       # 453 lines - API client implementation
└── client_test.go  # 254 lines - mock server tests
```

### API Interaction Categories

**Category 1: Repository Fetching** (3 methods)
- `GetRepo(owner, name)` — Single repository
- `ListOrgRepos(org, includeForks, includeArchived)` — Organization repos with pagination
- `ListUserRepos(user, includeForks, includeArchived)` — User repos with pagination

**Category 2: File Checks** (1 method, checks 12+ file patterns)
- `PopulateFileChecks(repo)` — Community files (README, LICENSE, CODE_OF_CONDUCT, CODEOWNERS, SECURITY, CONTRIBUTING, issue/PR templates)
- `CheckFileExists(owner, repo, path)` — Low-level file check helper

**Category 3: Extended Checks** (1 method, checks 8+ settings)
- `PopulateExtendedChecks(repo)` — Dependabot, CI, branch protection, rulesets, security settings
- Includes permission-aware error handling (403/404 distinction)

**Category 4: Branch & Tag Checks** (1 method)
- `PopulateBranchTagChecks(repo, since)` — Branch/tag counts, stale branch detection
- Handles pagination for repos with many branches/tags

## Implementation Approach (Actual)

### Phase 1: Client Initialization

**Client struct**:
```go
type Client struct {
    rest *api.RESTClient  // go-gh REST client
}
```

**Factory functions**:
- `NewClient()` — Creates client from default gh auth
- `NewClientFromREST(rest)` — Creates client from existing REST client (for testing)

**Design Decision**: Thin wrapper around go-gh library. All API calls delegate to `rest.Get()` with path and response struct.

### Phase 2: Repository Data Model

**Repository struct** (50+ fields):
```go
type Repository struct {
    // GitHub API standard fields
    FullName    string    `json:"full_name"`
    Name        string    `json:"name"`
    Owner       struct { Login string } `json:"owner"`
    Description string    `json:"description"`
    Homepage    string    `json:"homepage"`
    Topics      []string  `json:"topics"`
    PushedAt    time.Time `json:"pushed_at"`
    // ... feature flags, counts, security settings
    
    // Populated by client methods
    HasReadme         bool `json:"has_readme"`
    HasLicense        bool `json:"has_license"`
    // ... 20+ additional populated fields
}
```

**Security field structure**:
```go
SecurityAndAnalysis struct {
    SecretScanning struct {
        Status string `json:"status"`  // "enabled" or "disabled"
    } `json:"secret_scanning"`
    SecretScanningPushProtection struct {
        Status string `json:"status"`
    } `json:"secret_scanning_push_protection"`
} `json:"security_and_analysis"`
```

**Design Decision**: Single struct combines API response fields and computed fields. Populated incrementally via `Populate*` methods.

### Phase 3: Repository Fetching (User Story 1 + 6)

**Single Repository**:
```go
func (c *Client) GetRepo(owner, name string) (*Repository, error) {
    var repo Repository
    err := c.rest.Get(fmt.Sprintf("repos/%s/%s", owner, name), &repo)
    return &repo, err
}
```

**Paginated Lists**:
```go
func (c *Client) listRepos(basePath string, includeForks, includeArchived bool) ([]*Repository, error) {
    var all []*Repository
    page := 1
    for {
        var pageRepos []*Repository
        path := fmt.Sprintf("%s?per_page=100&page=%d", basePath, page)
        if err := c.rest.Get(path, &pageRepos); err != nil {
            return nil, err
        }
        // Filter forks and archived repos
        for _, r := range pageRepos {
            if r.Fork && !includeForks { continue }
            if r.Archived && !includeArchived { continue }
            all = append(all, r)
        }
        if len(pageRepos) < 100 { break }
        page++
    }
    return all, nil
}
```

**Design Decisions**:
- Pagination: Loop until page returns <100 results
- Filtering: Client-side (GitHub API doesn't support `?fork=false`)
- Error handling: Propagate errors immediately (no retry logic)

### Phase 4: File Existence Checks (User Story 2)

**Helper function**:
```go
func (c *Client) CheckFileExists(owner, repo, path string) (bool, error) {
    var result interface{}
    err := c.rest.Get(fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, path), &result)
    if err != nil {
        if isNotFound(err) {
            return false, nil  // 404 = file doesn't exist
        }
        return false, err  // Other errors propagated
    }
    return true, nil
}
```

**PopulateFileChecks pattern**:
1. **Dedicated endpoints** for README and LICENSE (GitHub special handling)
2. **Multiple location checks** for community files (loop through possible paths)
3. **Short-circuit on first match** (stop checking once file found)

**Example** (CODE_OF_CONDUCT):
```go
for _, p := range []string{"CODE_OF_CONDUCT.md", ".github/CODE_OF_CONDUCT.md", "docs/CODE_OF_CONDUCT.md"} {
    ok, err := c.CheckFileExists(owner, name, p)
    if err != nil { return err }
    if ok {
        repo.HasCodeOfConduct = true
        break
    }
}
```

**Design Decisions**:
- 404 errors converted to `false` (not an error condition)
- First match wins (no duplicate file detection)
- All checks run even if some fail (collect as much data as possible)

### Phase 5: Extended Checks (User Story 3)

**Dependabot check**:
```go
for _, p := range []string{".github/dependabot.yml", ".github/dependabot.yaml"} {
    ok, err := c.CheckFileExists(owner, name, p)
    if err != nil { return err }
    if ok {
        repo.HasDependabot = true
        break
    }
}
```

**CI workflows check**:
```go
var contents []interface{}
err := c.rest.Get(fmt.Sprintf("repos/%s/%s/contents/.github/workflows", owner, name), &contents)
if err != nil {
    if !isNotFound(err) { return err }
} else {
    repo.HasCIWorkflows = len(contents) > 0
}
```

**Branch protection check** (403 = no admin access, treated as unprotected):
```go
var protection interface{}
err = c.rest.Get(fmt.Sprintf("repos/%s/%s/branches/%s/protection", owner, name, repo.DefaultBranch), &protection)
if err != nil {
    if !isNotFound(err) && !isForbidden(err) {
        return err
    }
} else {
    repo.DefaultBranchProtected = true
}
```

**Vulnerability alerts check** (403 = unknown, not disabled):
```go
var noBody interface{}
err = c.rest.Get(fmt.Sprintf("repos/%s/%s/vulnerability-alerts", owner, name), &noBody)
if err != nil {
    if isForbidden(err) {
        repo.VulnerabilityAlertsUnknown = true
    } else if !isNotFound(err) {
        return err
    }
} else {
    repo.VulnerabilityAlertsEnabled = true
}
```

**Secret scanning parsing** (from GetRepo response):
```go
if sa := repo.SecurityAndAnalysis.SecretScanning.Status; sa != "" {
    repo.SecretScanningEnabled = sa == "enabled"
} else {
    repo.SecretScanningUnknown = true
}
```

**Design Decisions**:
- Branch protection 403 → treated as unprotected (legacy decision)
- Vulnerability alerts 403 → marked as unknown (cannot determine)
- Secret scanning absence → marked as unknown (requires push/admin access)
- Empty status string indicates field not returned by API (permission issue)

### Phase 6: Branch & Tag Checks (User Story 4)

**Branch pagination with staleness check**:
```go
func (c *Client) PopulateBranchTagChecks(repo *Repository, since time.Time) error {
    page := 1
    for {
        var branches []branchItem
        path := fmt.Sprintf("repos/%s/%s/branches?per_page=100&page=%d", owner, name, page)
        if err := c.rest.Get(path, &branches); err != nil { return err }
        repo.BranchCount += len(branches)
        
        // Check each non-default branch for staleness
        sinceStr := since.UTC().Format(time.RFC3339)
        for _, b := range branches {
            if b.Name == repo.DefaultBranch { continue }
            var commits []interface{}
            cpath := fmt.Sprintf("repos/%s/%s/commits?sha=%s&since=%s&per_page=1", owner, name, b.Name, sinceStr)
            if err := c.rest.Get(cpath, &commits); err != nil {
                continue  // Skip branch on error (don't abort audit)
            }
            if len(commits) == 0 {
                repo.StaleBranchCount++
            }
        }
        
        if len(branches) < 100 { break }
        page++
    }
    return nil
}
```

**Tag pagination** (simpler, no staleness check):
```go
page := 1
for {
    var tags []interface{}
    path := fmt.Sprintf("repos/%s/%s/tags?per_page=100&page=%d", owner, name, page)
    if err := c.rest.Get(path, &tags); err != nil { return err }
    repo.TagCount += len(tags)
    if len(tags) < 100 { break }
    page++
}
```

**Design Decisions**:
- Stale branch check: 1 API call per non-default branch (expensive!)
- Error handling: Skip failed branches (continue audit with partial data)
- Default branch excluded from staleness (always considered active)
- `since` parameter: ISO 8601 format for GitHub API compatibility

### Phase 7: Error Handling (User Story 5)

**Error helper functions**:
```go
func isNotFound(err error) bool {
    if e, ok := err.(*api.HTTPError); ok {
        return e.StatusCode == 404
    }
    return false
}

func isForbidden(err error) bool {
    if e, ok := err.(*api.HTTPError); ok {
        return e.StatusCode == 403
    }
    return false
}
```

**Error handling patterns**:
1. **File checks**: 404 → `false, nil` (not found, not an error)
2. **Branch protection**: 403/404 → treated as unprotected
3. **Vulnerability alerts**: 403 → marked as unknown
4. **Branch staleness**: Any error → skip branch (partial data acceptable)
5. **All others**: Errors propagated to caller

## Architecture Decisions

### Decision 1: Populate Pattern (Incremental Field Setting)
**Rationale**: Repository struct populated in stages. GetRepo returns basic metadata; Populate* methods add computed fields.

**Trade-offs**:
- ✅ Separation of concerns (fetch vs. compute)
- ✅ Caller controls which checks to run
- ❌ Mutates repository object (not pure function)
- ❌ Easy to forget to call Populate* methods

### Decision 2: Permission-Aware Tristate Logic
**Rationale**: Security settings require elevated permissions. Can't distinguish "disabled" from "cannot check" without tristate.

**Trade-offs**:
- ✅ Honest reporting (shows ? when unknown)
- ✅ Doesn't fail audits due to permission issues
- ❌ Inconsistent handling (branch protection 403 = unprotected, vulnerability alerts 403 = unknown)
- ❌ Complexity in error handling logic

### Decision 3: Client-Side Fork/Archive Filtering
**Rationale**: GitHub API doesn't support `?fork=false` filter parameter. Must filter after fetching.

**Trade-offs**:
- ✅ Works with GitHub API limitations
- ❌ Wastes API calls fetching repos that get filtered
- ❌ May hit rate limits faster for orgs with many forks

### Decision 4: Expensive Stale Branch Checking
**Rationale**: No bulk API to check branch last commit dates. Must query each branch individually.

**Trade-offs**:
- ✅ Accurate staleness detection
- ❌ Slow (1 API call per non-default branch)
- ❌ May hit rate limits for repos with 100+ branches
- ❌ Partial data on errors (may under-count stale branches)

## Complexity Assessment

**Cyclomatic Complexity**: High
- Pagination loops, nested error handling, multiple check locations
- Branch staleness: O(n) API calls for n branches

**Lines of Code**: 453 lines (implementation) + 254 lines (tests) = 707 total

**Dependencies**: 
- go-gh library (external)
- Tight coupling to GitHub API response format

**Test Coverage**: Mock server tests
- HTTP test server with fixtures
- Tests verify API paths and pagination logic
- Limited real API testing (requires auth)

## Known Gaps & Limitations

### Gap 1: No Rate Limit Handling
**Issue**: Branch staleness checks can consume 50+ API calls per repository. No explicit rate limit detection or throttling.

**Impact**: May hit GitHub's 5000 requests/hour limit when auditing orgs with many repos/branches.

**Potential Fix**: Add rate limit detection (`X-RateLimit-Remaining` header) and sleep/abort when low.

### Gap 2: Inconsistent 403 Handling
**Issue**: Branch protection 403 → treated as unprotected. Vulnerability alerts 403 → marked as unknown.

**Impact**: Confusing behavior; unclear why different treatment.

**Rationale**: Legacy decision (branch protection check predates tristate logic).

### Gap 3: No Retry Logic
**Issue**: Transient network errors abort entire operation.

**Impact**: Audit fails completely on temporary API issues.

**Potential Fix**: Add exponential backoff retry for 5xx errors.

### Gap 4: Client-Side Filtering Waste
**Issue**: Fetches forks/archived repos then discards them (GitHub API limitation).

**Impact**: Unnecessary API calls, slower audits.

**No Fix**: GitHub API doesn't support server-side filtering for forks.

## Testing Strategy (Actual Implementation)

### Test Infrastructure

**Mock HTTP server** (`client_test.go`):
```go
func mockServer(t *testing.T, handler http.Handler) *httptest.Server {
    srv := httptest.NewServer(handler)
    t.Cleanup(srv.Close)
    return srv
}
```

### Test Coverage

**Repository Fetching Tests**:
- `TestGetRepo_Mock` — Single repo fetch with JSON fixtures
- `TestListOrgRepos_Pagination_Mock` — Multi-page pagination logic

**File Checks Tests**:
- `TestCheckFileExists_Mock` — 200 vs 404 handling
- Limited coverage of PopulateFileChecks (would require extensive fixtures)

**Limitations**:
- No tests for extended checks (complex fixtures required)
- No tests for branch staleness (expensive to mock)
- No integration tests against real GitHub API

## Integration Points

### Input: CLI Flags (via main.go)
- org, owner, repos flags → routing to appropriate fetch method
- includeForks, includeArchived flags → passed to list methods

### Output: Repository Struct
- Passed to `checks.Evaluate()` for health evaluation
- All fields must be populated before evaluation

### Authentication: GitHub CLI
- Relies on `gh auth login` having been run
- go-gh library handles token retrieval automatically

## Performance Considerations

**Single Repo**: ~5-10 API calls
- 1 GetRepo call
- 2 dedicated file endpoints (readme, license)
- 5+ contents checks for community files
- Extended checks add 5+ more calls

**Org with 100 Repos**: ~500-1000 API calls
- 1 org list call (per 100 repos)
- 5-10 calls per repo for file checks
- Branch checks add 5-50 calls per repo

**Rate Limit Impact**: 
- GitHub API: 5000 requests/hour
- Audit of 50 repos: ~500 calls = well within limit
- Audit of 500 repos with branch checks: ~5000-10000 calls = may hit limit

**Optimization Opportunities**:
- Cache file checks across repos (same files in .github/ template)
- Skip branch staleness for repos with <10 branches
- Batch API calls via GraphQL (major refactor)

## Maintenance Notes

**Adding New File Check**:
1. Add boolean field to `Repository` struct
2. Add check location loop in `PopulateFileChecks()`
3. Use `CheckFileExists()` helper for each location
4. Short-circuit on first match

**Adding New Extended Check**:
1. Add field(s) to `Repository` struct (consider tristate if permission-gated)
2. Add API call in `PopulateExtendedChecks()`
3. Handle 404/403 errors appropriately
4. Parse response and set field(s)

**Changing Error Handling**:
- Coordinate with checks package (affects failure collection)
- Consider impact on existing audits (behavior change)
- Add tests for new error scenarios
