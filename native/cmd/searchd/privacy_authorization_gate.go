package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

var (
	errPrivacyAuthorizationVerifierUnavailable = errors.New("Privacy Shield authorization verifier is unavailable")
	errPrivacyAuthorizationReferenceRequired   = errors.New("Privacy Shield capability reference is required")
	errPrivacyAuthorizationReferenceAmbiguous  = errors.New("Privacy Shield capability reference must be specified once")
	errPrivacyAuthorizationReferenceInvalid    = errors.New("Privacy Shield capability reference is invalid")
)

type searchPrivacyAuthorizationContext struct {
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

// searchPrivacyAuthorizationGate is the Search-side enforcement boundary for
// the capability-token reference published in the Search capability contract.
//
// Current Development runtime wires the gate with required=false and advertises
// not_enforced_development. A future production runtime must construct this gate
// with required=true and a real Privacy Shield verifier before the advertised
// enforcement state can change to required.
type searchPrivacyAuthorizationGate struct {
	required bool
	verifier searchPrivacyAuthorizationVerifier
}

func (g searchPrivacyAuthorizationGate) Enforced() bool {
	return g.required && g.verifier != nil
}

func (g searchPrivacyAuthorizationGate) Verify(r *http.Request) error {
	if !g.required {
		return nil
	}
	if g.verifier == nil {
		return errPrivacyAuthorizationVerifierUnavailable
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
	if !strings.HasPrefix(reference, "psc_") || len(reference) <= len("psc_") {
		return errPrivacyAuthorizationReferenceInvalid
	}

	return g.verifier.VerifySearchCapability(
		r.Context(),
		reference,
		searchPrivacyAuthorizationContext{
			Resource:       "goreecloud.search.query",
			Operation:      "search.query",
			Purpose:        "internet_search",
			ProcessingZone: "private_goreecloud",
			Destination:    "https://search.goreecloud.com",
			RetentionMode:  "none",
		},
	)
}
