package checks

import (
	"time"

	"github.com/gclhub/gh-repo-health-report/internal/api"
)

// Check name constants.
const (
	CheckHasDescription      = "has-description"
	CheckHasHomepage         = "has-homepage"
	CheckMissingReadme       = "missing-readme"
	CheckMissingLicense      = "missing-license"
	CheckMissingCodeowners   = "missing-codeowners"
	CheckMissingSecurityMd   = "missing-security"
	CheckMissingContributing = "missing-contributing"
	CheckStale               = "stale"
	CheckHasIssues           = "has-issues"
	CheckHasProjects         = "has-projects"
	CheckHasWiki             = "has-wiki"
)

// Options configures the health checks.
type Options struct {
	Since time.Duration
}

// Result holds the health check results for a single repository.
type Result struct {
	Repository      *api.Repository
	Stale           bool
	HasDescription  bool
	HasHomepage     bool
	TopicsCount     int
	HasReadme       bool
	HasLicense      bool
	HasCodeowners   bool
	HasSecurity     bool
	HasContributing bool
	HasIssues       bool
	HasProjects     bool
	HasWiki         bool
	FailedChecks    []string
}

// Evaluate runs all health checks against a repository.
func Evaluate(repo *api.Repository, opts Options) *Result {
	r := &Result{
		Repository:      repo,
		HasDescription:  repo.Description != "",
		HasHomepage:     repo.Homepage != "",
		TopicsCount:     len(repo.Topics),
		HasReadme:       repo.HasReadme,
		HasLicense:      repo.HasLicense,
		HasCodeowners:   repo.HasCodeowners,
		HasSecurity:     repo.HasSecurity,
		HasContributing: repo.HasContributing,
		HasIssues:       repo.HasIssuesEnabled,
		HasProjects:     repo.HasProjectsEnabled,
		HasWiki:         repo.HasWikiEnabled,
	}

	threshold := opts.Since
	if threshold == 0 {
		threshold = 180 * 24 * time.Hour
	}
	if !repo.PushedAt.IsZero() && time.Since(repo.PushedAt) > threshold {
		r.Stale = true
	}

	// Collect failed checks.
	if r.Stale {
		r.FailedChecks = append(r.FailedChecks, CheckStale)
	}
	if !r.HasDescription {
		r.FailedChecks = append(r.FailedChecks, CheckHasDescription)
	}
	if !r.HasHomepage {
		r.FailedChecks = append(r.FailedChecks, CheckHasHomepage)
	}
	if !r.HasReadme {
		r.FailedChecks = append(r.FailedChecks, CheckMissingReadme)
	}
	if !r.HasLicense {
		r.FailedChecks = append(r.FailedChecks, CheckMissingLicense)
	}
	if !r.HasCodeowners {
		r.FailedChecks = append(r.FailedChecks, CheckMissingCodeowners)
	}
	if !r.HasSecurity {
		r.FailedChecks = append(r.FailedChecks, CheckMissingSecurityMd)
	}
	if !r.HasContributing {
		r.FailedChecks = append(r.FailedChecks, CheckMissingContributing)
	}
	if !r.HasIssues {
		r.FailedChecks = append(r.FailedChecks, CheckHasIssues)
	}
	if !r.HasProjects {
		r.FailedChecks = append(r.FailedChecks, CheckHasProjects)
	}
	if !r.HasWiki {
		r.FailedChecks = append(r.FailedChecks, CheckHasWiki)
	}

	return r
}
