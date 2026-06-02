package formatter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"text/tabwriter"

	"github.com/gclhub/gh-repo-health-report/internal/checks"
)

const tableHeader = "REPO\tSTALE\tDESCRIPTION\tTOPICS\tREADME\tLICENSE\tCODE_CONDUCT\tCODEOWNERS\tSECURITY\tCONTRIBUTING\tISSUE_TMPL\tPR_TMPL\tISSUES\tWIKI\tPROJECTS\tDEPENDABOT\tCI\tBR_PROTECT\tRULESETS\tVULN_ALERTS\tSECRET_SCAN\tPUSH_PROT\tAUTO_DEL_BR\tBRANCHES\tSTALE_BR\tTAGS\tOPEN_ISSUES\tSIZE_KB"

func bool2check(v bool) string {
	if v {
		return "✓"
	}
	return "✗"
}

// tristate returns ✓ when ok is true, ? when unknown is true, and ✗ otherwise.
// Use this for security settings that require sufficient GitHub permissions
// (typically push or admin access) to read: a ? means the caller lacked the
// necessary permissions to determine the status.
func tristate(ok, unknown bool) string {
	if unknown {
		return "?"
	}
	return bool2check(ok)
}

func staleStr(v bool) string {
	if v {
		return "YES"
	}
	return "NO"
}

func skippedChecksSet(r *checks.Result) map[string]struct{} {
	if len(r.SkippedChecks) == 0 {
		return nil
	}
	skipped := make(map[string]struct{}, len(r.SkippedChecks))
	for _, sc := range r.SkippedChecks {
		skipped[sc.Name] = struct{}{}
	}
	return skipped
}

func checkDisplay(checkName string, passed bool, skipped map[string]struct{}) string {
	if _, ok := skipped[checkName]; ok {
		return "[SKIP]"
	}
	return bool2check(passed)
}

func tristateCheckDisplay(checkName string, ok, unknown bool, skipped map[string]struct{}) string {
	if _, isSkipped := skipped[checkName]; isSkipped {
		return "[SKIP]"
	}
	return tristate(ok, unknown)
}

func staleDisplay(stale bool, skipped map[string]struct{}) string {
	if _, isSkipped := skipped[checks.CheckStale]; isSkipped {
		return "[SKIP]"
	}
	return staleStr(stale)
}

func isSkipped(checkName string, skipped map[string]struct{}) bool {
	_, ok := skipped[checkName]
	return ok
}

func csvCheckDisplay(checkName string, passed bool, skipped map[string]struct{}) string {
	if isSkipped(checkName, skipped) {
		return "SKIPPED"
	}
	return strconv.FormatBool(passed)
}

func csvTriStateCheckDisplay(checkName string, ok, unknown bool, skipped map[string]struct{}) string {
	if isSkipped(checkName, skipped) {
		return "SKIPPED"
	}
	return tristate(ok, unknown)
}

func csvStaleDisplay(stale bool, skipped map[string]struct{}) string {
	if isSkipped(checks.CheckStale, skipped) {
		return "SKIPPED"
	}
	return staleStr(stale)
}

// Format writes results in the requested format to w.
func Format(results []*checks.Result, format string, w io.Writer) error {
	switch format {
	case "json":
		return formatJSON(results, w)
	case "csv":
		return formatCSV(results, w)
	case "md":
		return formatMD(results, w)
	default:
		return formatTable(results, w)
	}
}

