package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

func TestRenderResultsEscapesProviderContentAndHidesInternalScore(t *testing.T) {
	recorder := httptest.NewRecorder()
	published := time.Date(2026, time.August, 31, 14, 30, 0, 0, time.UTC)
	RenderResults(recorder, searchcore.Response{
		Query: "goreecloud <script>alert(1)</script>",
		Results: []searchcore.Result{{
			Title:             "<img src=x onerror=alert(1)>",
			URL:               "https://example.com/path?q=one&two=2",
			Snippet:           "<b>provider content</b>",
			Provider:          "example <script>",
			Score:             42,
			SourceCount:       2,
			Sources:           []string{"example <script>", "other"},
			PublishedAt:       &published,
			PublishedAtSource: "example <script>",
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
	for _, expected := range []string{
		"GoreeCloud Search",
		`data-glaze-version="1.1"`,
		`data-glaze-density-profile="comfortable"`,
		`/assets/appearance.js`,
		"result-card",
		"example",
		"2 sources agree",
		"Source agreement",
		"https://example.com/path",
		"trustworthy freshness when requested",
		"Click history is not used",
		"Published Aug 31, 2026",
		`datetime="2026-08-31T14:30:00Z"`,
		"512 results · limit applied",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("rendered body missing %q", expected)
		}
	}
	if strings.Contains(body, "Score 42") {
		t.Fatal("results UI exposed internal ranking score")
	}
	if strings.Contains(body, "PublishedAtSource") {
		t.Fatal("results UI exposed internal publication metadata field name")
	}
}

func TestRenderResultsOmitsPublicationTimeWhenTimestampMissing(t *testing.T) {
	recorder := httptest.NewRecorder()
	RenderResults(recorder, searchcore.Response{
		Query: "goreecloud search",
		Results: []searchcore.Result{{
			Title:    "GoreeCloud Search",
			URL:      "https://example.com/search",
			Provider: "example",
		}},
	})
	if strings.Contains(recorder.Body.String(), "result-published") {
		t.Fatal("results UI rendered publication metadata without a timestamp")
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
	for _, expected := range []string{".result-card", ".result-published", ".results-layout", ".results-sidebar", "prefers-reduced-motion", "forced-colors"} {
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

func TestAppStylesAdoptStableGlazeV11IdentityAndFallbacks(t *testing.T) {
	recorder := httptest.NewRecorder()
	Styles(recorder, httptest.NewRequest(http.MethodGet, "/assets/app.css", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, expected := range []string{
		`html[data-glaze-version="1.1"]`,
		`--glz11-deep-teal:#0f6b6f`,
		`--glz11-soft-amber:#d9a35f`,
		`data-glz-appearance="dark"`,
		`data-glz-appearance="deep-dark"`,
		`data-glaze-density-profile="productive"`,
		"prefers-reduced-motion:reduce",
		"prefers-reduced-transparency:reduce",
		"prefers-contrast:more",
		"forced-colors:active",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("app stylesheet missing Glaze V1.1 contract %q", expected)
		}
	}
	for _, legacy := range []string{"#5d62ff", "#7c4dff"} {
		if strings.Contains(strings.ToLower(body), legacy) {
			t.Fatalf("app stylesheet retained legacy violet identity %q", legacy)
		}
	}
}

func TestPreferencesStylesContainCompactNavigationOverflow(t *testing.T) {
	recorder := httptest.NewRecorder()
	PreferencesStyles(recorder, httptest.NewRequest(http.MethodGet, "/assets/preferences.css", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/css; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	body := recorder.Body.String()
	for _, expected := range []string{
		".preferences-layout,.settings-nav,.settings-content{min-width:0}",
		".settings-nav nav{display:flex;min-width:0;max-width:100%;overflow-x:auto",
		".settings-nav nav a{flex:0 0 auto;white-space:nowrap}",
		"min-height:48px",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("Preferences stylesheet missing compact overflow contract %q", expected)
		}
	}
}

func TestHomepageUsesOnlySharedLocalAppearanceBootstrap(t *testing.T) {
	recorder := httptest.NewRecorder()
	Homepage(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	body := recorder.Body.String()

	for _, expected := range []string{
		`data-glaze-version="1.1"`,
		`data-glaze-density-profile="comfortable"`,
		`/assets/appearance.js`,
		"/assets/home.css",
		"No advertising",
		"No click tracking",
		"Provider transparency",
		"name=\"category\"",
		"value=\"images\"",
		"value=\"news\"",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("homepage missing %q", expected)
		}
	}
	if strings.Count(strings.ToLower(body), "<script") != 1 {
		t.Fatal("homepage must execute only the shared local appearance bootstrap")
	}
	for _, forbidden := range []string{"/assets/preferences.js", "/assets/results.js", "http://", "https://"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("homepage unexpectedly references %q", forbidden)
		}
	}
}

func TestAppearanceScriptIsLocalOnlyAndGlazeV11Bound(t *testing.T) {
	recorder := httptest.NewRecorder()
	AppearanceScript(recorder, httptest.NewRequest(http.MethodGet, "/assets/appearance.js", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/javascript; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	body := recorder.Body.String()
	for _, expected := range []string{
		"goreecloud.search.preferences.v1",
		`root.dataset.glazeVersion = "1.1"`,
		`"deep-dark"`,
		`comfortable: "comfortable"`,
		`compact: "productive"`,
		"localStorage",
		"GoreeCloudSearchAppearance",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("appearance bootstrap missing %q", expected)
		}
	}
	for _, forbidden := range []string{"fetch(", "XMLHttpRequest", "sendBeacon", "WebSocket", "http://", "https://"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("appearance bootstrap unexpectedly contains network primitive/reference %q", forbidden)
		}
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
		`"deep-dark"`,
		"GoreeCloudSearchAppearance",
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
		`data-glaze-version="1.1"`,
		`/assets/appearance.js`,
		"/assets/preferences.js",
		"data-settings-filter",
		"data-export-preferences",
		"data-import-preferences",
		"data-reset-preferences",
		"data-preference=\"search.autocomplete\"",
		"data-preference=\"privacy.recent_queries\"",
		`value="deep-dark"`,
		"Glaze UI V1.1",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("Preferences page missing %q", expected)
		}
	}
}

type errPrivate string

func (e errPrivate) Error() string { return string(e) }
