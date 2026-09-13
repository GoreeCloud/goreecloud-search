package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GoreeCloud/goreecloud-search/native/internal/webautomation"
)

func TestWebAutomationAuthStatusUnconfigured(t *testing.T) {
	app := server{}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/api/v1/web-intelligence/authentication/status", nil)
	app.webAutomationAuthStatus(recorder, request)
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
	for _, key := range []string{"credentials_exposed", "profile_ids_exposed", "credential_item_ids_exposed", "production_approved"} {
		if value, ok := payload[key].(bool); !ok || value {
			t.Fatalf("%s = %#v", key, payload[key])
		}
	}
}

func TestWebAutomationAuthStatusConfiguredSnapshotRedactsReferences(t *testing.T) {
	bindings, err := webautomation.ParseAuthBindings([]byte(`{
  "schema_version": 1,
  "bindings": [
    {
      "host": "example.com",
      "profile_id": "prof_private_reference",
      "credential_item_ids": ["credential://vault/private-reference"],
      "browser_profile": "stealth"
    }
  ]
}`))
	if err != nil {
		t.Fatalf("parse bindings: %v", err)
	}
	acceptance, err := webautomation.NewAuthAcceptanceRegistry(bindings)
	if err != nil {
		t.Fatalf("acceptance registry: %v", err)
	}
	control := &webautomation.AuthControl{Bindings: bindings, Acceptance: acceptance}
	app := server{webAutomationAuth: control, webAutomationAuthConfigured: true}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/api/v1/web-intelligence/authentication/status", nil)
	app.webAutomationAuthStatus(recorder, request)
	if recorder.Code != 200 {
		t.Fatalf("status = %d", recorder.Code)
	}
	body := recorder.Body.String()
	for _, forbidden := range []string{"prof_private_reference", "credential://vault/private-reference"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("status leaked provider reference %q: %s", forbidden, body)
		}
	}
	var payload struct {
		Configured                bool                                  `json:"configured"`
		CredentialsExposed        bool                                  `json:"credentials_exposed"`
		ProfileIDsExposed         bool                                  `json:"profile_ids_exposed"`
		CredentialItemIDsExposed  bool                                  `json:"credential_item_ids_exposed"`
		ProductionApproved        bool                                  `json:"production_approved"`
		Snapshot                  []webautomation.AuthAcceptanceStatus  `json:"snapshot"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.Configured || payload.CredentialsExposed || payload.ProfileIDsExposed || payload.CredentialItemIDsExposed || payload.ProductionApproved {
		t.Fatalf("unexpected authority flags: %#v", payload)
	}
	if len(payload.Snapshot) != 1 {
		t.Fatalf("snapshot length = %d", len(payload.Snapshot))
	}
	if payload.Snapshot[0].Host != "example.com" || payload.Snapshot[0].State != webautomation.AuthAcceptanceUnverified {
		t.Fatalf("snapshot = %#v", payload.Snapshot[0])
	}
	if !payload.Snapshot[0].VaultConfigured || payload.Snapshot[0].BrowserProfile != webautomation.BrowserProfileStealth {
		t.Fatalf("snapshot flags = %#v", payload.Snapshot[0])
	}
}
