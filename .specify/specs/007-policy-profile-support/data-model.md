# Data Model: Policy Profile Support

**Feature**: 007-policy-profile-support  
**Date**: 2026-06-01  
**Phase**: 1 - Design

## Core Entities

### 1. EnforcementLevel (Enum)

Defines how a check is applied within a profile.

**Type**: `int` (Go iota enum)

**Values**:
- `EnforcementRequired` (0) — Check must pass; failure contributes to failed check count and affects exit codes
- `EnforcementRecommended` (1) — Check is evaluated and reported but doesn't cause strict failures (informational)
- `EnforcementIgnored` (2) — Check is skipped entirely; not evaluated or displayed in failure lists

**String Representation**:
- Required: `"required"`
- Recommended: `"recommended"`
- Ignored: `"ignored"`

**Validation Rules**:
- Must be one of the three values (0, 1, 2)
- Used as map values in Profile.Checks

**State Transitions**: None (stateless enum)

---

### 2. Profile

Defines governance expectations for a repository type.

**Fields**:
| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `Name` | `string` | Yes | Unique identifier (e.g., "open-source", "internal-service") | Lowercase, hyphen-separated, non-empty |
| `Description` | `string` | Yes | Human-readable explanation of profile purpose | Non-empty |
| `Checks` | `map[string]EnforcementLevel` | Yes | Mapping of check name → enforcement level | All 28 check names must be present |

**Example**:
```go
Profile{
    Name:        "open-source",
    Description: "Public libraries focused on community collaboration and transparency",
    Checks: map[string]EnforcementLevel{
        CheckMissingReadme:            EnforcementRequired,
        CheckMissingLicense:           EnforcementRequired,
        CheckMissingCodeOfConduct:     EnforcementRequired,
        CheckMissingContributing:      EnforcementRequired,
        CheckMissingSecurityMd:        EnforcementRequired,
        CheckHasDescription:           EnforcementRequired,
        CheckHasIssues:                EnforcementRequired,
        CheckNoSecretScanning:         EnforcementRequired,
        CheckNoVulnerabilityAlerts:    EnforcementRequired,
        CheckMissingIssueTemplates:    EnforcementRecommended,
        CheckMissingPRTemplate:        EnforcementRecommended,
        CheckMissingCI:                EnforcementRecommended,
        CheckHasHomepage:              EnforcementRecommended,
        CheckStale:                    EnforcementRecommended,
        CheckTooManyBranches:          EnforcementRecommended,
        CheckHasStaleBranches:         EnforcementRecommended,
        CheckMissingCodeowners:        EnforcementIgnored,
        CheckMissingDependabot:        EnforcementIgnored,
        CheckHasProjects:              EnforcementIgnored,
        CheckHasWiki:                  EnforcementIgnored,
        CheckNoBranchProtection:       EnforcementIgnored,
        CheckNoRulesets:               EnforcementIgnored,
        CheckNoPushProtection:         EnforcementIgnored,
        CheckNoDeleteBranchOnMerge:    EnforcementIgnored,
        CheckTooManyTags:              EnforcementIgnored,
    },
}
```

**Validation Rules**:
- All 28 existing check constants must have an enforcement level
- No duplicate check names
- No unknown check names

**Relationships**:
- Profiles are immutable once defined (no runtime modification)
- Multiple repositories can use the same profile
- Profile selected by CLI flag, config file, or auto-detection

---

### 3. Config

Configuration file data structure for default profile setting.

**Fields**:
| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `DefaultProfile` | `string` | No | Name of profile to use when no --profile flag specified | Must match a predefined profile name or be empty |

**File Format (YAML)**:
```yaml
default_profile: internal-service
```

**File Format (JSON)**:
```json
{
  "default_profile": "internal-service"
}
```

**Validation Rules**:
- If `default_profile` is set, it must be one of: `open-source`, `internal-service`, `application`, `archived`, `prototype`, `auto`
- Empty/missing config file is valid (no default profile)

**File Locations** (searched in order):
1. `./.gh-repo-health-report.yml` (current directory)
2. `./.gh-repo-health-report.json` (current directory)
3. `~/.gh-repo-health-report.yml` (home directory)
4. `~/.gh-repo-health-report.json` (home directory)

**Loading Precedence**:
- First file found is used
- Invalid YAML/JSON → error (fail fast)
- Missing file → no default profile (legacy behavior)

---

### 4. Result (Extended)

Existing `checks.Result` struct extended to include skipped checks.

**New Fields**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `SkippedChecks` | `[]SkippedCheck` | No | List of checks skipped due to profile enforcement level |