func formatTable(results []*checks.Result, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, tableHeader)
	for _, r := range results {
		skipped := skippedChecksSet(r)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\n",
			r.Repository.FullName,
			staleDisplay(r.Stale, skipped),
			checkDisplay(checks.CheckHasDescription, r.HasDescription, skipped),
			r.TopicsCount,
			checkDisplay(checks.CheckMissingReadme, r.HasReadme, skipped),
			checkDisplay(checks.CheckMissingLicense, r.HasLicense, skipped),
			checkDisplay(checks.CheckMissingCodeOfConduct, r.HasCodeOfConduct, skipped),
			checkDisplay(checks.CheckMissingCodeowners, r.HasCodeowners, skipped),
			checkDisplay(checks.CheckMissingSecurityMd, r.HasSecurity, skipped),
			checkDisplay(checks.CheckMissingContributing, r.HasContributing, skipped),
			checkDisplay(checks.CheckMissingIssueTemplates, r.HasIssueTemplates, skipped),
			checkDisplay(checks.CheckMissingPRTemplate, r.HasPRTemplate, skipped),
			checkDisplay(checks.CheckHasIssues, r.HasIssues, skipped),
			checkDisplay(checks.CheckHasWiki, r.HasWiki, skipped),
			checkDisplay(checks.CheckHasProjects, r.HasProjects, skipped),
			checkDisplay(checks.CheckMissingDependabot, r.HasDependabot, skipped),
			checkDisplay(checks.CheckMissingCI, r.HasCIWorkflows, skipped),
			checkDisplay(checks.CheckNoBranchProtection, r.DefaultBranchProtected, skipped),
			checkDisplay(checks.CheckNoRulesets, r.HasRulesets, skipped),
			tristateCheckDisplay(checks.CheckNoVulnerabilityAlerts, r.VulnerabilityAlertsEnabled, r.VulnerabilityAlertsUnknown, skipped),
			tristateCheckDisplay(checks.CheckNoSecretScanning, r.SecretScanningEnabled, r.SecretScanningUnknown, skipped),
			tristateCheckDisplay(checks.CheckNoPushProtection, r.PushProtectionEnabled, r.PushProtectionUnknown, skipped),
			checkDisplay(checks.CheckNoDeleteBranchOnMerge, r.DeleteBranchOnMerge, skipped),
			r.BranchCount,
			r.StaleBranchCount,
			r.TagCount,
			r.OpenIssueCount,
			r.SizeKB,
		)
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	// Show skipped checks summary if any profiles are active
	hasSkipped := false
	for _, r := range results {
		if len(r.SkippedChecks) > 0 {
			hasSkipped = true
			break
		}
	}

	if hasSkipped {
		fmt.Fprintln(w, "\nNote: Some checks were skipped due to policy profiles.")
		for _, r := range results {
			if len(r.SkippedChecks) > 0 {
				fmt.Fprintf(w, "  %s: %d checks skipped\n", r.Repository.FullName, len(r.SkippedChecks))
			}
		}
	}

	return nil
}

type skippedCheckRow struct {
	Check  string `json:"check"`
	Reason string `json:"reason"`
}

type jsonRow struct {
	Repo                       string            `json:"repo"`
	Stale                      bool              `json:"stale"`
	Description                bool              `json:"has_description"`
	Topics                     int               `json:"topics_count"`
	Readme                     bool              `json:"has_readme"`
	License                    bool              `json:"has_license"`
	CodeOfConduct              bool              `json:"has_code_of_conduct"`
	Codeowners                 bool              `json:"has_codeowners"`
	Security                   bool              `json:"has_security"`
	Contributing               bool              `json:"has_contributing"`
	IssueTemplates             bool              `json:"has_issue_templates"`
	PRTemplate                 bool              `json:"has_pr_template"`
	Issues                     bool              `json:"has_issues"`
	Wiki                       bool              `json:"has_wiki"`
	Projects                   bool              `json:"has_projects"`
	Dependabot                 bool              `json:"has_dependabot"`
	CIWorkflows                bool              `json:"has_ci_workflows"`
	DefaultBranchProtected     bool              `json:"default_branch_protected"`
	HasRulesets                bool              `json:"has_rulesets"`
	VulnerabilityAlertsEnabled bool              `json:"vulnerability_alerts_enabled"`
	VulnerabilityAlertsUnknown bool              `json:"vulnerability_alerts_unknown"`
	SecretScanningEnabled      bool              `json:"secret_scanning_enabled"`
	SecretScanningUnknown      bool              `json:"secret_scanning_unknown"`
	PushProtectionEnabled      bool              `json:"push_protection_enabled"`
	PushProtectionUnknown      bool              `json:"push_protection_unknown"`
	DeleteBranchOnMerge        bool              `json:"delete_branch_on_merge"`
	BranchCount                int               `json:"branch_count"`
	StaleBranchCount           int               `json:"stale_branch_count"`
	TagCount                   int               `json:"tag_count"`
	OpenIssueCount             int               `json:"open_issue_count"`
	SizeKB                     int               `json:"size_kb"`
	SkippedChecks              []skippedCheckRow `json:"skipped_checks,omitempty"`
}

