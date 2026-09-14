package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"unicode"
)

const maxPrivacyAuthorizationReferenceBytes = 512

var (
	errPrivacyAuthorizationVerifierUnavailable  = errors.New("Privacy Shield authorization verifier is unavailable")
	errPrivacyRequesterResolverUnavailable      = errors.New("authenticated Search requester resolver is unavailable")
	errPrivacyRequesterIdentityRequired         = errors.New("authenticated Search requester identity is required")
	errPrivacyAuthorizationReferenceRequired    = errors.New("Privacy Shield capability reference is required")
	errPrivacyAuthorizationReferenceAmbiguous   = errors.New("Privacy Shield capability reference must be specified once")
	errPrivacyAuthorizationReferenceInvalid     = errors.New("Privacy Shield capability reference is invalid")
	errPrivacyAuthorizationMethodRequired        = errors.New("Privacy Shield protected Search requires POST")
	errPrivacyAuthorizationMediaTypeRequired     = errors.New("Privacy Shield protected Search requires application/json")
)

type searchPrivacyAuthorizationContext struct {
	RequesterID    string
	Resource       string
	Operation      string
	Purpose        string
	ProcessingZone string
	Destination    string
	RetentionMode  string
}

type searchPrivacyAuthorizationVerifier interface {
	VerifySearchCapability(
		ctx context.Context,
		capabilityReference string,
		authorizationContext searchPrivacyAuthorizationContext,
	) error
}

type searchRequesterIdentityResolver interface {
	ResolveSearchRequester(r *http.Request) (string, error)
}

// searchPrivacyAuthorizationGate is the Search-side enforcement boundary for
// the capability-token reference published in the Search capability contract.
//
// Current Development runtime wires the gate with required=false and advertises
// not_enforced_development. A future production runtime must construct this gate
// with required=true, a real Privacy Shield verifier, and an authenticated
// requester-identity resolver before the advertised enforcement state can change
// to required. The requester resolver must derive identity from an authenticated
// runtime/transport boundary; arbitrary client-provided identity headers are not
// sufficient authority.
//
// Required mode also owns the private transport invariant: authorization cannot
// make a query-bearing GET request acceptable. Protected first-party Search is
// POST-only with an application/json body so query text does not need to appear
// in the request URL.
type searchPrivacyAuthorizationGate struct {
	required          bool
	verifier          searchPrivacyAuthorizationVerifier
	requesterResolver searchRequesterIdentityResolver
}

func (g searchPrivacyAuthorizationGate) Enforced() bool {
	return g.required && g.verifier != nil && g.requesterResolver != nil
}

func (g searchPrivacyAuthorizationGate) Verify(r *http.Request) error {
	if !g.required {
		return nil
	}
	if g.verifier == nil {
		return errPrivacyAuthorizationVerifierUnavailable
	}
	if g.requesterResolver == nil {
		return errPrivacyRequesterResolverUnavailable
	}

	values := r.Header.Values(searchPrivacyAuthorizationHeader)
	if len(values) == 0 {
		return errPrivacyAuthorizationReferenceRequired
	}
	if len(values) != 1 {
		return errPrivacyAuthorizationReferenceAmbiguous
	}
	reference := strings.TrimSpace(values[0])
	if reference == "" {
		return errPrivacyAuthorizationReferenceRequired
	}
	if !validPrivacyAuthorizationReference(reference) {
		return errPrivacyAuthorizationReferenceInvalid
	}

	if r.Method != http.MethodPost {
		return errPrivacyAuthorizationMethodRequired
	}
	mediaType := strings.TrimSpace(strings.SplitN(r.Header.Get("Content-Type"), ";", 2)[0])
	if mediaType != "application/json" {
		return errPrivacyAuthorizationMediaTypeRequired
	}

	requesterID, err := g.requesterResolver.ResolveSearchRequester(r)
	if err != nil {
		return err
	}
	requesterID = strings.TrimSpace(requesterID)
	if requesterID == "" {
		return errPrivacyRequesterIdentityRequired
	}

	return g.verifier.VerifySearchCapability(
		r.Context(),
		reference,
		searchPrivacyAuthorizationContext{
			RequesterID:    requesterID,
			Resource:       "goreecloud.search.query",
			Operation:      "search.query",
			Purpose:        "internet_search",
			ProcessingZone: "private_goreecloud",
			Destination:    "https://search.goreecloud.com",
			RetentionMode:  "none",
		},
	)
}

func validPrivacyAuthorizationReference(reference string) bool {
	if len(reference) <= len("psc_") || len(reference) > maxPrivacyAuthorizationReferenceBytes {
		return false
	}
	if !strings.HasPrefix(reference, "psc_") {
		return false
	}
	return strings.IndexFunc(reference, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) == -1
}
