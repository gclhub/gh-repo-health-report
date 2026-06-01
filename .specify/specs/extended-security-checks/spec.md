# Feature Specification: Extended Security Checks

**Feature Branch**: N/A (existing feature)  
**Created**: 2026-06-01  
**Status**: migrated  
**Input**: Reverse-engineered from existing implementation in `internal/api/client.go` and `internal/checks/checks.go`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Security Feature Detection with Tristate Logic (Priority: P0) 🎯 MVP

As a security engineer auditing GitHub repositories, I need to determine whether security features (vulnerability alerts, secret scanning, push protection) are enabled, disabled, or unknown due to permission restrictions, so I can identify actionable security gaps versus access issues.

**Why this priority**: Critical security compliance feature; tristate logic (enabled/disabled/unknown) prevents false negatives when API returns 403 due to insufficient permissions. This is the core differentiator from basic health checks.

**Independent Test**: Run against a public repo without push access and verify security checks show "?" for unknown values instead of false failures.

**Acceptance Scenarios**:

1. **Given** a repository with push access and vulnerability alerts enabled, **When** health check runs, **Then** VulnerabilityAlertsEnabled=true and VulnerabilityAlertsUnknown=false
2. **Given** a repository without push access (403 response), **When** checking vulnerability alerts, **Then** VulnerabilityAlertsUnknown=true and check does NOT appear in FailedChecks
3. **Given** a repository with secret scanning disabled, **When** health check runs, **Then** SecretScanningEnabled=false and check appears in FailedChecks
4. **Given** a repository with unknown secret scanning status, **When** health check runs, **Then** SecretScanningUnknown=true and check does NOT appear in FailedChecks
5. **Given** a repository with push protection enabled, **When** health check runs, **Then** PushProtectionEnabled=true (parsed from security_and_analysis field)
6. **Given** a repository where security_and_analysis field is empty, **When** health check runs, **Then** PushProtectionUnknown=true

---

### User Story 2 - Branch Protection and Rulesets Validation (Priority: P1) 🎯 MVP

As a DevOps engineer enforcing repository standards, I need to verify whether repositories have default branch protection rules or modern rulesets configured, so I can ensure code review policies are enforced before merging.

**Why this priority**: Essential for preventing unreviewed code from reaching production; supports both legacy branch protection API and modern rulesets API.

**Independent Test**: Run against repos with/without branch protection and verify detection works correctly for both APIs.

**Acceptance Scenarios**:

1. **Given** a repository with branch protection on default branch, **When** health check runs, **Then** DefaultBranchProtected=true and no-branch-protection check passes
2. **Given** a repository without branch protection (404 response), **When** health check runs, **Then** DefaultBranchProtected=false and no-branch-protection check fails
3. **Given** a repository with rulesets configured (array length > 0), **When** health check runs, **Then** HasRulesets=true and no-rulesets check passes
4. **Given** a repository without admin access (403 on protection endpoint), **When** health check runs, **Then** treats as unprotected (DefaultBranchProtected=false)
5. **Given** a repository using rulesets instead of legacy protection, **When** health check runs, **Then** HasRulesets=true compensates for DefaultBranchProtected=false

---

### User Story 3 - CI/CD and Automation Checks (Priority: P1) 🎯 MVP

As a repository maintainer, I need to verify whether repositories have automated CI workflows (GitHub Actions) and dependency management (Dependabot) configured, so I can ensure code quality gates and security updates are automated.

**Why this priority**: Core DevOps best practice; easy to detect via file existence checks without requiring special permissions.

**Independent Test**: Run against repos with/without .github/workflows and .github/dependabot.yml files.

**Acceptance Scenarios**:

1. **Given** a repository with .github/workflows/ directory containing files, **When** health check runs, **Then** HasCIWorkflows=true and missing-ci check passes
2. **Given** a repository without .github/workflows/, **When** health check runs, **Then** HasCIWorkflows=false and missing-ci check fails
3. **Given** a repository with .github/dependabot.yml, **When** health check runs, **Then** HasDependabot=true and missing-dependabot check passes
4. **Given** a repository with .github/dependabot.yaml (alternate extension), **When** health check runs, **Then** HasDependabot=true
5. **Given** a repository without either dependabot file, **When** health check runs, **Then** HasDependabot=false and missing-dependabot check fails

---

### User Story 4 - Repository Settings Validation (Priority: P2)

As a repository administrator, I need to verify whether automatic branch deletion after merge is enabled, so I can ensure stale branches don't accumulate after PRs are merged.

