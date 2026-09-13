package webintelligence

import (
	"fmt"
	"strings"
	"time"
)

const MaxHealthPolicyMinimumSamples = 100000

// ProviderHealthPolicy defines a bounded, provider-neutral reliability gate over
// privacy-minimized Controller observations. Percent thresholds are whole
// percentages (0-100); zero disables that threshold. Cancelled and policy-
// rejected operations are excluded by default because they are not necessarily
// evidence of provider failure.
type ProviderHealthPolicy struct {
	MinimumSamples            uint64 `json:"minimum_samples"`
	DegradedSuccessPercent    uint64 `json:"degraded_success_percent"`
	UnavailableSuccessPercent uint64 `json:"unavailable_success_percent"`
	MaxAverageDurationMS      uint64 `json:"max_average_duration_ms,omitempty"`
	CountCancelledAsFailure   bool   `json:"count_cancelled_as_failure"`
	CountRejectedAsFailure    bool   `json:"count_rejected_as_failure"`
}

// ProviderHealthEvaluation is operational evidence only. It does not prove
// production acceptance, Privacy Shield authorization, or Wardveil protection.
// It intentionally contains no query, URL, goal, page, credential, citation, or
// provider-response content.
type ProviderHealthEvaluation struct {
	ProviderID             string `json:"provider_id"`
	CurrentHealth          Health `json:"current_health"`
	RecommendedHealth      Health `json:"recommended_health"`
	EnoughSamples          bool   `json:"enough_samples"`
	DecisionSamples        uint64 `json:"decision_samples"`
	Successes              uint64 `json:"successes"`
	Failures               uint64 `json:"failures"`
	Timeouts               uint64 `json:"timeouts"`
	Cancelled              uint64 `json:"cancelled"`
	Rejected               uint64 `json:"rejected"`
	SuccessRateBasisPoints uint64 `json:"success_rate_basis_points"`
	AverageDurationMS      uint64 `json:"average_duration_ms"`
	LatencyDegraded        bool   `json:"latency_degraded"`
}

func normalizeProviderHealthPolicy(input ProviderHealthPolicy) (ProviderHealthPolicy, error) {
	if input.MinimumSamples < 1 || input.MinimumSamples > MaxHealthPolicyMinimumSamples {
		return ProviderHealthPolicy{}, fmt.Errorf("%w: health policy minimum samples must be between 1 and %d", ErrInvalidOperation, MaxHealthPolicyMinimumSamples)
	}
	if input.DegradedSuccessPercent > 100 || input.UnavailableSuccessPercent > 100 {
		return ProviderHealthPolicy{}, fmt.Errorf("%w: health policy success thresholds must be between 0 and 100", ErrInvalidOperation)
	}
	if input.DegradedSuccessPercent > 0 && input.UnavailableSuccessPercent > input.DegradedSuccessPercent {
		return ProviderHealthPolicy{}, fmt.Errorf("%w: unavailable success threshold cannot exceed degraded threshold", ErrInvalidOperation)
	}
	return input, nil
}

