package search

import (
	"testing"
	"time"
)

func TestFreshnessDoesNotActivateForLiteralOrNounCurrent(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	published := now.Add(-time.Hour)
	result := Result{PublishedAt: &published, publishedAtTrusted: true}

	for _, query := range []string{`"latest goreecloud" search`, "electric current"} {
		if got := freshnessScore(parseQueryIntent(query), CategoryGeneral, result, now); got != 0 {
			t.Fatalf("query %q should remain lexical rather than freshness intent, got bonus %d", query, got)
		}
	}
	if got := freshnessScore(parseQueryIntent("current weather"), CategoryGeneral, result, now); got != maxFreshnessBonus {
		t.Fatalf("leading current should activate freshness, got bonus %d", got)
	}
}

func TestTemporalModifierDoesNotDiluteSubjectRelevance(t *testing.T) {
	intent := parseQueryIntent("latest goreecloud search privacy")
	if intent.normalized != "goreecloud search privacy" {
		t.Fatalf("temporal modifier leaked into lexical query: %q", intent.normalized)
	}

	subject := Result{Title: "GoreeCloud Search privacy", URL: "https://subject.example/", Score: 1}
	generic := Result{Title: "Latest cloud privacy news", URL: "https://generic.example/", Score: 300}
	if relevanceScoreIntent(intent, subject) <= relevanceScoreIntent(intent, generic) {
		t.Fatalf("subject relevance should outrank a generic title that merely repeats the temporal modifier")
	}
}

func TestTemporalOnlyQueryCanUseFreshnessWithoutSyntheticLexicalTerms(t *testing.T) {
	intent := parseQueryIntent("latest")
	if !intent.freshnessRequested || intent.normalized != "" || len(intent.tokens) != 0 {
		t.Fatalf("latest-only intent should be temporal without lexical terms: %#v", intent)
	}

	fresh := Result{Title: "Fresh item", URL: "https://fresh.example/", recencyBonus: maxFreshnessBonus}
	old := Result{Title: "Older item", URL: "https://old.example/"}
	ranked := rankResults("latest", []Result{old, fresh})
	if len(ranked) != 2 || ranked[0].URL != fresh.URL {
		t.Fatalf("freshness should order a temporal-only query without synthetic text relevance: %#v", ranked)
	}
}
