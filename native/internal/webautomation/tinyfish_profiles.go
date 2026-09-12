package webautomation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	TinyFishProfilesEndpoint        = "https://agent.tinyfish.ai/v1/profiles"
	MaxTinyFishProfileResponseBytes = 2 << 20
)

type TinyFishProfileManager struct {
	apiKey string
	client *http.Client
}

type tinyFishCreateProfileRequest struct {
	Name             string  `json:"name"`
	SetAsDefault     bool    `json:"set_as_default,omitempty"`
	ProxyCountryCode *string `json:"proxy_country_code,omitempty"`
}

type tinyFishCreateProfileResponse struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	ProxyCountryCode *string `json:"proxy_country_code"`
}

type tinyFishStartSetupRequest struct {
	URL            string `json:"url"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

type tinyFishStartSetupResponse struct {
	SessionID      string `json:"session_id"`
	CDPURL         string `json:"cdp_url"`
	BaseURL        string `json:"base_url"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	ExpiresAt      string `json:"expires_at"`
}

type tinyFishSetupSessionRequest struct {
	SessionID string `json:"session_id"`
}

type tinyFishSaveSetupResponse struct {
	DomainsUpdated []string `json:"domains_updated"`
	DomainsFailed  []string `json:"domains_failed"`
	CookieCount    int      `json:"cookie_count"`
	PagesCaptured  int      `json:"pages_captured"`
}

func NewTinyFishProfileManager(apiKey string) (*TinyFishProfileManager, error) {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return nil, fmt.Errorf("%w: TinyFish API key is required", ErrInvalidRequest)
	}
	return &TinyFishProfileManager{apiKey: key, client: secureTinyFishAgentClient()}, nil
}

func newTinyFishProfileManagerWithHTTPClient(apiKey string, client *http.Client) (*TinyFishProfileManager, error) {
	manager, err := NewTinyFishProfileManager(apiKey)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("%w: HTTP client is required", ErrInvalidRequest)
	}
	manager.client = client
	return manager, nil
}

func (m *TinyFishProfileManager) CreateProfile(ctx context.Context, input CreateProfileRequest) (CreateProfileResult, error) {
	request, err := normalizeCreateProfileRequest(input)
	if err != nil {
		return CreateProfileResult{}, err
	}
	payload := tinyFishCreateProfileRequest{Name: request.Name, SetAsDefault: request.SetAsDefault}
	if request.ProxyCountryCode != "" {
		country := request.ProxyCountryCode
		payload.ProxyCountryCode = &country
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return CreateProfileResult{}, fmt.Errorf("%w: encode profile request", ErrInvalidRequest)
	}
	responseBytes, err := m.doJSON(ctx, http.MethodPost, TinyFishProfilesEndpoint, body)
	if err != nil {
		return CreateProfileResult{}, err
	}
	var response tinyFishCreateProfileResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return CreateProfileResult{}, fmt.Errorf("%w: profile create response is invalid", ErrAutomationFailed)
	}
	profileID, err := normalizeProviderReference(response.ID, "profile ID")
	if err != nil {
		return CreateProfileResult{}, fmt.Errorf("%w: provider returned invalid profile identity", ErrAutomationFailed)
	}
	name := strings.TrimSpace(response.Name)
	if name == "" || len([]rune(name)) > MaxProfileNameRunes {
		return CreateProfileResult{}, fmt.Errorf("%w: provider returned invalid profile name", ErrAutomationFailed)
	}
	country := ""
	if response.ProxyCountryCode != nil {
		country = strings.ToUpper(strings.TrimSpace(*response.ProxyCountryCode))
		if country != "" && !validProfileProxyCountry(country) {
			return CreateProfileResult{}, fmt.Errorf("%w: provider returned unsupported profile proxy country", ErrAutomationFailed)
		}
	}
	return CreateProfileResult{ProfileID: profileID, Name: name, ProxyCountryCode: country}, nil
}

func (m *TinyFishProfileManager) StartProfileSetup(ctx context.Context, input StartProfileSetupRequest) (ProfileSetupSession, error) {
	request, err := normalizeStartProfileSetupRequest(input)
	if err != nil {
		return ProfileSetupSession{}, err
	}
	payload, err := json.Marshal(tinyFishStartSetupRequest{URL: request.TargetURL, TimeoutSeconds: durationSeconds(request.Timeout)})
	if err != nil {
		return ProfileSetupSession{}, fmt.Errorf("%w: encode profile setup request", ErrInvalidRequest)
	}
	endpoint := TinyFishProfilesEndpoint + "/" + url.PathEscape(request.ProfileID) + "/setup-session"
	responseBytes, err := m.doJSON(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return ProfileSetupSession{}, err
	}
	var response tinyFishStartSetupResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return ProfileSetupSession{}, fmt.Errorf("%w: profile setup response is invalid", ErrAutomationFailed)
	}
	sessionID, err := normalizeProviderReference(response.SessionID, "setup session ID")
	if err != nil {
		return ProfileSetupSession{}, fmt.Errorf("%w: provider returned invalid setup session identity", ErrAutomationFailed)
	}
	cdpURL, err := validateSensitiveSetupURL(response.CDPURL, "wss")
	if err != nil {
		return ProfileSetupSession{}, err
	}
	baseURL, err := validateSensitiveSetupURL(response.BaseURL, "https")
	if err != nil {
		return ProfileSetupSession{}, err
	}
	if response.TimeoutSeconds < 1 || response.TimeoutSeconds > int(MaxProfileSetupTimeout/time.Second) {
		return ProfileSetupSession{}, fmt.Errorf("%w: provider returned invalid profile setup timeout", ErrAutomationFailed)
	}
	expiresAt, err := time.Parse(time.RFC3339, strings.TrimSpace(response.ExpiresAt))
	if err != nil || expiresAt.IsZero() {
		return ProfileSetupSession{}, fmt.Errorf("%w: provider returned invalid profile setup expiry", ErrAutomationFailed)
	}
	return ProfileSetupSession{
		ProfileID: request.ProfileID,
		SessionID: sessionID,
		CDPURL:    cdpURL,
		BaseURL:   baseURL,
		Timeout:   time.Duration(response.TimeoutSeconds) * time.Second,
		ExpiresAt: expiresAt,
	}, nil
}

