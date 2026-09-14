package main

import (
	"context"
	"errors"
	"net/http/httptest"
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
	r := httptest.NewRequest("POST", "/api/v1/search", nil)
	r.Header.Set(searchPrivacyAuthorizationHeader, "privacy-shield:capability:test")

	if err := gate.Verify(r); !errors.Is(err, errPrivacyAuthorizationVerifierUnavailable) {
		t.Fatalf("verify error = %v, want verifier unavailable", err)
	}
	if gate.Enforced() {
		t.Fatal("required gate without verifier must not claim enforcement")
	}
}

func TestSearchPrivacyAuthorizationGateRejectsMissingOrAmbiguousReference(t *testing.T) {
	verifier := &recordingSearchPrivacyVerifier{}
	gate := searchPrivacyAuthorizationGate{required: true, verifier: verifier}

	missing := httptest.NewRequest("POST", "/api/v1/search", nil)
	if err := gate.Verify(missing); !errors.Is(err, errPrivacyAuthorizationReferenceRequired) {
		t.Fatalf("missing reference error = %v, want required", err)
	}

	ambiguous := httptest.NewRequest("POST", "/api/v1/search", nil)
	ambiguous.Header.Add(searchPrivacyAuthorizationHeader, "privacy-shield:capability:one")
	ambiguous.Header.Add(searchPrivacyAuthorizationHeader, "privacy-shield:capability:two")
	if err := gate.Verify(ambiguous); !errors.Is(err, errPrivacyAuthorizationReferenceAmbiguous) {
		t.Fatalf("ambiguous reference error = %v, want ambiguous", err)
	}

	if verifier.calls != 0 {
		t.Fatalf("verifier calls = %d, want 0", verifier.calls)
	}
}

func TestSearchPrivacyAuthorizationGatePassesExactOperationContextToVerifier(t *testing.T) {
	verifier := &recordingSearchPrivacyVerifier{}
	gate := searchPrivacyAuthorizationGate{required: true, verifier: verifier}
	r := httptest.NewRequest("POST", "/api/v1/search", nil)
	r.Header.Set(searchPrivacyAuthorizationHeader, " privacy-shield:capability:test ")

	if err := gate.Verify(r); err != nil {
		t.Fatalf("verify = %v", err)
	}
	if !gate.Enforced() {
		t.Fatal("required gate with verifier must report enforcement")
	}
	if verifier.calls != 1 {
		t.Fatalf("verifier calls = %d, want 1", verifier.calls)
	}
	if verifier.capabilityReference != "privacy-shield:capability:test" {
		t.Fatalf("capability reference = %q", verifier.capabilityReference)
	}
	want := searchPrivacyAuthorizationContext{
		Operation:      "search.query",
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
	r := httptest.NewRequest("POST", "/api/v1/search", nil)
	r.Header.Set(searchPrivacyAuthorizationHeader, "privacy-shield:capability:revoked")

	if err := gate.Verify(r); !errors.Is(err, wantErr) {
		t.Fatalf("verify error = %v, want %v", err, wantErr)
	}
}
