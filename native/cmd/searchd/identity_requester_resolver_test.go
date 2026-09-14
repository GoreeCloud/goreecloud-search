package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type recordingIdentitySearchRequesterVerifier struct {
	calls      int
	credential string
	verified   identityVerifiedSearchRequester
	err        error
}

func (v *recordingIdentitySearchRequesterVerifier) VerifySearchRequester(
	_ context.Context,
	bearerCredential string,
) (identityVerifiedSearchRequester, error) {
	v.calls++
	v.credential = bearerCredential
	return v.verified, v.err
}

func TestIdentityRequesterResolverRequiresInjectedVerifier(t *testing.T) {
	if _, err := newIdentityBearerSearchRequesterResolver(nil); !errors.Is(err, errIdentityRequesterVerifierUnavailable) {
		t.Fatalf("nil verifier error = %v", err)
	}
}

func TestIdentityRequesterResolverVerifiesOpaqueBearerAndReturnsPrincipal(t *testing.T) {
	verifier := &recordingIdentitySearchRequesterVerifier{
		verified: identityVerifiedSearchRequester{PrincipalID: "identity:subject:browser-user"},
	}
	resolver, err := newIdentityBearerSearchRequesterResolver(verifier)
	if err != nil {
		t.Fatalf("create resolver: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/search", nil)
	request.Header.Set("Authorization", "Bearer opaque.identity.credential")
	request.Header.Set("X-GoreeCloud-Requester", "attacker-selected-principal")

	principal, err := resolver.ResolveSearchRequester(request)
	if err != nil {
		t.Fatalf("resolve requester: %v", err)
	}
	if principal != "identity:subject:browser-user" {
		t.Fatalf("principal = %q", principal)
	}
	if verifier.calls != 1 || verifier.credential != "opaque.identity.credential" {
		t.Fatalf("verifier calls/credential = %d/%q", verifier.calls, verifier.credential)
	}
}

func TestIdentityRequesterResolverRejectsMissingDuplicateAndMalformedAuthorization(t *testing.T) {
	cases := []struct {
		name      string
		configure func(*http.Request)
		want      error
	}{
		{
			name: "missing",
			configure: func(_ *http.Request) {},
			want: errIdentityRequesterAuthorizationMissing,
		},
		{
			name: "duplicate",
			configure: func(request *http.Request) {
				request.Header.Add("Authorization", "Bearer one")
				request.Header.Add("Authorization", "Bearer two")
			},
			want: errIdentityRequesterAuthorizationAmbiguous,
		},
		{
			name: "wrong scheme",
			configure: func(request *http.Request) {
				request.Header.Set("Authorization", "Basic opaque")
			},
			want: errIdentityRequesterAuthorizationInvalid,
		},
		{
			name: "empty bearer",
			configure: func(request *http.Request) {
				request.Header.Set("Authorization", "Bearer ")
			},
			want: errIdentityRequesterAuthorizationInvalid,
		},
		{
			name: "trim-dependent",
			configure: func(request *http.Request) {
				request.Header.Set("Authorization", " Bearer opaque ")
			},
			want: errIdentityRequesterAuthorizationInvalid,
		},
		{
			name: "credential whitespace",
			configure: func(request *http.Request) {
				request.Header.Set("Authorization", "Bearer opaque credential")
			},
			want: errIdentityRequesterAuthorizationInvalid,
		},
		{
			name: "oversized",
			configure: func(request *http.Request) {
				request.Header.Set("Authorization", "Bearer "+strings.Repeat("x", maxIdentityRequesterBearerBytes+1))
			},
			want: errIdentityRequesterAuthorizationInvalid,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			verifier := &recordingIdentitySearchRequesterVerifier{
				verified: identityVerifiedSearchRequester{PrincipalID: "identity:subject:user"},
			}
			resolver, err := newIdentityBearerSearchRequesterResolver(verifier)
			if err != nil {
				t.Fatalf("create resolver: %v", err)
			}
			request := httptest.NewRequest(http.MethodPost, "/api/v1/search", nil)
			testCase.configure(request)
			_, err = resolver.ResolveSearchRequester(request)
			if !errors.Is(err, testCase.want) {
				t.Fatalf("resolve error = %v, want %v", err, testCase.want)
			}
			if verifier.calls != 0 {
				t.Fatalf("verifier calls = %d, want 0", verifier.calls)
			}
		})
	}
}

func TestIdentityRequesterResolverPropagatesVerifierRejection(t *testing.T) {
	identityRejected := errors.New("Identity requester credential rejected")
	verifier := &recordingIdentitySearchRequesterVerifier{err: identityRejected}
	resolver, err := newIdentityBearerSearchRequesterResolver(verifier)
	if err != nil {
		t.Fatalf("create resolver: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/search", nil)
	request.Header.Set("Authorization", "Bearer opaque.identity.credential")

	_, err = resolver.ResolveSearchRequester(request)
	if !errors.Is(err, identityRejected) {
		t.Fatalf("resolve error = %v", err)
	}
	if verifier.calls != 1 {
		t.Fatalf("verifier calls = %d, want 1", verifier.calls)
	}
}

func TestIdentityRequesterResolverRejectsInvalidVerifiedPrincipal(t *testing.T) {
	invalidPrincipals := []string{
		"",
		" identity:subject:user ",
		"identity:subject:\nuser",
		strings.Repeat("x", maxIdentityRequesterIDBytes+1),
	}

	for _, principal := range invalidPrincipals {
		verifier := &recordingIdentitySearchRequesterVerifier{
			verified: identityVerifiedSearchRequester{PrincipalID: principal},
		}
		resolver, err := newIdentityBearerSearchRequesterResolver(verifier)
		if err != nil {
			t.Fatalf("create resolver: %v", err)
		}
		request := httptest.NewRequest(http.MethodPost, "/api/v1/search", nil)
		request.Header.Set("Authorization", "Bearer opaque.identity.credential")
		_, err = resolver.ResolveSearchRequester(request)
		if !errors.Is(err, errIdentityRequesterPrincipalInvalid) {
			t.Fatalf("principal %q error = %v", principal, err)
		}
	}
}
