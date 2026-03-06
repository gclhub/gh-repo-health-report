# gh repo-health-report

A GitHub CLI extension that audits the health of GitHub repositories — checking for README, LICENSE, CODEOWNERS, SECURITY.md, CONTRIBUTING.md, staleness, topics, branch hygiene, and more.

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

# Use custom branch/tag thresholds
gh repo-health-report --org myorg --max-branches 30 --max-tags 50
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--org` | | Organization to audit |
| `--owner` | | User to audit |
| `--repo` | | Specific repo(s) in `owner/name` format (repeatable) |
| `--include-forks` | false | Include forked repos |
| `--include-archived` | false | Include archived repos |
| `--since` | `180d` | Staleness threshold (`180d`, `6m`, `1y`, `2024-01-01`); also used for stale-branch detection |
| `--format` | `table` | Output format: `table`, `json`, `csv`, `md` |
| `--output` | stdout | Output file path |
| `--fail-on` | | Comma-separated check names; exit 1 if any repo fails them (use `any`) |
| `--max-branches` | `50` | Branch count threshold for `too-many-branches` check (0 to disable) |
| `--max-tags` | `100` | Tag count threshold for `too-many-tags` check (0 to disable) |

## Check names

| Check name | Fails when |
|---|---|
| `missing-readme` | No README detected |
| `missing-license` | No LICENSE detected |
| `missing-codeowners` | No CODEOWNERS file |
| `missing-security` | No SECURITY.md |
| `missing-contributing` | No CONTRIBUTING.md |
| `stale` | No push since `--since` threshold |
| `has-description` | Repository description is empty |
| `has-homepage` | Repository homepage URL is empty |
| `has-issues` | Issues are disabled |
| `has-projects` | Projects are disabled |
| `has-wiki` | Wiki is disabled |
| `missing-dependabot` | No `.github/dependabot.yml` |
| `missing-ci` | No files in `.github/workflows/` |
| `no-branch-protection` | Default branch has no protection rules |
| `no-vulnerability-alerts` | Dependabot security alerts are not enabled |
| `no-delete-branch-on-merge` | Auto-delete head branches is disabled |
| `too-many-branches` | Branch count exceeds `--max-branches` (default 50) |
| `has-stale-branches` | One or more non-default branches have no commits since `--since` |
| `too-many-tags` | Tag count exceeds `--max-tags` (default 100) |

## Example output (table)

```
REPO              STALE  DESCRIPTION  TOPICS  README  LICENSE  CODEOWNERS  SECURITY  CONTRIBUTING  ISSUES  WIKI  PROJECTS  DEPENDABOT  CI  BR_PROTECT  VULN_ALERTS  AUTO_DEL_BR  BRANCHES  STALE_BR  TAGS  OPEN_ISSUES  SIZE_KB
owner/my-project  NO     ✓            3       ✓       ✓        ✗           ✗         ✗             ✓       ✓     ✗         ✓           ✓   ✗           ✓            ✗            4         1         12    12           4096
```

Columns:
- **REPO** — full repository name
- **STALE** — no push since `--since` threshold
- **DESCRIPTION**, **TOPICS** — metadata completeness
- **README**, **LICENSE**, **CODEOWNERS**, **SECURITY**, **CONTRIBUTING** — standard repo files
- **ISSUES**, **WIKI**, **PROJECTS** — GitHub features enabled
- **DEPENDABOT**, **CI**, **BR_PROTECT** — automation and security baseline
- **VULN_ALERTS** — Dependabot security alerts enabled
- **AUTO_DEL_BR** — auto-delete head branches on merge enabled
- **BRANCHES** — total branch count; **STALE_BR** — branches with no commits since `--since`
- **TAGS** — total tag count
- **OPEN_ISSUES** — open issues + PRs count
- **SIZE_KB** — repository disk size in KB
