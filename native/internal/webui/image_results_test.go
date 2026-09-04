package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

func TestRenderImageResultsUsesMappedMediaURLsAndViewerContract(t *testing.T) {
	response := httptest.NewRecorder()
	RenderResultsWithMedia(response, searchcore.Response{
		Query:    "golden gate",
		Category: searchcore.CategoryImages,
		Results: []searchcore.Result{{
			Title:    "Golden <Gate>",
			URL:      "https://source.example/page",
			Provider: "Images Provider",
			Media: &searchcore.Media{
				Kind:         searchcore.MediaKindImage,
				ThumbnailURL: "https://media.example/thumb.jpg",
				ContentURL:   "https://media.example/full.jpg",
				Width:        1600,
				Height:       900,
				Alt:          "Bridge <photo>",
			},
		}},
	}, func(raw string) string {
		if strings.Contains(raw, "thumb") {
			return "/media/image?fixture=thumb"
		}
		return "/media/image?fixture=full"
	})

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, forbidden := range []string{"https://media.example/thumb.jpg", "https://media.example/full.jpg", "Bridge <photo>", "Golden <Gate>"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("rendered unbounded provider media content %q", forbidden)
		}
	}
	for _, expected := range []string{
		`class="image-results-grid"`,
		`data-image-open`,
		`data-image-dialog`,
		`data-image-close`,
		`data-image-previous`,
		`data-image-next`,
		`/media/image?fixture=thumb`,
		`/media/image?fixture=full`,
		`1600 × 900`,
		`Open source`,
		`https://source.example/page`,
		`Bridge &lt;photo&gt;`,
		`Golden &lt;Gate&gt;`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("image results missing %q", expected)
		}
	}
}

func TestRenderResultsWithoutMediaMapperNeverEmbedsExternalMedia(t *testing.T) {
	response := httptest.NewRecorder()
	RenderResults(response, searchcore.Response{
		Query:    "image",
		Category: searchcore.CategoryImages,
		Results: []searchcore.Result{{
			Title: "Image",
			URL:   "https://source.example/page",
			Media: &searchcore.Media{
				Kind:         searchcore.MediaKindImage,
				ThumbnailURL: "https://media.example/thumb.jpg",
				ContentURL:   "https://media.example/full.jpg",
			},
		}},
	})
	body := response.Body.String()
	if strings.Contains(body, "media.example") {
		t.Fatalf("external media URL rendered without mapper: %s", body)
	}
	if !strings.Contains(body, "Preview unavailable") {
		t.Fatal("image result without media mapper lost safe fallback")
	}
}

func TestImageResultsStylesPreserveTouchAndResilienceContract(t *testing.T) {
	response := httptest.NewRecorder()
	ImageResultsStyles(response, httptest.NewRequest(http.MethodGet, "/assets/image-results.css", nil))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "text/css; charset=utf-8" {
		t.Fatalf("unexpected image stylesheet response: status=%d type=%q", response.Code, response.Header().Get("Content-Type"))
	}
	body := response.Body.String()
	for _, expected := range []string{".image-results-grid", ".image-viewer", "min-height:48px", "prefers-reduced-motion", "prefers-reduced-transparency", "forced-colors"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("image stylesheet missing %q", expected)
		}
	}
}

func TestResultsScriptIsLocalOnlyAndKeyboardAware(t *testing.T) {
	response := httptest.NewRecorder()
	ResultsScript(response, httptest.NewRequest(http.MethodGet, "/assets/results.js", nil))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "text/javascript; charset=utf-8" {
		t.Fatalf("unexpected results script response: status=%d type=%q", response.Code, response.Header().Get("Content-Type"))
	}
	body := response.Body.String()
	for _, expected := range []string{"showModal", "ArrowLeft", "ArrowRight", "data-image-close", "data-image-open"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("results script missing %q", expected)
		}
	}
	for _, forbidden := range []string{"fetch(", "XMLHttpRequest", "sendBeacon", "WebSocket"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("results script unexpectedly contains network primitive %q", forbidden)
		}
	}
}
