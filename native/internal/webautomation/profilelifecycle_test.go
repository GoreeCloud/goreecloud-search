package webautomation

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type profileAuthorizerFunc func(context.Context, ProfileAuthorization) error

func (f profileAuthorizerFunc) AuthorizeProfile(ctx context.Context, authorization ProfileAuthorization) error {
	return f(ctx, authorization)
}

type profileProviderStub struct {
	createFn func(context.Context, CreateProfileRequest) (CreateProfileResult, error)
	startFn  func(context.Context, StartProfileSetupRequest) (ProfileSetupSession, error)
	saveFn   func(context.Context, SaveProfileSetupRequest) (SaveProfileSetupResult, error)
	cancelFn func(context.Context, CancelProfileSetupRequest) error
}

func (p profileProviderStub) CreateProfile(ctx context.Context, request CreateProfileRequest) (CreateProfileResult, error) {
	return p.createFn(ctx, request)
}

func (p profileProviderStub) StartProfileSetup(ctx context.Context, request StartProfileSetupRequest) (ProfileSetupSession, error) {
	return p.startFn(ctx, request)
}

func (p profileProviderStub) SaveProfileSetup(ctx context.Context, request SaveProfileSetupRequest) (SaveProfileSetupResult, error) {
	return p.saveFn(ctx, request)
}

func (p profileProviderStub) CancelProfileSetup(ctx context.Context, request CancelProfileSetupRequest) error {
	return p.cancelFn(ctx, request)
}

func TestProfileLifecycleFailsClosedAndKeepsPurposeLocal(t *testing.T) {
	providerCalls := 0
	provider := profileProviderStub{
		createFn: func(_ context.Context, request CreateProfileRequest) (CreateProfileResult, error) {
			providerCalls++
			if request.Name != "GitHub GoreeCloud" || request.Purpose != "Establish reusable authenticated GitHub session" {
				t.Fatalf("unexpected normalized request: %#v", request)
			}
			return CreateProfileResult{ProfileID: "prof_github", Name: request.Name}, nil
		},
		startFn: func(context.Context, StartProfileSetupRequest) (ProfileSetupSession, error) { return ProfileSetupSession{}, nil },
		saveFn:  func(context.Context, SaveProfileSetupRequest) (SaveProfileSetupResult, error) { return SaveProfileSetupResult{}, nil },
		cancelFn: func(context.Context, CancelProfileSetupRequest) error { return nil },
	}

	denied, err := NewProfileLifecycle(profileAuthorizerFunc(func(context.Context, ProfileAuthorization) error {
		return ErrProfileNotAuthorized
	}), provider)
	if err != nil {
		t.Fatalf("NewProfileLifecycle() error = %v", err)
	}
	_, err = denied.Create(context.Background(), CreateProfileRequest{Name: "GitHub GoreeCloud", Purpose: "Establish reusable authenticated GitHub session"})
	if !errors.Is(err, ErrProfileNotAuthorized) || providerCalls != 0 {
		t.Fatalf("expected fail-closed authorization, err=%v calls=%d", err, providerCalls)
	}

	var authorization ProfileAuthorization
	allowed, err := NewProfileLifecycle(profileAuthorizerFunc(func(_ context.Context, value ProfileAuthorization) error {
		authorization = value
		return nil
	}), provider)
	if err != nil {
		t.Fatalf("NewProfileLifecycle() error = %v", err)
	}
	result, err := allowed.Create(context.Background(), CreateProfileRequest{Name: " GitHub GoreeCloud ", Purpose: " Establish reusable authenticated GitHub session "})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if result.ProfileID != "prof_github" || providerCalls != 1 {
		t.Fatalf("unexpected result=%#v calls=%d", result, providerCalls)
	}
	if authorization.Operation != ProfileOperationCreate || authorization.Purpose != "Establish reusable authenticated GitHub session" {
		t.Fatalf("unexpected authorization: %#v", authorization)
	}
}

func TestProfileLifecycleStartSetupNormalizesTargetAndTimeout(t *testing.T) {
	provider := profileProviderStub{
		createFn: func(context.Context, CreateProfileRequest) (CreateProfileResult, error) { return CreateProfileResult{}, nil },
		startFn: func(_ context.Context, request StartProfileSetupRequest) (ProfileSetupSession, error) {
			if request.ProfileID != "prof_example" || request.TargetURL != "https://example.com/login" || request.Timeout != DefaultSetupTimeout {
				t.Fatalf("unexpected start request: %#v", request)
			}
			return ProfileSetupSession{ProfileID: request.ProfileID, SessionID: "sess_example"}, nil
		},
		saveFn:  func(context.Context, SaveProfileSetupRequest) (SaveProfileSetupResult, error) { return SaveProfileSetupResult{}, nil },
		cancelFn: func(context.Context, CancelProfileSetupRequest) error { return nil },
	}
	lifecycle, err := NewProfileLifecycle(profileAuthorizerFunc(func(_ context.Context, authorization ProfileAuthorization) error {
		if authorization.Operation != ProfileOperationStartSetup || authorization.TargetURL != "https://example.com/login" {
			t.Fatalf("unexpected authorization: %#v", authorization)
		}
		return nil
	}), provider)
	if err != nil {
		t.Fatalf("NewProfileLifecycle() error = %v", err)
	}
	_, err = lifecycle.StartSetup(context.Background(), StartProfileSetupRequest{
		ProfileID: "prof_example",
		TargetURL: "https://example.com/login#fragment",
		Purpose:   "Repair saved session",
	})
	if err != nil {
		t.Fatalf("StartSetup() error = %v", err)
	}
}