**Why this priority**: Nice-to-have housekeeping feature; less critical than security checks but helpful for repository hygiene.

**Independent Test**: Run against repos with/without delete-branch-on-merge setting enabled.

**Acceptance Scenarios**:

1. **Given** a repository with DeleteBranchOnMerge=true, **When** health check runs, **Then** no-delete-branch-on-merge check passes
2. **Given** a repository with DeleteBranchOnMerge=false, **When** health check runs, **Then** no-delete-branch-on-merge check fails
3. **Given** a repository metadata response includes delete_branch_on_merge field, **When** parsing, **Then** field is correctly mapped to Repository struct

---

## Functional Requirements *(mandatory)*

### Core Requirements

#### R1: Tristate Security Status Tracking (Priority: P0)

**Must have**: System must distinguish between three states for permission-gated security features:
- **Enabled**: Feature is confirmed enabled (API returned success with status="enabled")
- **Disabled**: Feature is confirmed disabled (API returned success with status="disabled" or feature not enabled)
- **Unknown**: Feature status cannot be determined due to permission restrictions (API returned 403)

**Details**:
- VulnerabilityAlertsEnabled + VulnerabilityAlertsUnknown fields
- SecretScanningEnabled + SecretScanningUnknown fields  
- PushProtectionEnabled + PushProtectionUnknown fields
- Unknown status prevents check from appearing in FailedChecks
- Tests must verify tristate logic for all three security features

**Why this priority**: Prevents false negatives and provides actionable audit signals — "?" means get admin access, "✗" means enable the feature.

---

#### R2: Security Feature API Integration (Priority: P0)

**Must have**: System must query GitHub API endpoints to determine security feature status:
- **Vulnerability alerts**: GET `/repos/{owner}/{repo}/vulnerability-alerts` (204=enabled, 404=disabled, 403=unknown)
- **Secret scanning**: Parse `security_and_analysis.secret_scanning.status` from repo metadata (requires push access)
- **Push protection**: Parse `security_and_analysis.secret_scanning_push_protection.status` from repo metadata

**Details**:
- Use existing RESTClient from go-gh package
- Handle HTTP status codes: 200/204 (success), 404 (not found), 403 (forbidden)
- Parse nested JSON structures from security_and_analysis field
- Empty status string indicates field not present (no permission)

**Why this priority**: Core functionality; API integration is the only way to determine security feature status.

---

#### R3: Branch Protection Detection (Priority: P1)

**Must have**: System must check default branch protection status via GitHub API:
- GET `/repos/{owner}/{repo}/branches/{branch}/protection`
- 200 response = protected, 404 = not protected, 403 = treat as not protected
- DefaultBranchProtected boolean field in Repository struct
- CheckNoBranchProtection in FailedChecks when false

**Details**:
- Use DefaultBranch field from repository metadata
- 403 response treated as unprotected (aligns with security-first approach)
- Does not validate specific protection rules, only checks existence

**Why this priority**: Essential for enforcing code review requirements; widely used legacy API.

---

#### R4: Repository Rulesets Detection (Priority: P1)

**Must have**: System must detect modern GitHub rulesets as alternative to branch protection:
- GET `/repos/{owner}/{repo}/rulesets`
- Array response with length > 0 = has rulesets
- HasRulesets boolean field in Repository struct
- CheckNoRulesets in FailedChecks when false

**Details**:
- Rulesets API is newer than branch protection API
- Rulesets visible to anyone with read access (no 403 issues)
- Complements branch protection check (repo might use one or both)

**Why this priority**: Modern replacement for branch protection; organizations migrating to rulesets need visibility.

---

#### R5: CI Workflow Detection (Priority: P1)

**Must have**: System must detect GitHub Actions workflows:
- GET `/repos/{owner}/{repo}/contents/.github/workflows`
- Array response with length > 0 = has CI workflows
- HasCIWorkflows boolean field in Repository struct
- CheckMissingCI in FailedChecks when false

**Details**:
- Checks directory existence, not individual workflow validity
- 404 response means no workflows directory
- Any file in workflows/ directory counts as CI present

**Why this priority**: CI/CD is fundamental to modern software delivery; easy to detect without permissions.

---

#### R6: Dependabot Configuration Detection (Priority: P1)

**Must have**: System must detect Dependabot configuration files:
- Check both `.github/dependabot.yml` and `.github/dependabot.yaml`
- Use CheckFileExists helper from api client
- HasDependabot boolean field in Repository struct
- CheckMissingDependabot in FailedChecks when false

