# gh repo-health-report

A GitHub CLI extension that audits the health of GitHub repositories — checking for README, LICENSE, CODEOWNERS, SECURITY.md, CONTRIBUTING.md, staleness, topics, and more.

## Installation

```bash
gh extension install gclhub/gh-repo-health-report
```

## Usage

```bash
# Audit all repos in an organization
gh repo-health-report --org myorg

# Audit all repos for a user
gh repo-health-report --owner myuser

# Audit one or more specific repos
gh repo-health-report --repo owner/repo

# Output as JSON
gh repo-health-report --org myorg --format json

# Output as CSV to a file
gh repo-health-report --org myorg --format csv --output report.csv

# Exit 1 if any repo is missing a README or LICENSE
gh repo-health-report --org myorg --fail-on missing-readme,missing-license
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--org` | | Organization to audit |
| `--owner` | | User to audit |
| `--repo` | | Specific repo(s) in `owner/name` format (repeatable) |
| `--include-forks` | false | Include forked repos |
| `--include-archived` | false | Include archived repos |
| `--since` | `180d` | Staleness threshold (`180d`, `6m`, `1y`, `2024-01-01`) |
| `--format` | `table` | Output format: `table`, `json`, `csv`, `md` |
| `--output` | stdout | Output file path |
| `--fail-on` | | Comma-separated check names; exit 1 if any repo fails them (use `any`) |

## Check names

`missing-readme`, `missing-license`, `missing-codeowners`, `missing-security`, `missing-contributing`, `stale`, `has-description`, `has-homepage`, `has-issues`, `has-projects`, `has-wiki`

## Example output (table)

```
REPO              STALE  DESCRIPTION  TOPICS  README  LICENSE  CODEOWNERS  SECURITY  CONTRIBUTING  ISSUES  WIKI  PROJECTS
owner/my-project  NO     ✓            3       ✓       ✓        ✗           ✗         ✗             ✓       ✓     ✗
```

Columns: **REPO**, **STALE**, **DESCRIPTION**, **TOPICS**, **README**, **LICENSE**, **CODEOWNERS**, **SECURITY**, **CONTRIBUTING**, **ISSUES**, **WIKI**, **PROJECTS**
