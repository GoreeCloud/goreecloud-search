package webautomation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxProfileNameRunes      = 100
	MaxProfilePurposeRunes   = 512
	DefaultSetupTimeout      = 15 * time.Minute
	MaxProfileSetupTimeout   = 15 * time.Minute
)

type ProfileOperation string

const (
	ProfileOperationCreate      ProfileOperation = "create"
	ProfileOperationStartSetup  ProfileOperation = "start_setup"
	ProfileOperationSaveSetup   ProfileOperation = "save_setup"
	ProfileOperationCancelSetup ProfileOperation = "cancel_setup"
)

var (
	ErrProfileNotAuthorized = errors.New("browser profile lifecycle operation is not authorized")
	ErrProfileUnavailable   = errors.New("browser profile is unavailable")
	ErrProfileConflict      = errors.New("browser profile lifecycle conflict")
)

// ProfileAuthorization is the local GoreeCloud authority record for a profile
// lifecycle operation. Purpose remains local and must not be forwarded to an
// external provider merely because a provider transport is used.
type ProfileAuthorization struct {
	Operation ProfileOperation
	ProfileID string
	TargetURL string
	Purpose   string
}

// ProfileAuthorizer is the fail-closed authority seam for Privacy Shield,
// Wardveil, caller/session authority, intended site, and operational policy.
// No allow-all implementation is shipped from this package.
type ProfileAuthorizer interface {
	AuthorizeProfile(context.Context, ProfileAuthorization) error
}

type CreateProfileRequest struct {
	Name             string
	Purpose          string
	SetAsDefault     bool
	ProxyCountryCode string
}

type CreateProfileResult struct {
	ProfileID        string
	Name             string
	ProxyCountryCode string
}

type StartProfileSetupRequest struct {
	ProfileID string
	TargetURL string
	Purpose   string
	Timeout   time.Duration
}

// ProfileSetupSession contains short-lived browser-control endpoints and must be
// treated as sensitive runtime state. It must not be persisted in normal logs,
// status endpoints, analytics, or long-lived configuration.
type ProfileSetupSession struct {
	ProfileID      string
	SessionID      string
	CDPURL         string
	BaseURL        string
	Timeout        time.Duration
	ExpiresAt      time.Time
}

type SaveProfileSetupRequest struct {
	ProfileID string
	SessionID string
	Purpose   string
}

type SaveProfileSetupResult struct {
	DomainsUpdated []string
	DomainsFailed  []string
	CookieCount    int
	PagesCaptured  int
}

type CancelProfileSetupRequest struct {
	ProfileID string
	SessionID string
	Purpose   string
}

// ProfileProvider is the provider-neutral lifecycle transport. It intentionally
// excludes raw cookie upload, password material, and arbitrary provider mutation
// surfaces from the stable GoreeCloud contract.
type ProfileProvider interface {
	CreateProfile(context.Context, CreateProfileRequest) (CreateProfileResult, error)
	StartProfileSetup(context.Context, StartProfileSetupRequest) (ProfileSetupSession, error)
	SaveProfileSetup(context.Context, SaveProfileSetupRequest) (SaveProfileSetupResult, error)
	CancelProfileSetup(context.Context, CancelProfileSetupRequest) error
}

type ProfileLifecycle struct {
	authorizer ProfileAuthorizer
	provider   ProfileProvider
}

func NewProfileLifecycle(authorizer ProfileAuthorizer, provider ProfileProvider) (*ProfileLifecycle, error) {
	if authorizer == nil {
		return nil, fmt.Errorf("%w: profile authorizer is required", ErrProfileNotAuthorized)
	}
	if provider == nil {
		return nil, fmt.Errorf("%w: profile provider is required", ErrProfileUnavailable)
	}
	return &ProfileLifecycle{authorizer: authorizer, provider: provider}, nil
}

func (l *ProfileLifecycle) Create(ctx context.Context, input CreateProfileRequest) (CreateProfileResult, error) {
	request, err := normalizeCreateProfileRequest(input)
	if err != nil {
		return CreateProfileResult{}, err
	}
	if err := l.authorize(ctx, ProfileAuthorization{Operation: ProfileOperationCreate, Purpose: request.Purpose}); err != nil {
		return CreateProfileResult{}, err
	}
	return l.provider.CreateProfile(ctx, request)
}

func (l *ProfileLifecycle) StartSetup(ctx context.Context, input StartProfileSetupRequest) (ProfileSetupSession, error) {
	request, err := normalizeStartProfileSetupRequest(input)
	if err != nil {
		return ProfileSetupSession{}, err
	}
	if err := l.authorize(ctx, ProfileAuthorization{Operation: ProfileOperationStartSetup, ProfileID: request.ProfileID, TargetURL: request.TargetURL, Purpose: request.Purpose}); err != nil {
		return ProfileSetupSession{}, err
	}
	return l.provider.StartProfileSetup(ctx, request)
}

func (l *ProfileLifecycle) SaveSetup(ctx context.Context, input SaveProfileSetupRequest) (SaveProfileSetupResult, error) {
	request, err := normalizeSaveProfileSetupRequest(input)
	if err != nil {
		return SaveProfileSetupResult{}, err
	}
	if err := l.authorize(ctx, ProfileAuthorization{Operation: ProfileOperationSaveSetup, ProfileID: request.ProfileID, Purpose: request.Purpose}); err != nil {
		return SaveProfileSetupResult{}, err
	}
	return l.provider.SaveProfileSetup(ctx, request)
}