func toRow(r *checks.Result) jsonRow {
	var skippedChecks []skippedCheckRow
	if len(r.SkippedChecks) > 0 {
		skippedChecks = make([]skippedCheckRow, len(r.SkippedChecks))
		for i, sc := range r.SkippedChecks {
			skippedChecks[i] = skippedCheckRow{Check: sc.Name, Reason: sc.Reason}
		}
	}

	return jsonRow{
		Repo:                       r.Repository.FullName,
		Stale:                      r.Stale,
		Description:                r.HasDescription,
		Topics:                     r.TopicsCount,
		Readme:                     r.HasReadme,
		License:                    r.HasLicense,
		CodeOfConduct:              r.HasCodeOfConduct,
		Codeowners:                 r.HasCodeowners,
		Security:                   r.HasSecurity,
		Contributing:               r.HasContributing,
		IssueTemplates:             r.HasIssueTemplates,
		PRTemplate:                 r.HasPRTemplate,
		Issues:                     r.HasIssues,
		Wiki:                       r.HasWiki,
		Projects:                   r.HasProjects,
		Dependabot:                 r.HasDependabot,
		CIWorkflows:                r.HasCIWorkflows,
		DefaultBranchProtected:     r.DefaultBranchProtected,
		HasRulesets:                r.HasRulesets,
		VulnerabilityAlertsEnabled: r.VulnerabilityAlertsEnabled,
		VulnerabilityAlertsUnknown: r.VulnerabilityAlertsUnknown,
		SecretScanningEnabled:      r.SecretScanningEnabled,
		SecretScanningUnknown:      r.SecretScanningUnknown,
		PushProtectionEnabled:      r.PushProtectionEnabled,
		PushProtectionUnknown:      r.PushProtectionUnknown,
		DeleteBranchOnMerge:        r.DeleteBranchOnMerge,
		BranchCount:                r.BranchCount,
		StaleBranchCount:           r.StaleBranchCount,
		TagCount:                   r.TagCount,
		OpenIssueCount:             r.OpenIssueCount,
		SizeKB:                     r.SizeKB,
		SkippedChecks:              skippedChecks,
	}
}

