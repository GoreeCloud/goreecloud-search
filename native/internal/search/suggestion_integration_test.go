package search

import (
	"context"
	"testing"
	"time"
)

type recordingSuggestionProvider struct {
	name    string
	results []Result
	queries chan<- string
}

func (p recordingSuggestionProvider) Name() string { return p.name }
func (p recordingSuggestionProvider) Search(_ context.Context, query string) ([]Result, error) {
	p.queries <- query
	return append([]Result(nil), p.results...), nil
}

func TestEngineSuggestionDoesNotRewriteProviderQuery(t *testing.T) {
	queries := make(chan string, 2)
	engine := NewEngine(time.Second,
		recordingSuggestionProvider{
			name:    "one",
			queries: queries,
			results: []Result{{Title: "GoreeCloud Search", URL: "https://one.example/search"}},
		},
		recordingSuggestionProvider{
			name:    "two",
			queries: queries,
			results: []Result{{Title: "GoreeCloud private search", URL: "https://two.example/search"}},
		},
	)

	const submitted = "goreecluod search"
	response, err := engine.Search(context.Background(), submitted)
	if err != nil {
		t.Fatal(err)
	}
	if response.Query != submitted {
		t.Fatalf("response query = %q, want original %q", response.Query, submitted)
	}
	if response.SuggestedQuery != "goreecloud search" {
		t.Fatalf("suggested query = %q, want explicit correction", response.SuggestedQuery)
	}

	for i := 0; i < 2; i++ {
		if got := <-queries; got != submitted {
			t.Fatalf("provider received rewritten query %q, want original %q", got, submitted)
		}
	}
}
