package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
	"github.com/GoreeCloud/goreecloud-search/native/internal/webui"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/search", http.StatusFound)
	})
	mux.HandleFunc("GET /search", resultsFixture)
	mux.HandleFunc("GET /assets/app.css", webui.Styles)
	mux.HandleFunc("GET /assets/results.css", webui.ResultsStyles)
	mux.HandleFunc("GET /assets/categories.css", webui.CategoryStyles)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})

	addr := os.Getenv("GOREECLOUD_SEARCH_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8091"
	}

	log.Printf("GoreeCloud Search native results acceptance fixture listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, acceptanceHeaders(mux)))
}

func resultsFixture(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Query().Get("case") {
	case "empty":
		webui.RenderResults(w, searchcore.Response{
			Query:    "unusually specific private search phrase",
			Category: searchcore.CategoryGeneral,
			Results:  []searchcore.Result{},
			Providers: []searchcore.ProviderStatus{
				{Name: "GoreeCloud Index", State: searchcore.ProviderStateAvailable, Count: 0},
				{Name: "Research Catalog", State: searchcore.ProviderStateAvailable, Count: 0},
			},
		})
	case "error":
		webui.RenderSearchError(w, "goreecloud search", errors.New("acceptance fixture error"))
	case "suggestion":
		response := representativeResponse()
		response.Query = "goreecluod search privacy"
		response.SuggestedQuery = "goreecloud search privacy"
		webui.RenderResults(w, response)
	default:
		webui.RenderResults(w, representativeResponse())
	}
}

func representativeResponse() searchcore.Response {
	return searchcore.Response{
		Query:    "goreecloud search privacy",
		Category: searchcore.CategoryGeneral,
		Degraded: true,
		Results: []searchcore.Result{
			{
				Title:             "GoreeCloud Search — Private Search Architecture",
				URL:               "https://docs.goreecloud.example/search/private-search-architecture",
				Snippet:           "How GoreeCloud Search keeps request-local relevance, source agreement, privacy boundaries, and deterministic ranking separate from behavioral profiling.",
				Provider:          "GoreeCloud Index",
				SourceCount:       3,
				Sources:           []string{"Docs Mirror", "GoreeCloud Index", "Research Catalog"},
				PublishedAt:       fixedPublishedAt(2026, time.August, 31, 14, 30),
				PublishedAtSource: "GoreeCloud Index",
			},
			{
				Title:             "Search ranking without click-history profiling",
				URL:               "https://research.goreecloud.example/search/ranking-without-profiling",
				Snippet:           "A relevance model that favors query-title alignment, bounded source evidence, transparent consensus, and source diversity without personal behavioral tracking.",
				Provider:          "Research Catalog",
				PublishedAt:       fixedPublishedAt(2026, time.August, 28, 9, 15),
				PublishedAtSource: "Research Catalog",
			},
			{
				Title:    "Understanding source agreement in GoreeCloud Search",
				URL:      "https://docs.goreecloud.example/search/source-agreement",
				Snippet:  "Source agreement is supporting evidence rather than universal authority, so multiple weak duplicates cannot automatically outrank a clearly more relevant result.",
				Provider: "Docs Mirror",
			},
			{
				Title:    "Privacy Shield integration for search requests",
				URL:      "https://privacy.goreecloud.example/search/request-boundaries",
				Snippet:  "Search request handling is designed around data minimization, explicit product boundaries, and no advertising-driven ranking signals.",
				Provider: "GoreeCloud Index",
			},
			{
				Title:    "Search operators: site, file type, and quoted phrases",
				URL:      "https://help.goreecloud.example/search/operators/site-filetype-phrases",
				Snippet:  "Use explicit operators to express domain, file-type, and phrase intent while the native reranker preserves the query submitted to configured providers.",
				Provider: "Docs Mirror",
			},
			{
				Title:    "Why GoreeCloud Search diversifies the first result viewport",
				URL:      "https://research.goreecloud.example/search/source-diversity",
				Snippet:  "First-page diversity prevents one hostname from crowding out useful alternatives when the query is not explicitly directed to that domain.",
				Provider: "Research Catalog",
			},
			{
				Title:    "Native Search provider health and bounded processing",
				URL:      "https://status.goreecloud.example/search/provider-health-and-limits",
				Snippet:  "Provider health remains separate from result relevance, and Search discloses when its bounded per-provider processing ceiling is applied.",
				Provider: "GoreeCloud Index",
			},
		},
		Providers: []searchcore.ProviderStatus{
			{Name: "GoreeCloud Index", State: searchcore.ProviderStateAvailable, Count: 7},
			{Name: "Docs Mirror", State: searchcore.ProviderStateAvailable, Count: searchcore.MaxResultsPerProvider, Truncated: true},
			{Name: "Research Catalog", State: searchcore.ProviderStateUnavailable, Code: searchcore.ProviderCodeTimeout, Count: 0},
		},
	}
}

func fixedPublishedAt(year int, month time.Month, day, hour, minute int) *time.Time {
	value := time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
	return &value
}

func acceptanceHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; img-src 'self' data:; form-action 'self'; frame-ancestors 'none'; base-uri 'none'")
		next.ServeHTTP(w, r)
	})
}
