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
		PreferredMethod              string `json:"preferred_method"`
		PreferredQueryTransport      string `json:"preferred_query_transport"`
		RequestMediaType             string `json:"request_media_type"`
		ResponseMediaType            string `json:"response_media_type"`
		PrivacyAuthorizationRequired bool   `json:"privacy_authorization_required"`
		MaxRequestBytes              int    `json:"max_request_bytes"`
	}
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("unmarshal capability evidence: %v", err)
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
	if document.MaxRequestBytes != maxSearchAPIRequestBytes {
		t.Fatalf("max request bytes = %d, want %d", document.MaxRequestBytes, maxSearchAPIRequestBytes)
	}
}
