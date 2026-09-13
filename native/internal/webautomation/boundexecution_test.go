package webautomation

import (
	"context"
	"errors"
	"testing"
	"time"
)

type recordingExecutor struct {
	request Request
	result  Result
	err     error
	calls   int
}

func (e *recordingExecutor) Run(_ context.Context, request Request) (Result, error) {
	e.calls++
	e.request = request
	return e.result, e.err
}

func testBoundAuthExecutor(t *testing.T, withVault bool, executor *recordingExecutor) (*BoundAuthExecutor, *AuthAcceptanceRegistry) {
	t.Helper()
	binding := AuthBinding{Host: "example.com", ProfileID: "prof_example", BrowserProfile: BrowserProfileLite}
	if withVault {
		binding.CredentialItemIDs = []string{"vault://example-login"}
	}
	bindings, err := NewAuthBindings([]AuthBinding{binding})
	if err != nil {
		t.Fatal(err)
	}
	acceptance, err := NewAuthAcceptanceRegistry(bindings)
	if err != nil {
		t.Fatal(err)
	}
	bound, err := NewBoundAuthExecutor(bindings, acceptance, executor)
	if err != nil {
		t.Fatal(err)
	}
	bound.now = func() time.Time { return time.Date(2026, 9, 13, 2, 30, 0, 0, time.UTC) }
	return bound, acceptance
}

func TestBoundAuthExecutorAppliesBindingAndRecordsReuse(t *testing.T) {
	executor := &recordingExecutor{result: Result{Capability: CapabilityAgent, Output: "ok"}}
	bound, acceptance := testBoundAuthExecutor(t, true, executor)
	result, err := bound.Run(context.Background(), Request{
		URL:        "https://example.com/account",
		Goal:       "Read the authenticated account heading without changing anything.",
		Purpose:    "authenticated acceptance verification",
		Capability: CapabilityAgent,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Output != "ok" || executor.calls != 1 {
		t.Fatalf("unexpected execution result: %#v calls=%d", result, executor.calls)
	}
	if !executor.request.UseProfile || executor.request.ProfileID != "prof_example" {
		t.Fatalf("managed profile was not applied: %#v", executor.request)
	}
	if !executor.request.UseVault || len(executor.request.CredentialItemIDs) != 1 || executor.request.CredentialItemIDs[0] != "vault://example-login" {
		t.Fatalf("scoped Vault reference was not applied: %#v", executor.request)
	}
	status, err := acceptance.Status("example.com")
	if err != nil {
		t.Fatal(err)
	}
	if status.State != AuthAcceptanceSessionReuseVerified || !status.SessionReuseVerified || !status.VaultConfigured {
		t.Fatalf("unexpected acceptance status: %#v", status)
	}
	if status.VaultRepairVerified {
		t.Fatal("successful run must not infer Vault repair")
	}
}

func TestBoundAuthExecutorRecordsNormalizedSiteBlocked(t *testing.T) {
	executor := &recordingExecutor{err: ErrSiteBlocked}
	bound, acceptance := testBoundAuthExecutor(t, false, executor)
	_, err := bound.Run(context.Background(), Request{URL: "https://example.com/private", Goal: "Verify access.", Purpose: "acceptance", Capability: CapabilityAgent})
	if !errors.Is(err, ErrSiteBlocked) {
		t.Fatalf("expected site blocked, got %v", err)
	}
	status, err := acceptance.Status("example.com")
	if err != nil {
		t.Fatal(err)
	}
	if status.State != AuthAcceptanceSiteBlocked || status.ObservationCount != 1 {
		t.Fatalf("unexpected status: %#v", status)
	}
}

func TestBoundAuthExecutorRecordsGoalFailureButNotProviderFailure(t *testing.T) {
	goalExecutor := &recordingExecutor{err: ErrGoalFailed}
	goalBound, goalAcceptance := testBoundAuthExecutor(t, false, goalExecutor)
	_, err := goalBound.Run(context.Background(), Request{URL: "https://example.com/private", Goal: "Verify access.", Purpose: "acceptance", Capability: CapabilityAgent})
	if !errors.Is(err, ErrGoalFailed) {
		t.Fatalf("expected goal failure, got %v", err)
	}
	status, _ := goalAcceptance.Status("example.com")
	if status.State != AuthAcceptanceFailed || status.ObservationCount != 1 {
		t.Fatalf("unexpected goal-failure status: %#v", status)
	}

	providerExecutor := &recordingExecutor{err: ErrAutomationFailed}
	providerBound, providerAcceptance := testBoundAuthExecutor(t, false, providerExecutor)
	_, err = providerBound.Run(context.Background(), Request{URL: "https://example.com/private", Goal: "Verify access.", Purpose: "acceptance", Capability: CapabilityAgent})
	if !errors.Is(err, ErrAutomationFailed) {
		t.Fatalf("expected provider failure, got %v", err)
	}
	status, _ = providerAcceptance.Status("example.com")
	if status.State != AuthAcceptanceUnverified || status.ObservationCount != 0 {
		t.Fatalf("provider failure must not become auth evidence: %#v", status)
	}
}

func TestBoundAuthExecutorRejectsUnboundAndCallerAuthOverrides(t *testing.T) {
	executor := &recordingExecutor{}
	bound, _ := testBoundAuthExecutor(t, false, executor)
	_, err := bound.Run(context.Background(), Request{URL: "https://other.example.com/", Goal: "Verify access.", Purpose: "acceptance", Capability: CapabilityAgent})
	if !errors.Is(err, ErrAuthBindingUnavailable) || executor.calls != 0 {
		t.Fatalf("expected unbound request rejection before execution, err=%v calls=%d", err, executor.calls)
	}
	_, err = bound.Run(context.Background(), Request{URL: "https://example.com/", Goal: "Verify access.", Purpose: "acceptance", Capability: CapabilityAgent, UseProfile: true, ProfileID: "prof_other"})
	if !errors.Is(err, ErrInvalidRequest) || executor.calls != 0 {
		t.Fatalf("expected caller auth override rejection before execution, err=%v calls=%d", err, executor.calls)
	}
}

func TestBoundAuthExecutorDoesNotRecordCancellation(t *testing.T) {
	executor := &recordingExecutor{err: context.Canceled}
	bound, acceptance := testBoundAuthExecutor(t, false, executor)
	_, err := bound.Run(context.Background(), Request{URL: "https://example.com/", Goal: "Verify access.", Purpose: "acceptance", Capability: CapabilityAgent})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	status, _ := acceptance.Status("example.com")
	if status.State != AuthAcceptanceUnverified || status.ObservationCount != 0 {
		t.Fatalf("cancellation must not become auth evidence: %#v", status)
	}
}
