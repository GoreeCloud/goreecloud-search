package webautomation

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAuthAcceptanceExactHostAndProfileBinding(t *testing.T) {
	bindings, err := NewAuthBindings([]AuthBinding{{
		Host:              "github.com",
		ProfileID:         "prof_github",
		CredentialItemIDs: []string{"cred_github"},
		BrowserProfile:    BrowserProfileLite,
	}})
	if err != nil {
		t.Fatalf("NewAuthBindings() error = %v", err)
	}
	registry, err := NewAuthAcceptanceRegistry(bindings)
	if err != nil {
		t.Fatalf("NewAuthAcceptanceRegistry() error = %v", err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	status, err := registry.Record(AuthAcceptanceObservation{
		Host:       "GITHUB.com.",
		ProfileID:  "prof_github",
		State:      "session reuse verified",
		ObservedAt: now,
	})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if status.Host != "github.com" || status.State != AuthAcceptanceSessionReuseVerified || !status.SessionReuseVerified || !status.SetupSavedObserved || !status.VaultConfigured {
		t.Fatalf("unexpected status: %#v", status)
	}

	_, err = registry.Record(AuthAcceptanceObservation{Host: "api.github.com", ProfileID: "prof_github", State: AuthAcceptanceSessionReuseVerified, ObservedAt: now})
	if !errors.Is(err, ErrAuthBindingUnavailable) {
		t.Fatalf("expected exact-host isolation, got %v", err)
	}
	_, err = registry.Record(AuthAcceptanceObservation{Host: "github.com", ProfileID: "prof_wrong", State: AuthAcceptanceSessionReuseVerified, ObservedAt: now})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected profile mismatch rejection, got %v", err)
	}
}

func TestAuthAcceptanceVaultRepairRequiresScopedCredentialBinding(t *testing.T) {
	bindings, err := NewAuthBindings([]AuthBinding{{Host: "example.com", ProfileID: "prof_example"}})
	if err != nil {
		t.Fatalf("NewAuthBindings() error = %v", err)
	}
	registry, err := NewAuthAcceptanceRegistry(bindings)
	if err != nil {
		t.Fatalf("NewAuthAcceptanceRegistry() error = %v", err)
	}
	_, err = registry.Record(AuthAcceptanceObservation{
		Host: "example.com", ProfileID: "prof_example", State: AuthAcceptanceVaultRepairVerified, ObservedAt: time.Now().UTC(),
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected profile-only Vault repair rejection, got %v", err)
	}
}

func TestAuthAcceptanceRejectsStaleObservationAndPreservesHistoricalEvidence(t *testing.T) {
	bindings, err := NewAuthBindings([]AuthBinding{{
		Host: "example.com", ProfileID: "prof_example", CredentialItemIDs: []string{"cred_example"},
	}})
	if err != nil {
		t.Fatalf("NewAuthBindings() error = %v", err)
	}
	registry, err := NewAuthAcceptanceRegistry(bindings)
	if err != nil {
		t.Fatalf("NewAuthAcceptanceRegistry() error = %v", err)
	}
	base := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)
	if _, err := registry.Record(AuthAcceptanceObservation{Host: "example.com", ProfileID: "prof_example", State: AuthAcceptanceSessionReuseVerified, ObservedAt: base}); err != nil {
		t.Fatalf("Record reuse error = %v", err)
	}
	if _, err := registry.Record(AuthAcceptanceObservation{Host: "example.com", ProfileID: "prof_example", State: AuthAcceptanceVaultRepairVerified, ObservedAt: base.Add(time.Second)}); err != nil {
		t.Fatalf("Record Vault repair error = %v", err)
	}
	status, err := registry.Record(AuthAcceptanceObservation{Host: "example.com", ProfileID: "prof_example", State: AuthAcceptanceExpired, ObservedAt: base.Add(2 * time.Second)})
	if err != nil {
		t.Fatalf("Record expired error = %v", err)
	}
	if status.State != AuthAcceptanceExpired || !status.SessionReuseVerified || !status.VaultRepairVerified || status.ObservationCount != 3 {
		t.Fatalf("historical evidence lost: %#v", status)
	}
	_, err = registry.Record(AuthAcceptanceObservation{Host: "example.com", ProfileID: "prof_example", State: AuthAcceptanceFailed, ObservedAt: base})
	if !errors.Is(err, ErrAuthAcceptanceStale) {
		t.Fatalf("expected stale observation rejection, got %v", err)
	}
}

func TestAuthAcceptanceSecretDisclosureIsStickyAndSnapshotOmitsProviderReferences(t *testing.T) {
	bindings, err := NewAuthBindings([]AuthBinding{{
		Host: "portal.example.com", ProfileID: "prof_sensitive_reference", CredentialItemIDs: []string{"cred_sensitive_reference"}, BrowserProfile: BrowserProfileStealth,
	}})
	if err != nil {
		t.Fatalf("NewAuthBindings() error = %v", err)
	}
	registry, err := NewAuthAcceptanceRegistry(bindings)
	if err != nil {
		t.Fatalf("NewAuthAcceptanceRegistry() error = %v", err)
	}
	base := time.Now().UTC().Add(-time.Minute)
	if _, err := registry.Record(AuthAcceptanceObservation{Host: "portal.example.com", ProfileID: "prof_sensitive_reference", State: AuthAcceptanceSecretDisclosureDetected, ObservedAt: base}); err != nil {
		t.Fatalf("Record secret-disclosure observation error = %v", err)
	}
	status, err := registry.Record(AuthAcceptanceObservation{Host: "portal.example.com", ProfileID: "prof_sensitive_reference", State: AuthAcceptanceSessionReuseVerified, ObservedAt: base.Add(time.Second)})
	if err != nil {
		t.Fatalf("Record reuse observation error = %v", err)
	}
	if !status.SecretDisclosureDetected || !status.SessionReuseVerified {
		t.Fatalf("expected sticky disclosure evidence: %#v", status)
	}
	encoded, err := json.Marshal(registry.Snapshot())
	if err != nil {
		t.Fatalf("Marshal snapshot error = %v", err)
	}
	text := string(encoded)
	if strings.Contains(text, "prof_sensitive_reference") || strings.Contains(text, "cred_sensitive_reference") {
		t.Fatalf("provider authentication references leaked in snapshot: %s", text)
	}
	if !strings.Contains(text, `"browser_profile":"stealth"`) || !strings.Contains(text, `"secret_disclosure_detected":true`) {
		t.Fatalf("expected privacy-minimized evidence missing: %s", text)
	}
}

func TestAuthAcceptanceUnobservedBindingIsExplicitlyUnverified(t *testing.T) {
	bindings, err := NewAuthBindings([]AuthBinding{{Host: "example.com", ProfileID: "prof_example"}})
	if err != nil {
		t.Fatalf("NewAuthBindings() error = %v", err)
	}
	registry, err := NewAuthAcceptanceRegistry(bindings)
	if err != nil {
		t.Fatalf("NewAuthAcceptanceRegistry() error = %v", err)
	}
	status, err := registry.Status("example.com")
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.State != AuthAcceptanceUnverified || status.ObservationCount != 0 || !status.ObservedAt.IsZero() {
		t.Fatalf("unexpected unverified status: %#v", status)
	}
}
