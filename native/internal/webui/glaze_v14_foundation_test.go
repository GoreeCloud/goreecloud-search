package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

func TestGlazeV141HTMLMigrationIsServedAcrossNativePages(t *testing.T) {
	request := httptest.NewRequest("GET", "/", nil)

	for name, handler := range map[string]func(http.ResponseWriter, *http.Request){
		"home":        Homepage,
		"preferences": Preferences,
	} {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler(recorder, request)
			body := recorder.Body.String()
			if !strings.Contains(body, `data-glaze-version="1.4.1"`) {
				t.Fatalf("served %s page is not bound to Glaze 1.4.1", name)
			}
			if !strings.Contains(body, `data-glaze-optical-v14="adaptive-optical"`) {
				t.Fatalf("served %s page is missing bounded V1.4.1 optical mode", name)
			}
			if strings.Contains(body, `data-glaze-version="1.1"`) || strings.Contains(body, "Glaze UI V1.1") {
				t.Fatalf("served %s page still exposes the retired local V1.1 presentation marker", name)
			}
		})
	}
}

func TestResultsTemplateUsesGlazeV141RuntimeMarkup(t *testing.T) {
	recorder := httptest.NewRecorder()
	RenderResults(recorder, searchcore.Response{Query: "goreecloud", Category: searchcore.CategoryGeneral})
	body := recorder.Body.String()
	if !strings.Contains(body, `data-glaze-version="1.4.1"`) {
		t.Fatal("results page is not bound to Glaze 1.4.1")
	}
	if strings.Contains(body, `data-glaze-version="1.1"`) {
		t.Fatal("results page still exposes the V1.1 marker")
	}
}

func TestStylesPrependGlazeV141AccessibilitySourceProjection(t *testing.T) {
	recorder := httptest.NewRecorder()
	Styles(recorder, httptest.NewRequest("GET", "/assets/app.css", nil))
	body := recorder.Body.String()

	for _, required := range []string{
		`html[data-glaze-version="1.4.1"]`,
		`--glz14-frost-strength`,
		`--glz14-memory-tint-influence: 0`,
		`prefers-reduced-transparency`,
		`forced-colors: active`,
		`solid-accessible`,
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("combined Search stylesheet is missing %q", required)
		}
	}
}
