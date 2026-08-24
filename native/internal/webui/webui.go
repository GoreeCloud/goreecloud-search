package webui

import (
	"embed"
	"html/template"
	"net/http"
	"strings"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

//go:embed assets/*
var assets embed.FS

var resultsTemplate = template.Must(template.ParseFS(assets, "assets/results.html"))

type resultsPageData struct {
	Query     string
	Results   []searchcore.Result
	Providers []searchcore.ProviderStatus
	Degraded  bool
	Error     string
}

func Homepage(w http.ResponseWriter, _ *http.Request) {
	serveAsset(w, "assets/index.html", "text/html; charset=utf-8")
}

func Preferences(w http.ResponseWriter, _ *http.Request) {
	serveAsset(w, "assets/preferences.html", "text/html; charset=utf-8")
}

func Styles(w http.ResponseWriter, _ *http.Request) {
	serveAsset(w, "assets/app.css", "text/css; charset=utf-8")
}

func ResultsStyles(w http.ResponseWriter, _ *http.Request) {
	serveAsset(w, "assets/results.css", "text/css; charset=utf-8")
}

func RenderResults(w http.ResponseWriter, response searchcore.Response) {
	renderResults(w, http.StatusOK, resultsPageData{
		Query:     response.Query,
		Results:   response.Results,
		Providers: response.Providers,
		Degraded:  response.Degraded,
	})
}

func RenderSearchError(w http.ResponseWriter, query string, err error) {
	message := "Search could not run. Check the query and try again."
	if strings.TrimSpace(query) == "" {
		message = "Enter something to search for."
	}
	if err == nil {
		message = "Search could not run."
	}
	renderResults(w, http.StatusBadRequest, resultsPageData{Query: strings.TrimSpace(query), Error: message})
}

func renderResults(w http.ResponseWriter, status int, data resultsPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := resultsTemplate.ExecuteTemplate(w, "results.html", data); err != nil {
		http.Error(w, "Unable to render results", http.StatusInternalServerError)
	}
}

func serveAsset(w http.ResponseWriter, name, contentType string) {
	content, err := assets.ReadFile(name)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}
