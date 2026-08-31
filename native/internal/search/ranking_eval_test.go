package search

import "testing"

type rankingExpectation struct {
	name       string
	query      string
	candidates []Result
	wantURL    string
}

func TestRankingEvaluationCorpus(t *testing.T) {
	cases := []rankingExpectation{
		{
			name:  "navigational domain beats provider score",
			query: "goreecloud.com",
			candidates: []Result{
				{Title: "GoreeCloud", URL: "https://goreecloud.com/", Provider: "low", Score: 1},
				{Title: "GoreeCloud community mirror", URL: "https://example.net/goreecloud", Provider: "high", Score: 9999},
			},
			wantURL: "https://goreecloud.com/",
		},
		{
			name:  "informational exact title beats generic coverage",
			query: "privacy shield architecture",
			candidates: []Result{
				{Title: "Privacy Shield architecture", URL: "https://docs.example/privacy-shield", Provider: "docs", Score: 1},
				{Title: "Architecture reference", URL: "https://example.net/reference", Snippet: "Privacy controls and shield design for architecture teams", Provider: "index", Score: 300},
			},
			wantURL: "https://docs.example/privacy-shield",
		},
		{
			name:  "explicit site phrase and file type cooperate",
			query: `site:docs.example "privacy shield" filetype:pdf`,
			candidates: []Result{
				{Title: "Privacy Shield", URL: "https://docs.example/privacy-shield.pdf", Provider: "docs", Score: 1},
				{Title: "Privacy Shield", URL: "https://other.example/privacy-shield.pdf", Provider: "other", Score: 300},
				{Title: "Privacy Shield", URL: "https://docs.example/privacy-shield", Provider: "docs-html", Score: 300},
			},
			wantURL: "https://docs.example/privacy-shield.pdf",
		},
		{
			name:  "bounded transposition tolerance rescues likely target",
			query: "goreecluod search",
			candidates: []Result{
				{Title: "GoreeCloud Search", URL: "https://goreecloud.com/search", Provider: "native", Score: 1},
				{Title: "Cloud Search", URL: "https://example.net/cloud-search", Provider: "index", Score: 300},
			},
			wantURL: "https://goreecloud.com/search",
		},
		{
			name:  "consensus cannot overpower clearly relevant result",
			query: "goreecloud security model",
			candidates: []Result{
				{Title: "GoreeCloud security model", URL: "https://security.example/model", Provider: "security", Score: 1},
				{Title: "Cloud security news", URL: "https://news.example/cloud", Provider: "one", Score: 300},
				{Title: "Cloud security news mirror", URL: "https://news.example/cloud", Provider: "two", Score: 300},
				{Title: "Cloud security news index", URL: "https://news.example/cloud", Provider: "three", Score: 300},
				{Title: "Cloud security news archive", URL: "https://news.example/cloud", Provider: "four", Score: 300},
			},
			wantURL: "https://security.example/model",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			ranked := rankResults(test.query, test.candidates)
			if len(ranked) == 0 {
				t.Fatal("expected ranked results")
			}
			if ranked[0].URL != test.wantURL {
				t.Fatalf("top URL = %q, want %q; ranked=%#v", ranked[0].URL, test.wantURL, ranked)
			}
		})
	}
}
