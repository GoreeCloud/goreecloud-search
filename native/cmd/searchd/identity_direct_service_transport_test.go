package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type testIdentityDirectServiceCredentialSource struct {
	credential identityDirectServiceCredential
	err        error
	calls      int
	audience   string
	scopes     []string
}

func (s *testIdentityDirectServiceCredentialSource) IssueDirectServiceCredential(
	_ context.Context,
	audience string,
	scopes []string,
) (identityDirectServiceCredential, error) {
	s.calls++
	s.audience = audience
	s.scopes = append([]string(nil), scopes...)
	return s.credential, s.err
}

type testRoundTripper struct {
	calls int
	last  *http.Request
}

func (t *testRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	t.calls++
	t.last = request
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("ok")),
	}, nil
}

func validIdentityDirectServiceCredential() identityDirectServiceCredential {
	return identityDirectServiceCredential{
		BearerToken: "eyJhbGciOiJSUzI1NiJ9.test.signature",
		ServiceID:   searchPrivacyVerificationConsumerID,
		Audience:    searchPrivacyDirectServiceAudience,
		Scopes:      []string{searchPrivacyDirectServiceScope},
	}
}

func TestIdentityDirectServiceTransportAddsOnlyBoundAuthorizationHeader(t *testing.T) {
	source := &testIdentityDirectServiceCredentialSource{
		credential: validIdentityDirectServiceCredential(),
	}
	base := &testRoundTripper{}
	transport, err := newIdentityDirectServiceRoundTripper(base, source)
	if err != nil {
		t.Fatalf("create transport: %v", err)
	}
	if transport.AuthenticatedConsumerID() != searchPrivacyVerificationConsumerID {
		t.Fatalf("authenticated consumer = %q", transport.AuthenticatedConsumerID())
	}

	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"https://privacy.goreecloud.test/v1/capabilities/verify",
		strings.NewReader(`{"consumer_id":"goreecloud-search"}`),
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := transport.RoundTrip(request)
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	response.Body.Close()

	if source.calls != 1 {
		t.Fatalf("credential source calls = %d, want 1", source.calls)
	}
	if source.audience != searchPrivacyDirectServiceAudience {
		t.Fatalf("credential audience = %q", source.audience)
	}
	if len(source.scopes) != 1 || source.scopes[0] != searchPrivacyDirectServiceScope {
		t.Fatalf("credential scopes = %#v", source.scopes)
	}
	if base.calls != 1 {
		t.Fatalf("base transport calls = %d, want 1", base.calls)
	}
	if got := base.last.Header.Get("Authorization"); got != "Bearer "+source.credential.BearerToken {
		t.Fatalf("forwarded authorization = %q", got)
	}
	if request.Header.Get("Authorization") != "" {
		t.Fatal("original request must not be mutated with bearer credential")
	}
}

func TestIdentityDirectServiceTransportRejectsMissingDependencies(t *testing.T) {
	base := &testRoundTripper{}
	source := &testIdentityDirectServiceCredentialSource{
		credential: validIdentityDirectServiceCredential(),
	}
	if _, err := newIdentityDirectServiceRoundTripper(nil, source); err != errIdentityDirectServiceCredentialSourceUnavailable {
		t.Fatalf("nil base error = %v", err)
	}
	if _, err := newIdentityDirectServiceRoundTripper(base, nil); err != errIdentityDirectServiceCredentialSourceUnavailable {
		t.Fatalf("nil source error = %v", err)
	}
}

func TestIdentityDirectServiceTransportRejectsExistingAuthorizationHeader(t *testing.T) {
	source := &testIdentityDirectServiceCredentialSource{
		credential: validIdentityDirectServiceCredential(),
	}
	base := &testRoundTripper{}
	transport, err := newIdentityDirectServiceRoundTripper(base, source)
	if err != nil {
		t.Fatalf("create transport: %v", err)
	}
	request, _ := http.NewRequest(http.MethodPost, "https://privacy.goreecloud.test/verify", nil)
	request.Header.Set("Authorization", "Bearer injected")

	_, err = transport.RoundTrip(request)
	if err != errIdentityDirectServiceAuthorizationHeaderPresent {
		t.Fatalf("authorization-header error = %v", err)
	}
	if source.calls != 0 || base.calls != 0 {
		t.Fatalf("source/base calls = %d/%d, want 0/0", source.calls, base.calls)
	}
}

func TestIdentityDirectServiceTransportRejectsCredentialMetadataMismatch(t *testing.T) {
	cases := []struct {
		name       string
		credential identityDirectServiceCredential
	}{
		{
			name: "wrong service",
			credential: identityDirectServiceCredential{
				BearerToken: "token",
				ServiceID:   "goreecloud-browser",
				Audience:    searchPrivacyDirectServiceAudience,
				Scopes:      []string{searchPrivacyDirectServiceScope},
			},
		},
		{
			name: "wrong audience",
			credential: identityDirectServiceCredential{
				BearerToken: "token",
				ServiceID:   searchPrivacyVerificationConsumerID,
				Audience:    "goreecloud-mesh",
				Scopes:      []string{searchPrivacyDirectServiceScope},
			},
		},
		{
			name: "scope escalation",
			credential: identityDirectServiceCredential{
				BearerToken: "token",
				ServiceID:   searchPrivacyVerificationConsumerID,
				Audience:    searchPrivacyDirectServiceAudience,
				Scopes: []string{
					searchPrivacyDirectServiceScope,
					"privacy.capability-reference.verify",
				},
			},
		},
		{
			name: "whitespace token",
			credential: identityDirectServiceCredential{
				BearerToken: " bearer token ",
				ServiceID:   searchPrivacyVerificationConsumerID,
				Audience:    searchPrivacyDirectServiceAudience,
				Scopes:      []string{searchPrivacyDirectServiceScope},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			source := &testIdentityDirectServiceCredentialSource{credential: testCase.credential}
			base := &testRoundTripper{}
			transport, err := newIdentityDirectServiceRoundTripper(base, source)
			if err != nil {
				t.Fatalf("create transport: %v", err)
			}
			request, _ := http.NewRequest(http.MethodPost, "https://privacy.goreecloud.test/verify", nil)
			_, err = transport.RoundTrip(request)
			if err != errIdentityDirectServiceCredentialInvalid {
				t.Fatalf("credential validation error = %v", err)
			}
			if base.calls != 0 {
				t.Fatalf("base transport calls = %d, want 0", base.calls)
			}
		})
	}
}

func TestIdentityDirectServiceTransportPropagatesCredentialSourceFailure(t *testing.T) {
	identityUnavailable := errors.New("identity runtime unavailable")
	source := &testIdentityDirectServiceCredentialSource{err: identityUnavailable}
	base := &testRoundTripper{}
	transport, err := newIdentityDirectServiceRoundTripper(base, source)
	if err != nil {
		t.Fatalf("create transport: %v", err)
	}
	request, _ := http.NewRequest(http.MethodPost, "https://privacy.goreecloud.test/verify", nil)

	_, err = transport.RoundTrip(request)
	if !errors.Is(err, identityUnavailable) {
		t.Fatalf("credential source error = %v", err)
	}
	if base.calls != 0 {
		t.Fatalf("base transport calls = %d, want 0", base.calls)
	}
}
