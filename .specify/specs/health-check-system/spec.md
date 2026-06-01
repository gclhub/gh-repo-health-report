# Feature Specification: Health Check System

**Feature Branch**: N/A (existing feature)  
**Created**: 2026-06-01  
**Status**: migrated  
**Input**: Reverse-engineered from existing implementation in `internal/checks/`

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Community File Detection (Priority: P1) 🎯 MVP

A repository maintainer wants to quickly identify whether their repositories contain essential community health files (README, LICENSE, CODE_OF_CONDUCT, CODEOWNERS, SECURITY, CONTRIBUTING, issue/PR templates) to ensure compliance with open-source best practices and organizational policies.

**Why this priority**: Core value proposition of the tool; most organizations require these files for legal compliance (LICENSE), security disclosure (SECURITY.md), and contributor guidance.

**Independent Test**: Run `gh repo-health-report --repo owner/name` and verify output shows ✓ or ✗ for each community file check. Can be tested against any public repository.

**Acceptance Scenarios**:

1. **Given** a repository with README.md and LICENSE, **When** health check runs, **Then** displays ✓ for `missing-readme` and `missing-license` checks
2. **Given** a repository without CODE_OF_CONDUCT.md, **When** health check runs, **Then** displays ✗ for `missing-code-of-conduct` check
3. **Given** a repository with CODEOWNERS in `.github/`, **When** health check runs, **Then** displays ✓ for `missing-codeowners` check
4. **Given** a repository with issue templates directory, **When** health check runs, **Then** displays ✓ for `missing-issue-templates` check

---

### User Story 2 - Repository Metadata Assessment (Priority: P1) 🎯 MVP

A repository auditor needs to verify that repositories have proper metadata (description, homepage URL, topics) and enabled features (issues, wiki, projects) to ensure discoverability and usability.

**Why this priority**: Part of MVP — essential for understanding repository configuration and ensuring proper documentation.

**Independent Test**: Run against a repo with incomplete metadata and verify stale status, description, and feature flags are correctly evaluated.

**Acceptance Scenarios**:

1. **Given** a repository with no description, **When** health check runs, **Then** `has-description` check fails
2. **Given** a repository with issues disabled, **When** health check runs, **Then** `has-issues` check fails
3. **Given** a repository not pushed to in 200 days and `--since 180d`, **When** health check runs, **Then** `stale` check fails and displays "YES"
4. **Given** a repository with topics configured, **When** health check runs, **Then** topics count is displayed

---

### User Story 3 - Security & Automation Checks (Priority: P2)

An organization security officer wants to audit whether repositories have security features enabled (secret scanning, push protection, vulnerability alerts, branch protection, rulesets) and automation configured (Dependabot, CI workflows).

**Why this priority**: Critical for security compliance but requires admin/push access; can fail gracefully with "?" when permissions insufficient.

**Independent Test**: Run against an org's repos with admin access and verify security settings are correctly detected or show "?" when access denied.

**Acceptance Scenarios**:

1. **Given** a repository with `.github/dependabot.yml`, **When** health check runs, **Then** `missing-dependabot` check passes
2. **Given** a repository with workflows in `.github/workflows/`, **When** health check runs, **Then** `missing-ci` check passes
3. **Given** a repository with branch protection on default branch, **When** health check runs, **Then** `no-branch-protection` check passes
4. **Given** a repository with repository rulesets configured, **When** health check runs, **Then** `no-rulesets` check passes
5. **Given** a repository where user lacks admin access, **When** checking vulnerability alerts, **Then** displays "?" instead of ✓ or ✗
6. **Given** a repository with secret scanning enabled, **When** health check runs with push access, **Then** `no-secret-scanning` check passes
7. **Given** a repository with push protection enabled, **When** health check runs, **Then** `no-push-protection` check passes

---

### User Story 4 - Branch & Tag Management (Priority: P2)

A repository maintainer wants to identify repositories with excessive branches or tags, or branches that haven't been updated in a long time, to reduce clutter and maintenance overhead.

