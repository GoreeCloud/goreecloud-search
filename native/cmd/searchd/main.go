package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GoreeCloud/goreecloud-search/native/internal/preferences"
	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
	"github.com/GoreeCloud/goreecloud-search/native/internal/syncstate"
	"github.com/GoreeCloud/goreecloud-search/native/internal/webui"
)

type server struct {
	engine *searchcore.Engine
}

func main() {
	engine := searchcore.NewEngine(8 * time.Second)
	app := server{engine: engine}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", webui.Homepage)
	mux.HandleFunc("GET /search", app.searchPage)
	mux.HandleFunc("GET /preferences", webui.Preferences)
	mux.HandleFunc("GET /assets/app.css", webui.Styles)
	mux.HandleFunc("GET /assets/home.css", webui.HomepageStyles)
	mux.HandleFunc("GET /assets/preferences.css", webui.PreferencesStyles)
	mux.HandleFunc("GET /assets/preferences.js", webui.PreferencesScript)
	mux.HandleFunc("GET /assets/results.css", webui.ResultsStyles)
	mux.HandleFunc("GET /assets/categories.css", webui.CategoryStyles)
	mux.HandleFunc("GET /healthz", app.health)
	mux.HandleFunc("GET /api/v1/search", app.searchAPI)
	mux.HandleFunc("GET /api/v1/preferences/definitions", app.preferenceDefinitions)
	mux.HandleFunc("GET /api/v1/providers/definitions", app.providerDefinitions)
	mux.HandleFunc("GET /api/v1/sync/capabilities", app.syncCapabilities)

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
	webui.RenderResults(w, response)
}

func (s server) searchAPI(w http.ResponseWriter, r *http.Request) {
	category, err := requestedCategory(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported search category"})
		return
	}
	if !s.engine.SupportsCategory(category) {
		writeJSON(w, http.StatusNotImplemented, map[string]string{
			"error":    "search category is not implemented in the native provider layer",
			"category": category,
		})
		return
	}
	response, err := s.engine.SearchCategory(r.Context(), r.URL.Query().Get("q"), category)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s server) preferenceDefinitions(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
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
	providers := s.engine.ProviderDefinitions()
	writeJSON(w, http.StatusOK, map[string]any{
		"schema_version":             1,
		"providers":                  providers,
		"configured_provider_count": len(providers),
		"supported_categories":      searchcore.SupportedCategories,
		"executable_categories":     executableCategories(s.engine),
		"category_execution_scope":  "current-native-engine",
		"management_scope":           "deployment-controlled",
		"credentials_exposed":       false,
		"production_approved":       false,
	})
}

func (s server) syncCapabilities(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
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

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}
