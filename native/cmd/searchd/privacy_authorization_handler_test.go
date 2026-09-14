package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchAPIRequiredPrivacyGateRejectsBeforeParsingRequest(t *testing.T) {
	app := server{
		privacyAuthorizationGate: searchPrivacyAuthorizationGate{
			required: true,
			verifier: &recordingSearchPrivacyVerifier{},
		},
	}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/search", strings.NewReader("not-json"))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.searchAPI(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
	if !strings.Contains(w.Body.String(), "Privacy Shield authorization is required") {
		t.Fatalf("body = %q", w.Body.String())
	}
}

func TestSearchAPIRequiredPrivacyGateNeedsVerifier(t *testing.T) {
	app := server{
		privacyAuthorizationGate: searchPrivacyAuthorizationGate{required: true},
	}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/search", strings.NewReader("not-json"))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set(searchPrivacyAuthorizationHeader, "privacy-shield:capability:test")
	w := httptest.NewRecorder()

	app.searchAPI(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(w.Body.String(), "Privacy Shield authorization verifier is unavailable") {
		t.Fatalf("body = %q", w.Body.String())
	}
}

func TestSearchAPIVerifiedPrivacyReferenceReachesRequestParsing(t *testing.T) {
	verifier := &recordingSearchPrivacyVerifier{}
	app := server{
		privacyAuthorizationGate: searchPrivacyAuthorizationGate{
			required: true,
			verifier: verifier,
		},
	}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/search", strings.NewReader("not-json"))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set(searchPrivacyAuthorizationHeader, "privacy-shield:capability:test")
	w := httptest.NewRecorder()

	app.searchAPI(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if verifier.calls != 1 {
		t.Fatalf("verifier calls = %d, want 1", verifier.calls)
	}
	if !strings.Contains(w.Body.String(), "invalid JSON search request") {
		t.Fatalf("body = %q", w.Body.String())
	}
}

func TestSearchAPIVerifierRejectionIsSanitized(t *testing.T) {
	verifier := &recordingSearchPrivacyVerifier{err: errors.New("internal token signature detail")}
	app := server{
		privacyAuthorizationGate: searchPrivacyAuthorizationGate{
			required: true,
			verifier: verifier,
		},
	}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/search", strings.NewReader("not-json"))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set(searchPrivacyAuthorizationHeader, "privacy-shield:capability:rejected")
	w := httptest.NewRecorder()

	app.searchAPI(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
	if strings.Contains(w.Body.String(), "signature") {
		t.Fatalf("verifier details leaked in body: %q", w.Body.String())
	}
}

func TestReadinessFailsClosedWhenPrivacyEnforcementRequiredWithoutVerifier(t *testing.T) {
	app := server{
		privacyAuthorizationGate: searchPrivacyAuthorizationGate{required: true},
	}
	r := httptest.NewRequest(http.MethodGet, "/api/v1/readiness", nil)
	w := httptest.NewRecorder()

	app.readiness(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(w.Body.String(), `"privacy_authorization_boundary_ready":false`) {
		t.Fatalf("readiness body = %q", w.Body.String())
	}
}
