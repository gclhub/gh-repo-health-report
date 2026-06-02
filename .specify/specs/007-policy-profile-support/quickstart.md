# Quickstart: Policy Profile Support

**Feature**: 007-policy-profile-support
**Audience**: Developers implementing this feature

## Overview

This feature adds policy profile support to gh-repo-health-report, enabling users to apply different health check expectations based on repository type. Implementation spans three internal packages plus CLI integration.

---

## Implementation Checklist

### Phase 1: Core Profile System (internal/checks)

**File**: `internal/checks/profile.go`

- [ ] Define `EnforcementLevel` enum (Required, Recommended, Ignored)
- [ ] Define `Profile` struct (Name, Description, Checks map)
- [ ] Implement `GetProfile(name string) *Profile` function
- [ ] Define five predefined profiles as package variables:
  - [ ] `ProfileOpenSource`
  - [ ] `ProfileInternalService`
  - [ ] `ProfileApplication`
  - [ ] `ProfileArchived`
  - [ ] `ProfilePrototype`
- [ ] Implement auto-detection logic: `DetectProfile(repo *api.Repository) *Profile`

**File**: `internal/checks/profile_test.go`

- [ ] Test each predefined profile has all 28 checks defined
- [ ] Test `GetProfile()` returns correct profile for valid names
- [ ] Test `GetProfile()` returns nil for invalid names
- [ ] Test auto-detection logic with archived repositories
- [ ] Test auto-detection logic with topic matching
- [ ] Test auto-detection fallback logic

---

### Phase 2: Config File Support (internal/checks)

**File**: `internal/checks/config.go`

- [ ] Define `Config` struct (DefaultProfile field)
- [ ] Implement `LoadConfig(path string) (*Config, error)` for explicit path
- [ ] Implement `DiscoverConfig() (*Config, error)` for auto-discovery:
  - [ ] Search `./.gh-repo-health-report.yml`
  - [ ] Search `./.gh-repo-health-report.json`
  - [ ] Search `~/.gh-repo-health-report.yml`
  - [ ] Search `~/.gh-repo-health-report.json`
- [ ] Support YAML parsing with `gopkg.in/yaml.v3`
- [ ] Support JSON parsing with `encoding/json`

**File**: `internal/checks/config_test.go`

- [ ] Test config loading from YAML file
- [ ] Test config loading from JSON file
- [ ] Test config discovery order (current dir before home dir)
- [ ] Test error handling for invalid YAML/JSON
- [ ] Test error handling for invalid profile names in config
- [ ] Test missing config file returns nil without error

---

### Phase 3: Profile-Aware Evaluation (internal/checks)

**File**: `internal/checks/checks.go`

- [ ] Add `Profile *Profile` field to `Options` struct
- [ ] Add `SkippedChecks []SkippedCheck` field to `Result` struct
- [ ] Define `SkippedCheck` struct (Name, Reason)
- [ ] Update `Evaluate()` function:
  - [ ] Handle `nil` profile (backward compatibility path)
  - [ ] For each check, check enforcement level from profile
  - [ ] Skip checks with `EnforcementIgnored`, add to `SkippedChecks`
  - [ ] Evaluate checks with `EnforcementRequired` or `EnforcementRecommended`
  - [ ] Add failed checks to `FailedChecks` as before

**File**: `internal/checks/checks_test.go`

- [ ] Test `Evaluate()` with nil profile (legacy behavior)
- [ ] Test `Evaluate()` with open-source profile
- [ ] Test skipped checks are recorded in `SkippedChecks`
- [ ] Test failed checks exclude ignored checks
- [ ] Test required checks that fail are in `FailedChecks`
- [ ] Test recommended checks that fail are in `FailedChecks`

---

### Phase 4: Output Format Changes (internal/formatter)

**File**: `internal/formatter/formatter.go`

- [ ] Update `formatTable()`:
  - [ ] Mark skipped checks with `[SKIP]` or blank cells
  - [ ] Add footer legend when profile is active
- [ ] Update `formatJSON()`:
  - [ ] Add `"profile"` field with profile name
  - [ ] Add `"skipped_checks"` array with check objects (name, reason)
- [ ] Update `formatCSV()`:
  - [ ] Add `profile` column
  - [ ] Add `skipped_checks` column (comma-separated names)
- [ ] Update `formatMD()`:
  - [ ] Mark skipped checks with `[SKIP]` in table
  - [ ] Add profile name and skipped checks summary below table

**File**: `internal/formatter/formatter_test.go`

- [ ] Test table format with profile (includes `[SKIP]` markers)
- [ ] Test table format without profile (no `[SKIP]` markers)
- [ ] Test JSON format includes `skipped_checks` field when profile active
- [ ] Test CSV format includes `skipped_checks` column
- [ ] Test markdown format includes profile summary

---

### Phase 5: CLI Integration (cmd/gh-repo-health-report)

**File**: `cmd/gh-repo-health-report/main.go`

- [ ] Add `--profile` string flag
- [ ] Add `--profile-config` string flag
- [ ] Implement profile resolution logic:
  - [ ] If `--profile` specified, use `GetProfile()` or `DetectProfile()` for "auto"
  - [ ] Else if `--profile-config` specified, use `LoadConfig()`
  - [ ] Else use `DiscoverConfig()`
  - [ ] Else no profile (nil)
- [ ] Pass profile to `checks.Options` in `Evaluate()` calls
- [ ] Update `--fail-on` logic:
  - [ ] If `--fail-on any`, only fail on non-ignored checks
  - [ ] If `--fail-on <check>`, fail even if profile ignores check (explicit override)
- [ ] Update help text with profile documentation

---

### Phase 6: Documentation

**File**: `README.md`

