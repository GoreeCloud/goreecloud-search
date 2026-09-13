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

type indexConsumerContractProvider struct{}

func (indexConsumerContractProvider) Name() string { return "index-contract-fixture" }

func (indexConsumerContractProvider) Search(context.Context, string) ([]searchcore.Result, error) {
	return []searchcore.Result{
		{Title: "First result", URL: "https://example.test/first", Score: 300},
		{Title: "Second result", URL: "https://example.test/second", Score: 200},
		{Title: "Third result", URL: "https://example.test/third", Score: 100},
	}, nil
}

func TestSearchAPIHonorsBoundedResultLimit(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second, indexConsumerContractProvider{})}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=%20test%20&category=general&limit=2", nil)
	response := httptest.NewRecorder()

	app.searchAPI(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}

	var payload searchcore.Response
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Query != "test" {
		t.Fatalf("normalized query = %q, want test", payload.Query)
	}
	if payload.Category != searchcore.CategoryGeneral {
		t.Fatalf("category = %q, want general", payload.Category)
	}
	if len(payload.Results) != 2 {
		t.Fatalf("result count = %d, want 2", len(payload.Results))
	}
	if response.Header().Get("X-GoreeCloud-API-Version") != apiVersion {
		t.Fatalf("search API missing version header")
	}
}

func TestSearchAPIResultLimitIsAdditiveWhenOmitted(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second, indexConsumerContractProvider{})}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=test&category=general", nil)
	response := httptest.NewRecorder()

	app.searchAPI(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}

	var payload searchcore.Response
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Results) != 3 {
		t.Fatalf("result count without limit = %d, want 3", len(payload.Results))
	}
}

func TestSearchAPIRejectsInvalidOrAmbiguousResultLimit(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second, indexConsumerContractProvider{})}
	paths := []string{
		"/api/v1/search?q=test&limit=0",
		"/api/v1/search?q=test&limit=101",
		"/api/v1/search?q=test&limit=not-a-number",
		"/api/v1/search?q=test&limit=",
		"/api/v1/search?q=test&limit=1&limit=2",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()
			app.searchAPI(response, request)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body.String())
			}
		})
	}
}

func TestRequestedResultLimitCapsFirstPartyConsumerRequests(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=test&limit=100", nil)
	limit, present, err := requestedResultLimit(request)
	if err != nil || !present || limit != maxAPISearchResults {
		t.Fatalf("limit = %d, present = %v, err = %v", limit, present, err)
	}
}
