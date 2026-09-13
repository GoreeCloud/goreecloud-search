package webautomation

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type AuthAcceptanceState string

const (
	AuthAcceptanceUnverified              AuthAcceptanceState = "unverified"
	AuthAcceptanceSetupSaved              AuthAcceptanceState = "setup_saved"
	AuthAcceptanceSessionReuseVerified    AuthAcceptanceState = "session_reuse_verified"
	AuthAcceptanceVaultRepairVerified     AuthAcceptanceState = "vault_repair_verified"
	AuthAcceptanceMFARequired             AuthAcceptanceState = "mfa_required"
	AuthAcceptanceSiteBlocked             AuthAcceptanceState = "site_blocked"
	AuthAcceptanceExpired                 AuthAcceptanceState = "expired"
	AuthAcceptanceRevoked                 AuthAcceptanceState = "revoked"
	AuthAcceptanceFailed                  AuthAcceptanceState = "failed"
	AuthAcceptanceSecretDisclosureDetected AuthAcceptanceState = "secret_disclosure_detected"
)

var (
	ErrAuthAcceptanceUnavailable = errors.New("authenticated-session acceptance evidence is unavailable")
	ErrAuthAcceptanceStale       = errors.New("authenticated-session acceptance evidence is stale")
)

// AuthAcceptanceObservation records only a normalized outcome for one exact-host
// authentication binding. It intentionally has no fields capable of carrying a
// password, OTP, cookie, API key, CDP URL, browser session ID, or page content.
type AuthAcceptanceObservation struct {
	Host       string
	ProfileID  string
	State      AuthAcceptanceState
	ObservedAt time.Time
}

// AuthAcceptanceStatus is privacy-minimized operational evidence. It proves only
// what has actually been observed for a binding; it never establishes production
// approval, Privacy Shield authorization, or Wardveil protected state.
type AuthAcceptanceStatus struct {
	Host                     string              `json:"host"`
	State                    AuthAcceptanceState `json:"state"`
	ObservedAt               time.Time           `json:"observed_at"`
	BrowserProfile           BrowserProfile      `json:"browser_profile"`
	VaultConfigured          bool                `json:"vault_configured"`
	SetupSavedObserved       bool                `json:"setup_saved_observed"`
	SessionReuseVerified     bool                `json:"session_reuse_verified"`
	VaultRepairVerified      bool                `json:"vault_repair_verified"`
	SecretDisclosureDetected bool                `json:"secret_disclosure_detected"`
	ObservationCount         uint64              `json:"observation_count"`
}

// AuthAcceptanceRegistry stores in-memory Development evidence for managed exact-
// host bindings. Persistence/distribution is deliberately outside this primitive.
type AuthAcceptanceRegistry struct {
	mu       sync.RWMutex
	bindings *AuthBindings
	status   map[string]AuthAcceptanceStatus
}

func NewAuthAcceptanceRegistry(bindings *AuthBindings) (*AuthAcceptanceRegistry, error) {
	if bindings == nil || len(bindings.byHost) == 0 {
		return nil, ErrAuthAcceptanceUnavailable
	}
	return &AuthAcceptanceRegistry{bindings: bindings, status: make(map[string]AuthAcceptanceStatus)}, nil
}

