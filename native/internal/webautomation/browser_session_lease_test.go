package webautomation

import (
	"context"
	"errors"
	"testing"
	"time"
)

type leaseBrowserProvider struct {
	startResult       BrowserSession
	startErr          error
	startCalls        int
	terminateCalls    int
	terminateErr      error
	terminateCtxError error
}

func (p *leaseBrowserProvider) StartBrowserSession(_ context.Context, _ StartBrowserSessionRequest) (BrowserSession, error) {
	p.startCalls++
	return p.startResult, p.startErr
}

func (p *leaseBrowserProvider) TerminateBrowserSession(ctx context.Context, _ string) error {
	p.terminateCalls++
	p.terminateCtxError = ctx.Err()
	return p.terminateErr
}

func newLeaseTestRunner(t *testing.T, provider *leaseBrowserProvider) *BrowserSessionLeaseRunner {
	t.Helper()
	authorizer := browserSessionAuthorizerFunc(func(context.Context, BrowserSessionAuthorization) error { return nil })
	manager, err := NewBrowserSessionManager(authorizer, provider)
	if err != nil {
		t.Fatal(err)
	}
	runner, err := NewBrowserSessionLeaseRunner(manager, 0)
	if err != nil {
		t.Fatal(err)
	}
	return runner
}

func validLeaseSession() BrowserSession {
	return BrowserSession{
		SessionID: "lease_session_1",
		CDPURL:    "wss://browser.example.test/cdp",
		BaseURL:   "https://browser.example.test",
	}
}

func validLeaseRequest() BrowserSessionLeaseRequest {
	return BrowserSessionLeaseRequest{
		Start: StartBrowserSessionRequest{
			TargetURL: "https://example.com/account",
			Purpose:   "perform one authorized direct Browser task",
		},
		CleanupPurpose: "terminate the direct Browser session after the authorized task",
	}
}

func TestBrowserSessionLeaseAlwaysTerminatesAfterSuccessfulWork(t *testing.T) {
	provider := &leaseBrowserProvider{startResult: validLeaseSession()}
	runner := newLeaseTestRunner(t, provider)
	workCalls := 0
	err := runner.Run(context.Background(), validLeaseRequest(), func(_ context.Context, session BrowserSession) error {
		workCalls++
		if session.SessionID != "lease_session_1" {
			t.Fatalf("unexpected session: %#v", session)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if provider.startCalls != 1 || workCalls != 1 || provider.terminateCalls != 1 {
		t.Fatalf("calls start=%d work=%d terminate=%d", provider.startCalls, workCalls, provider.terminateCalls)
	}
}

func TestBrowserSessionLeaseUsesFreshCleanupContextAfterCancellation(t *testing.T) {
	provider := &leaseBrowserProvider{startResult: validLeaseSession()}
	runner := newLeaseTestRunner(t, provider)
	ctx, cancel := context.WithCancel(context.Background())
	err := runner.Run(ctx, validLeaseRequest(), func(ctx context.Context, _ BrowserSession) error {
		cancel()
		return ctx.Err()
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected work cancellation, got %v", err)
	}
	if provider.terminateCalls != 1 {
		t.Fatalf("terminate calls = %d", provider.terminateCalls)
	}
	if provider.terminateCtxError != nil {
		t.Fatalf("cleanup inherited canceled caller context: %v", provider.terminateCtxError)
	}
}

func TestBrowserSessionLeaseJoinsWorkAndCleanupFailures(t *testing.T) {
	workErr := errors.New("controller failed")
	provider := &leaseBrowserProvider{startResult: validLeaseSession(), terminateErr: ErrBrowserSessionTerminationUnknown}
	runner := newLeaseTestRunner(t, provider)
	err := runner.Run(context.Background(), validLeaseRequest(), func(context.Context, BrowserSession) error {
		return workErr
	})
	if !errors.Is(err, workErr) || !errors.Is(err, ErrBrowserSessionTerminationUnknown) {
		t.Fatalf("expected joined work and cleanup errors, got %v", err)
	}
	if provider.terminateCalls != 1 {
		t.Fatalf("terminate calls = %d", provider.terminateCalls)
	}
}

func TestBrowserSessionLeaseSurfacesCleanupFailureAfterSuccessfulWork(t *testing.T) {
	provider := &leaseBrowserProvider{startResult: validLeaseSession(), terminateErr: ErrBrowserSessionTerminationUnknown}
	runner := newLeaseTestRunner(t, provider)
	err := runner.Run(context.Background(), validLeaseRequest(), func(context.Context, BrowserSession) error { return nil })
	if !errors.Is(err, ErrBrowserSessionTerminationUnknown) {
		t.Fatalf("expected cleanup failure, got %v", err)
	}
}

func TestBrowserSessionLeaseDoesNotRunWorkOrCleanupWhenCreateIsUnconfirmed(t *testing.T) {
	provider := &leaseBrowserProvider{startErr: ErrBrowserSessionCreationUnknown}
	runner := newLeaseTestRunner(t, provider)
	workCalls := 0
	err := runner.Run(context.Background(), validLeaseRequest(), func(context.Context, BrowserSession) error {
		workCalls++
		return nil
	})
	if !errors.Is(err, ErrBrowserSessionCreationUnknown) {
		t.Fatalf("expected create uncertainty, got %v", err)
	}
	if workCalls != 0 || provider.terminateCalls != 0 {
		t.Fatalf("unexpected work/cleanup after unconfirmed create: work=%d terminate=%d", workCalls, provider.terminateCalls)
	}
}

func TestBrowserSessionLeaseValidatesCleanupBeforeCreatingSession(t *testing.T) {
	provider := &leaseBrowserProvider{startResult: validLeaseSession()}
	runner := newLeaseTestRunner(t, provider)
	request := validLeaseRequest()
	request.CleanupPurpose = ""
	if err := runner.Run(context.Background(), request, func(context.Context, BrowserSession) error { return nil }); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid cleanup purpose, got %v", err)
	}
	if provider.startCalls != 0 {
		t.Fatalf("provider start called before cleanup validation: %d", provider.startCalls)
	}
	if err := runner.Run(context.Background(), validLeaseRequest(), nil); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected nil work rejection, got %v", err)
	}
	if provider.startCalls != 0 {
		t.Fatalf("provider start called before work validation: %d", provider.startCalls)
	}
}

func TestBrowserSessionLeaseCleanupTimeoutBounds(t *testing.T) {
	authorizer := browserSessionAuthorizerFunc(func(context.Context, BrowserSessionAuthorization) error { return nil })
	provider := &leaseBrowserProvider{}
	manager, err := NewBrowserSessionManager(authorizer, provider)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewBrowserSessionLeaseRunner(manager, 500*time.Millisecond); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected too-short cleanup timeout rejection, got %v", err)
	}
	if _, err := NewBrowserSessionLeaseRunner(manager, MaxBrowserSessionCleanupTimeout+time.Second); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected too-long cleanup timeout rejection, got %v", err)
	}
}