func (m *TinyFishProfileManager) SaveProfileSetup(ctx context.Context, input SaveProfileSetupRequest) (SaveProfileSetupResult, error) {
	request, err := normalizeSaveProfileSetupRequest(input)
	if err != nil {
		return SaveProfileSetupResult{}, err
	}
	payload, err := json.Marshal(tinyFishSetupSessionRequest{SessionID: request.SessionID})
	if err != nil {
		return SaveProfileSetupResult{}, fmt.Errorf("%w: encode profile save request", ErrInvalidRequest)
	}
	endpoint := TinyFishProfilesEndpoint + "/" + url.PathEscape(request.ProfileID) + "/save"
	responseBytes, err := m.doJSON(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return SaveProfileSetupResult{}, err
	}
	var response tinyFishSaveSetupResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return SaveProfileSetupResult{}, fmt.Errorf("%w: profile save response is invalid", ErrAutomationFailed)
	}
	updated, err := normalizeSavedDomains(response.DomainsUpdated)
	if err != nil {
		return SaveProfileSetupResult{}, err
	}
	failed, err := normalizeSavedDomains(response.DomainsFailed)
	if err != nil {
		return SaveProfileSetupResult{}, err
	}
	if response.CookieCount < 0 || response.PagesCaptured < 0 {
		return SaveProfileSetupResult{}, fmt.Errorf("%w: profile save counters are invalid", ErrAutomationFailed)
	}
	return SaveProfileSetupResult{DomainsUpdated: updated, DomainsFailed: failed, CookieCount: response.CookieCount, PagesCaptured: response.PagesCaptured}, nil
}

func (m *TinyFishProfileManager) CancelProfileSetup(ctx context.Context, input CancelProfileSetupRequest) error {
	request, err := normalizeCancelProfileSetupRequest(input)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(tinyFishSetupSessionRequest{SessionID: request.SessionID})
	if err != nil {
		return fmt.Errorf("%w: encode profile cancel request", ErrInvalidRequest)
	}
	endpoint := TinyFishProfilesEndpoint + "/" + url.PathEscape(request.ProfileID) + "/setup-session/cancel"
	_, err = m.doJSON(ctx, http.MethodPost, endpoint, payload)
	return err
}

func (m *TinyFishProfileManager) doJSON(ctx context.Context, method, endpoint string, payload []byte) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		body = bytes.NewReader(payload)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("%w: create profile provider request", ErrAutomationFailed)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-API-Key", m.apiKey)
	request.Header.Set("User-Agent", "GoreeCloud-Search-Native/1 tinyfish-profiles-v1")
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := m.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("%w: profile provider request failed", ErrAutomationFailed)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		switch response.StatusCode {
		case http.StatusBadRequest:
			return nil, fmt.Errorf("%w: profile provider rejected request", ErrInvalidRequest)
		case http.StatusUnauthorized, http.StatusNotFound:
			return nil, ErrProfileUnavailable
		case http.StatusConflict:
			return nil, ErrProfileConflict
		default:
			return nil, fmt.Errorf("%w: profile provider returned HTTP %d", ErrAutomationFailed, response.StatusCode)
		}
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, MaxTinyFishProfileResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: profile provider response could not be read", ErrAutomationFailed)
	}
	if len(content) > MaxTinyFishProfileResponseBytes {
		return nil, fmt.Errorf("%w: profile provider response exceeds maximum size", ErrAutomationFailed)
	}
	if len(bytes.TrimSpace(content)) == 0 {
		return []byte("{}"), nil
	}
	if contentType := response.Header.Get("Content-Type"); contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		return nil, fmt.Errorf("%w: profile provider returned unsupported content type", ErrAutomationFailed)
	}
	return content, nil
}

func validateSensitiveSetupURL(raw, expectedScheme string) (string, error) {
	value := strings.TrimSpace(raw)
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != expectedScheme || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return "", fmt.Errorf("%w: provider returned invalid profile setup endpoint", ErrAutomationFailed)
	}
	return parsed.String(), nil
}

func normalizeSavedDomains(values []string) ([]string, error) {
	output := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, raw := range values {
		domain := canonicalBindingHost(raw)
		if !validBindingHost(domain) {
			return nil, fmt.Errorf("%w: provider returned invalid saved domain", ErrAutomationFailed)
		}
		if !seen[domain] {
			seen[domain] = true
			output = append(output, domain)
		}
	}
	return output, nil
}
