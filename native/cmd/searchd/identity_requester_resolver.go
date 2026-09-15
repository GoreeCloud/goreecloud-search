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
	errIdentityRequesterVerifierUnavailable    = errors.New("GoreeCloud Identity requester verifier is unavailable")
	errIdentityRequesterAuthorizationMissing   = errors.New("authenticated requester credential is required")
	errIdentityRequesterAuthorizationAmbiguous = errors.New("authenticated requester credential must be specified once")
	errIdentityRequesterAuthorizationInvalid   = errors.New("authenticated requester credential is invalid")
	errIdentityRequesterApplicationInvalid     = errors.New("verified requester application is invalid")
	errIdentityRequesterPrincipalInvalid       = errors.New("verified requester principal is invalid")
)

// identityVerifiedSearchRequester is minimized Identity-owned verification
// output. Search needs two independent identities from the authenticated native
// application credential: the registered application that is the Privacy Shield
// requester and the user principal bound to the Identity session. Neither value
// is accepted from caller-selected headers. Search receives no Identity signing
// material, reusable session secret, or arbitrary caller-selected claim.
type identityVerifiedSearchRequester struct {
	ApplicationID string
	PrincipalID   string
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
// Shield gate. Privacy Shield capabilities issued to first-party applications
// bind expected.requester_id to the registered application requester, not to
// Search's own service identity and not to the end-user principal. The principal
// remains independently required and validated as evidence that the credential
// represents a user-bound native application session.
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
	if !validIdentityRequesterID(verified.ApplicationID) {
		return "", errIdentityRequesterApplicationInvalid
	}
	if !validIdentityRequesterID(verified.PrincipalID) {
		return "", errIdentityRequesterPrincipalInvalid
	}
	return verified.ApplicationID, nil
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

func validIdentityRequesterID(value string) bool {
	if value == "" || len(value) > maxIdentityRequesterIDBytes {
		return false
	}
	if strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return false
		}
	}
	return true
}
