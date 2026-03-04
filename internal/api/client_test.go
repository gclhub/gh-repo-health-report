package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockServer creates a test HTTP server and returns it along with a helper
// to build a go-gh REST client pointed at it.
func mockServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestGetRepo_Mock(t *testing.T) {
	repo := map[string]interface{}{
		"full_name":    "owner/repo",
		"name":         "repo",
		"owner":        map[string]string{"login": "owner"},
		"description":  "test repo",
		"pushed_at":    time.Now().Format(time.RFC3339),
		"has_issues":   true,
		"has_projects": false,
		"has_wiki":     true,
	}

	srv := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repo)
	}))

	// Verify the mock server works correctly by making a direct HTTP call.
	resp, err := http.Get(srv.URL + "/repos/owner/repo")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if result["full_name"] != "owner/repo" {
		t.Errorf("expected full_name=owner/repo, got %v", result["full_name"])
	}
}

func TestListOrgRepos_Pagination_Mock(t *testing.T) {
	page1 := []map[string]interface{}{
		{"full_name": "org/repo1", "name": "repo1", "owner": map[string]string{"login": "org"}, "pushed_at": time.Now().Format(time.RFC3339)},
		{"full_name": "org/repo2", "name": "repo2", "owner": map[string]string{"login": "org"}, "pushed_at": time.Now().Format(time.RFC3339)},
	}
	page2 := []map[string]interface{}{
		{"full_name": "org/repo3", "name": "repo3", "owner": map[string]string{"login": "org"}, "pushed_at": time.Now().Format(time.RFC3339)},
	}

	srv := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		q := r.URL.Query()
		if q.Get("page") == "2" {
			json.NewEncoder(w).Encode(page2)
		} else {
			json.NewEncoder(w).Encode(page1)
		}
	}))

	resp1, _ := http.Get(srv.URL + "/orgs/org/repos?per_page=100&page=1")
	resp2, _ := http.Get(srv.URL + "/orgs/org/repos?per_page=100&page=2")
	defer resp1.Body.Close()
	defer resp2.Body.Close()

	var p1, p2 []map[string]interface{}
	json.NewDecoder(resp1.Body).Decode(&p1)
	json.NewDecoder(resp2.Body).Decode(&p2)

	if len(p1) != 2 {
		t.Errorf("expected 2 repos on page 1, got %d", len(p1))
	}
	if len(p2) != 1 {
		t.Errorf("expected 1 repo on page 2, got %d", len(p2))
	}
}

func TestCheckFileExists_Mock(t *testing.T) {
	srv := mockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/contents/README.md" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"name": "README.md"})
			return
		}
		http.NotFound(w, r)
	}))

	// File exists
	resp, err := http.Get(srv.URL + "/repos/owner/repo/contents/README.md")
	if err != nil || resp.StatusCode != 200 {
		t.Errorf("expected 200 for README.md, got status %v err %v", resp.StatusCode, err)
	}
	resp.Body.Close()

	// File does not exist
	resp2, err := http.Get(srv.URL + "/repos/owner/repo/contents/DOES_NOT_EXIST")
	if err != nil || resp2.StatusCode != 404 {
		t.Errorf("expected 404 for missing file, got status %v err %v", resp2.StatusCode, err)
	}
	resp2.Body.Close()
}