**Why this priority**: Helps identify maintenance issues but not critical for basic compliance; can be disabled via flags.

**Independent Test**: Run with custom thresholds `--max-branches 30 --max-tags 50` and verify repositories exceeding these limits are flagged.

**Acceptance Scenarios**:

1. **Given** a repository with 60 branches and `--max-branches 50`, **When** health check runs, **Then** `too-many-branches` check fails
2. **Given** a repository with 5 branches and default threshold, **When** health check runs, **Then** `too-many-branches` check passes
3. **Given** a repository with branches not updated since `--since` threshold, **When** health check runs, **Then** `has-stale-branches` check fails and displays stale branch count
4. **Given** a repository with 150 tags and `--max-tags 100`, **When** health check runs, **Then** `too-many-tags` check fails
5. **Given** `--max-branches 0`, **When** health check runs, **Then** branch count check is disabled

---

### User Story 5 - Configurable Thresholds & Exit Codes (Priority: P3)

A CI/CD pipeline operator wants to configure specific thresholds for checks and fail the build if critical checks fail, enabling automated policy enforcement.

**Why this priority**: Advanced feature for automation; not needed for basic auditing use cases.

**Independent Test**: Run with `--fail-on missing-readme,missing-license` against a repo missing README and verify exit code 1.

**Acceptance Scenarios**:

1. **Given** `--fail-on missing-readme` and a repo without README, **When** health check runs, **Then** CLI exits with code 1
2. **Given** `--fail-on any` and a repo with any failed checks, **When** health check runs, **Then** CLI exits with code 1
3. **Given** `--fail-on missing-readme` and a repo with README, **When** health check runs, **Then** CLI exits with code 0
4. **Given** custom thresholds `--since 1y`, **When** health check runs, **Then** staleness is evaluated against 365 days instead of default 180 days

---

### Edge Cases

- What happens when a file exists in multiple locations (e.g., CODE_OF_CONDUCT.md in root, `.github/`, and `docs/`)? → System checks all locations and passes if any is found
- How does system handle permission-restricted security settings? → Displays "?" for unknown status; does not count as failure unless explicitly determinable
- What happens when branch/tag pagination hits rate limits? → Continues with partial data; may under-count stale branches but reports total count
- How are archived or forked repositories handled? → Filtered out by default unless `--include-archived` or `--include-forks` flags are used
- What happens with empty repositories (no commits)? → PushedAt is zero; staleness check may behave unexpectedly — this is a known edge case

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST evaluate 25+ distinct health checks per repository
- **FR-002**: System MUST support configurable staleness threshold via `--since` flag (format: `180d`, `6m`, `1y`, `YYYY-MM-DD`)
- **FR-003**: System MUST support configurable branch count threshold via `--max-branches` flag (default: 50, 0 to disable)
- **FR-004**: System MUST support configurable tag count threshold via `--max-tags` flag (default: 100, 0 to disable)
- **FR-005**: System MUST detect README via GitHub's dedicated `/readme` API endpoint
- **FR-006**: System MUST detect LICENSE via GitHub's dedicated `/license` API endpoint
- **FR-007**: System MUST check CODE_OF_CONDUCT.md in root, `.github/`, and `docs/` directories
- **FR-008**: System MUST check CODEOWNERS in `.github/`, root, and `docs/` directories
- **FR-009**: System MUST check SECURITY.md in root and `.github/` directories
- **FR-010**: System MUST check CONTRIBUTING.md in root and `.github/` directories
- **FR-011**: System MUST detect issue templates in `.github/ISSUE_TEMPLATE/` directory or `ISSUE_TEMPLATE.md` file
- **FR-012**: System MUST detect PR templates in multiple locations (`.github/PULL_REQUEST_TEMPLATE.md`, root, docs)
- **FR-013**: System MUST evaluate repository staleness based on `pushed_at` timestamp
- **FR-014**: System MUST check for Dependabot config in `.github/dependabot.yml` or `.github/dependabot.yaml`
- **FR-015**: System MUST check for CI workflows by listing `.github/workflows/` directory contents
- **FR-016**: System MUST check default branch protection via `/branches/{branch}/protection` API
- **FR-017**: System MUST check repository rulesets via `/rulesets` API
- **FR-018**: System MUST check vulnerability alerts via `/vulnerability-alerts` API (admin access required)
- **FR-019**: System MUST check secret scanning status from `security_and_analysis` field (push/admin access required)
- **FR-020**: System MUST check push protection status from `security_and_analysis` field (push/admin access required)
- **FR-021**: System MUST track failed checks in `FailedChecks` slice for each repository
- **FR-022**: System MUST support tristate output for security checks (enabled/disabled/unknown)
- **FR-023**: System MUST exclude default branch when counting stale branches
- **FR-024**: System MUST determine branch staleness by checking for commits since `--since` threshold
- **FR-025**: System MUST use 180 days as default staleness threshold when `--since` not specified
- **FR-026**: System MUST handle 403 Forbidden errors gracefully for restricted security settings
- **FR-027**: System MUST handle 404 Not Found errors gracefully for missing resources

