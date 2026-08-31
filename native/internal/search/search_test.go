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

type blockingProvider struct {
	name    string
	release <-chan struct{}
}

func (p blockingProvider) Name() string { return p.name }
func (p blockingProvider) Search(context.Context, string) ([]Result, error) {
	<-p.release // Intentionally ignores context to exercise the engine boundary.
	return []Result{{Title: "Late", URL: "https://late.example/", Score: 100}}, nil
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

func TestValidateCategory(t *testing.T) {
	for _, category := range SupportedCategories {
		got, err := ValidateCategory("  " + strings.ToUpper(category) + "  ")
		if err != nil || got != category {
			t.Fatalf("category %q normalized to %q with %v", category, got, err)
		}
	}
	got, err := ValidateCategory("")
	if err != nil || got != CategoryGeneral {
		t.Fatalf("blank category = %q, %v", got, err)
	}
	if _, err := ValidateCategory("shopping"); err == nil {
		t.Fatal("expected unknown category rejection")
	}

	engine := NewEngine(time.Second)
	if !engine.SupportsCategory(CategoryGeneral) {
		t.Fatal("general category must be implemented")
	}
	for _, category := range []string{CategoryImages, CategoryVideos, CategoryNews, CategoryFiles} {
		if engine.SupportsCategory(category) {
			t.Fatalf("category %q must remain fail-closed until adapters are implemented", category)
		}
	}
}

func TestEngineAggregatesDeduplicatesAndDegrades(t *testing.T) {
	engine := NewEngine(time.Second,
		fakeProvider{name: "one", results: []Result{{Title: "A", URL: "https://example.com/a#fragment", Score: 4}, {Title: "B", URL: "javascript:alert(1)", Score: 10}, {Title: "Credential URL", URL: "https://user:secret@example.net/private", Score: 20}}},
		fakeProvider{name: "two", results: []Result{{Title: "Duplicate", URL: "https://example.com/a", Score: 9}, {Title: "C", URL: "https://example.org/c", Score: 5}}},
		fakeProvider{name: "three", err: errors.New("provider unavailable: credential=do-not-expose")},
	)

	response, err := engine.Search(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if response.Category != CategoryGeneral {
		t.Fatalf("expected general response category, got %q", response.Category)
	}
	if !response.Degraded {
		t.Fatal("expected degraded response when one provider fails")
	}
	if len(response.Results) != 2 {
		t.Fatalf("expected 2 sanitized unique results, got %d", len(response.Results))
	}
	if response.Results[0].URL != "https://example.com/a" || response.Results[0].Provider != "two" {
		t.Fatalf("expected query-aware duplicate cluster to retain deterministic representative: %#v", response.Results)
	}
	if response.Results[0].SourceCount != 2 || len(response.Results[0].Sources) != 2 {
		t.Fatalf("expected duplicate-provider consensus evidence: %#v", response.Results[0])
	}
	if response.Results[1].URL != "https://example.org/c" || response.Results[0].Score <= response.Results[1].Score {
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

func TestEngineDeadlineBoundsProviderIgnoringContext(t *testing.T) {
	release := make(chan struct{})
	defer close(release)

	engine := NewEngine(30*time.Millisecond,
		fakeProvider{name: "healthy", results: []Result{{Title: "Healthy", URL: "https://healthy.example/", Score: 1}}},
		blockingProvider{name: "stuck", release: release},
	)

	started := time.Now()
	response, err := engine.Search(context.Background(), "bounded")
	elapsed := time.Since(started)
	if err != nil {
		t.Fatal(err)
	}
	if elapsed > time.Second {
		t.Fatalf("search exceeded bounded request deadline: %s", elapsed)
	}
	if !response.Degraded {
		t.Fatal("expected degraded response for provider that ignored context")
	}
	if len(response.Results) != 1 || response.Results[0].Provider != "healthy" {
		t.Fatalf("expected healthy provider results to survive timeout: %#v", response.Results)
	}
	if len(response.Providers) != 2 {
		t.Fatalf("expected status for both providers: %#v", response.Providers)
	}
	var stuck ProviderStatus
	for _, status := range response.Providers {
		if status.Name == "stuck" {
			stuck = status
		}
	}
	if stuck.State != ProviderStateUnavailable || stuck.Code != ProviderCodeTimeout || stuck.Count != 0 {
		t.Fatalf("unexpected timeout status for stuck provider: %#v", stuck)
	}
}

func TestInvalidProviderIdentityIsNotAdvertisedOrExecuted(t *testing.T) {
	invalid := fakeProvider{name: "bad\nprovider", results: []Result{{Title: "Bad", URL: "https://bad.example/", Score: 99}}}
	blank := fakeProvider{name: "   ", results: []Result{{Title: "Blank", URL: "https://blank.example/", Score: 99}}}
	valid := fakeProvider{name: "  good-provider  ", results: []Result{{Title: "Good", URL: "https://good.example/", Score: 1}}}
	engine := NewEngine(time.Second, invalid, blank, valid)

	definitions := engine.ProviderDefinitions()
	if len(definitions) != 1 || definitions[0].Name != "good-provider" {
		t.Fatalf("unexpected provider definitions: %#v", definitions)
	}
	response, err := engine.Search(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Providers) != 1 || response.Providers[0].Name != "good-provider" {
		t.Fatalf("invalid provider identities must not enter runtime evidence: %#v", response.Providers)
	}
	if len(response.Results) != 1 || response.Results[0].Provider != "good-provider" {
		t.Fatalf("invalid provider identities must not execute: %#v", response.Results)
	}
}

func TestProviderNameBounds(t *testing.T) {
	if _, ok := normalizeProviderName(strings.Repeat("p", MaxProviderNameRunes)); !ok {
		t.Fatal("provider name at maximum length should be valid")
	}
	if _, ok := normalizeProviderName(strings.Repeat("p", MaxProviderNameRunes+1)); ok {
		t.Fatal("oversized provider name should be rejected")
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
