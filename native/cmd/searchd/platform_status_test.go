package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

func TestStatusAPIDiscoversPlatformStatusWithoutApproval(t *testing.T) {
	app := server{engine: searchcore.NewEngine(time.Second)}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	response := httptest.NewRecorder()
	app.status(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, required := range []string{
		`"platform_status":true`,
		`"platform_status":"/api/v1/platform/status"`,
		`"production_approved":false`,
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("status response missing %s: %s", required, body)
		}
	}
}
