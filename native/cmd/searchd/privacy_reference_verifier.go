package main

import "context"

const (
	searchPrivacyVerificationContractVersion = 1
	searchPrivacyVerificationConsumerID       = "goreecloud-search"
)

type privacyReferenceVerificationExpected struct {
	RequesterID    string
	ResourceID     string
	Purpose        string
	Operation      string
	ProcessingZone string
	Destination    string
	RetentionMode  string
}

type privacyReferenceVerificationRequest struct {
	ContractVersion     int
	ConsumerID          string
	CapabilityReference string
	Expected            privacyReferenceVerificationExpected
	Consume             bool
}

type privacyReferenceVerificationClient interface {
	VerifyReference(ctx context.Context, request privacyReferenceVerificationRequest) error
}

// privacyShieldReferenceVerifier adapts Search's narrow authorization gate to
// the authority-owned Privacy Shield capability-reference verification service.
// The concrete IPC/network client remains injected: Search never receives
// Privacy Shield signing keys and never interprets the signed bearer token.
//
// The verification envelope has its own version so Search can fail closed when
// the authority-side IPC contract changes independently from token format.
//
// Search uses consume=true because one remote query is one authorization use.
// This allows Privacy Shield to enforce single-use capabilities and replay state
// without Search owning that authority. Requester identity comes from the
// gate's authenticated requester resolver, never from an arbitrary HTTP field.
type privacyShieldReferenceVerifier struct {
	client privacyReferenceVerificationClient
}

func (v privacyShieldReferenceVerifier) VerifySearchCapability(
	ctx context.Context,
	capabilityReference string,
	authorizationContext searchPrivacyAuthorizationContext,
) error {
	return v.client.VerifyReference(ctx, privacyReferenceVerificationRequest{
		ContractVersion:     searchPrivacyVerificationContractVersion,
		ConsumerID:          searchPrivacyVerificationConsumerID,
		CapabilityReference: capabilityReference,
		Expected: privacyReferenceVerificationExpected{
			RequesterID:    authorizationContext.RequesterID,
			ResourceID:     authorizationContext.Resource,
			Purpose:        authorizationContext.Purpose,
			Operation:      authorizationContext.Operation,
			ProcessingZone: authorizationContext.ProcessingZone,
			Destination:    authorizationContext.Destination,
			RetentionMode:  authorizationContext.RetentionMode,
		},
		Consume: true,
	})
}
