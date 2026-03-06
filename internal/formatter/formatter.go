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

const tableHeader = "REPO\tSTALE\tDESCRIPTION\tTOPICS\tREADME\tLICENSE\tCODEOWNERS\tSECURITY\tCONTRIBUTING\tISSUES\tWIKI\tPROJECTS\tDEPENDABOT\tCI\tBR_PROTECT\tVULN_ALERTS\tAUTO_DEL_BR\tBRANCHES\tSTALE_BR\tTAGS\tOPEN_ISSUES\tSIZE_KB"

func bool2check(v bool) string {
	if v {
		return "✓"
	}
	return "✗"
}

func staleStr(v bool) string {
	if v {
		return "YES"
	}
	return "NO"
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
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\n",
			r.Repository.FullName,
			staleStr(r.Stale),
			bool2check(r.HasDescription),
			r.TopicsCount,
			bool2check(r.HasReadme),
			bool2check(r.HasLicense),
			bool2check(r.HasCodeowners),
			bool2check(r.HasSecurity),
			bool2check(r.HasContributing),
			bool2check(r.HasIssues),
			bool2check(r.HasWiki),
			bool2check(r.HasProjects),
			bool2check(r.HasDependabot),
			bool2check(r.HasCIWorkflows),
			bool2check(r.DefaultBranchProtected),
			bool2check(r.VulnerabilityAlertsEnabled),
			bool2check(r.DeleteBranchOnMerge),
			r.BranchCount,
			r.StaleBranchCount,
			r.TagCount,
			r.OpenIssueCount,
			r.SizeKB,
		)
	}
	return tw.Flush()
}

type jsonRow struct {
	Repo                       string `json:"repo"`
	Stale                      bool   `json:"stale"`
	Description                bool   `json:"has_description"`
	Topics                     int    `json:"topics_count"`
	Readme                     bool   `json:"has_readme"`
	License                    bool   `json:"has_license"`
	Codeowners                 bool   `json:"has_codeowners"`
	Security                   bool   `json:"has_security"`
	Contributing               bool   `json:"has_contributing"`
	Issues                     bool   `json:"has_issues"`
	Wiki                       bool   `json:"has_wiki"`
	Projects                   bool   `json:"has_projects"`
	Dependabot                 bool   `json:"has_dependabot"`
	CIWorkflows                bool   `json:"has_ci_workflows"`
	DefaultBranchProtected     bool   `json:"default_branch_protected"`
	VulnerabilityAlertsEnabled bool   `json:"vulnerability_alerts_enabled"`
	DeleteBranchOnMerge        bool   `json:"delete_branch_on_merge"`
	BranchCount                int    `json:"branch_count"`
	StaleBranchCount           int    `json:"stale_branch_count"`
	TagCount                   int    `json:"tag_count"`
	OpenIssueCount             int    `json:"open_issue_count"`
	SizeKB                     int    `json:"size_kb"`
}

func toRow(r *checks.Result) jsonRow {
	return jsonRow{
		Repo:                       r.Repository.FullName,
		Stale:                      r.Stale,
		Description:                r.HasDescription,
		Topics:                     r.TopicsCount,
		Readme:                     r.HasReadme,
		License:                    r.HasLicense,
		Codeowners:                 r.HasCodeowners,
		Security:                   r.HasSecurity,
		Contributing:               r.HasContributing,
		Issues:                     r.HasIssues,
		Wiki:                       r.HasWiki,
		Projects:                   r.HasProjects,
		Dependabot:                 r.HasDependabot,
		CIWorkflows:                r.HasCIWorkflows,
		DefaultBranchProtected:     r.DefaultBranchProtected,
		VulnerabilityAlertsEnabled: r.VulnerabilityAlertsEnabled,
		DeleteBranchOnMerge:        r.DeleteBranchOnMerge,
		BranchCount:                r.BranchCount,
		StaleBranchCount:           r.StaleBranchCount,
		TagCount:                   r.TagCount,
		OpenIssueCount:             r.OpenIssueCount,
		SizeKB:                     r.SizeKB,
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

var csvHeader = []string{"REPO", "STALE", "DESCRIPTION", "TOPICS", "README", "LICENSE", "CODEOWNERS", "SECURITY", "CONTRIBUTING", "ISSUES", "WIKI", "PROJECTS", "DEPENDABOT", "CI", "BR_PROTECT", "VULN_ALERTS", "AUTO_DEL_BR", "BRANCHES", "STALE_BR", "TAGS", "OPEN_ISSUES", "SIZE_KB"}

func formatCSV(results []*checks.Result, w io.Writer) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(csvHeader); err != nil {
		return err
	}
	for _, r := range results {
		row := []string{
			r.Repository.FullName,
			staleStr(r.Stale),
			strconv.FormatBool(r.HasDescription),
			strconv.Itoa(r.TopicsCount),
			strconv.FormatBool(r.HasReadme),
			strconv.FormatBool(r.HasLicense),
			strconv.FormatBool(r.HasCodeowners),
			strconv.FormatBool(r.HasSecurity),
			strconv.FormatBool(r.HasContributing),
			strconv.FormatBool(r.HasIssues),
			strconv.FormatBool(r.HasWiki),
			strconv.FormatBool(r.HasProjects),
			strconv.FormatBool(r.HasDependabot),
			strconv.FormatBool(r.HasCIWorkflows),
			strconv.FormatBool(r.DefaultBranchProtected),
			strconv.FormatBool(r.VulnerabilityAlertsEnabled),
			strconv.FormatBool(r.DeleteBranchOnMerge),
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
	fmt.Fprintln(w, "| REPO | STALE | DESCRIPTION | TOPICS | README | LICENSE | CODEOWNERS | SECURITY | CONTRIBUTING | ISSUES | WIKI | PROJECTS | DEPENDABOT | CI | BR_PROTECT | VULN_ALERTS | AUTO_DEL_BR | BRANCHES | STALE_BR | TAGS | OPEN_ISSUES | SIZE_KB |")
	fmt.Fprintln(w, "|------|-------|-------------|--------|--------|---------|------------|----------|--------------|--------|------|----------|------------|----|-----------:|------------:|------------:|---------:|---------:|-----:|------------:|--------:|")
	for _, r := range results {
		fmt.Fprintf(w, "| %s | %s | %s | %d | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %d | %d | %d | %d | %d |\n",
			r.Repository.FullName,
			staleStr(r.Stale),
			bool2check(r.HasDescription),
			r.TopicsCount,
			bool2check(r.HasReadme),
			bool2check(r.HasLicense),
			bool2check(r.HasCodeowners),
			bool2check(r.HasSecurity),
			bool2check(r.HasContributing),
			bool2check(r.HasIssues),
			bool2check(r.HasWiki),
			bool2check(r.HasProjects),
			bool2check(r.HasDependabot),
			bool2check(r.HasCIWorkflows),
			bool2check(r.DefaultBranchProtected),
			bool2check(r.VulnerabilityAlertsEnabled),
			bool2check(r.DeleteBranchOnMerge),
			r.BranchCount,
			r.StaleBranchCount,
			r.TagCount,
			r.OpenIssueCount,
			r.SizeKB,
		)
	}
	return nil
}
