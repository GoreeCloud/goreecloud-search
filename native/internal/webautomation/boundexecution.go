package webautomation

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// BoundAuthExecutor applies one managed exact-host authentication binding before
// invoking an existing web-automation executor. It prevents recurring
// authenticated runs from accidentally omitting the intended Browser Context
// Profile or from selecting broad/ad-hoc Vault credentials.
//
// Successful execution records session-reuse evidence. The wrapper deliberately
// does not infer Vault repair merely because scoped Vault references were made
// available to the provider; repair evidence requires a separate trusted signal.
type BoundAuthExecutor struct {
	bindings   *AuthBindings
	acceptance *AuthAcceptanceRegistry
	executor   Executor
	now        func() time.Time
}

func NewBoundAuthExecutor(bindings *AuthBindings, acceptance *AuthAcceptanceRegistry, executor Executor) (*BoundAuthExecutor, error) {
	if bindings == nil || len(bindings.byHost) == 0 {
		return nil, ErrAuthBindingUnavailable
	}
	if acceptance == nil {
		return nil, ErrAuthAcceptanceUnavailable
	}
	if executor == nil {
		return nil, fmt.Errorf("%w: authenticated executor is required", ErrCapabilityUnavailable)
	}
	return &BoundAuthExecutor{bindings: bindings, acceptance: acceptance, executor: executor, now: time.Now}, nil
}

func (e *BoundAuthExecutor) Run(ctx context.Context, input Request) (Result, error) {
	if e == nil || e.bindings == nil || e.acceptance == nil || e.executor == nil {
		return Result{}, ErrAuthAcceptanceUnavailable
	}
	request, err := e.bindings.Apply(input)
	if err != nil {
		return Result{}, err
	}

	result, runErr := e.executor.Run(ctx, request)
	state, shouldRecord := boundAuthAcceptanceState(runErr)
	if !shouldRecord {
		return result, runErr
	}

	host, err := boundAuthTargetHost(request.URL)
	if err != nil {
		if runErr != nil {
			return Result{}, runErr
		}
		return Result{}, err
	}
	_, recordErr := e.acceptance.Record(AuthAcceptanceObservation{
		Host:       host,
		ProfileID:  request.ProfileID,
		State:      state,
		ObservedAt: e.now().UTC(),
	})
	if recordErr != nil {
		if runErr != nil {
			return Result{}, runErr
		}
		return Result{}, fmt.Errorf("%w: authenticated execution evidence could not be recorded", ErrAuthAcceptanceUnavailable)
	}
	return result, runErr
}

func boundAuthAcceptanceState(err error) (AuthAcceptanceState, bool) {
	if err == nil {
		return AuthAcceptanceSessionReuseVerified, true
	}
	if errors.Is(err, ErrSiteBlocked) {
		return AuthAcceptanceSiteBlocked, true
	}
	if errors.Is(err, ErrGoalFailed) {
		return AuthAcceptanceFailed, true
	}
	// Caller cancellation/deadline and provider/infrastructure failures are not
	// authentication evidence. Recording them as profile failures would create
	// misleading acceptance state.
	return "", false
}

func boundAuthTargetHost(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Hostname() == "" {
		return "", fmt.Errorf("%w: authenticated target URL is invalid", ErrInvalidRequest)
	}
	host := canonicalBindingHost(parsed.Hostname())
	if !validBindingHost(host) {
		return "", fmt.Errorf("%w: authenticated target host is invalid", ErrInvalidRequest)
	}
	return host, nil
}
