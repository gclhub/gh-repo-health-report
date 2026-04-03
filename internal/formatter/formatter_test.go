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
		HasCodeOfConduct:           false,
		HasCodeowners:              false,
		HasSecurity:                false,
		HasContributing:            false,
		HasIssueTemplates:          false,
		HasPRTemplate:              false,
		HasDependabot:              true,
		HasCIWorkflows:             true,
		DefaultBranchProtected:     false,
		HasRulesets:                false,
		VulnerabilityAlertsEnabled: true,
		VulnerabilityAlertsUnknown: false,
		SecretScanningEnabled:      false,
		SecretScanningUnknown:      true, // status not available → show "?"
		PushProtectionEnabled:      false,
		PushProtectionUnknown:      true, // status not available → show "?"
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
	if !strings.Contains(out, "CODE_CONDUCT") {
		t.Error("table output should contain CODE_CONDUCT header")
	}
	if !strings.Contains(out, "RULESETS") {
		t.Error("table output should contain RULESETS header")
	}
	if !strings.Contains(out, "SECRET_SCAN") {
		t.Error("table output should contain SECRET_SCAN header")
	}
	if !strings.Contains(out, "PUSH_PROT") {
		t.Error("table output should contain PUSH_PROT header")
	}
	if !strings.Contains(out, "ISSUE_TMPL") {
		t.Error("table output should contain ISSUE_TMPL header")
	}
	if !strings.Contains(out, "PR_TMPL") {
		t.Error("table output should contain PR_TMPL header")
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
	// SecretScanningUnknown and PushProtectionUnknown are true → expect "?"
	if !strings.Contains(out, "?") {
		t.Error("table output should contain ? for unknown security settings")
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
	for _, field := range []string{
		"has_readme", "has_code_of_conduct", "has_codeowners",
		"has_issue_templates", "has_pr_template",
		"branch_count", "stale_branch_count", "tag_count",
		"has_rulesets",
		"vulnerability_alerts_enabled", "vulnerability_alerts_unknown",
		"secret_scanning_enabled", "secret_scanning_unknown",
		"push_protection_enabled", "push_protection_unknown",
		"delete_branch_on_merge",
	} {
		if _, ok := rows[0][field]; !ok {
			t.Errorf("JSON should contain %s field", field)
		}
	}
	// VulnerabilityAlertsUnknown is false → should be false in JSON
	if rows[0]["vulnerability_alerts_unknown"] != false {
		t.Errorf("expected vulnerability_alerts_unknown=false, got %v", rows[0]["vulnerability_alerts_unknown"])
	}
	// SecretScanningUnknown is true → should be true in JSON
	if rows[0]["secret_scanning_unknown"] != true {
		t.Errorf("expected secret_scanning_unknown=true, got %v", rows[0]["secret_scanning_unknown"])
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
	for _, col := range []string{"CODE_CONDUCT", "RULESETS", "SECRET_SCAN", "PUSH_PROT", "ISSUE_TMPL", "PR_TMPL"} {
		if !strings.Contains(lines[0], col) {
			t.Errorf("CSV header should contain %s, got: %s", col, lines[0])
		}
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
	if !strings.Contains(out, "CODE_CONDUCT") {
		t.Error("md output should contain CODE_CONDUCT column")
	}
	if !strings.Contains(out, "RULESETS") {
		t.Error("md output should contain RULESETS column")
	}
	if !strings.Contains(out, "|---") {
		t.Error("md output should contain separator row")
	}
	if !strings.Contains(out, "owner/repo") {
		t.Error("md output should contain repo name")
	}
}
