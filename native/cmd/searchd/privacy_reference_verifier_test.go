package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type recordingPrivacyReferenceClient struct {
	request  privacyReferenceVerificationRequest
	response privacyReferenceVerificationResponse
	calls    int
	err      error
}

func (c *recordingPrivacyReferenceClient) VerifyReference(
	_ context.Context,
	request privacyReferenceVerificationRequest,
) (privacyReferenceVerificationResponse, error) {
	c.calls++
	c.request = request
	return c.response, c.err
}

func authorizedPrivacyReferenceResponse(
	reference string,
	ctx searchPrivacyAuthorizationContext,
) privacyReferenceVerificationResponse {
	return privacyReferenceVerificationResponse{
		ContractVersion:     searchPrivacyVerificationContractVersion,
		Authorized:          true,
		CapabilityReference: reference,
		Constraints: privacyReferenceVerificationConstraints{
			ProcessingZone: ctx.ProcessingZone,
			Destination:    ctx.Destination,
			RetentionMode:  ctx.RetentionMode,
		},
	}
}

func searchAuthorizationTestContext() searchPrivacyAuthorizationContext {
	return searchPrivacyAuthorizationContext{
		RequesterID:    "goreecloud-index",
		Resource:       "goreecloud.search.query",
		Operation:      "search.query",
		Purpose:        "internet_search",
		ProcessingZone: "private_goreecloud",
		Destination:    "https://search.goreecloud.com",
		RetentionMode:  "none",
	}
}

func TestPrivacyShieldReferenceVerifierBuildsVersionedMinimalSingleUseVerificationRequest(t *testing.T) {
	ctx := searchAuthorizationTestContext()
	client := &recordingPrivacyReferenceClient{
		response: authorizedPrivacyReferenceResponse("psc_operation", ctx),
	}
	verifier := privacyShieldReferenceVerifier{client: client}

	if err := verifier.VerifySearchCapability(context.Background(), "psc_operation", ctx); err != nil {
		t.Fatalf("VerifySearchCapability = %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("client calls = %d, want 1", client.calls)
	}
	want := privacyReferenceVerificationRequest{
		ContractVersion:     searchPrivacyVerificationContractVersion,
		ConsumerID:          "goreecloud-search",
		CapabilityReference: "psc_operation",
		Expected: privacyReferenceVerificationExpected{
			RequesterID:    "goreecloud-index",
			ResourceID:     "goreecloud.search.query",
			Purpose:        "internet_search",
			Operation:      "search.query",
			ProcessingZone: "private_goreecloud",
			Destination:    "https://search.goreecloud.com",
			RetentionMode:  "none",
		},
		Consume: true,
	}
	if client.request != want {
		t.Fatalf("verification request = %#v, want %#v", client.request, want)
	}
}

func TestPrivacyShieldReferenceVerificationRequestUsesAuthoritySchemaWireNames(t *testing.T) {
	request := privacyReferenceVerificationRequest{
		ContractVersion:     searchPrivacyVerificationContractVersion,
		ConsumerID:          searchPrivacyVerificationConsumerID,
		CapabilityReference: "psc_operation",
		Expected: privacyReferenceVerificationExpected{
			RequesterID:    "goreecloud-browser",
			ResourceID:     "goreecloud.search.query",
			Purpose:        "internet_search",
			Operation:      "search.query",
			ProcessingZone: "private_goreecloud",
			Destination:    "https://search.goreecloud.com",
			RetentionMode:  "none",
		},
		Consume: true,
	}

	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal verification request: %v", err)
	}
	body := string(encoded)
	for _, expectedField := range []string{
		`"contract_version":1`,
		`"consumer_id":"goreecloud-search"`,
		`"capability_reference":"psc_operation"`,
		`"requester_id":"goreecloud-browser"`,
		`"resource_id":"goreecloud.search.query"`,
		`"processing_zone":"private_goreecloud"`,
		`"retention_mode":"none"`,
		`"consume":true`,
	} {
		if !strings.Contains(body, expectedField) {
			t.Fatalf("verification JSON %s missing %s", body, expectedField)
		}
	}
	for _, forbiddenField := range []string{"ContractVersion", "ConsumerID", "RequesterID", "ResourceID"} {
		if strings.Contains(body, forbiddenField) {
			t.Fatalf("verification JSON leaked Go field name %q: %s", forbiddenField, body)
		}
	}
}

func TestPrivacyShieldReferenceVerifierPropagatesAuthorityRejection(t *testing.T) {
	wantErr := errors.New("CAPABILITY_ALREADY_CONSUMED")
	client := &recordingPrivacyReferenceClient{err: wantErr}
	verifier := privacyShieldReferenceVerifier{client: client}

	err := verifier.VerifySearchCapability(
		context.Background(),
		"psc_operation",
		searchAuthorizationTestContext(),
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("VerifySearchCapability error = %v, want %v", err, wantErr)
	}
}

func TestPrivacyShieldReferenceVerifierFailsClosedWithoutClient(t *testing.T) {
	verifier := privacyShieldReferenceVerifier{}
	err := verifier.VerifySearchCapability(
		context.Background(),
		"psc_operation",
		searchAuthorizationTestContext(),
	)
	if !errors.Is(err, errPrivacyReferenceVerificationClientUnavailable) {
		t.Fatalf("VerifySearchCapability error = %v, want %v", err, errPrivacyReferenceVerificationClientUnavailable)
	}
}

func TestPrivacyShieldReferenceVerifierValidatesAuthorityResponse(t *testing.T) {
	ctx := searchAuthorizationTestContext()
	valid := authorizedPrivacyReferenceResponse("psc_operation", ctx)

	tests := []struct {
		name     string
		mutate   func(*privacyReferenceVerificationResponse)
		wantErr  error
	}{
		{
			name: "contract version",
			mutate: func(response *privacyReferenceVerificationResponse) {
				response.ContractVersion++
			},
			wantErr: errPrivacyReferenceVerificationContractMismatch,
		},
		{
			name: "authorization denied",
			mutate: func(response *privacyReferenceVerificationResponse) {
				response.Authorized = false
			},
			wantErr: errPrivacyReferenceVerificationDenied,
		},
		{
			name: "reference echo",
			mutate: func(response *privacyReferenceVerificationResponse) {
				response.CapabilityReference = "psc_other"
			},
			wantErr: errPrivacyReferenceVerificationReferenceMismatch,
		},
		{
			name: "processing-zone constraint",
			mutate: func(response *privacyReferenceVerificationResponse) {
				response.Constraints.ProcessingZone = "external"
			},
			wantErr: errPrivacyReferenceVerificationConstraintMismatch,
		},
		{
			name: "destination constraint",
			mutate: func(response *privacyReferenceVerificationResponse) {
				response.Constraints.Destination = "https://other.example"
			},
			wantErr: errPrivacyReferenceVerificationConstraintMismatch,
		},
		{
			name: "retention constraint",
			mutate: func(response *privacyReferenceVerificationResponse) {
				response.Constraints.RetentionMode = "session"
			},
			wantErr: errPrivacyReferenceVerificationConstraintMismatch,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := valid
			test.mutate(&response)
			client := &recordingPrivacyReferenceClient{response: response}
			verifier := privacyShieldReferenceVerifier{client: client}

			err := verifier.VerifySearchCapability(context.Background(), "psc_operation", ctx)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("VerifySearchCapability error = %v, want %v", err, test.wantErr)
			}
		})
	}
}
