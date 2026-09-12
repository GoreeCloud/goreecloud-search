package webautomation

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxGoalRunes             = 4096
	MaxPurposeRunes          = 512
	MaxProfileIDRunes        = 256
	MaxCredentialItemIDRunes = 256
	MaxCredentialItems       = 16
	MaxAgentSteps            = 50
	DefaultMaxDuration       = 3 * time.Minute
	MaxMaxDuration           = 10 * time.Minute
	maxTargetURLBytes        = 8192
)

type Capability string

const (
	CapabilityAgent   Capability = "agent"
	CapabilityBrowser Capability = "browser"
)

type BrowserProfile string

const (
	BrowserProfileLite    BrowserProfile = "lite"
	BrowserProfileStealth BrowserProfile = "stealth"
)

var (
	ErrInvalidRequest        = errors.New("web automation request is invalid")
	ErrNotAuthorized         = errors.New("web automation is not authorized")
	ErrCapabilityUnavailable = errors.New("web automation capability is unavailable")
	ErrAutomationFailed      = errors.New("web automation failed")
)

// Request is GoreeCloud's provider-neutral web-automation request. Purpose is a
// local authorization/audit binding and must not be forwarded to an external
// provider unless a separately reviewed provider contract explicitly requires it.
type Request struct {
	URL               string
	Goal              string
	Purpose           string
	Capability        Capability
	BrowserProfile    BrowserProfile
	UseProfile        bool
	ProfileID         string
	UseVault          bool
	CredentialItemIDs []string
	MaxSteps          int
	MaxDuration       time.Duration
}

// Result intentionally exposes only normalized caller-facing output. Provider
// traces, screenshots, HTML, internal evidence, raw errors, and credential data
// are outside this stable contract.
type Result struct {
	Capability Capability `json:"capability"`
	Output     string     `json:"output,omitempty"`
}

// Authorizer is the fail-closed policy seam for Privacy Shield, Wardveil, caller
// authority, purpose, cost, session-state, and capability checks. This package
// deliberately ships no allow-all authorizer.
type Authorizer interface {
	Authorize(context.Context, Request) error
}

type Executor interface {
	Run(context.Context, Request) (Result, error)
}

type Escalator struct {
	authorizer Authorizer
	agent      Executor
	browser    Executor
}

func NewEscalator(authorizer Authorizer, agent Executor, browser Executor) (*Escalator, error) {
	if authorizer == nil {
		return nil, fmt.Errorf("%w: authorizer is required", ErrNotAuthorized)
	}
	return &Escalator{authorizer: authorizer, agent: agent, browser: browser}, nil
}

func (e *Escalator) Run(ctx context.Context, input Request) (Result, error) {
	request, err := normalizeRequest(input)
	if err != nil {
		return Result{}, err
	}
	if err := e.authorizer.Authorize(ctx, request); err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return Result{}, err
		}
		return Result{}, fmt.Errorf("%w: policy rejected request", ErrNotAuthorized)
	}

	var executor Executor
	switch request.Capability {
	case CapabilityAgent:
		executor = e.agent
	case CapabilityBrowser:
		executor = e.browser
	}
	if executor == nil {
		return Result{}, fmt.Errorf("%w: %s", ErrCapabilityUnavailable, request.Capability)
	}
	return executor.Run(ctx, request)
}

