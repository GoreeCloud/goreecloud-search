package webintelligence

import (
	"errors"
	"testing"
	"time"
)

func TestPlanUsesExactCapabilityAndPrefersFreeProvider(t *testing.T) {
	controller, err := NewController([]Provider{
		{ID: "tinyfish-search-free", Capability: CapabilitySearch, Metered: false, Priority: 50},
		{ID: "search-metered", Capability: CapabilitySearch, Metered: true, Priority: 0},
		{ID: "tinyfish-fetch-free", Capability: CapabilityFetch, Metered: false, Priority: 0},
	}, BudgetPolicy{LimitMicros: 1_000_000, PerOperationMicros: 500_000})
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}

	plan, err := controller.Plan(Request{
		Capability:          CapabilitySearch,
		AllowMetered:        true,
		EstimatedCostMicros: 125_000,
		MaxAttempts:         2,
	})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if plan.Capability != CapabilitySearch {
		t.Fatalf("Plan().Capability = %q, want %q", plan.Capability, CapabilitySearch)
	}
	if len(plan.Providers) != 2 {
		t.Fatalf("len(Plan().Providers) = %d, want 2", len(plan.Providers))
	}
	if got := plan.Providers[0].ID; got != "tinyfish-search-free" {
		t.Fatalf("first provider = %q, want free search provider", got)
	}
	if got := plan.Providers[1].ID; got != "search-metered" {
		t.Fatalf("second provider = %q, want metered search fallback", got)
	}
	if plan.ReservationID == 0 {
		t.Fatal("Plan().ReservationID = 0, want reservation for metered fallback")
	}
	if plan.ReservedCostMicros != 125_000 {
		t.Fatalf("Plan().ReservedCostMicros = %d, want 125000", plan.ReservedCostMicros)
	}

	for _, provider := range plan.Providers {
		if provider.Capability != CapabilitySearch {
			t.Fatalf("provider %q capability = %q, want exact search capability", provider.ID, provider.Capability)
		}
	}
}

func TestPlanCanStayFreeWithoutBudget(t *testing.T) {
	controller, err := NewController([]Provider{
		{ID: "tinyfish-search-v1", Capability: CapabilitySearch, Metered: false},
		{ID: "metered-search", Capability: CapabilitySearch, Metered: true},
	}, BudgetPolicy{})
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}

	plan, err := controller.Plan(Request{Capability: CapabilitySearch, MaxAttempts: 2})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if len(plan.Providers) != 1 || plan.Providers[0].ID != "tinyfish-search-v1" {
		t.Fatalf("Plan().Providers = %#v, want only free provider", plan.Providers)
	}
	if plan.ReservationID != 0 || plan.ReservedCostMicros != 0 {
		t.Fatalf("free plan reserved budget: id=%d cost=%d", plan.ReservationID, plan.ReservedCostMicros)
	}
}

func TestMeteredPlanRequiresConfiguredBudgetAndEstimate(t *testing.T) {
	controller, err := NewController([]Provider{
		{ID: "tinyfish-research", Capability: CapabilityResearch, Metered: true},
	}, BudgetPolicy{})
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}

	_, err = controller.Plan(Request{
		Capability:          CapabilityResearch,
		AllowMetered:        true,
		EstimatedCostMicros: 100,
	})
	if !errors.Is(err, ErrBudgetRequired) {
		t.Fatalf("Plan() error = %v, want ErrBudgetRequired", err)
	}

	controller, err = NewController([]Provider{
		{ID: "tinyfish-research", Capability: CapabilityResearch, Metered: true},
	}, BudgetPolicy{LimitMicros: 1_000, PerOperationMicros: 500})
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}
	_, err = controller.Plan(Request{Capability: CapabilityResearch, AllowMetered: true})
	if !errors.Is(err, ErrBudgetRequired) {
		t.Fatalf("Plan() without estimate error = %v, want ErrBudgetRequired", err)
	}
}

