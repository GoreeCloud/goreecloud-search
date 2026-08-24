package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

type server struct {
	engine *searchcore.Engine
}

func main() {
	engine := searchcore.NewEngine(8 * time.Second)
	app := server{engine: engine}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", app.health)
	mux.HandleFunc("GET /api/v1/search", app.search)

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

func (s server) search(w http.ResponseWriter, r *http.Request) {
	response, err := s.engine.Search(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}
