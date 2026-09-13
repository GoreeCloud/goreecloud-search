package webautomation

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	DefaultBrowserSessionCleanupTimeout = 15 * time.Second
	MaxBrowserSessionCleanupTimeout     = 30 * time.Second
)

// BrowserSessionWork receives sensitive ephemeral browser-control state for one
// already-authorized direct Browser session. Implementations must not log,
// persist, or expose SessionID, CDPURL, or BaseURL through ordinary status or
// analytics surfaces.
type BrowserSessionWork func(context.Context, BrowserSession) error

type BrowserSessionLeaseRequest struct {
	Start          StartBrowserSessionRequest
	CleanupPurpose string
}

// BrowserSessionLeaseRunner scopes one direct Browser session to one unit of
// caller work and always attempts termination after a confirmed create. Cleanup
// deliberately uses a fresh bounded context so caller cancellation cannot turn
// into a silent remote-session leak.
type BrowserSessionLeaseRunner struct {
	manager        *BrowserSessionManager
	cleanupTimeout time.Duration
}

func NewBrowserSessionLeaseRunner(manager *BrowserSessionManager, cleanupTimeout time.Duration) (*BrowserSessionLeaseRunner, error) {
	if manager == nil || manager.authorizer == nil || manager.provider == nil {
		return nil, ErrBrowserSessionUnavailable
	}
	if cleanupTimeout == 0 {
		cleanupTimeout = DefaultBrowserSessionCleanupTimeout
	}
	if cleanupTimeout < time.Second || cleanupTimeout > MaxBrowserSessionCleanupTimeout {
		return nil, fmt.Errorf("%w: browser session cleanup timeout must be between 1 second and %s", ErrInvalidRequest, MaxBrowserSessionCleanupTimeout)
	}
	return &BrowserSessionLeaseRunner{manager: manager, cleanupTimeout: cleanupTimeout}, nil
}

func (r *BrowserSessionLeaseRunner) Run(ctx context.Context, input BrowserSessionLeaseRequest, work BrowserSessionWork) (resultErr error) {
	if r == nil || r.manager == nil || r.cleanupTimeout <= 0 {
		return ErrBrowserSessionUnavailable
	}
	if work == nil {
		return fmt.Errorf("%w: browser session work is required", ErrInvalidRequest)
	}
	cleanupPurpose, err := normalizeBrowserSessionPurpose(input.CleanupPurpose)
	if err != nil {
		return err
	}

	session, err := r.manager.Start(ctx, input.Start)
	if err != nil {
		return err
	}

	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), r.cleanupTimeout)
		defer cancel()
		cleanupErr := r.manager.Terminate(cleanupCtx, TerminateBrowserSessionRequest{
			SessionID: session.SessionID,
			Purpose:   cleanupPurpose,
		})
		if cleanupErr == nil {
			return
		}
		wrapped := fmt.Errorf("browser session cleanup failed: %w", cleanupErr)
		if resultErr == nil {
			resultErr = wrapped
			return
		}
		resultErr = errors.Join(resultErr, wrapped)
	}()

	return work(ctx, session)
}
