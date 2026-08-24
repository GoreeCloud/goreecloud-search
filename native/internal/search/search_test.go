package search

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeProvider struct {
	name    string
	results []Result
	err     error
}

func (p fakeProvider) Name() string { return p.name }
func (p fakeProvider) Search(context.Context, string) ([]Result, error) {
	return append([]Result(nil), p.results...), p.err
}

func TestValidateQuery(t *testing.T) {
	if _, err := ValidateQuery("   "); err == nil {
		t.Fatal("expected blank query rejection")
	}
	if _, err := ValidateQuery(strings.Repeat("a", MaxQueryRunes+1)); err == nil {
		t.Fatal("expected oversized query rejection")
	}
	query, err := ValidateQuery("  goreecloud search  ")
	if err != nil || query != "goreecloud search" {
		t.Fatalf("unexpected normalization: %q %v", query, err)
	}
}

func TestEngineAggregatesDeduplicatesAndDegrades(t *testing.T) {
	engine := NewEngine(time.Second,
		fakeProvider{name: "one", results: []Result{{Title: "A", URL: "https://example.com/a#fragment", Score: 4}, {Title: "B", URL: "javascript:alert(1)", Score: 10}}},
		fakeProvider{name: "two", results: []Result{{Title: "Duplicate", URL: "https://example.com/a", Score: 9}, {Title: "C", URL: "https://example.org/c", Score: 5}}},
		fakeProvider{name: "three", err: errors.New("provider unavailable: credential=do-not-expose")},
	)

	response, err := engine.Search(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if !response.Degraded {
		t.Fatal("expected degraded response when one provider fails")
	}
	if len(response.Results) != 2 {
		t.Fatalf("expected 2 sanitized unique results, got %d", len(response.Results))
	}
	if response.Results[0].URL != "https://example.com/a" || response.Results[0].Score != 9 || response.Results[0].Provider != "two" {
		t.Fatalf("expected highest-scoring duplicate to win deterministically: %#v", response.Results)
	}
	if response.Results[1].URL != "https://example.org/c" || response.Results[1].Score != 5 {
		t.Fatalf("unexpected deterministic ranking: %#v", response.Results)
	}
	if len(response.Providers) != 3 {
		t.Fatalf("expected provider evidence, got %#v", response.Providers)
	}
	var failed ProviderStatus
	found := false
	for _, status := range response.Providers {
		if status.Name == "three" {
			failed = status
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing failed provider evidence: %#v", response.Providers)
	}
	if failed.State != ProviderStateUnavailable || failed.Code != ProviderCodeUnavailable {
		t.Fatalf("unexpected sanitized provider failure: %#v", failed)
	}
	if failed.Count != 0 {
		t.Fatalf("failed provider must not report discarded results: %#v", failed)
	}
}

func TestResultTieBreakIsStable(t *testing.T) {
	one := Result{Title: "B", URL: "https://example.com/a", Provider: "zeta", Score: 5}
	two := Result{Title: "A", URL: "https://example.com/a", Provider: "alpha", Score: 5}
	if !resultBetterThan(two, one) {
		t.Fatal("expected provider-name tie break to be deterministic")
	}
}

func TestProviderFailureClassification(t *testing.T) {
	if code := classifyProviderFailure(context.DeadlineExceeded); code != ProviderCodeTimeout {
		t.Fatalf("expected timeout code, got %q", code)
	}
	if code := classifyProviderFailure(errors.New("token=secret")); code != ProviderCodeUnavailable {
		t.Fatalf("expected bounded unavailable code, got %q", code)
	}
}
