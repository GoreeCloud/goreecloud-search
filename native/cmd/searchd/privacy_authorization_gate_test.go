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

type staticSearchRequesterResolver struct {
	requesterID string
	err         error
}

func (r staticSearchRequesterResolver) ResolveSearchRequester(_ *http.Request) (string, error) {
	return r.requesterID, r.err
}

func protectedSearchRequest(method string) *http.Request {
	r := httptest.NewRequest(method, "/api/v1/search", nil)
	r.Header.Set("Content-Type", "application/json")
	return r
}

func requiredSearchPrivacyGate(verifier searchPrivacyAuthorizationVerifier) searchPrivacyAuthorizationGate {
	return searchPrivacyAuthorizationGate{
		required:          true,
		verifier:          verifier,
		requesterResolver: staticSearchRequesterResolver{requesterID: "goreecloud-browser"},
	}
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

func TestSearchPrivacyAuthorizationGateRequiredModeNeedsVerifierAndRequesterResolver(t *testing.T) {
	r := protectedSearchRequest(http.MethodPost)
	r.Header.Set(searchPrivacyAuthorizationHeader, "psc_test")

	withoutVerifier := searchPrivacyAuthorizationGate{
		required:          true,
		requesterResolver: staticSearchRequesterResolver{requesterID: "goreecloud-browser"},
	}
	if err := withoutVerifier.Verify(r); !errors.Is(err, errPrivacyAuthorizationVerifierUnavailable) {
		t.Fatalf("verify error = %v, want verifier unavailable", err)
	}
	if withoutVerifier.Enforced() {
		t.Fatal("required gate without verifier must not claim enforcement")
	}

	withoutResolver := searchPrivacyAuthorizationGate{
		required: true,
		verifier: &recordingSearchPrivacyVerifier{},
	}
	if err := withoutResolver.Verify(r); !errors.Is(err, errPrivacyRequesterResolverUnavailable) {
		t.Fatalf("verify error = %v, want requester resolver unavailable", err)
	}
	if withoutResolver.Enforced() {
		t.Fatal("required gate without requester resolver must not claim enforcement")
	}
}

func TestSearchPrivacyAuthorizationGateRejectsMissingAuthenticatedRequesterIdentity(t *testing.T) {
	verifier := &recordingSearchPrivacyVerifier{}
	gate := searchPrivacyAuthorizationGate{
		required:          true,
		verifier:          verifier,
		requesterResolver: staticSearchRequesterResolver{},
	}
	r := protectedSearchRequest(http.MethodPost)
	r.Header.Set(searchPrivacyAuthorizationHeader, "psc_test")

	if err := gate.Verify(r); !errors.Is(err, errPrivacyRequesterIdentityRequired) {
		t.Fatalf("verify error = %v, want requester identity required", err)
	}
	if verifier.calls != 0 {
		t.Fatalf("verifier calls = %d, want 0", verifier.calls)
	}
}

func TestSearchPrivacyAuthorizationGateRejectsMissingAmbiguousOrInvalidReference(t *testing.T) {
	verifier := &recordingSearchPrivacyVerifier{}
	gate := requiredSearchPrivacyGate(verifier)

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
	gate := requiredSearchPrivacyGate(verifier)

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

func TestSearchPrivacyAuthorizationGatePassesAuthenticatedRequesterAndExactOperationContext(t *testing.T) {
	verifier := &recordingSearchPrivacyVerifier{}
	gate := requiredSearchPrivacyGate(verifier)
	r := protectedSearchRequest(http.MethodPost)
	r.Header.Set(searchPrivacyAuthorizationHeader, " psc_test-capability ")

	if err := gate.Verify(r); err != nil {
		t.Fatalf("verify = %v", err)
	}
	if !gate.Enforced() {
		t.Fatal("required gate with verifier and resolver must report enforcement")
	}
	if verifier.calls != 1 {
		t.Fatalf("verifier calls = %d, want 1", verifier.calls)
	}
	if verifier.capabilityReference != "psc_test-capability" {
		t.Fatalf("capability reference = %q", verifier.capabilityReference)
	}
	want := searchPrivacyAuthorizationContext{
		RequesterID:    "goreecloud-browser",
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

func TestSearchPrivacyAuthorizationGatePropagatesRequesterAndVerifierRejection(t *testing.T) {
	requesterErr := errors.New("requester authentication failed")
	verifier := &recordingSearchPrivacyVerifier{}
	gate := searchPrivacyAuthorizationGate{
		required:          true,
		verifier:          verifier,
		requesterResolver: staticSearchRequesterResolver{err: requesterErr},
	}
	r := protectedSearchRequest(http.MethodPost)
	r.Header.Set(searchPrivacyAuthorizationHeader, "psc_test")
	if err := gate.Verify(r); !errors.Is(err, requesterErr) {
		t.Fatalf("requester error = %v, want %v", err, requesterErr)
	}
	if verifier.calls != 0 {
		t.Fatalf("verifier calls = %d, want 0", verifier.calls)
	}

	wantErr := errors.New("revoked capability")
	verifier.err = wantErr
	gate = requiredSearchPrivacyGate(verifier)
	if err := gate.Verify(r); !errors.Is(err, wantErr) {
		t.Fatalf("verify error = %v, want %v", err, wantErr)
	}
}
