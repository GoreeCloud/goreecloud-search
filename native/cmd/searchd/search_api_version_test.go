package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

type apiVersionFixtureProvider struct{}

func (apiVersionFixtureProvider) Name() string { return "api-version-fixture" }
func (apiVersionFixtureProvider) Search(context.Context, string) ([]searchcore.Result, error) {
	return []searchcore.Result{{Title: "Result", URL: "https://example.test/result"}}, nil
}

func TestSearchAPIBodyCarriesContractVersion(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second, apiVersionFixtureProvider{})}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=goreecloud&category=general&limit=1", nil)
	response := httptest.NewRecorder()

	app.searchAPI(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}

	var payload struct {
		APIVersion string `json:"api_version"`
		Query      string `json:"query"`
		Category   string `json:"category"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.APIVersion != apiVersion {
		t.Fatalf("api_version = %q, want %q", payload.APIVersion, apiVersion)
	}
	if payload.Query != "goreecloud" || payload.Category != searchcore.CategoryGeneral {
		t.Fatalf("unexpected response identity: query=%q category=%q", payload.Query, payload.Category)
	}
	if response.Header().Get("X-GoreeCloud-API-Version") != apiVersion {
		t.Fatalf("header version does not match body version")
	}
}
