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

func TestStatusAPIIdentifiesNativePreStableContract(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second)}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/status?ignored=query", nil)
	response := httptest.NewRecorder()
	app.status(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Header().Get("X-GoreeCloud-API-Version") != apiVersion {
		t.Fatalf("missing API version header: %q", response.Header().Get("X-GoreeCloud-API-Version"))
	}
	body := response.Body.String()
	for _, required := range []string{
		`"product":"GoreeCloud Search"`,
		`"implementation":"native"`,
		`"lifecycle":"pre-stable"`,
		`"machine_readable_search_api":true`,
		`"production_approved":false`,
		`"readiness":"/api/v1/readiness"`,
		`"id":"search.query"`,
		`"contract_version":"1"`,
		`"authoritative":true`,
		`"current":true`,
		`"production_accepted":false`,
		`"endpoint":"/api/v1/search"`,
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("status response missing %s: %s", required, body)
		}
	}
	if strings.Contains(body, "ignored") {
		t.Fatalf("status endpoint echoed query input: %s", body)
	}
}

func TestSearchCapabilityEvidenceRemainsPreStable(t *testing.T) {
	evidence := searchCapabilityEvidence()
	if len(evidence) != 1 {
		t.Fatalf("capability evidence count = %d, want 1", len(evidence))
	}
	query := evidence[0]
	if query.ID != "search.query" || query.ContractVersion != apiVersion || !query.Authoritative || !query.Current {
		t.Fatalf("unexpected Search query capability evidence: %+v", query)
	}
	if query.ProductionAccepted {
		t.Fatal("pre-Stable Search must not claim production acceptance")
	}
	if query.Endpoint != "/api/v1/search" {
		t.Fatalf("Search query endpoint = %q", query.Endpoint)
	}
}

func TestReadinessAPIBoundsItselfToLocalNativeApplication(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second)}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/readiness", nil)
	response := httptest.NewRecorder()
	app.readiness(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, required := range []string{
		`"ready":true`,
		`"readiness_scope":"local_native_application"`,
		`"native_engine_initialized":true`,
		`"general_category_ready":true`,
		`"external_search_providers"`,
		`"backup_restore_and_rollback"`,
		`"production_approved":false`,
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("readiness response missing %s: %s", required, body)
		}
	}
}

func TestReadinessAPIFailsClosedWithoutNativeEngine(t *testing.T) {
	app := server{}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/readiness", nil)
	response := httptest.NewRecorder()
	app.readiness(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	body := response.Body.String()
	if !strings.Contains(body, `"ready":false`) || !strings.Contains(body, `"status":"not_ready"`) {
		t.Fatalf("readiness must fail closed without an engine: %s", body)
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
	if response.Header().Get("X-GoreeCloud-API-Version") != apiVersion {
		t.Fatalf("search API missing version header")
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
