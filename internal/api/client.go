package api

import (
	"fmt"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

// branchItem is used internally for branch listing.
type branchItem struct {
	Name string `json:"name"`
}

// securityFeatureStatus holds the enabled/disabled status of a GitHub Advanced
// Security feature as returned in the security_and_analysis API response.
type securityFeatureStatus struct {
	Status string `json:"status"`
}

// Repository represents a GitHub repository with health-check fields.
type Repository struct {
	FullName string `json:"full_name"`
	Name     string `json:"name"`
	Owner    struct {
		Login string `json:"login"`
	} `json:"owner"`
	Description        string    `json:"description"`
	Homepage           string    `json:"homepage"`
	Topics             []string  `json:"topics"`
	PushedAt           time.Time `json:"pushed_at"`
	DefaultBranch      string    `json:"default_branch"`
	Fork               bool      `json:"fork"`
	Archived           bool      `json:"archived"`
	HasIssuesEnabled   bool      `json:"has_issues"`
	HasProjectsEnabled bool      `json:"has_projects"`
	HasWikiEnabled     bool      `json:"has_wiki"`
	// From GitHub API metadata (returned alongside basic repo fields)
	OpenIssueCount      int  `json:"open_issues_count"`
	SizeKB              int  `json:"size"`
	DeleteBranchOnMerge bool `json:"delete_branch_on_merge"`
	// security_and_analysis is returned by the GitHub API when the caller has
	// push access or admin rights. Fields are empty strings when unavailable.
	SecurityAndAnalysis struct {
		SecretScanning struct {
			securityFeatureStatus
		} `json:"secret_scanning"`
		SecretScanningPushProtection struct {
			securityFeatureStatus
		} `json:"secret_scanning_push_protection"`
	} `json:"security_and_analysis"`
	// Populated by PopulateFileChecks
	HasReadme         bool `json:"has_readme"`
	HasLicense        bool `json:"has_license"`
	HasCodeOfConduct  bool `json:"has_code_of_conduct"`
	HasCodeowners     bool `json:"has_codeowners"`
	HasSecurity       bool `json:"has_security"`
	HasContributing   bool `json:"has_contributing"`
	HasIssueTemplates bool `json:"has_issue_templates"`
	HasPRTemplate     bool `json:"has_pr_template"`
	// Populated by PopulateExtendedChecks
	HasDependabot              bool `json:"has_dependabot"`
	HasCIWorkflows             bool `json:"has_ci_workflows"`
	DefaultBranchProtected     bool `json:"default_branch_protected"`
	HasRulesets                bool `json:"has_rulesets"`
	VulnerabilityAlertsEnabled bool `json:"vulnerability_alerts_enabled"`
	VulnerabilityAlertsUnknown bool `json:"vulnerability_alerts_unknown"`
	SecretScanningEnabled      bool `json:"secret_scanning_enabled"`
	SecretScanningUnknown      bool `json:"secret_scanning_unknown"`
	PushProtectionEnabled      bool `json:"push_protection_enabled"`
	PushProtectionUnknown      bool `json:"push_protection_unknown"`
	// Populated by PopulateBranchTagChecks
	BranchCount      int `json:"branch_count"`
	StaleBranchCount int `json:"stale_branch_count"`
	TagCount         int `json:"tag_count"`
}

// Client wraps the go-gh REST client.
type Client struct {
	rest *api.RESTClient
}

// NewClient creates a new Client using the default go-gh REST client.
func NewClient() (*Client, error) {
	rest, err := api.DefaultRESTClient()
	if err != nil {
		return nil, err
	}
	return &Client{rest: rest}, nil
}

// NewClientFromREST creates a Client from an existing RESTClient (for testing).
func NewClientFromREST(rest *api.RESTClient) *Client {
	return &Client{rest: rest}
}

// GetRepo fetches a single repository.
func (c *Client) GetRepo(owner, name string) (*Repository, error) {
	var repo Repository
	err := c.rest.Get(fmt.Sprintf("repos/%s/%s", owner, name), &repo)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

// ListOrgRepos lists all repos in an organization, handling pagination.
func (c *Client) ListOrgRepos(org string, includeForks, includeArchived bool) ([]*Repository, error) {
	return c.listRepos(fmt.Sprintf("orgs/%s/repos", org), includeForks, includeArchived)
}

// ListUserRepos lists all repos for a user, handling pagination.
func (c *Client) ListUserRepos(user string, includeForks, includeArchived bool) ([]*Repository, error) {
	return c.listRepos(fmt.Sprintf("users/%s/repos", user), includeForks, includeArchived)
}

func (c *Client) listRepos(basePath string, includeForks, includeArchived bool) ([]*Repository, error) {
	var all []*Repository
	page := 1
	for {
		var pageRepos []*Repository
		path := fmt.Sprintf("%s?per_page=100&page=%d", basePath, page)
		if err := c.rest.Get(path, &pageRepos); err != nil {
			return nil, err
		}
		for _, r := range pageRepos {
			if r.Fork && !includeForks {
				continue
			}
			if r.Archived && !includeArchived {
				continue
			}
			all = append(all, r)
		}
		if len(pageRepos) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// CheckFileExists returns true if the given path exists in the repository.
func (c *Client) CheckFileExists(owner, repo, path string) (bool, error) {
	var result interface{}
	err := c.rest.Get(fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, path), &result)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// PopulateFileChecks fills HasReadme, HasLicense, HasCodeOfConduct,
// HasCodeowners, HasSecurity, HasContributing, HasIssueTemplates, and
// HasPRTemplate on repo.
func (c *Client) PopulateFileChecks(repo *Repository) error {
	owner := repo.Owner.Login
	name := repo.Name

	// README – use the dedicated endpoint
	var ignore interface{}
	if err := c.rest.Get(fmt.Sprintf("repos/%s/%s/readme", owner, name), &ignore); err != nil {
		if !isNotFound(err) {
			return err
		}
	} else {
		repo.HasReadme = true
	}

	// LICENSE – use the dedicated endpoint
	if err := c.rest.Get(fmt.Sprintf("repos/%s/%s/license", owner, name), &ignore); err != nil {
		if !isNotFound(err) {
			return err
		}
	} else {
		repo.HasLicense = true
	}

	// CODE_OF_CONDUCT.md – checked in root, .github/, and docs/
	for _, p := range []string{"CODE_OF_CONDUCT.md", ".github/CODE_OF_CONDUCT.md", "docs/CODE_OF_CONDUCT.md"} {
		ok, err := c.CheckFileExists(owner, name, p)
		if err != nil {
			return err
		}
		if ok {
			repo.HasCodeOfConduct = true
			break
		}
	}

	// CODEOWNERS
	for _, p := range []string{".github/CODEOWNERS", "CODEOWNERS", "docs/CODEOWNERS"} {
		ok, err := c.CheckFileExists(owner, name, p)
		if err != nil {
			return err
		}
		if ok {
			repo.HasCodeowners = true
			break
		}
	}

	// SECURITY.md
	for _, p := range []string{"SECURITY.md", ".github/SECURITY.md"} {
		ok, err := c.CheckFileExists(owner, name, p)
		if err != nil {
			return err
		}
		if ok {
			repo.HasSecurity = true
			break
		}
	}

	// CONTRIBUTING.md
	for _, p := range []string{"CONTRIBUTING.md", ".github/CONTRIBUTING.md"} {
		ok, err := c.CheckFileExists(owner, name, p)
		if err != nil {
			return err
		}
		if ok {
			repo.HasContributing = true
			break
		}
	}

	// Issue templates — CheckFileExists handles both directories and files, so
	// checking .github/ISSUE_TEMPLATE will return true for the directory itself
	// when it exists. Fall back to single-file variants for repos that use a
	// simpler layout.
	for _, p := range []string{".github/ISSUE_TEMPLATE", ".github/ISSUE_TEMPLATE.md", "ISSUE_TEMPLATE.md"} {
		ok, err := c.CheckFileExists(owner, name, p)
		if err != nil {
			return err
		}
		if ok {
			repo.HasIssueTemplates = true
			break
		}
	}

	// PR template – checked in several supported locations
	for _, p := range []string{
		".github/PULL_REQUEST_TEMPLATE.md",
		".github/PULL_REQUEST_TEMPLATE",
		"PULL_REQUEST_TEMPLATE.md",
		"docs/PULL_REQUEST_TEMPLATE.md",
	} {
		ok, err := c.CheckFileExists(owner, name, p)
		if err != nil {
			return err
		}
		if ok {
			repo.HasPRTemplate = true
			break
		}
	}

	return nil
}

// GetCurrentUser fetches the authenticated user and unmarshals into v.
func (c *Client) GetCurrentUser(v interface{}) error {
	return c.rest.Get("user", v)
}

// PopulateExtendedChecks fills HasDependabot, HasCIWorkflows,
// DefaultBranchProtected, HasRulesets, VulnerabilityAlertsEnabled,
// VulnerabilityAlertsUnknown, SecretScanningEnabled, SecretScanningUnknown,
// PushProtectionEnabled, and PushProtectionUnknown on repo. These require
// extra API round-trips beyond the basic file checks.
func (c *Client) PopulateExtendedChecks(repo *Repository) error {
	owner := repo.Owner.Login
	name := repo.Name

	// Dependabot — check both .yml and .yaml extensions
	for _, p := range []string{".github/dependabot.yml", ".github/dependabot.yaml"} {
		ok, err := c.CheckFileExists(owner, name, p)
		if err != nil {
			return err
		}
		if ok {
			repo.HasDependabot = true
			break
		}
	}

	// CI workflows — check if .github/workflows/ directory has any entries
	var contents []interface{}
	err := c.rest.Get(fmt.Sprintf("repos/%s/%s/contents/.github/workflows", owner, name), &contents)
	if err != nil {
		if !isNotFound(err) {
			return err
		}
	} else {
		repo.HasCIWorkflows = len(contents) > 0
	}

	// Default branch protection — 404 means no protection; 403 means the
	// caller lacks admin access, which we treat as "unknown / unprotected"
	// so that the check still surfaces actionable signal.
	var protection interface{}
	err = c.rest.Get(
		fmt.Sprintf("repos/%s/%s/branches/%s/protection", owner, name, repo.DefaultBranch),
		&protection,
	)
	if err != nil {
		if !isNotFound(err) && !isForbidden(err) {
			return err
		}
	} else {
		repo.DefaultBranchProtected = true
	}

	// Repository rulesets — an array is returned; non-empty means rulesets
	// are configured. Rulesets are the modern replacement for classic branch
	// protection rules and are visible to anyone with read access.
	var rulesets []interface{}
	err = c.rest.Get(fmt.Sprintf("repos/%s/%s/rulesets", owner, name), &rulesets)
	if err != nil {
		if !isNotFound(err) && !isForbidden(err) {
			return err
		}
	} else {
		repo.HasRulesets = len(rulesets) > 0
	}

	// Vulnerability alerts — 204 means enabled; 404 means not enabled.
	// A 403 means the caller lacks admin rights to check the setting; in
	// that case we set VulnerabilityAlertsUnknown so the formatter can
	// distinguish "disabled" from "can't check".
	var noBody interface{}
	err = c.rest.Get(fmt.Sprintf("repos/%s/%s/vulnerability-alerts", owner, name), &noBody)
	if err != nil {
		if isForbidden(err) {
			repo.VulnerabilityAlertsUnknown = true
		} else if !isNotFound(err) {
			return err
		}
	} else {
		repo.VulnerabilityAlertsEnabled = true
	}

	// Secret scanning and push protection — derived from the
	// security_and_analysis field that is populated when GetRepo() is called.
	// The field is only present in the API response when the caller has push
	// or admin access; an empty Status string means the data is unavailable.
	if sa := repo.SecurityAndAnalysis.SecretScanning.Status; sa != "" {
		repo.SecretScanningEnabled = sa == "enabled"
	} else {
		repo.SecretScanningUnknown = true
	}
	if sa := repo.SecurityAndAnalysis.SecretScanningPushProtection.Status; sa != "" {
		repo.PushProtectionEnabled = sa == "enabled"
	} else {
		repo.PushProtectionUnknown = true
	}

	return nil
}

// PopulateBranchTagChecks fills BranchCount, StaleBranchCount, and TagCount
// on repo. A branch is considered stale if it has no commits after the
// provided cutoff time (this excludes the default branch). Each non-default
// branch requires one extra API call to check for recent commits; callers
// should be aware of the resulting rate-limit cost on repos with many branches.
func (c *Client) PopulateBranchTagChecks(repo *Repository, since time.Time) error {
	owner := repo.Owner.Login
	name := repo.Name

	// Paginate all branches.
	page := 1
	for {
		var branches []branchItem
		path := fmt.Sprintf("repos/%s/%s/branches?per_page=100&page=%d", owner, name, page)
		if err := c.rest.Get(path, &branches); err != nil {
			return err
		}
		repo.BranchCount += len(branches)

		// Check staleness for every non-default branch.
		sinceStr := since.UTC().Format(time.RFC3339)
		for _, b := range branches {
			if b.Name == repo.DefaultBranch {
				continue
			}
			var commits []interface{}
			cpath := fmt.Sprintf(
				"repos/%s/%s/commits?sha=%s&since=%s&per_page=1",
				owner, name, b.Name, sinceStr,
			)
			if err := c.rest.Get(cpath, &commits); err != nil {
				// On transient errors (rate limit, deleted branch race, etc.)
				// skip the branch rather than aborting the whole report.
				// BranchCount still reflects the total; only StaleBranchCount
				// may be under-counted in such cases.
				continue
			}
			if len(commits) == 0 {
				repo.StaleBranchCount++
			}
		}

		if len(branches) < 100 {
			break
		}
		page++
	}

	// Paginate all tags.
	page = 1
	for {
		var tags []interface{}
		path := fmt.Sprintf("repos/%s/%s/tags?per_page=100&page=%d", owner, name, page)
		if err := c.rest.Get(path, &tags); err != nil {
			return err
		}
		repo.TagCount += len(tags)
		if len(tags) < 100 {
			break
		}
		page++
	}

	return nil
}

// isNotFound checks whether an error from go-gh is an HTTP 404.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	var httpErr *api.HTTPError
	if e, ok := err.(*api.HTTPError); ok {
		httpErr = e
		return httpErr.StatusCode == 404
	}
	return false
}

// isForbidden checks whether an error from go-gh is an HTTP 403.
func isForbidden(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*api.HTTPError); ok {
		return e.StatusCode == 403
	}
	return false
}
