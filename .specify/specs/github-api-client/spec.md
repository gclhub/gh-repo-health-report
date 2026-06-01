# Feature Specification: GitHub API Client

**Feature Branch**: N/A (existing feature)  
**Created**: 2026-06-01  
**Status**: migrated  
**Input**: Reverse-engineered from existing implementation in `internal/api/`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Repository Fetching (Priority: P1) 🎯 MVP

A user wants to fetch repository health data from GitHub, whether for a single repository, all repositories in an organization, or all repositories owned by a user, with automatic pagination handling for large result sets.

**Why this priority**: Core functionality — without repository fetching, no health checks are possible.

**Independent Test**: Run `gh repo-health-report --repo owner/name` to fetch single repo, `--org orgname` for org repos, `--owner username` for user repos. Verify all repos returned with pagination working.

**Acceptance Scenarios**:

1. **Given** a repository name `owner/repo`, **When** fetching via `GetRepo()`, **Then** repository metadata is returned with all fields populated
2. **Given** an organization name, **When** fetching via `ListOrgRepos()`, **Then** all non-archived, non-forked repos are returned (unless flags specify otherwise)
3. **Given** a user name, **When** fetching via `ListUserRepos()`, **Then** all repositories owned by that user are returned
4. **Given** `--include-forks` flag, **When** fetching org repos, **Then** forked repositories are included in results
5. **Given** `--include-archived` flag, **When** fetching org repos, **Then** archived repositories are included in results
6. **Given** an org with 250 repositories, **When** fetching via `ListOrgRepos()`, **Then** pagination automatically handles multiple pages (100 repos per page)

---

### User Story 2 - Community File Detection (Priority: P1) 🎯 MVP

A user wants to check whether repositories contain essential community health files (README, LICENSE, CODE_OF_CONDUCT, CODEOWNERS, SECURITY, CONTRIBUTING, issue/PR templates) by querying GitHub's Contents API and dedicated file endpoints.

**Why this priority**: Core value proposition — community file detection is the primary audit function.

**Independent Test**: Call `PopulateFileChecks()` on a repo and verify all community file boolean flags are correctly set based on GitHub API responses.

**Acceptance Scenarios**:

1. **Given** a repository with README.md, **When** calling GitHub `/readme` endpoint, **Then** `HasReadme` is set to `true`
2. **Given** a repository with LICENSE file, **When** calling GitHub `/license` endpoint, **Then** `HasLicense` is set to `true`
3. **Given** CODE_OF_CONDUCT.md in `.github/` directory, **When** checking file locations, **Then** `HasCodeOfConduct` is set to `true`
4. **Given** CODEOWNERS in root directory, **When** checking file locations, **Then** `HasCodeowners` is set to `true`
5. **Given** SECURITY.md missing, **When** checking all locations (root, `.github/`), **Then** `HasSecurity` remains `false`
6. **Given** issue templates in `.github/ISSUE_TEMPLATE/` directory, **When** checking file existence, **Then** `HasIssueTemplates` is set to `true`
7. **Given** PR template in `.github/PULL_REQUEST_TEMPLATE.md`, **When** checking file locations, **Then** `HasPRTemplate` is set to `true`
8. **Given** 404 error when checking file, **When** handling error, **Then** file is considered absent (no error propagated)

---

### User Story 3 - Extended Security & Automation Checks (Priority: P2)

A user wants to check for automation configuration (Dependabot, CI workflows) and security settings (branch protection, rulesets, vulnerability alerts, secret scanning, push protection) to ensure repository best practices.

**Why this priority**: Important for security compliance but depends on elevated permissions; gracefully degrades when access denied.

**Independent Test**: Call `PopulateExtendedChecks()` on repos with varying permission levels and verify security settings correctly detected or marked as unknown.

**Acceptance Scenarios**:

1. **Given** `.github/dependabot.yml` exists, **When** checking file, **Then** `HasDependabot` is set to `true`
2. **Given** `.github/workflows/` directory with files, **When** listing directory contents, **Then** `HasCIWorkflows` is set to `true`
3. **Given** default branch has protection rules, **When** calling `/branches/{branch}/protection` endpoint, **Then** `DefaultBranchProtected` is set to `true`
4. **Given** repository has rulesets configured, **When** calling `/rulesets` endpoint, **Then** `HasRulesets` is set to `true` (non-empty array)
5. **Given** vulnerability alerts enabled and user has admin access, **When** calling `/vulnerability-alerts`, **Then** `VulnerabilityAlertsEnabled` is `true`
6. **Given** user lacks admin access, **When** calling `/vulnerability-alerts` returns 403, **Then** `VulnerabilityAlertsUnknown` is `true`
7. **Given** secret scanning enabled and visible in `security_and_analysis` field, **When** parsing response, **Then** `SecretScanningEnabled` is `true`
8. **Given** `security_and_analysis` field absent (insufficient permissions), **When** parsing response, **Then** `SecretScanningUnknown` is `true`
9. **Given** 404 on branch protection check, **When** handling error, **Then** protection is considered disabled (not unknown)

---

### User Story 4 - Branch & Tag Analysis (Priority: P2)

