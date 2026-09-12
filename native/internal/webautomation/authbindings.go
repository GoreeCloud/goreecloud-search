package webautomation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	AuthBindingsSchemaVersion = 1
	MaxAuthBindings           = 64
	MaxAuthBindingsBytes      = 64 << 10
)

var ErrAuthBindingUnavailable = errors.New("web automation authentication binding is unavailable")

// AuthBinding binds one exact public website host to one saved Browser Context
// Profile and, optionally, an explicitly scoped set of Vault credential item
// references. It contains references only; credential values never belong here.
//
// Host matching is deliberately exact. A binding for example.com does not
// silently authorize credentials for login.example.com or any other subdomain.
type AuthBinding struct {
	Host              string         `json:"host"`
	ProfileID         string         `json:"profile_id"`
	CredentialItemIDs []string       `json:"credential_item_ids,omitempty"`
	BrowserProfile    BrowserProfile `json:"browser_profile,omitempty"`
}

type authBindingsDocument struct {
	SchemaVersion int           `json:"schema_version"`
	Bindings      []AuthBinding `json:"bindings"`
}

// AuthBindings is an immutable, exact-host authentication-reference registry.
// It exists to prevent a caller from accidentally pairing a target website with
// the wrong saved browser session or with broad/unscoped Vault access.
type AuthBindings struct {
	byHost map[string]AuthBinding
}

func NewAuthBindings(bindings []AuthBinding) (*AuthBindings, error) {
	if len(bindings) == 0 || len(bindings) > MaxAuthBindings {
		return nil, fmt.Errorf("%w: binding count must be between 1 and %d", ErrInvalidRequest, MaxAuthBindings)
	}

	registry := &AuthBindings{byHost: make(map[string]AuthBinding, len(bindings))}
	for _, input := range bindings {
		binding, err := normalizeAuthBinding(input)
		if err != nil {
			return nil, err
		}
		if _, exists := registry.byHost[binding.Host]; exists {
			return nil, fmt.Errorf("%w: duplicate authentication binding host", ErrInvalidRequest)
		}
		registry.byHost[binding.Host] = binding
	}
	return registry, nil
}

// ParseAuthBindings parses a bounded, strict deployment-controlled JSON
// document. The document carries profile and credential references, not secrets.
func ParseAuthBindings(data []byte) (*AuthBindings, error) {
	if len(data) == 0 || len(data) > MaxAuthBindingsBytes {
		return nil, fmt.Errorf("%w: authentication binding document size is invalid", ErrInvalidRequest)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var document authBindingsDocument
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("%w: authentication binding document is invalid", ErrInvalidRequest)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return nil, err
	}
	if document.SchemaVersion != AuthBindingsSchemaVersion {
		return nil, fmt.Errorf("%w: unsupported authentication binding schema version", ErrInvalidRequest)
	}
	return NewAuthBindings(document.Bindings)
}

// Apply resolves the exact target host and injects only the corresponding saved
// profile/Vault references. Callers are not allowed to mix ad-hoc profile or
// Vault references with a managed binding because doing so defeats the binding.
func (b *AuthBindings) Apply(input Request) (Request, error) {
	if b == nil || len(b.byHost) == 0 {
		return Request{}, ErrAuthBindingUnavailable
	}
	if input.UseProfile || input.ProfileID != "" || input.UseVault || len(input.CredentialItemIDs) != 0 {
		return Request{}, fmt.Errorf("%w: managed authentication bindings cannot be combined with caller-supplied authentication references", ErrInvalidRequest)
	}

	target, err := normalizeTargetURL(input.URL)
	if err != nil {
		return Request{}, err
	}
	parsed, err := url.Parse(target)
	if err != nil {
		return Request{}, fmt.Errorf("%w: target URL is invalid", ErrInvalidRequest)
	}
	host := canonicalBindingHost(parsed.Hostname())
	binding, ok := b.byHost[host]
	if !ok {
		return Request{}, ErrAuthBindingUnavailable
	}

	output := input
	output.URL = target
	output.UseProfile = true
	output.ProfileID = binding.ProfileID
	output.BrowserProfile = binding.BrowserProfile
	if len(binding.CredentialItemIDs) > 0 {
		output.UseVault = true
		output.CredentialItemIDs = append([]string(nil), binding.CredentialItemIDs...)
	}
	return output, nil
}

func normalizeAuthBinding(input AuthBinding) (AuthBinding, error) {
	host := canonicalBindingHost(input.Host)
	if !validBindingHost(host) {
		return AuthBinding{}, fmt.Errorf("%w: authentication binding host is invalid", ErrInvalidRequest)
	}
	profileID := strings.TrimSpace(input.ProfileID)
	if profileID == "" || utf8.RuneCountInString(profileID) > MaxProfileIDRunes || strings.ContainsAny(profileID, "\r\n\x00") {
		return AuthBinding{}, fmt.Errorf("%w: authentication binding profile ID is invalid", ErrInvalidRequest)
	}

	profile := input.BrowserProfile
	if profile == "" {
		profile = BrowserProfileLite
	}
	if profile != BrowserProfileLite && profile != BrowserProfileStealth {
		return AuthBinding{}, fmt.Errorf("%w: authentication binding browser profile is invalid", ErrInvalidRequest)
	}

	if len(input.CredentialItemIDs) > MaxCredentialItems {
		return AuthBinding{}, fmt.Errorf("%w: too many authentication binding credential item IDs", ErrInvalidRequest)
	}
	credentialIDs := make([]string, 0, len(input.CredentialItemIDs))
	seen := map[string]bool{}
	for _, raw := range input.CredentialItemIDs {
		value := strings.TrimSpace(raw)
		if value == "" || utf8.RuneCountInString(value) > MaxCredentialItemIDRunes || strings.ContainsAny(value, "\r\n\x00") {
			return AuthBinding{}, fmt.Errorf("%w: authentication binding credential item ID is invalid", ErrInvalidRequest)
		}
		if !seen[value] {
			seen[value] = true
			credentialIDs = append(credentialIDs, value)
		}
	}

	return AuthBinding{
		Host:              host,
		ProfileID:         profileID,
		CredentialItemIDs: credentialIDs,
		BrowserProfile:    profile,
	}, nil
}

func canonicalBindingHost(value string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
}

func validBindingHost(host string) bool {
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || !strings.Contains(host, ".") {
		return false
	}
	if strings.ContainsAny(host, "/:@?#[]\r\n\x00") {
		return false
	}
	if net.ParseIP(host) != nil {
		return false
	}
	if len(host) > 253 {
		return false
	}
	labels := strings.Split(host, ".")
	for _, label := range labels {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, char := range label {
			if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
				continue
			}
			return false
		}
	}
	return true
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err == io.EOF {
		return nil
	} else if err != nil {
		return fmt.Errorf("%w: authentication binding document contains trailing data", ErrInvalidRequest)
	}
	return fmt.Errorf("%w: authentication binding document contains multiple JSON values", ErrInvalidRequest)
}
