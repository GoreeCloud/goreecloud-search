package webui

import (
	"embed"
	"net/http"
)

//go:embed assets/*
var assets embed.FS

func Homepage(w http.ResponseWriter, _ *http.Request) {
	serveAsset(w, "assets/index.html", "text/html; charset=utf-8")
}

func Preferences(w http.ResponseWriter, _ *http.Request) {
	serveAsset(w, "assets/preferences.html", "text/html; charset=utf-8")
}

func Styles(w http.ResponseWriter, _ *http.Request) {
	serveAsset(w, "assets/app.css", "text/css; charset=utf-8")
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
