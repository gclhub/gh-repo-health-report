# CLI Contract: Profile Support Flags

**Feature**: 007-policy-profile-support  
**Contract Type**: Command-line interface flags and behavior

## New CLI Flags

### --profile

**Type**: `string`  
**Default**: "" (empty, no profile)  
**Valid Values**: `open-source`, `internal-service`, `application`, `archived`, `prototype`, `auto`

**Description**: Specifies which policy profile to apply when evaluating health checks. Each profile defines which checks are required, recommended, or ignored based on repository type.

**Usage Examples**:
```bash
# Evaluate a single repository with open-source profile
gh repo-health-report --repo owner/library --profile open-source

# Audit an organization with internal-service profile
gh repo-health-report --org mycompany --profile internal-service

# Use auto-detection to infer profile from repository metadata
gh repo-health-report --org mycompany --profile auto

# No profile specified (legacy behavior: all 28 checks evaluated)
gh repo-health-report --repo owner/repo
```

**Error Handling**:
- Invalid profile name → exit with error: `Error: unknown profile "xyz". Valid profiles: open-source, internal-service, application, archived, prototype, auto`
- Conflicting flags: None (compatible with all existing flags)

**Interaction with Other Flags**:
- **--fail-on**: Profiles filter which checks contribute to exit code (see below)
- **--format**: All output formats include skipped check indicators when profile is active
- **--include-archived**: Works normally; archived repos can still use any profile (profile != repository archived status)

---

### --profile-config

**Type**: `string`  
**Default**: "" (auto-discover config file)  
**Valid Values**: Path to YAML or JSON config file

**Description**: Explicitly specifies the config file path to load default profile settings from. If not provided, the tool searches for config files in standard locations (current directory, home directory).

**Usage Examples**:
```bash
# Use config file from custom path
gh repo-health-report --org mycompany --profile-config /path/to/config.yml

# Let tool auto-discover config file (searches ./.gh-repo-health-report.yml, ~/.gh-repo-health-report.yml)
gh repo-health-report --org mycompany
```

**Config File Format (YAML)**:
```yaml
default_profile: internal-service
```

**Config File Format (JSON)**:
```json
{
  "default_profile": "internal-service"
}
```

**Error Handling**:
- File not found → exit with error: `Error: config file not found: /path/to/config.yml`
- Invalid YAML/JSON → exit with error: `Error: failed to parse config file: ...`
- Invalid profile name in config → exit with error: `Error: unknown profile "xyz" in config file. Valid profiles: ...`

**Precedence**:
1. `--profile` flag (if specified) → overrides config file
2. Config file `default_profile` (if exists) → used when no `--profile` flag
3. No profile → legacy behavior (all checks)

---

## Modified Behavior: --fail-on with Profiles

**Existing Flag**: `--fail-on [check-names]` (comma-separated, or "any")

**New Behavior with Profiles**:
- `--fail-on any` → exit 1 if ANY non-ignored check fails
- `--fail-on [specific-check]` → exit 1 if that specific check fails, EVEN IF profile ignores it (explicit user override)

**Examples**:

```bash
# Exit 1 if any required/recommended check fails (ignored checks don't count)
gh repo-health-report --repo owner/repo --profile open-source --fail-on any

# Exit 1 if missing-codeowners fails, even though open-source profile ignores it
# (explicit flag overrides profile)
gh repo-health-report --repo owner/repo --profile open-source --fail-on missing-codeowners

# Backward compatibility: no profile = all checks contribute to failure
gh repo-health-report --repo owner/repo --fail-on any
```

