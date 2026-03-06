package formatter_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/gclhub/gh-repo-health-report/internal/api"
	"github.com/gclhub/gh-repo-health-report/internal/checks"
	"github.com/gclhub/gh-repo-health-report/internal/formatter"
)

func sampleResults() []*checks.Result {
	repo := &api.Repository{
		FullName:                   "owner/repo",
		Name:                       "repo",
		Description:                "desc",
		Topics:                     []string{"go"},
		PushedAt:                   time.Now().Add(-10 * 24 * time.Hour),
		HasIssuesEnabled:           true,
		HasProjectsEnabled:         false,
		HasWikiEnabled:             true,
		HasReadme:                  true,
		HasLicense:                 false,
		HasCodeowners:              false,
		HasSecurity:                false,
		HasContributing:            false,
		HasDependabot:              true,
		HasCIWorkflows:             true,
		DefaultBranchProtected:     false,
		VulnerabilityAlertsEnabled: true,
		DeleteBranchOnMerge:        false,
		OpenIssueCount:             5,
		SizeKB:                     2048,
		BranchCount:                4,
		StaleBranchCount:           1,
		TagCount:                   10,
	}
	return []*checks.Result{checks.Evaluate(repo, checks.Options{Since: 180 * 24 * time.Hour})}
}

func TestFormatTable(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.Format(sampleResults(), "table", &buf); err != nil {
		t.Fatalf("Format table error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "REPO") {
		t.Error("table output should contain REPO header")
	}
	if !strings.Contains(out, "owner/repo") {
		t.Error("table output should contain repo name")
	}
	if !strings.Contains(out, "✓") {
		t.Error("table output should contain ✓")
	}
	if !strings.Contains(out, "✗") {
		t.Error("table output should contain ✗")
	}
}

func TestFormatJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.Format(sampleResults(), "json", &buf); err != nil {
		t.Fatalf("Format json error: %v", err)
	}
	var rows []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &rows); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["repo"] != "owner/repo" {
		t.Errorf("expected repo=owner/repo, got %v", rows[0]["repo"])
	}
	if _, ok := rows[0]["has_readme"]; !ok {
		t.Error("JSON should contain has_readme field")
	}
	if _, ok := rows[0]["branch_count"]; !ok {
		t.Error("JSON should contain branch_count field")
	}
	if _, ok := rows[0]["stale_branch_count"]; !ok {
		t.Error("JSON should contain stale_branch_count field")
	}
	if _, ok := rows[0]["tag_count"]; !ok {
		t.Error("JSON should contain tag_count field")
	}
	if _, ok := rows[0]["vulnerability_alerts_enabled"]; !ok {
		t.Error("JSON should contain vulnerability_alerts_enabled field")
	}
	if _, ok := rows[0]["delete_branch_on_merge"]; !ok {
		t.Error("JSON should contain delete_branch_on_merge field")
	}
}

func TestFormatCSV(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.Format(sampleResults(), "csv", &buf); err != nil {
		t.Fatalf("Format csv error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected header + at least 1 data row, got %d lines", len(lines))
	}
	if !strings.HasPrefix(lines[0], "REPO") {
		t.Errorf("first line should be header, got: %s", lines[0])
	}
	if !strings.Contains(lines[1], "owner/repo") {
		t.Errorf("data row should contain repo name: %s", lines[1])
	}
}

func TestFormatMD(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.Format(sampleResults(), "md", &buf); err != nil {
		t.Fatalf("Format md error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "| REPO |") {
		t.Error("md output should contain | REPO | header")
	}
	if !strings.Contains(out, "|---") {
		t.Error("md output should contain separator row")
	}
	if !strings.Contains(out, "owner/repo") {
		t.Error("md output should contain repo name")
	}
}
