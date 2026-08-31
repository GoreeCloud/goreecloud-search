package search

import (
	"context"
	"testing"
	"time"
)

type timestampProvider struct {
	name          string
	authoritative bool
	results       []Result
}

func (p timestampProvider) Name() string { return p.name }
func (p timestampProvider) Search(context.Context, string) ([]Result, error) {
	return append([]Result(nil), p.results...), nil
}
func (p timestampProvider) PublishedAtAuthoritative() bool { return p.authoritative }

type legacyTimestampProvider struct {
	name    string
	results []Result
}

func (p legacyTimestampProvider) Name() string { return p.name }
func (p legacyTimestampProvider) Search(context.Context, string) ([]Result, error) {
	return append([]Result(nil), p.results...), nil
}

func TestNormalizeAuthoritativePublishedAtRejectsUntrustedAndImplausibleValues(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	valid := now.Add(-2 * time.Hour)
	if got, trusted := normalizeAuthoritativePublishedAt(&valid, false, now); got != nil || trusted {
		t.Fatalf("untrusted timestamp must be stripped, got %#v trusted=%v", got, trusted)
	}

	future := now.Add(25 * time.Hour)
	if got, trusted := normalizeAuthoritativePublishedAt(&future, true, now); got != nil || trusted {
		t.Fatalf("implausibly future timestamp must be stripped, got %#v trusted=%v", got, trusted)
	}

	beforeUnix := time.Date(1960, 1, 1, 0, 0, 0, 0, time.UTC)
	if got, trusted := normalizeAuthoritativePublishedAt(&beforeUnix, true, now); got != nil || trusted {
		t.Fatalf("pre-Unix timestamp must be stripped, got %#v trusted=%v", got, trusted)
	}

	got, trusted := normalizeAuthoritativePublishedAt(&valid, true, now)
	if !trusted || got == nil || !got.Equal(valid) || got.Location() != time.UTC {
		t.Fatalf("authoritative timestamp should be retained in UTC, got %#v trusted=%v", got, trusted)
	}
}

func TestFreshnessScoreRequiresTemporalIntentOrNewsCategory(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	published := now.Add(-3 * time.Hour)
	result := Result{PublishedAt: &published, publishedAtTrusted: true}

	if got := freshnessScore(parseQueryIntent("goreecloud search"), CategoryGeneral, result, now); got != 0 {
		t.Fatalf("ordinary general search must not receive freshness bias, got %d", got)
	}
	if got := freshnessScore(parseQueryIntent("latest goreecloud search"), CategoryGeneral, result, now); got != maxFreshnessBonus {
		t.Fatalf("explicit temporal query should receive max fresh-result bonus, got %d", got)
	}
	newsBonus := freshnessScore(parseQueryIntent("goreecloud search"), CategoryNews, result, now)
	if newsBonus <= 0 || newsBonus >= maxFreshnessBonus {
		t.Fatalf("news category should receive positive but lower implicit freshness bonus, got %d", newsBonus)
	}
}

func TestFreshnessBonusDoesNotOverpowerClearRelevance(t *testing.T) {
	fresh := Result{
		Title:        "Unrelated infrastructure notice",
		URL:          "https://fresh.example/item",
		Provider:     "fresh",
		Score:        0,
		recencyBonus: maxFreshnessBonus,
	}
	strong := Result{
		Title:    "Latest GoreeCloud Search privacy update",
		URL:      "https://relevant.example/item",
		Provider: "relevant",
		Score:    0,
	}

	ranked := rankResults("latest goreecloud search privacy update", []Result{fresh, strong})
	if len(ranked) != 2 || ranked[0].URL != strong.URL {
		t.Fatalf("bounded freshness must not overpower a clearly stronger relevance match: %#v", ranked)
	}
}

func TestEngineRetainsOnlyProviderAuthoritativePublishedAt(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	trustedTime := now.Add(-2 * time.Hour)
	spoofedTime := now.Add(-30 * time.Minute)

	engine := NewEngine(time.Second,
		timestampProvider{
			name:          "trusted",
			authoritative: true,
			results: []Result{{
				Title: "Trusted publication time",
				URL: "https://trusted.example/item",
				PublishedAt: &trustedTime,
			}},
		},
		legacyTimestampProvider{
			name: "legacy",
			results: []Result{{
				Title: "Untrusted publication time",
				URL: "https://legacy.example/item",
				PublishedAt: &spoofedTime,
				PublishedAtSource: "spoofed-authority",
			}},
		},
	)

	response, err := engine.Search(context.Background(), "publication time")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(response.Results) != 2 {
		t.Fatalf("expected two results, got %#v", response.Results)
	}

	byProvider := map[string]Result{}
	for _, result := range response.Results {
		byProvider[result.Provider] = result
	}
	trusted := byProvider["trusted"]
	if trusted.PublishedAt == nil || !trusted.PublishedAt.Equal(trustedTime) || trusted.PublishedAtSource != "trusted" {
		t.Fatalf("trusted provider timestamp/provenance was not retained: %#v", trusted)
	}
	legacy := byProvider["legacy"]
	if legacy.PublishedAt != nil || legacy.PublishedAtSource != "" {
		t.Fatalf("legacy provider timestamp must be stripped: %#v", legacy)
	}
}

func TestEngineFreshnessRanksRecentAuthoritativeResultForLatestQuery(t *testing.T) {
	now := time.Now().UTC()
	freshTime := now.Add(-time.Hour)
	oldTime := now.Add(-60 * 24 * time.Hour)
	engine := NewEngine(time.Second,
		timestampProvider{name: "fresh", authoritative: true, results: []Result{{
			Title: "GoreeCloud Search update",
			URL: "https://fresh.example/update",
			Snippet: "Latest GoreeCloud Search update",
			PublishedAt: &freshTime,
		}}},
		timestampProvider{name: "old", authoritative: true, results: []Result{{
			Title: "GoreeCloud Search update",
			URL: "https://old.example/update",
			Snippet: "Latest GoreeCloud Search update",
			PublishedAt: &oldTime,
		}}},
	)

	response, err := engine.Search(context.Background(), "latest goreecloud search update")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(response.Results) != 2 || response.Results[0].Provider != "fresh" {
		t.Fatalf("fresh authoritative result should win equal-relevance latest query: %#v", response.Results)
	}
	if response.Results[0].Score-response.Results[1].Score <= 0 || response.Results[0].Score-response.Results[1].Score > maxFreshnessBonus {
		t.Fatalf("freshness delta must remain positive and bounded, got scores %d vs %d", response.Results[0].Score, response.Results[1].Score)
	}
}
