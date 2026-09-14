package main

import "encoding/json"

const preferredSearchQueryTransport = "json_body"

// MarshalJSON makes the machine-readable Search capability explicit enough for
// independent consumers to validate transport and privacy requirements before
// delegating a query. GET remains available for compatibility, but the
// preferred inter-application contract is POST with a JSON body so query text
// does not need to appear in a URL.
func (e capabilityEvidence) MarshalJSON() ([]byte, error) {
	type capabilityEvidenceJSON struct {
		ID                           string   `json:"id"`
		ContractVersion              string   `json:"contract_version"`
		Authoritative                bool     `json:"authoritative"`
		Current                      bool     `json:"current"`
		ProductionAccepted           bool     `json:"production_accepted"`
		Endpoint                     string   `json:"endpoint"`
		Methods                      []string `json:"methods,omitempty"`
		PreferredMethod              string   `json:"preferred_method,omitempty"`
		PreferredQueryTransport      string   `json:"preferred_query_transport,omitempty"`
		RequestMediaType             string   `json:"request_media_type,omitempty"`
		ResponseMediaType            string   `json:"response_media_type,omitempty"`
		PrivacyAuthorizationRequired bool     `json:"privacy_authorization_required"`
		MaxRequestBytes              int      `json:"max_request_bytes,omitempty"`
		MaxResults                   int      `json:"max_results,omitempty"`
	}

	return json.Marshal(capabilityEvidenceJSON{
		ID:                           e.ID,
		ContractVersion:              e.ContractVersion,
		Authoritative:                e.Authoritative,
		Current:                      e.Current,
		ProductionAccepted:           e.ProductionAccepted,
		Endpoint:                     e.Endpoint,
		Methods:                      e.Methods,
		PreferredMethod:              e.PreferredMethod,
		PreferredQueryTransport:      preferredSearchQueryTransport,
		RequestMediaType:             "application/json",
		ResponseMediaType:            "application/json",
		PrivacyAuthorizationRequired: true,
		MaxRequestBytes:              maxSearchAPIRequestBytes,
		MaxResults:                   e.MaxResults,
	})
}
