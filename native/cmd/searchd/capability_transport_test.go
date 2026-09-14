package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestSearchCapabilityJSONPublishesPreferredPrivateTransport(t *testing.T) {
	evidence := searchCapabilityEvidence()
	if len(evidence) != 1 {
		t.Fatalf("capability evidence count = %d, want 1", len(evidence))
	}

	encoded, err := json.Marshal(evidence[0])
	if err != nil {
		t.Fatalf("marshal capability evidence: %v", err)
	}

	var document struct {
		DiscoveryEndpoint               string `json:"discovery_endpoint"`
		DiscoveryCollection             string `json:"discovery_collection"`
		PreferredMethod                 string `json:"preferred_method"`
		PreferredQueryTransport         string `json:"preferred_query_transport"`
		RequestMediaType                string `json:"request_media_type"`
		ResponseMediaType               string `json:"response_media_type"`
		PrivacyAuthorizationRequired    bool   `json:"privacy_authorization_required"`
		PrivacyAuthorizationScheme      string `json:"privacy_authorization_scheme"`
		PrivacyAuthorizationHeader      string `json:"privacy_authorization_header"`
		PrivacyAuthorizationEnforcement string `json:"privacy_authorization_enforcement"`
		MaxRequestBytes                 int    `json:"max_request_bytes"`
	}
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("unmarshal capability evidence: %v", err)
	}

	if document.DiscoveryEndpoint != searchCapabilityDiscoveryEndpoint {
		t.Fatalf("discovery endpoint = %q, want %q", document.DiscoveryEndpoint, searchCapabilityDiscoveryEndpoint)
	}
	if document.DiscoveryCollection != searchCapabilityDiscoveryCollection {
		t.Fatalf("discovery collection = %q, want %q", document.DiscoveryCollection, searchCapabilityDiscoveryCollection)
	}
	if document.PreferredMethod != http.MethodPost {
		t.Fatalf("preferred method = %q, want POST", document.PreferredMethod)
	}
	if document.PreferredQueryTransport != preferredSearchQueryTransport {
		t.Fatalf("preferred query transport = %q, want %q", document.PreferredQueryTransport, preferredSearchQueryTransport)
	}
	if document.RequestMediaType != "application/json" || document.ResponseMediaType != "application/json" {
		t.Fatalf("media types = request %q response %q, want application/json", document.RequestMediaType, document.ResponseMediaType)
	}
	if !document.PrivacyAuthorizationRequired {
		t.Fatal("privacy authorization requirement must be published")
	}
	if document.PrivacyAuthorizationScheme != searchPrivacyAuthorizationScheme {
		t.Fatalf("privacy authorization scheme = %q, want %q", document.PrivacyAuthorizationScheme, searchPrivacyAuthorizationScheme)
	}
	if document.PrivacyAuthorizationHeader != searchPrivacyAuthorizationHeader {
		t.Fatalf("privacy authorization header = %q, want %q", document.PrivacyAuthorizationHeader, searchPrivacyAuthorizationHeader)
	}
	if document.PrivacyAuthorizationEnforcement != searchPrivacyAuthorizationEnforcementDev {
		t.Fatalf("privacy authorization enforcement = %q, want %q", document.PrivacyAuthorizationEnforcement, searchPrivacyAuthorizationEnforcementDev)
	}
	if document.MaxRequestBytes != maxSearchAPIRequestBytes {
		t.Fatalf("max request bytes = %d, want %d", document.MaxRequestBytes, maxSearchAPIRequestBytes)
	}
}
