package webui

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

//go:embed assets/*
var assets embed.FS

var resultsTemplate = template.Must(template.ParseFS(assets, "assets/results.html"))

type imageResultView struct {
	ThumbnailURL string
	ContentURL   string
	Alt          string
	Dimensions   string
	DialogID     string
	OpenerID     string
}

type resultView struct {
	searchcore.Result
	Index int
	Image *imageResultView
}

type resultsPageData struct {
	Query          string
	SuggestedQuery string
	Category       string
	Results        []resultView
	Providers      []searchcore.ProviderStatus
	Degraded       bool
	Error          string
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

func HomepageStyles(w http.ResponseWriter, _ *http.Request) {
	serveAsset(w, "assets/home.css", "text/css; charset=utf-8")
}

func PreferencesStyles(w http.ResponseWriter, _ *http.Request) {
	serveAsset(w, "assets/preferences.css", "text/css; charset=utf-8")
}

func PreferencesScript(w http.ResponseWriter, _ *http.Request) {
	serveAsset(w, "assets/preferences.js", "text/javascript; charset=utf-8")
}

func ResultsStyles(w http.ResponseWriter, _ *http.Request) {
	serveAsset(w, "assets/results.css", "text/css; charset=utf-8")
}

func ImageResultsStyles(w http.ResponseWriter, _ *http.Request) {
	serveAsset(w, "assets/image-results.css", "text/css; charset=utf-8")
}

func ResultsScript(w http.ResponseWriter, _ *http.Request) {
	serveAsset(w, "assets/results.js", "text/javascript; charset=utf-8")
}

func CategoryStyles(w http.ResponseWriter, _ *http.Request) {
	serveAsset(w, "assets/categories.css", "text/css; charset=utf-8")
}

func RenderResults(w http.ResponseWriter, response searchcore.Response) {
	RenderResultsWithMedia(w, response, nil)
}

// RenderResultsWithMedia renders provider media only through the supplied URL
// mapper. Production passes the signed same-origin media proxy mapper; tests may
// pass a deterministic local/data mapper. Without a mapper, external media is
// never embedded in the page.
func RenderResultsWithMedia(w http.ResponseWriter, response searchcore.Response, mediaURL func(string) string) {
	renderResults(w, http.StatusOK, resultsPageData{
		Query:          response.Query,
		SuggestedQuery: response.SuggestedQuery,
		Category:       response.Category,
		Results:        buildResultViews(response.Results, mediaURL),
		Providers:      response.Providers,
		Degraded:       response.Degraded,
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
	renderResults(w, http.StatusBadRequest, resultsPageData{Query: strings.TrimSpace(query), Category: searchcore.CategoryGeneral, Error: message})
}

func RenderCategoryError(w http.ResponseWriter, query, category string, status int) {
	message := "That search category is not supported."
	if status == http.StatusNotImplemented {
		message = "This category is preserved in the native Search contract, but its native provider adapters are not implemented yet."
	}
	renderResults(w, status, resultsPageData{Query: strings.TrimSpace(query), Category: category, Error: message})
}

func buildResultViews(results []searchcore.Result, mediaURL func(string) string) []resultView {
	views := make([]resultView, 0, len(results))
	for index, result := range results {
		view := resultView{Result: result, Index: index + 1}
		if mediaURL != nil && result.Media != nil && result.Media.Kind == searchcore.MediaKindImage {
			thumbnail := mediaURL(result.Media.ThumbnailURL)
			content := mediaURL(result.Media.ContentURL)
			if thumbnail != "" && content != "" {
				alt := strings.TrimSpace(result.Media.Alt)
				if alt == "" {
					alt = strings.TrimSpace(result.Title)
				}
				if alt == "" {
					alt = "Search result image"
				}
				dimensions := ""
				if result.Media.Width > 0 && result.Media.Height > 0 {
					dimensions = fmt.Sprintf("%d × %d", result.Media.Width, result.Media.Height)
				}
				view.Image = &imageResultView{
					ThumbnailURL: thumbnail,
					ContentURL:   content,
					Alt:          alt,
					Dimensions:   dimensions,
					DialogID:     fmt.Sprintf("image-viewer-%d", index+1),
					OpenerID:     fmt.Sprintf("image-opener-%d", index+1),
				}
			}
		}
		views = append(views, view)
	}
	return views
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
