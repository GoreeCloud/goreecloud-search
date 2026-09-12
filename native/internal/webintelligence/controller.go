package webintelligence

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	MaxProviderIDRunes = 128
	MaxPlanAttempts    = 8
)

type Capability string

const (
	CapabilitySearch   Capability = "search"
	CapabilityFetch    Capability = "fetch"
	CapabilityResearch Capability = "research"
	CapabilityAgent    Capability = "agent"
	CapabilityBrowser  Capability = "browser"
)

type Health string

const (
	HealthHealthy     Health = "healthy"
	HealthDegraded    Health = "degraded"
	HealthUnavailable Health = "unavailable"
)

type Outcome string

const (
	OutcomeSuccess   Outcome = "success"
	OutcomeFailure   Outcome = "failure"
	OutcomeTimeout   Outcome = "timeout"
	OutcomeCancelled Outcome = "cancelled"
	OutcomeRejected  Outcome = "rejected"
)

var (
	ErrInvalidRequest   = errors.New("web intelligence routing request is invalid")
	ErrNoProvider       = errors.New("no eligible web intelligence provider is available")
	ErrBudgetRequired   = errors.New("metered web intelligence requires a configured budget")
	ErrBudgetExceeded   = errors.New("web intelligence budget would be exceeded")
	ErrUnknownProvider  = errors.New("web intelligence provider is unknown")
	ErrInvalidOperation = errors.New("web intelligence operation is invalid")
)

type Provider struct {
	ID         string     `json:"id"`
	Capability Capability `json:"capability"`
	Metered    bool       `json:"metered"`
	Priority   int        `json:"priority"`
}

type ProviderSnapshot struct {
	ID         string     `json:"id"`
	Capability Capability `json:"capability"`
	Metered    bool       `json:"metered"`
	Priority   int        `json:"priority"`
	Health     Health     `json:"health"`
}

type BudgetPolicy struct {
	LimitMicros        uint64 `json:"limit_micros"`
	PerOperationMicros uint64 `json:"per_operation_micros"`
}

type BudgetSnapshot struct {
	LimitMicros        uint64 `json:"limit_micros"`
	PerOperationMicros uint64 `json:"per_operation_micros"`
	ReservedMicros     uint64 `json:"reserved_micros"`
	SpentMicros        uint64 `json:"spent_micros"`
	AvailableMicros    uint64 `json:"available_micros"`
}

type Request struct {
	Capability          Capability `json:"capability"`
	AllowMetered        bool       `json:"allow_metered"`
	EstimatedCostMicros uint64     `json:"estimated_cost_micros"`
	MaxAttempts         int        `json:"max_attempts"`
}

type Plan struct {
	Capability         Capability `json:"capability"`
	Providers          []Provider `json:"providers"`
	ReservationID      uint64     `json:"reservation_id,omitempty"`
	ReservedCostMicros uint64     `json:"reserved_cost_micros"`
}

type Observation struct {
	ProviderID string  `json:"provider_id"`
	Outcome    Outcome `json:"outcome"`
	Count      uint64  `json:"count"`
	DurationMS uint64  `json:"duration_ms"`
}

type Snapshot struct {
	Budget       BudgetSnapshot     `json:"budget"`
	Providers    []ProviderSnapshot `json:"providers"`
	Observations []Observation      `json:"observations"`
}

type providerState struct {
	provider Provider
	health   Health
}

type reservation struct {
	amount uint64
}

type observationKey struct {
	providerID string
	outcome    Outcome
}

type observationValue struct {
	count      uint64
	durationMS uint64
}

// Controller is a provider-neutral GoreeCloud web-intelligence policy primitive.
// It selects only providers that implement the exact requested capability. It
// never silently escalates Search to Fetch, Fetch to Research, Research to Agent,
// or Agent to Browser. Cross-capability escalation must be separately authorized
// by the caller before Plan is invoked.
type Controller struct {
	mu            sync.RWMutex
	providers     map[string]*providerState
	budgetPolicy  BudgetPolicy
	reserved      uint64
	spent         uint64
	reservations  map[uint64]reservation
	nextReserveID uint64
	observations  map[observationKey]observationValue
}

func NewController(providers []Provider, budget BudgetPolicy) (*Controller, error) {
	if budget.PerOperationMicros > 0 && budget.LimitMicros > 0 && budget.PerOperationMicros > budget.LimitMicros {
		return nil, fmt.Errorf("%w: per-operation budget exceeds total budget", ErrInvalidOperation)
	}
	controller := &Controller{
		providers:    make(map[string]*providerState, len(providers)),
		budgetPolicy: budget,
		reservations: make(map[uint64]reservation),
		observations: make(map[observationKey]observationValue),
	}
	for _, raw := range providers {
		provider, err := normalizeProvider(raw)
		if err != nil {
			return nil, err
		}
		if _, exists := controller.providers[provider.ID]; exists {
			return nil, fmt.Errorf("%w: duplicate provider ID %q", ErrInvalidOperation, provider.ID)
		}
		controller.providers[provider.ID] = &providerState{provider: provider, health: HealthHealthy}
	}
	return controller, nil
}

