package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-search/native/internal/buildinfo"
	"github.com/GoreeCloud/goreecloud-search/native/internal/mediaproxy"
	"github.com/GoreeCloud/goreecloud-search/native/internal/platformstate"
	"github.com/GoreeCloud/goreecloud-search/native/internal/preferences"
	"github.com/GoreeCloud/goreecloud-search/native/internal/providers"
	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
	"github.com/GoreeCloud/goreecloud-search/native/internal/syncstate"
	"github.com/GoreeCloud/goreecloud-search/native/internal/webautomation"
	"github.com/GoreeCloud/goreecloud-search/native/internal/webintelligence"
	"github.com/GoreeCloud/goreecloud-search/native/internal/webui"
)

const (
	apiVersion               = "1"
	maxAPISearchResults      = 100
	maxSearchAPIRequestBytes = 16 * 1024
)

type capabilityEvidence struct {
	ID                 string   `json:"id"`
	ContractVersion    string   `json:"contract_version"`
	Authoritative      bool     `json:"authoritative"`
	Current            bool     `json:"current"`
	ProductionAccepted bool     `json:"production_accepted"`
	Endpoint           string   `json:"endpoint"`
	Methods            []string `json:"methods,omitempty"`
	PreferredMethod    string   `json:"preferred_method,omitempty"`
	MaxResults         int      `json:"max_results,omitempty"`
}

type searchAPIResponse struct {
	APIVersion string `json:"api_version"`
	searchcore.Response
}

type searchAPIBodyRequest struct {
	Query    string `json:"query"`
	Category string `json:"category,omitempty"`
	Limit    *int   `json:"limit,omitempty"`
}

type parsedSearchAPIRequest struct {
	query    string
	category string
	limit    int
	hasLimit bool
}

type server struct {
	engine                      *searchcore.Engine
	media                       *mediaproxy.Proxy
	build                       buildinfo.Provenance
	webIntelligence             *webintelligence.Controller
	webIntelligenceConfigured   bool
	webAutomationAuth           *webautomation.AuthControl
	webAutomationAuthConfigured bool
	privacyAuthorizationGate    searchPrivacyAuthorizationGate
}

func searchCapabilityEvidence() []capabilityEvidence {
	return []capabilityEvidence{
		{
			ID:                 "search.query",
			ContractVersion:    apiVersion,
			Authoritative:      true,
			Current:            true,
			ProductionAccepted: false,
			Endpoint:           "/api/v1/search",
			Methods:            []string{http.MethodPost, http.MethodGet},
			PreferredMethod:    http.MethodPost,
			MaxResults:         maxAPISearchResults,
		},
	}
}

