package webintelligence

import (
	"context"
	"errors"
	"testing"
)

type fakeExecutor struct {
	id    string
	cap   Capability
	calls int
	run   func(context.Context, any) (ExecutionResult, error)
}

func (f *fakeExecutor) ProviderID() string     { return f.id }
func (f *fakeExecutor) Capability() Capability { return f.cap }
func (f *fakeExecutor) Execute(ctx context.Context, input any) (ExecutionResult, error) {
	f.calls++
	if f.run == nil {
		return ExecutionResult{}, nil
	}
	return f.run(ctx, input)
}

func mustController(t *testing.T, providers []Provider, budget BudgetPolicy) *Controller {
	t.Helper()
	controller, err := NewController(providers, budget)
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}
	return controller
}

func observationCount(snapshot Snapshot, providerID string, outcome Outcome) uint64 {
	for _, observation := range snapshot.Observations {
		if observation.ProviderID == providerID && observation.Outcome == outcome {
			return observation.Count
		}
	}
	return 0
}

func TestCoordinatorPrefersFreeProvider(t *testing.T) {
	controller := mustController(t, []Provider{
		{ID: "free-search", Capability: CapabilitySearch, Priority: 10},
		{ID: "metered-search", Capability: CapabilitySearch, Metered: true, Priority: 1},
	}, BudgetPolicy{LimitMicros: 1000, PerOperationMicros: 500})

	free := &fakeExecutor{id: "free-search", cap: CapabilitySearch, run: func(context.Context, any) (ExecutionResult, error) {
		return ExecutionResult{Value: "free"}, nil
	}}
	metered := &fakeExecutor{id: "metered-search", cap: CapabilitySearch, run: func(context.Context, any) (ExecutionResult, error) {
		return ExecutionResult{Value: "metered", CostMicros: 100, CostObserved: true}, nil
	}}
	coordinator, err := NewCoordinator(controller, []Executor{free, metered})
	if err != nil {
		t.Fatalf("NewCoordinator() error = %v", err)
	}

	result, err := coordinator.Execute(context.Background(), Request{
		Capability: CapabilitySearch,
		MaxAttempts: 1,
	}, "opaque")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.ProviderID != "free-search" || result.Value != "free" {
		t.Fatalf("Execute() result = %#v", result)
	}
	if free.calls != 1 || metered.calls != 0 {
		t.Fatalf("calls: free=%d metered=%d", free.calls, metered.calls)
	}
}

