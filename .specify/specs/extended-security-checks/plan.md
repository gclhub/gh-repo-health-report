# Implementation Plan: Extended Security Checks

**Feature**: Extended Security Checks  
**Status**: migrated  
**Created**: 2026-06-01

## Technical Context

### Existing Architecture

This feature extends the existing health check system with security-focused checks that require additional API calls and permission-aware tristate logic.

**Key packages**:
- `internal/api/client.go`: GitHub API client with REST methods
- `internal/checks/checks.go`: Health check evaluation logic
- `cmd/gh-repo-health-report/main.go`: CLI entry point

**Dependencies**:
- `github.com/cli/go-gh/v2/pkg/api`: GitHub CLI REST client
- `github.com/spf13/cobra`: CLI framework

### Integration Points

1. **Repository struct** (internal/api/client.go):
   - Added security-related boolean fields (lines 62-71)
   - Added tristate fields for permission-gated checks
   - Uses nested `SecurityAndAnalysis` struct from GitHub API response

2. **PopulateExtendedChecks** (internal/api/client.go:270-363):
   - Called after PopulateFileChecks in command flow
   - Makes additional REST API calls for security features
   - Handles HTTP status codes: 200/204 (success), 403 (forbidden), 404 (not found)

3. **Check evaluation** (internal/checks/checks.go:96-230):
   - Extended check constants defined (lines 25-32)
   - Result struct includes extended fields (lines 74-87)
   - Evaluate() function includes tristate logic (lines 196-205)

### Technical Decisions

#### Decision 1: Tristate Logic for Permission-Gated Features

**Chosen approach**: Use paired boolean fields (Enabled + Unknown) instead of pointer types or enums.

**Rationale**:
- Keeps JSON serialization simple (no nil values)
- Explicit in Result struct (self-documenting)
- Test assertions are clear: `if !result.VulnerabilityAlertsUnknown && !result.VulnerabilityAlertsEnabled`
- Formatter can distinguish ✓ (enabled), ✗ (disabled), ? (unknown)

**Alternatives considered**:
- `*bool` pointers: Complicates JSON output, nil checks everywhere
- `string` with "enabled"/"disabled"/"unknown": Loses type safety
- Custom enum type: Overkill for simple tristate

---

#### Decision 2: 403 Means "Unknown" Not "Disabled"

**Chosen approach**: When API returns 403 (Forbidden), set *Unknown=true and exclude check from FailedChecks.

**Rationale**:
- User lacks permission to determine status, not that feature is disabled
- Prevents false negatives in security audits
- Aligns with principle: "If you can't verify it, don't claim it's failing"
- Formatter displays "?" to indicate "need admin access to check"

**Alternatives considered**:
- Treat 403 as disabled: Would flag repos user doesn't control, creating noise
- Treat 403 as error: Would abort entire report for permission issues
- Skip check entirely: Would hide actionable signal (user should get admin access)

---

#### Decision 3: Parse security_and_analysis from Existing API Response

**Chosen approach**: Secret scanning and push protection status extracted from `security_and_analysis` field already returned by GET /repos/{owner}/{repo}.

**Rationale**:
- No additional API call required (data already fetched)
- GitHub API only includes field when caller has push/admin access
- Empty Status string reliably indicates insufficient permissions

**Implementation**:
```go
// lines 351-360 in client.go
if sa := repo.SecurityAndAnalysis.SecretScanning.Status; sa != "" {
    repo.SecretScanningEnabled = sa == "enabled"
} else {
    repo.SecretScanningUnknown = true
}
```

**Alternatives considered**:
- Dedicated API endpoint: Doesn't exist for these features
- GraphQL query: Would require separate client setup, added complexity

---

#### Decision 4: Rulesets as Complement to Branch Protection

**Chosen approach**: Check both rulesets (modern) and branch protection (legacy) as separate checks.

**Rationale**:
- Organizations migrating from protection to rulesets need visibility into both
- Rulesets API has better permission model (visible to readers, not just admins)
- A repo might use rulesets, protection, both, or neither