A user wants to count total branches and tags, and identify stale branches (branches with no commits since a threshold date), to detect repository maintenance issues.

**Why this priority**: Valuable for maintenance audits but computationally expensive (requires many API calls); can be slow for large repos.

**Independent Test**: Call `PopulateBranchTagChecks()` on a repo with multiple branches and verify counts are accurate, stale branches correctly identified.

**Acceptance Scenarios**:

1. **Given** a repository with 10 branches, **When** paginating branch list, **Then** `BranchCount` is set to 10
2. **Given** a repository with 3 non-default branches with no commits in 200 days, **When** checking commits since 180 days ago, **Then** `StaleBranchCount` is set to 3
3. **Given** default branch, **When** evaluating staleness, **Then** default branch is excluded from stale branch count
4. **Given** a repository with 150 tags, **When** paginating tag list, **Then** `TagCount` is set to 150
5. **Given** pagination with 250 branches, **When** fetching branches, **Then** multiple pages are fetched (100 per page) until all counted
6. **Given** API rate limit hit during branch checks, **When** error occurs on individual branch, **Then** branch is skipped (total count maintained, stale count may be under-reported)

---

### User Story 5 - Error Handling & Permissions (Priority: P2)

A user wants the API client to gracefully handle insufficient permissions (403 Forbidden), missing resources (404 Not Found), and transient errors without aborting the entire audit.

**Why this priority**: Real-world GitHub API access varies by user permissions; must degrade gracefully.

**Independent Test**: Call API methods against repos with various permission levels and verify appropriate error handling (mark unknown vs. fail).

**Acceptance Scenarios**:

1. **Given** 404 error on file check, **When** processing response, **Then** file is considered absent (returns `false, nil`)
2. **Given** 403 error on branch protection check, **When** processing response, **Then** protection is considered disabled (not an error)
3. **Given** 403 error on vulnerability alerts check, **When** processing response, **Then** status marked as unknown (not an error)
4. **Given** non-404/403 error on API call, **When** processing response, **Then** error is propagated to caller
5. **Given** transient error during stale branch check for one branch, **When** error occurs, **Then** branch is skipped (no audit failure)
6. **Given** invalid API response format, **When** parsing JSON, **Then** error is returned to caller

---

### User Story 6 - Client Initialization & Authentication (Priority: P1) 🎯 MVP

A user wants the API client to automatically authenticate using GitHub CLI credentials, leveraging the existing `gh` auth token without requiring separate configuration.

**Why this priority**: Must work out-of-the-box as a GitHub CLI extension; authentication is prerequisite for all API calls.

**Independent Test**: Run `NewClient()` after `gh auth login` and verify REST client is initialized with valid credentials.

**Acceptance Scenarios**:

1. **Given** user is authenticated via `gh auth login`, **When** calling `NewClient()`, **Then** REST client is created with auth token from gh
2. **Given** user is not authenticated, **When** calling `NewClient()`, **Then** error is returned explaining authentication required
3. **Given** multiple GitHub hosts configured, **When** creating client, **Then** default GitHub.com host is used
4. **Given** valid REST client, **When** calling API methods, **Then** requests include authentication headers automatically

---

### Edge Cases

- What happens when a repository is deleted between listing and fetching? → API returns 404; individual fetch fails but doesn't abort batch processing
- How does system handle empty organizations (no repositories)? → Returns empty slice, no error
- What happens when a file exists in multiple checked locations? → First match is accepted; subsequent locations not checked
- How are directory vs. file checks distinguished in Contents API? → Both return 200; system treats any 200 as "exists"
- What happens when branch/tag count exceeds 1000 (10+ pages)? → Continues pagination until API returns <100 results; may be slow but accurate
- How are API rate limits handled? → No explicit rate limit handling; relies on go-gh client's built-in handling (may return errors)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST authenticate using GitHub CLI credentials via `gh` token
- **FR-002**: System MUST fetch single repository metadata via `/repos/{owner}/{repo}` endpoint
- **FR-003**: System MUST fetch all org repositories via `/orgs/{org}/repos` with pagination (100 per page)
- **FR-004**: System MUST fetch all user repositories via `/users/{user}/repos` with pagination
- **FR-005**: System MUST filter out forks unless `--include-forks` flag is set
- **FR-006**: System MUST filter out archived repos unless `--include-archived` flag is set
- **FR-007**: System MUST check README existence via dedicated `/readme` endpoint
- **FR-008**: System MUST check LICENSE existence via dedicated `/license` endpoint
- **FR-009**: System MUST check CODE_OF_CONDUCT.md in three locations: root, `.github/`, `docs/`
- **FR-010**: System MUST check CODEOWNERS in three locations: `.github/`, root, `docs/`
- **FR-011**: System MUST check SECURITY.md in two locations: root, `.github/`
- **FR-012**: System MUST check CONTRIBUTING.md in two locations: root, `.github/`
- **FR-013**: System MUST check issue templates in directory `.github/ISSUE_TEMPLATE/` and files `ISSUE_TEMPLATE.md`, `.github/ISSUE_TEMPLATE.md`
- **FR-014**: System MUST check PR templates in four locations: `.github/PULL_REQUEST_TEMPLATE.md`, `.github/PULL_REQUEST_TEMPLATE`, `PULL_REQUEST_TEMPLATE.md`, `docs/PULL_REQUEST_TEMPLATE.md`
- **FR-015**: System MUST check Dependabot config for both `.yml` and `.yaml` extensions
- **FR-016**: System MUST list `.github/workflows/` directory contents to detect CI workflows
- **FR-017**: System MUST check default branch protection via `/branches/{branch}/protection` endpoint
- **FR-018**: System MUST check repository rulesets via `/rulesets` endpoint
- **FR-019**: System MUST check vulnerability alerts via `/vulnerability-alerts` endpoint
- **FR-020**: System MUST parse `security_and_analysis` field for secret scanning and push protection status
- **FR-021**: System MUST paginate branch list via `/branches` endpoint (100 per page)
- **FR-022**: System MUST check for commits on each non-default branch since provided timestamp
- **FR-023**: System MUST paginate tag list via `/tags` endpoint (100 per page)
- **FR-024**: System MUST treat 404 errors as "not found" without propagating error (for file checks)
- **FR-025**: System MUST treat 403 errors as "unknown" for permission-restricted checks (vulnerability alerts)
- **FR-026**: System MUST treat 403 errors as "disabled" for branch protection (legacy behavior)
- **FR-027**: System MUST handle transient errors during batch operations by skipping failed items