func (r *AuthAcceptanceRegistry) Record(input AuthAcceptanceObservation) (AuthAcceptanceStatus, error) {
	if r == nil || r.bindings == nil {
		return AuthAcceptanceStatus{}, ErrAuthAcceptanceUnavailable
	}
	host := canonicalBindingHost(input.Host)
	binding, ok := r.bindings.byHost[host]
	if !ok {
		return AuthAcceptanceStatus{}, ErrAuthBindingUnavailable
	}
	profileID, err := normalizeProviderReference(input.ProfileID, "profile ID")
	if err != nil {
		return AuthAcceptanceStatus{}, err
	}
	if profileID != binding.ProfileID {
		return AuthAcceptanceStatus{}, fmt.Errorf("%w: observation profile does not match exact-host binding", ErrInvalidRequest)
	}
	state := normalizeAuthAcceptanceState(input.State)
	if !validAuthAcceptanceState(state) {
		return AuthAcceptanceStatus{}, fmt.Errorf("%w: authenticated-session acceptance state is invalid", ErrInvalidRequest)
	}
	observedAt := input.ObservedAt.UTC()
	if observedAt.IsZero() {
		return AuthAcceptanceStatus{}, fmt.Errorf("%w: observation time is required", ErrInvalidRequest)
	}
	if observedAt.After(time.Now().UTC().Add(5 * time.Minute)) {
		return AuthAcceptanceStatus{}, fmt.Errorf("%w: observation time is implausibly in the future", ErrInvalidRequest)
	}
	if state == AuthAcceptanceVaultRepairVerified && len(binding.CredentialItemIDs) == 0 {
		return AuthAcceptanceStatus{}, fmt.Errorf("%w: Vault repair cannot be verified for a profile-only binding", ErrInvalidRequest)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	status := r.status[host]
	if !status.ObservedAt.IsZero() && observedAt.Before(status.ObservedAt) {
		return AuthAcceptanceStatus{}, ErrAuthAcceptanceStale
	}
	if status.Host == "" {
		status.Host = host
		status.BrowserProfile = binding.BrowserProfile
		status.VaultConfigured = len(binding.CredentialItemIDs) > 0
	}
	status.State = state
	status.ObservedAt = observedAt
	status.ObservationCount++

	switch state {
	case AuthAcceptanceSetupSaved:
		status.SetupSavedObserved = true
	case AuthAcceptanceSessionReuseVerified:
		status.SetupSavedObserved = true
		status.SessionReuseVerified = true
	case AuthAcceptanceVaultRepairVerified:
		status.SetupSavedObserved = true
		status.VaultRepairVerified = true
	case AuthAcceptanceSecretDisclosureDetected:
		status.SecretDisclosureDetected = true
	}

	r.status[host] = status
	return status, nil
}

func (r *AuthAcceptanceRegistry) Status(host string) (AuthAcceptanceStatus, error) {
	if r == nil || r.bindings == nil {
		return AuthAcceptanceStatus{}, ErrAuthAcceptanceUnavailable
	}
	canonical := canonicalBindingHost(host)
	binding, ok := r.bindings.byHost[canonical]
	if !ok {
		return AuthAcceptanceStatus{}, ErrAuthBindingUnavailable
	}
	r.mu.RLock()
	status, exists := r.status[canonical]
	r.mu.RUnlock()
	if exists {
		return status, nil
	}
	return AuthAcceptanceStatus{
		Host:            canonical,
		State:           AuthAcceptanceUnverified,
		BrowserProfile:  binding.BrowserProfile,
		VaultConfigured: len(binding.CredentialItemIDs) > 0,
	}, nil
}

func (r *AuthAcceptanceRegistry) Snapshot() []AuthAcceptanceStatus {
	if r == nil || r.bindings == nil {
		return nil
	}
	hosts := make([]string, 0, len(r.bindings.byHost))
	for host := range r.bindings.byHost {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)
	output := make([]AuthAcceptanceStatus, 0, len(hosts))
	for _, host := range hosts {
		status, err := r.Status(host)
		if err == nil {
			output = append(output, status)
		}
	}
	return output
}

func normalizeAuthAcceptanceState(input AuthAcceptanceState) AuthAcceptanceState {
	value := strings.ToLower(strings.TrimSpace(string(input)))
	value = strings.NewReplacer("-", "_", " ", "_").Replace(value)
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}
	return AuthAcceptanceState(value)
}

func validAuthAcceptanceState(state AuthAcceptanceState) bool {
	switch state {
	case AuthAcceptanceSetupSaved,
		AuthAcceptanceSessionReuseVerified,
		AuthAcceptanceVaultRepairVerified,
		AuthAcceptanceMFARequired,
		AuthAcceptanceSiteBlocked,
		AuthAcceptanceExpired,
		AuthAcceptanceRevoked,
		AuthAcceptanceFailed,
		AuthAcceptanceSecretDisclosureDetected:
		return true
	default:
		return false
	}
}