**Implementation**:
- CheckNoBranchProtection: Fails if DefaultBranchProtected=false
- CheckNoRulesets: Fails if HasRulesets=false
- Independent checks allow org to choose which policy they enforce

**Alternatives considered**:
- Single "protection" check: Would miss repos using only one approach
- Only check rulesets: Would ignore legacy protection configs still in use

---

## Project Structure

### Modified Files

**internal/api/client.go** (~453 lines):
- Lines 16-19: securityFeatureStatus struct for parsing status fields
- Lines 44-51: SecurityAndAnalysis nested struct in Repository
- Lines 62-71: Extended check fields (HasDependabot, VulnerabilityAlerts*, etc.)
- Lines 270-363: PopulateExtendedChecks() implementation
- Lines 444-451: isForbidden() helper for 403 detection

**internal/checks/checks.go** (~230 lines):
- Lines 25-32: Extended check name constants
- Lines 74-87: Extended fields in Result struct
- Lines 96-230: Evaluate() includes extended check logic (lines 182-208)

**internal/checks/checks_test.go** (~343 lines):
- Lines 145-177: TestEvaluate_ExtendedChecks_AllPresent
- Lines 179-206: TestEvaluate_ExtendedChecks_Missing
- Lines 208-243: Tristate logic tests (VulnerabilityAlerts, SecretScanning, PushProtection Unknown)
- Lines 245-258: OpenIssueCount and SizeKB tests

### File Organization

```
internal/
  api/
    client.go          # PopulateExtendedChecks implementation
    client_test.go     # Mock tests for API behavior
  checks/
    checks.go          # Extended check evaluation logic
    checks_test.go     # Tristate logic and check result tests
```

## Implementation Phases

### Phase 1: Repository Struct Extensions ✅ COMPLETED

**Implemented**:
- Added security feature fields to Repository struct
- Added tristate boolean pairs (Enabled + Unknown)
- Added SecurityAndAnalysis nested struct for JSON parsing
- Added OpenIssueCount and SizeKB metadata fields

**Lines of code**: ~30 lines (struct definitions)

**Test coverage**: Verified via check evaluation tests

---

### Phase 2: API Integration for Security Features ✅ COMPLETED

**Implemented**:
- PopulateExtendedChecks() method on Client
- Dependabot file existence checks (.yml and .yaml)
- CI workflows directory check
- Branch protection endpoint call (with 403/404 handling)
- Rulesets endpoint call
- Vulnerability alerts endpoint call (with 403 handling)
- Secret scanning / push protection parsing from security_and_analysis

**Lines of code**: ~90 lines (API integration)

**Key functions**:
- `PopulateExtendedChecks(repo *Repository) error`
- Uses existing `rest.Get()` from go-gh client
- Uses existing `CheckFileExists()` helper
- Uses existing `isNotFound()` and new `isForbidden()` helpers

**Error handling**:
- 403 → set *Unknown=true, continue
- 404 → set feature as disabled/missing, continue
- Other errors → return error, abort check

---

### Phase 3: Check Evaluation Logic ✅ COMPLETED

**Implemented**:
- Extended check constants (8 new checks)
- Result struct fields populated from Repository
- Tristate logic in Evaluate() function (lines 196-205)
- FailedChecks population with permission awareness

**Lines of code**: ~60 lines (check logic)

**Key logic**:
```go
// Only flag as failed when status is known
if !r.VulnerabilityAlertsUnknown && !r.VulnerabilityAlertsEnabled {
    r.FailedChecks = append(r.FailedChecks, CheckNoVulnerabilityAlerts)
}
```

**Design principle**: Unknown status never triggers FailedChecks entry

---

### Phase 4: Test Coverage ✅ COMPLETED

**Implemented**:
- TestEvaluate_ExtendedChecks_AllPresent: Healthy repo passes all checks
- TestEvaluate_ExtendedChecks_Missing: All features disabled triggers failures
- TestEvaluate_VulnerabilityAlerts_Unknown: 403 handling
- TestEvaluate_SecretScanning_Unknown: Empty status field handling
- TestEvaluate_PushProtection_Unknown: Empty status field handling
- TestEvaluate_OpenIssueCountAndSize: Metadata field population

