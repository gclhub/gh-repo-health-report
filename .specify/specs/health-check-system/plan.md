# Implementation Plan: Health Check System

**Status**: migrated (reverse-engineered from existing implementation)  
**Input**: Analysis of `internal/checks/checks.go` and `internal/checks/checks_test.go`

## Technical Context

### Technology Stack (Actual Implementation)

- **Language**: Go 1.x
- **Package**: `internal/checks` — isolated health evaluation logic
- **Dependencies**: 
  - `time` package for staleness calculations
  - `internal/api` for Repository struct definition
- **Testing**: Table-driven tests in `checks_test.go` with comprehensive scenarios

### Project Structure

```
internal/checks/
├── checks.go           # 230 lines - core evaluation logic
└── checks_test.go      # 343 lines - test scenarios
```

### Check Categories (25+ checks implemented)

**Category 1: Repository Metadata** (4 checks)
- `has-description` — Repository has non-empty description
- `has-homepage` — Repository has non-empty homepage URL
- `stale` — No push since `--since` threshold
- Topics count (not a check, but tracked for output)

**Category 2: Community Files** (8 checks)
- `missing-readme` — Uses GitHub `/readme` API endpoint
- `missing-license` — Uses GitHub `/license` API endpoint
- `missing-code-of-conduct` — Checks root, `.github/`, `docs/`
- `missing-codeowners` — Checks `.github/`, root, `docs/`
- `missing-security` — Checks root, `.github/`
- `missing-contributing` — Checks root, `.github/`
- `missing-issue-templates` — Checks `.github/ISSUE_TEMPLATE/` dir or `.md` files
- `missing-pr-template` — Checks multiple locations

**Category 3: GitHub Features** (3 checks)
- `has-issues` — Issues feature enabled
- `has-projects` — Projects feature enabled
- `has-wiki` — Wiki feature enabled

**Category 4: Security & Automation** (7 checks)
- `missing-dependabot` — `.github/dependabot.yml` or `.yaml`
- `missing-ci` — `.github/workflows/` directory has files
- `no-branch-protection` — Default branch has classic protection rules
- `no-rulesets` — Repository rulesets configured (modern protection)
- `no-vulnerability-alerts` — Dependabot security alerts enabled (admin required)
- `no-secret-scanning` — Secret scanning enabled (push/admin required)
- `no-push-protection` — Push protection enabled (push/admin required)
- `no-delete-branch-on-merge` — Auto-delete head branches on merge

**Category 5: Branch & Tag Management** (3 checks)
- `too-many-branches` — Branch count exceeds `--max-branches` (default 50, 0 to disable)
- `has-stale-branches` — Non-default branches with no commits since `--since`
- `too-many-tags` — Tag count exceeds `--max-tags` (default 100, 0 to disable)

## Implementation Approach (Actual)

### Phase 1: Data Structure Design

**Result struct** (`checks.Result`):
```go
type Result struct {
    Repository        *api.Repository
    Stale             bool
    HasDescription    bool
    HasHomepage       bool
    TopicsCount       int
    // ... 28+ boolean fields for each check
    // Security tristate: both Enabled and Unknown flags
    VulnerabilityAlertsEnabled bool
    VulnerabilityAlertsUnknown bool
    SecretScanningEnabled      bool
    SecretScanningUnknown      bool
    PushProtectionEnabled      bool
    PushProtectionUnknown      bool
    // Counts
    BranchCount      int
    StaleBranchCount int
    TagCount         int
    FailedChecks     []string // Check names that failed
}
```

**Options struct** (`checks.Options`):
```go
type Options struct {
    Since       time.Duration // Staleness threshold
    MaxBranches int           // 0 = disabled
    MaxTags     int           // 0 = disabled
}
```

### Phase 2: Check Constants

All check names defined as constants in `checks.go`:
```go
const (
    CheckHasDescription        = "has-description"
    CheckMissingReadme         = "missing-readme"
    CheckStale                 = "stale"
    CheckNoBranchProtection    = "no-branch-protection"
    CheckTooManyBranches       = "too-many-branches"
    // ... 20+ more constants
)
```

**Design Decision**: String constants enable CLI `--fail-on` flag matching and output formatting.

### Phase 3: Evaluation Logic

**Core function**: `Evaluate(repo *api.Repository, opts Options) *Result`

**Algorithm**:
1. **Initialize Result** — Copy repository metadata to result struct
2. **Evaluate staleness** — Compare `repo.PushedAt` against `opts.Since` threshold
3. **Map boolean fields** — Direct mapping from `api.Repository` to `Result` (file checks populated by API client)
4. **Collect failures** — Iterate through all checks, append failed check names to `FailedChecks` slice
5. **Handle thresholds** — Apply `MaxBranches` and `MaxTags` thresholds (0 = use defaults)
6. **Tristate security checks** — Only add to `FailedChecks` when status is known and disabled

**Key Implementation Details**:
- Default staleness threshold: 180 days if `opts.Since` is zero
- Default `MaxBranches`: 50 (constant `DefaultMaxBranches`)
- Default `MaxTags`: 100 (constant `DefaultMaxTags`)
- Security checks: Only fail when `!Unknown && !Enabled` to avoid false positives
- Stale branch check: Fails when `StaleBranchCount > 0` (any stale branches present)

### Phase 4: Failed Check Collection

**Pattern**:
```go
if !r.HasReadme {
    r.FailedChecks = append(r.FailedChecks, CheckMissingReadme)
}
```

