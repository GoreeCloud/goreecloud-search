package main

import "encoding/json"

const (
	preferredSearchQueryTransport            = "json_body"
	searchCapabilityDiscoveryEndpoint        = "/api/v1/status"
	searchCapabilityDiscoveryCollection      = "capability_evidence"
	searchPrivacyAuthorizationScheme         = "privacy_shield_capability_token_reference"
	searchPrivacyAuthorizationHeader         = "X-GoreeCloud-Privacy-Capability"
	searchPrivacyAuthorizationEnforcementDev = "not_enforced_development"
)

// MarshalJSON makes the machine-readable Search capability explicit enough for
// independent consumers to validate discovery, transport, and privacy
// requirements before delegating a query. GET remains available for
// compatibility, but the preferred inter-application contract is POST with a
// JSON body so query text does not need to appear in a URL.
//
// Privacy Shield authorization is part of the intended contract, but this
// Development service does not yet enforce the capability-token reference at
// the server boundary. The evidence says so explicitly; production consumers
// must reject this Development evidence rather than inferring enforcement from
// endpoint reachability.
func (e capabilityEvidence) MarshalJSON() ([]byte, error) {
	type capabilityEvidenceJSON struct {
		ID                              string   `json:"id"`
		ContractVersion                 string   `json:"contract_version"`
		Authoritative                   bool     `json:"authoritative"`
		Current                         bool     `json:"current"`
		ProductionAccepted              bool     `json:"production_accepted"`
		Endpoint                        string   `json:"endpoint"`
		DiscoveryEndpoint               string   `json:"discovery_endpoint"`
		DiscoveryCollection             string   `json:"discovery_collection"`
		Methods                         []string `json:"methods,omitempty"`
		PreferredMethod                 string   `json:"preferred_method,omitempty"`
		PreferredQueryTransport         string   `json:"preferred_query_transport,omitempty"`
		RequestMediaType                string   `json:"request_media_type,omitempty"`
		ResponseMediaType               string   `json:"response_media_type,omitempty"`
		PrivacyAuthorizationRequired    bool     `json:"privacy_authorization_required"`
		PrivacyAuthorizationScheme      string   `json:"privacy_authorization_scheme,omitempty"`
		PrivacyAuthorizationHeader      string   `json:"privacy_authorization_header,omitempty"`
		PrivacyAuthorizationEnforcement string   `json:"privacy_authorization_enforcement,omitempty"`
		MaxRequestBytes                 int      `json:"max_request_bytes,omitempty"`
		MaxResults                      int      `json:"max_results,omitempty"`
	}

	return json.Marshal(capabilityEvidenceJSON{
		ID:                              e.ID,
		ContractVersion:                 e.ContractVersion,
		Authoritative:                   e.Authoritative,
		Current:                         e.Current,
		ProductionAccepted:              e.ProductionAccepted,
		Endpoint:                        e.Endpoint,
		DiscoveryEndpoint:               searchCapabilityDiscoveryEndpoint,
		DiscoveryCollection:             searchCapabilityDiscoveryCollection,
		Methods:                         e.Methods,
		PreferredMethod:                 e.PreferredMethod,
		PreferredQueryTransport:         preferredSearchQueryTransport,
		RequestMediaType:                "application/json",
		ResponseMediaType:               "application/json",
		PrivacyAuthorizationRequired:    true,
		PrivacyAuthorizationScheme:      searchPrivacyAuthorizationScheme,
		PrivacyAuthorizationHeader:      searchPrivacyAuthorizationHeader,
		PrivacyAuthorizationEnforcement: searchPrivacyAuthorizationEnforcementDev,
		MaxRequestBytes:                 maxSearchAPIRequestBytes,
		MaxResults:                      e.MaxResults,
	})
}
