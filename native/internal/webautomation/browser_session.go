package webautomation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	DefaultBrowserSessionTimeout = 5 * time.Minute
	MaxBrowserSessionTimeout     = 10 * time.Minute
)

type BrowserSessionOperation string

const (
	BrowserSessionOperationStart     BrowserSessionOperation = "start"
	BrowserSessionOperationTerminate BrowserSessionOperation = "terminate"
)

var (
	ErrBrowserSessionNotAuthorized      = errors.New("browser session operation is not authorized")
	ErrBrowserSessionUnavailable        = errors.New("browser session is unavailable")
	ErrBrowserSessionCreationUnknown    = errors.New("browser session creation is not confirmed")
	ErrBrowserSessionTerminationUnknown = errors.New("browser session termination is not confirmed")
)

// BrowserSessionAuthorization is the local GoreeCloud authority record for a
// direct low-level browser session operation. Purpose is local authorization
// context and must never be forwarded to the provider solely because a browser
// session is being created or terminated.
type BrowserSessionAuthorization struct {
	Operation BrowserSessionOperation
	TargetURL string
	SessionID string
	Purpose   string
}

// BrowserSessionAuthorizer is the fail-closed authority seam for Privacy Shield,
// Wardveil, caller/session authority, target-site policy, and cost policy.
// This package intentionally ships no allow-all implementation.
type BrowserSessionAuthorizer interface {
	AuthorizeBrowserSession(context.Context, BrowserSessionAuthorization) error
}

type StartBrowserSessionRequest struct {
	TargetURL string
	Purpose   string
	Timeout   time.Duration
}

// BrowserSession contains short-lived provider control endpoints. SessionID,
// CDPURL, and BaseURL are sensitive runtime state and must not be emitted through
// normal logs, analytics, status endpoints, or long-lived configuration.
type BrowserSession struct {
	SessionID string
	CDPURL    string
	BaseURL   string
	TargetURL string
	Timeout   time.Duration
}

type TerminateBrowserSessionRequest struct {
	SessionID string
	Purpose   string
}

// BrowserSessionProvider is the provider-neutral lifecycle transport for direct
// low-level browser sessions. It does not expose arbitrary provider mutation or
// account-management operations.
type BrowserSessionProvider interface {
	StartBrowserSession(context.Context, StartBrowserSessionRequest) (BrowserSession, error)
	TerminateBrowserSession(context.Context, string) error
}

// BrowserSessionManager owns the authorized lifetime of direct browser sessions
// created through one runtime instance. It deliberately refuses to terminate a
// session it did not create, preventing a caller from presenting an arbitrary
// provider session identifier for destructive action.
type BrowserSessionManager struct {
	authorizer BrowserSessionAuthorizer
	provider   BrowserSessionProvider
	mu         sync.Mutex
	active     map[string]string
}

func NewBrowserSessionManager(authorizer BrowserSessionAuthorizer, provider BrowserSessionProvider) (*BrowserSessionManager, error) {
	if authorizer == nil {
		return nil, fmt.Errorf("%w: browser session authorizer is required", ErrBrowserSessionNotAuthorized)
	}
	if provider == nil {
		return nil, fmt.Errorf("%w: browser session provider is required", ErrBrowserSessionUnavailable)
	}
	return &BrowserSessionManager{authorizer: authorizer, provider: provider, active: make(map[string]string)}, nil
}

func (m *BrowserSessionManager) Start(ctx context.Context, input StartBrowserSessionRequest) (BrowserSession, error) {
	if m == nil || m.authorizer == nil || m.provider == nil {
		return BrowserSession{}, ErrBrowserSessionUnavailable
	}
	request, err := normalizeStartBrowserSessionRequest(input)
	if err != nil {
		return BrowserSession{}, err
	}
	if err := m.authorize(ctx, BrowserSessionAuthorization{
		Operation: BrowserSessionOperationStart,
		TargetURL: request.TargetURL,
		Purpose:   request.Purpose,
	}); err != nil {
		return BrowserSession{}, err
	}

	session, err := m.provider.StartBrowserSession(ctx, request)
	if err != nil {
		return BrowserSession{}, err
	}
	normalized, err := normalizeBrowserSessionResult(session, request)
	if err != nil {
		return BrowserSession{}, err
	}
	m.mu.Lock()
	m.active[normalized.SessionID] = request.TargetURL
	m.mu.Unlock()
	return normalized, nil
}

