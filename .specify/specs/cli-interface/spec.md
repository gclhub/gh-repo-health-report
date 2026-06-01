# CLI Interface Specification

**Status**: migrated  
**Feature**: Command-line interface for repository health reporting  
**Migration Date**: 2026-06-01

## Overview

A command-line tool that analyzes GitHub repositories and generates health reports based on configurable metrics. The CLI accepts repository identifiers and configuration flags, then outputs health scores and detailed metric reports.

## User Scenarios

### Scenario 1: Basic Repository Analysis
**As a** developer  
**I want to** analyze a single repository with default settings  
**So that** I can quickly assess its health status

```bash
gh-repo-health-report --repo owner/repo
```

**Expected Outcome**:
- Tool fetches repository data from GitHub API
- Calculates health metrics using default weights
- Displays overall health score and metric breakdown
- Exits with code 0 on success

### Scenario 2: Custom Metric Weights
**As a** team lead  
**I want to** customize metric weights based on team priorities  
**So that** the health score reflects what matters most to our workflow

```bash
gh-repo-health-report --repo owner/repo \
  --documentation-weight 0.3 \
  --activity-weight 0.2 \
  --community-weight 0.2 \
  --maintenance-weight 0.3
```

**Expected Outcome**:
- Tool applies custom weights to each metric category
- Health score reflects the prioritized weighting
- Report shows which metrics contributed most to the score

### Scenario 3: Output Format Selection
**As a** CI/CD pipeline  
**I want to** receive health reports in JSON format  
**So that** I can parse and process the results programmatically

```bash
gh-repo-health-report --repo owner/repo --format json
```

**Expected Outcome**:
- Tool outputs structured JSON with health metrics
- JSON includes all metric scores, overall health, and metadata
- Format is machine-readable and parseable

### Scenario 4: Verbose Debugging
**As a** developer troubleshooting issues  
**I want to** see detailed execution information  
**So that** I can understand what the tool is doing and diagnose problems

```bash
gh-repo-health-report --repo owner/repo --verbose
```

**Expected Outcome**:
- Tool displays detailed logging information
- Shows API calls being made, data retrieved, and calculations performed
- Helps identify where issues occur in the analysis pipeline

### Scenario 5: Configuration File Usage
**As a** team member  
**I want to** use a shared configuration file for consistent analysis  
**So that** all team members use the same settings without memorizing flags

```bash
gh-repo-health-report --repo owner/repo --config team-config.yaml
```

**Expected Outcome**:
- Tool loads configuration from specified YAML file
- Configuration overrides default settings
- Allows standardization across team analyses

## Requirements

### Functional Requirements

#### FR1: Repository Identification
- **FR1.1**: MUST accept `--repo` flag with format `owner/repository`
- **FR1.2**: MUST validate repository format before making API calls
- **FR1.3**: MUST support both public and private repositories (with authentication)

#### FR2: Metric Weight Customization
- **FR2.1**: MUST accept `--documentation-weight` flag (float, 0.0-1.0)
- **FR2.2**: MUST accept `--activity-weight` flag (float, 0.0-1.0)
- **FR2.3**: MUST accept `--community-weight` flag (float, 0.0-1.0)
- **FR2.4**: MUST accept `--maintenance-weight` flag (float, 0.0-1.0)
- **FR2.5**: MUST validate that all weights are non-negative
- **FR2.6**: MUST normalize weights to sum to 1.0 if they don't already

#### FR3: Output Format Control
- **FR3.1**: MUST accept `--format` flag with values: `text`, `json`, `markdown`
- **FR3.2**: MUST default to `text` format if not specified
- **FR3.3**: MUST produce valid JSON when `json` format is selected
- **FR3.4**: MUST produce properly formatted Markdown when `markdown` format is selected

#### FR4: Verbosity Control
- **FR4.1**: MUST accept `--verbose` boolean flag
- **FR4.2**: MUST display detailed logging when verbose mode is enabled
- **FR4.3**: MUST suppress detailed logging in normal mode (default)

