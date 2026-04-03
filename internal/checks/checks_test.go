package checks_test

import (
	"testing"
	"time"

	"github.com/gclhub/gh-repo-health-report/internal/api"
	"github.com/gclhub/gh-repo-health-report/internal/checks"
)

func baseRepo() *api.Repository {
	return &api.Repository{
		FullName:                   "owner/repo",
		Name:                       "repo",
		Description:                "a description",
		Homepage:                   "https://example.com",
		Topics:                     []string{"go", "cli"},
		PushedAt:                   time.Now().Add(-10 * 24 * time.Hour),
		HasIssuesEnabled:           true,
		HasProjectsEnabled:         true,
		HasWikiEnabled:             true,
		HasReadme:                  true,
		HasLicense:                 true,
		HasCodeOfConduct:           true,
		HasCodeowners:              true,
		HasSecurity:                true,
		HasContributing:            true,
		HasIssueTemplates:          true,
		HasPRTemplate:              true,
		HasDependabot:              true,
		HasCIWorkflows:             true,
		DefaultBranchProtected:     true,
		HasRulesets:                true,
		VulnerabilityAlertsEnabled: true,
		SecretScanningEnabled:      true,
		PushProtectionEnabled:      true,
		DeleteBranchOnMerge:        true,
		BranchCount:                3,
		StaleBranchCount:           0,
		TagCount:                   5,
	}
}

func TestEvaluate_Healthy(t *testing.T) {
	repo := baseRepo()
	opts := checks.Options{Since: 180 * 24 * time.Hour}
	result := checks.Evaluate(repo, opts)

	if result.Stale {
		t.Error("expected not stale")
	}
	if !result.HasDescription {
		t.Error("expected HasDescription")
	}
	if !result.HasReadme {
		t.Error("expected HasReadme")
	}
	if !result.HasLicense {
		t.Error("expected HasLicense")
	}
	if len(result.FailedChecks) != 0 {
		t.Errorf("expected no failed checks, got %v", result.FailedChecks)
	}
}

func TestEvaluate_Stale(t *testing.T) {
	repo := baseRepo()
	repo.PushedAt = time.Now().Add(-400 * 24 * time.Hour)
	opts := checks.Options{Since: 180 * 24 * time.Hour}
	result := checks.Evaluate(repo, opts)

	if !result.Stale {
		t.Error("expected stale")
	}
	if !contains(result.FailedChecks, checks.CheckStale) {
		t.Errorf("expected %s in FailedChecks, got %v", checks.CheckStale, result.FailedChecks)
	}
}

func TestEvaluate_NotStale(t *testing.T) {
	repo := baseRepo()
	repo.PushedAt = time.Now().Add(-10 * 24 * time.Hour)
	opts := checks.Options{Since: 180 * 24 * time.Hour}
	result := checks.Evaluate(repo, opts)

	if result.Stale {
		t.Error("expected not stale for recent push")
	}
}

func TestEvaluate_MissingFiles(t *testing.T) {
	repo := baseRepo()
	repo.HasReadme = false
	repo.HasLicense = false
	repo.HasCodeOfConduct = false
	repo.HasCodeowners = false
	repo.HasSecurity = false
	repo.HasContributing = false
	repo.HasIssueTemplates = false
	repo.HasPRTemplate = false
	opts := checks.Options{Since: 180 * 24 * time.Hour}
	result := checks.Evaluate(repo, opts)

	for _, check := range []string{
		checks.CheckMissingReadme,
		checks.CheckMissingLicense,
		checks.CheckMissingCodeOfConduct,
		checks.CheckMissingCodeowners,
		checks.CheckMissingSecurityMd,
		checks.CheckMissingContributing,
		checks.CheckMissingIssueTemplates,
		checks.CheckMissingPRTemplate,
	} {
		if !contains(result.FailedChecks, check) {
			t.Errorf("expected %s in FailedChecks, got %v", check, result.FailedChecks)
		}
	}
}

func TestEvaluate_MissingDescription(t *testing.T) {
	repo := baseRepo()
	repo.Description = ""
	opts := checks.Options{Since: 180 * 24 * time.Hour}
	result := checks.Evaluate(repo, opts)

	if result.HasDescription {
		t.Error("expected HasDescription to be false")
	}
	if !contains(result.FailedChecks, checks.CheckHasDescription) {
		t.Errorf("expected %s in FailedChecks", checks.CheckHasDescription)
	}
}

