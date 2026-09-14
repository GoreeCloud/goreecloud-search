package main

import (
	"context"
	"errors"
)

const (
	searchPrivacyVerificationContractVersion = 1
	searchPrivacyVerificationConsumerID       = "goreecloud-search"
)

var (
	errPrivacyReferenceVerificationClientUnavailable = errors.New("Privacy Shield reference verification client is unavailable")
	errPrivacyReferenceVerificationContractMismatch  = errors.New("Privacy Shield reference verification contract version mismatch")
	errPrivacyReferenceVerificationDenied            = errors.New("Privacy Shield reference verification denied authorization")
	errPrivacyReferenceVerificationReferenceMismatch = errors.New("Privacy Shield reference verification capability reference mismatch")
	errPrivacyReferenceVerificationConstraintMismatch = errors.New("Privacy Shield reference verification constraints mismatch")
)

type privacyReferenceVerificationExpected struct {
	RequesterID    string `json:"requester_id"`
	ResourceID     string `json:"resource_id"`
	Purpose        string `json:"purpose"`
	Operation      string `json:"operation"`
	ProcessingZone string `json:"processing_zone"`
	Destination    string `json:"destination"`
	RetentionMode  string `json:"retention_mode"`
}

type privacyReferenceVerificationRequest struct {
	ContractVersion     int                                  `json:"contract_version"`
	ConsumerID          string                               `json:"consumer_id"`
	CapabilityReference string                               `json:"capability_reference"`
	Expected            privacyReferenceVerificationExpected `json:"expected"`
	Consume             bool                                 `json:"consume"`
}

type privacyReferenceVerificationConstraints struct {
	ProcessingZone string `json:"processing_zone"`
	Destination    string `json:"destination"`
	RetentionMode  string `json:"retention_mode"`
}

type privacyReferenceVerificationResponse struct {
	ContractVersion     int                                     `json:"contract_version"`
	Authorized          bool                                    `json:"authorized"`
	CapabilityReference string                                  `json:"capability_reference"`
	Constraints         privacyReferenceVerificationConstraints `json:"constraints"`
}

type privacyReferenceVerificationClient interface {
	VerifyReference(
		ctx context.Context,
		request privacyReferenceVerificationRequest,
	) (privacyReferenceVerificationResponse, error)
}

// privacyShieldReferenceVerifier adapts Search's narrow authorization gate to
// the authority-owned Privacy Shield capability-reference verification service.
// The concrete IPC/network client remains injected: Search never receives
// Privacy Shield signing keys and never interprets the signed bearer token.
//
// The verification envelope has its own version and pinned JSON field names so
// Search can fail closed when the authority-side IPC contract changes
// independently from token format.
//
// Search uses consume=true because one remote query is one authorization use.
// This allows Privacy Shield to enforce single-use capabilities and replay state
// without Search owning that authority. Requester identity comes from the
// gate's authenticated requester resolver, never from an arbitrary HTTP field.
//
// The transport client is deliberately not trusted to collapse the authority
// response into a success/error bit. Search validates the versioned response,
// positive authorization, echoed opaque reference, and enforceable constraints
// itself before the request is admitted.
type privacyShieldReferenceVerifier struct {
	client privacyReferenceVerificationClient
}

func (v privacyShieldReferenceVerifier) VerifySearchCapability(
	ctx context.Context,
	capabilityReference string,
	authorizationContext searchPrivacyAuthorizationContext,
) error {
	if v.client == nil {
		return errPrivacyReferenceVerificationClientUnavailable
	}

	request := privacyReferenceVerificationRequest{
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
	}

	response, err := v.client.VerifyReference(ctx, request)
	if err != nil {
		return err
	}
	if response.ContractVersion != searchPrivacyVerificationContractVersion {
		return errPrivacyReferenceVerificationContractMismatch
	}
	if !response.Authorized {
		return errPrivacyReferenceVerificationDenied
	}
	if response.CapabilityReference != capabilityReference {
		return errPrivacyReferenceVerificationReferenceMismatch
	}

	expectedConstraints := privacyReferenceVerificationConstraints{
		ProcessingZone: authorizationContext.ProcessingZone,
		Destination:    authorizationContext.Destination,
		RetentionMode:  authorizationContext.RetentionMode,
	}
	if response.Constraints != expectedConstraints {
		return errPrivacyReferenceVerificationConstraintMismatch
	}

	return nil
}