func TestCoordinatorFallsBackWithinCapability(t *testing.T) {
	controller := mustController(t, []Provider{
		{ID: "search-a", Capability: CapabilitySearch, Priority: 1},
		{ID: "search-b", Capability: CapabilitySearch, Priority: 2},
	}, BudgetPolicy{})

	first := &fakeExecutor{id: "search-a", cap: CapabilitySearch, run: func(context.Context, any) (ExecutionResult, error) {
		return ExecutionResult{}, errors.New("provider detail that must not escape")
	}}
	second := &fakeExecutor{id: "search-b", cap: CapabilitySearch, run: func(context.Context, any) (ExecutionResult, error) {
		return ExecutionResult{Value: "ok"}, nil
	}}
	coordinator, err := NewCoordinator(controller, []Executor{first, second})
	if err != nil {
		t.Fatalf("NewCoordinator() error = %v", err)
	}

	result, err := coordinator.Execute(context.Background(), Request{Capability: CapabilitySearch, MaxAttempts: 2}, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.ProviderID != "search-b" || result.Attempts != 2 {
		t.Fatalf("Execute() result = %#v", result)
	}
	snapshot := controller.Snapshot()
	if observationCount(snapshot, "search-a", OutcomeFailure) != 1 || observationCount(snapshot, "search-b", OutcomeSuccess) != 1 {
		t.Fatalf("unexpected observations: %#v", snapshot.Observations)
	}
}

func TestCoordinatorDoesNotEscalateCapability(t *testing.T) {
	controller := mustController(t, []Provider{
		{ID: "search", Capability: CapabilitySearch},
		{ID: "fetch", Capability: CapabilityFetch},
	}, BudgetPolicy{})
	search := &fakeExecutor{id: "search", cap: CapabilitySearch, run: func(context.Context, any) (ExecutionResult, error) {
		return ExecutionResult{}, errors.New("search failed")
	}}
	fetch := &fakeExecutor{id: "fetch", cap: CapabilityFetch, run: func(context.Context, any) (ExecutionResult, error) {
		return ExecutionResult{Value: "must-not-run"}, nil
	}}
	coordinator, err := NewCoordinator(controller, []Executor{search, fetch})
	if err != nil {
		t.Fatalf("NewCoordinator() error = %v", err)
	}

	_, err = coordinator.Execute(context.Background(), Request{Capability: CapabilitySearch, MaxAttempts: 2}, nil)
	if !errors.Is(err, ErrExecutionFailed) {
		t.Fatalf("Execute() error = %v, want ErrExecutionFailed", err)
	}
	if fetch.calls != 0 {
		t.Fatalf("fetch executor calls = %d, want 0", fetch.calls)
	}
}

func TestCoordinatorRejectsExecutorCapabilityMismatch(t *testing.T) {
	controller := mustController(t, []Provider{{ID: "provider", Capability: CapabilitySearch}}, BudgetPolicy{})
	_, err := NewCoordinator(controller, []Executor{&fakeExecutor{id: "provider", cap: CapabilityFetch}})
	if !errors.Is(err, ErrExecutorContract) {
		t.Fatalf("NewCoordinator() error = %v, want ErrExecutorContract", err)
	}
}

func TestCoordinatorFailsClosedWhenExecutorMissing(t *testing.T) {
	controller := mustController(t, []Provider{{ID: "search", Capability: CapabilitySearch}}, BudgetPolicy{})
	coordinator, err := NewCoordinator(controller, nil)
	if err != nil {
		t.Fatalf("NewCoordinator() error = %v", err)
	}
	_, err = coordinator.Execute(context.Background(), Request{Capability: CapabilitySearch}, nil)
	if !errors.Is(err, ErrExecutorUnavailable) {
		t.Fatalf("Execute() error = %v, want ErrExecutorUnavailable", err)
	}
	if observationCount(controller.Snapshot(), "search", OutcomeRejected) != 1 {
		t.Fatalf("missing executor rejection was not recorded")
	}
}

func TestCoordinatorCommitsMeteredActualCost(t *testing.T) {
	controller := mustController(t, []Provider{{ID: "research", Capability: CapabilityResearch, Metered: true}}, BudgetPolicy{
		LimitMicros: 1000, PerOperationMicros: 500,
	})
	executor := &fakeExecutor{id: "research", cap: CapabilityResearch, run: func(context.Context, any) (ExecutionResult, error) {
		return ExecutionResult{Value: "report", CostMicros: 300, CostObserved: true}, nil
	}}
	coordinator, err := NewCoordinator(controller, []Executor{executor})
	if err != nil {
		t.Fatalf("NewCoordinator() error = %v", err)
	}
	result, err := coordinator.Execute(context.Background(), Request{
		Capability: CapabilityResearch, AllowMetered: true, EstimatedCostMicros: 400,
	}, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.CostMicros != 300 {
		t.Fatalf("result cost = %d, want 300", result.CostMicros)
	}
	budget := controller.Snapshot().Budget
	if budget.SpentMicros != 300 || budget.ReservedMicros != 0 || budget.AvailableMicros != 700 {
		t.Fatalf("budget = %#v", budget)
	}
}

func TestCoordinatorCommitsFailedMeteredAttemptBeforeFallback(t *testing.T) {
	controller := mustController(t, []Provider{
		{ID: "research-a", Capability: CapabilityResearch, Metered: true, Priority: 1},
		{ID: "research-b", Capability: CapabilityResearch, Metered: true, Priority: 2},
	}, BudgetPolicy{LimitMicros: 2000, PerOperationMicros: 1000})
	first := &fakeExecutor{id: "research-a", cap: CapabilityResearch, run: func(context.Context, any) (ExecutionResult, error) {
		return ExecutionResult{CostMicros: 100, CostObserved: true}, errors.New("failed after metered work")
	}}
	second := &fakeExecutor{id: "research-b", cap: CapabilityResearch, run: func(context.Context, any) (ExecutionResult, error) {
		return ExecutionResult{Value: "ok", CostMicros: 200, CostObserved: true}, nil
	}}
	coordinator, err := NewCoordinator(controller, []Executor{first, second})
	if err != nil {
		t.Fatalf("NewCoordinator() error = %v", err)
	}
	result, err := coordinator.Execute(context.Background(), Request{
		Capability: CapabilityResearch, AllowMetered: true, EstimatedCostMicros: 400, MaxAttempts: 2,
	}, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.CostMicros != 300 || controller.Snapshot().Budget.SpentMicros != 300 {
		t.Fatalf("result=%#v budget=%#v", result, controller.Snapshot().Budget)
	}
}

func TestCoordinatorReleasesUnusedMeteredReservationAfterFreeSuccess(t *testing.T) {
	controller := mustController(t, []Provider{
		{ID: "free", Capability: CapabilitySearch, Priority: 1},
		{ID: "metered", Capability: CapabilitySearch, Metered: true, Priority: 2},
	}, BudgetPolicy{LimitMicros: 1000, PerOperationMicros: 500})
	free := &fakeExecutor{id: "free", cap: CapabilitySearch, run: func(context.Context, any) (ExecutionResult, error) {
		return ExecutionResult{Value: "ok"}, nil
	}}
	metered := &fakeExecutor{id: "metered", cap: CapabilitySearch, run: func(context.Context, any) (ExecutionResult, error) {
		return ExecutionResult{CostMicros: 100, CostObserved: true}, nil
	}}
	coordinator, err := NewCoordinator(controller, []Executor{free, metered})
	if err != nil {
		t.Fatalf("NewCoordinator() error = %v", err)
	}
	_, err = coordinator.Execute(context.Background(), Request{
		Capability: CapabilitySearch, AllowMetered: true, EstimatedCostMicros: 100, MaxAttempts: 2,
	}, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	budget := controller.Snapshot().Budget
	if budget.ReservedMicros != 0 || budget.SpentMicros != 0 {
		t.Fatalf("budget = %#v", budget)
	}
	if metered.calls != 0 {
		t.Fatalf("metered calls = %d, want 0", metered.calls)
	}
}

func TestCoordinatorPreservesCancellationAndAccountsCost(t *testing.T) {
	controller := mustController(t, []Provider{{ID: "agent", Capability: CapabilityAgent, Metered: true}}, BudgetPolicy{
		LimitMicros: 1000, PerOperationMicros: 500,
	})
	executor := &fakeExecutor{id: "agent", cap: CapabilityAgent, run: func(context.Context, any) (ExecutionResult, error) {
		return ExecutionResult{CostMicros: 50, CostObserved: true}, context.Canceled
	}}
	coordinator, err := NewCoordinator(controller, []Executor{executor})
	if err != nil {
		t.Fatalf("NewCoordinator() error = %v", err)
	}
	_, err = coordinator.Execute(context.Background(), Request{
		Capability: CapabilityAgent, AllowMetered: true, EstimatedCostMicros: 100,
	}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Execute() error = %v, want context.Canceled", err)
	}
	if controller.Snapshot().Budget.SpentMicros != 50 {
		t.Fatalf("spent = %d, want 50", controller.Snapshot().Budget.SpentMicros)
	}
	if observationCount(controller.Snapshot(), "agent", OutcomeCancelled) != 1 {
		t.Fatalf("cancellation observation missing")
	}
}

func TestCoordinatorRetainsReservationWhenMeteredCostUnverified(t *testing.T) {
	controller := mustController(t, []Provider{{ID: "agent", Capability: CapabilityAgent, Metered: true}}, BudgetPolicy{
		LimitMicros: 1000, PerOperationMicros: 500,
	})
	executor := &fakeExecutor{id: "agent", cap: CapabilityAgent, run: func(context.Context, any) (ExecutionResult, error) {
		return ExecutionResult{Value: "provider says success but billing is unresolved"}, nil
	}}
	coordinator, err := NewCoordinator(controller, []Executor{executor})
	if err != nil {
		t.Fatalf("NewCoordinator() error = %v", err)
	}
	_, err = coordinator.Execute(context.Background(), Request{
		Capability: CapabilityAgent, AllowMetered: true, EstimatedCostMicros: 100,
	}, nil)
	if !errors.Is(err, ErrCostEvidenceRequired) {
		t.Fatalf("Execute() error = %v, want ErrCostEvidenceRequired", err)
	}
	budget := controller.Snapshot().Budget
	if budget.ReservedMicros != 100 || budget.SpentMicros != 0 || budget.AvailableMicros != 900 {
		t.Fatalf("budget = %#v", budget)
	}
	if observationCount(controller.Snapshot(), "agent", OutcomeRejected) != 1 {
		t.Fatalf("unverified-cost rejection missing")
	}
}

func TestCoordinatorRejectsCostFromFreeProvider(t *testing.T) {
	controller := mustController(t, []Provider{{ID: "fetch", Capability: CapabilityFetch}}, BudgetPolicy{})
	executor := &fakeExecutor{id: "fetch", cap: CapabilityFetch, run: func(context.Context, any) (ExecutionResult, error) {
		return ExecutionResult{Value: "content", CostMicros: 1}, nil
	}}
	coordinator, err := NewCoordinator(controller, []Executor{executor})
	if err != nil {
		t.Fatalf("NewCoordinator() error = %v", err)
	}
	_, err = coordinator.Execute(context.Background(), Request{Capability: CapabilityFetch}, nil)
	if !errors.Is(err, ErrExecutorContract) {
		t.Fatalf("Execute() error = %v, want ErrExecutorContract", err)
	}
	if observationCount(controller.Snapshot(), "fetch", OutcomeRejected) != 1 {
		t.Fatalf("free-provider cost rejection missing")
	}
}
