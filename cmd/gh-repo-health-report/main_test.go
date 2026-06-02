package main

import (
	"testing"

	"github.com/gclhub/gh-repo-health-report/internal/checks"
)

func TestShouldFailSpecificCheckOverridesSkippedProfile(t *testing.T) {
	result := &checks.Result{
		HasCodeowners: false,
		SkippedChecks: []checks.SkippedCheck{
			{Name: checks.CheckMissingCodeowners, Reason: "Ignored by profile: open-source"},
		},
	}

	if !shouldFail(result, []string{checks.CheckMissingCodeowners}, 0, 0) {
		t.Fatalf("expected specific fail-on check to fail even when skipped by profile")
	}
}

func TestShouldFailAnyRespectsProfileFailures(t *testing.T) {
	result := &checks.Result{
		HasCodeowners: false,
		SkippedChecks: []checks.SkippedCheck{
			{Name: checks.CheckMissingCodeowners, Reason: "Ignored by profile: open-source"},
		},
	}

	if shouldFail(result, []string{"any"}, 0, 0) {
		t.Fatalf("expected fail-on any to ignore profile-skipped failures")
	}
}