- [ ] Add "Policy Profiles" section explaining feature
- [ ] Document five predefined profiles with check lists
- [ ] Document `--profile` and `--profile-config` flags
- [ ] Provide config file examples (YAML and JSON)
- [ ] Add usage examples for each profile
- [ ] Document auto-detection behavior
- [ ] Explain `--fail-on` interaction with profiles

---

### Phase 7: Testing & Validation

- [ ] Run `go test ./...` — all tests pass
- [ ] Run `go build` — builds successfully
- [ ] Manual test: Run without `--profile` flag (verify backward compatibility)
- [ ] Manual test: Run with `--profile open-source` (verify skipped checks marked)
- [ ] Manual test: Create config file with `default_profile`, verify it's used
- [ ] Manual test: Use `--profile auto` on diverse repositories (verify detection)
- [ ] Manual test: Test `--fail-on any` with profile (verify ignored checks don't exit 1)
- [ ] Manual test: Test `--fail-on <check>` with profile ignoring that check (verify explicit override)
- [ ] Manual test: All output formats (table, JSON, CSV, markdown) show skipped checks correctly

---

## Quick Reference: Key Functions

### Profile Selection
```go
// Get predefined profile by name
profile := checks.GetProfile("open-source") // returns *Profile or nil

// Auto-detect profile from repository
profile := checks.DetectProfile(repo) // returns *Profile
```

### Config Loading
```go
// Load config from explicit path
config, err := checks.LoadConfig("/path/to/config.yml")

// Auto-discover config from standard locations
config, err := checks.DiscoverConfig()
```

### Evaluation with Profile
```go
opts := checks.Options{
    Since:       180 * 24 * time.Hour,
    MaxBranches: 50,
    MaxTags:     100,
    Profile:     profile, // can be nil for legacy behavior
}
result := checks.Evaluate(repo, opts)
```

### Checking Enforcement Level
```go
if profile != nil {
    enforcement := profile.Checks[checks.CheckMissingCodeowners]
    if enforcement == checks.EnforcementIgnored {
        // Skip this check
    }
}
```

---

## Development Tips

### Testing Profiles
- Use table-driven tests with multiple profiles and repositories
- Test edge cases: nil profile, empty profile, profile with only ignored checks
- Validate all 28 checks have explicit enforcement levels in each profile

### Auto-Detection
- Test priority order: archived status > topics > visibility > fallback
- Test topic conflict resolution (first matching topic wins)
- Test fallback behavior when no heuristics match

### Backward Compatibility
- Always test with `Profile: nil` in Options to verify legacy behavior
- Ensure existing tests pass without modification
- Verify output formats show no profile indicators when profile is nil

### Config Files
- Test both YAML and JSON formats
- Test file discovery order (current dir before home dir)
- Test invalid syntax handling (should fail fast with clear error)

### Output Formats
- Verify `[SKIP]` markers only appear when profile is active
- Ensure JSON structure is backward compatible (new fields are additive)
- Test CSV with spreadsheet tools to verify format correctness

---

## Common Pitfalls

❌ **Don't**: Modify existing check evaluation logic
✅ **Do**: Add profile filtering before/after existing evaluation

❌ **Don't**: Change behavior when profile is nil
✅ **Do**: Preserve exact legacy behavior for backward compatibility

❌ **Don't**: Make profiles required
✅ **Do**: Make profiles optional with sensible defaults

❌ **Don't**: Hard-code check names in multiple places
✅ **Do**: Use existing check constants from `checks.go`

❌ **Don't**: Skip validation of profile definitions
✅ **Do**: Add unit tests that verify all 28 checks are defined in each profile

❌ **Don't**: Ignore config file parsing errors
✅ **Do**: Fail fast with clear error messages for invalid configs

---

## Dependencies

### New Dependency
```bash
go get gopkg.in/yaml.v3
```

**Justification**: Standard library has no YAML support; `yaml.v3` is the de facto standard for YAML in Go.

---

## Estimated Implementation Time

| Phase | Time Estimate | Complexity |
|-------|---------------|------------|
| Profile system (profile.go) | 2-3 hours | Medium (data structures, enum, predefined profiles) |
| Config loading (config.go) | 1-2 hours | Low (file I/O, YAML/JSON parsing) |
| Profile-aware evaluation | 2-3 hours | Medium (update Evaluate(), add filtering) |
| Output format changes | 2-3 hours | Medium (update 4 formatters, maintain backward compat) |
| CLI integration | 1-2 hours | Low (add flags, wire up profile resolution) |
| Testing | 3-4 hours | High (comprehensive test coverage for all components) |
| Documentation | 1-2 hours | Low (README updates, examples) |
| **Total** | **12-19 hours** | **Medium** |

---

## Success Criteria

✅ **Feature Complete When**:
- All 5 predefined profiles defined with correct enforcement levels
- Config file loading works for YAML and JSON
- Auto-detection correctly infers profile from repository metadata
- `Evaluate()` skips checks marked as ignored by profile
- All 4 output formats show skipped check indicators when profile is active
- `--fail-on` respects profile filtering for "any", but overrides for specific checks
- Backward compatibility: no profile = identical behavior to current version
- All tests pass (`go test ./...`)
- README documents the feature with examples

✅ **Validation Tests Pass**:
- Run tool on 100+ repositories with different profiles
- Verify no unexpected failures or crashes
- Verify skipped checks match profile definitions
- Verify exit codes work correctly with `--fail-on`
- Verify output formats are correct in all modes

---

## Next Steps After Implementation

1. **Merge to main**: After code review and CI passes
2. **Release**: Tag new version with profile support
3. **Documentation**: Update GitHub release notes with examples
4. **Adoption**: Share with early adopters for feedback
5. **Iteration**: Gather feedback, tune profile definitions if needed
6. **Future**: Consider custom profile support in later releases
