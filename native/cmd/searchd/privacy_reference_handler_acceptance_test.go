package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func requiredReferenceBackedSearchServer(
	requesterID string,
	client privacyReferenceVerificationClient,
) server {
	return server{
		privacyAuthorizationGate: searchPrivacyAuthorizationGate{
			required: true,
			verifier: privacyShieldReferenceVerifier{
				client: client,
			},
			requesterResolver: staticSearchRequesterResolver{
				requesterID: requesterID,
			},
		},
	}
}

func TestSearchAPIRejectsPrivacyShieldConstraintMismatchBeforeSearch(t *testing.T) {
	reference := "psc_constraint-mismatch"
	ctx := searchAuthorizationTestContext()
	response := authorizedPrivacyReferenceResponse(reference, ctx)
	response.Constraints.Destination = "https://other.example"
	client := &recordingPrivacyReferenceClient{response: response}
	app := requiredReferenceBackedSearchServer(ctx.RequesterID, client)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/search",
		strings.NewReader(`{"query":"must not execute","category":"general","limit":1}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(searchPrivacyAuthorizationHeader, reference)
	responseRecorder := httptest.NewRecorder()

	app.searchAPI(responseRecorder, request)

	if responseRecorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", responseRecorder.Code, http.StatusForbidden, responseRecorder.Body.String())
	}
	if client.calls != 1 {
		t.Fatalf("Privacy Shield client calls = %d, want 1", client.calls)
	}
	if !client.request.Consume {
		t.Fatal("handler verification request must preserve consume=true")
	}
	if client.request.Expected.RequesterID != ctx.RequesterID {
		t.Fatalf("verified requester = %q, want %q", client.request.Expected.RequesterID, ctx.RequesterID)
	}
	body := responseRecorder.Body.String()
	if !strings.Contains(body, "Privacy Shield authorization is required") {
		t.Fatalf("body = %q", body)
	}
	if strings.Contains(body, "other.example") || strings.Contains(body, "constraints mismatch") {
		t.Fatalf("constraint details leaked in body: %q", body)
	}
}

func TestSearchAPIAuthorityConsumedCapabilityRejectionIsOpaque(t *testing.T) {
	reference := "psc_consumed"
	ctx := searchAuthorizationTestContext()
	authorityDetail := errors.New("CAPABILITY_ALREADY_CONSUMED replay state 42")
	client := &recordingPrivacyReferenceClient{err: authorityDetail}
	app := requiredReferenceBackedSearchServer(ctx.RequesterID, client)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/search",
		strings.NewReader(`{"query":"must not execute","category":"general","limit":1}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(searchPrivacyAuthorizationHeader, reference)
	responseRecorder := httptest.NewRecorder()

	app.searchAPI(responseRecorder, request)

	if responseRecorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", responseRecorder.Code, http.StatusForbidden, responseRecorder.Body.String())
	}
	if client.calls != 1 {
		t.Fatalf("Privacy Shield client calls = %d, want 1", client.calls)
	}
	if !client.request.Consume {
		t.Fatal("handler verification request must preserve consume=true")
	}
	if client.request.CapabilityReference != reference {
		t.Fatalf("capability reference = %q, want %q", client.request.CapabilityReference, reference)
	}
	body := responseRecorder.Body.String()
	if !strings.Contains(body, "Privacy Shield authorization is required") {
		t.Fatalf("body = %q", body)
	}
	if strings.Contains(body, "CAPABILITY_ALREADY_CONSUMED") || strings.Contains(body, "replay state") {
		t.Fatalf("authority replay detail leaked in body: %q", body)
	}
}
