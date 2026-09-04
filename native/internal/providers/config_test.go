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
	requireOriginal, requireExisted := os.LookupEnv(RequireReleaseProviderCoverageEnvironment)
	if err := os.Unsetenv(RequireReleaseProviderCoverageEnvironment); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(ProviderConfigEnvironment, original)
		} else {
			_ = os.Unsetenv(ProviderConfigEnvironment)
		}
		if requireExisted {
			_ = os.Setenv(RequireReleaseProviderCoverageEnvironment, requireOriginal)
		} else {
			_ = os.Unsetenv(RequireReleaseProviderCoverageEnvironment)
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

func TestReleaseProviderCoverageRequirementParsing(t *testing.T) {
	for _, raw := range []string{"", "0", "false", " FALSE "} {
		required, err := parseReleaseProviderCoverageRequirement(raw)
		if err != nil || required {
			t.Fatalf("%q = required %v, err %v; want disabled", raw, required, err)
		}
	}
	for _, raw := range []string{"1", "true", " TRUE "} {
		required, err := parseReleaseProviderCoverageRequirement(raw)
		if err != nil || !required {
			t.Fatalf("%q = required %v, err %v; want enabled", raw, required, err)
		}
	}
	if _, err := parseReleaseProviderCoverageRequirement("yes"); err == nil {
		t.Fatal("ambiguous release provider coverage flag unexpectedly succeeded")
	}
}

func TestReleaseProviderCoverageAcceptsExecutableAllCategoryConfig(t *testing.T) {
	config := `{
		"schema_version":1,
		"providers":[{
			"name":"Structural coverage fixture",
			"adapter":"goreecloud-http-v1",
			"endpoint":"https://provider.example/search",
			"categories":["general","images","videos","news","files"]
		}]
	}`
	configured, err := loadConfigBytes([]byte(config), func(string) (string, bool) { return "", false })
	if err != nil {
		t.Fatal(err)
	}
	if err := validateReleaseProviderCoverage(configured); err != nil {
		t.Fatalf("all-category executable config failed coverage validation: %v", err)
	}
}

func TestReleaseProviderCoverageFailsClosedWithMissingCategories(t *testing.T) {
	config := `{
		"schema_version":1,
		"providers":[{
			"name":"General only",
			"adapter":"goreecloud-http-v1",
			"endpoint":"https://provider.example/search",
			"categories":["general"]
		}]
	}`
	configured, err := loadConfigBytes([]byte(config), func(string) (string, bool) { return "", false })
	if err != nil {
		t.Fatal(err)
	}
	err = validateReleaseProviderCoverage(configured)
	if err == nil {
		t.Fatal("incomplete provider coverage unexpectedly succeeded")
	}
	want := "missing categories: images, videos, news, files"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("coverage error = %q, want %q", err, want)
	}
}

func TestLoadFromEnvironmentRequiresProviderConfigWhenReleaseCoverageEnabled(t *testing.T) {
	t.Setenv(ProviderConfigEnvironment, "")
	t.Setenv(RequireReleaseProviderCoverageEnvironment, "true")
	if _, err := LoadFromEnvironment(); err == nil || !strings.Contains(err.Error(), "missing categories: general, images, videos, news, files") {
		t.Fatalf("unexpected coverage error: %v", err)
	}
}

func TestLoadFromEnvironmentEnforcesReleaseCoverage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "providers.json")
	config := `{"schema_version":1,"providers":[{"name":"All","adapter":"goreecloud-http-v1","endpoint":"https://provider.example/search","categories":["general","images","videos","news","files"]}]}`
	if err := os.WriteFile(path, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(ProviderConfigEnvironment, path)
	t.Setenv(RequireReleaseProviderCoverageEnvironment, "1")
	configured, err := LoadFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if len(configured) != 1 {
		t.Fatalf("configured provider count = %d, want 1", len(configured))
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
