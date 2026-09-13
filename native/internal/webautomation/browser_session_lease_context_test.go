package webautomation

import (
	"context"
	"errors"
	"testing"
)

type cleanupContextKey struct{}

func TestBrowserSessionLeasePreservesContextValuesForCleanupAfterCancellation(t *testing.T) {
	provider := &leaseBrowserProvider{startResult: validLeaseSession()}
	var terminateValue any
	authorizer := browserSessionAuthorizerFunc(func(ctx context.Context, input BrowserSessionAuthorization) error {
		if input.Operation == BrowserSessionOperationTerminate {
			terminateValue = ctx.Value(cleanupContextKey{})
			if ctx.Err() != nil {
				t.Fatalf("cleanup authorization inherited cancellation: %v", ctx.Err())
			}
		}
		return nil
	})
	manager, err := NewBrowserSessionManager(authorizer, provider)
	if err != nil {
		t.Fatal(err)
	}
	runner, err := NewBrowserSessionLeaseRunner(manager, 0)
	if err != nil {
		t.Fatal(err)
	}

	base := context.WithValue(context.Background(), cleanupContextKey{}, "request-authority")
	ctx, cancel := context.WithCancel(base)
	err = runner.Run(ctx, validLeaseRequest(), func(ctx context.Context, _ BrowserSession) error {
		cancel()
		return ctx.Err()
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected work cancellation, got %v", err)
	}
	if terminateValue != "request-authority" {
		t.Fatalf("cleanup lost caller-scoped context value: %#v", terminateValue)
	}
}

func TestBrowserSessionLeaseRejectsNilContextBeforeCreation(t *testing.T) {
	provider := &leaseBrowserProvider{startResult: validLeaseSession()}
	runner := newLeaseTestRunner(t, provider)

	if err := runner.Run(nil, validLeaseRequest(), func(context.Context, BrowserSession) error { return nil }); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid request for nil context, got %v", err)
	}
	if provider.startCalls != 0 {
		t.Fatalf("provider start called with invalid nil context: %d", provider.startCalls)
	}
}
