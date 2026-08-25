package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

func TestRequestedCategoryDefaultsToGeneral(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/search?q=test", nil)
	category, err := requestedCategory(request)
	if err != nil || category != searchcore.CategoryGeneral {
		t.Fatalf("category = %q, err = %v", category, err)
	}
}

func TestSearchAPIRejectsUnknownCategory(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second)}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=test&category=shopping", nil)
	response := httptest.NewRecorder()
	app.searchAPI(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), "unsupported search category") {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}

func TestSearchAPIFailsClosedForPreservedUnimplementedCategory(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second)}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=test&category=images", nil)
	response := httptest.NewRecorder()
	app.searchAPI(response, request)
	if response.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotImplemented)
	}
	body := response.Body.String()
	if !strings.Contains(body, `"category":"images"`) || !strings.Contains(body, "not implemented") {
		t.Fatalf("unexpected response: %s", body)
	}
}

func TestGeneralSearchAPIIncludesCategory(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second)}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=test&category=general", nil)
	response := httptest.NewRecorder()
	app.searchAPI(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"category":"general"`) {
		t.Fatalf("general response missing category contract: %s", response.Body.String())
	}
}
