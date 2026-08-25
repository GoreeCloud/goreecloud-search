package search

import (
	"context"
	"reflect"
	"testing"
	"time"
)

type registryLegacyProvider struct{ name string }

func (p registryLegacyProvider) Name() string { return p.name }
func (p registryLegacyProvider) Search(context.Context, string) ([]Result, error) { return nil, nil }

type registryCategoryProvider struct {
	name       string
	categories []string
}

func (p registryCategoryProvider) Name() string { return p.name }
func (p registryCategoryProvider) Search(context.Context, string) ([]Result, error) { return nil, nil }
func (p registryCategoryProvider) Categories() []string { return append([]string(nil), p.categories...) }

func TestProviderDefinitionsAreSanitizedAndDeterministic(t *testing.T) {
	engine := NewEngine(time.Second,
		registryCategoryProvider{name: "zeta", categories: []string{"news", "general", "NEWS", "invalid"}},
		registryLegacyProvider{name: "alpha"},
	)

	got := engine.ProviderDefinitions()
	want := []ProviderDefinition{
		{Name: "alpha", Categories: []string{CategoryGeneral}, Legacy: true},
		{Name: "zeta", Categories: []string{CategoryGeneral, CategoryNews}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ProviderDefinitions() = %#v, want %#v", got, want)
	}
}

func TestProviderDefinitionsDoNotExposeRuntimeState(t *testing.T) {
	engine := NewEngine(time.Second, registryLegacyProvider{name: "example"})
	definitions := engine.ProviderDefinitions()
	if len(definitions) != 1 {
		t.Fatalf("definitions = %d, want 1", len(definitions))
	}
	if definitions[0].Name != "example" || !definitions[0].Legacy {
		t.Fatalf("unexpected definition: %#v", definitions[0])
	}
}
