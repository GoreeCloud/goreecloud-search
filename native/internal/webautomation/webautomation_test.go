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
		{URL: "https://example.com", Goal: "x", Purpose: "y", UseVault: true},
		{URL: "https://example.com", Goal: "x", Purpose: "y", OutputSchema: json.RawMessage(`{"type":"array"}`)},
		{URL: "https://example.com", Goal: "x", Purpose: "y", OutputSchema: json.RawMessage(`not-json`)},
		{URL: "https://example.com", Goal: "x", Purpose: "y", MaxSteps: MaxAgentSteps + 1},
	}
	for index, request := range tests {
		if _, err := normalizeRequest(request); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("case %d: expected invalid request, got %v", index, err)
		}
	}
}

func TestNormalizeRequestAllowsScopedVaultAndObjectSchema(t *testing.T) {
	request, err := normalizeRequest(Request{
		URL:               "https://example.com",
		Goal:              "inspect",
		Purpose:           "test",
		UseVault:          true,
		CredentialItemIDs: []string{"cred-1", " cred-1 ", "cred-2"},
		OutputSchema:      json.RawMessage(`{"type":"object","properties":{"state":{"type":"string"}}}`),
	})
	if err != nil {
		t.Fatalf("normalize request: %v", err)
	}
	if len(request.CredentialItemIDs) != 2 || request.CredentialItemIDs[0] != "cred-1" || request.CredentialItemIDs[1] != "cred-2" {
		t.Fatalf("unexpected credential scope: %#v", request.CredentialItemIDs)
	}
	if len(request.OutputSchema) == 0 {
		t.Fatal("output schema was not preserved")
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
			if payload["url"] != "https://example.com/path" {
				t.Fatalf("unexpected request URL: %#v", payload)
			}
			goal, ok := payload["goal"].(string)
			if !ok || !strings.Contains(goal, "check status") || !strings.Contains(goal, "saved browser session") || !strings.Contains(goal, "scoped vault credential") {
				t.Fatalf("unexpected reliability goal: %#v", payload["goal"])
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
			schema, ok := payload["output_schema"].(map[string]any)
			if !ok || schema["type"] != "object" {
				t.Fatalf("unexpected output schema: %#v", payload["output_schema"])
			}
			return jsonResponse(http.StatusOK, `{"run_id":"run_123","status":"PENDING"}`), nil
		case request.Method == http.MethodGet && request.URL.String() == TinyFishAgentRunsBaseURL+"run_123":
			mu.Lock()
			defer mu.Unlock()
			polls++
			if polls == 1 {
				return jsonResponse(http.StatusOK, `{"run_id":"run_123","status":"RUNNING"}`), nil
			}
			return jsonResponse(http.StatusOK, `{"run_id":"run_123","status":"COMPLETED","result_json":{"state":"ok"}}`), nil
		default:
			t.Fatalf("unexpected provider request: %s %s", request.Method, request.URL)
			return nil, nil
		}
	})}

	agent, err := newTinyFishAgentWithHTTPClient(TinyFishAgentConfig{APIKey: "secret-key", PollInterval: 10 * time.Millisecond, AllowBetaMaxSteps: true}, client)
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
		OutputSchema:      json.RawMessage(`{"type":"object","properties":{"state":{"type":"string"}}}`),
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

func TestTinyFishAgentFailsClosedWhenBetaMaxStepsIsNotEnabled(t *testing.T) {
	called := false
	client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		called = true
		return nil, errors.New("must not call provider")
	})}
	agent, err := newTinyFishAgentWithHTTPClient(TinyFishAgentConfig{APIKey: "key"}, client)
	if err != nil {
		t.Fatalf("construct agent: %v", err)
	}
	_, err = agent.Run(context.Background(), Request{URL: "https://example.com", Goal: "work", Purpose: "test", MaxSteps: 10})
	if !errors.Is(err, ErrProviderControlUnavailable) {
		t.Fatalf("expected provider-control error, got %v", err)
	}
	if called {
		t.Fatal("provider was called even though max_steps was not authorized")
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

func TestTinyFishAgentClassifiesRunLevelFailuresReturnedWithHTTP200(t *testing.T) {
	tests := []struct {
		code string
		want error
	}{
		{code: "SITE_BLOCKED", want: ErrSiteBlocked},
		{code: "TASK_FAILED", want: ErrGoalFailed},
		{code: "MAX_STEPS_EXCEEDED", want: ErrAutomationLimit},
		{code: "TIMEOUT", want: ErrAutomationLimit},
		{code: "BILLING_REJECTED", want: ErrBillingRejected},
	}
	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if request.Method == http.MethodPost {
					return jsonResponse(http.StatusOK, `{"run_id":"run_error","status":"PENDING"}`), nil
				}
				return jsonResponse(http.StatusOK, `{"run_id":"run_error","status":"COMPLETED","error":{"code":"`+test.code+`","message":"provider detail must not escape"}}`), nil
			})}
			agent, err := newTinyFishAgentWithHTTPClient(TinyFishAgentConfig{APIKey: "key", PollInterval: 10 * time.Millisecond}, client)
			if err != nil {
				t.Fatalf("construct agent: %v", err)
			}
			_, err = agent.Run(context.Background(), Request{URL: "https://example.com", Goal: "work", Purpose: "test"})
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if strings.Contains(err.Error(), "provider detail") {
				t.Fatal("provider error detail leaked through normalized error")
			}
		})
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
