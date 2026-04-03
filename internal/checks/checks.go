package checks

import (
	"time"

	"github.com/gclhub/gh-repo-health-report/internal/api"
)

// Check name constants.
const (
	CheckHasDescription        = "has-description"
	CheckHasHomepage           = "has-homepage"
	CheckMissingReadme         = "missing-readme"
	CheckMissingLicense        = "missing-license"
	CheckMissingCodeOfConduct  = "missing-code-of-conduct"
	CheckMissingCodeowners     = "missing-codeowners"
	CheckMissingSecurityMd     = "missing-security"
	CheckMissingContributing   = "missing-contributing"
	CheckMissingIssueTemplates = "missing-issue-templates"
	CheckMissingPRTemplate     = "missing-pr-template"
	CheckStale                 = "stale"
	CheckHasIssues             = "has-issues"
	CheckHasProjects           = "has-projects"
	CheckHasWiki               = "has-wiki"
	// Extended checks
	CheckMissingDependabot     = "missing-dependabot"
	CheckMissingCI             = "missing-ci"
	CheckNoBranchProtection    = "no-branch-protection"
	CheckNoRulesets            = "no-rulesets"
	CheckNoVulnerabilityAlerts = "no-vulnerability-alerts"
	CheckNoSecretScanning      = "no-secret-scanning"
	CheckNoPushProtection      = "no-push-protection"
	CheckNoDeleteBranchOnMerge = "no-delete-branch-on-merge"
	// Branch and tag checks
	CheckTooManyBranches  = "too-many-branches"
	CheckHasStaleBranches = "has-stale-branches"
	CheckTooManyTags      = "too-many-tags"
)

// DefaultMaxBranches is the default threshold for the too-many-branches check.
// Repositories with more than this many branches are flagged as having excessive
// overhead; the value can be overridden via Options.MaxBranches.
const DefaultMaxBranches = 50

// DefaultMaxTags is the default threshold for the too-many-tags check.
// Repositories with more than this many tags are flagged; the value can be
// overridden via Options.MaxTags.
const DefaultMaxTags = 100

type Options struct {
	Since       time.Duration
	MaxBranches int // flag too-many-branches if BranchCount > MaxBranches; 0 = disabled
	MaxTags     int // flag too-many-tags if TagCount > MaxTags; 0 = disabled
}

// Result holds the health check results for a single repository.
type Result struct {
	Repository        *api.Repository
	Stale             bool
	HasDescription    bool
	HasHomepage       bool
	TopicsCount       int
	HasReadme         bool
	HasLicense        bool
	HasCodeOfConduct  bool
	HasCodeowners     bool
	HasSecurity       bool
	HasContributing   bool
	HasIssueTemplates bool
	HasPRTemplate     bool
	HasIssues         bool
	HasProjects       bool
	HasWiki           bool
	// Extended check results
	OpenIssueCount             int
	SizeKB                     int
	HasDependabot              bool
	HasCIWorkflows             bool
	DefaultBranchProtected     bool
	HasRulesets                bool
	VulnerabilityAlertsEnabled bool
	VulnerabilityAlertsUnknown bool
	SecretScanningEnabled      bool
	SecretScanningUnknown      bool
	PushProtectionEnabled      bool
	PushProtectionUnknown      bool
	DeleteBranchOnMerge        bool
	// Branch and tag check results
	BranchCount      int
	StaleBranchCount int
	TagCount         int
	FailedChecks     []string
}

