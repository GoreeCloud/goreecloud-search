package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

func TestRenderResultsEscapesProviderContentAndHidesInternalScore(t *testing.T) {
	recorder := httptest.NewRecorder()
	RenderResults(recorder, searchcore.Response{
		Query: "goreecloud <script>alert(1)</script>",
		Results: []searchcore.Result{{
			Title:       "<img src=x onerror=alert(1)>",
			URL:         "https://example.com/path?q=one&two=2",
			Snippet:     "<b>provider content</b>",
			Provider:    "example <script>",
			Score:       42,
			SourceCount: 2,
			Sources:     []string{"example <script>", "other"},
		}},
		Providers: []searchcore.ProviderStatus{{
			Name: "example", State: searchcore.ProviderStateAvailable, Count: 512, Truncated: true,
		}},
	})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, unsafe := range []string{"<script>alert(1)</script>", "<img src=x onerror=alert(1)>", "<b>provider content</b>"} {
		if strings.Contains(body, unsafe) {
			t.Fatalf("rendered unsafe provider HTML %q", unsafe)
		}
	}
	for _, expected := range []string{"GoreeCloud Search", "result-card", "example", "2 sources agree", "Source agreement", "https://example.com/path", "Click history is not used", "512 results · limit applied"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("rendered body missing %q", expected)
		}
	}
	if strings.Contains(body, "Score 42") {
		t.Fatal("results UI exposed internal ranking score")
	}
}

func TestRenderSearchErrorDoesNotExposeErrorDetails(t *testing.T) {
	recorder := httptest.NewRecorder()
	RenderSearchError(recorder, "query", errPrivate("provider token=secret"))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	body := recorder.Body.String()
	if strings.Contains(body, "token=secret") {
		t.Fatal("search error exposed internal error details")
	}
	if !strings.Contains(body, "Search could not run") {
		t.Fatal("search error did not render bounded user-facing message")
	}
}

func TestResultsStylesAreServedAsCSS(t *testing.T) {
	recorder := httptest.NewRecorder()
	ResultsStyles(recorder, httptest.NewRequest(http.MethodGet, "/assets/results.css", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/css; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	body := recorder.Body.String()
	for _, expected := range []string{".result-card", ".results-layout", ".results-sidebar", "prefers-reduced-motion", "forced-colors"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("results stylesheet missing %q", expected)
		}
	}
}

func TestHomepageStylesAreServedAsCSS(t *testing.T) {
	recorder := httptest.NewRecorder()
	HomepageStyles(recorder, httptest.NewRequest(http.MethodGet, "/assets/home.css", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/css; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	body := recorder.Body.String()
	for _, expected := range []string{".trust-strip", ".home-insights", ":has(input:checked)", "prefers-reduced-motion", "forced-colors"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("homepage stylesheet missing %q", expected)
		}
	}
}

func TestHomepageKeepsTrustAndCategoryControlsScriptFree(t *testing.T) {
	recorder := httptest.NewRecorder()
	Homepage(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	body := recorder.Body.String()

	for _, expected := range []string{"/assets/home.css", "No advertising", "No click tracking", "Provider transparency", "name=\"category\"", "value=\"images\"", "value=\"news\""} {
		if !strings.Contains(body, expected) {
			t.Fatalf("homepage missing %q", expected)
		}
	}
	if strings.Contains(strings.ToLower(body), "<script") {
		t.Fatal("homepage unexpectedly executes client-side script")
	}
}

func TestPreferencesScriptIsLocalOnlyAndSchemaBound(t *testing.T) {
	recorder := httptest.NewRecorder()
	PreferencesScript(recorder, httptest.NewRequest(http.MethodGet, "/assets/preferences.js", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/javascript; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	body := recorder.Body.String()
	for _, expected := range []string{
		"goreecloud.search.preferences.v1",
		"schema_version",
		"localStorage",
		"privacy.recent_queries",
		"search.autocomplete",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("Preferences script missing %q", expected)
		}
	}
	for _, forbidden := range []string{"fetch(", "XMLHttpRequest", "sendBeacon", "WebSocket"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("Preferences script unexpectedly contains network primitive %q", forbidden)
		}
	}
}

func TestPreferencesPageWiresLocalControlsAndPortability(t *testing.T) {
	recorder := httptest.NewRecorder()
	Preferences(recorder, httptest.NewRequest(http.MethodGet, "/preferences", nil))
	body := recorder.Body.String()

	for _, expected := range []string{
		"/assets/preferences.js",
		"data-settings-filter",
		"data-export-preferences",
		"data-import-preferences",
		"data-reset-preferences",
		"data-preference=\"search.autocomplete\"",
		"data-preference=\"privacy.recent_queries\"",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("Preferences page missing %q", expected)
		}
	}
}

type errPrivate string

func (e errPrivate) Error() string { return string(e) }
