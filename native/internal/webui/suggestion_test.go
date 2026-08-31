package webui

import (
	"net/http/httptest"
	"strings"
	"testing"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

func TestRenderResultsShowsEscapedExplicitCorrectionAndPreservesCategory(t *testing.T) {
	recorder := httptest.NewRecorder()
	RenderResults(recorder, searchcore.Response{
		Query:          "goreecluod search",
		SuggestedQuery: `goreecloud <script>alert(1)</script>`,
		Category:       searchcore.CategoryNews,
		Results: []searchcore.Result{{
			Title:    "GoreeCloud Search",
			URL:      "https://example.com/search",
			Provider: "example",
		}},
	})

	body := recorder.Body.String()
	for _, expected := range []string{
		`class="query-suggestion"`,
		"Did you mean",
		"goreecloud &lt;script&gt;alert(1)&lt;/script&gt;",
		"category=news",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("rendered suggestion missing %q", expected)
		}
	}
	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Fatal("suggested query rendered unsafe HTML")
	}
	if !strings.Contains(body, "q=goreecloud") || !strings.Contains(body, "%3Cscript%3E") {
		t.Fatalf("suggested query href was not URL-escaped: %s", body)
	}
}

func TestRenderResultsOmitsCorrectionWhenSuggestionMissing(t *testing.T) {
	recorder := httptest.NewRecorder()
	RenderResults(recorder, searchcore.Response{
		Query:    "goreecloud search",
		Category: searchcore.CategoryGeneral,
		Results: []searchcore.Result{{
			Title:    "GoreeCloud Search",
			URL:      "https://example.com/search",
			Provider: "example",
		}},
	})
	if strings.Contains(recorder.Body.String(), "query-suggestion") {
		t.Fatal("results UI rendered correction affordance without a suggestion")
	}
}

func TestResultsStylesIncludeCorrectionTargetAndAccessibilityFallbacks(t *testing.T) {
	recorder := httptest.NewRecorder()
	ResultsStyles(recorder, httptest.NewRequest("GET", "/assets/results.css", nil))
	body := recorder.Body.String()
	for _, expected := range []string{".query-suggestion", ".query-suggestion a", "min-height:44px", "min-height:48px", "prefers-contrast:more", "forced-colors"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("results stylesheet missing correction contract %q", expected)
		}
	}
}
