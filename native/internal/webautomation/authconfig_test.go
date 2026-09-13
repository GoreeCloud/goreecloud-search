package webautomation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAuthControlFromEnvironmentAbsent(t *testing.T) {
	t.Setenv(AuthBindingsEnvironment, "")
	control, configured, err := LoadAuthControlFromEnvironment()
	if err != nil {
		t.Fatalf("load control: %v", err)
	}
	if configured || control != nil {
		t.Fatalf("unexpected configured control: configured=%v control=%#v", configured, control)
	}
}

func TestLoadAuthControlFileSnapshotRedactsProviderReferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth-bindings.json")
	body := `{
  "schema_version": 1,
  "bindings": [
    {
      "host": "example.com",
      "profile_id": "prof_secret_reference",
      "credential_item_ids": ["credential://vault/secret-reference"],
      "browser_profile": "lite"
    }
  ]
}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	control, err := LoadAuthControlFile(path)
	if err != nil {
		t.Fatalf("load control: %v", err)
	}
	snapshot := control.Snapshot()
	if len(snapshot) != 1 {
		t.Fatalf("snapshot length = %d", len(snapshot))
	}
	if snapshot[0].Host != "example.com" || snapshot[0].State != AuthAcceptanceUnverified {
		t.Fatalf("snapshot = %#v", snapshot[0])
	}
	if !snapshot[0].VaultConfigured || snapshot[0].BrowserProfile != BrowserProfileLite {
		t.Fatalf("snapshot flags = %#v", snapshot[0])
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	text := string(encoded)
	for _, forbidden := range []string{"prof_secret_reference", "credential://vault/secret-reference"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("snapshot leaked provider reference %q: %s", forbidden, text)
		}
	}
}

func TestLoadAuthControlFileRejectsOversizedDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "oversized.json")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", MaxAuthBindingsBytes+1)), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := LoadAuthControlFile(path); err == nil || !strings.Contains(err.Error(), "maximum size") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadAuthControlFileRejectsInvalidDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":1,"bindings":[]}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := LoadAuthControlFile(path); err == nil {
		t.Fatal("expected invalid binding document error")
	}
}