**Rationale**: The word "any" is context-aware (respects profile's check filtering), but explicit check names represent immediate user intent (override profile).

---

## Output Format Changes

All output formats include indicators for skipped checks when a profile is active.

### Table Format

**Before Profile Feature** (existing):
```
REPO              STALE  DESCRIPTION  README  LICENSE  CODEOWNERS  ...
owner/library     NO     ✓            ✓       ✓        ✗           ...
```

**After Profile Feature** (with profile):
```
REPO              STALE  DESCRIPTION  README  LICENSE  CODEOWNERS  ...
owner/library     NO     ✓            ✓       ✓        [SKIP]      ...

Note: Checks marked [SKIP] are ignored by the active profile (open-source)
```

**Alternative** (blank cells instead of [SKIP]):
```
REPO              STALE  DESCRIPTION  README  LICENSE  CODEOWNERS  ...
owner/library     NO     ✓            ✓       ✓                    ...
```

---

### JSON Format

**Before Profile Feature** (existing):
```json
{
  "repository": "owner/library",
  "stale": false,
  "has_description": true,
  "has_readme": true,
  "has_license": true,
  "has_codeowners": false,
  "failed_checks": ["missing-codeowners"]
}
```

**After Profile Feature** (with profile):
```json
{
  "repository": "owner/library",
  "profile": "open-source",
  "stale": false,
  "has_description": true,
  "has_readme": true,
  "has_license": true,
  "has_codeowners": false,
  "failed_checks": ["missing-readme"],
  "skipped_checks": [
    {
      "check": "missing-codeowners",
      "reason": "Ignored by profile: open-source"
    },
    {
      "check": "has-projects",
      "reason": "Ignored by profile: open-source"
    }
  ]
}
```

**New Fields**:
- `profile` (string): Name of active profile, or omitted if no profile
- `skipped_checks` (array): List of checks skipped due to profile, each with `check` (name) and `reason` (explanation)

---

### CSV Format

**Before Profile Feature** (existing):
```csv
repository,stale,description,readme,license,codeowners,failed_checks
owner/library,NO,YES,YES,YES,NO,"missing-codeowners"
```

**After Profile Feature** (with profile):
```csv
repository,profile,stale,description,readme,license,codeowners,failed_checks,skipped_checks
owner/library,open-source,NO,YES,YES,YES,NO,"missing-readme","missing-codeowners,has-projects,has-wiki"
```

**New Columns**:
- `profile`: Name of active profile, or empty if no profile
- `skipped_checks`: Comma-separated list of skipped check names

---

### Markdown Format

**Before Profile Feature** (existing):
```markdown
| Repository    | Stale | Description | README | LICENSE | CODEOWNERS |
|---------------|-------|-------------|--------|---------|------------|
| owner/library | NO    | ✓           | ✓      | ✓       | ✗          |
```

**After Profile Feature** (with profile):
```markdown
| Repository    | Stale | Description | README | LICENSE | CODEOWNERS |
|---------------|-------|-------------|--------|---------|------------|
| owner/library | NO    | ✓           | ✓      | ✓       | [SKIP]     |

**Profile**: open-source  
**Skipped Checks**: missing-codeowners, has-projects, has-wiki (ignored by profile)
```

---

## Exit Code Behavior

### Without Profile (Backward Compatible)

```bash
# Exit 0: All checks pass
gh repo-health-report --repo owner/perfect-repo

# Exit 1: Any check fails when --fail-on any specified
gh repo-health-report --repo owner/repo --fail-on any

# Exit 1: Specific check fails
gh repo-health-report --repo owner/repo --fail-on missing-readme
```

### With Profile

```bash
# Exit 0: Ignored check fails, but --fail-on any respects profile
gh repo-health-report --repo owner/repo --profile open-source --fail-on any
# (missing-codeowners fails, but it's ignored by open-source profile)

# Exit 1: Required check fails
gh repo-health-report --repo owner/repo --profile open-source --fail-on any
# (missing-readme fails, which is required by open-source profile)

# Exit 1: Explicit check override (user wants this check even if profile ignores it)
gh repo-health-report --repo owner/repo --profile open-source --fail-on missing-codeowners
```

---

## Help Text

### Updated --help Output

```
gh repo-health-report --help

Report on the health of GitHub repositories

Usage:
  gh-repo-health-report [flags]

Flags:
      --org string              Organization to audit
      --owner string            User to audit
      --repo strings            Specific repo(s) in owner/name format (may be repeated)
      --include-forks           Include forked repos
      --include-archived        Include archived repos
      --since string            Staleness threshold (e.g. 180d, 6m, 1y, 2024-01-01) (default "180d")
      --format string           Output format: table, json, csv, md (default "table")
      --output string           Output file path (default stdout)
      --fail-on string          Comma-separated check names; exit 1 if any repo fails (use 'any' to fail on any failure)
      --max-branches int        Branch count threshold for too-many-branches check (0 to disable) (default 50)
      --max-tags int            Tag count threshold for too-many-tags check (0 to disable) (default 100)
      --profile string          Policy profile to apply: open-source, internal-service, application, archived, prototype, auto
      --profile-config string   Path to config file for default profile (default: auto-discover)
  -h, --help                    help for gh-repo-health-report

Profiles:
  open-source       Public libraries focused on community collaboration
  internal-service  Production services and APIs maintained by internal teams
  application       End-user applications (web apps, mobile apps, desktop software)
  archived          Repositories no longer under active development
  prototype         Experimental or proof-of-concept repositories
  auto              Automatically detect profile based on repository metadata

Config File:
  Create .gh-repo-health-report.yml (or .json) in current directory or home directory:
    default_profile: internal-service

Examples:
  # Evaluate open-source library
  gh repo-health-report --repo owner/library --profile open-source

  # Audit organization with internal-service profile
  gh repo-health-report --org mycompany --profile internal-service

  # Use auto-detection
  gh repo-health-report --org mycompany --profile auto

  # Fail CI if required checks fail
  gh repo-health-report --repo owner/repo --profile internal-service --fail-on any
```

---

## Backward Compatibility Guarantees

**No Breaking Changes**:
- All existing flags work identically when `--profile` is not specified
- Existing invocations produce identical output (no `[SKIP]` markers when no profile)
- Existing `--fail-on` logic unchanged when no profile is active
- All existing tests pass without modification

**Opt-In Feature**:
- Profiles are opt-in via explicit `--profile` flag or config file
- Users who don't adopt profiles see zero behavior change
- No performance impact when profiles are not used (same code paths as before)

---

## Error Messages

### Invalid Profile Name
```
Error: unknown profile "my-custom-profile"
Valid profiles: open-source, internal-service, application, archived, prototype, auto
```

### Config File Not Found (when --profile-config specified)
```
Error: config file not found: /path/to/config.yml
```

### Invalid Config File Syntax
```
Error: failed to parse config file: yaml: line 2: mapping values are not allowed in this context
```

### Invalid Profile in Config File
```
Error: unknown profile "custom" in config file .gh-repo-health-report.yml
Valid profiles: open-source, internal-service, application, archived, prototype, auto
```

---

## Future Extensibility

**Not in Scope for Initial Release** (but design allows):
- Custom profile definitions in config file (beyond the five predefined)
- Per-repository profile overrides (e.g., `.github/health-profile.yml` in repo)
- Profile inheritance (e.g., `my-service` extends `internal-service`)
- Org-level profile API (fetch default profile from GitHub org settings)

**Current Design Compatibility**:
- Profile data structure supports custom profiles (just add more `Profile` variables)
- Config loading can be extended to support `profiles:` block with custom definitions
- CLI contract allows future flags like `--profile-file` for custom profile definitions
