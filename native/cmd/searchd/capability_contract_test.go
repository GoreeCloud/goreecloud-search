package main

import "testing"

func TestSearchCapabilityEvidencePublishesConsumerBounds(t *testing.T) {
	evidence := searchCapabilityEvidence()
	if len(evidence) != 1 {
		t.Fatalf("capability evidence count = %d, want 1", len(evidence))
	}

	query := evidence[0]
	if query.ID != "search.query" {
		t.Fatalf("capability id = %q, want search.query", query.ID)
	}
	if query.ContractVersion != apiVersion {
		t.Fatalf("contract version = %q, want %q", query.ContractVersion, apiVersion)
	}
	if query.Endpoint != "/api/v1/search" {
		t.Fatalf("endpoint = %q, want /api/v1/search", query.Endpoint)
	}
	if query.MaxResults != maxAPISearchResults {
		t.Fatalf("max results = %d, want %d", query.MaxResults, maxAPISearchResults)
	}
	if !query.Authoritative || !query.Current {
		t.Fatalf("search.query capability must remain authoritative and current")
	}
	if query.ProductionAccepted {
		t.Fatalf("development capability must not claim production acceptance")
	}
}