func normalizeRequest(input Request) (Request, error) {
	target, err := normalizeTargetURL(input.URL)
	if err != nil {
		return Request{}, err
	}
	goal := strings.TrimSpace(input.Goal)
	purpose := strings.TrimSpace(input.Purpose)
	if goal == "" || utf8.RuneCountInString(goal) > MaxGoalRunes {
		return Request{}, fmt.Errorf("%w: goal is required and must be at most %d characters", ErrInvalidRequest, MaxGoalRunes)
	}
	if purpose == "" || utf8.RuneCountInString(purpose) > MaxPurposeRunes {
		return Request{}, fmt.Errorf("%w: purpose is required and must be at most %d characters", ErrInvalidRequest, MaxPurposeRunes)
	}

	capability := input.Capability
	if capability == "" {
		capability = CapabilityAgent
	}
	if capability != CapabilityAgent && capability != CapabilityBrowser {
		return Request{}, fmt.Errorf("%w: unsupported capability", ErrInvalidRequest)
	}

	profile := input.BrowserProfile
	if profile == "" {
		profile = BrowserProfileLite
	}
	if profile != BrowserProfileLite && profile != BrowserProfileStealth {
		return Request{}, fmt.Errorf("%w: unsupported browser profile", ErrInvalidRequest)
	}

	profileID := strings.TrimSpace(input.ProfileID)
	if utf8.RuneCountInString(profileID) > MaxProfileIDRunes || strings.ContainsAny(profileID, "\r\n\x00") {
		return Request{}, fmt.Errorf("%w: profile ID is invalid", ErrInvalidRequest)
	}
	if input.UseProfile && profileID == "" {
		return Request{}, fmt.Errorf("%w: explicit profile ID is required when profile reuse is enabled", ErrInvalidRequest)
	}
	if !input.UseProfile && profileID != "" {
		return Request{}, fmt.Errorf("%w: profile ID requires profile reuse", ErrInvalidRequest)
	}

	if len(input.CredentialItemIDs) > MaxCredentialItems {
		return Request{}, fmt.Errorf("%w: too many credential item IDs", ErrInvalidRequest)
	}
	credentialIDs := make([]string, 0, len(input.CredentialItemIDs))
	seenCredentials := map[string]bool{}
	for _, raw := range input.CredentialItemIDs {
		value := strings.TrimSpace(raw)
		if value == "" || utf8.RuneCountInString(value) > MaxCredentialItemIDRunes || strings.ContainsAny(value, "\r\n\x00") {
			return Request{}, fmt.Errorf("%w: credential item ID is invalid", ErrInvalidRequest)
		}
		if !seenCredentials[value] {
			seenCredentials[value] = true
			credentialIDs = append(credentialIDs, value)
		}
	}
	if len(credentialIDs) > 0 && !input.UseVault {
		return Request{}, fmt.Errorf("%w: credential item IDs require Vault", ErrInvalidRequest)
	}

	if input.MaxSteps < 0 || input.MaxSteps > MaxAgentSteps {
		return Request{}, fmt.Errorf("%w: max steps must be between 0 and %d", ErrInvalidRequest, MaxAgentSteps)
	}
	maxDuration := input.MaxDuration
	if maxDuration == 0 {
		maxDuration = DefaultMaxDuration
	}
	if maxDuration < time.Second || maxDuration > MaxMaxDuration {
		return Request{}, fmt.Errorf("%w: max duration must be between 1 second and %s", ErrInvalidRequest, MaxMaxDuration)
	}

	return Request{
		URL:               target,
		Goal:              goal,
		Purpose:           purpose,
		Capability:        capability,
		BrowserProfile:    profile,
		UseProfile:        input.UseProfile,
		ProfileID:         profileID,
		UseVault:          input.UseVault,
		CredentialItemIDs: credentialIDs,
		MaxSteps:          input.MaxSteps,
		MaxDuration:       maxDuration,
	}, nil
}

func normalizeTargetURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" || len(value) > maxTargetURLBytes {
		return "", fmt.Errorf("%w: target URL is invalid", ErrInvalidRequest)
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("%w: target URL must be HTTP or HTTPS without embedded credentials", ErrInvalidRequest)
	}
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || !strings.Contains(host, ".") {
		return "", fmt.Errorf("%w: target URL host is not allowed", ErrInvalidRequest)
	}
	if ip, parseErr := netip.ParseAddr(host); parseErr == nil && !publicAddress(ip) {
		return "", fmt.Errorf("%w: target URL address is not public", ErrInvalidRequest)
	}
	parsed.Fragment = ""
	return parsed.String(), nil
}

func publicAddress(address netip.Addr) bool {
	address = address.Unmap()
	return address.IsValid() && !address.IsUnspecified() && !address.IsLoopback() && !address.IsPrivate() &&
		!address.IsLinkLocalUnicast() && !address.IsLinkLocalMulticast() && !address.IsMulticast()
}
