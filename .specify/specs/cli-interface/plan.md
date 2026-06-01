# CLI Interface Implementation Plan

**Status**: migrated  
**Feature**: Command-line interface for repository health reporting  
**Migration Date**: 2026-06-01

## Technical Context

### Framework Selection: Cobra
The CLI is built using [spf13/cobra](https://github.com/spf13/cobra), a popular Go library for creating powerful CLI applications. Cobra provides:
- Automatic help text generation
- Flag parsing with type validation
- Subcommand support (extensible for future features)
- POSIX-compliant flags
- Integration with Viper for configuration management

**Rationale**: Cobra is the de facto standard for Go CLI tools, used by kubectl, gh, and many other prominent projects. It reduces boilerplate and provides battle-tested flag parsing.

### Project Structure
```
gh-repo-health-report/
├── cmd/
│   └── root.go           # Cobra root command and flag definitions
├── internal/
│   ├── analyzer/         # Core analysis logic
│   ├── config/           # Configuration loading and validation
│   ├── formatter/        # Output format implementations
│   └── reporter/         # Report generation
├── go.mod
├── go.sum
└── main.go              # Entry point
```

### Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `gopkg.in/yaml.v3` - YAML parsing
- GitHub API client - Repository data fetching

## Architecture Overview

### Command Execution Flow
```
1. main.go entry point
   ↓
2. cmd/root.go initializes Cobra command
   ↓
3. Cobra parses flags and validates inputs
   ↓
4. Config loader merges CLI flags + config file
   ↓
5. Analyzer fetches repository data
   ↓
6. Analyzer calculates health metrics
   ↓
7. Formatter generates output (text/json/markdown)
   ↓
8. Reporter writes to stdout or file
   ↓
9. Exit with appropriate code
```

### Flag Definition Strategy
All flags are defined as persistent flags on the root command, making them available globally. Each flag includes:
- Long name (e.g., `--documentation-weight`)
- Short name where applicable (e.g., `-r` for `--repo`)
- Type (string, float64, bool, int)
- Default value
- Usage description
- Validation function (executed during parsing)

### Configuration Hierarchy
Configuration values are resolved in the following order (highest to lowest priority):
1. CLI flags (explicit user input)
2. Configuration file values (if `--config` specified)
3. Default values (defined in code)

## Implementation Phases

### Phase 1: Core CLI Scaffold
**Goal**: Set up Cobra framework and basic command structure

**Tasks**:
- Initialize Cobra root command in `cmd/root.go`
- Define `--repo` flag with validation
- Implement basic help text
- Set up entry point in `main.go`
- Define exit code constants

**Validation**: Running `gh-repo-health-report --help` displays usage information

### Phase 2: Flag Implementation
**Goal**: Implement all 11 CLI flags with validation

**Flags to Implement**:
1. `--repo` (string, required) - Repository identifier
2. `--documentation-weight` (float64) - Documentation metric weight
3. `--activity-weight` (float64) - Activity metric weight
4. `--community-weight` (float64) - Community metric weight
5. `--maintenance-weight` (float64) - Maintenance metric weight
6. `--format` (string) - Output format (text/json/markdown)
7. `--output` (string) - Output file path
8. `--config` (string) - Configuration file path
9. `--threshold` (int) - Health score threshold
10. `--verbose` (bool) - Enable verbose logging
11. `--help` (bool, built-in) - Display help

**Validation Logic**:
- Repository format: Must match `owner/repo` pattern
- Weight flags: Must be 0.0-1.0 range
- Format flag: Must be one of: text, json, markdown
- Threshold flag: Must be 0-100
- Config file: Must exist and be readable
- Output path: Must be writable

**Validation**: All flags parse correctly and invalid inputs produce clear error messages

### Phase 3: Configuration File Integration
**Goal**: Support YAML configuration files with CLI flag override

**Configuration Schema**:
```yaml
repository: owner/repo
weights:
  documentation: 0.25
  activity: 0.25
  community: 0.25
  maintenance: 0.25
output:
  format: text
  file: report.txt
threshold: 70
verbose: false
```

**Implementation**:
- Use Viper to load YAML files
- Bind CLI flags to Viper configuration keys
- Implement configuration validation
- Merge configuration sources (CLI flags override file values)

**Validation**: Configuration file settings are applied, and CLI flags override them correctly

### Phase 4: Output and Exit Code Handling
**Goal**: Implement output formatting and proper exit codes

**Output Implementation**:
- Text formatter: Human-readable console output with colors (if TTY)
- JSON formatter: Structured JSON with all metrics and metadata
- Markdown formatter: GitHub-flavored Markdown report
- File writer: Handle `--output` flag to write to files

**Exit Code Strategy**:
- `0`: Success - Analysis completed, score meets or exceeds threshold
- `1`: Threshold failure - Analysis completed, score below threshold
- `2`: Error - Invalid input, configuration error, or execution failure

**Validation**: Tool outputs correct formats and exits with appropriate codes

## Technical Decisions

### TD1: Flag Naming Convention
**Decision**: Use kebab-case for flag names (e.g., `--documentation-weight`)  
**Rationale**: Follows standard CLI conventions; consistent with `gh` CLI and other tools  
**Alternatives Considered**: camelCase, snake_case (rejected for poor CLI ergonomics)

### TD2: Weight Normalization
**Decision**: Automatically normalize weights to sum to 1.0 if they don't  
**Rationale**: Provides better UX; users don't need to calculate exact proportions  
**Implementation**: Sum all weight flags, divide each by sum to normalize

### TD3: Required vs. Optional Flags
**Decision**: Only `--repo` is required; all other flags have sensible defaults  
**Rationale**: Minimizes friction for basic use cases; advanced users can customize  
**Default Values**:
- Weights: 0.25 each (equal weighting)
- Format: text
- Threshold: 0 (no threshold check)
- Verbose: false

### TD4: Configuration File Format
**Decision**: Use YAML for configuration files  
**Rationale**: Human-readable, widely used in DevOps, easy to edit  
**Alternatives Considered**: JSON (too verbose), TOML (less familiar)

### TD5: Output Redirection
**Decision**: Support both stdout and file output via `--output` flag  
**Rationale**: Allows piping to other tools or saving for later analysis  
**Implementation**: If `--output` specified, write to file; else write to stdout

### TD6: Verbose Logging
**Decision**: Use boolean flag (`--verbose`) rather than log level enum  
**Rationale**: Simplifies UX for initial version; can expand to levels later if needed  
**Implementation**: Control logging output in analyzer and reporter components

### TD7: Exit Code Granularity
**Decision**: Use 3 exit codes (0, 1, 2) rather than more granular codes  
**Rationale**: Keeps it simple; 0=success, 1=threshold fail, 2=error covers most needs  
**Future Enhancement**: Could add more specific error codes if needed

### TD8: Error Message Format
**Decision**: Prefix all error messages with "Error: " and write to stderr  
**Rationale**: Follows CLI conventions; allows separation of errors from output  
**Implementation**: Use Cobra's error handling to automatically format errors

## Complexity Assessment

### Metrics
- **Files**: 5 files (1 main, 1 cmd, 3 internal packages)
- **Lines of Code**: ~420 lines total
  - cmd/root.go: ~180 lines (flag definitions + validation)
  - internal/config: ~80 lines (config loading)
  - internal/formatter: ~120 lines (output formatting)
  - main.go: ~40 lines (entry point + error handling)
- **External Dependencies**: 2 major (Cobra, Viper)
- **Flag Count**: 11 flags with 6 requiring validation

### Complexity Rating
**Medium Complexity**
- Flag parsing and validation logic is straightforward
- Configuration file handling adds moderate complexity
- Output formatting is isolated and testable
- Cobra handles most boilerplate, reducing custom code

### Risk Areas
1. **Flag Validation**: Ensuring all edge cases are caught before execution
2. **Configuration Merging**: Correctly prioritizing CLI flags over config file
3. **Weight Normalization**: Handling division by zero if all weights are 0
4. **File I/O**: Handling permissions, missing directories, disk full scenarios

## Testing Strategy

### Test Coverage Gaps (Identified During Migration)
⚠️ **No tests currently exist for CLI layer**

### Recommended Test Approach
1. **Unit Tests** (for future implementation):
   - Flag validation functions
   - Configuration loading and merging
   - Weight normalization logic
   - Output formatters (text, JSON, markdown)

2. **Integration Tests** (for future implementation):
   - End-to-end CLI execution with various flag combinations
   - Configuration file loading and override behavior
   - Exit code verification for success/failure scenarios
   - Error message validation

3. **Test Framework**: Use Go's built-in `testing` package with table-driven tests

## Migration Notes

This plan was reverse-engineered from the existing implementation. The actual build process may have evolved organically rather than following this exact phase sequence. All described functionality is currently implemented and working.

## Identified Gaps

1. **⚠️ No Tests**: CLI layer has no unit or integration tests
2. **⚠️ Limited Exit Codes**: Only 0 and 2 are actually implemented; exit code 1 (threshold failure) is defined but not used
3. **⚠️ Missing --version Flag**: No version flag implemented (common CLI convention)
4. **⚠️ No Shell Completion**: No bash/zsh completion scripts (Cobra can generate these)