func TestEvaluate_TopicsCount(t *testing.T) {
	repo := baseRepo()
	repo.Topics = []string{"go", "cli", "github"}
	opts := checks.Options{Since: 180 * 24 * time.Hour}
	result := checks.Evaluate(repo, opts)

	if result.TopicsCount != 3 {
		t.Errorf("expected TopicsCount=3, got %d", result.TopicsCount)
	}
}

func TestEvaluate_ExtendedChecks_AllPresent(t *testing.T) {
	repo := baseRepo()
	opts := checks.Options{Since: 180 * 24 * time.Hour}
	result := checks.Evaluate(repo, opts)

	if !result.HasDependabot {
		t.Error("expected HasDependabot")
	}
	if !result.HasCIWorkflows {
		t.Error("expected HasCIWorkflows")
	}
	if !result.DefaultBranchProtected {
		t.Error("expected DefaultBranchProtected")
	}
	if !result.HasRulesets {
		t.Error("expected HasRulesets")
	}
	if !result.VulnerabilityAlertsEnabled {
		t.Error("expected VulnerabilityAlertsEnabled")
	}
	if !result.SecretScanningEnabled {
		t.Error("expected SecretScanningEnabled")
	}
	if !result.PushProtectionEnabled {
		t.Error("expected PushProtectionEnabled")
	}
	if !result.DeleteBranchOnMerge {
		t.Error("expected DeleteBranchOnMerge")
	}
	if len(result.FailedChecks) != 0 {
		t.Errorf("expected no failed checks for healthy repo, got %v", result.FailedChecks)
	}
}

func TestEvaluate_ExtendedChecks_Missing(t *testing.T) {
	repo := baseRepo()
	repo.HasDependabot = false
	repo.HasCIWorkflows = false
	repo.DefaultBranchProtected = false
	repo.HasRulesets = false
	repo.VulnerabilityAlertsEnabled = false
	repo.SecretScanningEnabled = false
	repo.PushProtectionEnabled = false
	repo.DeleteBranchOnMerge = false
	opts := checks.Options{Since: 180 * 24 * time.Hour}
	result := checks.Evaluate(repo, opts)

	for _, check := range []string{
		checks.CheckMissingDependabot,
		checks.CheckMissingCI,
		checks.CheckNoBranchProtection,
		checks.CheckNoRulesets,
		checks.CheckNoVulnerabilityAlerts,
		checks.CheckNoSecretScanning,
		checks.CheckNoPushProtection,
		checks.CheckNoDeleteBranchOnMerge,
	} {
		if !contains(result.FailedChecks, check) {
			t.Errorf("expected %s in FailedChecks, got %v", check, result.FailedChecks)
		}
	}
}

func TestEvaluate_VulnerabilityAlerts_Unknown(t *testing.T) {
	repo := baseRepo()
	repo.VulnerabilityAlertsEnabled = false
	repo.VulnerabilityAlertsUnknown = true
	opts := checks.Options{Since: 180 * 24 * time.Hour}
	result := checks.Evaluate(repo, opts)

	// When unknown, the check should NOT appear in FailedChecks.
	if contains(result.FailedChecks, checks.CheckNoVulnerabilityAlerts) {
		t.Errorf("expected %s NOT in FailedChecks when status is unknown, got %v", checks.CheckNoVulnerabilityAlerts, result.FailedChecks)
	}
}

func TestEvaluate_SecretScanning_Unknown(t *testing.T) {
	repo := baseRepo()
	repo.SecretScanningEnabled = false
	repo.SecretScanningUnknown = true
	opts := checks.Options{Since: 180 * 24 * time.Hour}
	result := checks.Evaluate(repo, opts)

	if contains(result.FailedChecks, checks.CheckNoSecretScanning) {
		t.Errorf("expected %s NOT in FailedChecks when status is unknown, got %v", checks.CheckNoSecretScanning, result.FailedChecks)
	}
}

