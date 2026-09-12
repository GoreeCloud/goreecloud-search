package webautomation

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type authorizerFunc func(context.Context, Request) error

func (f authorizerFunc) Authorize(ctx context.Context, request Request) error {
	return f(ctx, request)
}

type executorFunc func(context.Context, Request) (Result, error)

func (f executorFunc) Run(ctx context.Context, request Request) (Result, error) {
	return f(ctx, request)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestEscalatorFailsClosedAndRequiresExecutor(t *testing.T) {
	calls := 0
	agent := executorFunc(func(_ context.Context, request Request) (Result, error) {
		calls++
		return Result{Capability: request.Capability}, nil
	})
	denied, err := NewEscalator(authorizerFunc(func(_ context.Context, _ Request) error {
		return ErrNotAuthorized
	}), agent, nil)
	if err != nil {
		t.Fatalf("new escalator: %v", err)
	}
	_, err = denied.Run(context.Background(), Request{URL: "https://example.com", Goal: "inspect", Purpose: "test"})
	if !errors.Is(err, ErrNotAuthorized) || calls != 0 {
		t.Fatalf("expected fail-closed authorization, err=%v calls=%d", err, calls)
	}

	allowed, err := NewEscalator(authorizerFunc(func(_ context.Context, _ Request) error { return nil }), agent, nil)
	if err != nil {
		t.Fatalf("new escalator: %v", err)
	}
	_, err = allowed.Run(context.Background(), Request{
		URL:        "https://example.com",
		Goal:       "inspect",
		Purpose:    "test",
		Capability: CapabilityBrowser,
	})
	if !errors.Is(err, ErrCapabilityUnavailable) {
		t.Fatalf("expected unavailable browser capability, got %v", err)
	}
}

func TestNormalizeRequestRejectsUnsafeSessionInputs(t *testing.T) {
	tests := []Request{
		{URL: "http://127.0.0.1/private", Goal: "x", Purpose: "y"},
		{URL: "https://localhost/", Goal: "x", Purpose: "y"},
		{URL: "https://example.com", Goal: "x", Purpose: "y", UseProfile: true},
		{URL: "https://example.com", Goal: "x", Purpose: "y", ProfileID: "profile-without-use"},
		{URL: "https://example.com", Goal: "x", Purpose: "y", CredentialItemIDs: []string{"cred"}},
		{URL: "https://example.com", Goal: "x", Purpose: "y", MaxSteps: MaxAgentSteps + 1},
	}
	for index, request := range tests {
		if _, err := normalizeRequest(request); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("case %d: expected invalid request, got %v", index, err)
		}
	}
}