func formatJSON(results []*checks.Result, w io.Writer) error {
	rows := make([]jsonRow, len(results))
	for i, r := range results {
		rows[i] = toRow(r)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}

var csvHeader = []string{"REPO", "STALE", "DESCRIPTION", "TOPICS", "README", "LICENSE", "CODE_CONDUCT", "CODEOWNERS", "SECURITY", "CONTRIBUTING", "ISSUE_TMPL", "PR_TMPL", "ISSUES", "WIKI", "PROJECTS", "DEPENDABOT", "CI", "BR_PROTECT", "RULESETS", "VULN_ALERTS", "SECRET_SCAN", "PUSH_PROT", "AUTO_DEL_BR", "BRANCHES", "STALE_BR", "TAGS", "OPEN_ISSUES", "SIZE_KB"}

func formatCSV(results []*checks.Result, w io.Writer) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(csvHeader); err != nil {
		return err
	}
	for _, r := range results {
		skipped := skippedChecksSet(r)
		row := []string{
			r.Repository.FullName,
			csvStaleDisplay(r.Stale, skipped),
			csvCheckDisplay(checks.CheckHasDescription, r.HasDescription, skipped),
			strconv.Itoa(r.TopicsCount),
			csvCheckDisplay(checks.CheckMissingReadme, r.HasReadme, skipped),
			csvCheckDisplay(checks.CheckMissingLicense, r.HasLicense, skipped),
			csvCheckDisplay(checks.CheckMissingCodeOfConduct, r.HasCodeOfConduct, skipped),
			csvCheckDisplay(checks.CheckMissingCodeowners, r.HasCodeowners, skipped),
			csvCheckDisplay(checks.CheckMissingSecurityMd, r.HasSecurity, skipped),
			csvCheckDisplay(checks.CheckMissingContributing, r.HasContributing, skipped),
			csvCheckDisplay(checks.CheckMissingIssueTemplates, r.HasIssueTemplates, skipped),
			csvCheckDisplay(checks.CheckMissingPRTemplate, r.HasPRTemplate, skipped),
			csvCheckDisplay(checks.CheckHasIssues, r.HasIssues, skipped),
			csvCheckDisplay(checks.CheckHasWiki, r.HasWiki, skipped),
			csvCheckDisplay(checks.CheckHasProjects, r.HasProjects, skipped),
			csvCheckDisplay(checks.CheckMissingDependabot, r.HasDependabot, skipped),
			csvCheckDisplay(checks.CheckMissingCI, r.HasCIWorkflows, skipped),
			csvCheckDisplay(checks.CheckNoBranchProtection, r.DefaultBranchProtected, skipped),
			csvCheckDisplay(checks.CheckNoRulesets, r.HasRulesets, skipped),
			csvTriStateCheckDisplay(checks.CheckNoVulnerabilityAlerts, r.VulnerabilityAlertsEnabled, r.VulnerabilityAlertsUnknown, skipped),
			csvTriStateCheckDisplay(checks.CheckNoSecretScanning, r.SecretScanningEnabled, r.SecretScanningUnknown, skipped),
			csvTriStateCheckDisplay(checks.CheckNoPushProtection, r.PushProtectionEnabled, r.PushProtectionUnknown, skipped),
			csvCheckDisplay(checks.CheckNoDeleteBranchOnMerge, r.DeleteBranchOnMerge, skipped),
			strconv.Itoa(r.BranchCount),
			strconv.Itoa(r.StaleBranchCount),
			strconv.Itoa(r.TagCount),
			strconv.Itoa(r.OpenIssueCount),
			strconv.Itoa(r.SizeKB),
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func formatMD(results []*checks.Result, w io.Writer) error {
	fmt.Fprintln(w, "| REPO | STALE | DESCRIPTION | TOPICS | README | LICENSE | CODE_CONDUCT | CODEOWNERS | SECURITY | CONTRIBUTING | ISSUE_TMPL | PR_TMPL | ISSUES | WIKI | PROJECTS | DEPENDABOT | CI | BR_PROTECT | RULESETS | VULN_ALERTS | SECRET_SCAN | PUSH_PROT | AUTO_DEL_BR | BRANCHES | STALE_BR | TAGS | OPEN_ISSUES | SIZE_KB |")
	fmt.Fprintln(w, "|------|-------|-------------|--------|--------|---------|--------------|------------|----------|--------------|------------|---------|--------|------|----------|------------|----|-----------:|----------:|------------:|------------:|----------:|------------:|---------:|---------:|-----:|------------:|--------:|")
	for _, r := range results {
		skipped := skippedChecksSet(r)
		fmt.Fprintf(w, "| %s | %s | %s | %d | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %d | %d | %d | %d | %d |\n",
			r.Repository.FullName,
			staleDisplay(r.Stale, skipped),
			checkDisplay(checks.CheckHasDescription, r.HasDescription, skipped),
			r.TopicsCount,
			checkDisplay(checks.CheckMissingReadme, r.HasReadme, skipped),
			checkDisplay(checks.CheckMissingLicense, r.HasLicense, skipped),
			checkDisplay(checks.CheckMissingCodeOfConduct, r.HasCodeOfConduct, skipped),
			checkDisplay(checks.CheckMissingCodeowners, r.HasCodeowners, skipped),
			checkDisplay(checks.CheckMissingSecurityMd, r.HasSecurity, skipped),
			checkDisplay(checks.CheckMissingContributing, r.HasContributing, skipped),
			checkDisplay(checks.CheckMissingIssueTemplates, r.HasIssueTemplates, skipped),
			checkDisplay(checks.CheckMissingPRTemplate, r.HasPRTemplate, skipped),
			checkDisplay(checks.CheckHasIssues, r.HasIssues, skipped),
			checkDisplay(checks.CheckHasWiki, r.HasWiki, skipped),
			checkDisplay(checks.CheckHasProjects, r.HasProjects, skipped),
			checkDisplay(checks.CheckMissingDependabot, r.HasDependabot, skipped),
			checkDisplay(checks.CheckMissingCI, r.HasCIWorkflows, skipped),
			checkDisplay(checks.CheckNoBranchProtection, r.DefaultBranchProtected, skipped),
			checkDisplay(checks.CheckNoRulesets, r.HasRulesets, skipped),
			tristateCheckDisplay(checks.CheckNoVulnerabilityAlerts, r.VulnerabilityAlertsEnabled, r.VulnerabilityAlertsUnknown, skipped),
			tristateCheckDisplay(checks.CheckNoSecretScanning, r.SecretScanningEnabled, r.SecretScanningUnknown, skipped),
			tristateCheckDisplay(checks.CheckNoPushProtection, r.PushProtectionEnabled, r.PushProtectionUnknown, skipped),
			checkDisplay(checks.CheckNoDeleteBranchOnMerge, r.DeleteBranchOnMerge, skipped),
			r.BranchCount,
			r.StaleBranchCount,
			r.TagCount,
			r.OpenIssueCount,
			r.SizeKB,
		)
	}
	return nil
}