#### FR5: Configuration File Support
- **FR5.1**: MUST accept `--config` flag with path to YAML configuration file
- **FR5.2**: MUST validate configuration file exists and is readable
- **FR5.3**: MUST parse YAML configuration and apply settings
- **FR5.4**: MUST allow CLI flags to override configuration file settings

#### FR6: Output File
- **FR6.1**: MUST accept `--output` flag with path to output file
- **FR6.2**: MUST write report to specified file instead of stdout
- **FR6.3**: MUST create output file if it doesn't exist
- **FR6.4**: MUST fail gracefully if output path is not writable

#### FR7: Threshold Configuration
- **FR7.1**: MUST accept `--threshold` flag (integer, 0-100)
- **FR7.2**: MUST use threshold to determine pass/fail status
- **FR7.3**: MUST exit with code 1 if health score is below threshold

#### FR8: Help and Usage
- **FR8.1**: MUST accept `--help` or `-h` flag to display usage information
- **FR8.2**: MUST display all available flags with descriptions
- **FR8.3**: MUST show example usage patterns

### Non-Functional Requirements

#### NFR1: Input Validation
- All flag values MUST be validated before processing
- Invalid inputs MUST produce clear error messages
- Error messages MUST indicate which flag contains invalid data

#### NFR2: Exit Codes
- MUST exit with code 0 on successful analysis
- MUST exit with code 1 when health score is below threshold
- MUST exit with code 2 for invalid input or configuration errors

#### NFR3: Performance
- Flag parsing MUST complete in under 100ms
- CLI initialization MUST not block on network calls

#### NFR4: Usability
- Error messages MUST be user-friendly and actionable
- Help text MUST be clear and comprehensive
- Flag names MUST follow standard CLI conventions (kebab-case)

## Success Criteria

### SC1: Repository Analysis Execution
- ✅ User can analyze any repository with just `--repo owner/name`
- ✅ Tool successfully fetches data and calculates health metrics
- ✅ Results are displayed in human-readable format

### SC2: Flag Parsing and Validation
- ✅ All 11 flags are recognized and parsed correctly
- ✅ Invalid flag values produce clear error messages
- ✅ Flag combinations work together without conflicts

### SC3: Output Format Support
- ✅ Text format produces readable console output
- ✅ JSON format produces valid, parseable JSON
- ✅ Markdown format produces properly formatted Markdown
- ✅ Output can be redirected to files using `--output`

### SC4: Configuration Management
- ✅ YAML configuration files are loaded and applied
- ✅ CLI flags override configuration file settings
- ✅ Invalid configuration produces helpful error messages

### SC5: Help and Documentation
- ✅ `--help` displays comprehensive usage information
- ✅ All flags are documented in help text
- ✅ Example usage patterns are provided

### SC6: Error Handling
- ✅ Invalid repository format is caught and reported
- ✅ Missing required flags are detected and reported
- ✅ File I/O errors are handled gracefully
- ✅ Network errors are caught and reported clearly

## Assumptions

- **A1**: GitHub authentication is handled externally (via `gh` CLI or environment variables)
- **A2**: Users have `gh` CLI tool installed and authenticated
- **A3**: Repository analysis is synchronous (no background processing)
- **A4**: Default metric weights are equal (0.25 each for 4 categories)
- **A5**: YAML configuration files follow a documented schema
- **A6**: Output formats are mutually exclusive (cannot output multiple formats simultaneously)

## Out of Scope

- Interactive mode / TUI interface
- Batch processing of multiple repositories in a single command
- Historical trend analysis across multiple runs
- Automatic remediation suggestions
- Integration with other CLI tools beyond `gh`
- GUI or web interface

## Open Questions

*None at time of migration*

## Notes

This specification was reverse-engineered from the existing codebase during migration to spec-kit. Some requirements may reflect implementation details rather than original design intent.

## Identified Gaps

1. **⚠️ No Tests**: No unit or integration tests exist for CLI flag parsing and validation
2. **⚠️ Limited Exit Codes**: Only basic exit codes implemented; missing granular error codes
3. **⚠️ Missing --version Flag**: No `--version` flag to display tool version information