**Lines of code**: ~140 lines (test code)

**Test strategy**:
- Unit tests with baseRepo() fixture
- Table-driven tests for different states
- Explicit tests for tristate logic paths
- Mock tests in client_test.go for API behavior

---

## Validation Approach

### Automated Testing

**Unit tests** (internal/checks/checks_test.go):
- ✅ All extended checks pass on healthy repo
- ✅ All extended checks fail when features missing
- ✅ Unknown status prevents FailedChecks entry
- ✅ Tristate logic validated for all three security features

**Mock tests** (internal/api/client_test.go):
- ✅ Mock HTTP server simulates GitHub API responses
- ✅ Pagination tested for org/user repo listings
- ✅ File existence checks tested (200 vs 404)

### Manual Testing

**Test commands**:
```bash
# Test against public repo without admin access (should show ? for security checks)
go build && ./gh-repo-health-report --repo cli/cli

# Test against own repo with admin access (should show ✓ or ✗ for all checks)
go build && ./gh-repo-health-report --repo yourusername/yourrepo

# Test organization scan with mixed permissions
go build && ./gh-repo-health-report --org yourorg
```

**Expected outcomes**:
- Public repos → security checks show "?" (unknown status)
- Owned repos with security features → checks show "✓"
- Owned repos without security features → checks show "✗"
- No crashes on permission errors

### Quality Gates

All checks passed:
- ✅ `go build ./...` — builds successfully
- ✅ `go test ./...` — all tests pass
- ✅ `go vet ./...` — no warnings
- ✅ Integration with existing formatter (table/JSON/CSV output)

---

## Complexity Assessment

### Code Metrics

| Component | Files | Lines | Functions | Test Lines |
|-----------|-------|-------|-----------|------------|
| API Integration | 1 | ~90 | 1 main | ~50 (mocks) |
| Check Evaluation | 1 | ~60 | 1 main | ~140 |
| Struct Definitions | 2 | ~30 | 0 | 0 |
| **Total** | **4** | **~180** | **2** | **~190** |

### Complexity Factors

**Low complexity**:
- Extends existing patterns (PopulateFileChecks → PopulateExtendedChecks)
- Uses existing RESTClient infrastructure
- Boolean field logic (simple state management)

**Medium complexity**:
- Tristate logic requires careful handling to avoid false negatives
- Multiple API endpoints with different permission models
- HTTP status code interpretation (200, 204, 403, 404)

**High complexity**:
- Permission awareness across different security features
- Nested JSON parsing (security_and_analysis field)
- Error handling without aborting entire report

**Overall**: Medium complexity feature with high test coverage requirements due to tristate logic paths.

---

## Risk Mitigation

### Identified Risks

1. **Permission errors abort report**: Mitigated by gracefully handling 403 and setting *Unknown=true
2. **False negatives in security audits**: Mitigated by tristate logic (unknown ≠ disabled)
3. **API changes breaking parsing**: Mitigated by unit tests and error handling on missing fields
4. **Rate limiting on large org scans**: Accepted constraint (documented in spec)

### Testing Gaps

- ⚠️ No integration tests against live GitHub API
- ⚠️ Mock tests don't verify actual go-gh HTTPError structure
- ⚠️ No test for security_and_analysis completely missing from response

**Mitigation**: Existing manual testing provides confidence; consider adding optional integration test suite.

---

## Future Enhancements

Potential improvements not in current implementation:

1. **Caching layer**: Cache API responses for repeated checks on same repo
2. **Detailed protection rules**: Validate specific branch protection requirements (required reviewers, etc.)
3. **Workflow validation**: Parse .github/workflows/*.yml to verify CI quality
4. **Dependabot config validation**: Check for required package ecosystems
5. **SARIF integration**: Export results in standard security report format
6. **Enterprise version detection**: Document which features require specific GitHub versions

These were intentionally excluded to maintain simplicity and avoid scope creep.
