package search

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestEnginePropagatesCallerCancellation(t *testing.T) {
	release := make(chan struct{})
	defer close(release)

	engine := NewEngine(time.Second, blockingProvider{name: "stuck", release: release})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	response, err := engine.Search(ctx, "superseded query")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("search error = %v, want context canceled", err)
	}
	if response.Query != "" || len(response.Results) != 0 || len(response.Providers) != 0 {
		t.Fatalf("canceled search must not manufacture partial response evidence: %#v", response)
	}
}

func TestEnginePropagatesCancellationWithoutConfiguredProviders(t *testing.T) {
	engine := NewEngine(time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	response, err := engine.Search(ctx, "providerless cancellation")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("search error = %v, want context canceled", err)
	}
	if response.Query != "" || len(response.Results) != 0 || len(response.Providers) != 0 {
		t.Fatalf("providerless canceled search must not return an empty success: %#v", response)
	}
}

func TestEnginePropagatesCallerDeadlineInsteadOfProviderTimeoutEvidence(t *testing.T) {
	release := make(chan struct{})
	defer close(release)

	engine := NewEngine(time.Second, blockingProvider{name: "stuck", release: release})
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	started := time.Now()
	response, err := engine.Search(ctx, "caller deadline")
	elapsed := time.Since(started)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("search error = %v, want caller deadline exceeded", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("caller deadline was not propagated promptly: %s", elapsed)
	}
	if response.Degraded || len(response.Providers) != 0 {
		t.Fatalf("caller deadline must not be rewritten as provider degradation: %#v", response)
	}
}