**Details**:
- Supports both .yml and .yaml extensions
- First match returns true (early exit optimization)
- File must exist, content validation not required

**Why this priority**: Automated dependency updates are security best practice; configuration is simple file existence check.

---

#### R7: Delete Branch on Merge Setting (Priority: P2)

**Must have**: System must check whether automatic branch deletion is enabled:
- Parse `delete_branch_on_merge` field from repository metadata (already fetched)
- DeleteBranchOnMerge boolean field in Repository struct
- CheckNoDeleteBranchOnMerge in FailedChecks when false

**Details**:
- No additional API call required (field included in base repo response)
- Helps prevent stale branch accumulation
- Less critical than security checks but improves repository hygiene

**Why this priority**: Nice-to-have feature for housekeeping; minimal implementation cost since data already available.

---

#### R8: Check Name Constants (Priority: P1)

**Must have**: System must define constant identifiers for all extended checks:
- `CheckMissingDependabot = "missing-dependabot"`
- `CheckMissingCI = "missing-ci"`
- `CheckNoBranchProtection = "no-branch-protection"`
- `CheckNoRulesets = "no-rulesets"`
- `CheckNoVulnerabilityAlerts = "no-vulnerability-alerts"`
- `CheckNoSecretScanning = "no-secret-scanning"`
- `CheckNoPushProtection = "no-push-protection"`
- `CheckNoDeleteBranchOnMerge = "no-delete-branch-on-merge"`

**Details**:
- Follow kebab-case naming convention for consistency with existing checks
- Used in FailedChecks slice for reporting
- Must match formatter expectations for display names

**Why this priority**: Required for check evaluation and reporting; prevents typos and ensures consistency.

---

## Success Criteria *(mandatory)*

### Test Coverage

- ✅ All tristate logic paths tested (enabled, disabled, unknown)
- ✅ HTTP status code handling tested (200, 204, 403, 404)
- ✅ Security feature detection tested for all three features
- ✅ Branch protection and rulesets detection tested
- ✅ CI workflow and Dependabot detection tested
- ✅ Unknown status prevents FailedChecks entry (no false negatives)
- ✅ All check constants defined and used correctly

### Integration Points

- ✅ PopulateExtendedChecks() called from main command flow
- ✅ Check results included in Result struct
- ✅ FailedChecks correctly populated based on tristate logic
- ✅ Formatter displays security status (✓, ✗, or ?)
- ✅ OpenIssueCount and SizeKB fields populated from base repo metadata

### Error Handling

- ✅ 403 responses handled gracefully (set *Unknown=true)
- ✅ 404 responses correctly interpreted as disabled/missing
- ✅ Empty security_and_analysis field handled (set *Unknown=true)
- ✅ Transient errors don't crash evaluation (skip and continue)
- ✅ Missing permissions communicated to user via "?" display

---

## Assumptions & Constraints *(mandatory)*

### Assumptions

1. **GitHub API availability**: Assumes GitHub API endpoints are available and follow documented behavior
2. **Permission model**: Assumes vulnerability alerts endpoint requires admin access (403 when insufficient)
3. **security_and_analysis field**: Assumes field only present in API response when caller has push/admin access
4. **Rulesets API stability**: Assumes rulesets API is stable (introduced in GitHub Enterprise Server 3.8+)
5. **File existence semantics**: Assumes CheckFileExists returning true means file/directory exists (404 = does not exist)
6. **Workflow definition**: Assumes any file in .github/workflows/ indicates CI is configured

### Constraints

1. **Permission gating**: Some checks require push or admin access; must handle gracefully when unavailable
2. **API rate limits**: Each extended check may require additional API calls; impacts performance on large org scans
3. **No validation depth**: Checks only detect presence, not correctness (e.g., doesn't validate Dependabot config syntax)
4. **GitHub-specific**: Tied to GitHub API; not portable to other Git hosting platforms
5. **No caching**: Each run fetches fresh data; no caching layer for repeated checks
6. **Enterprise compatibility**: Rulesets may not be available on older GitHub Enterprise Server versions

### Migration Gaps Identified

- ⚠️ No explicit test for 403 handling on branch protection endpoint (behavior documented in code comments)
- ⚠️ No test coverage for empty security_and_analysis field edge case
- ⚠️ Tests mock HTTP responses but don't test actual go-gh RESTClient error types
- ⚠️ No documentation of which GitHub Enterprise versions support each feature
- ⚠️ No rate limit handling or backoff strategy for API calls

**Recommendation**: Add integration tests against real GitHub API (optional test suite with API token) to validate error handling behavior across different permission scenarios.
