package main

import (
	"context"
	"errors"
	"testing"
)

type recordingPrivacyReferenceClient struct {
	request privacyReferenceVerificationRequest
	calls   int
	err     error
}

func (c *recordingPrivacyReferenceClient) VerifyReference(
	_ context.Context,
	request privacyReferenceVerificationRequest,
) error {
	c.calls++
	c.request = request
	return c.err
}

func TestPrivacyShieldReferenceVerifierBuildsMinimalSingleUseVerificationRequest(t *testing.T) {
	client := &recordingPrivacyReferenceClient{}
	verifier := privacyShieldReferenceVerifier{client: client}
	ctx := searchPrivacyAuthorizationContext{
		RequesterID:    "goreecloud-index",
		Resource:       "goreecloud.search.query",
		Operation:      "search.query",
		Purpose:        "internet_search",
		ProcessingZone: "private_goreecloud",
		Destination:    "https://search.goreecloud.com",
		RetentionMode:  "none",
	}

	if err := verifier.VerifySearchCapability(context.Background(), "psc_operation", ctx); err != nil {
		t.Fatalf("VerifySearchCapability = %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("client calls = %d, want 1", client.calls)
	}
	want := privacyReferenceVerificationRequest{
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

func TestPrivacyShieldReferenceVerifierPropagatesAuthorityRejection(t *testing.T) {
	wantErr := errors.New("CAPABILITY_ALREADY_CONSUMED")
	client := &recordingPrivacyReferenceClient{err: wantErr}
	verifier := privacyShieldReferenceVerifier{client: client}

	err := verifier.VerifySearchCapability(
		context.Background(),
		"psc_operation",
		searchPrivacyAuthorizationContext{
			RequesterID:    "goreecloud-browser",
			Resource:       "goreecloud.search.query",
			Operation:      "search.query",
			Purpose:        "internet_search",
			ProcessingZone: "private_goreecloud",
			Destination:    "https://search.goreecloud.com",
			RetentionMode:  "none",
		},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("VerifySearchCapability error = %v, want %v", err, wantErr)
	}
}