// EvaluateHealth computes a point-in-time recommendation from the Controller's
// in-process aggregate observations. It never mutates provider health. Because
// the current observation store is cumulative for the runtime lifetime, callers
// must not describe this as a rolling-window or durable service-level measure.
func (c *Controller) EvaluateHealth(providerID string, policy ProviderHealthPolicy) (ProviderHealthEvaluation, error) {
	if c == nil {
		return ProviderHealthEvaluation{}, ErrUnknownProvider
	}
	normalized, err := normalizeProviderHealthPolicy(policy)
	if err != nil {
		return ProviderHealthEvaluation{}, err
	}
	providerID = strings.TrimSpace(providerID)

	c.mu.RLock()
	defer c.mu.RUnlock()
	state, exists := c.providers[providerID]
	if !exists {
		return ProviderHealthEvaluation{}, ErrUnknownProvider
	}

	evaluation := ProviderHealthEvaluation{
		ProviderID:        providerID,
		CurrentHealth:     state.health,
		RecommendedHealth: state.health,
	}
	var decisionDurationMS uint64
	for key, value := range c.observations {
		if key.providerID != providerID {
			continue
		}
		switch key.outcome {
		case OutcomeSuccess:
			evaluation.Successes = saturatingAdd(evaluation.Successes, value.count)
			evaluation.DecisionSamples = saturatingAdd(evaluation.DecisionSamples, value.count)
			decisionDurationMS = saturatingAdd(decisionDurationMS, value.durationMS)
		case OutcomeFailure:
			evaluation.Failures = saturatingAdd(evaluation.Failures, value.count)
			evaluation.DecisionSamples = saturatingAdd(evaluation.DecisionSamples, value.count)
			decisionDurationMS = saturatingAdd(decisionDurationMS, value.durationMS)
		case OutcomeTimeout:
			evaluation.Timeouts = saturatingAdd(evaluation.Timeouts, value.count)
			evaluation.DecisionSamples = saturatingAdd(evaluation.DecisionSamples, value.count)
			decisionDurationMS = saturatingAdd(decisionDurationMS, value.durationMS)
		case OutcomeCancelled:
			evaluation.Cancelled = saturatingAdd(evaluation.Cancelled, value.count)
			if normalized.CountCancelledAsFailure {
				evaluation.DecisionSamples = saturatingAdd(evaluation.DecisionSamples, value.count)
				decisionDurationMS = saturatingAdd(decisionDurationMS, value.durationMS)
			}
		case OutcomeRejected:
			evaluation.Rejected = saturatingAdd(evaluation.Rejected, value.count)
			if normalized.CountRejectedAsFailure {
				evaluation.DecisionSamples = saturatingAdd(evaluation.DecisionSamples, value.count)
				decisionDurationMS = saturatingAdd(decisionDurationMS, value.durationMS)
			}
		}
	}

	if evaluation.DecisionSamples > 0 {
		if evaluation.Successes > ^uint64(0)/10000 {
			evaluation.SuccessRateBasisPoints = 10000
		} else {
			evaluation.SuccessRateBasisPoints = (evaluation.Successes * 10000) / evaluation.DecisionSamples
			if evaluation.SuccessRateBasisPoints > 10000 {
				evaluation.SuccessRateBasisPoints = 10000
			}
		}
		evaluation.AverageDurationMS = decisionDurationMS / evaluation.DecisionSamples
	}
	if evaluation.DecisionSamples < normalized.MinimumSamples {
		return evaluation, nil
	}
	evaluation.EnoughSamples = true

	unavailableThresholdBPS := normalized.UnavailableSuccessPercent * 100
	degradedThresholdBPS := normalized.DegradedSuccessPercent * 100
	if unavailableThresholdBPS > 0 && evaluation.SuccessRateBasisPoints < unavailableThresholdBPS {
		evaluation.RecommendedHealth = HealthUnavailable
		return evaluation, nil
	}
	if degradedThresholdBPS > 0 && evaluation.SuccessRateBasisPoints < degradedThresholdBPS {
		evaluation.RecommendedHealth = HealthDegraded
		return evaluation, nil
	}
	if normalized.MaxAverageDurationMS > 0 && evaluation.AverageDurationMS > normalized.MaxAverageDurationMS {
		evaluation.RecommendedHealth = HealthDegraded
		evaluation.LatencyDegraded = true
		return evaluation, nil
	}
	evaluation.RecommendedHealth = HealthHealthy
	return evaluation, nil
}

// ApplyHealthPolicy evaluates and, only when the minimum evidence threshold is
// met, applies the resulting health state. The evaluation is point-in-time; a
// concurrent Record may land immediately after it, which is acceptable for this
// Development control-plane primitive and must not be presented as durable SLO
// accounting.
func (c *Controller) ApplyHealthPolicy(providerID string, policy ProviderHealthPolicy) (ProviderHealthEvaluation, error) {
	evaluation, err := c.EvaluateHealth(providerID, policy)
	if err != nil || !evaluation.EnoughSamples {
		return evaluation, err
	}
	if err := c.SetHealth(providerID, evaluation.RecommendedHealth); err != nil {
		return ProviderHealthEvaluation{}, err
	}
	evaluation.CurrentHealth = evaluation.RecommendedHealth
	return evaluation, nil
}

func saturatingAdd(left, right uint64) uint64 {
	if ^uint64(0)-left < right {
		return ^uint64(0)
	}
	return left + right
}

// recordHealthSample is a test/helper convenience that preserves Controller's
// existing bounded observation semantics.
func recordHealthSample(c *Controller, providerID string, outcome Outcome, duration time.Duration, count int) error {
	for i := 0; i < count; i++ {
		if err := c.Record(providerID, outcome, duration); err != nil {
			return err
		}
	}
	return nil
}
