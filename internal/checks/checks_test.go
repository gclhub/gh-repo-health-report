package checks_test

import (
	"testing"
	"time"

	"github.com/gclhub/gh-repo-health-report/internal/api"
	"github.com/gclhub/gh-repo-health-report/internal/checks"
)

func baseRepo() *api.Repository {
	return &api.Repository{
		FullName:           "owner/repo",
		Name:               "repo",
		Description:        "a description",
		Homepage:           "https://example.com",
		Topics:             []string{"go", "cli"},
		PushedAt:           time.Now().Add(-10 * 24 * time.Hour),
		HasIssuesEnabled:   true,
		HasProjectsEnabled: true,
		HasWikiEnabled:     true,
		HasReadme:          true,
		HasLicense:         true,
		HasCodeowners:      true,
		HasSecurity:        true,
		HasContributing:    true,
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
	repo.HasCodeowners = false
	repo.HasSecurity = false
	repo.HasContributing = false
	opts := checks.Options{Since: 180 * 24 * time.Hour}
	result := checks.Evaluate(repo, opts)

	for _, check := range []string{
		checks.CheckMissingReadme,
		checks.CheckMissingLicense,
		checks.CheckMissingCodeowners,
		checks.CheckMissingSecurityMd,
		checks.CheckMissingContributing,
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

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
