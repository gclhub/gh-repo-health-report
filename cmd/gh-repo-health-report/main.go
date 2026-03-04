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
	)

	cmd := &cobra.Command{
		Use:   "gh-repo-health-report",
		Short: "Report on the health of GitHub repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse --since
			sinceThreshold, err := parseSince(since)
			if err != nil {
				return fmt.Errorf("invalid --since value %q: %w", since, err)
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
			default:
				// Default: current authenticated user
				var user struct {
					Login string `json:"login"`
				}
				if err := client.GetCurrentUser(&user); err != nil {
					return fmt.Errorf("failed to get current user: %w", err)
				}
				repoList, err = client.ListUserRepos(user.Login, includeForks, includeArchived)
				if err != nil {
					return fmt.Errorf("failed to list repos for current user: %w", err)
				}
			}

			// Populate file checks and evaluate
			opts := checks.Options{Since: sinceThreshold}
			var results []*checks.Result
			for _, repo := range repoList {
				if err := client.PopulateFileChecks(repo); err != nil {
					return fmt.Errorf("failed to check files for %s: %w", repo.FullName, err)
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
				failChecks := strings.Split(failOn, ",")
				for _, r := range results {
					if shouldFail(r.FailedChecks, failChecks) {
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

// shouldFail returns true if any of the wanted checks are in failed.
func shouldFail(failed, wanted []string) bool {
	for _, w := range wanted {
		if w == "any" && len(failed) > 0 {
			return true
		}
		for _, f := range failed {
			if f == w {
				return true
			}
		}
	}
	return false
}