func TestBudgetReservationsEnforcePerOperationAndTotalCeilings(t *testing.T) {
	controller, err := NewController([]Provider{
		{ID: "agent-a", Capability: CapabilityAgent, Metered: true, Priority: 0},
		{ID: "agent-b", Capability: CapabilityAgent, Metered: true, Priority: 1},
	}, BudgetPolicy{LimitMicros: 500, PerOperationMicros: 300})
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}

	_, err = controller.Plan(Request{
		Capability:          CapabilityAgent,
		AllowMetered:        true,
		EstimatedCostMicros: 200,
		MaxAttempts:         2,
	})
	if !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("Plan() error = %v, want per-operation ErrBudgetExceeded", err)
	}

	first, err := controller.Plan(Request{
		Capability:          CapabilityAgent,
		AllowMetered:        true,
		EstimatedCostMicros: 100,
		MaxAttempts:         2,
	})
	if err != nil {
		t.Fatalf("first Plan() error = %v", err)
	}
	second, err := controller.Plan(Request{
		Capability:          CapabilityAgent,
		AllowMetered:        true,
		EstimatedCostMicros: 100,
		MaxAttempts:         2,
	})
	if err != nil {
		t.Fatalf("second Plan() error = %v", err)
	}

	_, err = controller.Plan(Request{
		Capability:          CapabilityAgent,
		AllowMetered:        true,
		EstimatedCostMicros: 150,
		MaxAttempts:         1,
	})
	if !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("third Plan() error = %v, want total-budget ErrBudgetExceeded", err)
	}

	snapshot := controller.Snapshot()
	if snapshot.Budget.ReservedMicros != 400 || snapshot.Budget.AvailableMicros != 100 {
		t.Fatalf("budget snapshot = %#v, want reserved=400 available=100", snapshot.Budget)
	}

	if err := controller.Release(first.ReservationID); err != nil {
		t.Fatalf("Release(first) error = %v", err)
	}
	if err := controller.Commit(second.ReservationID, 150); err != nil {
		t.Fatalf("Commit(second) error = %v", err)
	}
	snapshot = controller.Snapshot()
	if snapshot.Budget.ReservedMicros != 0 || snapshot.Budget.SpentMicros != 150 || snapshot.Budget.AvailableMicros != 350 {
		t.Fatalf("budget snapshot after release/commit = %#v", snapshot.Budget)
	}
}

func TestCommitRecordsRealSpendWhenReservationIsExceeded(t *testing.T) {
	controller, err := NewController([]Provider{
		{ID: "tinyfish-agent", Capability: CapabilityAgent, Metered: true},
	}, BudgetPolicy{LimitMicros: 1_000, PerOperationMicros: 500})
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}
	plan, err := controller.Plan(Request{
		Capability:          CapabilityAgent,
		AllowMetered:        true,
		EstimatedCostMicros: 200,
	})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	err = controller.Commit(plan.ReservationID, 250)
	if !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("Commit() error = %v, want ErrBudgetExceeded", err)
	}
	snapshot := controller.Snapshot()
	if snapshot.Budget.ReservedMicros != 0 || snapshot.Budget.SpentMicros != 250 || snapshot.Budget.AvailableMicros != 750 {
		t.Fatalf("budget snapshot = %#v, want real spend recorded", snapshot.Budget)
	}
}

func TestHealthControlsProviderEligibility(t *testing.T) {
	controller, err := NewController([]Provider{
		{ID: "search-a", Capability: CapabilitySearch, Metered: false, Priority: 0},
		{ID: "search-b", Capability: CapabilitySearch, Metered: false, Priority: 1},
	}, BudgetPolicy{})
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}
	if err := controller.SetHealth("search-a", HealthUnavailable); err != nil {
		t.Fatalf("SetHealth(search-a) error = %v", err)
	}
	if err := controller.SetHealth("search-b", HealthDegraded); err != nil {
		t.Fatalf("SetHealth(search-b) error = %v", err)
	}

	plan, err := controller.Plan(Request{Capability: CapabilitySearch, MaxAttempts: 2})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if len(plan.Providers) != 1 || plan.Providers[0].ID != "search-b" {
		t.Fatalf("Plan().Providers = %#v, want only degraded but available search-b", plan.Providers)
	}

	if err := controller.SetHealth("missing", HealthHealthy); !errors.Is(err, ErrUnknownProvider) {
		t.Fatalf("SetHealth(missing) error = %v, want ErrUnknownProvider", err)
	}
}

