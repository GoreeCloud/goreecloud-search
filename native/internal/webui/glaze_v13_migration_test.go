package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGlazeV13SourceMarkers(t *testing.T) {
	for _, name := range []string{
		"assets/index.html",
		"assets/preferences.html",
		"assets/results.html",
	} {
		content, err := assets.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !strings.Contains(string(content), `data-glaze-version="1.3"`) {
			t.Fatalf("%s does not declare GLAZE UI V1.3", name)
		}
	}

	appearance, err := assets.ReadFile("assets/appearance.js")
	if err != nil {
		t.Fatalf("read appearance bootstrap: %v", err)
	}
	if !strings.Contains(string(appearance), `root.dataset.glazeVersion = "1.3";`) {
		t.Fatal("appearance bootstrap does not enforce GLAZE UI V1.3")
	}
}

func TestStylesAppendGlazeV13MappingAfterLegacyComponents(t *testing.T) {
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
	v13Index := strings.Index(body, "--glz13-frost-white")
	if legacyIndex < 0 {
		t.Fatal("combined stylesheet is missing the existing Search component vocabulary")
	}
	if v13Index < 0 {
		t.Fatal("combined stylesheet is missing the GLAZE UI V1.3 source mapping")
	}
	if v13Index <= legacyIndex {
		t.Fatal("GLAZE UI V1.3 source mapping must be appended after the legacy component stylesheet")
	}

	for _, required := range []string{
		"fc7cc91d2eace8da2371371c2855c24cbcb326a1",
		`html[data-glaze-version="1.3"]`,
		"--glz13-base-glass: rgba(255, 255, 255, 0.58)",
		"--glz13-clear-sky-blue: #68aee0",
		"@media (prefers-reduced-transparency: reduce)",
		"@media (forced-colors: active)",
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("combined stylesheet is missing %q", required)
		}
	}
}

func TestGlazeV13MappingDoesNotClaimUnownedAdaptiveAuthorities(t *testing.T) {
	content, err := assets.ReadFile("assets/glaze-v1.3.css")
	if err != nil {
		t.Fatalf("read V1.3 mapping: %v", err)
	}
	body := string(content)
	for _, required := range []string{
		"no wallpaper/environment sampling",
		"remote dynamic color",
		"contextual intelligence",
		"personalization persistence",
		"system-shell authority",
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("V1.3 mapping does not preserve the bounded-adoption note %q", required)
		}
	}
}
