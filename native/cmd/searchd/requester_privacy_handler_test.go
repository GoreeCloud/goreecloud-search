package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

func protectedSearchAPIRequest(body string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/search", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(searchPrivacyAuthorizationHeader, "psc_handler-test")
	request.Header.Set("Authorization", "Bearer opaque.identity.credential")
	return request
}

func requiredIdentityBackedSearchServer(
	t *testing.T,
	identityVerifier identitySearchRequesterVerifier,
	privacyVerifier searchPrivacyAuthorizationVerifier,
) server {
	t.Helper()
	resolver, err := newIdentityBearerSearchRequesterResolver(identityVerifier)
	if err != nil {
		t.Fatalf("create Identity requester resolver: %v", err)
	}
	return server{
		engine: searchcore.NewEngine(time.Second),
		privacyAuthorizationGate: searchPrivacyAuthorizationGate{
			required:          true,
			verifier:          privacyVerifier,
			requesterResolver: resolver,
		},
	}
}

func TestSearchAPIUsesIdentityVerifiedApplicationAsPrivacyShieldRequester(t *testing.T) {
	identityVerifier := &recordingIdentitySearchRequesterVerifier{
		verified: identityVerifiedSearchRequester{
			ApplicationID: "goreecloud-browser",
			PrincipalID:   "identity:subject:browser-user",
		},
	}
	privacyVerifier := &recordingSearchPrivacyVerifier{}
	app := requiredIdentityBackedSearchServer(t, identityVerifier, privacyVerifier)
	request := protectedSearchAPIRequest(`{"query":"test","category":"general","limit":1}`)
	request.Header.Set("X-GoreeCloud-Requester", "attacker-selected-requester")
	response := httptest.NewRecorder()

	app.searchAPI(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if identityVerifier.calls != 1 || identityVerifier.credential != "opaque.identity.credential" {
		t.Fatalf("Identity verifier calls/credential = %d/%q", identityVerifier.calls, identityVerifier.credential)
	}
	if privacyVerifier.calls != 1 {
		t.Fatalf("Privacy Shield verifier calls = %d, want 1", privacyVerifier.calls)
	}
	if privacyVerifier.authorizationContext.RequesterID != "goreecloud-browser" {
		t.Fatalf("Privacy Shield requester = %q, want verified application ID", privacyVerifier.authorizationContext.RequesterID)
	}
	if privacyVerifier.authorizationContext.RequesterID == "identity:subject:browser-user" {
		t.Fatal("user principal must not become Privacy Shield requester_id")
	}
	if privacyVerifier.authorizationContext.RequesterID == "attacker-selected-requester" {
		t.Fatal("caller-selected requester header must not override verified application identity")
	}
}

func TestSearchAPIRejectsInvalidIdentityApplicationOrPrincipalBeforePrivacyShield(t *testing.T) {
	tests := []struct {
		name     string
		verified identityVerifiedSearchRequester
	}{
		{
			name: "missing application",
			verified: identityVerifiedSearchRequester{
				PrincipalID: "identity:subject:browser-user",
			},
		},
		{
			name: "missing principal",
			verified: identityVerifiedSearchRequester{
				ApplicationID: "goreecloud-browser",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			identityVerifier := &recordingIdentitySearchRequesterVerifier{verified: testCase.verified}
			privacyVerifier := &recordingSearchPrivacyVerifier{}
			app := requiredIdentityBackedSearchServer(t, identityVerifier, privacyVerifier)
			request := protectedSearchAPIRequest(`{"query":"test","category":"general"}`)
			response := httptest.NewRecorder()

			app.searchAPI(response, request)

			if response.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusForbidden, response.Body.String())
			}
			if identityVerifier.calls != 1 {
				t.Fatalf("Identity verifier calls = %d, want 1", identityVerifier.calls)
			}
			if privacyVerifier.calls != 0 {
				t.Fatalf("Privacy Shield verifier calls = %d, want 0", privacyVerifier.calls)
			}
		})
	}
}

func TestSearchAPIStopsBeforePrivacyShieldWhenIdentityRejectsRequester(t *testing.T) {
	identityRejected := errors.New("Identity requester credential rejected")
	identityVerifier := &recordingIdentitySearchRequesterVerifier{err: identityRejected}
	privacyVerifier := &recordingSearchPrivacyVerifier{}
	app := requiredIdentityBackedSearchServer(t, identityVerifier, privacyVerifier)
	request := protectedSearchAPIRequest(`{"query":"test","category":"general"}`)
	response := httptest.NewRecorder()

	app.searchAPI(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusForbidden, response.Body.String())
	}
	if identityVerifier.calls != 1 {
		t.Fatalf("Identity verifier calls = %d, want 1", identityVerifier.calls)
	}
	if privacyVerifier.calls != 0 {
		t.Fatalf("Privacy Shield verifier calls = %d, want 0", privacyVerifier.calls)
	}
	if strings.Contains(response.Body.String(), identityRejected.Error()) {
		t.Fatalf("handler exposed Identity verifier error detail: %s", response.Body.String())
	}
}
