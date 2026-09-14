package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

func TestSearchAPIPostKeepsQueryOutOfURLAndHonorsBodyContract(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second, indexConsumerContractProvider{})}
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/search",
		strings.NewReader(`{"query":"  goreecloud  ","category":"general","limit":2}`),
	)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	response := httptest.NewRecorder()

	app.searchAPI(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if request.URL.RawQuery != "" {
		t.Fatalf("POST query leaked into URL: %q", request.URL.RawQuery)
	}

	var payload struct {
		APIVersion string              `json:"api_version"`
		Query      string              `json:"query"`
		Category   string              `json:"category"`
		Results    []searchcore.Result `json:"results"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.APIVersion != apiVersion {
		t.Fatalf("api_version = %q, want %q", payload.APIVersion, apiVersion)
	}
	if payload.Query != "goreecloud" {
		t.Fatalf("normalized query = %q, want goreecloud", payload.Query)
	}
	if payload.Category != searchcore.CategoryGeneral {
		t.Fatalf("category = %q, want general", payload.Category)
	}
	if len(payload.Results) != 2 {
		t.Fatalf("result count = %d, want 2", len(payload.Results))
	}
}

func TestSearchAPIPostRejectsNonJSONAndUnknownFields(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second, indexConsumerContractProvider{})}
	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{name: "non-json", contentType: "text/plain", body: `goreecloud`},
		{name: "unknown-field", contentType: "application/json", body: `{"query":"goreecloud","unexpected":true}`},
		{name: "multiple-objects", contentType: "application/json", body: `{"query":"one"}{"query":"two"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/search", strings.NewReader(test.body))
			request.Header.Set("Content-Type", test.contentType)
			response := httptest.NewRecorder()

			app.searchAPI(response, request)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body.String())
			}
		})
	}
}