### Key Entities *(include if feature involves data)*

- **Result**: Represents health check results for a single repository
  - Repository reference, boolean flags for each check type
  - Counts (topics, open issues, size KB, branches, tags, stale branches)
  - Security tristate flags (enabled, unknown) for permission-restricted settings
  - FailedChecks slice containing check names that failed
  
- **Options**: Configuration for check evaluation
  - Since duration (staleness threshold)
  - MaxBranches threshold (0 = disabled)
  - MaxTags threshold (0 = disabled)

### Health Check Requirements *(for new checks)*

**Standard Check Names** (23 checks implemented):
- `has-description`, `has-homepage`
- `missing-readme`, `missing-license`, `missing-code-of-conduct`, `missing-codeowners`, `missing-security`, `missing-contributing`, `missing-issue-templates`, `missing-pr-template`
- `stale`
- `has-issues`, `has-projects`, `has-wiki`
- `missing-dependabot`, `missing-ci`
- `no-branch-protection`, `no-rulesets`
- `no-vulnerability-alerts`, `no-secret-scanning`, `no-push-protection`
- `no-delete-branch-on-merge`
- `too-many-branches`, `has-stale-branches`, `too-many-tags`

**Check Evaluation Rules**:
- **CHK-001**: Community file checks fail when file not found in any supported location
- **CHK-002**: Staleness check fails when `time.Since(repo.PushedAt) > opts.Since`
- **CHK-003**: Branch count check fails when `BranchCount > MaxBranches` (unless MaxBranches = 0)
- **CHK-004**: Tag count check fails when `TagCount > MaxTags` (unless MaxTags = 0)
- **CHK-005**: Stale branch check fails when `StaleBranchCount > 0`
- **CHK-006**: Security checks only fail when status is determinable and disabled (unknown status not counted as failure)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: System correctly evaluates all 25+ health checks against a repository in under 30 seconds per repo
- **SC-002**: System accurately detects community files in all documented locations (root, `.github/`, `docs/`)
- **SC-003**: System correctly handles permission errors (403) for security checks without failing the entire audit
- **SC-004**: System provides clear ✓/✗/? indicators for each check result in output
- **SC-005**: System correctly counts failed checks for use with `--fail-on` flag
- **SC-006**: Configurable thresholds (`--max-branches`, `--max-tags`, `--since`) correctly affect check evaluation
- **SC-007**: System processes repositories with 100+ branches without exceeding GitHub API rate limits (within pagination limits)
- **SC-008**: Stale branch detection correctly excludes the default branch from staleness evaluation

## Assumptions

- GitHub API provides consistent response formats for file checks, branch protection, and security settings
- Repositories follow standard GitHub conventions for community file locations
- Users running the tool have at least read access to target repositories
- Admin or push access is required to view certain security settings (tool degrades gracefully when unavailable)
- Branch and tag checks may be expensive for large repositories; users accept potential API rate limit impact
- Default threshold values (50 branches, 100 tags, 180 days) are reasonable for most organizations
- File existence checks via `/contents/{path}` API work consistently for both files and directories
