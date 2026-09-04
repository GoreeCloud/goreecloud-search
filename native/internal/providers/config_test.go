package providers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

func TestLoadConfigBuildsSanitizedProviderCapabilities(t *testing.T) {
	config := `{
		"schema_version":1,
		"providers":[{
			"name":"Primary",
			"adapter":"goreecloud-http-v1",
			"endpoint":"https://provider.example/search",
			"categories":["general","images","images"],
			"credential_env":"SEARCH_PROVIDER_TOKEN",
			"published_at_authoritative":false
		}]
	}`
	providers, err := loadConfigBytes([]byte(config), func(name string) (string, bool) {
		if name == "SEARCH_PROVIDER_TOKEN" {
			return "secret", true
		}
		return "", false
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(providers) != 1 {
		t.Fatalf("provider count = %d", len(providers))
	}
	categorized, ok := providers[0].(searchcore.CategoryProvider)
	if !ok {
		t.Fatal("configured provider does not expose category capabilities")
	}
	if got := strings.Join(categorized.Categories(), ","); got != "general,images" {
		t.Fatalf("categories = %q", got)
	}
}

func TestLoadConfigFailsClosedOnCredentialsAndSchema(t *testing.T) {
	tests := []string{
		`{"schema_version":2,"providers":[]}`,
		`{"schema_version":1,"providers":[{"name":"A","adapter":"unknown","endpoint":"https://provider.example/search","categories":["general"]}]}`,
		`{"schema_version":1,"providers":[{"name":"A","adapter":"goreecloud-http-v1","endpoint":"https://provider.example/search","categories":["general"],"credential_env":"bad-name"}]}`,
		`{"schema_version":1,"providers":[{"name":"A","adapter":"goreecloud-http-v1","endpoint":"https://provider.example/search","categories":["general"],"credential_env":"MISSING_TOKEN"}]}`,
		`{"schema_version":1,"providers":[{"name":"A","adapter":"goreecloud-http-v1","endpoint":"https://provider.example/search","categories":["general"]},{"name":"a","adapter":"goreecloud-http-v1","endpoint":"https://provider.example/search","categories":["general"]}]}`,
		`{"schema_version":1,"providers":[],"unexpected":true}`,
	}
	for _, config := range tests {
		if _, err := loadConfigBytes([]byte(config), func(string) (string, bool) { return "", false }); err == nil {
			t.Fatalf("invalid provider config unexpectedly succeeded: %s", config)
		}
	}
}

func TestLoadFromEnvironmentDefaultsToNoProviders(t *testing.T) {
	original, existed := os.LookupEnv(ProviderConfigEnvironment)
	if err := os.Unsetenv(ProviderConfigEnvironment); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(ProviderConfigEnvironment, original)
		} else {
			_ = os.Unsetenv(ProviderConfigEnvironment)
		}
	})
	providers, err := LoadFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if len(providers) != 0 {
		t.Fatalf("default provider count = %d", len(providers))
	}
}

func TestLoadConfigFileRejectsOversizedConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "providers.json")
	if err := os.WriteFile(path, []byte(strings.Repeat(" ", MaxProviderConfigBytes+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfigFile(path, func(string) (string, bool) { return "", false }); err == nil || !strings.Contains(err.Error(), "maximum size") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadConfigFileReadsDeploymentControlledFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "providers.json")
	config := `{"schema_version":1,"providers":[{"name":"News","adapter":"goreecloud-http-v1","endpoint":"https://provider.example/search","categories":["news"]}]}`
	if err := os.WriteFile(path, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	providers, err := LoadConfigFile(path, func(string) (string, bool) { return "", false })
	if err != nil {
		t.Fatal(err)
	}
	if len(providers) != 1 || providers[0].Name() != "News" {
		t.Fatalf("unexpected providers: %+v", providers)
	}
}