### Key Entities *(include if feature involves data)*

- **Client**: Wraps go-gh REST client for GitHub API interactions
  - Initialized from default GitHub CLI authentication
  - Provides high-level methods for repository operations
  
- **Repository**: Represents GitHub repository with health check fields
  - Basic metadata: full_name, owner, description, homepage, topics, timestamps
  - Feature flags: has_issues, has_projects, has_wiki
  - Security fields: security_and_analysis with nested secret scanning/push protection
  - Populated fields: File checks, extended checks, branch/tag counts
  - 50+ fields total combining API response and computed values

### GitHub API Requirements *(for API-dependent features)*

**Core Endpoints**:
- `GET /repos/{owner}/{repo}` — Single repository fetch
- `GET /orgs/{org}/repos` — Organization repository list (paginated)
- `GET /users/{user}/repos` — User repository list (paginated)
- `GET /repos/{owner}/{repo}/readme` — README detection
- `GET /repos/{owner}/{repo}/license` — LICENSE detection
- `GET /repos/{owner}/{repo}/contents/{path}` — File existence checks

**Extended Endpoints**:
- `GET /repos/{owner}/{repo}/branches` — Branch list (paginated)
- `GET /repos/{owner}/{repo}/branches/{branch}/protection` — Branch protection status
- `GET /repos/{owner}/{repo}/rulesets` — Repository rulesets
- `GET /repos/{owner}/{repo}/vulnerability-alerts` — Dependabot alerts status
- `GET /repos/{owner}/{repo}/commits?sha={branch}&since={timestamp}&per_page=1` — Recent commits check
- `GET /repos/{owner}/{repo}/tags` — Tag list (paginated)

**Required Permissions**:
- `repo` scope: Read repository metadata and contents (minimum required)
- `push` or `admin`: Required for `security_and_analysis` field visibility
- `admin`: Required for vulnerability alerts endpoint (returns 403 otherwise)

**Rate Limiting**:
- Single repo: ~5-10 API calls (readme, license, contents checks)
- Org/user repos: 1 call per 100 repos + per-repo checks
- Branch staleness: 1 additional call per non-default branch (can be expensive)
- Typical audit: 50-100 API calls per repository for full checks
- GitHub API rate limit: 5000 requests/hour for authenticated users

**Data Fetching Strategy**:
- REST API (via go-gh library)
- Pagination: Automatic via `per_page=100&page={n}` query parameters
- Error handling: Distinguish 404 (not found) vs. 403 (forbidden) vs. other errors
- Batching: None — each repository checked independently

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Single repository fetch completes in under 5 seconds for repos with standard files
- **SC-002**: Organization with 100 repositories fetches all repos in under 60 seconds (excluding per-repo checks)
- **SC-003**: File existence checks correctly identify files in all documented locations (8+ patterns checked)
- **SC-004**: System gracefully handles 403 Forbidden errors without failing entire audit
- **SC-005**: Pagination correctly handles organizations with 500+ repositories
- **SC-006**: Security settings marked as "unknown" when user lacks permissions (not incorrectly failed)
- **SC-007**: Branch staleness check completes for repos with up to 50 branches without timeout

## Assumptions

- Users have GitHub CLI (`gh`) installed and authenticated
- API rate limits (5000/hour) are sufficient for typical organization audits
- GitHub API response formats remain stable
- File existence checks via Contents API work consistently for both files and directories
- Users accept performance impact of branch staleness checks (1 API call per non-default branch)
- Transient API errors during branch checks are acceptable (tool reports partial data)
- `security_and_analysis` field availability indicates push/admin access (no explicit permission check)
- GitHub.com is the target host (no GitHub Enterprise Server support specified)
- go-gh library handles authentication token refresh automatically