func main() {
	configuredProviders, err := providers.LoadFromEnvironment()
	if err != nil {
		log.Fatalf("initialize GoreeCloud Search providers: %v", err)
	}
	webIntelligenceController, webIntelligenceConfigured, err := webintelligence.LoadFromEnvironment()
	if err != nil {
		log.Fatalf("initialize GoreeCloud web intelligence control plane: %v", err)
	}
	webAutomationAuth, webAutomationAuthConfigured, err := webautomation.LoadAuthControlFromEnvironment()
	if err != nil {
		log.Fatalf("initialize GoreeCloud authenticated web automation control plane: %v", err)
	}
	engine := searchcore.NewEngine(8*time.Second, configuredProviders...)
	mediaProxy, err := mediaproxy.New()
	if err != nil {
		log.Fatalf("initialize GoreeCloud Search media boundary: %v", err)
	}
	app := server{
		engine:                      engine,
		media:                       mediaProxy,
		build:                       buildinfo.Current(),
		webIntelligence:             webIntelligenceController,
		webIntelligenceConfigured:   webIntelligenceConfigured,
		webAutomationAuth:           webAutomationAuth,
		webAutomationAuthConfigured: webAutomationAuthConfigured,
		privacyAuthorizationGate: searchPrivacyAuthorizationGate{
			required: false,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", webui.Homepage)
	mux.HandleFunc("GET /search", app.searchPage)
	mux.HandleFunc("GET /preferences", webui.Preferences)
	mux.HandleFunc("GET /assets/app.css", webui.Styles)
	mux.HandleFunc("GET /assets/appearance.js", webui.AppearanceScript)
	mux.HandleFunc("GET /assets/home.css", webui.HomepageStyles)
	mux.HandleFunc("GET /assets/preferences.css", webui.PreferencesStyles)
	mux.HandleFunc("GET /assets/preferences.js", webui.PreferencesScript)
	mux.HandleFunc("GET /assets/results.css", webui.ResultsStyles)
	mux.HandleFunc("GET /assets/image-results.css", webui.ImageResultsStyles)
	mux.HandleFunc("GET /assets/results.js", webui.ResultsScript)
	mux.HandleFunc("GET /assets/categories.css", webui.CategoryStyles)
	mux.Handle("GET /media/image", mediaProxy)
	mux.HandleFunc("GET /healthz", app.health)
	mux.HandleFunc("GET /api/v1/status", app.status)
	mux.HandleFunc("GET /api/v1/readiness", app.readiness)
	mux.HandleFunc("GET /api/v1/search", app.searchAPI)
	mux.HandleFunc("POST /api/v1/search", app.searchAPI)
	mux.HandleFunc("GET /api/v1/preferences/definitions", app.preferenceDefinitions)
	mux.HandleFunc("GET /api/v1/providers/definitions", app.providerDefinitions)
	mux.HandleFunc("GET /api/v1/web-intelligence/status", app.webIntelligenceStatus)
	mux.HandleFunc("GET /api/v1/web-intelligence/authentication/status", app.webAutomationAuthStatus)
	mux.HandleFunc("GET /api/v1/sync/capabilities", app.syncCapabilities)
	mux.HandleFunc("GET /api/v1/platform/status", platformstate.Handler)

	addr := os.Getenv("GOREECLOUD_SEARCH_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("GoreeCloud Search native development service listening on %s", addr)
	log.Fatal(httpServer.ListenAndServe())
}

func (s server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service":             "goreecloud-search",
		"implementation":      "native-development-foundation",
		"production_approved": false,
	})
}

func (s server) status(w http.ResponseWriter, _ *http.Request) {
	writeAPIV1JSON(w, http.StatusOK, map[string]any{
		"api_version":                    apiVersion,
		"product":                        "GoreeCloud Search",
		"service":                        "search",
		"status":                         "ok",
		"implementation":                 "native",
		"lifecycle":                      "development",
		"production_approved":            false,
		"privacy_authorization_enforced": s.privacyAuthorizationGate.Enforced(),
		"build":                          s.build,
		"capabilities": map[string]bool{
			"html_search":                 true,
			"machine_readable_search_api": true,
			"preferences_definitions":     true,
			"provider_definitions":        true,
			"web_intelligence_status":     true,
			"web_automation_auth_status":  true,
			"sync_capabilities":           true,
			"platform_status":             true,
		},
		"capability_evidence": searchCapabilityEvidence(),
		"endpoints": map[string]string{
			"health":                     "/healthz",
			"status":                     "/api/v1/status",
			"readiness":                  "/api/v1/readiness",
			"search":                     "/api/v1/search",
			"interactive_search":         "/search",
			"preferences_definitions":    "/api/v1/preferences/definitions",
			"provider_definitions":       "/api/v1/providers/definitions",
			"web_intelligence_status":    "/api/v1/web-intelligence/status",
			"web_automation_auth_status": "/api/v1/web-intelligence/authentication/status",
			"sync_capabilities":          "/api/v1/sync/capabilities",
			"platform_status":            "/api/v1/platform/status",
		},
	})
}

func (s server) readiness(w http.ResponseWriter, _ *http.Request) {
	engineInitialized := s.engine != nil
	generalCategoryReady := false
	if engineInitialized {
		generalCategoryReady = s.engine.SupportsCategory(searchcore.CategoryGeneral)
	}
	privacyAuthorizationBoundaryReady :=
		!s.privacyAuthorizationGate.required || s.privacyAuthorizationGate.Enforced()
	ready := engineInitialized && generalCategoryReady && privacyAuthorizationBoundaryReady
	status := "ready"
	httpStatus := http.StatusOK
	if !ready {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	writeAPIV1JSON(w, httpStatus, map[string]any{
		"api_version":         apiVersion,
		"product":             "GoreeCloud Search",
		"service":             "search",
		"status":              status,
		"ready":               ready,
		"readiness_scope":     "local_native_application",
		"production_approved": false,
		"checks": map[string]bool{
			"native_engine_initialized":             engineInitialized,
			"general_category_ready":                 generalCategoryReady,
			"privacy_authorization_boundary_ready": privacyAuthorizationBoundaryReady,
		},
		"not_evaluated": []string{
			"external_search_providers",
			"production_provider_credentials",
			"private_dns_and_reverse_proxy",
			"monitoring_and_alert_delivery",
			"backup_restore_and_rollback",
			"physical_device_acceptance",
			"production_cutover",
		},
	})
}

func requestedSingleQueryValue(r *http.Request, key, duplicateError string) (string, error) {
	values, present := r.URL.Query()[key]
	if !present {
		return "", nil
	}
	if len(values) != 1 {
		return "", errors.New(duplicateError)
	}
	return values[0], nil
}

func requestedCategory(r *http.Request) (string, error) {
	raw, err := requestedSingleQueryValue(r, "category", "search category must be specified once")
	if err != nil {
		return "", err
	}
	return searchcore.ValidateCategory(raw)
}

func requestedSearchQuery(r *http.Request) (string, error) {
	return requestedSingleQueryValue(r, "q", "query must be specified once")
}

func requestedResultLimit(r *http.Request) (int, bool, error) {
	values, present := r.URL.Query()["limit"]
	if !present {
		return 0, false, nil
	}
	if len(values) != 1 {
		return 0, true, errors.New("result limit must be specified once")
	}
	raw := strings.TrimSpace(values[0])
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 || limit > maxAPISearchResults {
		return 0, true, errors.New("result limit must be an integer between 1 and 100")
	}
	return limit, true, nil
}

func requestedGETSearchAPIRequest(r *http.Request) (parsedSearchAPIRequest, error) {
	rawQuery, err := requestedSearchQuery(r)
	if err != nil {
		return parsedSearchAPIRequest{}, err
	}
	category, err := requestedCategory(r)
	if err != nil {
		return parsedSearchAPIRequest{}, err
	}
	limit, hasLimit, err := requestedResultLimit(r)
	if err != nil {
		return parsedSearchAPIRequest{}, err
	}
	return parsedSearchAPIRequest{
		query:    rawQuery,
		category: category,
		limit:    limit,
		hasLimit: hasLimit,
	}, nil
}

func requestedPOSTSearchAPIRequest(w http.ResponseWriter, r *http.Request) (parsedSearchAPIRequest, error) {
	contentType := strings.TrimSpace(strings.SplitN(r.Header.Get("Content-Type"), ";", 2)[0])
	if contentType != "application/json" {
		return parsedSearchAPIRequest{}, errors.New("POST search requests require application/json")
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxSearchAPIRequestBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var body searchAPIBodyRequest
	if err := decoder.Decode(&body); err != nil {
		return parsedSearchAPIRequest{}, errors.New("invalid JSON search request")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return parsedSearchAPIRequest{}, errors.New("search request must contain one JSON object")
	}

	category, err := searchcore.ValidateCategory(body.Category)
	if err != nil {
		return parsedSearchAPIRequest{}, err
	}

	parsed := parsedSearchAPIRequest{
		query:    body.Query,
		category: category,
	}
	if body.Limit != nil {
		if *body.Limit < 1 || *body.Limit > maxAPISearchResults {
			return parsedSearchAPIRequest{}, errors.New("result limit must be an integer between 1 and 100")
		}
		parsed.limit = *body.Limit
		parsed.hasLimit = true
	}
	return parsed, nil
}

func requestedSearchAPIRequest(w http.ResponseWriter, r *http.Request) (parsedSearchAPIRequest, error) {
	if r.Method == http.MethodPost {
		return requestedPOSTSearchAPIRequest(w, r)
	}
	return requestedGETSearchAPIRequest(r)
}

func (s server) searchPage(w http.ResponseWriter, r *http.Request) {
	category, err := requestedCategory(r)
	if err != nil {
		webui.RenderCategoryError(w, r.URL.Query().Get("q"), "", http.StatusBadRequest)
		return
	}
	if !s.engine.SupportsCategory(category) {
		webui.RenderCategoryError(w, r.URL.Query().Get("q"), category, http.StatusNotImplemented)
		return
	}
	response, err := s.engine.SearchCategory(r.Context(), r.URL.Query().Get("q"), category)
	if err != nil {
		if r.Context().Err() != nil {
			return
		}
		webui.RenderSearchError(w, r.URL.Query().Get("q"), err)
		return
	}
	if s.media != nil {
		webui.RenderResultsWithMedia(w, response, s.media.URL)
		return
	}
	webui.RenderResults(w, response)
}

func (s server) searchAPI(w http.ResponseWriter, r *http.Request) {
	if err := s.privacyAuthorizationGate.Verify(r); err != nil {
		if errors.Is(err, errPrivacyAuthorizationVerifierUnavailable) {
			writeAPIV1JSON(w, http.StatusServiceUnavailable, map[string]string{
				"error": "Privacy Shield authorization verifier is unavailable",
			})
			return
		}
		writeAPIV1JSON(w, http.StatusForbidden, map[string]string{
			"error": "Privacy Shield authorization is required",
		})
		return
	}

	request, err := requestedSearchAPIRequest(w, r)
	if err != nil {
		writeAPIV1JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if !s.engine.SupportsCategory(request.category) {
		writeAPIV1JSON(w, http.StatusNotImplemented, map[string]string{
			"error":    "search category is not implemented in the native provider layer",
			"category": request.category,
		})
		return
	}
	response, err := s.engine.SearchCategory(r.Context(), request.query, request.category)
	if err != nil {
		if r.Context().Err() != nil {
			return
		}
		writeAPIV1JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if request.hasLimit && len(response.Results) > request.limit {
		response.Results = append([]searchcore.Result(nil), response.Results[:request.limit]...)
	}
	writeAPIV1JSON(w, http.StatusOK, searchAPIResponse{
		APIVersion: apiVersion,
		Response:   response,
	})
}

func (s server) preferenceDefinitions(w http.ResponseWriter, _ *http.Request) {
	writeAPIV1JSON(w, http.StatusOK, map[string]any{
		"schema_version": 1,
		"sections": []string{"search", "sources", "appearance", "privacy", "security", "data-resilience", "advanced"},
		"definitions": preferences.Definitions(),
	})
}

func executableCategories(engine *searchcore.Engine) []string {
	categories := make([]string, 0, len(searchcore.SupportedCategories))
	for _, category := range searchcore.SupportedCategories {
		if engine.SupportsCategory(category) {
			categories = append(categories, category)
		}
	}
	return categories
}

func (s server) providerDefinitions(w http.ResponseWriter, _ *http.Request) {
	providerDefinitions := s.engine.ProviderDefinitions()
	writeAPIV1JSON(w, http.StatusOK, map[string]any{
		"schema_version":             1,
		"providers":                  providerDefinitions,
		"configured_provider_count": len(providerDefinitions),
		"supported_categories":      searchcore.SupportedCategories,
		"executable_categories":     executableCategories(s.engine),
		"category_execution_scope":  "current-native-engine",
		"management_scope":           "deployment-controlled",
		"credentials_exposed":       false,
		"production_approved":       false,
	})
}

func (s server) webIntelligenceStatus(w http.ResponseWriter, _ *http.Request) {
	response := map[string]any{
		"schema_version":       1,
		"configured":           s.webIntelligenceConfigured,
		"management_scope":     "development-control-plane",
		"credentials_exposed": false,
		"production_approved": false,
	}
	if s.webIntelligence != nil {
		response["snapshot"] = s.webIntelligence.Snapshot()
	}
	writeAPIV1JSON(w, http.StatusOK, response)
}

func (s server) syncCapabilities(w http.ResponseWriter, _ *http.Request) {
	writeAPIV1JSON(w, http.StatusOK, map[string]any{
		"schema_version":       1,
		"application":          "search",
		"capabilities":         syncstate.Capabilities(),
		"credentials_exposed": false,
		"production_approved": false,
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; img-src 'self' data:; script-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'")
		next.ServeHTTP(w, r)
	})
}

func writeAPIV1JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("X-GoreeCloud-API-Version", apiVersion)
	writeJSON(w, status, value)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}
