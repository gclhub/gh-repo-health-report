# Research: Policy Profile Support

**Feature**: 007-policy-profile-support
**Date**: 2026-06-01
**Phase**: 0 - Research & Discovery

## Research Tasks

### 1. Profile Data Structure Design

**Question**: How should profiles be represented in Go to efficiently map checks to enforcement levels?

**Decision**: Use a `map[string]EnforcementLevel` for check-to-enforcement mapping within a `Profile` struct.

**Rationale**:
- O(1) lookup for check enforcement level during evaluation
- Flexible: easy to add custom profiles or modify existing ones
- Type-safe: enum for enforcement levels prevents invalid values
- Compact: only stores non-default enforcement levels if using a baseline default

**Structure**:
```go
type EnforcementLevel int

const (
    EnforcementRequired EnforcementLevel = iota
    EnforcementRecommended
    EnforcementIgnored
)

type Profile struct {
    Name        string
    Description string
    Checks      map[string]EnforcementLevel // check name -> enforcement level
}
```

**Alternatives Considered**:
- Slice of check names per enforcement level: Requires 3 slices per profile, slower lookup (O(n) contains check)
- Separate boolean fields per check: Not scalable to 28 checks, verbose
- String-based enforcement: Less type-safe, prone to typos

---

### 2. Config File Format and Loading

**Question**: What config file format and loading strategy provides the best user experience?

**Decision**: Support YAML (primary) with JSON fallback; use `gopkg.in/yaml.v3` library.

**Rationale**:
- YAML is human-friendly for config files (comments, less syntax noise)
- JSON support provides programmatic alternative
- `gopkg.in/yaml.v3` is stable, widely used in Go ecosystem, maintained
- Search order: `./.gh-repo-health-report.yml` → `~/.gh-repo-health-report.yml` (allows project-level overrides)

**Config Structure**:
```yaml
default_profile: internal-service
```

**Alternatives Considered**:
- TOML: Less common in Go CLI tools, requires additional dependency
- JSON only: More verbose, no comment support, less user-friendly
- Environment variable: Harder to version control, less discoverable
- CLI config subcommands: Over-engineered for a single setting

---

### 3. Auto-Detection Heuristics

**Question**: What repository metadata is most reliable for auto-detecting profile type?

**Decision**: Prioritize explicit GitHub metadata (archived status, visibility) over inferred metadata (topics).

**Heuristics Priority**:
1. **Archived status** (definitive) → `archived` profile
2. **Topics** (explicit tags) → match keywords to profiles
3. **Visibility** (public) → default to `open-source`
4. **Fallback** (private) → use `internal-service` or config default

**Rationale**:
- Archived status is an explicit GitHub setting (high signal)
- Topics are user-maintained (medium signal, but opt-in indicates intent)
- Visibility is a reasonable proxy for open-source vs internal (low signal but widely available)
- Fallback ensures every repo gets a profile

**Topic Mapping**:
- `prototype`, `experimental`, `poc`, `spike` → `prototype`
- `library`, `package`, `npm-package`, `gem`, `pypi` → `open-source`
- `service`, `api`, `microservice` → `internal-service`
- `app`, `webapp`, `mobile-app`, `desktop` → `application`

**Alternatives Considered**:
- Repository name patterns (e.g., `-service` suffix): Too fragile, organization-specific
- Commit frequency: Expensive to compute, overlaps with staleness check
- Language detection: Not indicative of governance needs (Go services vs Go libraries both need governance)
- Organization-level defaults: Requires GitHub API calls to org settings (rate limit impact)

---

### 4. Profile-Aware Scoring Algorithm

**Question**: How should scoring change when checks are ignored by a profile?

**Decision**: Exclude ignored checks from pass/fail counts and percentages; only required/recommended checks contribute to scoring.

**Algorithm**:
```
total_checks = count(required_checks) + count(recommended_checks)
passed_checks = count(required_checks_passed) + count(recommended_checks_passed)
score_percentage = (passed_checks / total_checks) * 100
```

**Rationale**:
- Scoring reflects only applicable checks for the repository type
- Matches user expectation: ignored checks shouldn't penalize score
- Maintains backward compatibility: with no profile, all checks are "required" (legacy behavior)

**Alternatives Considered**:
- Include ignored checks in denominator as "auto-pass": Inflates scores artificially, misleading
- Separate scores for required vs recommended: Over-complicates output, users want one number
- Weight required checks higher than recommended: Adds complexity without clear benefit for initial version

---

### 5. Backward Compatibility Strategy

**Question**: How do we ensure existing users see no behavior change when profiles are not used?

**Decision**: Default behavior (no profile specified, no config file) evaluates all 28 checks as "required" with existing logic.

**Compatibility Guarantees**:
- No `--profile` flag + no config file → all checks evaluated (legacy `Evaluate()` path)
- Existing `--fail-on` behavior unchanged when no profile specified
- Output formats unchanged when no profile specified (no `[IGNORED]` markers appear)
- All existing tests pass without modification

**Implementation**:
- `Evaluate()` accepts optional `*Profile` parameter
- `nil` profile → legacy behavior (treat all checks as required)
- Non-nil profile → filter checks by enforcement level

**Alternatives Considered**:
- Make profile required, provide "all-checks" profile: Breaking change, requires flag on every invocation
- Enable profiles by default with auto-detection: Too aggressive, changes behavior for existing users
- Separate `EvaluateWithProfile()` function: Code duplication, maintenance burden