**SkippedCheck Struct**:
```go
type SkippedCheck struct {
    Name   string // Check name constant (e.g., "missing-codeowners")
    Reason string // Human-readable reason (e.g., "Ignored by profile: open-source")
}
```

**Example**:
```go
Result{
    Repository:   repo,
    FailedChecks: []string{"missing-readme", "missing-license"},
    SkippedChecks: []SkippedCheck{
        {Name: "missing-codeowners", Reason: "Ignored by profile: open-source"},
        {Name: "has-projects", Reason: "Ignored by profile: open-source"},
        {Name: "has-wiki", Reason: "Ignored by profile: open-source"},
    },
}
```

**State Transitions**: None (read-only result after evaluation)

---

### 5. Options (Extended)

Existing `checks.Options` struct extended to include profile.

**New Fields**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Profile` | `*Profile` | No | Profile to apply during evaluation; nil = legacy behavior (all checks required) |

**Example**:
```go
opts := checks.Options{
    Since:       180 * 24 * time.Hour,
    MaxBranches: 50,
    MaxTags:     100,
    Profile:     &checks.ProfileOpenSource, // New field
}
```

**Validation Rules**:
- Profile pointer can be nil (backward compatibility)
- If non-nil, Profile must be a valid predefined profile

---

## Predefined Profiles

### Profile: open-source

**Name**: `open-source`  
**Description**: "Public libraries focused on community collaboration and transparency"

**Check Enforcement**:
- **Required** (9 checks): `missing-readme`, `missing-license`, `missing-code-of-conduct`, `missing-contributing`, `missing-security`, `has-description`, `has-issues`, `no-secret-scanning`, `no-vulnerability-alerts`
- **Recommended** (6 checks): `missing-issue-templates`, `missing-pr-template`, `missing-ci`, `has-homepage`, `stale`, `too-many-branches`, `has-stale-branches`
- **Ignored** (13 checks): `missing-codeowners`, `missing-dependabot`, `has-projects`, `has-wiki`, `no-branch-protection`, `no-rulesets`, `no-push-protection`, `no-delete-branch-on-merge`, `too-many-tags`

---

### Profile: internal-service

**Name**: `internal-service`  
**Description**: "Production services and APIs maintained by internal teams"

**Check Enforcement**:
- **Required** (10 checks): `missing-readme`, `missing-codeowners`, `missing-ci`, `no-branch-protection`, `no-rulesets`, `no-vulnerability-alerts`, `no-secret-scanning`, `no-push-protection`, `no-delete-branch-on-merge`, `has-description`
- **Recommended** (7 checks): `missing-license`, `missing-security`, `missing-dependabot`, `stale`, `too-many-branches`, `has-stale-branches`, `missing-issue-templates`
- **Ignored** (11 checks): `missing-code-of-conduct`, `missing-contributing`, `missing-pr-template`, `has-homepage`, `has-issues`, `has-projects`, `has-wiki`, `too-many-tags`

---

### Profile: application

**Name**: `application`  
**Description**: "End-user applications (web apps, mobile apps, desktop software)"

**Check Enforcement**:
- **Required** (6 checks): `missing-readme`, `missing-license`, `missing-ci`, `no-vulnerability-alerts`, `no-secret-scanning`, `has-description`
- **Recommended** (9 checks): `missing-security`, `missing-dependabot`, `missing-codeowners`, `no-branch-protection`, `no-rulesets`, `stale`, `too-many-branches`, `has-stale-branches`, `has-homepage`
- **Ignored** (13 checks): `missing-code-of-conduct`, `missing-contributing`, `missing-issue-templates`, `missing-pr-template`, `has-issues`, `has-projects`, `has-wiki`, `no-push-protection`, `no-delete-branch-on-merge`, `too-many-tags`

---

### Profile: archived

**Name**: `archived`  
**Description**: "Repositories no longer under active development but retained for reference"

**Check Enforcement**:
- **Required** (2 checks): `missing-readme`, `missing-license`
- **Recommended** (2 checks): `has-description`, `has-homepage`
- **Ignored** (24 checks): All other checks (staleness, CI, governance, security)

---

### Profile: prototype

**Name**: `prototype`  
**Description**: "Experimental or proof-of-concept repositories for exploration and learning"

**Check Enforcement**:
- **Required** (2 checks): `missing-readme`, `has-description`
- **Recommended** (2 checks): `missing-license`, `has-homepage`
- **Ignored** (24 checks): All other checks (governance, security, CI, staleness)

---

## Auto-Detection Logic

When `--profile auto` is specified, apply these rules in priority order:

### Priority 1: Archived Status
**Condition**: `repo.IsArchived == true`  
**Result**: Use `archived` profile

### Priority 2: Topic Matching
**Conditions**:
- Topics include `prototype`, `experimental`, `poc`, `spike` → `prototype` profile
- Topics include `library`, `package`, `npm-package`, `gem`, `pypi` → `open-source` profile
- Topics include `service`, `api`, `microservice` → `internal-service` profile
- Topics include `app`, `webapp`, `mobile-app`, `desktop` → `application` profile

**Conflict Resolution**: First matching topic wins (order above)

### Priority 3: Visibility
**Condition**: `repo.IsPrivate == false`  
**Result**: Use `open-source` profile

### Priority 4: Fallback
**Condition**: No above conditions matched  
**Result**: Use `internal-service` profile (or config default if set)

---

## Evaluation Algorithm

```
function Evaluate(repo, opts):
    result = new Result(repo)
    profile = opts.Profile
    
    // Backward compatibility: no profile = all checks required
    if profile == nil:
        return evaluateLegacy(repo, opts)
    
    // For each of the 28 checks:
    for each check in AllChecks:
        enforcement = profile.Checks[check]
        
        if enforcement == EnforcementIgnored:
            result.SkippedChecks.append({
                Name: check,
                Reason: "Ignored by profile: " + profile.Name
            })
            continue
        
        // Evaluate check (existing logic)
        passed = evaluateCheck(check, repo, opts)
        
        if !passed:
            if enforcement == EnforcementRequired || enforcement == EnforcementRecommended:
                result.FailedChecks.append(check)
    
    return result
