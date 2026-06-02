package checks

import (
	"testing"

	"github.com/gclhub/gh-repo-health-report/internal/api"
)

// TestProfileDefinitions verifies that each predefined profile has all check constants defined.
func TestProfileDefinitions(t *testing.T) {
	// List of all check constants
	allChecks := []string{
		CheckHasDescription,
		CheckHasHomepage,
		CheckMissingReadme,
		CheckMissingLicense,
		CheckMissingCodeOfConduct,
		CheckMissingCodeowners,
		CheckMissingSecurityMd,
		CheckMissingContributing,
		CheckMissingIssueTemplates,
		CheckMissingPRTemplate,
		CheckStale,
		CheckHasIssues,
		CheckHasProjects,
		CheckHasWiki,
		CheckMissingDependabot,
		CheckMissingCI,
		CheckNoBranchProtection,
		CheckNoRulesets,
		CheckNoVulnerabilityAlerts,
		CheckNoSecretScanning,
		CheckNoPushProtection,
		CheckNoDeleteBranchOnMerge,
		CheckTooManyBranches,
		CheckHasStaleBranches,
		CheckTooManyTags,
	}

	profiles := []Profile{
		ProfileOpenSource,
		ProfileInternalService,
		ProfileApplication,
		ProfileArchived,
		ProfilePrototype,
	}

	for _, profile := range profiles {
		t.Run(profile.Name, func(t *testing.T) {
			if len(profile.Checks) != len(allChecks) {
				t.Errorf("Profile %s has %d checks, expected %d", profile.Name, len(profile.Checks), len(allChecks))
			}

			for _, checkName := range allChecks {
				if _, exists := profile.Checks[checkName]; !exists {
					t.Errorf("Profile %s missing check: %s", profile.Name, checkName)
				}
			}
		})
	}
}

// TestGetProfile verifies that GetProfile returns correct profiles for valid names.
func TestGetProfile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Profile
	}{
		{"open-source lowercase", "open-source", &ProfileOpenSource},
		{"open-source uppercase", "OPEN-SOURCE", &ProfileOpenSource},
		{"open-source mixed case", "Open-Source", &ProfileOpenSource},
		{"internal-service", "internal-service", &ProfileInternalService},
		{"application", "application", &ProfileApplication},
		{"archived", "archived", &ProfileArchived},
		{"prototype", "prototype", &ProfilePrototype},
		{"invalid name", "invalid-profile", nil},
		{"empty string", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetProfile(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil for %q, got %v", tt.input, result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected profile for %q, got nil", tt.input)
				} else if result.Name != tt.expected.Name {
					t.Errorf("Expected profile %q, got %q", tt.expected.Name, result.Name)
				}
			}
		})
	}
}

// TestDetectProfile verifies auto-detection logic.
func TestDetectProfile(t *testing.T) {
	tests := []struct {
		name     string
		repo     *api.Repository
		expected string
	}{
		{
			name:     "archived repository",
			repo:     &api.Repository{Archived: true, Private: false},
			expected: "archived",
		},
		{
			name:     "prototype topic",
			repo:     &api.Repository{Topics: []string{"prototype"}, Private: true},
			expected: "prototype",
		},
		{
			name:     "experimental topic",
			repo:     &api.Repository{Topics: []string{"experimental"}, Private: true},
			expected: "prototype",
		},
		{
			name:     "library topic",
			repo:     &api.Repository{Topics: []string{"library"}, Private: true},
			expected: "open-source",
		},
		{
			name:     "npm-package topic",
			repo:     &api.Repository{Topics: []string{"npm-package"}, Private: true},
			expected: "open-source",
		},
		{
			name:     "service topic",
			repo:     &api.Repository{Topics: []string{"service"}, Private: true},
			expected: "internal-service",
		},
		{
			name:     "api topic",
			repo:     &api.Repository{Topics: []string{"api"}, Private: true},
			expected: "internal-service",
		},
		{
			name:     "app topic",
			repo:     &api.Repository{Topics: []string{"app"}, Private: true},
			expected: "application",
		},
		{
			name:     "webapp topic",
			repo:     &api.Repository{Topics: []string{"webapp"}, Private: true},
			expected: "application",
		},
		{
			name:     "public repository no topics",
			repo:     &api.Repository{Private: false},
			expected: "open-source",
		},
		{
			name:     "private repository no topics",
			repo:     &api.Repository{Private: true},
			expected: "internal-service",
		},
		{
			name:     "topic conflict - prototype priority wins after library",
			repo:     &api.Repository{Topics: []string{"library", "prototype"}, Private: true},
			expected: "prototype",
		},
		{
			name:     "archived trumps everything",
			repo:     &api.Repository{Archived: true, Topics: []string{"service"}, Private: false},
			expected: "archived",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := DetectProfile(tt.repo)
			if profile == nil {
				t.Fatalf("DetectProfile returned nil")
			}
			if profile.Name != tt.expected {
				t.Errorf("Expected profile %q, got %q", tt.expected, profile.Name)
			}
		})
	}
}

// TestEnforcementLevelString verifies String() method for EnforcementLevel.
func TestEnforcementLevelString(t *testing.T) {
	tests := []struct {
		level    EnforcementLevel
		expected string
	}{
		{EnforcementRequired, "required"},
		{EnforcementRecommended, "recommended"},
		{EnforcementIgnored, "ignored"},
		{EnforcementLevel(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
