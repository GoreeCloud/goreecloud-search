package webautomation

import (
	"context"
	"errors"
	"testing"
	"time"
)

type browserSessionAuthorizerFunc func(context.Context, BrowserSessionAuthorization) error

func (f browserSessionAuthorizerFunc) AuthorizeBrowserSession(ctx context.Context, input BrowserSessionAuthorization) error {
	return f(ctx, input)
}

type fakeBrowserSessionProvider struct {
	startRequest   StartBrowserSessionRequest
	startResult    BrowserSession
	startErr       error
	startCalls     int
	terminateIDs   []string
	terminateErr   error
}

func (p *fakeBrowserSessionProvider) StartBrowserSession(_ context.Context, request StartBrowserSessionRequest) (BrowserSession, error) {
	p.startCalls++
	p.startRequest = request
	return p.startResult, p.startErr
}

func (p *fakeBrowserSessionProvider) TerminateBrowserSession(_ context.Context, sessionID string) error {
	p.terminateIDs = append(p.terminateIDs, sessionID)
	return p.terminateErr
}

func TestBrowserSessionManagerRequiresAuthorityAndProvider(t *testing.T) {
	provider := &fakeBrowserSessionProvider{}
	if _, err := NewBrowserSessionManager(nil, provider); !errors.Is(err, ErrBrowserSessionNotAuthorized) {
		t.Fatalf("expected missing-authorizer rejection, got %v", err)
	}
	authorizer := browserSessionAuthorizerFunc(func(context.Context, BrowserSessionAuthorization) error { return nil })
	if _, err := NewBrowserSessionManager(authorizer, nil); !errors.Is(err, ErrBrowserSessionUnavailable) {
		t.Fatalf("expected missing-provider rejection, got %v", err)
	}
}

func TestBrowserSessionManagerAuthorizesNormalizedStartAndTerminate(t *testing.T) {
	provider := &fakeBrowserSessionProvider{startResult: BrowserSession{
		SessionID: "br-session_1",
		CDPURL:    "wss://browser.example.test/cdp",
		BaseURL:   "https://browser.example.test",
	}}
	var observed []BrowserSessionAuthorization
	authorizer := browserSessionAuthorizerFunc(func(_ context.Context, input BrowserSessionAuthorization) error {
		observed = append(observed, input)
		return nil
	})
	manager, err := NewBrowserSessionManager(authorizer, provider)
	if err != nil {
		t.Fatal(err)
	}

	session, err := manager.Start(context.Background(), StartBrowserSessionRequest{
		TargetURL: "https://example.com/path#fragment",
		Purpose:   "  inspect a user-authorized page  ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if provider.startCalls != 1 {
		t.Fatalf("start calls = %d", provider.startCalls)
	}
	if provider.startRequest.TargetURL != "https://example.com/path" || provider.startRequest.Purpose != "inspect a user-authorized page" {
		t.Fatalf("provider received unnormalized request: %#v", provider.startRequest)
	}
	if provider.startRequest.Timeout != DefaultBrowserSessionTimeout {
		t.Fatalf("timeout = %s", provider.startRequest.Timeout)
	}
	if session.SessionID != "br-session_1" || session.TargetURL != "https://example.com/path" || session.Timeout != DefaultBrowserSessionTimeout {
		t.Fatalf("unexpected session: %#v", session)
	}
	if len(observed) != 1 || observed[0].Operation != BrowserSessionOperationStart || observed[0].TargetURL != "https://example.com/path" || observed[0].Purpose != "inspect a user-authorized page" {
		t.Fatalf("unexpected start authorization: %#v", observed)
	}

	if err := manager.Terminate(context.Background(), TerminateBrowserSessionRequest{
		SessionID: "br-session_1",
		Purpose:   "finish the authorized browser task",
	}); err != nil {
		t.Fatal(err)
	}
	if len(provider.terminateIDs) != 1 || provider.terminateIDs[0] != "br-session_1" {
		t.Fatalf("unexpected terminate calls: %#v", provider.terminateIDs)
	}
	if len(observed) != 2 || observed[1].Operation != BrowserSessionOperationTerminate || observed[1].TargetURL != "https://example.com/path" || observed[1].SessionID != "br-session_1" {
		t.Fatalf("unexpected terminate authorization: %#v", observed)
	}
	if err := manager.Terminate(context.Background(), TerminateBrowserSessionRequest{SessionID: "br-session_1", Purpose: "retry"}); !errors.Is(err, ErrBrowserSessionUnavailable) {
		t.Fatalf("terminated session must no longer be active, got %v", err)
	}
}

func TestBrowserSessionManagerDenialPreventsProviderExecution(t *testing.T) {
	provider := &fakeBrowserSessionProvider{}
	authorizer := browserSessionAuthorizerFunc(func(_ context.Context, _ BrowserSessionAuthorization) error {
		return ErrBrowserSessionNotAuthorized
	})
	manager, err := NewBrowserSessionManager(authorizer, provider)
	if err != nil {
		t.Fatal(err)
	}
	_, err = manager.Start(context.Background(), StartBrowserSessionRequest{
		TargetURL: "https://example.com",
		Purpose:   "test denied browser session",
	})
	if !errors.Is(err, ErrBrowserSessionNotAuthorized) || provider.startCalls != 0 {
		t.Fatalf("expected fail-closed denial before provider, err=%v calls=%d", err, provider.startCalls)
	}
}

func TestBrowserSessionManagerKeepsSessionActiveWhenTerminationUnconfirmed(t *testing.T) {
	provider := &fakeBrowserSessionProvider{startResult: BrowserSession{
		SessionID: "br-retry",
		CDPURL:    "wss://browser.example.test/cdp",
		BaseURL:   "https://browser.example.test",
	}}
	authorizer := browserSessionAuthorizerFunc(func(context.Context, BrowserSessionAuthorization) error { return nil })
	manager, err := NewBrowserSessionManager(authorizer, provider)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Start(context.Background(), StartBrowserSessionRequest{TargetURL: "https://example.com", Purpose: "test termination retry"}); err != nil {
		t.Fatal(err)
	}
	provider.terminateErr = ErrBrowserSessionTerminationUnknown
	if err := manager.Terminate(context.Background(), TerminateBrowserSessionRequest{SessionID: "br-retry", Purpose: "stop session"}); !errors.Is(err, ErrBrowserSessionTerminationUnknown) {
		t.Fatalf("expected unconfirmed termination, got %v", err)
	}
	provider.terminateErr = nil
	if err := manager.Terminate(context.Background(), TerminateBrowserSessionRequest{SessionID: "br-retry", Purpose: "retry stop"}); err != nil {
		t.Fatalf("session should remain active for retry: %v", err)
	}
	if len(provider.terminateIDs) != 2 {
		t.Fatalf("terminate calls = %d", len(provider.terminateIDs))
	}
}

func TestNormalizeBrowserSessionRequestBounds(t *testing.T) {
	for index, request := range []StartBrowserSessionRequest{
		{TargetURL: "http://127.0.0.1/private", Purpose: "test"},
		{TargetURL: "https://example.com", Purpose: ""},
		{TargetURL: "https://example.com", Purpose: "test", Timeout: 4 * time.Second},
		{TargetURL: "https://example.com", Purpose: "test", Timeout: MaxBrowserSessionTimeout + time.Second},
	} {
		if _, err := normalizeStartBrowserSessionRequest(request); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("case %d: expected invalid request, got %v", index, err)
		}
	}
}
