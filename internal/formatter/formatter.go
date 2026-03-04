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

const tableHeader = "REPO\tSTALE\tDESCRIPTION\tTOPICS\tREADME\tLICENSE\tCODEOWNERS\tSECURITY\tCONTRIBUTING\tISSUES\tWIKI\tPROJECTS"

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
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
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
		)
	}
	return tw.Flush()
}

type jsonRow struct {
	Repo         string `json:"repo"`
	Stale        bool   `json:"stale"`
	Description  bool   `json:"has_description"`
	Topics       int    `json:"topics_count"`
	Readme       bool   `json:"has_readme"`
	License      bool   `json:"has_license"`
	Codeowners   bool   `json:"has_codeowners"`
	Security     bool   `json:"has_security"`
	Contributing bool   `json:"has_contributing"`
	Issues       bool   `json:"has_issues"`
	Wiki         bool   `json:"has_wiki"`
	Projects     bool   `json:"has_projects"`
}

func toRow(r *checks.Result) jsonRow {
	return jsonRow{
		Repo:         r.Repository.FullName,
		Stale:        r.Stale,
		Description:  r.HasDescription,
		Topics:       r.TopicsCount,
		Readme:       r.HasReadme,
		License:      r.HasLicense,
		Codeowners:   r.HasCodeowners,
		Security:     r.HasSecurity,
		Contributing: r.HasContributing,
		Issues:       r.HasIssues,
		Wiki:         r.HasWiki,
		Projects:     r.HasProjects,
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

var csvHeader = []string{"REPO", "STALE", "DESCRIPTION", "TOPICS", "README", "LICENSE", "CODEOWNERS", "SECURITY", "CONTRIBUTING", "ISSUES", "WIKI", "PROJECTS"}

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
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func formatMD(results []*checks.Result, w io.Writer) error {
	fmt.Fprintln(w, "| REPO | STALE | DESCRIPTION | TOPICS | README | LICENSE | CODEOWNERS | SECURITY | CONTRIBUTING | ISSUES | WIKI | PROJECTS |")
	fmt.Fprintln(w, "|------|-------|-------------|--------|--------|---------|------------|----------|--------------|--------|------|----------|")
	for _, r := range results {
		fmt.Fprintf(w, "| %s | %s | %s | %d | %s | %s | %s | %s | %s | %s | %s | %s |\n",
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
		)
	}
	return nil
}