func (c *Controller) Plan(input Request) (Plan, error) {
	request, err := normalizeRequest(input)
	if err != nil {
		return Plan{}, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	candidates := make([]ProviderSnapshot, 0, len(c.providers))
	for _, state := range c.providers {
		if state.provider.Capability != request.Capability || state.health == HealthUnavailable {
			continue
		}
		if state.provider.Metered && !request.AllowMetered {
			continue
		}
		candidates = append(candidates, ProviderSnapshot{
			ID:         state.provider.ID,
			Capability: state.provider.Capability,
			Metered:    state.provider.Metered,
			Priority:   state.provider.Priority,
			Health:     state.health,
		})
	}
	if len(candidates) == 0 {
		return Plan{}, ErrNoProvider
	}

	sort.Slice(candidates, func(i, j int) bool {
		// Free providers are preferred over metered providers at the same
		// capability, then healthy over degraded, then configured priority.
		if candidates[i].Metered != candidates[j].Metered {
			return !candidates[i].Metered
		}
		if healthRank(candidates[i].Health) != healthRank(candidates[j].Health) {
			return healthRank(candidates[i].Health) < healthRank(candidates[j].Health)
		}
		if candidates[i].Priority != candidates[j].Priority {
			return candidates[i].Priority < candidates[j].Priority
		}
		return candidates[i].ID < candidates[j].ID
	})

	attempts := request.MaxAttempts
	if attempts > len(candidates) {
		attempts = len(candidates)
	}
	selectedSnapshots := candidates[:attempts]
	selected := make([]Provider, 0, len(selectedSnapshots))
	meteredAttempts := uint64(0)
	for _, candidate := range selectedSnapshots {
		selected = append(selected, Provider{
			ID:         candidate.ID,
			Capability: candidate.Capability,
			Metered:    candidate.Metered,
			Priority:   candidate.Priority,
		})
		if candidate.Metered {
			meteredAttempts++
		}
	}

	reservedCost := uint64(0)
	reservationID := uint64(0)
	if meteredAttempts > 0 {
		if request.EstimatedCostMicros == 0 || c.budgetPolicy.LimitMicros == 0 || c.budgetPolicy.PerOperationMicros == 0 {
			return Plan{}, ErrBudgetRequired
		}
		if request.EstimatedCostMicros > ^uint64(0)/meteredAttempts {
			return Plan{}, fmt.Errorf("%w: estimated cost overflow", ErrBudgetExceeded)
		}
		reservedCost = request.EstimatedCostMicros * meteredAttempts
		if reservedCost > c.budgetPolicy.PerOperationMicros {
			return Plan{}, ErrBudgetExceeded
		}
		if c.spent > c.budgetPolicy.LimitMicros || c.reserved > c.budgetPolicy.LimitMicros-c.spent {
			return Plan{}, ErrBudgetExceeded
		}
		available := c.budgetPolicy.LimitMicros - c.spent - c.reserved
		if reservedCost > available {
			return Plan{}, ErrBudgetExceeded
		}
		c.nextReserveID++
		if c.nextReserveID == 0 {
			c.nextReserveID++
		}
		reservationID = c.nextReserveID
		c.reservations[reservationID] = reservation{amount: reservedCost}
		c.reserved += reservedCost
	}

	return Plan{
		Capability:         request.Capability,
		Providers:          selected,
		ReservationID:      reservationID,
		ReservedCostMicros: reservedCost,
	}, nil
}

// Commit records actual metered cost after an operation. If actual cost exceeds
// the reserved ceiling, accounting still records the real spend and returns an
// error so an operator can see the policy overrun instead of hiding it.
func (c *Controller) Commit(reservationID, actualCostMicros uint64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if reservationID == 0 {
		if actualCostMicros != 0 {
			return fmt.Errorf("%w: unreserved metered cost", ErrBudgetExceeded)
		}
		return nil
	}
	entry, exists := c.reservations[reservationID]
	if !exists {
		return fmt.Errorf("%w: reservation does not exist", ErrInvalidOperation)
	}
	delete(c.reservations, reservationID)
	if entry.amount > c.reserved {
		return fmt.Errorf("%w: reservation accounting is inconsistent", ErrInvalidOperation)
	}
	c.reserved -= entry.amount
	if ^uint64(0)-c.spent < actualCostMicros {
		c.spent = ^uint64(0)
		return ErrBudgetExceeded
	}
	c.spent += actualCostMicros
	if actualCostMicros > entry.amount || (c.budgetPolicy.LimitMicros > 0 && c.spent > c.budgetPolicy.LimitMicros) {
		return ErrBudgetExceeded
	}
	return nil
}

func (c *Controller) Release(reservationID uint64) error {
	if reservationID == 0 {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, exists := c.reservations[reservationID]
	if !exists {
		return fmt.Errorf("%w: reservation does not exist", ErrInvalidOperation)
	}
	delete(c.reservations, reservationID)
	if entry.amount > c.reserved {
		return fmt.Errorf("%w: reservation accounting is inconsistent", ErrInvalidOperation)
	}
	c.reserved -= entry.amount
	return nil
}

func (c *Controller) SetHealth(providerID string, health Health) error {
	if !validHealth(health) {
		return fmt.Errorf("%w: invalid provider health", ErrInvalidOperation)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	state, exists := c.providers[strings.TrimSpace(providerID)]
	if !exists {
		return ErrUnknownProvider
	}
	state.health = health
	return nil
}

// Record stores only bounded operational metadata. It intentionally has no
// fields for query text, URLs, goals, page content, credentials, citations, or
// provider response bodies.
func (c *Controller) Record(providerID string, outcome Outcome, duration time.Duration) error {
	if !validOutcome(outcome) {
		return fmt.Errorf("%w: invalid observation outcome", ErrInvalidOperation)
	}
	providerID = strings.TrimSpace(providerID)
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.providers[providerID]; !exists {
		return ErrUnknownProvider
	}
	key := observationKey{providerID: providerID, outcome: outcome}
	value := c.observations[key]
	value.count++
	if duration > 0 {
		milliseconds := uint64(duration / time.Millisecond)
		if ^uint64(0)-value.durationMS < milliseconds {
			value.durationMS = ^uint64(0)
		} else {
			value.durationMS += milliseconds
		}
	}
	c.observations[key] = value
	return nil
}

func (c *Controller) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	providers := make([]ProviderSnapshot, 0, len(c.providers))
	for _, state := range c.providers {
		providers = append(providers, ProviderSnapshot{
			ID:         state.provider.ID,
			Capability: state.provider.Capability,
			Metered:    state.provider.Metered,
			Priority:   state.provider.Priority,
			Health:     state.health,
		})
	}
	sort.Slice(providers, func(i, j int) bool { return providers[i].ID < providers[j].ID })

	observations := make([]Observation, 0, len(c.observations))
	for key, value := range c.observations {
		observations = append(observations, Observation{
			ProviderID: key.providerID,
			Outcome:    key.outcome,
			Count:      value.count,
			DurationMS: value.durationMS,
		})
	}
	sort.Slice(observations, func(i, j int) bool {
		if observations[i].ProviderID != observations[j].ProviderID {
			return observations[i].ProviderID < observations[j].ProviderID
		}
		return observations[i].Outcome < observations[j].Outcome
	})

	available := uint64(0)
	if c.budgetPolicy.LimitMicros > c.spent && c.budgetPolicy.LimitMicros-c.spent > c.reserved {
		available = c.budgetPolicy.LimitMicros - c.spent - c.reserved
	}
	return Snapshot{
		Budget: BudgetSnapshot{
			LimitMicros:        c.budgetPolicy.LimitMicros,
			PerOperationMicros: c.budgetPolicy.PerOperationMicros,
			ReservedMicros:     c.reserved,
			SpentMicros:        c.spent,
			AvailableMicros:    available,
		},
		Providers:    providers,
		Observations: observations,
	}
}

func normalizeProvider(input Provider) (Provider, error) {
	id := strings.TrimSpace(input.ID)
	if id == "" || utf8.RuneCountInString(id) > MaxProviderIDRunes || !validProviderID(id) {
		return Provider{}, fmt.Errorf("%w: provider ID is invalid", ErrInvalidOperation)
	}
	if !validCapability(input.Capability) {
		return Provider{}, fmt.Errorf("%w: provider capability is invalid", ErrInvalidOperation)
	}
	return Provider{ID: id, Capability: input.Capability, Metered: input.Metered, Priority: input.Priority}, nil
}

func normalizeRequest(input Request) (Request, error) {
	if !validCapability(input.Capability) {
		return Request{}, fmt.Errorf("%w: capability is required", ErrInvalidRequest)
	}
	attempts := input.MaxAttempts
	if attempts == 0 {
		attempts = 1
	}
	if attempts < 1 || attempts > MaxPlanAttempts {
		return Request{}, fmt.Errorf("%w: max attempts must be between 1 and %d", ErrInvalidRequest, MaxPlanAttempts)
	}
	return Request{
		Capability:          input.Capability,
		AllowMetered:        input.AllowMetered,
		EstimatedCostMicros: input.EstimatedCostMicros,
		MaxAttempts:         attempts,
	}, nil
}

func validCapability(capability Capability) bool {
	switch capability {
	case CapabilitySearch, CapabilityFetch, CapabilityResearch, CapabilityAgent, CapabilityBrowser:
		return true
	default:
		return false
	}
}

func validHealth(health Health) bool {
	switch health {
	case HealthHealthy, HealthDegraded, HealthUnavailable:
		return true
	default:
		return false
	}
}

func validOutcome(outcome Outcome) bool {
	switch outcome {
	case OutcomeSuccess, OutcomeFailure, OutcomeTimeout, OutcomeCancelled, OutcomeRejected:
		return true
	default:
		return false
	}
}

func validProviderID(value string) bool {
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' {
			continue
		}
		return false
	}
	return true
}

func healthRank(health Health) int {
	switch health {
	case HealthHealthy:
		return 0
	case HealthDegraded:
		return 1
	default:
		return 2
	}
}
