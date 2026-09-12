package webintelligence

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrExecutorContract    = errors.New("web intelligence executor contract is invalid")
	ErrExecutorUnavailable = errors.New("web intelligence executor is unavailable")
	ErrExecutionFailed     = errors.New("web intelligence execution failed")
)

// Executor is a provider-specific execution boundary registered under one exact
// GoreeCloud web-intelligence capability. Implementations own their provider
// transport, response validation, credential handling, and content boundaries.
// The coordinator never interprets or records Input or Value.
type Executor interface {
	ProviderID() string
	Capability() Capability
	Execute(context.Context, any) (ExecutionResult, error)
}

// ExecutionResult contains only the opaque provider result and actual metered
// cost reported by the executor. CostMicros must be zero for providers declared
// free in the Controller configuration.
type ExecutionResult struct {
	Value      any
	CostMicros uint64
}

// Result is returned only after a provider succeeds and budget accounting has
// been finalized. CostMicros is the total metered cost incurred across all
// attempted providers for this operation, including failed metered attempts.
type Result struct {
	ProviderID string
	Value      any
	CostMicros uint64
	Attempts   int
}

// Coordinator executes a Controller plan without changing its capability. It
// can fall back only through the same-capability providers already selected by
// Controller.Plan. Cross-capability escalation remains the caller's separate
// authorization decision.
type Coordinator struct {
	controller *Controller
	executors  map[string]Executor
}

func NewCoordinator(controller *Controller, executors []Executor) (*Coordinator, error) {
	if controller == nil {
		return nil, fmt.Errorf("%w: controller is required", ErrExecutorContract)
	}

	configured := make(map[string]Capability)
	for _, provider := range controller.Snapshot().Providers {
		configured[provider.ID] = provider.Capability
	}

	registry := make(map[string]Executor, len(executors))
	for index, executor := range executors {
		if executor == nil {
			return nil, fmt.Errorf("%w: executor %d is nil", ErrExecutorContract, index+1)
		}
		providerID := strings.TrimSpace(executor.ProviderID())
		if providerID == "" {
			return nil, fmt.Errorf("%w: executor %d has no provider ID", ErrExecutorContract, index+1)
		}
		if _, exists := registry[providerID]; exists {
			return nil, fmt.Errorf("%w: duplicate executor provider ID", ErrExecutorContract)
		}
		configuredCapability, exists := configured[providerID]
		if !exists {
			return nil, fmt.Errorf("%w: executor provider is not configured", ErrExecutorContract)
		}
		if executor.Capability() != configuredCapability {
			return nil, fmt.Errorf("%w: executor capability does not match configured provider", ErrExecutorContract)
		}
		registry[providerID] = executor
	}

	return &Coordinator{controller: controller, executors: registry}, nil
}

// Execute obtains one exact-capability plan and runs its provider sequence. The
// input and provider result remain opaque to this package and are never copied
// into Controller observations.
func (c *Coordinator) Execute(ctx context.Context, request Request, input any) (Result, error) {
	if c == nil || c.controller == nil {
		return Result{}, fmt.Errorf("%w: coordinator is not initialized", ErrExecutorContract)
	}
	if ctx == nil {
		return Result{}, fmt.Errorf("%w: context is required", ErrInvalidRequest)
	}

	plan, err := c.controller.Plan(request)
	if err != nil {
		return Result{}, err
	}

	var (
		spentMicros     uint64
		meteredAttempt  bool
		executorAttempts int
	)

	finalize := func() error {
		if plan.ReservationID == 0 {
			return c.controller.Commit(0, spentMicros)
		}
		if meteredAttempt || spentMicros != 0 {
			return c.controller.Commit(plan.ReservationID, spentMicros)
		}
		return c.controller.Release(plan.ReservationID)
	}

	for _, provider := range plan.Providers {
		if err := ctx.Err(); err != nil {
			if finalizeErr := finalize(); finalizeErr != nil {
				return Result{}, finalizeErr
			}
			return Result{}, err
		}

		executor, ok := c.executors[provider.ID]
		if !ok {
			if recordErr := c.controller.Record(provider.ID, OutcomeRejected, 0); recordErr != nil {
				_ = finalize()
				return Result{}, ErrInvalidOperation
			}
			continue
		}
		if executor.Capability() != plan.Capability || executor.Capability() != provider.Capability {
			if recordErr := c.controller.Record(provider.ID, OutcomeRejected, 0); recordErr != nil {
				_ = finalize()
				return Result{}, ErrInvalidOperation
			}
			if finalizeErr := finalize(); finalizeErr != nil {
				return Result{}, finalizeErr
			}
			return Result{}, ErrExecutorContract
		}

		executorAttempts++
		started := time.Now()
		attemptResult, executeErr := executor.Execute(ctx, input)
		duration := time.Since(started)

		if provider.Metered {
			meteredAttempt = true
			var overflow bool
			spentMicros, overflow = addMicros(spentMicros, attemptResult.CostMicros)
			if overflow {
				_ = c.controller.Record(provider.ID, OutcomeRejected, duration)
				if finalizeErr := finalize(); finalizeErr != nil {
					return Result{}, finalizeErr
				}
				return Result{}, ErrBudgetExceeded
			}
		} else if attemptResult.CostMicros != 0 {
			if recordErr := c.controller.Record(provider.ID, OutcomeRejected, duration); recordErr != nil {
				_ = finalize()
				return Result{}, ErrInvalidOperation
			}
			if finalizeErr := finalize(); finalizeErr != nil {
				return Result{}, finalizeErr
			}
			return Result{}, ErrExecutorContract
		}

		outcome := executionOutcome(ctx, executeErr)
		if recordErr := c.controller.Record(provider.ID, outcome, duration); recordErr != nil {
			_ = finalize()
			return Result{}, ErrInvalidOperation
		}

		if executeErr == nil {
			if finalizeErr := finalize(); finalizeErr != nil {
				return Result{}, finalizeErr
			}
			return Result{
				ProviderID: provider.ID,
				Value:      attemptResult.Value,
				CostMicros: spentMicros,
				Attempts:   executorAttempts,
			}, nil
		}

		if outcome == OutcomeCancelled {
			if finalizeErr := finalize(); finalizeErr != nil {
				return Result{}, finalizeErr
			}
			return Result{}, context.Canceled
		}
		if outcome == OutcomeTimeout {
			if finalizeErr := finalize(); finalizeErr != nil {
				return Result{}, finalizeErr
			}
			return Result{}, context.DeadlineExceeded
		}
	}

	if finalizeErr := finalize(); finalizeErr != nil {
		return Result{}, finalizeErr
	}
	if executorAttempts == 0 {
		return Result{}, ErrExecutorUnavailable
	}
	return Result{}, ErrExecutionFailed
}

func executionOutcome(ctx context.Context, err error) Outcome {
	if err == nil {
		return OutcomeSuccess
	}
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		return OutcomeCancelled
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return OutcomeTimeout
	}
	return OutcomeFailure
}

func addMicros(current, next uint64) (uint64, bool) {
	if ^uint64(0)-current < next {
		return ^uint64(0), true
	}
	return current + next, false
}
