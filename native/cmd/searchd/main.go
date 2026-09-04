package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GoreeCloud/goreecloud-search/native/internal/mediaproxy"
	"github.com/GoreeCloud/goreecloud-search/native/internal/platformstate"
	"github.com/GoreeCloud/goreecloud-search/native/internal/preferences"
	"github.com/GoreeCloud/goreecloud-search/native/internal/providers"
	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
	"github.com/GoreeCloud/goreecloud-search/native/internal/syncstate"
	"github.com/GoreeCloud/goreecloud-search/native/internal/webui"
)

const apiVersion = "1"

type capabilityEvidence struct {
	ID                 string `json:"id"`
	ContractVersion    string `json:"contract_version"`
	Authoritative      bool   `json:"authoritative"`
	Current            bool   `json:"current"`
	ProductionAccepted bool   `json:"production_accepted"`
	Endpoint           string `json:"endpoint"`
}

type server struct {
	engine *searchcore.Engine
	media  *mediaproxy.Proxy
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
		},
	}
}

func main() {
	configuredProviders, err := providers.LoadFromEnvironment()
	if err != nil {
		log.Fatalf("initialize GoreeCloud Search providers: %v", err)
	}
	engine := searchcore.NewEngine(8*time.Second, configuredProviders...)
	mediaProxy, err := mediaproxy.New()
	if err != nil {
		log.Fatalf("initialize GoreeCloud Search media boundary: %v", err)
	}
	app := server{engine: engine, media: mediaProxy}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", webui.Homepage)
	mux.HandleFunc("GET /search", app.searchPage)
	mux.HandleFunc("GET /preferences", webui.Preferences)
	mux.HandleFunc("GET /assets/app.css", webui.Styles)
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
	mux.HandleFunc("GET /api/v1/preferences/definitions", app.preferenceDefinitions)
	mux.HandleFunc("GET /api/v1/providers/definitions", app.providerDefinitions)
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
		"api_version":         apiVersion,
		"product":             "GoreeCloud Search",
		"service":             "search",
		"status":              "ok",
		"implementation":      "native",
		"lifecycle":           "pre-stable",
		"production_approved": false,
		"capabilities": map[string]bool{
			"html_search":                 true,
			"machine_readable_search_api": true,
			"preferences_definitions":     true,
			"provider_definitions":        true,
			"sync_capabilities":           true,
			"platform_status":             true,
		},
		"capability_evidence": searchCapabilityEvidence(),
		"endpoints": map[string]string{
			"health":                  "/healthz",
			"status":                  "/api/v1/status",
			"readiness":               "/api/v1/readiness",
			"search":                  "/api/v1/search",
			"interactive_search":      "/search",
			"preferences_definitions": "/api/v1/preferences/definitions",
			"provider_definitions":    "/api/v1/providers/definitions",
			"sync_capabilities":       "/api/v1/sync/capabilities",
			"platform_status":         "/api/v1/platform/status",
		},
	})
}

func (s server) readiness(w http.ResponseWriter, _ *http.Request) {
	engineInitialized := s.engine != nil
	generalCategoryReady := false
	if engineInitialized {
		generalCategoryReady = s.engine.SupportsCategory(searchcore.CategoryGeneral)
	}
	ready := engineInitialized && generalCategoryReady
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
			"native_engine_initialized": engineInitialized,
			"general_category_ready":     generalCategoryReady,
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

func requestedCategory(r *http.Request) (string, error) {
	return searchcore.ValidateCategory(r.URL.Query().Get("category"))
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
	category, err := requestedCategory(r)
	if err != nil {
		writeAPIV1JSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported search category"})
		return
	}
	if !s.engine.SupportsCategory(category) {
		writeAPIV1JSON(w, http.StatusNotImplemented, map[string]string{
			"error":    "search category is not implemented in the native provider layer",
			"category": category,
		})
		return
	}
	response, err := s.engine.SearchCategory(r.Context(), r.URL.Query().Get("q"), category)
	if err != nil {
		writeAPIV1JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeAPIV1JSON(w, http.StatusOK, response)
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
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; img-src 'self' data:; form-action 'self'; frame-ancestors 'none'; base-uri 'none'")
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
