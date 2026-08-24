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
		fakeProvider{name: "three", err: errors.New("provider unavailable")},
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
	if response.Results[0].URL != "https://example.org/c" || response.Results[1].URL != "https://example.com/a" {
		t.Fatalf("unexpected deterministic ranking: %#v", response.Results)
	}
	if len(response.Providers) != 3 {
		t.Fatalf("expected provider evidence, got %#v", response.Providers)
	}
}
