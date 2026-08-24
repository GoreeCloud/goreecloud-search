package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

func TestRenderResultsEscapesProviderContent(t *testing.T) {
	recorder := httptest.NewRecorder()
	RenderResults(recorder, searchcore.Response{
		Query: "goreecloud <script>alert(1)</script>",
		Results: []searchcore.Result{{
			Title:    "<img src=x onerror=alert(1)>",
			URL:      "https://example.com/path?q=one&two=2",
			Snippet:  "<b>provider content</b>",
			Provider: "example <script>",
			Score:    42,
		}},
		Providers: []searchcore.ProviderStatus{{Name: "example", State: searchcore.ProviderStateAvailable, Count: 1}},
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
	for _, expected := range []string{"GoreeCloud Search", "result-card", "example", "Score 42", "https://example.com/path"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("rendered body missing %q", expected)
		}
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
	if !strings.Contains(recorder.Body.String(), ".result-card") {
		t.Fatal("results stylesheet missing result-card contract")
	}
}

type errPrivate string

func (e errPrivate) Error() string { return string(e) }
