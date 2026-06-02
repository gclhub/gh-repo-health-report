package checks

import (
	"strings"

	"github.com/gclhub/gh-repo-health-report/internal/api"
)

// EnforcementLevel defines how a check is applied within a profile.
type EnforcementLevel int

const (
	EnforcementRequired EnforcementLevel = iota
	EnforcementRecommended
	EnforcementIgnored
)

// String returns the string representation of the enforcement level.
func (e EnforcementLevel) String() string {
	switch e {
	case EnforcementRequired:
		return "required"
	case EnforcementRecommended:
		return "recommended"
	case EnforcementIgnored:
		return "ignored"
	default:
		return "unknown"
	}
}

// Profile defines governance expectations for a repository type.
type Profile struct {
	Name        string
	Description string
	Checks      map[string]EnforcementLevel
}

// Predefined profiles
var (
	ProfileOpenSource = Profile{
		Name:        "open-source",
		Description: "Public libraries focused on community collaboration and transparency",
		Checks: map[string]EnforcementLevel{
			CheckMissingReadme:         EnforcementRequired,
			CheckMissingLicense:        EnforcementRequired,
			CheckMissingCodeOfConduct:  EnforcementRequired,
			CheckMissingContributing:   EnforcementRequired,
			CheckMissingSecurityMd:     EnforcementRequired,
			CheckHasDescription:        EnforcementRequired,
			CheckHasIssues:             EnforcementRequired,
			CheckNoSecretScanning:      EnforcementRequired,
			CheckNoVulnerabilityAlerts: EnforcementRequired,
			CheckMissingIssueTemplates: EnforcementRecommended,
			CheckMissingPRTemplate:     EnforcementRecommended,
			CheckMissingCI:             EnforcementRecommended,
			CheckHasHomepage:           EnforcementRecommended,
			CheckStale:                 EnforcementRecommended,
			CheckTooManyBranches:       EnforcementRecommended,
			CheckHasStaleBranches:      EnforcementRecommended,
			CheckMissingCodeowners:     EnforcementIgnored,
			CheckMissingDependabot:     EnforcementIgnored,
			CheckHasProjects:           EnforcementIgnored,
			CheckHasWiki:               EnforcementIgnored,
			CheckNoBranchProtection:    EnforcementIgnored,
			CheckNoRulesets:            EnforcementIgnored,
			CheckNoPushProtection:      EnforcementIgnored,
			CheckNoDeleteBranchOnMerge: EnforcementIgnored,
			CheckTooManyTags:           EnforcementIgnored,
		},
	}

	ProfileInternalService = Profile{
		Name:        "internal-service",
		Description: "Production services and APIs maintained by internal teams",
		Checks: map[string]EnforcementLevel{
			CheckMissingReadme:         EnforcementRequired,
			CheckMissingCodeowners:     EnforcementRequired,
			CheckMissingCI:             EnforcementRequired,
			CheckNoBranchProtection:    EnforcementRequired,
			CheckNoRulesets:            EnforcementRequired,
			CheckNoVulnerabilityAlerts: EnforcementRequired,
			CheckNoSecretScanning:      EnforcementRequired,
			CheckNoPushProtection:      EnforcementRequired,
			CheckNoDeleteBranchOnMerge: EnforcementRequired,
			CheckHasDescription:        EnforcementRequired,
			CheckMissingLicense:        EnforcementRecommended,
			CheckMissingSecurityMd:     EnforcementRecommended,
			CheckMissingDependabot:     EnforcementRecommended,
			CheckStale:                 EnforcementRecommended,
			CheckTooManyBranches:       EnforcementRecommended,
			CheckHasStaleBranches:      EnforcementRecommended,
			CheckMissingIssueTemplates: EnforcementRecommended,
			CheckMissingCodeOfConduct:  EnforcementIgnored,
			CheckMissingContributing:   EnforcementIgnored,
			CheckMissingPRTemplate:     EnforcementIgnored,
			CheckHasHomepage:           EnforcementIgnored,
			CheckHasIssues:             EnforcementIgnored,
			CheckHasProjects:           EnforcementIgnored,
			CheckHasWiki:               EnforcementIgnored,
			CheckTooManyTags:           EnforcementIgnored,
		},
	}

	ProfileApplication = Profile{
		Name:        "application",
		Description: "End-user applications (web apps, mobile apps, desktop software)",
		Checks: map[string]EnforcementLevel{
			CheckMissingReadme:         EnforcementRequired,
			CheckMissingLicense:        EnforcementRequired,
			CheckMissingCI:             EnforcementRequired,
			CheckNoVulnerabilityAlerts: EnforcementRequired,
			CheckNoSecretScanning:      EnforcementRequired,
			CheckHasDescription:        EnforcementRequired,
			CheckMissingSecurityMd:     EnforcementRecommended,
			CheckMissingDependabot:     EnforcementRecommended,
			CheckMissingCodeowners:     EnforcementRecommended,
			CheckNoBranchProtection:    EnforcementRecommended,
			CheckNoRulesets:            EnforcementRecommended,
			CheckStale:                 EnforcementRecommended,
			CheckTooManyBranches:       EnforcementRecommended,
			CheckHasStaleBranches:      EnforcementRecommended,
			CheckHasHomepage:           EnforcementRecommended,
			CheckMissingCodeOfConduct:  EnforcementIgnored,
			CheckMissingContributing:   EnforcementIgnored,
			CheckMissingIssueTemplates: EnforcementIgnored,
			CheckMissingPRTemplate:     EnforcementIgnored,
			CheckHasIssues:             EnforcementIgnored,
			CheckHasProjects:           EnforcementIgnored,
			CheckHasWiki:               EnforcementIgnored,
			CheckNoPushProtection:      EnforcementIgnored,
			CheckNoDeleteBranchOnMerge: EnforcementIgnored,
			CheckTooManyTags:           EnforcementIgnored,
		},
	}

	ProfileArchived = Profile{
		Name:        "archived",
		Description: "Repositories no longer under active development but retained for reference",
		Checks: map[string]EnforcementLevel{
			CheckMissingReadme:         EnforcementRequired,
			CheckMissingLicense:        EnforcementRequired,
			CheckHasDescription:        EnforcementRecommended,
			CheckHasHomepage:           EnforcementRecommended,
			CheckMissingCodeOfConduct:  EnforcementIgnored,
			CheckMissingContributing:   EnforcementIgnored,
			CheckMissingCodeowners:     EnforcementIgnored,
			CheckMissingSecurityMd:     EnforcementIgnored,
			CheckMissingIssueTemplates: EnforcementIgnored,
			CheckMissingPRTemplate:     EnforcementIgnored,
			CheckMissingDependabot:     EnforcementIgnored,
			CheckMissingCI:             EnforcementIgnored,
			CheckStale:                 EnforcementIgnored,
			CheckHasIssues:             EnforcementIgnored,
			CheckHasProjects:           EnforcementIgnored,
			CheckHasWiki:               EnforcementIgnored,
			CheckNoBranchProtection:    EnforcementIgnored,
			CheckNoRulesets:            EnforcementIgnored,
			CheckNoVulnerabilityAlerts: EnforcementIgnored,
			CheckNoSecretScanning:      EnforcementIgnored,
			CheckNoPushProtection:      EnforcementIgnored,
			CheckNoDeleteBranchOnMerge: EnforcementIgnored,
			CheckTooManyBranches:       EnforcementIgnored,
			CheckHasStaleBranches:      EnforcementIgnored,
			CheckTooManyTags:           EnforcementIgnored,
		},
	}

	ProfilePrototype = Profile{
		Name:        "prototype",
		Description: "Experimental or proof-of-concept repositories for exploration and learning",
		Checks: map[string]EnforcementLevel{
			CheckMissingReadme:         EnforcementRequired,
			CheckHasDescription:        EnforcementRequired,
			CheckMissingLicense:        EnforcementRecommended,
			CheckHasHomepage:           EnforcementRecommended,
			CheckMissingCodeOfConduct:  EnforcementIgnored,
			CheckMissingContributing:   EnforcementIgnored,
			CheckMissingCodeowners:     EnforcementIgnored,
			CheckMissingSecurityMd:     EnforcementIgnored,
			CheckMissingIssueTemplates: EnforcementIgnored,
			CheckMissingPRTemplate:     EnforcementIgnored,
			CheckMissingDependabot:     EnforcementIgnored,
			CheckMissingCI:             EnforcementIgnored,
			CheckStale:                 EnforcementIgnored,
			CheckHasIssues:             EnforcementIgnored,
			CheckHasProjects:           EnforcementIgnored,
			CheckHasWiki:               EnforcementIgnored,
			CheckNoBranchProtection:    EnforcementIgnored,
			CheckNoRulesets:            EnforcementIgnored,
			CheckNoVulnerabilityAlerts: EnforcementIgnored,
			CheckNoSecretScanning:      EnforcementIgnored,
			CheckNoPushProtection:      EnforcementIgnored,
			CheckNoDeleteBranchOnMerge: EnforcementIgnored,
			CheckTooManyBranches:       EnforcementIgnored,
			CheckHasStaleBranches:      EnforcementIgnored,
			CheckTooManyTags:           EnforcementIgnored,
		},
	}
)