func TestTinyFishProfileCreateDoesNotForwardPurpose(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost || request.URL.String() != TinyFishProfilesEndpoint {
			t.Fatalf("unexpected request %s %s", request.Method, request.URL)
		}
		if request.Header.Get("X-API-Key") != "secret" {
			t.Fatal("missing TinyFish API key")
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request: %v", err)
		}
		if strings.Contains(string(body), "Do not forward this purpose") {
			t.Fatalf("purpose leaked to provider: %s", body)
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["name"] != "GitHub GoreeCloud" || payload["proxy_country_code"] != "US" || payload["set_as_default"] != true {
			t.Fatalf("unexpected payload: %#v", payload)
		}
		return jsonResponse(http.StatusOK, `{"id":"prof_abc123","name":"GitHub GoreeCloud","proxy_country_code":"US"}`), nil
	})}
	manager, err := newTinyFishProfileManagerWithHTTPClient("secret", client)
	if err != nil {
		t.Fatalf("newTinyFishProfileManagerWithHTTPClient() error = %v", err)
	}
	result, err := manager.CreateProfile(context.Background(), CreateProfileRequest{
		Name:             "GitHub GoreeCloud",
		Purpose:          "Do not forward this purpose",
		SetAsDefault:     true,
		ProxyCountryCode: "us",
	})
	if err != nil {
		t.Fatalf("CreateProfile() error = %v", err)
	}
	if result.ProfileID != "prof_abc123" || result.ProxyCountryCode != "US" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestTinyFishProfileSetupSaveAndCancel(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		calls++
		switch calls {
		case 1:
			if request.URL.String() != TinyFishProfilesEndpoint+"/prof_abc/setup-session" {
				t.Fatalf("unexpected setup URL: %s", request.URL)
			}
			return jsonResponse(http.StatusOK, `{"session_id":"sess_123","cdp_url":"wss://browser.example/session","base_url":"https://browser.example","timeout_seconds":900,"expires_at":"2026-09-12T23:59:00Z"}`), nil
		case 2:
			if request.URL.String() != TinyFishProfilesEndpoint+"/prof_abc/save" {
				t.Fatalf("unexpected save URL: %s", request.URL)
			}
			return jsonResponse(http.StatusOK, `{"domains_updated":["Example.COM","example.com"],"domains_failed":["login.example.com"],"cookie_count":12,"pages_captured":1}`), nil
		case 3:
			if request.URL.String() != TinyFishProfilesEndpoint+"/prof_abc/setup-session/cancel" {
				t.Fatalf("unexpected cancel URL: %s", request.URL)
			}
			return jsonResponse(http.StatusOK, `{}`), nil
		default:
			t.Fatalf("unexpected extra request")
			return nil, nil
		}
	})}
	manager, err := newTinyFishProfileManagerWithHTTPClient("secret", client)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	session, err := manager.StartProfileSetup(context.Background(), StartProfileSetupRequest{
		ProfileID: "prof_abc",
		TargetURL: "https://example.com/login",
		Purpose:   "Repair session",
		Timeout:   15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("StartProfileSetup() error = %v", err)
	}
	if session.SessionID != "sess_123" || session.CDPURL != "wss://browser.example/session" || session.Timeout != 15*time.Minute {
		t.Fatalf("unexpected setup session: %#v", session)
	}
	saved, err := manager.SaveProfileSetup(context.Background(), SaveProfileSetupRequest{ProfileID: "prof_abc", SessionID: "sess_123", Purpose: "Persist repaired session"})
	if err != nil {
		t.Fatalf("SaveProfileSetup() error = %v", err)
	}
	if len(saved.DomainsUpdated) != 1 || saved.DomainsUpdated[0] != "example.com" || saved.CookieCount != 12 || saved.PagesCaptured != 1 {
		t.Fatalf("unexpected save result: %#v", saved)
	}
	if err := manager.CancelProfileSetup(context.Background(), CancelProfileSetupRequest{ProfileID: "prof_abc", SessionID: "sess_123", Purpose: "Discard setup"}); err != nil {
		t.Fatalf("CancelProfileSetup() error = %v", err)
	}
}

func TestTinyFishProfileErrorMappingAndUnsafeResponse(t *testing.T) {
	tests := []struct {
		status int
		want   error
	}{
		{http.StatusNotFound, ErrProfileUnavailable},
		{http.StatusConflict, ErrProfileConflict},
		{http.StatusBadRequest, ErrInvalidRequest},
	}
	for _, test := range tests {
		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return jsonResponse(test.status, `{"error":"provider detail must not escape"}`), nil
		})}
		manager, err := newTinyFishProfileManagerWithHTTPClient("secret", client)
		if err != nil {
			t.Fatalf("new manager: %v", err)
		}
		_, err = manager.StartProfileSetup(context.Background(), StartProfileSetupRequest{ProfileID: "prof_abc", TargetURL: "https://example.com", Purpose: "test"})
		if !errors.Is(err, test.want) {
			t.Fatalf("status %d: got %v want %v", test.status, err, test.want)
		}
		if strings.Contains(err.Error(), "provider detail") {
			t.Fatalf("provider response leaked: %v", err)
		}
	}

	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{"session_id":"sess_123","cdp_url":"https://wrong.example","base_url":"https://browser.example","timeout_seconds":900,"expires_at":"2026-09-12T23:59:00Z"}`), nil
	})}
	manager, err := newTinyFishProfileManagerWithHTTPClient("secret", client)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	_, err = manager.StartProfileSetup(context.Background(), StartProfileSetupRequest{ProfileID: "prof_abc", TargetURL: "https://example.com", Purpose: "test"})
	if !errors.Is(err, ErrAutomationFailed) {
		t.Fatalf("expected invalid provider endpoint failure, got %v", err)
	}
}
