package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"unicode"
)

const (
	searchPrivacyDirectServiceAudience = "goreecloud-privacy-shield"
	searchPrivacyDirectServiceScope    = "privacy.capability-reference.consume"
	maxIdentityDirectServiceTokenBytes = 16 * 1024
)

var (
	errIdentityDirectServiceCredentialSourceUnavailable = errors.New("GoreeCloud Identity direct-service credential source is unavailable")
	errIdentityDirectServiceCredentialInvalid           = errors.New("GoreeCloud Identity direct-service credential is invalid for Privacy Shield")
	errIdentityDirectServiceAuthorizationHeaderPresent  = errors.New("Privacy Shield request already contains an Authorization header")
)

// identityDirectServiceCredential is minimized metadata returned alongside an
// opaque Identity-issued bearer credential. Search never interprets JWT claims
// or receives Identity signing keys; the target Privacy Shield runtime remains
// responsible for verifying the credential against the authoritative Identity
// contract/JWKS before deriving authenticatedConsumerId.
type identityDirectServiceCredential struct {
	BearerToken string
	ServiceID   string
	Audience    string
	Scopes      []string
}

// identityDirectServiceCredentialSource is the intentionally injected seam for
// the eventual GoreeCloud Identity runtime. Implementations must obtain a fresh
// short-lived credential from Identity; Search does not mint, self-sign, cache,
// or transform service credentials locally.
type identityDirectServiceCredentialSource interface {
	IssueDirectServiceCredential(
		ctx context.Context,
		audience string,
		scopes []string,
	) (identityDirectServiceCredential, error)
}

// identityDirectServiceRoundTripper authenticates Search as a service caller
// to Privacy Shield. This is service authentication only. It does not resolve
// the end-user/requester identity carried in the Privacy Shield operation and
// it does not create Privacy Shield authorization by itself.
type identityDirectServiceRoundTripper struct {
	base   http.RoundTripper
	source identityDirectServiceCredentialSource
}

func newIdentityDirectServiceRoundTripper(
	base http.RoundTripper,
	source identityDirectServiceCredentialSource,
) (*identityDirectServiceRoundTripper, error) {
	if base == nil || source == nil {
		return nil, errIdentityDirectServiceCredentialSourceUnavailable
	}
	return &identityDirectServiceRoundTripper{base: base, source: source}, nil
}

func (t *identityDirectServiceRoundTripper) AuthenticatedConsumerID() string {
	if t == nil || t.source == nil || t.base == nil {
		return ""
	}
	return searchPrivacyVerificationConsumerID
}

func (t *identityDirectServiceRoundTripper) RoundTrip(
	request *http.Request,
) (*http.Response, error) {
	if t == nil || t.source == nil || t.base == nil || request == nil {
		return nil, errIdentityDirectServiceCredentialSourceUnavailable
	}
	if strings.TrimSpace(request.Header.Get("Authorization")) != "" {
		return nil, errIdentityDirectServiceAuthorizationHeaderPresent
	}

	credential, err := t.source.IssueDirectServiceCredential(
		request.Context(),
		searchPrivacyDirectServiceAudience,
		[]string{searchPrivacyDirectServiceScope},
	)
	if err != nil {
		return nil, err
	}
	if err := validateSearchPrivacyDirectServiceCredential(credential); err != nil {
		return nil, err
	}

	forwarded := request.Clone(request.Context())
	forwarded.Header = request.Header.Clone()
	forwarded.Header.Set("Authorization", "Bearer "+credential.BearerToken)
	return t.base.RoundTrip(forwarded)
}

func validateSearchPrivacyDirectServiceCredential(
	credential identityDirectServiceCredential,
) error {
	if strings.TrimSpace(credential.ServiceID) != searchPrivacyVerificationConsumerID {
		return errIdentityDirectServiceCredentialInvalid
	}
	if strings.TrimSpace(credential.Audience) != searchPrivacyDirectServiceAudience {
		return errIdentityDirectServiceCredentialInvalid
	}
	if len(credential.Scopes) != 1 || credential.Scopes[0] != searchPrivacyDirectServiceScope {
		return errIdentityDirectServiceCredentialInvalid
	}

	token := credential.BearerToken
	if token == "" || len(token) > maxIdentityDirectServiceTokenBytes || strings.TrimSpace(token) != token {
		return errIdentityDirectServiceCredentialInvalid
	}
	for _, character := range token {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return errIdentityDirectServiceCredentialInvalid
		}
	}
	return nil
}
