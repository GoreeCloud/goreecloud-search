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

type imageProvider struct{}

func (imageProvider) Name() string { return "images" }
func (imageProvider) Search(context.Context, string) ([]searchcore.Result, error) {
	return nil, nil
}
func (imageProvider) Categories() []string { return []string{searchcore.CategoryImages} }

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

func TestExecutableCategoriesReflectCurrentEngine(t *testing.T) {
	engine := searchcore.NewEngine(time.Second, imageProvider{})
	categories := executableCategories(engine)
	if strings.Join(categories, ",") != "general,images" {
		t.Fatalf("executable categories = %v, want [general images]", categories)
	}
}

func TestProviderDefinitionsAPIIsReadOnlyAndNonProduction(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second)}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/providers/definitions", nil)
	response := httptest.NewRecorder()
	app.providerDefinitions(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, required := range []string{
		`"configured_provider_count":0`,
		`"credentials_exposed":false`,
		`"management_scope":"deployment-controlled"`,
		`"production_approved":false`,
		`"supported_categories"`,
		`"executable_categories":["general"]`,
		`"category_execution_scope":"current-native-engine"`,
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("provider registry response missing %s: %s", required, body)
		}
	}
}

func TestSyncCapabilitiesAPIIsReadOnlyAndNonProduction(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second)}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/sync/capabilities", nil)
	response := httptest.NewRecorder()
	app.syncCapabilities(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, required := range []string{
		`"application":"search"`,
		`"credentials_exposed":false`,
		`"production_approved":false`,
		`"dataset":"search.preferences"`,
		`"dataset":"search.history"`,
		`"dataset":"search.sources"`,
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("sync capability response missing %s: %s", required, body)
		}
	}
}
