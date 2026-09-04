package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

type allCategoryAcceptanceProvider struct{}

func (allCategoryAcceptanceProvider) Name() string { return "acceptance" }
func (allCategoryAcceptanceProvider) Categories() []string {
	return append([]string(nil), searchcore.SupportedCategories...)
}
func (allCategoryAcceptanceProvider) Search(ctx context.Context, query string) ([]searchcore.Result, error) {
	return allCategoryAcceptanceProvider{}.SearchCategory(ctx, query, searchcore.CategoryGeneral)
}
func (allCategoryAcceptanceProvider) SearchCategory(_ context.Context, query, category string) ([]searchcore.Result, error) {
	result := searchcore.Result{
		Title:    "GoreeCloud " + category + " result",
		URL:      "https://results.example/" + category,
		Snippet:  "Deterministic result for " + query,
		Score:    50,
	}
	if category == searchcore.CategoryImages {
		result.Media = &searchcore.Media{
			Kind:         searchcore.MediaKindImage,
			ThumbnailURL: "https://media.example/thumb.jpg",
			ContentURL:   "https://media.example/full.jpg",
			MIMEType:     "image/jpeg",
			Width:        1200,
			Height:       800,
			Alt:          "Acceptance image",
		}
	}
	return []searchcore.Result{result}, nil
}

func TestEverySupportedCategoryExecutesThroughSearchAPI(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second, allCategoryAcceptanceProvider{})}
	for _, category := range searchcore.SupportedCategories {
		t.Run(category, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=goreecloud&category="+category, nil)
			response := httptest.NewRecorder()
			app.searchAPI(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			body := response.Body.String()
			for _, expected := range []string{
				`"category":"` + category + `"`,
				`"provider":"acceptance"`,
				`"url":"https://results.example/` + category + `"`,
			} {
				if !strings.Contains(body, expected) {
					t.Fatalf("%s API response missing %q: %s", category, expected, body)
				}
			}
			if category == searchcore.CategoryImages && !strings.Contains(body, `"kind":"image"`) {
				t.Fatalf("image API response lost media contract: %s", body)
			}
		})
	}
}

func TestEverySupportedCategoryExecutesThroughHTMLRoute(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second, allCategoryAcceptanceProvider{})}
	for _, category := range searchcore.SupportedCategories {
		t.Run(category, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/search?q=goreecloud&category="+category, nil)
			response := httptest.NewRecorder()
			app.searchPage(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			body := response.Body.String()
			if !strings.Contains(body, `aria-current="page"`) || !strings.Contains(body, "GoreeCloud "+category+" result") {
				t.Fatalf("%s HTML route did not render selected native category result: %s", category, body)
			}
			if strings.Contains(body, "not implemented") {
				t.Fatalf("%s HTML route unexpectedly reported preserved-only category: %s", category, body)
			}
		})
	}
}
