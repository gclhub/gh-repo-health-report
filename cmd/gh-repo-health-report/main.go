package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gclhub/gh-repo-health-report/internal/api"
	"github.com/gclhub/gh-repo-health-report/internal/checks"
	"github.com/gclhub/gh-repo-health-report/internal/formatter"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var (
		org             string
		owner           string
		repos           []string
		includeForks    bool
		includeArchived bool
		since           string
		format          string
		output          string
		failOn          string
		maxBranches     int
		maxTags         int
		profile         string
		profileConfig   string
	)

	cmd := &cobra.Command{
		Use:   "gh-repo-health-report",
		Short: "Report on the health of GitHub repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Require at least one of --org, --owner, or --repo
			if org == "" && owner == "" && len(repos) == 0 {
				return fmt.Errorf("must specify at least one of --org, --owner, or --repo")
			}

			// Parse --since
			sinceThreshold, err := parseSince(since)
			if err != nil {
				return fmt.Errorf("invalid --since value %q: %w", since, err)
			}

			// Resolve profile: --profile flag > config file > nil
			var selectedProfile *checks.Profile
			var autoMode bool
			var cfg *checks.Config

			if profile != "" {
				// Explicit --profile flag takes precedence
				if strings.EqualFold(profile, "auto") {
					// Auto-detection will be done per-repository
					autoMode = true
				} else {
					selectedProfile = checks.GetProfile(profile)
					if selectedProfile == nil {
						return fmt.Errorf("unknown profile %q; valid profiles: open-source, internal-service, application, archived, prototype, auto", profile)
					}
				}
			} else {
				// Try loading config file
				if profileConfig != "" {
					cfg, err = checks.LoadConfig(profileConfig)
					if err != nil {
						return fmt.Errorf("failed to load config file: %w", err)
					}
				} else {
					cfg, err = checks.DiscoverConfig()
					if err != nil {
						return fmt.Errorf("failed to discover config file: %w", err)
					}
				}

				if cfg != nil && cfg.DefaultProfile != "" {
					if strings.EqualFold(cfg.DefaultProfile, "auto") {
						// Auto mode from config
						autoMode = true
					} else {
						selectedProfile = checks.GetProfile(cfg.DefaultProfile)
						if selectedProfile == nil {
							return fmt.Errorf("unknown profile %q in config file; valid profiles: open-source, internal-service, application, archived, prototype, auto", cfg.DefaultProfile)
						}
					}
				}
			}

			client, err := api.NewClient()
			if err != nil {
				return fmt.Errorf("failed to create API client: %w", err)
			}

			var repoList []*api.Repository

			switch {
			case len(repos) > 0:
				for _, r := range repos {
					parts := strings.SplitN(r, "/", 2)
					if len(parts) != 2 {
						return fmt.Errorf("invalid repo format %q, expected owner/name", r)
					}
					repo, err := client.GetRepo(parts[0], parts[1])
					if err != nil {
						return fmt.Errorf("failed to get repo %s: %w", r, err)
					}
					repoList = append(repoList, repo)
				}
			case org != "":
				repoList, err = client.ListOrgRepos(org, includeForks, includeArchived)
				if err != nil {
					return fmt.Errorf("failed to list org repos: %w", err)
				}
			case owner != "":
				repoList, err = client.ListUserRepos(owner, includeForks, includeArchived)
				if err != nil {
					return fmt.Errorf("failed to list user repos: %w", err)
				}
			}

			// Populate file checks and evaluate
			sinceTime := time.Now().Add(-sinceThreshold)
			var results []*checks.Result
			for _, repo := range repoList {
				if err := client.PopulateFileChecks(repo); err != nil {
					return fmt.Errorf("failed to check files for %s: %w", repo.FullName, err)
				}
				if err := client.PopulateExtendedChecks(repo); err != nil {
					return fmt.Errorf("failed to run extended checks for %s: %w", repo.FullName, err)
				}
				if err := client.PopulateBranchTagChecks(repo, sinceTime); err != nil {
					return fmt.Errorf("failed to run branch/tag checks for %s: %w", repo.FullName, err)
				}

				// Determine profile for this repository
				repoProfile := selectedProfile
				if autoMode {
					// Auto-detect profile based on repository metadata
					repoProfile = checks.DetectProfile(repo)
				}

				opts := checks.Options{
					Since:       sinceThreshold,
					MaxBranches: maxBranches,
					MaxTags:     maxTags,
					Profile:     repoProfile,
				}
				results = append(results, checks.Evaluate(repo, opts))
			}

			// Open output writer
			w := os.Stdout
			if output != "" {
				f, err := os.Create(output)
				if err != nil {
					return fmt.Errorf("failed to open output file: %w", err)
				}
				defer f.Close()
				w = f
			}

			if err := formatter.Format(results, format, w); err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			// --fail-on logic
			if failOn != "" {
				failChecks := splitCheckNames(failOn)
				for _, r := range results {
					if shouldFail(r, failChecks, maxBranches, maxTags) {
						os.Exit(1)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&org, "org", "", "Organization to audit")
	cmd.Flags().StringVar(&owner, "owner", "", "User to audit")
	cmd.Flags().StringArrayVar(&repos, "repo", nil, "Specific repo(s) in owner/name format (may be repeated)")
	cmd.Flags().BoolVar(&includeForks, "include-forks", false, "Include forked repos")
	cmd.Flags().BoolVar(&includeArchived, "include-archived", false, "Include archived repos")
	cmd.Flags().StringVar(&since, "since", "180d", "Staleness threshold (e.g. 180d, 6m, 1y, 2024-01-01)")
	cmd.Flags().StringVar(&format, "format", "table", "Output format: table, json, csv, md")
	cmd.Flags().StringVar(&output, "output", "", "Output file path (default stdout)")
	cmd.Flags().StringVar(&failOn, "fail-on", "", "Comma-separated check names; exit 1 if any repo fails (use 'any' to fail on any failure)")
	cmd.Flags().IntVar(&maxBranches, "max-branches", 50, "Branch count threshold for too-many-branches check (0 to disable)")
	cmd.Flags().IntVar(&maxTags, "max-tags", 100, "Tag count threshold for too-many-tags check (0 to disable)")
	cmd.Flags().StringVar(&profile, "profile", "", "Policy profile to apply (open-source, internal-service, application, archived, prototype, auto)")
	cmd.Flags().StringVar(&profileConfig, "profile-config", "", "Path to config file with default profile (YAML or JSON)")

	return cmd
}

// parseSince converts strings like "180d", "6m", "1y", or "2006-01-02" to a duration.
func parseSince(s string) (time.Duration, error) {
	if s == "" {
		return 180 * 24 * time.Hour, nil
	}
	// Try absolute date first
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return time.Since(t), nil
	}
	// Suffix-based duration
	if len(s) < 2 {
		return 0, fmt.Errorf("unrecognized duration %q", s)
	}
	suffix := s[len(s)-1]
	numStr := s[:len(s)-1]
	var n int
	if _, err := fmt.Sscanf(numStr, "%d", &n); err != nil {
		return 0, fmt.Errorf("unrecognized duration %q", s)
	}
	switch suffix {
	case 'd':
		return time.Duration(n) * 24 * time.Hour, nil
	case 'm':
		return time.Duration(n) * 30 * 24 * time.Hour, nil
	case 'y':
		return time.Duration(n) * 365 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unrecognized duration suffix %q in %q", suffix, s)
	}
}

func splitCheckNames(s string) []string {
	parts := strings.Split(s, ",")
	checks := make([]string, 0, len(parts))
	for _, part := range parts {
		if check := strings.TrimSpace(part); check != "" {
			checks = append(checks, check)
		}
	}
	return checks
}

// shouldFail returns true if any wanted check fails.
func shouldFail(r *checks.Result, wanted []string, maxBranches, maxTags int) bool {
	for _, w := range wanted {
		if w == "any" {
			if len(r.FailedChecks) > 0 {
				return true
			}
			continue
		}
		if checkFailed(r, w, maxBranches, maxTags) {
			return true
		}
	}
	return false
}

func checkFailed(r *checks.Result, checkName string, maxBranches, maxTags int) bool {
	switch checkName {
	case checks.CheckStale:
		return r.Stale
	case checks.CheckHasDescription:
		return !r.HasDescription
	case checks.CheckHasHomepage:
		return !r.HasHomepage
	case checks.CheckMissingReadme:
		return !r.HasReadme
	case checks.CheckMissingLicense:
		return !r.HasLicense
	case checks.CheckMissingCodeOfConduct:
		return !r.HasCodeOfConduct
	case checks.CheckMissingCodeowners:
		return !r.HasCodeowners
	case checks.CheckMissingSecurityMd:
		return !r.HasSecurity
	case checks.CheckMissingContributing:
		return !r.HasContributing
	case checks.CheckMissingIssueTemplates:
		return !r.HasIssueTemplates
	case checks.CheckMissingPRTemplate:
		return !r.HasPRTemplate
	case checks.CheckHasIssues:
		return !r.HasIssues
	case checks.CheckHasProjects:
		return !r.HasProjects
	case checks.CheckHasWiki:
		return !r.HasWiki
	case checks.CheckMissingDependabot:
		return !r.HasDependabot
	case checks.CheckMissingCI:
		return !r.HasCIWorkflows
	case checks.CheckNoBranchProtection:
		return !r.DefaultBranchProtected
	case checks.CheckNoRulesets:
		return !r.HasRulesets
	case checks.CheckNoVulnerabilityAlerts:
		return !r.VulnerabilityAlertsUnknown && !r.VulnerabilityAlertsEnabled
	case checks.CheckNoSecretScanning:
		return !r.SecretScanningUnknown && !r.SecretScanningEnabled
	case checks.CheckNoPushProtection:
		return !r.PushProtectionUnknown && !r.PushProtectionEnabled
	case checks.CheckNoDeleteBranchOnMerge:
		return !r.DeleteBranchOnMerge
	case checks.CheckTooManyBranches:
		if maxBranches == 0 {
			maxBranches = checks.DefaultMaxBranches
		}
		return r.BranchCount > maxBranches
	case checks.CheckHasStaleBranches:
		return r.StaleBranchCount > 0
	case checks.CheckTooManyTags:
		if maxTags == 0 {
			maxTags = checks.DefaultMaxTags
		}
		return r.TagCount > maxTags
	default:
		for _, failed := range r.FailedChecks {
			if failed == checkName {
				return true
			}
		}
		return false
	}
}