// Evaluate runs all health checks against a repository.
func Evaluate(repo *api.Repository, opts Options) *Result {
	r := &Result{
		Repository:                 repo,
		HasDescription:             repo.Description != "",
		HasHomepage:                repo.Homepage != "",
		TopicsCount:                len(repo.Topics),
		HasReadme:                  repo.HasReadme,
		HasLicense:                 repo.HasLicense,
		HasCodeOfConduct:           repo.HasCodeOfConduct,
		HasCodeowners:              repo.HasCodeowners,
		HasSecurity:                repo.HasSecurity,
		HasContributing:            repo.HasContributing,
		HasIssueTemplates:          repo.HasIssueTemplates,
		HasPRTemplate:              repo.HasPRTemplate,
		HasIssues:                  repo.HasIssuesEnabled,
		HasProjects:                repo.HasProjectsEnabled,
		HasWiki:                    repo.HasWikiEnabled,
		OpenIssueCount:             repo.OpenIssueCount,
		SizeKB:                     repo.SizeKB,
		HasDependabot:              repo.HasDependabot,
		HasCIWorkflows:             repo.HasCIWorkflows,
		DefaultBranchProtected:     repo.DefaultBranchProtected,
		HasRulesets:                repo.HasRulesets,
		VulnerabilityAlertsEnabled: repo.VulnerabilityAlertsEnabled,
		VulnerabilityAlertsUnknown: repo.VulnerabilityAlertsUnknown,
		SecretScanningEnabled:      repo.SecretScanningEnabled,
		SecretScanningUnknown:      repo.SecretScanningUnknown,
		PushProtectionEnabled:      repo.PushProtectionEnabled,
		PushProtectionUnknown:      repo.PushProtectionUnknown,
		DeleteBranchOnMerge:        repo.DeleteBranchOnMerge,
		BranchCount:                repo.BranchCount,
		StaleBranchCount:           repo.StaleBranchCount,
		TagCount:                   repo.TagCount,
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
	if !r.HasCodeOfConduct {
		r.FailedChecks = append(r.FailedChecks, CheckMissingCodeOfConduct)
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
	if !r.HasIssueTemplates {
		r.FailedChecks = append(r.FailedChecks, CheckMissingIssueTemplates)
	}
	if !r.HasPRTemplate {
		r.FailedChecks = append(r.FailedChecks, CheckMissingPRTemplate)
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
	if !r.HasDependabot {
		r.FailedChecks = append(r.FailedChecks, CheckMissingDependabot)
	}
	if !r.HasCIWorkflows {
		r.FailedChecks = append(r.FailedChecks, CheckMissingCI)
	}
	if !r.DefaultBranchProtected {
		r.FailedChecks = append(r.FailedChecks, CheckNoBranchProtection)
	}
	if !r.HasRulesets {
		r.FailedChecks = append(r.FailedChecks, CheckNoRulesets)
	}
	// Only flag vulnerability alerts as failed when we can actually determine
	// the status (i.e. not a 403 / unknown case).
	if !r.VulnerabilityAlertsUnknown && !r.VulnerabilityAlertsEnabled {
		r.FailedChecks = append(r.FailedChecks, CheckNoVulnerabilityAlerts)
	}
	// Only flag secret scanning / push protection when the status is known.
	if !r.SecretScanningUnknown && !r.SecretScanningEnabled {
		r.FailedChecks = append(r.FailedChecks, CheckNoSecretScanning)
	}
	if !r.PushProtectionUnknown && !r.PushProtectionEnabled {
		r.FailedChecks = append(r.FailedChecks, CheckNoPushProtection)
	}
	if !r.DeleteBranchOnMerge {
		r.FailedChecks = append(r.FailedChecks, CheckNoDeleteBranchOnMerge)
	}
	// Branch count threshold (0 = use default).
	maxBranches := opts.MaxBranches
	if maxBranches == 0 {
		maxBranches = DefaultMaxBranches
	}
	if r.BranchCount > maxBranches {
		r.FailedChecks = append(r.FailedChecks, CheckTooManyBranches)
	}
	if r.StaleBranchCount > 0 {
		r.FailedChecks = append(r.FailedChecks, CheckHasStaleBranches)
	}
	// Tag count threshold (0 = use default).
	maxTags := opts.MaxTags
	if maxTags == 0 {
		maxTags = DefaultMaxTags
	}
	if r.TagCount > maxTags {
		r.FailedChecks = append(r.FailedChecks, CheckTooManyTags)
	}

	return r
}
