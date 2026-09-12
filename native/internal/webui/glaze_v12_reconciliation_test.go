package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGlazeV12SourceMarkers(t *testing.T) {
	for _, name := range []string{
		"assets/index.html",
		"assets/preferences.html",
		"assets/results.html",
	} {
		content, err := assets.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !strings.Contains(string(content), `data-glaze-version="1.2"`) {
			t.Fatalf("%s does not declare Glaze UI 1.2", name)
		}
	}

	appearance, err := assets.ReadFile("assets/appearance.js")
	if err != nil {
		t.Fatalf("read appearance bootstrap: %v", err)
	}
	if !strings.Contains(string(appearance), `root.dataset.glazeVersion = "1.2";`) {
		t.Fatal("appearance bootstrap does not enforce Glaze UI 1.2")
	}
}

func TestStylesAppendGlazeV12ReconciliationAfterLegacyComponents(t *testing.T) {
	recorder := httptest.NewRecorder()
	Styles(recorder, nil)

	if recorder.Code != http.StatusOK {
		t.Fatalf("Styles status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/css; charset=utf-8" {
		t.Fatalf("Styles Content-Type = %q", got)
	}

	body := recorder.Body.String()
	legacyIndex := strings.Index(body, "--glz11-deep-teal")
	v12Index := strings.Index(body, "--glz12-frost-white")
	if legacyIndex < 0 {
		t.Fatal("combined stylesheet is missing the existing component vocabulary")
	}
	if v12Index < 0 {
		t.Fatal("combined stylesheet is missing the Glaze UI 1.2 reconciliation layer")
	}
	if v12Index <= legacyIndex {
		t.Fatal("Glaze UI 1.2 reconciliation must be appended after the legacy component stylesheet")
	}

	for _, required := range []string{
		"Neutral glass is the material. Color is an accent.",
		`html[data-glaze-version="1.2"]`,
		"--glz12-base-glass: rgba(255, 255, 255, 0.58)",
		"--glz12-clear-sky-blue: #68aee0",
		"@media (forced-colors: active)",
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("combined stylesheet is missing %q", required)
		}
	}
}