func (l *ProfileLifecycle) CancelSetup(ctx context.Context, input CancelProfileSetupRequest) error {
	request, err := normalizeCancelProfileSetupRequest(input)
	if err != nil {
		return err
	}
	if err := l.authorize(ctx, ProfileAuthorization{Operation: ProfileOperationCancelSetup, ProfileID: request.ProfileID, Purpose: request.Purpose}); err != nil {
		return err
	}
	return l.provider.CancelProfileSetup(ctx, request)
}

func (l *ProfileLifecycle) authorize(ctx context.Context, authorization ProfileAuthorization) error {
	if err := l.authorizer.AuthorizeProfile(ctx, authorization); err != nil {
		if errors.Is(err, ErrProfileNotAuthorized) {
			return err
		}
		return fmt.Errorf("%w: profile policy rejected operation", ErrProfileNotAuthorized)
	}
	return nil
}

func normalizeCreateProfileRequest(input CreateProfileRequest) (CreateProfileRequest, error) {
	name := strings.TrimSpace(input.Name)
	purpose, err := normalizeProfilePurpose(input.Purpose)
	if err != nil {
		return CreateProfileRequest{}, err
	}
	if name == "" || utf8.RuneCountInString(name) > MaxProfileNameRunes || strings.ContainsAny(name, "\r\n\x00") {
		return CreateProfileRequest{}, fmt.Errorf("%w: profile name is invalid", ErrInvalidRequest)
	}
	country := strings.ToUpper(strings.TrimSpace(input.ProxyCountryCode))
	if country != "" && !validProfileProxyCountry(country) {
		return CreateProfileRequest{}, fmt.Errorf("%w: profile proxy country is unsupported", ErrInvalidRequest)
	}
	return CreateProfileRequest{Name: name, Purpose: purpose, SetAsDefault: input.SetAsDefault, ProxyCountryCode: country}, nil
}

func normalizeStartProfileSetupRequest(input StartProfileSetupRequest) (StartProfileSetupRequest, error) {
	profileID, err := normalizeProviderReference(input.ProfileID, "profile ID")
	if err != nil {
		return StartProfileSetupRequest{}, err
	}
	target, err := normalizeTargetURL(input.TargetURL)
	if err != nil {
		return StartProfileSetupRequest{}, err
	}
	purpose, err := normalizeProfilePurpose(input.Purpose)
	if err != nil {
		return StartProfileSetupRequest{}, err
	}
	timeout := input.Timeout
	if timeout == 0 {
		timeout = DefaultSetupTimeout
	}
	if timeout < time.Second || timeout > MaxProfileSetupTimeout {
		return StartProfileSetupRequest{}, fmt.Errorf("%w: profile setup timeout must be between 1 second and %s", ErrInvalidRequest, MaxProfileSetupTimeout)
	}
	return StartProfileSetupRequest{ProfileID: profileID, TargetURL: target, Purpose: purpose, Timeout: timeout}, nil
}

func normalizeSaveProfileSetupRequest(input SaveProfileSetupRequest) (SaveProfileSetupRequest, error) {
	profileID, err := normalizeProviderReference(input.ProfileID, "profile ID")
	if err != nil {
		return SaveProfileSetupRequest{}, err
	}
	sessionID, err := normalizeProviderReference(input.SessionID, "setup session ID")
	if err != nil {
		return SaveProfileSetupRequest{}, err
	}
	purpose, err := normalizeProfilePurpose(input.Purpose)
	if err != nil {
		return SaveProfileSetupRequest{}, err
	}
	return SaveProfileSetupRequest{ProfileID: profileID, SessionID: sessionID, Purpose: purpose}, nil
}

func normalizeCancelProfileSetupRequest(input CancelProfileSetupRequest) (CancelProfileSetupRequest, error) {
	normalized, err := normalizeSaveProfileSetupRequest(SaveProfileSetupRequest{ProfileID: input.ProfileID, SessionID: input.SessionID, Purpose: input.Purpose})
	if err != nil {
		return CancelProfileSetupRequest{}, err
	}
	return CancelProfileSetupRequest{ProfileID: normalized.ProfileID, SessionID: normalized.SessionID, Purpose: normalized.Purpose}, nil
}

func normalizeProfilePurpose(raw string) (string, error) {
	purpose := strings.TrimSpace(raw)
	if purpose == "" || utf8.RuneCountInString(purpose) > MaxProfilePurposeRunes || strings.ContainsAny(purpose, "\x00") {
		return "", fmt.Errorf("%w: profile purpose is required and must be bounded", ErrInvalidRequest)
	}
	return purpose, nil
}

func normalizeProviderReference(raw, label string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" || utf8.RuneCountInString(value) > MaxProfileIDRunes {
		return "", fmt.Errorf("%w: %s is invalid", ErrInvalidRequest, label)
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' {
			continue
		}
		return "", fmt.Errorf("%w: %s is invalid", ErrInvalidRequest, label)
	}
	return value, nil
}

func validProfileProxyCountry(country string) bool {
	switch country {
	case "US", "GB", "CA", "DE", "FR", "JP", "AU":
		return true
	default:
		return false
	}
}
