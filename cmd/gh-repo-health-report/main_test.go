package main

import "testing"

func TestShouldFailSpecificCheckUsesUnprofiledFailures(t *testing.T) {
	filteredFailed := []string{"missing-readme"}
	unprofiledFailed := []string{"missing-readme", "missing-codeowners"}

	if !shouldFail(filteredFailed, unprofiledFailed, []string{"missing-codeowners"}) {
		t.Fatal("expected specific fail-on check to use unprofiled failures")
	}
}

func TestShouldFailAnyUsesProfileFilteredFailures(t *testing.T) {
	filteredFailed := []string{}
	unprofiledFailed := []string{"missing-codeowners"}

	if shouldFail(filteredFailed, unprofiledFailed, []string{"any"}) {
		t.Fatal("expected fail-on any to use profile-filtered failures")
	}
}
