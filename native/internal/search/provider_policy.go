package search

import "time"

const (
	MaxProviderRequestTimeout = 8 * time.Second
	MaxProviderResponseBytes  = 8 * 1024 * 1024
)

// ProviderExecutionPolicy declares the transport/resource bounds an external
// provider adapter is designed to enforce. RequestTimeout is also enforced by
// the native engine as a child context. MaxResults is enforced after the
// provider returns and may not exceed the Search-owned global ceiling.
//
// MaxResponseBytes is an adapter obligation: an HTTP-backed adapter must apply
// the limit to the response body before unbounded decode/allocation. Merely
// declaring this value does not prove that a live provider is production-safe.
type ProviderExecutionPolicy struct {
	RequestTimeout   time.Duration
	MaxResponseBytes int64
	MaxResults       int
}

// BoundedProvider is the opt-in contract for providers that declare explicit
// transport/resource bounds. Providers implementing this interface fail closed
// if any declared bound is invalid.
type BoundedProvider interface {
	Provider
	ExecutionPolicy() ProviderExecutionPolicy
}

func validatedProviderExecutionPolicy(provider Provider) (ProviderExecutionPolicy, bool, bool) {
	bounded, declared := provider.(BoundedProvider)
	if !declared {
		return ProviderExecutionPolicy{}, false, true
	}
	policy := bounded.ExecutionPolicy()
	valid := policy.RequestTimeout > 0 &&
		policy.RequestTimeout <= MaxProviderRequestTimeout &&
		policy.MaxResponseBytes > 0 &&
		policy.MaxResponseBytes <= MaxProviderResponseBytes &&
		policy.MaxResults > 0 &&
		policy.MaxResults <= MaxResultsPerProvider
	return policy, true, valid
}