**Special Cases**:
- Security checks: `if !r.VulnerabilityAlertsUnknown && !r.VulnerabilityAlertsEnabled { ... }`
- Branch threshold: `if r.BranchCount > maxBranches { ... }` (with default fallback)
- Staleness: `if time.Since(repo.PushedAt) > threshold { ... }`

## Architecture Decisions

### Decision 1: Single-Pass Evaluation
**Rationale**: All checks evaluated in one function call for efficiency. Repository data is pre-populated by API client before evaluation.

**Trade-offs**:
- ✅ Fast evaluation (no additional API calls)
- ✅ Simple interface (one function)
- ❌ Tight coupling to `api.Repository` structure
- ❌ Cannot evaluate checks incrementally

### Decision 2: Tristate Security Flags
**Rationale**: Security settings require elevated permissions. Using `Enabled` + `Unknown` flags allows distinguishing "disabled" from "cannot determine".

**Implementation**:
```go
VulnerabilityAlertsEnabled bool
VulnerabilityAlertsUnknown bool
```

**Output Mapping**:
- `Unknown = true` → Display "?"
- `Unknown = false, Enabled = true` → Display "✓"
- `Unknown = false, Enabled = false` → Display "✗", add to FailedChecks

### Decision 3: Configurable Thresholds with Defaults
**Rationale**: Organizations have different policies. Providing defaults (50 branches, 100 tags, 180 days) works for most cases but allows customization.

**Implementation**:
- Flag value `0` disables the check entirely
- Missing flag value uses constant defaults
- No validation of threshold reasonableness (user responsibility)

### Decision 4: Failed Check Name List
**Rationale**: CLI needs to match check names for `--fail-on` flag. Collecting failed check names as strings enables flexible failure policies.

**Usage**:
```go
// In main.go
if shouldFail(result.FailedChecks, failOnFlags) {
    os.Exit(1)
}
```

## Complexity Assessment

**Cyclomatic Complexity**: Medium
- Single evaluation function with 25+ conditional branches
- Straightforward logic (mostly boolean checks and comparisons)
- No complex algorithms or recursion

**Lines of Code**: 230 lines (implementation) + 343 lines (tests) = 573 total

**Dependencies**: Minimal
- Depends on `api.Repository` struct definition
- No external libraries beyond standard library

**Test Coverage**: Comprehensive
- Table-driven tests for various check scenarios
- Healthy repo baseline test
- Individual check failure tests
- Threshold configuration tests
- Unknown security status tests

## Known Gaps & Limitations

### Gap 1: Empty Repository Handling
**Issue**: Repositories with zero commits have `PushedAt = time.Zero`. Staleness calculation may behave unexpectedly.

**Impact**: May incorrectly flag brand-new repos as extremely stale.

**Potential Fix**: Add zero-time check before staleness evaluation.

### Gap 2: No Check Prioritization
**Issue**: All checks treated equally. No distinction between critical (LICENSE) and nice-to-have (wiki enabled).

**Impact**: Users must manually interpret which failures matter most.

**Workaround**: Use `--fail-on` with specific check names to enforce critical checks only.

### Gap 3: Limited Inline Documentation
**Issue**: Complex logic (threshold defaults, tristate handling) not extensively documented in code comments.

**Impact**: Future maintainers may need to reverse-engineer intent.

**Recommendation**: Add comments explaining tristate logic and default threshold behavior.

## Testing Strategy (Actual Implementation)

### Test Structure

**Baseline Test** (`TestEvaluate_Healthy`):
- Repository with all checks passing
- Verifies no failed checks when repo is healthy

**Individual Check Tests**:
- `TestEvaluate_Stale` — Repository older than threshold
- `TestEvaluate_NotStale` — Recent repository
- `TestEvaluate_MissingFiles` — Files absent, checks fail
- Additional tests for security settings, thresholds, etc.

**Test Helper**:
```go
func baseRepo() *api.Repository {
    // Returns fully-populated healthy repository
}
```

**Pattern**: Start with healthy baseline, modify specific fields to trigger failures, verify correct checks fail.

## Integration Points

### Input: API Client (`internal/api`)
- Receives `api.Repository` with all fields pre-populated
- File checks completed via `PopulateFileChecks()`
- Extended checks via `PopulateExtendedChecks()`
- Branch/tag checks via `PopulateBranchTagChecks()`

### Output: Formatter (`internal/formatter`)
- Passes `Result` slice to formatter
- Formatter uses boolean flags and counts for display
- FailedChecks slice enables failure detection

### CLI Integration (`cmd/gh-repo-health-report`)
- Options constructed from CLI flags (`--since`, `--max-branches`, `--max-tags`)
- Evaluate called for each repository
- FailedChecks used for `--fail-on` logic

## Performance Considerations

**Time Complexity**: O(1) per repository
- No loops (except appending to FailedChecks slice)
- All data pre-computed by API client

**Space Complexity**: O(n) for n repositories
- Each Result stores ~30 fields + FailedChecks slice
- Typical memory: ~500 bytes per Result

**Scalability**: Can evaluate 1000+ repositories efficiently
- Bottleneck is API client, not evaluation logic
- No shared state; thread-safe if needed

## Maintenance Notes

**Adding New Checks**:
1. Add constant to check name list
2. Add boolean field to `Result` struct
3. Add evaluation logic in `Evaluate()` function
4. Add failure collection logic
5. Update formatter to display new check
6. Add tests for new check

**Modifying Thresholds**:
- Update `DefaultMaxBranches` or `DefaultMaxTags` constants
- No code changes needed; defaults apply when flags not set

**Changing Tristate Logic**:
- Careful: Affects failure counting and output display
- Must coordinate changes with formatter package
