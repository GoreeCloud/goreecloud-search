package webintelligence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	ConfigEnvironment       = "GOREECLOUD_SEARCH_WEB_INTELLIGENCE_CONFIG_FILE"
	ConfigSchemaVersion     = 1
	MaxConfigBytes          = 256 << 10
	MaxConfiguredProviders  = 32
)

type configFile struct {
	SchemaVersion int              `json:"schema_version"`
	Budget        BudgetPolicy     `json:"budget"`
	Providers     []configProvider `json:"providers"`
}

type configProvider struct {
	ID         string     `json:"id"`
	Capability Capability `json:"capability"`
	Metered    bool       `json:"metered"`
	Priority   int        `json:"priority"`
	Health     Health     `json:"health,omitempty"`
}

// LoadFromEnvironment loads the optional GoreeCloud web-intelligence control
// configuration. Absence is not an error and means the control plane is not
// configured. This loader intentionally accepts no provider credentials,
// queries, URLs, goals, page content, or account identifiers.
func LoadFromEnvironment() (*Controller, bool, error) {
	path := strings.TrimSpace(os.Getenv(ConfigEnvironment))
	if path == "" {
		return nil, false, nil
	}
	controller, err := LoadConfigFile(path)
	if err != nil {
		return nil, false, err
	}
	return controller, true, nil
}

func LoadConfigFile(path string) (*Controller, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("web intelligence config path is required")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open web intelligence config: %w", err)
	}
	defer file.Close()

	body, err := io.ReadAll(io.LimitReader(file, MaxConfigBytes+1))
	if err != nil {
		return nil, errors.New("read web intelligence config")
	}
	if len(body) > MaxConfigBytes {
		return nil, errors.New("web intelligence config exceeds maximum size")
	}
	return loadConfigBytes(body)
}

func loadConfigBytes(body []byte) (*Controller, error) {
	var config configFile
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return nil, errors.New("web intelligence config is invalid JSON")
	}
	if err := ensureConfigEOF(decoder); err != nil {
		return nil, errors.New("web intelligence config contains trailing data")
	}
	if config.SchemaVersion != ConfigSchemaVersion {
		return nil, errors.New("web intelligence config schema version is unsupported")
	}
	if len(config.Providers) > MaxConfiguredProviders {
		return nil, errors.New("web intelligence config exceeds provider limit")
	}

	providers := make([]Provider, 0, len(config.Providers))
	for _, configured := range config.Providers {
		providers = append(providers, Provider{
			ID:         configured.ID,
			Capability: configured.Capability,
			Metered:    configured.Metered,
			Priority:   configured.Priority,
		})
	}
	controller, err := NewController(providers, config.Budget)
	if err != nil {
		return nil, fmt.Errorf("web intelligence config is invalid: %w", err)
	}
	for _, configured := range config.Providers {
		health := configured.Health
		if health == "" {
			health = HealthHealthy
		}
		if err := controller.SetHealth(configured.ID, health); err != nil {
			return nil, fmt.Errorf("web intelligence config provider %q health is invalid: %w", strings.TrimSpace(configured.ID), err)
		}
	}
	return controller, nil
}

func ensureConfigEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("trailing data")
	}
	return nil
}
