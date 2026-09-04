package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

const (
	ProviderConfigEnvironment                  = "GOREECLOUD_SEARCH_PROVIDER_CONFIG_FILE"
	RequireReleaseProviderCoverageEnvironment = "GOREECLOUD_SEARCH_REQUIRE_RELEASE_PROVIDER_COVERAGE"
	ProviderConfigVersion                      = 1
	MaxProviderConfigBytes                     = 256 << 10
	MaxConfiguredProviders                     = 32
)

var credentialEnvironmentName = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,127}$`)

type providerConfigFile struct {
	SchemaVersion int                  `json:"schema_version"`
	Providers     []configuredProvider `json:"providers"`
}

type configuredProvider struct {
	Name                     string   `json:"name"`
	Adapter                  string   `json:"adapter"`
	Endpoint                 string   `json:"endpoint"`
	Categories               []string `json:"categories"`
	CredentialEnvironment    string   `json:"credential_env,omitempty"`
	PublishedAtAuthoritative bool     `json:"published_at_authoritative,omitempty"`
}

type environmentLookup func(string) (string, bool)

func LoadFromEnvironment() ([]searchcore.Provider, error) {
	requireCoverage, err := parseReleaseProviderCoverageRequirement(
		os.Getenv(RequireReleaseProviderCoverageEnvironment),
	)
	if err != nil {
		return nil, err
	}

	path := strings.TrimSpace(os.Getenv(ProviderConfigEnvironment))
	configured := []searchcore.Provider{}
	if path != "" {
		configured, err = LoadConfigFile(path, os.LookupEnv)
		if err != nil {
			return nil, err
		}
	}
	if requireCoverage {
		if err := validateReleaseProviderCoverage(configured); err != nil {
			return nil, err
		}
	}
	return configured, nil
}

func parseReleaseProviderCoverageRequirement(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "0", "false":
		return false, nil
	case "1", "true":
		return true, nil
	default:
		return false, errors.New("release provider coverage requirement is invalid")
	}
}

func validateReleaseProviderCoverage(configured []searchcore.Provider) error {
	covered := make(map[string]bool, len(searchcore.SupportedCategories))
	for _, category := range searchcore.ProviderBackedCategories(configured...) {
		covered[category] = true
	}
	missing := make([]string, 0, len(searchcore.SupportedCategories))
	for _, category := range searchcore.SupportedCategories {
		if !covered[category] {
			missing = append(missing, category)
		}
	}
	if len(missing) != 0 {
		return fmt.Errorf(
			"release provider coverage is incomplete; missing categories: %s",
			strings.Join(missing, ", "),
		)
	}
	return nil
}

func LoadConfigFile(path string, lookup environmentLookup) ([]searchcore.Provider, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("provider config path is required")
	}
	if lookup == nil {
		return nil, errors.New("provider credential lookup is required")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open provider config: %w", err)
	}
	defer file.Close()

	body, err := io.ReadAll(io.LimitReader(file, MaxProviderConfigBytes+1))
	if err != nil {
		return nil, errors.New("read provider config")
	}
	if len(body) > MaxProviderConfigBytes {
		return nil, errors.New("provider config exceeds maximum size")
	}
	return loadConfigBytes(body, lookup)
}

func loadConfigBytes(body []byte, lookup environmentLookup) ([]searchcore.Provider, error) {
	var config providerConfigFile
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return nil, errors.New("provider config is invalid JSON")
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return nil, errors.New("provider config contains trailing data")
	}
	if config.SchemaVersion != ProviderConfigVersion {
		return nil, errors.New("provider config schema version is unsupported")
	}
	if len(config.Providers) > MaxConfiguredProviders {
		return nil, errors.New("provider config exceeds provider limit")
	}

	providers := make([]searchcore.Provider, 0, len(config.Providers))
	seenNames := map[string]bool{}
	for index, configured := range config.Providers {
		name := strings.TrimSpace(configured.Name)
		key := strings.ToLower(name)
		if key == "" || seenNames[key] {
			return nil, fmt.Errorf("provider %d has a missing or duplicate name", index+1)
		}
		seenNames[key] = true
		if configured.Adapter != "goreecloud-http-v1" {
			return nil, fmt.Errorf("provider %q uses an unsupported adapter", name)
		}

		credential := ""
		credentialEnvironment := strings.TrimSpace(configured.CredentialEnvironment)
		if credentialEnvironment != "" {
			if !credentialEnvironmentName.MatchString(credentialEnvironment) {
				return nil, fmt.Errorf("provider %q has an invalid credential environment name", name)
			}
			value, ok := lookup(credentialEnvironment)
			if !ok || strings.TrimSpace(value) == "" {
				return nil, fmt.Errorf("provider %q credential is unavailable", name)
			}
			credential = value
		}

		provider, err := NewHTTPJSONProvider(HTTPJSONConfig{
			Name:                     name,
			Endpoint:                 configured.Endpoint,
			Categories:               configured.Categories,
			BearerToken:              credential,
			PublishedAtAuthoritative: configured.PublishedAtAuthoritative,
		})
		if err != nil {
			return nil, fmt.Errorf("provider %q configuration is invalid: %w", name, err)
		}
		providers = append(providers, provider)
	}
	return providers, nil
}
