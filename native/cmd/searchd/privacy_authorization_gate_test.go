package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type recordingSearchPrivacyVerifier struct {
	calls                int
	capabilityReference  string
	authorizationContext searchPrivacyAuthorizationContext
	err                  error
}

func (v *recordingSearchPrivacyVerifier) VerifySearchCapability(
	_ context.Context,
	capabilityReference string,
	authorizationContext searchPrivacyAuthorizationContext,
) error {
	v.calls++
	v.capabilityReference = capabilityReference
	v.authorizationContext = authorizationContext
	return v.err
}

func protectedSearchRequest(method string) *http.Request {
	r := httptest.NewRequest(method, "/api/v1/search", nil)
	r.Header.Set("Content-Type", "application/json")
	return r
}

func TestSearchPrivacyAuthorizationGateDevelopmentModeDoesNotFabricateEnforcement(t *testing.T) {
	gate := searchPrivacyAuthorizationGate{required: false}
	r := httptest.NewRequest("POST", "/api/v1/search", nil)

	if err := gate.Verify(r); err != nil {
		t.Fatalf("development gate verify = %v, want nil", err)
	}
	if gate.Enforced() {
		t.Fatal("development gate must not claim enforcement")
	}
}

func TestSearchPrivacyAuthorizationGateRequiredModeNeedsVerifier(t *testing.T) {
	gate := searchPrivacyAuthorizationGate{required: true}
	r := protectedSearchRequest(http.MethodPost)
	r.Header.Set(searchPrivacyAuthorizationHeader, "psc_test")

	if err := gate.Verify(r); !errors.Is(err, errPrivacyAuthorizationVerifierUnavailable) {
		t.Fatalf("verify error = %v, want verifier unavailable", err)
	}
	if gate.Enforced() {
		t.Fatal("required gate without verifier must not claim enforcement")
	}
}

func TestSearchPrivacyAuthorizationGateRejectsMissingAmbiguousOrInvalidReference(t *testing.T) {
	verifier := &recordingSearchPrivacyVerifier{}
	gate := searchPrivacyAuthorizationGate{required: true, verifier: verifier}

	missing := protectedSearchRequest(http.MethodPost)
	if err := gate.Verify(missing); !errors.Is(err, errPrivacyAuthorizationReferenceRequired) {
		t.Fatalf("missing reference error = %v, want required", err)
	}

	ambiguous := protectedSearchRequest(http.MethodPost)
	ambiguous.Header.Add(searchPrivacyAuthorizationHeader, "psc_one")
	ambiguous.Header.Add(searchPrivacyAuthorizationHeader, "psc_two")
	if err := gate.Verify(ambiguous); !errors.Is(err, errPrivacyAuthorizationReferenceAmbiguous) {
		t.Fatalf("ambiguous reference error = %v, want ambiguous", err)
	}

	invalidReferences := []string{
		"privacy-shield:capability:test",
		"psc_has whitespace",
		"psc_line\nbreak",
		"psc_" + strings.Repeat("x", maxPrivacyAuthorizationReferenceBytes),
	}
	for _, reference := range invalidReferences {
		invalid := protectedSearchRequest(http.MethodPost)
		invalid.Header.Set(searchPrivacyAuthorizationHeader, reference)
		if err := gate.Verify(invalid); !errors.Is(err, errPrivacyAuthorizationReferenceInvalid) {
			t.Fatalf("invalid reference %q error = %v, want invalid", reference, err)
		}
	}

	if verifier.calls != 0 {
		t.Fatalf("verifier calls = %d, want 0", verifier.calls)
	}
}

func TestSearchPrivacyAuthorizationGateRejectsNonPrivateTransportBeforeVerification(t *testing.T) {
	verifier := &recordingSearchPrivacyVerifier{}
	gate := searchPrivacyAuthorizationGate{required: true, verifier: verifier}

	getRequest := protectedSearchRequest(http.MethodGet)
	getRequest.Header.Set(searchPrivacyAuthorizationHeader, "psc_test")
	if err := gate.Verify(getRequest); !errors.Is(err, errPrivacyAuthorizationMethodRequired) {
		t.Fatalf("GET verify error = %v, want POST required", err)
	}

	wrongMediaType := httptest.NewRequest(http.MethodPost, "/api/v1/search", nil)
	wrongMediaType.Header.Set(searchPrivacyAuthorizationHeader, "psc_test")
	wrongMediaType.Header.Set("Content-Type", "text/plain")
	if err := gate.Verify(wrongMediaType); !errors.Is(err, errPrivacyAuthorizationMediaTypeRequired) {
		t.Fatalf("media type verify error = %v, want application/json required", err)
	}

	if verifier.calls != 0 {
		t.Fatalf("verifier calls = %d, want 0", verifier.calls)
	}
}

func TestSearchPrivacyAuthorizationGatePassesExactOperationContextToVerifier(t *testing.T) {
	verifier := &recordingSearchPrivacyVerifier{}
	gate := searchPrivacyAuthorizationGate{required: true, verifier: verifier}
	r := protectedSearchRequest(http.MethodPost)
	r.Header.Set(searchPrivacyAuthorizationHeader, " psc_test-capability ")

	if err := gate.Verify(r); err != nil {
		t.Fatalf("verify = %v", err)
	}
	if !gate.Enforced() {
		t.Fatal("required gate with verifier must report enforcement")
	}
	if verifier.calls != 1 {
		t.Fatalf("verifier calls = %d, want 1", verifier.calls)
	}
	if verifier.capabilityReference != "psc_test-capability" {
		t.Fatalf("capability reference = %q", verifier.capabilityReference)
	}
	want := searchPrivacyAuthorizationContext{
		Resource:       "goreecloud.search.query",
		Operation:      "search.query",
		Purpose:        "internet_search",
		ProcessingZone: "private_goreecloud",
		Destination:    "https://search.goreecloud.com",
		RetentionMode:  "none",
	}
	if verifier.authorizationContext != want {
		t.Fatalf("authorization context = %#v, want %#v", verifier.authorizationContext, want)
	}
}

func TestSearchPrivacyAuthorizationGatePropagatesVerifierRejection(t *testing.T) {
	wantErr := errors.New("revoked capability")
	verifier := &recordingSearchPrivacyVerifier{err: wantErr}
	gate := searchPrivacyAuthorizationGate{required: true, verifier: verifier}
	r := protectedSearchRequest(http.MethodPost)
	r.Header.Set(searchPrivacyAuthorizationHeader, "psc_revoked")

	if err := gate.Verify(r); !errors.Is(err, wantErr) {
		t.Fatalf("verify error = %v, want %v", err, wantErr)
	}
}