func TestEvaluate_PushProtection_Unknown(t *testing.T) {
	repo := baseRepo()
	repo.PushProtectionEnabled = false
	repo.PushProtectionUnknown = true
	opts := checks.Options{Since: 180 * 24 * time.Hour}
	result := checks.Evaluate(repo, opts)

	if contains(result.FailedChecks, checks.CheckNoPushProtection) {
		t.Errorf("expected %s NOT in FailedChecks when status is unknown, got %v", checks.CheckNoPushProtection, result.FailedChecks)
	}
}

func TestEvaluate_OpenIssueCountAndSize(t *testing.T) {
	repo := baseRepo()
	repo.OpenIssueCount = 42
	repo.SizeKB = 8192
	opts := checks.Options{Since: 180 * 24 * time.Hour}
	result := checks.Evaluate(repo, opts)

	if result.OpenIssueCount != 42 {
		t.Errorf("expected OpenIssueCount=42, got %d", result.OpenIssueCount)
	}
	if result.SizeKB != 8192 {
		t.Errorf("expected SizeKB=8192, got %d", result.SizeKB)
	}
}

func TestEvaluate_TooManyBranches(t *testing.T) {
	repo := baseRepo()
	repo.BranchCount = 60
	opts := checks.Options{Since: 180 * 24 * time.Hour, MaxBranches: 50}
	result := checks.Evaluate(repo, opts)

	if result.BranchCount != 60 {
		t.Errorf("expected BranchCount=60, got %d", result.BranchCount)
	}
	if !contains(result.FailedChecks, checks.CheckTooManyBranches) {
		t.Errorf("expected %s in FailedChecks, got %v", checks.CheckTooManyBranches, result.FailedChecks)
	}
}

func TestEvaluate_BranchCountWithinLimit(t *testing.T) {
	repo := baseRepo()
	repo.BranchCount = 10
	opts := checks.Options{Since: 180 * 24 * time.Hour, MaxBranches: 50}
	result := checks.Evaluate(repo, opts)

	if contains(result.FailedChecks, checks.CheckTooManyBranches) {
		t.Errorf("expected %s not in FailedChecks, got %v", checks.CheckTooManyBranches, result.FailedChecks)
	}
}

func TestEvaluate_StaleBranches(t *testing.T) {
	repo := baseRepo()
	repo.StaleBranchCount = 3
	opts := checks.Options{Since: 180 * 24 * time.Hour}
	result := checks.Evaluate(repo, opts)

	if result.StaleBranchCount != 3 {
		t.Errorf("expected StaleBranchCount=3, got %d", result.StaleBranchCount)
	}
	if !contains(result.FailedChecks, checks.CheckHasStaleBranches) {
		t.Errorf("expected %s in FailedChecks, got %v", checks.CheckHasStaleBranches, result.FailedChecks)
	}
}

func TestEvaluate_TooManyTags(t *testing.T) {
	repo := baseRepo()
	repo.TagCount = 150
	opts := checks.Options{Since: 180 * 24 * time.Hour, MaxTags: 100}
	result := checks.Evaluate(repo, opts)

	if result.TagCount != 150 {
		t.Errorf("expected TagCount=150, got %d", result.TagCount)
	}
	if !contains(result.FailedChecks, checks.CheckTooManyTags) {
		t.Errorf("expected %s in FailedChecks, got %v", checks.CheckTooManyTags, result.FailedChecks)
	}
}

func TestEvaluate_TagCountWithinLimit(t *testing.T) {
	repo := baseRepo()
	repo.TagCount = 20
	opts := checks.Options{Since: 180 * 24 * time.Hour, MaxTags: 100}
	result := checks.Evaluate(repo, opts)

	if contains(result.FailedChecks, checks.CheckTooManyTags) {
		t.Errorf("expected %s not in FailedChecks, got %v", checks.CheckTooManyTags, result.FailedChecks)
	}
}

func TestEvaluate_DefaultBranchCountThresholds(t *testing.T) {
	// With MaxBranches=0, default of 50 should apply.
	repo := baseRepo()
	repo.BranchCount = 51
	opts := checks.Options{Since: 180 * 24 * time.Hour} // MaxBranches=0 → default 50
	result := checks.Evaluate(repo, opts)

	if !contains(result.FailedChecks, checks.CheckTooManyBranches) {
		t.Errorf("expected %s in FailedChecks with default threshold, got %v", checks.CheckTooManyBranches, result.FailedChecks)
	}
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
