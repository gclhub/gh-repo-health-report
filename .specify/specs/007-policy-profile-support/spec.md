# Feature Specification: Policy Profile Support

**Feature Branch**: `007-policy-profile-support`  
**Created**: 2026-06-01  
**Status**: Draft  
**Input**: User description: "Create a specification for adding policy profile support to gh-repo-health-report."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Tailor Health Checks by Repository Type (Priority: P1)

As a platform engineering lead, I want to apply different health check expectations based on repository type (open-source library vs. internal service vs. prototype), so that health reports reflect realistic governance policies for each context.

**Why this priority**: This is the core value proposition. Different repository types have fundamentally different governance needs (e.g., an archived repository shouldn't be penalized for lacking active CI, a prototype doesn't need CODEOWNERS). Without this, the tool produces misleading health scores.

**Independent Test**: Can be fully tested by running `gh repo-health-report --profile open-source --repo owner/my-lib` and verifying that checks marked as "ignored" in the open-source profile don't appear in the failed checks list, while "required" checks do. Delivers immediate value by eliminating false positives.

**Acceptance Scenarios**:

1. **Given** a user audits an open-source library repository, **When** they run `gh repo-health-report --profile open-source --repo owner/library`, **Then** the health report requires README, LICENSE, CODE_OF_CONDUCT, CONTRIBUTING, and security scanning, but ignores CODEOWNERS and PROJECT checks.

2. **Given** a user audits an internal service repository, **When** they run `gh repo-health-report --profile internal-service --repo owner/api-service`, **Then** the health report requires CODEOWNERS, CI workflows, branch protection, and secret scanning, but ignores CODE_OF_CONDUCT and CONTRIBUTING.

3. **Given** a user audits a prototype repository, **When** they run `gh repo-health-report --profile prototype --repo owner/experiment`, **Then** the health report requires only README and basic metadata, ignoring most governance checks like CODEOWNERS, CI, and branch protection.

4. **Given** a user audits an archived repository, **When** they run `gh repo-health-report --profile archived --repo owner/legacy-project`, **Then** the health report requires README and LICENSE for documentation, but ignores staleness, CI, branch protection, and security scanning checks.

5. **Given** a user runs the tool without specifying a profile, **When** they execute `gh repo-health-report --repo owner/repo`, **Then** all existing checks are evaluated with existing behavior (backward compatibility).

---

### User Story 2 - Set Organization-Wide Profile Defaults (Priority: P2)

As a platform engineering manager, I want to define a default profile in a configuration file that applies to all repositories in my organization, so that individual users don't need to remember to specify the profile flag each time.

**Why this priority**: Reduces cognitive load and ensures consistency across teams. After P1 proves the profile concept works, organizations need a way to encode their policy centrally rather than relying on each user to pass the correct flag.

**Independent Test**: Can be tested by creating a config file with `default_profile: internal-service`, running `gh repo-health-report --org myorg` without a profile flag, and verifying all repos are evaluated against the internal-service profile. Delivers value by reducing repetitive flag usage.

**Acceptance Scenarios**:

1. **Given** a configuration file exists with `default_profile: internal-service`, **When** a user runs `gh repo-health-report --org myorg`, **Then** all repositories are evaluated using the internal-service profile without requiring the `--profile` flag.

2. **Given** a configuration file specifies a default profile, **When** a user explicitly provides `--profile open-source`, **Then** the command-line flag overrides the config file default.

3. **Given** no configuration file exists and no profile flag is provided, **When** a user runs the tool, **Then** it falls back to legacy behavior (all checks enabled).

---

### User Story 3 - Understand Why Checks Were Skipped (Priority: P2)

As a developer reviewing my repository's health report, I want to see explanations for why certain checks were skipped or ignored, so that I understand the profile's policy decisions.

**Why this priority**: Transparency builds trust in the tool. Users need to understand why a check didn't run (was it ignored by policy, or did it fail due to missing data?). This prevents confusion and helps teams learn their organization's governance expectations.

**Independent Test**: Can be tested by running `gh repo-health-report --profile archived --format table --repo owner/old-repo` and verifying that the output includes indicators showing which checks were skipped due to the profile (e.g., `[IGNORED]` or `SKIPPED` in the table, or explanatory keys in JSON output). Delivers value by improving report comprehension.

