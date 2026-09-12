package webautomation

import (
	"errors"
	"strings"
	"testing"
)

func TestAuthBindingsApplyExactHost(t *testing.T) {
	bindings, err := NewAuthBindings([]AuthBinding{{
		Host:              "EXAMPLE.com.",
		ProfileID:         "prof_example",
		CredentialItemIDs: []string{"cred_example", "cred_example"},
	}})
	if err != nil {
		t.Fatalf("NewAuthBindings() error = %v", err)
	}

	request, err := bindings.Apply(Request{
		URL:     "https://example.com/account#fragment",
		Goal:    "Read the account page.",
		Purpose: "Authenticated account verification",
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if request.URL != "https://example.com/account" {
		t.Fatalf("URL = %q, want normalized URL", request.URL)
	}
	if !request.UseProfile || request.ProfileID != "prof_example" {
		t.Fatalf("profile binding not applied: %#v", request)
	}
	if !request.UseVault || len(request.CredentialItemIDs) != 1 || request.CredentialItemIDs[0] != "cred_example" {
		t.Fatalf("vault binding not applied or deduplicated: %#v", request.CredentialItemIDs)
	}
	if request.BrowserProfile != BrowserProfileLite {
		t.Fatalf("BrowserProfile = %q, want lite", request.BrowserProfile)
	}
}

func TestAuthBindingsDoNotMatchSubdomains(t *testing.T) {
	bindings, err := NewAuthBindings([]AuthBinding{{Host: "example.com", ProfileID: "prof_example"}})
	if err != nil {
		t.Fatalf("NewAuthBindings() error = %v", err)
	}
	_, err = bindings.Apply(Request{URL: "https://login.example.com/", Goal: "Open login.", Purpose: "Test"})
	if !errors.Is(err, ErrAuthBindingUnavailable) {
		t.Fatalf("Apply() error = %v, want ErrAuthBindingUnavailable", err)
	}
}

func TestAuthBindingsRejectCallerAuthenticationOverrides(t *testing.T) {
	bindings, err := NewAuthBindings([]AuthBinding{{Host: "example.com", ProfileID: "prof_example"}})
	if err != nil {
		t.Fatalf("NewAuthBindings() error = %v", err)
	}
	_, err = bindings.Apply(Request{
		URL:        "https://example.com/",
		Goal:       "Open account.",
		Purpose:    "Test",
		UseProfile: true,
		ProfileID:  "prof_other",
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("Apply() error = %v, want ErrInvalidRequest", err)
	}
}

func TestParseAuthBindingsStrictAndScoped(t *testing.T) {
	bindings, err := ParseAuthBindings([]byte(`{
		"schema_version": 1,
		"bindings": [{
			"host": "dash.example.com",
			"profile_id": "prof_dashboard",
			"credential_item_ids": ["cred_dashboard"],
			"browser_profile": "stealth"
		}]
	}`))
	if err != nil {
		t.Fatalf("ParseAuthBindings() error = %v", err)
	}
	request, err := bindings.Apply(Request{URL: "https://dash.example.com/settings", Goal: "Open settings.", Purpose: "Test"})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if request.BrowserProfile != BrowserProfileStealth {
		t.Fatalf("BrowserProfile = %q, want stealth", request.BrowserProfile)
	}
	if request.ProfileID != "prof_dashboard" || len(request.CredentialItemIDs) != 1 || request.CredentialItemIDs[0] != "cred_dashboard" {
		t.Fatalf("unexpected binding result: %#v", request)
	}
}

func TestParseAuthBindingsRejectsUnknownFieldsAndTrailingJSON(t *testing.T) {
	cases := [][]byte{
		[]byte(`{"schema_version":1,"bindings":[{"host":"example.com","profile_id":"prof"}],"secret":"no"}`),
		[]byte(`{"schema_version":1,"bindings":[{"host":"example.com","profile_id":"prof"}]} {}`),
	}
	for _, data := range cases {
		if _, err := ParseAuthBindings(data); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("ParseAuthBindings(%q) error = %v, want ErrInvalidRequest", data, err)
		}
	}
}

func TestAuthBindingsRejectUnsafeOrAmbiguousBindings(t *testing.T) {
	tests := []AuthBinding{
		{Host: "localhost", ProfileID: "prof"},
		{Host: "example.com/path", ProfileID: "prof"},
		{Host: "127.0.0.1", ProfileID: "prof"},
		{Host: "example.com", ProfileID: ""},
		{Host: "example.com", ProfileID: "prof", CredentialItemIDs: []string{""}},
	}
	for _, binding := range tests {
		if _, err := NewAuthBindings([]AuthBinding{binding}); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("NewAuthBindings(%#v) error = %v, want ErrInvalidRequest", binding, err)
		}
	}
}

func TestAuthBindingsRejectDuplicateCanonicalHost(t *testing.T) {
	_, err := NewAuthBindings([]AuthBinding{
		{Host: "example.com", ProfileID: "prof_one"},
		{Host: "EXAMPLE.COM.", ProfileID: "prof_two"},
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("NewAuthBindings() error = %v, want ErrInvalidRequest", err)
	}
}

func TestAuthBindingsDocumentBounds(t *testing.T) {
	if _, err := ParseAuthBindings(nil); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("ParseAuthBindings(nil) error = %v, want ErrInvalidRequest", err)
	}
	oversized := []byte(strings.Repeat("x", MaxAuthBindingsBytes+1))
	if _, err := ParseAuthBindings(oversized); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("ParseAuthBindings(oversized) error = %v, want ErrInvalidRequest", err)
	}
}