---

### 6. Skipped Check Indicators in Output

**Question**: How should skipped/ignored checks be communicated in different output formats?

**Decision**: Add visual indicators appropriate to each format, with explanatory legends.

**Format-Specific Approach**:

**Table Format**:
- Mark skipped checks with `[SKIP]` or blank cell (less visual noise)
- Add footer legend: `"Checks marked [SKIP] are ignored by the active profile"`

**JSON Format**:
```json
{
  "repository": "owner/repo",
  "failed_checks": ["missing-readme"],
  "skipped_checks": [
    {"check": "missing-codeowners", "reason": "Ignored by profile: open-source"}
  ],
  "score": "85%"
}
```

**CSV Format**:
- Add `skipped_checks` column with comma-separated check names
- Example: `"missing-codeowners,has-projects,has-wiki"`

**Markdown Format**:
- Similar to table, use `[SKIP]` notation
- Add legend section explaining profile

**Rationale**:
- Each format uses idioms familiar to that format's consumers
- JSON includes structured data (check name + reason) for programmatic use
- Table/Markdown optimize for readability
- CSV optimizes for spreadsheet import

**Alternatives Considered**:
- Omit skipped checks entirely from output: Confusing (users wonder why check is missing)
- List all 28 checks with pass/fail/skip status: Too verbose, clutters output
- Separate output section for skipped checks: Requires scrolling, harder to scan

---

### 7. Integration with `--fail-on` Flag

**Question**: How should `--fail-on` interact with profiles when a check is ignored?

**Decision**: Ignored checks never trigger exit code 1, unless explicitly specified in `--fail-on` (user intent override).

**Behavior**:
- `--fail-on any` + profile ignores check X → check X failure does NOT exit 1
- `--fail-on missing-codeowners` + profile ignores `missing-codeowners` → check failure DOES exit 1 (explicit override)
- No profile + `--fail-on any` → all check failures exit 1 (backward compatibility)

**Rationale**:
- Profile defines policy defaults, but explicit CLI flags represent immediate user intent
- `any` respects profile (users expect profile to filter checks)
- Specific check names override profile (users explicitly care about that check in this run)

**Alternatives Considered**:
- Profile always overrides `--fail-on`: Violates principle of least surprise (explicit flag should win)
- Profile never affects `--fail-on`: Defeats purpose of profiles (can't set org-level policy)
- Add `--fail-on-required` flag: More complexity, existing `--fail-on` is sufficient

---

### 8. Profile Definition Maintenance

**Question**: How are the five predefined profiles defined and maintained in code?

**Decision**: Define profiles as package-level variables in `internal/checks/profile.go` with clear documentation.

**Structure**:
```go
var (
    ProfileOpenSource = Profile{
        Name:        "open-source",
        Description: "Public libraries focused on community collaboration",
        Checks: map[string]EnforcementLevel{
            CheckMissingReadme:       EnforcementRequired,
            CheckMissingLicense:      EnforcementRequired,
            // ... (all 28 checks defined)
        },
    }

    ProfileInternalService = Profile{ /* ... */ }
    ProfileApplication = Profile{ /* ... */ }
    ProfileArchived = Profile{ /* ... */ }
    ProfilePrototype = Profile{ /* ... */ }
)

// GetProfile returns a predefined profile by name, or nil if not found
func GetProfile(name string) *Profile { /* ... */ }
```

**Rationale**:
- Centralized definitions are easy to review and maintain
- Package-level variables are initialized once, no runtime overhead
- Clear mapping from spec's profile definitions to code
- Easy to add custom profiles in future (just add another variable)

**Alternatives Considered**:
- JSON/YAML files embedded in binary: Requires parsing at runtime, more complexity
- Hardcoded in `Evaluate()`: Scattered logic, harder to test profiles independently
- Database/external config: Over-engineered for five static profiles

---

## Summary of Technical Decisions

| Area | Decision | Rationale |
|------|----------|-----------|
| **Data Structure** | `map[string]EnforcementLevel` in Profile struct | O(1) lookup, type-safe, flexible |
| **Config Format** | YAML primary, JSON fallback, `gopkg.in/yaml.v3` | User-friendly, standard in Go ecosystem |
| **Auto-Detection** | Prioritize archived status > topics > visibility | Explicit metadata more reliable than inferred |
| **Scoring** | Exclude ignored checks from pass/fail counts | Matches user expectation, contextually accurate |
| **Backward Compatibility** | No profile = all checks required (legacy behavior) | Zero impact on existing users |
| **Skipped Check Output** | Format-specific indicators (`[SKIP]`, JSON field, CSV column) | Optimized for each format's idioms |
| **--fail-on Interaction** | Explicit check names override profile | User intent takes precedence |
| **Profile Definitions** | Package-level variables in profile.go | Centralized, easy to maintain, no runtime overhead |

---

## Dependencies

**New Dependency**: `gopkg.in/yaml.v3`
- **Purpose**: Parse YAML config files for default profile setting
- **Justification**: Standard library has no YAML support; `yaml.v3` is stable, maintained, widely used
- **Size**: ~150KB, acceptable for config file parsing
- **Alternatives**: `encoding/json` (standard library, but JSON less user-friendly for config files)

---

## Open Questions

None. All research tasks resolved.
