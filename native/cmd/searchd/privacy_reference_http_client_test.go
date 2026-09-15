package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type testAuthenticatedPrivacyShieldTransport struct {
	consumerID string
	roundTrip  func(*http.Request) (*http.Response, error)
}

func (t testAuthenticatedPrivacyShieldTransport) AuthenticatedConsumerID() string {
	return t.consumerID
}

func (t testAuthenticatedPrivacyShieldTransport) RoundTrip(
	request *http.Request,
) (*http.Response, error) {
	return t.roundTrip(request)
}

func TestPrivacyReferenceHTTPClientRejectsUntrustedEndpointAndIdentity(t *testing.T) {
	transport := testAuthenticatedPrivacyShieldTransport{
		consumerID: searchPrivacyVerificationConsumerID,
		roundTrip: func(*http.Request) (*http.Response, error) {
			t.Fatal("transport must not be called")
			return nil, nil
		},
	}
	if _, err := newPrivacyReferenceHTTPClient(
		"http://privacy.internal.example/v1/capabilities/verify",
		transport,
	); err != errPrivacyReferenceHTTPEndpointInvalid {
		t.Fatalf("non-loopback cleartext endpoint error = %v", err)
	}

	wrongIdentity := testAuthenticatedPrivacyShieldTransport{
		consumerID: "goreecloud-browser",
		roundTrip:  transport.roundTrip,
	}
	if _, err := newPrivacyReferenceHTTPClient(
		"https://privacy.goreecloud.test/v1/capabilities/verify",
		wrongIdentity,
	); err != errPrivacyReferenceHTTPTransportUnauthenticated {
		t.Fatalf("wrong transport identity error = %v", err)
	}
}

func TestPrivacyReferenceHTTPClientUsesBoundedPrivateJSONContract(t *testing.T) {
	var observedBody string
	transport := testAuthenticatedPrivacyShieldTransport{
		consumerID: searchPrivacyVerificationConsumerID,
		roundTrip: func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodPost {
				t.Fatalf("method = %q", request.Method)
			}
			if request.Header.Get("Content-Type") != "application/json" {
				t.Fatalf("content type = %q", request.Header.Get("Content-Type"))
			}
			if request.Header.Get("Cache-Control") != "no-store" {
				t.Fatalf("cache control = %q", request.Header.Get("Cache-Control"))
			}
			encoded, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatalf("read request body: %v", err)
			}
			observedBody = string(encoded)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body: io.NopCloser(strings.NewReader(
					`{"contract_version":1,"authorized":true,"capability_reference":"psc_test","constraints":{"processing_zone":"private_goreecloud","destination":"https://search.goreecloud.com","retention_mode":"none"}}`,
				)),
			}, nil
		},
	}
	client, err := newPrivacyReferenceHTTPClient(
		"http://127.0.0.1:8787/v1/capabilities/verify",
		transport,
	)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	response, err := client.VerifyReference(context.Background(), privacyReferenceVerificationRequest{
		ContractVersion:     searchPrivacyVerificationContractVersion,
		ConsumerID:          searchPrivacyVerificationConsumerID,
		CapabilityReference: "psc_test",
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
	})
	if err != nil {
		t.Fatalf("verify reference: %v", err)
	}
	if !response.Authorized || response.CapabilityReference != "psc_test" {
		t.Fatalf("unexpected response: %+v", response)
	}
	if !strings.Contains(observedBody, `"consumer_id":"goreecloud-search"`) {
		t.Fatalf("request body missing bound consumer identity: %s", observedBody)
	}
	if strings.Contains(observedBody, "Authorization") || strings.Contains(observedBody, "Bearer") {
		t.Fatalf("verification envelope must not carry bearer credentials: %s", observedBody)
	}
}

func TestPrivacyReferenceHTTPClientRejectsConsumerMismatchBeforeTransport(t *testing.T) {
	calls := 0
	transport := testAuthenticatedPrivacyShieldTransport{
		consumerID: searchPrivacyVerificationConsumerID,
		roundTrip: func(*http.Request) (*http.Response, error) {
			calls++
			return nil, nil
		},
	}
	client, err := newPrivacyReferenceHTTPClient(
		"https://privacy.goreecloud.test/v1/capabilities/verify",
		transport,
	)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	_, err = client.VerifyReference(context.Background(), privacyReferenceVerificationRequest{
		ContractVersion: searchPrivacyVerificationContractVersion,
		ConsumerID:      "goreecloud-index",
	})
	if err != errPrivacyReferenceHTTPConsumerMismatch {
		t.Fatalf("consumer mismatch error = %v", err)
	}
	if calls != 0 {
		t.Fatalf("transport calls = %d, want 0", calls)
	}
}

func TestPrivacyReferenceHTTPClientRejectsUnboundedOrNonJSONResponse(t *testing.T) {
	cases := []struct {
		name        string
		contentType string
		body        string
	}{
		{name: "wrong media type", contentType: "text/plain", body: `{}`},
		{
			name:        "oversized",
			contentType: "application/json",
			body:        strings.Repeat("x", maxPrivacyReferenceVerificationResponseBytes+1),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			transport := testAuthenticatedPrivacyShieldTransport{
				consumerID: searchPrivacyVerificationConsumerID,
				roundTrip: func(*http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Header: http.Header{
							"Content-Type": []string{testCase.contentType},
						},
						Body: io.NopCloser(strings.NewReader(testCase.body)),
					}, nil
				},
			}
			client, err := newPrivacyReferenceHTTPClient(
				"https://privacy.goreecloud.test/v1/capabilities/verify",
				transport,
			)
			if err != nil {
				t.Fatalf("create client: %v", err)
			}
			_, err = client.VerifyReference(context.Background(), privacyReferenceVerificationRequest{
				ContractVersion:     searchPrivacyVerificationContractVersion,
				ConsumerID:          searchPrivacyVerificationConsumerID,
				CapabilityReference: "psc_test",
			})
			if err != errPrivacyReferenceHTTPResponseInvalid {
				t.Fatalf("response error = %v", err)
			}
		})
	}
}
