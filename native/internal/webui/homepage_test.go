package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHomepageCarriesPrivacyAndCategoryContracts(t *testing.T) {
	recorder := httptest.NewRecorder()
	Homepage(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, expected := range []string{
		"/assets/homepage.css",
		"No advertising",
		"No click tracking",
		"Provider transparency",
		"value=\"general\" checked",
		"value=\"images\"",
		"value=\"videos\"",
		"value=\"news\"",
		"value=\"files\"",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("homepage missing %q", expected)
		}
	}
	if strings.Contains(body, "<script") {
		t.Fatal("homepage unexpectedly introduced script execution")
	}
}

func TestHomepageStylesAreServedAsCSS(t *testing.T) {
	recorder := httptest.NewRecorder()
	HomepageStyles(recorder, httptest.NewRequest(http.MethodGet, "/assets/homepage.css", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/css; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	body := recorder.Body.String()
	for _, expected := range []string{".hero-stage", ".trust-row", ".category-row .chip:has(input:checked)", "prefers-contrast", "forced-colors"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("homepage stylesheet missing %q", expected)
		}
	}
}
