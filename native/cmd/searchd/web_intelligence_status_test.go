package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/GoreeCloud/goreecloud-search/native/internal/webintelligence"
)

func TestWebIntelligenceStatusUnconfigured(t *testing.T) {
	app := server{}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/api/v1/web-intelligence/status", nil)
	app.webIntelligenceStatus(recorder, request)
	if recorder.Code != 200 {
		t.Fatalf("status = %d", recorder.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if configured, ok := payload["configured"].(bool); !ok || configured {
		t.Fatalf("configured = %#v", payload["configured"])
	}
	if _, exists := payload["snapshot"]; exists {
		t.Fatal("unconfigured response must not expose a synthetic snapshot")
	}
	if exposed, ok := payload["credentials_exposed"].(bool); !ok || exposed {
		t.Fatalf("credentials_exposed = %#v", payload["credentials_exposed"])
	}
}

func TestWebIntelligenceStatusConfiguredSnapshot(t *testing.T) {
	controller, err := webintelligence.NewController([]webintelligence.Provider{
		{ID: "tinyfish-search-v1", Capability: webintelligence.CapabilitySearch, Metered: false, Priority: 10},
		{ID: "tinyfish-research-v1", Capability: webintelligence.CapabilityResearch, Metered: true, Priority: 20},
	}, webintelligence.BudgetPolicy{LimitMicros: 1000000, PerOperationMicros: 250000})
	if err != nil {
		t.Fatalf("new controller: %v", err)
	}
	if err := controller.SetHealth("tinyfish-research-v1", webintelligence.HealthDegraded); err != nil {
		t.Fatalf("set health: %v", err)
	}
	app := server{webIntelligence: controller, webIntelligenceConfigured: true}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/api/v1/web-intelligence/status", nil)
	app.webIntelligenceStatus(recorder, request)
	if recorder.Code != 200 {
		t.Fatalf("status = %d", recorder.Code)
	}
	var payload struct {
		Configured         bool                       `json:"configured"`
		CredentialsExposed bool                       `json:"credentials_exposed"`
		ProductionApproved bool                       `json:"production_approved"`
		Snapshot           webintelligence.Snapshot   `json:"snapshot"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.Configured {
		t.Fatal("configured=false")
	}
	if payload.CredentialsExposed || payload.ProductionApproved {
		t.Fatalf("unexpected authority flags: %#v", payload)
	}
	if len(payload.Snapshot.Providers) != 2 {
		t.Fatalf("providers = %d", len(payload.Snapshot.Providers))
	}
	if payload.Snapshot.Budget.LimitMicros != 1000000 {
		t.Fatalf("limit = %d", payload.Snapshot.Budget.LimitMicros)
	}
}
