package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"unicode"
)

const (
	maxIdentityRequesterBearerBytes = 16 * 1024
	maxIdentityRequesterIDBytes     = 512
)

var (
	errIdentityRequesterVerifierUnavailable = errors.New("GoreeCloud Identity requester verifier is unavailable")
	errIdentityRequesterAuthorizationMissing = errors.New("authenticated requester credential is required")
	errIdentityRequesterAuthorizationAmbiguous = errors.New("authenticated requester credential must be specified once")
	errIdentityRequesterAuthorizationInvalid = errors.New("authenticated requester credential is invalid")
	errIdentityRequesterPrincipalInvalid = errors.New("verified requester principal is invalid")
)

// identityVerifiedSearchRequester is minimized Identity-owned verification
// output. Search receives only the canonical principal identifier needed to bind
// the Privacy Shield operation. It does not receive Identity signing material,
// reusable session secrets, or arbitrary caller-selected authorization claims.
type identityVerifiedSearchRequester struct {
	PrincipalID string
}

// identitySearchRequesterVerifier is the producer-owned verification seam for
// the eventual GoreeCloud Identity runtime. Implementations must verify the
// opaque inbound bearer credential under an accepted Identity registration and
// return only minimized verified requester metadata. Search deliberately does
// not define an audience, scope, token format, JWKS location, or registration
// here because those remain Identity-owned contracts.
type identitySearchRequesterVerifier interface {
	VerifySearchRequester(
		ctx context.Context,
		bearerCredential string,
	) (identityVerifiedSearchRequester, error)
}

// identityBearerSearchRequesterResolver converts one independently verified
// Identity bearer credential into the requester ID used by the Search Privacy
// Shield gate. Caller-controlled identity headers are intentionally ignored.
type identityBearerSearchRequesterResolver struct {
	verifier identitySearchRequesterVerifier
}

func newIdentityBearerSearchRequesterResolver(
	verifier identitySearchRequesterVerifier,
) (*identityBearerSearchRequesterResolver, error) {
	if verifier == nil {
		return nil, errIdentityRequesterVerifierUnavailable
	}
	return &identityBearerSearchRequesterResolver{verifier: verifier}, nil
}

func (r *identityBearerSearchRequesterResolver) ResolveSearchRequester(
	request *http.Request,
) (string, error) {
	if r == nil || r.verifier == nil {
		return "", errIdentityRequesterVerifierUnavailable
	}
	if request == nil {
		return "", errIdentityRequesterAuthorizationMissing
	}

	values := request.Header.Values("Authorization")
	if len(values) == 0 {
		return "", errIdentityRequesterAuthorizationMissing
	}
	if len(values) != 1 {
		return "", errIdentityRequesterAuthorizationAmbiguous
	}

	credential, err := canonicalIdentityBearerCredential(values[0])
	if err != nil {
		return "", err
	}
	verified, err := r.verifier.VerifySearchRequester(request.Context(), credential)
	if err != nil {
		return "", err
	}
	if !validIdentityRequesterPrincipal(verified.PrincipalID) {
		return "", errIdentityRequesterPrincipalInvalid
	}
	return verified.PrincipalID, nil
}

func canonicalIdentityBearerCredential(value string) (string, error) {
	if value == "" || strings.TrimSpace(value) != value {
		return "", errIdentityRequesterAuthorizationInvalid
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(value, prefix) {
		return "", errIdentityRequesterAuthorizationInvalid
	}
	credential := strings.TrimPrefix(value, prefix)
	if credential == "" || len(credential) > maxIdentityRequesterBearerBytes {
		return "", errIdentityRequesterAuthorizationInvalid
	}
	for _, character := range credential {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return "", errIdentityRequesterAuthorizationInvalid
		}
	}
	return credential, nil
}

func validIdentityRequesterPrincipal(principalID string) bool {
	if principalID == "" || len(principalID) > maxIdentityRequesterIDBytes {
		return false
	}
	if strings.TrimSpace(principalID) != principalID {
		return false
	}
	for _, character := range principalID {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}
