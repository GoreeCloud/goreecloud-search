package search

import (
	"context"
	"reflect"
	"testing"
	"time"
)

type boundedTestProvider struct {
	name    string
	policy  ProviderExecutionPolicy
	results []Result
	release <-chan struct{}
	invoked *bool
}

func (p boundedTestProvider) Name() string { return p.name }
func (p boundedTestProvider) ExecutionPolicy() ProviderExecutionPolicy { return p.policy }
func (p boundedTestProvider) Search(ctx context.Context, _ string) ([]Result, error) {
	if p.invoked != nil {
		*p.invoked = true
	}
	if p.release != nil {
		<-p.release // Deliberately ignore ctx; the engine policy must still bound the request.
	}
	return append([]Result(nil), p.results...), nil
}

func TestBoundedProviderDefinitionIsSanitized(t *testing.T) {
	provider := boundedTestProvider{
		name: "bounded",
		policy: ProviderExecutionPolicy{
			RequestTimeout:   250 * time.Millisecond,
			MaxResponseBytes: 256 * 1024,
			MaxResults:       25,
		},
	}
	engine := NewEngine(time.Second, provider)

	got := engine.ProviderDefinitions()
	want := []ProviderDefinition{{
		Name:                    "bounded",
		Categories:              []string{CategoryGeneral},
		Legacy:                  true,
		TransportBoundsDeclared: true,
		RequestTimeoutMillis:    250,
		MaxResponseBytes:        256 * 1024,
		MaxResults:              25,
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ProviderDefinitions() = %#v, want %#v", got, want)
	}
}

func TestInvalidBoundedProviderIsNotAdvertisedOrExecuted(t *testing.T) {
	invoked := false
	provider := boundedTestProvider{
		name:    "invalid-bounds",
		invoked: &invoked,
		policy: ProviderExecutionPolicy{
			RequestTimeout:   0,
			MaxResponseBytes: MaxProviderResponseBytes + 1,
			MaxResults:       MaxResultsPerProvider + 1,
		},
		results: []Result{{Title: "Invalid", URL: "https://invalid.example/"}},
	}
	engine := NewEngine(time.Second, provider)

	if definitions := engine.ProviderDefinitions(); len(definitions) != 0 {
		t.Fatalf("invalid bounded provider was advertised: %#v", definitions)
	}
	response, err := engine.Search(context.Background(), "invalid")
	if err != nil {
		t.Fatal(err)
	}
	if invoked {
		t.Fatal("invalid bounded provider was executed")
	}
	if len(response.Providers) != 0 || len(response.Results) != 0 {
		t.Fatalf("invalid bounded provider entered response evidence: %#v", response)
	}
}

func TestBoundedProviderResultCeilingIsEnforced(t *testing.T) {
	provider := boundedTestProvider{
		name: "small-result-budget",
		policy: ProviderExecutionPolicy{
			RequestTimeout:   time.Second,
			MaxResponseBytes: 64 * 1024,
			MaxResults:       2,
		},
		results: []Result{
			{Title: "One", URL: "https://one.example/"},
			{Title: "Two", URL: "https://two.example/"},
			{Title: "Three", URL: "https://three.example/"},
		},
	}
	engine := NewEngine(2*time.Second, provider)

	response, err := engine.Search(context.Background(), "example")
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Results) != 2 {
		t.Fatalf("bounded provider returned %d results, want 2", len(response.Results))
	}
	if len(response.Providers) != 1 || !response.Providers[0].Truncated || response.Providers[0].Count != 2 {
		t.Fatalf("provider-specific result ceiling was not disclosed: %#v", response.Providers)
	}
}

func TestBoundedProviderDeadlineWinsBeforeGlobalDeadline(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	provider := boundedTestProvider{
		name:    "bounded-stuck",
		release: release,
		policy: ProviderExecutionPolicy{
			RequestTimeout:   30 * time.Millisecond,
			MaxResponseBytes: 64 * 1024,
			MaxResults:       10,
		},
	}
	engine := NewEngine(time.Second, provider)

	started := time.Now()
	response, err := engine.Search(context.Background(), "bounded")
	elapsed := time.Since(started)
	if err != nil {
		t.Fatal(err)
	}
	if elapsed >= 500*time.Millisecond {
		t.Fatalf("provider policy did not bound the request before global deadline: %s", elapsed)
	}
	if !response.Degraded || len(response.Providers) != 1 {
		t.Fatalf("expected bounded provider timeout evidence: %#v", response)
	}
	status := response.Providers[0]
	if status.State != ProviderStateUnavailable || status.Code != ProviderCodeTimeout {
		t.Fatalf("unexpected bounded provider timeout status: %#v", status)
	}
}

func TestProviderExecutionPolicyValidationBounds(t *testing.T) {
	valid := boundedTestProvider{name: "valid", policy: ProviderExecutionPolicy{
		RequestTimeout:   MaxProviderRequestTimeout,
		MaxResponseBytes: MaxProviderResponseBytes,
		MaxResults:       MaxResultsPerProvider,
	}}
	if _, declared, ok := validatedProviderExecutionPolicy(valid); !declared || !ok {
		t.Fatal("maximum supported provider policy should validate")
	}

	invalidCases := []ProviderExecutionPolicy{
		{RequestTimeout: MaxProviderRequestTimeout + time.Millisecond, MaxResponseBytes: 1, MaxResults: 1},
		{RequestTimeout: time.Second, MaxResponseBytes: MaxProviderResponseBytes + 1, MaxResults: 1},
		{RequestTimeout: time.Second, MaxResponseBytes: 1, MaxResults: MaxResultsPerProvider + 1},
		{RequestTimeout: time.Second, MaxResponseBytes: 1, MaxResults: 0},
	}
	for index, policy := range invalidCases {
		provider := boundedTestProvider{name: "invalid", policy: policy}
		if _, declared, ok := validatedProviderExecutionPolicy(provider); !declared || ok {
			t.Fatalf("invalid policy %d unexpectedly validated: %#v", index, policy)
		}
	}
}
