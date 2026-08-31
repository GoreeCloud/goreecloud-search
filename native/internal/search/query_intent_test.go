package search

import "testing"

func TestParseQueryIntentSeparatesOperators(t *testing.T) {
	intent := parseQueryIntent(`site:docs.example.com "privacy shield" filetype:pdf goreecloud`)
	if intent.normalized != "privacy shield goreecloud" {
		t.Fatalf("normalized=%q", intent.normalized)
	}
	if len(intent.siteHosts) != 1 || intent.siteHosts[0] != "docs.example.com" {
		t.Fatalf("siteHosts=%v", intent.siteHosts)
	}
	if len(intent.fileTypes) != 1 || intent.fileTypes[0] != "pdf" {
		t.Fatalf("fileTypes=%v", intent.fileTypes)
	}
	if len(intent.phrases) != 1 || intent.phrases[0] != "privacy shield" {
		t.Fatalf("phrases=%v", intent.phrases)
	}
}

func TestQueryIntentRejectsVersionAsDomain(t *testing.T) {
	if queryTargetsDomain("goreecloud 1.5.0 release") {
		t.Fatal("version token should not disable hostname diversity")
	}
	if !queryTargetsDomain("goreecloud.com") {
		t.Fatal("domain target should be recognized")
	}
}

func TestSiteIntentStronglyPrefersMatchingHost(t *testing.T) {
	matching := Result{Title: "GoreeCloud Search docs", URL: "https://docs.example.com/search", Provider: "a", Score: 1}
	wrong := Result{Title: "GoreeCloud Search docs", URL: "https://other.example.net/search", Provider: "b", Score: 9999}
	if relevanceScore("site:example.com goreecloud search", matching) <= relevanceScore("site:example.com goreecloud search", wrong) {
		t.Fatal("site intent should outrank a mismatched host despite provider score")
	}
}

func TestQuotedPhraseGetsOrderingBoost(t *testing.T) {
	phrase := Result{Title: "GoreeCloud Privacy Shield guide", URL: "https://a.example/guide", Score: 1}
	separated := Result{Title: "Privacy controls for GoreeCloud Shield users", URL: "https://b.example/guide", Score: 1}
	if relevanceScore(`"privacy shield" goreecloud`, phrase) <= relevanceScore(`"privacy shield" goreecloud`, separated) {
		t.Fatal("contiguous quoted phrase should receive a ranking boost")
	}
}

func TestBoundedTypoToleranceRecognizesTransposition(t *testing.T) {
	correct := Result{Title: "GoreeCloud Search", URL: "https://goreecloud.com/search", Score: 1}
	unrelated := Result{Title: "Cloud search index", URL: "https://example.net/search", Score: 300}
	if relevanceScore("goreecluod search", correct) <= relevanceScore("goreecluod search", unrelated) {
		t.Fatal("single transposition should support the clearly relevant title")
	}
}

func TestShortTokensDoNotUseFuzzyMatching(t *testing.T) {
	if fuzzyTokenCoverageScore([]string{"cat"}, "car", 700) != 0 {
		t.Fatal("short tokens should not receive fuzzy relevance")
	}
}

func TestFileTypeIntentPrefersRequestedExtension(t *testing.T) {
	pdf := Result{Title: "GoreeCloud architecture", URL: "https://docs.example/architecture.pdf", Score: 1}
	html := Result{Title: "GoreeCloud architecture", URL: "https://docs.example/architecture", Score: 300}
	if relevanceScore("filetype:pdf goreecloud architecture", pdf) <= relevanceScore("filetype:pdf goreecloud architecture", html) {
		t.Fatal("filetype intent should prefer the requested extension")
	}
}