// GetProfile returns a predefined profile by name, or nil if not found.
func GetProfile(name string) *Profile {
	switch strings.ToLower(name) {
	case "open-source":
		return &ProfileOpenSource
	case "internal-service":
		return &ProfileInternalService
	case "application":
		return &ProfileApplication
	case "archived":
		return &ProfileArchived
	case "prototype":
		return &ProfilePrototype
	default:
		return nil
	}
}

// DetectProfile automatically detects the appropriate profile based on repository metadata.
func DetectProfile(repo *api.Repository) *Profile {
	// Priority 1: Archived status
	if repo.Archived {
		return &ProfileArchived
	}

	// Priority 2: Topic matching
	for _, topic := range repo.Topics {
		topicLower := strings.ToLower(topic)
		// Prototype topics
		if topicLower == "prototype" || topicLower == "experimental" ||
			topicLower == "poc" || topicLower == "spike" {
			return &ProfilePrototype
		}
		// Library topics
		if topicLower == "library" || topicLower == "package" ||
			topicLower == "npm-package" || topicLower == "gem" || topicLower == "pypi" {
			return &ProfileOpenSource
		}
		// Service topics
		if topicLower == "service" || topicLower == "api" || topicLower == "microservice" {
			return &ProfileInternalService
		}
		// Application topics
		if topicLower == "app" || topicLower == "webapp" ||
			topicLower == "mobile-app" || topicLower == "desktop" {
			return &ProfileApplication
		}
	}

	// Priority 3: Visibility check
	if !repo.Private {
		return &ProfileOpenSource
	}

	// Priority 4: Fallback for private repositories
	return &ProfileInternalService
}
