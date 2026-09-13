package webintelligence

import (
	"errors"
	"testing"
	"time"
)

func newHealthTestController(t *testing.T) *Controller {
	t.Helper()
	controller, err := NewController([]Provider{{ID: "tinyfish-agent-v1", Capability: CapabilityAgent, Metered: true, Priority: 10}}, BudgetPolicy{LimitMicros: 1000000, PerOperationMicros: 100000})
	if err != nil {
		t.Fatalf("NewController() error = %v", err)
	}
	return controller
}

func recordSamples(t *testing.T, controller *Controller, outcome Outcome, duration time.Duration, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		if err := controller.Record("tinyfish-agent-v1", outcome, duration); err != nil {
			t.Fatalf("Record(%s) error = %v", outcome, err)
		}
	}
}

func TestEvaluateHealthCancellationTreatmentIsExplicit(t *testing.T) {
	controller := newHealthTestController(t)
	recordSamples(t, controller, OutcomeSuccess, 105*time.Second, 10)
	recordSamples(t, controller, OutcomeFailure, 640*time.Second, 8)
	recordSamples(t, controller, OutcomeCancelled, 200*time.Second, 5)

	policy := ProviderHealthPolicy{
		MinimumSamples:            18,
		DegradedSuccessPercent:    80,
		UnavailableSuccessPercent: 50,
	}
	evaluation, err := controller.EvaluateHealth("tinyfish-agent-v1", policy)
	if err != nil {
		t.Fatalf("EvaluateHealth() error = %v", err)
	}
	if !evaluation.EnoughSamples || evaluation.DecisionSamples != 18 {
		t.Fatalf("decision samples = %d, enough = %v; want 18,true", evaluation.DecisionSamples, evaluation.EnoughSamples)
	}
	if evaluation.SuccessRateBasisPoints != 5555 {
		t.Fatalf("success rate bps = %d; want 5555", evaluation.SuccessRateBasisPoints)
	}
	if evaluation.RecommendedHealth != HealthDegraded {
		t.Fatalf("recommended health = %q; want %q", evaluation.RecommendedHealth, HealthDegraded)
	}
	if evaluation.Cancelled != 5 {
		t.Fatalf("cancelled = %d; want 5", evaluation.Cancelled)
	}

	policy.CountCancelledAsFailure = true
	policy.MinimumSamples = 23
	evaluation, err = controller.EvaluateHealth("tinyfish-agent-v1", policy)
	if err != nil {
		t.Fatalf("EvaluateHealth(count cancelled) error = %v", err)
	}
	if evaluation.DecisionSamples != 23 || evaluation.SuccessRateBasisPoints != 4347 {
		t.Fatalf("decision samples/rate = %d/%d; want 23/4347", evaluation.DecisionSamples, evaluation.SuccessRateBasisPoints)
	}
	if evaluation.RecommendedHealth != HealthUnavailable {
		t.Fatalf("recommended health = %q; want %q", evaluation.RecommendedHealth, HealthUnavailable)
	}
}

func TestEvaluateHealthLatencyCanDegradeSuccessfulProvider(t *testing.T) {
	controller := newHealthTestController(t)
	recordSamples(t, controller, OutcomeSuccess, 650*time.Second, 4)

	evaluation, err := controller.EvaluateHealth("tinyfish-agent-v1", ProviderHealthPolicy{
		MinimumSamples:         4,
		DegradedSuccessPercent: 80,
		MaxAverageDurationMS:   600000,
	})
	if err != nil {
		t.Fatalf("EvaluateHealth() error = %v", err)
	}
	if evaluation.SuccessRateBasisPoints != 10000 {
		t.Fatalf("success rate bps = %d; want 10000", evaluation.SuccessRateBasisPoints)
	}
	if evaluation.AverageDurationMS != 650000 {
		t.Fatalf("average duration = %d; want 650000", evaluation.AverageDurationMS)
	}
	if !evaluation.LatencyDegraded || evaluation.RecommendedHealth != HealthDegraded {
		t.Fatalf("latency degraded/health = %v/%q; want true/%q", evaluation.LatencyDegraded, evaluation.RecommendedHealth, HealthDegraded)
	}
}

func TestApplyHealthPolicyCanRemoveUnavailableProviderFromPlan(t *testing.T) {
	controller := newHealthTestController(t)
	recordSamples(t, controller, OutcomeFailure, time.Second, 4)

	evaluation, err := controller.ApplyHealthPolicy("tinyfish-agent-v1", ProviderHealthPolicy{
		MinimumSamples:            4,
		DegradedSuccessPercent:    80,
		UnavailableSuccessPercent: 50,
	})
	if err != nil {
		t.Fatalf("ApplyHealthPolicy() error = %v", err)
	}
	if evaluation.RecommendedHealth != HealthUnavailable {
		t.Fatalf("recommended health = %q; want %q", evaluation.RecommendedHealth, HealthUnavailable)
	}
	_, err = controller.Plan(Request{Capability: CapabilityAgent, AllowMetered: true, EstimatedCostMicros: 1000, MaxAttempts: 1})
	if !errors.Is(err, ErrNoProvider) {
		t.Fatalf("Plan() error = %v; want ErrNoProvider", err)
	}
}

func TestApplyHealthPolicyInsufficientEvidencePreservesConfiguredHealth(t *testing.T) {
	controller := newHealthTestController(t)
	if err := controller.SetHealth("tinyfish-agent-v1", HealthDegraded); err != nil {
		t.Fatalf("SetHealth() error = %v", err)
	}
	recordSamples(t, controller, OutcomeSuccess, time.Second, 2)

	evaluation, err := controller.ApplyHealthPolicy("tinyfish-agent-v1", ProviderHealthPolicy{
		MinimumSamples:         5,
		DegradedSuccessPercent: 80,
	})
	if err != nil {
		t.Fatalf("ApplyHealthPolicy() error = %v", err)
	}
	if evaluation.EnoughSamples {
		t.Fatal("expected insufficient evidence")
	}
	snapshot := controller.Snapshot()
	if len(snapshot.Providers) != 1 || snapshot.Providers[0].Health != HealthDegraded {
		t.Fatalf("provider health = %+v; want degraded", snapshot.Providers)
	}
}

func TestHealthPolicyValidationRejectsInvertedThresholds(t *testing.T) {
	controller := newHealthTestController(t)
	_, err := controller.EvaluateHealth("tinyfish-agent-v1", ProviderHealthPolicy{
		MinimumSamples:            1,
		DegradedSuccessPercent:    40,
		UnavailableSuccessPercent: 60,
	})
	if !errors.Is(err, ErrInvalidOperation) {
		t.Fatalf("EvaluateHealth() error = %v; want ErrInvalidOperation", err)
	}
}