func TestTinyFishAgentMapsBoundedRequestAndPolls(t *testing.T) {
	var mu sync.Mutex
	polls := 0
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch {
		case request.Method == http.MethodPost && request.URL.String() == TinyFishAgentRunAsyncEndpoint:
			if request.Header.Get("X-API-Key") != "secret-key" {
				t.Fatal("missing API key header")
			}
			var payload map[string]any
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if payload["url"] != "https://example.com/path" || payload["goal"] != "check status" {
				t.Fatalf("unexpected request payload: %#v", payload)
			}
			if payload["browser_profile"] != "lite" || payload["use_profile"] != true || payload["profile_id"] != "prof-1" {
				t.Fatalf("unexpected profile mapping: %#v", payload)
			}
			if payload["use_vault"] != true {
				t.Fatalf("vault not enabled: %#v", payload)
			}
			credentials, ok := payload["credential_item_ids"].([]any)
			if !ok || len(credentials) != 1 || credentials[0] != "cred-1" {
				t.Fatalf("unexpected credential IDs: %#v", payload["credential_item_ids"])
			}
			if _, exists := payload["purpose"]; exists {
				t.Fatal("local purpose binding must not be sent to TinyFish")
			}
			config, ok := payload["agent_config"].(map[string]any)
			if !ok || config["max_steps"] != float64(12) || config["max_duration_seconds"] != float64(120) {
				t.Fatalf("unexpected agent_config: %#v", payload["agent_config"])
			}
			return jsonResponse(http.StatusOK, `{"run_id":"run_123","status":"PENDING"}`), nil
		case request.Method == http.MethodGet && request.URL.String() == TinyFishAgentRunsBaseURL+"run_123":
			mu.Lock()
			defer mu.Unlock()
			polls++
			if polls == 1 {
				return jsonResponse(http.StatusOK, `{"run_id":"run_123","status":"RUNNING"}`), nil
			}
			return jsonResponse(http.StatusOK, `{"run_id":"run_123","status":"COMPLETED","result":{"state":"ok"}}`), nil
		default:
			t.Fatalf("unexpected provider request: %s %s", request.Method, request.URL)
			return nil, nil
		}
	})}

	agent, err := newTinyFishAgentWithHTTPClient(TinyFishAgentConfig{APIKey: "secret-key", PollInterval: 10 * time.Millisecond}, client)
	if err != nil {
		t.Fatalf("construct agent: %v", err)
	}
	result, err := agent.Run(context.Background(), Request{
		URL:               "https://example.com/path#fragment",
		Goal:              " check status ",
		Purpose:           " operational verification ",
		UseProfile:        true,
		ProfileID:         "prof-1",
		UseVault:          true,
		CredentialItemIDs: []string{"cred-1"},
		MaxSteps:          12,
		MaxDuration:       2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("run agent: %v", err)
	}
	if result.Capability != CapabilityAgent || result.Output != `{"state":"ok"}` {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestTinyFishAgentCancelsAfterCallerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancelledProvider := false
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch {
		case request.Method == http.MethodPost && request.URL.String() == TinyFishAgentRunAsyncEndpoint:
			return jsonResponse(http.StatusOK, `{"run_id":"run_cancel","status":"PENDING"}`), nil
		case request.Method == http.MethodGet && request.URL.String() == TinyFishAgentRunsBaseURL+"run_cancel":
			cancel()
			return jsonResponse(http.StatusOK, `{"run_id":"run_cancel","status":"RUNNING"}`), nil
		case request.Method == http.MethodPost && request.URL.String() == TinyFishAgentRunsBaseURL+"run_cancel/cancel":
			cancelledProvider = true
			return jsonResponse(http.StatusOK, `{"run_id":"run_cancel","status":"CANCELLED"}`), nil
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL)
			return nil, nil
		}
	})}
	agent, err := newTinyFishAgentWithHTTPClient(TinyFishAgentConfig{APIKey: "key", PollInterval: 10 * time.Millisecond}, client)
	if err != nil {
		t.Fatalf("construct agent: %v", err)
	}
	_, err = agent.Run(ctx, Request{URL: "https://example.com", Goal: "work", Purpose: "test"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected caller cancellation, got %v", err)
	}
	if !cancelledProvider {
		t.Fatal("provider run was not cancelled")
	}
}

func TestTinyFishAgentSanitizesProviderFailure(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method == http.MethodPost {
			return jsonResponse(http.StatusOK, `{"run_id":"run_failed","status":"PENDING"}`), nil
		}
		return jsonResponse(http.StatusOK, `{"run_id":"run_failed","status":"FAILED","error":{"message":"secret provider detail"}}`), nil
	})}
	agent, err := newTinyFishAgentWithHTTPClient(TinyFishAgentConfig{APIKey: "key", PollInterval: 10 * time.Millisecond}, client)
	if err != nil {
		t.Fatalf("construct agent: %v", err)
	}
	_, err = agent.Run(context.Background(), Request{URL: "https://example.com", Goal: "work", Purpose: "test"})
	if !errors.Is(err, ErrAutomationFailed) {
		t.Fatalf("expected automation failure, got %v", err)
	}
	if strings.Contains(err.Error(), "secret provider detail") {
		t.Fatal("provider error body leaked through GoreeCloud error")
	}
}

func TestTinyFishAgentRejectsInvalidRunIdentity(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{"run_id":"../bad","status":"PENDING"}`), nil
	})}
	agent, err := newTinyFishAgentWithHTTPClient(TinyFishAgentConfig{APIKey: "key", PollInterval: 10 * time.Millisecond}, client)
	if err != nil {
		t.Fatalf("construct agent: %v", err)
	}
	_, err = agent.Run(context.Background(), Request{URL: "https://example.com", Goal: "work", Purpose: "test"})
	if !errors.Is(err, ErrAutomationFailed) {
		t.Fatalf("expected invalid run identity failure, got %v", err)
	}
}