func (m *BrowserSessionManager) Terminate(ctx context.Context, input TerminateBrowserSessionRequest) error {
	if m == nil || m.authorizer == nil || m.provider == nil {
		return ErrBrowserSessionUnavailable
	}
	request, err := normalizeTerminateBrowserSessionRequest(input)
	if err != nil {
		return err
	}
	m.mu.Lock()
	target, exists := m.active[request.SessionID]
	m.mu.Unlock()
	if !exists {
		return ErrBrowserSessionUnavailable
	}
	if err := m.authorize(ctx, BrowserSessionAuthorization{
		Operation: BrowserSessionOperationTerminate,
		TargetURL: target,
		SessionID: request.SessionID,
		Purpose:   request.Purpose,
	}); err != nil {
		return err
	}
	if err := m.provider.TerminateBrowserSession(ctx, request.SessionID); err != nil {
		return err
	}
	m.mu.Lock()
	delete(m.active, request.SessionID)
	m.mu.Unlock()
	return nil
}

func (m *BrowserSessionManager) authorize(ctx context.Context, input BrowserSessionAuthorization) error {
	if err := m.authorizer.AuthorizeBrowserSession(ctx, input); err != nil {
		if errors.Is(err, ErrBrowserSessionNotAuthorized) {
			return err
		}
		return fmt.Errorf("%w: browser session policy rejected operation", ErrBrowserSessionNotAuthorized)
	}
	return nil
}

func normalizeStartBrowserSessionRequest(input StartBrowserSessionRequest) (StartBrowserSessionRequest, error) {
	target, err := normalizeTargetURL(input.TargetURL)
	if err != nil {
		return StartBrowserSessionRequest{}, err
	}
	purpose, err := normalizeBrowserSessionPurpose(input.Purpose)
	if err != nil {
		return StartBrowserSessionRequest{}, err
	}
	timeout := input.Timeout
	if timeout == 0 {
		timeout = DefaultBrowserSessionTimeout
	}
	if timeout < 5*time.Second || timeout > MaxBrowserSessionTimeout {
		return StartBrowserSessionRequest{}, fmt.Errorf("%w: browser session timeout must be between 5 seconds and %s", ErrInvalidRequest, MaxBrowserSessionTimeout)
	}
	return StartBrowserSessionRequest{TargetURL: target, Purpose: purpose, Timeout: timeout}, nil
}

func normalizeTerminateBrowserSessionRequest(input TerminateBrowserSessionRequest) (TerminateBrowserSessionRequest, error) {
	sessionID, err := normalizeProviderReference(input.SessionID, "browser session ID")
	if err != nil {
		return TerminateBrowserSessionRequest{}, err
	}
	purpose, err := normalizeBrowserSessionPurpose(input.Purpose)
	if err != nil {
		return TerminateBrowserSessionRequest{}, err
	}
	return TerminateBrowserSessionRequest{SessionID: sessionID, Purpose: purpose}, nil
}

func normalizeBrowserSessionResult(input BrowserSession, request StartBrowserSessionRequest) (BrowserSession, error) {
	sessionID, err := normalizeProviderReference(input.SessionID, "browser session ID")
	if err != nil {
		return BrowserSession{}, fmt.Errorf("%w: provider returned invalid browser session identity", ErrAutomationFailed)
	}
	cdpURL, err := validateSensitiveSetupURL(input.CDPURL, "wss")
	if err != nil {
		return BrowserSession{}, fmt.Errorf("%w: provider returned invalid browser CDP endpoint", ErrAutomationFailed)
	}
	baseURL, err := validateSensitiveSetupURL(input.BaseURL, "https")
	if err != nil {
		return BrowserSession{}, fmt.Errorf("%w: provider returned invalid browser base endpoint", ErrAutomationFailed)
	}
	return BrowserSession{
		SessionID: sessionID,
		CDPURL:    cdpURL,
		BaseURL:   baseURL,
		TargetURL: request.TargetURL,
		Timeout:   request.Timeout,
	}, nil
}

func normalizeBrowserSessionPurpose(raw string) (string, error) {
	purpose := strings.TrimSpace(raw)
	if purpose == "" || utf8.RuneCountInString(purpose) > MaxPurposeRunes || strings.ContainsRune(purpose, '\x00') {
		return "", fmt.Errorf("%w: browser session purpose is required and must be bounded", ErrInvalidRequest)
	}
	return purpose, nil
}
