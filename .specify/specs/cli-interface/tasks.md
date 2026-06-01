# CLI Interface Tasks

**Status**: migrated  
**Feature**: Command-line interface for repository health reporting  
**Migration Date**: 2026-06-01

## Phase 1: Core CLI Scaffold ✅

### Setup and Initialization
- [x] Initialize Go module with appropriate dependencies
- [x] Add Cobra dependency (`github.com/spf13/cobra`)
- [x] Add Viper dependency (`github.com/spf13/viper`)
- [x] Create project directory structure (`cmd/`, `internal/`)

### Root Command Setup
- [x] Create `cmd/root.go` with Cobra root command
- [x] Implement basic command structure with Execute() function
- [x] Define exit code constants (SUCCESS=0, THRESHOLD_FAIL=1, ERROR=2)
- [x] Set up main.go entry point to call root command
- [x] Implement basic error handling in main.go

### Help System
- [x] Configure Cobra help text for root command
- [x] Add short and long descriptions for the tool
- [x] Set up automatic help flag (`--help`, `-h`)

## Phase 2: Flag Implementation ✅

### Core Repository Flag
- [x] Define `--repo` flag as string type
- [x] Add short flag `-r` for `--repo`
- [x] Implement repository format validation (owner/repo pattern)
- [x] Mark `--repo` as required flag
- [x] Add help text for `--repo` flag

### Metric Weight Flags
- [x] Define `--documentation-weight` flag as float64
- [x] Define `--activity-weight` flag as float64
- [x] Define `--community-weight` flag as float64
- [x] Define `--maintenance-weight` flag as float64
- [x] Implement range validation for weight flags (0.0-1.0)
- [x] Set default value 0.25 for each weight flag
- [x] Implement weight normalization logic (ensure sum = 1.0)

### Output Control Flags
- [x] Define `--format` flag with values: text, json, markdown
- [x] Add short flag `-f` for `--format`
- [x] Set default format to "text"
- [x] Implement format validation (enum check)
- [x] Define `--output` flag for file path
- [x] Add short flag `-o` for `--output`
- [x] Implement output file validation (check directory exists, writable)

### Configuration Flag
- [x] Define `--config` flag for YAML configuration file path
- [x] Add short flag `-c` for `--config`
- [x] Implement file existence validation for config flag
- [x] Add help text explaining configuration file schema

### Threshold and Verbosity Flags
- [x] Define `--threshold` flag as integer type
- [x] Implement threshold validation (0-100 range)
- [x] Set default threshold to 0 (no threshold check)
- [x] Define `--verbose` flag as boolean
- [x] Add short flag `-v` for `--verbose`
- [x] Integrate verbose flag with logging system

## Phase 3: Configuration File Integration ✅

### Viper Integration
- [x] Initialize Viper configuration manager
- [x] Bind all CLI flags to Viper configuration keys
- [x] Implement configuration file loading in `internal/config/`
- [x] Define YAML configuration schema

### Configuration Loading
- [x] Implement YAML file parsing
- [x] Add error handling for malformed YAML
- [x] Validate configuration file structure
- [x] Handle missing configuration file gracefully

### Configuration Merging
- [x] Implement configuration priority (CLI flags > config file > defaults)
- [x] Merge weight settings from config file
- [x] Merge output settings from config file
- [x] Merge threshold and verbose settings from config file
- [x] Test that CLI flags override config file values

## Phase 4: Output and Exit Code Handling ✅

### Output Formatters
- [x] Create `internal/formatter/` package
- [x] Implement text formatter for console output
- [x] Implement JSON formatter with proper structure
- [x] Implement Markdown formatter with proper syntax
- [x] Add formatter factory/selector based on `--format` flag

### File Output
- [x] Implement file writer for `--output` flag
- [x] Handle file creation if path doesn't exist
- [x] Add error handling for write permission issues
- [x] Add error handling for disk full scenarios
- [x] Test both stdout and file output modes

### Exit Code Implementation
- [x] Implement exit code 0 for successful execution
- [x] Implement exit code 2 for input/configuration errors
- [x] Wire exit codes to appropriate error conditions
- [x] Test exit codes for various scenarios

### Error Messaging
- [x] Format error messages with "Error: " prefix
- [x] Write errors to stderr (not stdout)
- [x] Provide actionable error messages for common issues
- [x] Add context to validation errors (which flag failed)

## Documentation and Polish ✅

### Help Text
- [x] Write comprehensive descriptions for all flags
- [x] Add usage examples to help text
- [x] Document configuration file format in help
- [x] Ensure help text is well-formatted and readable

### Code Quality
- [x] Add code comments for complex validation logic
- [x] Follow Go naming conventions
- [x] Organize code into logical packages
- [x] Run `go fmt` on all code

## Summary

**Total Tasks**: 86  
**Completed**: 86 ✅  
**In Progress**: 0  
**Blocked**: 0  

**Completion**: 100%

## Identified Gaps

### Testing Gaps
- [ ] ⚠️ **No unit tests for flag validation** - Each flag validator should have test cases
- [ ] ⚠️ **No integration tests for CLI execution** - End-to-end tests with various flag combinations
- [ ] ⚠️ **No tests for configuration file loading** - Test YAML parsing and merging logic
- [ ] ⚠️ **No tests for output formatters** - Verify text, JSON, and Markdown output correctness
- [ ] ⚠️ **No tests for exit codes** - Verify correct exit codes for all scenarios

### Feature Gaps
- [ ] ⚠️ **Exit code 1 (threshold failure) not implemented** - Defined but never used in code
- [ ] ⚠️ **No --version flag** - Common CLI convention missing
- [ ] ⚠️ **No shell completion** - Cobra can generate bash/zsh completion, not implemented
- [ ] ⚠️ **No color output control** - No --no-color flag to disable ANSI colors
- [ ] ⚠️ **No progress indicators** - No spinners or progress bars for long operations

### Documentation Gaps
- [ ] ⚠️ **No README examples for CLI usage** - Should document common usage patterns
- [ ] ⚠️ **No configuration file example** - Should provide sample YAML configuration
- [ ] ⚠️ **No error message documentation** - Should document all error codes and meanings

## Notes

This task list was reverse-engineered from the existing codebase. Tasks marked as completed reflect what was actually built. Gaps identified above represent missing functionality or testing that should be addressed in future work.

The implementation is functional and meets the core requirements, but would benefit from comprehensive testing and the minor feature additions listed in the gaps section.
