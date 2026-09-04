package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
	"github.com/GoreeCloud/goreecloud-search/native/internal/webui"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", webui.Homepage)
	mux.HandleFunc("GET /preferences", webui.Preferences)
	mux.HandleFunc("GET /search", resultsFixture)
	mux.HandleFunc("GET /assets/app.css", webui.Styles)
	mux.HandleFunc("GET /assets/home.css", webui.HomepageStyles)
	mux.HandleFunc("GET /assets/preferences.css", webui.PreferencesStyles)
	mux.HandleFunc("GET /assets/preferences.js", webui.PreferencesScript)
	mux.HandleFunc("GET /assets/results.css", webui.ResultsStyles)
	mux.HandleFunc("GET /assets/image-results.css", webui.ImageResultsStyles)
	mux.HandleFunc("GET /assets/results.js", webui.ResultsScript)
	mux.HandleFunc("GET /assets/categories.css", webui.CategoryStyles)
	mux.HandleFunc("GET /fixture/image/", fixtureImage)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})

	addr := os.Getenv("GOREECLOUD_SEARCH_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8091"
	}

	log.Printf("GoreeCloud Search native application acceptance fixture listening on %s", addr)
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
	case "images":
		webui.RenderResultsWithMedia(w, representativeImageResponse(), func(raw string) string { return raw })
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

func representativeImageResponse() searchcore.Response {
	return searchcore.Response{
		Query:    "macos golden gate",
		Category: searchcore.CategoryImages,
		Results: []searchcore.Result{
			{
				Title:    "Golden Gate morning",
				URL:      "https://photos.goreecloud.example/golden-gate-morning",
				Provider: "Image Catalog",
				Media: &searchcore.Media{
					Kind:         searchcore.MediaKindImage,
					ThumbnailURL: "/fixture/image/1.svg",
					ContentURL:   "/fixture/image/1.svg",
					Width:        1600,
					Height:       1000,
					Alt:          "Golden Gate bridge in morning light",
				},
			},
			{
				Title:    "Golden Gate fog",
				URL:      "https://photos.goreecloud.example/golden-gate-fog",
				Provider: "Image Catalog",
				Media: &searchcore.Media{
					Kind:         searchcore.MediaKindImage,
					ThumbnailURL: "/fixture/image/2.svg",
					ContentURL:   "/fixture/image/2.svg",
					Width:        1400,
					Height:       1050,
					Alt:          "Golden Gate bridge under fog",
				},
			},
			{
				Title:    "Golden Gate shoreline",
				URL:      "https://photos.goreecloud.example/golden-gate-shoreline",
				Provider: "Archive Images",
				Media: &searchcore.Media{
					Kind:         searchcore.MediaKindImage,
					ThumbnailURL: "/fixture/image/3.svg",
					ContentURL:   "/fixture/image/3.svg",
					Width:        1800,
					Height:       1200,
					Alt:          "Golden Gate bridge from the shoreline",
				},
			},
		},
		Providers: []searchcore.ProviderStatus{
			{Name: "Archive Images", State: searchcore.ProviderStateAvailable, Count: 1},
			{Name: "Image Catalog", State: searchcore.ProviderStateAvailable, Count: 2},
		},
	}
}

func fixtureImage(w http.ResponseWriter, r *http.Request) {
	label := "01"
	switch r.URL.Path {
	case "/fixture/image/2.svg":
		label = "02"
	case "/fixture/image/3.svg":
		label = "03"
	}
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 520"><rect width="800" height="520" fill="#dfe5ee"/><path d="M0 390 C180 320 300 370 440 310 C570 255 675 275 800 230 L800 520 L0 520 Z" fill="#aab7c8"/><path d="M110 345 L690 345" stroke="#7a8492" stroke-width="12"/><path d="M235 190 L235 395 M565 155 L565 395" stroke="#687487" stroke-width="22"/><path d="M235 205 C330 280 470 280 565 170" fill="none" stroke="#687487" stroke-width="8"/><text x="40" y="72" font-family="system-ui,sans-serif" font-size="44" fill="#2c3440">GoreeCloud image %s</text></svg>`, label)
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
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; img-src 'self' data:; script-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'")
		next.ServeHTTP(w, r)
	})
}
