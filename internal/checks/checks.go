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
	MaxBranches int      // flag too-many-branches if BranchCount > MaxBranches; 0 = disabled
	MaxTags     int      // flag too-many-tags if TagCount > MaxTags; 0 = disabled
	Profile     *Profile // profile to apply during evaluation; nil = legacy behavior
}

// SkippedCheck represents a check that was skipped due to profile enforcement.
type SkippedCheck struct {
	Name   string
	Reason string
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
	SkippedChecks    []SkippedCheck
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

	// Helper function to add a failed check, respecting profile enforcement
	addFailedCheck := func(checkName string, failed bool) {
		// If no profile, add to failed checks (backward compatibility)
		if opts.Profile == nil {
			if failed {
				r.FailedChecks = append(r.FailedChecks, checkName)
			}
			return
		}

		// Check profile enforcement level
		enforcement, exists := opts.Profile.Checks[checkName]
		if !exists {
			// If check not in profile, treat as required (backward compatibility)
			if failed {
				r.FailedChecks = append(r.FailedChecks, checkName)
			}
			return
		}

		// Handle enforcement levels
		switch enforcement {
		case EnforcementIgnored:
			// Skip this check entirely
			r.SkippedChecks = append(r.SkippedChecks, SkippedCheck{
				Name:   checkName,
				Reason: "Ignored by profile: " + opts.Profile.Name,
			})
		case EnforcementRequired, EnforcementRecommended:
			// Evaluate and add to failed if it fails
			if failed {
				r.FailedChecks = append(r.FailedChecks, checkName)
			}
		}
	}

	// Collect failed checks.
	addFailedCheck(CheckStale, r.Stale)
	addFailedCheck(CheckHasDescription, !r.HasDescription)
	addFailedCheck(CheckHasHomepage, !r.HasHomepage)
	addFailedCheck(CheckMissingReadme, !r.HasReadme)
	addFailedCheck(CheckMissingLicense, !r.HasLicense)
	addFailedCheck(CheckMissingCodeOfConduct, !r.HasCodeOfConduct)
	addFailedCheck(CheckMissingCodeowners, !r.HasCodeowners)
	addFailedCheck(CheckMissingSecurityMd, !r.HasSecurity)
	addFailedCheck(CheckMissingContributing, !r.HasContributing)
	addFailedCheck(CheckMissingIssueTemplates, !r.HasIssueTemplates)
	addFailedCheck(CheckMissingPRTemplate, !r.HasPRTemplate)
	addFailedCheck(CheckHasIssues, !r.HasIssues)
	addFailedCheck(CheckHasProjects, !r.HasProjects)
	addFailedCheck(CheckHasWiki, !r.HasWiki)
	addFailedCheck(CheckMissingDependabot, !r.HasDependabot)
	addFailedCheck(CheckMissingCI, !r.HasCIWorkflows)
	addFailedCheck(CheckNoBranchProtection, !r.DefaultBranchProtected)
	addFailedCheck(CheckNoRulesets, !r.HasRulesets)
	// Only flag vulnerability alerts as failed when we can actually determine the status
	addFailedCheck(CheckNoVulnerabilityAlerts, !r.VulnerabilityAlertsUnknown && !r.VulnerabilityAlertsEnabled)
	// Only flag secret scanning / push protection when the status is known
	addFailedCheck(CheckNoSecretScanning, !r.SecretScanningUnknown && !r.SecretScanningEnabled)
	addFailedCheck(CheckNoPushProtection, !r.PushProtectionUnknown && !r.PushProtectionEnabled)
	addFailedCheck(CheckNoDeleteBranchOnMerge, !r.DeleteBranchOnMerge)
	// Branch count threshold (0 = use default).
	maxBranches := opts.MaxBranches
	if maxBranches == 0 {
		maxBranches = DefaultMaxBranches
	}
	addFailedCheck(CheckTooManyBranches, r.BranchCount > maxBranches)
	addFailedCheck(CheckHasStaleBranches, r.StaleBranchCount > 0)

	// Tag count threshold (0 = use default).
	maxTags := opts.MaxTags
	if maxTags == 0 {
		maxTags = DefaultMaxTags
	}
	addFailedCheck(CheckTooManyTags, r.TagCount > maxTags)

	return r
}
