package webautomation

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const AuthBindingsEnvironment = "GOREECLOUD_SEARCH_WEB_AUTOMATION_AUTH_BINDINGS_FILE"

// AuthControl groups the deployment-controlled exact-host authentication
// bindings with their privacy-minimized in-memory acceptance evidence. It does
// not contain credential values or expose provider profile/credential IDs
// through its Snapshot method.
type AuthControl struct {
	Bindings   *AuthBindings
	Acceptance *AuthAcceptanceRegistry
}

// LoadAuthControlFromEnvironment loads the optional authenticated web-automation
// control configuration. Absence is not an error and means authenticated web
// automation is not configured for this Development runtime.
func LoadAuthControlFromEnvironment() (*AuthControl, bool, error) {
	path := strings.TrimSpace(os.Getenv(AuthBindingsEnvironment))
	if path == "" {
		return nil, false, nil
	}
	control, err := LoadAuthControlFile(path)
	if err != nil {
		return nil, false, err
	}
	return control, true, nil
}

func LoadAuthControlFile(path string) (*AuthControl, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("web automation authentication binding path is required")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open web automation authentication bindings: %w", err)
	}
	defer file.Close()

	body, err := io.ReadAll(io.LimitReader(file, MaxAuthBindingsBytes+1))
	if err != nil {
		return nil, errors.New("read web automation authentication bindings")
	}
	if len(body) > MaxAuthBindingsBytes {
		return nil, errors.New("web automation authentication bindings exceed maximum size")
	}
	bindings, err := ParseAuthBindings(body)
	if err != nil {
		return nil, fmt.Errorf("web automation authentication bindings are invalid: %w", err)
	}
	acceptance, err := NewAuthAcceptanceRegistry(bindings)
	if err != nil {
		return nil, fmt.Errorf("initialize web automation authentication acceptance: %w", err)
	}
	return &AuthControl{Bindings: bindings, Acceptance: acceptance}, nil
}

// Snapshot exposes only the privacy-minimized authenticated-session evidence.
// Browser Context Profile IDs and Vault credential item IDs intentionally remain
// internal to the managed binding registry.
func (c *AuthControl) Snapshot() []AuthAcceptanceStatus {
	if c == nil || c.Acceptance == nil {
		return nil
	}
	return c.Acceptance.Snapshot()
}
