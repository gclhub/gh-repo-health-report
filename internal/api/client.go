package api

import (
	"fmt"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

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
	OpenIssueCount int `json:"open_issues_count"`
	SizeKB         int `json:"size"`
	// Populated by PopulateFileChecks
	HasReadme       bool `json:"has_readme"`
	HasLicense      bool `json:"has_license"`
	HasCodeowners   bool `json:"has_codeowners"`
	HasSecurity     bool `json:"has_security"`
	HasContributing bool `json:"has_contributing"`
	// Populated by PopulateExtendedChecks
	HasDependabot          bool `json:"has_dependabot"`
	HasCIWorkflows         bool `json:"has_ci_workflows"`
	DefaultBranchProtected bool `json:"default_branch_protected"`
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

// PopulateFileChecks fills HasReadme, HasLicense, HasCodeowners, HasSecurity,
// and HasContributing on repo.
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

	return nil
}

// GetCurrentUser fetches the authenticated user and unmarshals into v.
func (c *Client) GetCurrentUser(v interface{}) error {
	return c.rest.Get("user", v)
}

// PopulateExtendedChecks fills HasDependabot, HasCIWorkflows, and
// DefaultBranchProtected on repo. These require extra API round-trips beyond
// the basic file checks.
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
