package search

import (
	"reflect"
	"testing"
)

func TestRankingPrefersQueryRelevanceOverUnboundedProviderScore(t *testing.T) {
	results := rankResults("goreecloud search", []Result{
		{Title: "Unrelated result", URL: "https://noise.example/", Provider: "loud", Score: 1000000, Snippet: "Nothing about the requested product."},
		{Title: "GoreeCloud Search", URL: "https://search.goreecloud.example/", Provider: "quiet", Score: 1, Snippet: "Private metasearch."},
	})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %#v", results)
	}
	if results[0].URL != "https://search.goreecloud.example/" {
		t.Fatalf("query relevance must outrank arbitrary provider score: %#v", results)
	}
	if results[1].Score > maxProviderScoreSignal {
		t.Fatalf("unrelated provider score must remain bounded, got %d", results[1].Score)
	}
}

func TestRankingWeightsTitleAboveSnippetAndURL(t *testing.T) {
	results := rankResults("private search", []Result{
		{Title: "Private Search", URL: "https://one.example/", Provider: "one", Score: 1},
		{Title: "Other", URL: "https://two.example/private/search", Provider: "two", Score: 1, Snippet: "Private search"},
	})
	if results[0].Provider != "one" {
		t.Fatalf("title relevance should be strongest local signal: %#v", results)
	}
}

func TestRankingPreservesProviderConsensus(t *testing.T) {
	results := rankResults("goreecloud", []Result{
		{Title: "GoreeCloud", URL: "https://example.com/product", Provider: "bravo", Score: 2},
		{Title: "GoreeCloud", URL: "https://example.com/product", Provider: "alpha", Score: 1},
	})
	if len(results) != 1 {
		t.Fatalf("expected duplicate URL cluster, got %#v", results)
	}
	if results[0].SourceCount != 2 {
		t.Fatalf("expected consensus count, got %#v", results[0])
	}
	if !reflect.DeepEqual(results[0].Sources, []string{"alpha", "bravo"}) {
		t.Fatalf("expected deterministic source evidence, got %#v", results[0].Sources)
	}
	if results[0].Score < consensusBonus(2) {
		t.Fatalf("expected consensus contribution, got %d", results[0].Score)
	}
}

func TestRankingRepresentativeUsesMoreRelevantDuplicateContent(t *testing.T) {
	results := rankResults("goreecloud search", []Result{
		{Title: "Generic", URL: "https://example.com/page", Provider: "high-raw", Score: 250},
		{Title: "GoreeCloud Search", URL: "https://example.com/page", Provider: "low-raw", Score: 1},
	})
	if len(results) != 1 || results[0].Provider != "low-raw" || results[0].Title != "GoreeCloud Search" {
		t.Fatalf("best duplicate presentation should be query-relevant: %#v", results)
	}
}

func TestRankingDiversifiesFirstViewportByHostname(t *testing.T) {
	results := rankResults("goreecloud", []Result{
		{Title: "GoreeCloud one", URL: "https://same.example/1", Provider: "p", Score: 20},
		{Title: "GoreeCloud two", URL: "https://same.example/2", Provider: "p", Score: 19},
		{Title: "GoreeCloud three", URL: "https://same.example/3", Provider: "p", Score: 18},
		{Title: "GoreeCloud alternative", URL: "https://other.example/", Provider: "q", Score: 1},
	})
	if len(results) != 4 {
		t.Fatalf("expected four results: %#v", results)
	}
	if resultHost(results[2]) != "other.example" {
		t.Fatalf("third slot should surface another relevant hostname: %#v", results)
	}
}

func TestRankingKeepsPureOrderForExplicitDomainQuery(t *testing.T) {
	ranked := []Result{
		{URL: "https://same.example/1", Score: 10},
		{URL: "https://same.example/2", Score: 9},
		{URL: "https://same.example/3", Score: 8},
		{URL: "https://other.example/", Score: 7},
	}
	results := diversifyTopResults("site:same.example goreecloud", ranked)
	if results[2].URL != "https://same.example/3" {
		t.Fatalf("site-directed query should not be diversified: %#v", results)
	}
}

func TestNormalizeSearchTextIsUnicodeAwareAndDeterministic(t *testing.T) {
	if got := normalizeSearchText("  GORÉECloud—Search / 2026 "); got != "goréecloud search 2026" {
		t.Fatalf("unexpected normalization %q", got)
	}
}