func TestRecordAggregatesOnlyOperationalMetadata(t *testing.T) {
	controller, err := NewController([]Provider{
		{ID: "tinyfish-fetch-v1", Capability: CapabilityFetch, Metered: false},
	}, BudgetPolicy{})
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}

	if err := controller.Record("tinyfish-fetch-v1", OutcomeSuccess, 1500*time.Millisecond); err != nil {
		t.Fatalf("Record(success 1) error = %v", err)
	}
	if err := controller.Record("tinyfish-fetch-v1", OutcomeSuccess, 500*time.Millisecond); err != nil {
		t.Fatalf("Record(success 2) error = %v", err)
	}
	if err := controller.Record("tinyfish-fetch-v1", OutcomeFailure, -time.Second); err != nil {
		t.Fatalf("Record(failure) error = %v", err)
	}

	snapshot := controller.Snapshot()
	if len(snapshot.Observations) != 2 {
		t.Fatalf("len(Snapshot().Observations) = %d, want 2", len(snapshot.Observations))
	}
	var success, failure Observation
	for _, observation := range snapshot.Observations {
		switch observation.Outcome {
		case OutcomeSuccess:
			success = observation
		case OutcomeFailure:
			failure = observation
		}
	}
	if success.Count != 2 || success.DurationMS != 2000 {
		t.Fatalf("success observation = %#v, want count=2 duration=2000ms", success)
	}
	if failure.Count != 1 || failure.DurationMS != 0 {
		t.Fatalf("failure observation = %#v, want count=1 duration=0ms", failure)
	}
	if err := controller.Record("missing", OutcomeSuccess, time.Second); !errors.Is(err, ErrUnknownProvider) {
		t.Fatalf("Record(missing) error = %v, want ErrUnknownProvider", err)
	}
}

func TestValidationFailsClosed(t *testing.T) {
	if _, err := NewController([]Provider{
		{ID: "bad provider", Capability: CapabilitySearch},
	}, BudgetPolicy{}); !errors.Is(err, ErrInvalidOperation) {
		t.Fatalf("NewController(invalid provider) error = %v, want ErrInvalidOperation", err)
	}
	if _, err := NewController([]Provider{
		{ID: "duplicate", Capability: CapabilitySearch},
		{ID: "duplicate", Capability: CapabilitySearch},
	}, BudgetPolicy{}); !errors.Is(err, ErrInvalidOperation) {
		t.Fatalf("NewController(duplicate provider) error = %v, want ErrInvalidOperation", err)
	}
	if _, err := NewController(nil, BudgetPolicy{LimitMicros: 100, PerOperationMicros: 101}); !errors.Is(err, ErrInvalidOperation) {
		t.Fatalf("NewController(invalid budget) error = %v, want ErrInvalidOperation", err)
	}

	controller, err := NewController([]Provider{{ID: "search", Capability: CapabilitySearch}}, BudgetPolicy{})
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}
	if _, err := controller.Plan(Request{Capability: Capability("invalid")}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("Plan(invalid capability) error = %v, want ErrInvalidRequest", err)
	}
	if _, err := controller.Plan(Request{Capability: CapabilitySearch, MaxAttempts: MaxPlanAttempts + 1}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("Plan(too many attempts) error = %v, want ErrInvalidRequest", err)
	}
	if err := controller.Commit(0, 1); !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("Commit(unreserved cost) error = %v, want ErrBudgetExceeded", err)
	}
	if err := controller.Release(999); !errors.Is(err, ErrInvalidOperation) {
		t.Fatalf("Release(unknown reservation) error = %v, want ErrInvalidOperation", err)
	}
}