**Acceptance Scenarios**:

1. **Given** a repository is evaluated with a profile that ignores specific checks, **When** the user views the table output, **Then** skipped checks are marked with `[IGNORED]` or similar notation in their columns.

2. **Given** a repository is evaluated with a profile that ignores specific checks, **When** the user requests JSON output, **Then** the output includes a `"skipped_checks"` field listing check names and reasons (e.g., `{"check": "missing-codeowners", "reason": "Ignored by profile: open-source"}`).

3. **Given** a user reviews a health report, **When** they see a check marked as `[IGNORED]`, **Then** they can reference the profile documentation to understand why that check doesn't apply to their repository type.

---

### User Story 4 - Define Profile-Aware Fail Thresholds (Priority: P3)

As a CI/CD pipeline maintainer, I want to use the `--fail-on` flag with profile-aware scoring, so that the exit code reflects only the checks that matter for my repository type.

**Why this priority**: Extends the existing `--fail-on` feature to work correctly with profiles. Without this, users might accidentally fail builds on checks their profile ignores, or miss failures on checks their profile requires. Lower priority because it builds on P1's foundation and many users may not use `--fail-on` initially.

**Independent Test**: Can be tested by running `gh repo-health-report --profile prototype --fail-on any --repo owner/experiment` where a prototype-ignored check fails, and verifying the command exits 0 (because ignored checks don't contribute to failure). Then test with a required check failing and verify exit code 1. Delivers value by making CI pipelines respect profile policies.

**Acceptance Scenarios**:

1. **Given** a user runs `gh repo-health-report --profile archived --fail-on stale --repo owner/old-project`, **When** the archived profile ignores the `stale` check, **Then** the command exits 0 even if the repository is stale.

2. **Given** a user runs `gh repo-health-report --profile internal-service --fail-on any --repo owner/service`, **When** a required check (e.g., `no-branch-protection`) fails, **Then** the command exits 1.

3. **Given** a user runs `gh repo-health-report --profile internal-service --fail-on any --repo owner/service`, **When** an ignored check (e.g., `missing-code-of-conduct`) fails, **Then** the command exits 0.

4. **Given** a user runs `gh repo-health-report --fail-on any` without a profile, **When** any check fails, **Then** the command exits 1 (backward compatibility with existing behavior).

---

### User Story 5 - Auto-Detect Profile from Repository Metadata (Priority: P3)

As a platform user with diverse repositories, I want the tool to automatically detect the appropriate profile based on repository metadata (archived status, topics, naming conventions), so that I don't need to manually specify profiles for every repository.

**Why this priority**: Nice-to-have automation that reduces manual work, but optional since P1 and P2 provide explicit control. Auto-detection can introduce edge cases and requires careful heuristics. Lower priority to validate P1-P3 work well before adding inference logic.

**Independent Test**: Can be tested by running `gh repo-health-report --profile auto --repo owner/archived-repo` (where the repository is marked archived on GitHub) and verifying it automatically uses the archived profile. Also test with a repo containing topic `prototype` and verify it selects the prototype profile. Delivers value by reducing the need to track which profile applies to which repo.

**Acceptance Scenarios**:

1. **Given** a repository is marked as archived on GitHub, **When** a user runs `gh repo-health-report --profile auto --repo owner/archived-repo`, **Then** the tool automatically applies the `archived` profile.

2. **Given** a repository has topic `prototype` or `experimental`, **When** a user runs `gh repo-health-report --profile auto --org myorg`, **Then** the tool applies the `prototype` profile to that repository.

3. **Given** a repository has topic `library` or `package`, **When** auto-detection is enabled, **Then** the tool applies the `open-source` profile.

4. **Given** a repository doesn't match any auto-detection heuristics, **When** auto-detection is enabled, **Then** the tool falls back to the `internal-service` profile or the configured default.

5. **Given** a user explicitly provides `--profile open-source`, **When** auto-detection would choose a different profile, **Then** the explicit flag takes precedence.

---

### Edge Cases

- What happens when a user specifies a non-existent profile name (e.g., `--profile unknown-profile`)? The tool should exit with an error listing available profiles.
- How does the system handle a repository that has both `archived` status and `prototype` topic during auto-detection? Archived status takes precedence (archival is a more definitive state).
- What happens when a profile ignores a check that a user explicitly requests via `--fail-on missing-codeowners`? The explicit `--fail-on` flag overrides the profile's ignore setting (user intent is clear).
- How does the tool behave when a profile is defined with conflicting states (e.g., a check is both "required" and "ignored")? This is a configuration error; the tool should reject the profile definition at startup.
- What happens when API permissions are insufficient to determine if a check passes (e.g., secret scanning requires admin access), and that check is marked "required" in the profile? The check result is "Unknown" (`?`), not counted as a failure, but flagged in output for user attention.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support five predefined profiles: `open-source`, `internal-service`, `application`, `archived`, and `prototype`.
- **FR-002**: System MUST allow users to specify a profile via the `--profile` CLI flag.
- **FR-003**: System MUST allow users to define a default profile in a configuration file that applies when no `--profile` flag is provided.
- **FR-004**: System MUST maintain backward compatibility by evaluating all 28 checks with existing behavior when no profile is specified and no config default exists.
- **FR-005**: Each profile MUST define which of the 28 existing checks are "required", "recommended", or "ignored".
- **FR-006**: System MUST evaluate only the checks marked "required" or "recommended" for the active profile; "ignored" checks MUST be skipped entirely.
- **FR-007**: System MUST include profile-aware scoring that counts only non-ignored checks when computing pass/fail ratios.
- **FR-008**: System MUST indicate skipped checks in all output formats (table, JSON, CSV, markdown) with clear notation (e.g., `[IGNORED]`, `SKIPPED`, or a dedicated field).
- **FR-009**: System MUST respect the existing `--fail-on` flag behavior, but only trigger exit code 1 for checks that are not ignored by the active profile.
- **FR-010**: System MUST allow the `--fail-on` flag to override profile settings when a user explicitly requests a specific check (user intent takes precedence).
- **FR-011**: System MUST support optional auto-detection mode (`--profile auto`) that infers the appropriate profile from repository metadata (archived status, topics, naming patterns).
- **FR-012**: System MUST prioritize profile selection in this order: explicit `--profile` flag > config file default > legacy behavior (all checks).
- **FR-013**: System MUST report an error and exit if a user specifies a non-existent profile name.
- **FR-014**: System MUST provide clear documentation for each predefined profile, listing which checks are required/recommended/ignored and the rationale.

### Key Entities

- **Profile**: A named configuration defining governance expectations for a repository type. Contains a name (string), description (string), and a mapping of check names to enforcement levels ("required", "recommended", "ignored").
  
- **Check Enforcement Level**: An enumeration (`required`, `recommended`, `ignored`) that determines whether a check is evaluated and how it impacts scoring:
  - `required`: Check must pass; failure contributes to fail count and affects exit codes.
  - `recommended`: Check is evaluated and reported but doesn't cause strict failures (informational).
  - `ignored`: Check is skipped entirely; not evaluated or displayed in failure lists.

- **Profile Configuration**: A file (JSON or YAML format) that stores the default profile setting for an organization or user. Located at `.gh-repo-health-report.yml` or similar convention.

### CLI Interface Requirements

- **CLI-001**: New flag: `--profile [name]` (string) — Specifies which profile to apply. Valid values: `open-source`, `internal-service`, `application`, `archived`, `prototype`, `auto`. No default (legacy behavior if omitted).
- **CLI-002**: Configuration file support: Tool reads `.gh-repo-health-report.yml` or `.gh-repo-health-report.json` in the current directory or user's home directory to load `default_profile` setting.
- **CLI-003**: Output format support: All formats (table, JSON, CSV, markdown) must indicate skipped checks:
  - Table: Use notation like `[IGNORED]` or `SKIP` in column cells for skipped checks.
  - JSON: Add a `"skipped_checks"` array field to each repository result object, listing check names and reasons.
  - CSV: Add a `"skipped_checks"` column with comma-separated check names.
  - Markdown: Similar to table, use notation in cells or a separate section listing skipped checks.
- **CLI-004**: Exit code behavior: Exit code 1 only when checks marked "required" or "recommended" fail AND they are specified in `--fail-on` (or `--fail-on any`). Ignored checks never cause exit code 1 unless explicitly overridden.
- **CLI-005**: Backward compatibility: All existing flags continue to work. When `--profile` is omitted and no config default exists, behavior is identical to current version (all 28 checks evaluated). `--max-branches` and `--max-tags` thresholds continue to work and apply to their respective checks if those checks are not ignored.
- **CLI-006**: Help output: `gh repo-health-report --help` must document the new `--profile` flag and list available profile names with brief descriptions.

### Profile Definitions

Each profile defines enforcement levels for all 28 checks. "Required" checks must pass or the repository fails that check. "Recommended" checks are informational but don't cause strict failures. "Ignored" checks are skipped entirely.

#### Open-Source Library Profile (`open-source`)
**Purpose**: Public libraries focused on community collaboration and transparency.

- **Required**: `missing-readme`, `missing-license`, `missing-code-of-conduct`, `missing-contributing`, `missing-security`, `has-description`, `has-issues`, `no-secret-scanning`, `no-vulnerability-alerts`
- **Recommended**: `missing-issue-templates`, `missing-pr-template`, `missing-ci`, `has-homepage`, `stale`, `too-many-branches`, `has-stale-branches`
- **Ignored**: `missing-codeowners`, `missing-dependabot`, `has-projects`, `has-wiki`, `no-branch-protection`, `no-rulesets`, `no-push-protection`, `no-delete-branch-on-merge`, `too-many-tags`

**Rationale**: Open-source projects prioritize community health files (CODE_OF_CONDUCT, CONTRIBUTING) and security transparency. CODEOWNERS and strict branch protection are less critical for community projects with diverse contributor models.

#### Internal Service Profile (`internal-service`)
**Purpose**: Production services and APIs maintained by internal teams.

- **Required**: `missing-readme`, `missing-codeowners`, `missing-ci`, `no-branch-protection`, `no-rulesets`, `no-vulnerability-alerts`, `no-secret-scanning`, `no-push-protection`, `no-delete-branch-on-merge`, `has-description`
- **Recommended**: `missing-license`, `missing-security`, `missing-dependabot`, `stale`, `too-many-branches`, `has-stale-branches`, `missing-issue-templates`
- **Ignored**: `missing-code-of-conduct`, `missing-contributing`, `missing-pr-template`, `has-homepage`, `has-issues`, `has-projects`, `has-wiki`, `too-many-tags`

**Rationale**: Internal services prioritize operational reliability (CI, branch protection, CODEOWNERS) and security (secret scanning, push protection). Community health files (CODE_OF_CONDUCT) are not applicable to internal-only repositories.

#### Application Profile (`application`)
**Purpose**: End-user applications (web apps, mobile apps, desktop software).

- **Required**: `missing-readme`, `missing-license`, `missing-ci`, `no-vulnerability-alerts`, `no-secret-scanning`, `has-description`
- **Recommended**: `missing-security`, `missing-dependabot`, `missing-codeowners`, `no-branch-protection`, `no-rulesets`, `stale`, `too-many-branches`, `has-stale-branches`, `has-homepage`
- **Ignored**: `missing-code-of-conduct`, `missing-contributing`, `missing-issue-templates`, `missing-pr-template`, `has-issues`, `has-projects`, `has-wiki`, `no-push-protection`, `no-delete-branch-on-merge`, `too-many-tags`

**Rationale**: Applications balance user-facing documentation (README, LICENSE) with development hygiene (CI, security scanning). Strict governance (CODEOWNERS, branch protection) is recommended but not required, allowing flexibility for smaller teams.

#### Archived Repository Profile (`archived`)
**Purpose**: Repositories no longer under active development but retained for reference.

- **Required**: `missing-readme`, `missing-license`
- **Recommended**: `has-description`, `has-homepage`
- **Ignored**: `missing-code-of-conduct`, `missing-codeowners`, `missing-security`, `missing-contributing`, `missing-issue-templates`, `missing-pr-template`, `stale`, `has-issues`, `has-projects`, `has-wiki`, `missing-dependabot`, `missing-ci`, `no-branch-protection`, `no-rulesets`, `no-vulnerability-alerts`, `no-secret-scanning`, `no-push-protection`, `no-delete-branch-on-merge`, `too-many-branches`, `has-stale-branches`, `too-many-tags`

**Rationale**: Archived repositories only need basic documentation for historical reference. Staleness, CI, and governance checks are irrelevant since no active development is expected.

#### Prototype Profile (`prototype`)
**Purpose**: Experimental or proof-of-concept repositories for exploration and learning.

- **Required**: `missing-readme`, `has-description`
- **Recommended**: `missing-license`, `has-homepage`
- **Ignored**: `missing-code-of-conduct`, `missing-codeowners`, `missing-security`, `missing-contributing`, `missing-issue-templates`, `missing-pr-template`, `stale`, `has-issues`, `has-projects`, `has-wiki`, `missing-dependabot`, `missing-ci`, `no-branch-protection`, `no-rulesets`, `no-vulnerability-alerts`, `no-secret-scanning`, `no-push-protection`, `no-delete-branch-on-merge`, `too-many-branches`, `has-stale-branches`, `too-many-tags`

**Rationale**: Prototypes focus on rapid experimentation. Only minimal documentation (README describing purpose) is required. All governance and security checks are ignored to avoid burdening exploratory work.

### Auto-Detection Heuristics

When `--profile auto` is specified, the system applies these rules in order:

1. **Archived status**: If repository is marked archived on GitHub → use `archived` profile
2. **Topic matching**: 
   - Topics `prototype`, `experimental`, `poc`, `spike` → use `prototype` profile
   - Topics `library`, `package`, `npm-package`, `gem`, `pypi` → use `open-source` profile
   - Topics `service`, `api`, `microservice` → use `internal-service` profile
   - Topics `app`, `webapp`, `mobile-app`, `desktop` → use `application` profile
3. **Visibility**: Public repositories without matching topics → use `open-source` profile
4. **Default fallback**: Private repositories without matching topics → use `internal-service` profile (or configured default)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can specify a profile via `--profile [name]` and see different health check results based on profile rules in under 10 seconds per repository.
- **SC-002**: Organizations can set a default profile in a config file, eliminating the need for users to specify `--profile` on 90%+ of executions.
- **SC-003**: Health reports clearly indicate skipped checks in all output formats (table, JSON, CSV, markdown), reducing user confusion about missing check results by 100%.
- **SC-004**: The tool maintains 100% backward compatibility: existing users who don't specify a profile see identical behavior to the current version (all 28 checks evaluated).
- **SC-005**: Profile-aware `--fail-on` logic correctly exits with code 1 only for required/recommended check failures, not for ignored checks, reducing false positive CI failures by 80%+ for teams using profiles.
- **SC-006**: Auto-detection (`--profile auto`) correctly infers the appropriate profile for 85%+ of repositories based on metadata, measured by user acceptance testing across diverse repository samples.

## Assumptions

- Users understand their repository types and can classify them into one of the five predefined profiles (open-source, internal-service, application, archived, prototype). If not, auto-detection provides a fallback.
- Configuration file is optional; users who prefer explicit CLI flags can continue using `--profile` without a config file.
- The five predefined profiles cover the vast majority of use cases. Custom profile definitions (beyond the five predefined) are out of scope for initial version.
- GitHub API access and permissions remain unchanged; no new API calls are required beyond existing checks. Profile logic is purely client-side filtering and scoring.
- The tool continues to use the existing Cobra CLI framework and internal package structure (`internal/api`, `internal/checks`, `internal/formatter`).
- Output format changes (adding skipped check indicators) are acceptable as they add information without removing existing data.
- The existing tristate check results (Pass/Fail/Unknown) remain unchanged; "Unknown" results (e.g., `?` for security settings requiring admin access) are not affected by profile logic.
- Auto-detection heuristics are best-effort and may require tuning based on user feedback; explicit profile flags always override auto-detection.
- The config file format (YAML or JSON) will follow common conventions (e.g., `.gh-repo-health-report.yml` in current directory or home directory), similar to tools like `eslint` or `prettier`.