```

**Key Points**:
- `nil` profile → legacy behavior (backward compatible)
- Ignored checks are skipped before evaluation (no API calls or computation wasted)
- Both Required and Recommended checks are evaluated and included in FailedChecks
- Distinction between Required/Recommended matters for scoring/output, not evaluation

---

## Scoring Algorithm

```
function CalculateScore(result):
    profile = result.Profile
    
    if profile == nil:
        // Legacy scoring: all checks equal weight
        total = 28
        passed = total - len(result.FailedChecks)
        return (passed / total) * 100
    
    // Profile-aware scoring: count only non-ignored checks
    total = 0
    for each check in AllChecks:
        if profile.Checks[check] != EnforcementIgnored:
            total++
    
    passed = total - len(result.FailedChecks)
    return (passed / total) * 100
```

**Example**:
- Profile: `open-source` (15 required + recommended, 13 ignored)
- Failed checks: 3
- Score: (15 - 3) / 15 = 80%

---

## CLI Flag Resolution Order

**Profile Selection Precedence** (highest to lowest):
1. `--profile` flag value (explicit user choice)
2. `default_profile` from config file (organizational default)
3. No profile / legacy behavior (all checks required)

**Auto-Detection Trigger**:
- `--profile auto` flag → run auto-detection for each repository
- Config file `default_profile: auto` → run auto-detection for all repositories

**Explicit Override**:
- `--profile open-source --fail-on missing-codeowners` → Even though `missing-codeowners` is ignored by the profile, the explicit `--fail-on` flag overrides this for exit code logic

---

## Relationships

```
Config (file) ─┐
               ├─> Profile (selected)
--profile flag─┘         │
                         ├─> Evaluate() ─> Result
Repository (metadata)────┘         │
                                   └─> SkippedChecks, FailedChecks
```

**Cardinality**:
- 1 Config → 0-1 default Profile (config may not exist)
- 1 Repository → 1 Result (after evaluation)
- 1 Result → 0-N SkippedChecks (none if no profile or all checks applicable)
- 1 Result → 0-N FailedChecks (none if all checks pass)

---

## Validation Rules Summary

| Entity | Validation | Error Handling |
|--------|------------|----------------|
| **Profile** | All 28 checks defined, no duplicates, no unknown checks | Compile-time validation (unit tests) |
| **Config** | Valid YAML/JSON syntax, `default_profile` matches predefined profile | Runtime error on invalid config file |
| **EnforcementLevel** | Must be 0, 1, or 2 | Type safety via Go enum (iota) |
| **Options.Profile** | Nil or valid profile pointer | Runtime check in Evaluate() |
| **--profile flag** | Must match predefined profile name or "auto" | CLI validation before execution |

---

## Migration Path

**Existing Users**:
- No changes required
- Existing invocations work identically (no profile = all checks)
- Existing tests pass without modification

**Adopting Profiles**:
1. Run tool once to understand current failures
2. Choose appropriate profile for repository type
3. Add `--profile` flag to invocations or create config file
4. Adjust CI pipelines to use profile-aware `--fail-on`

**Gradual Rollout**:
- Start with `--profile` flag for experimentation
- Move to config file once settled on defaults
- Optionally adopt `auto` for diverse repository portfolios
