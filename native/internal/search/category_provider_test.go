package search

import (
	"context"
	"testing"
	"time"
)

type categoryTestProvider struct {
	name       string
	categories []string
	calls      int
}

func (p *categoryTestProvider) Name() string { return p.name }
func (p *categoryTestProvider) Categories() []string { return append([]string(nil), p.categories...) }
func (p *categoryTestProvider) Search(_ context.Context, query string) ([]Result, error) {
	p.calls++
	return []Result{{Title: query, URL: "https://example.com/" + p.name, Score: 1}}, nil
}

type generalOnlyProvider struct{ calls int }

func (p *generalOnlyProvider) Name() string { return "general-only" }
func (p *generalOnlyProvider) Search(_ context.Context, query string) ([]Result, error) {
	p.calls++
	return []Result{{Title: query, URL: "https://example.com/general", Score: 1}}, nil
}

func TestLegacyProviderRemainsGeneralOnly(t *testing.T) {
	provider := &generalOnlyProvider{}
	engine := NewEngine(time.Second, provider)
	if !engine.SupportsCategory(CategoryGeneral) {
		t.Fatal("general category must remain implemented")
	}
	if engine.SupportsCategory(CategoryImages) {
		t.Fatal("legacy provider must not silently claim images")
	}
	if _, err := engine.SearchCategory(t.Context(), "cloud", CategoryImages); err == nil {
		t.Fatal("unimplemented category must fail closed")
	}
	if provider.calls != 0 {
		t.Fatalf("unimplemented category invoked provider %d times", provider.calls)
	}
}

func TestCategoryProviderEnablesOnlyDeclaredCategory(t *testing.T) {
	images := &categoryTestProvider{name: "images", categories: []string{CategoryImages}}
	news := &categoryTestProvider{name: "news", categories: []string{CategoryNews}}
	engine := NewEngine(time.Second, images, news)
	if !engine.SupportsCategory(CategoryImages) || !engine.SupportsCategory(CategoryNews) {
		t.Fatal("declared provider categories must be implemented")
	}
	if engine.SupportsCategory(CategoryVideos) {
		t.Fatal("undeclared category must remain unavailable")
	}

	response, err := engine.SearchCategory(t.Context(), "cloud", CategoryImages)
	if err != nil {
		t.Fatal(err)
	}
	if response.Category != CategoryImages {
		t.Fatalf("category = %q, want images", response.Category)
	}
	if images.calls != 1 || news.calls != 0 {
		t.Fatalf("category routing calls images=%d news=%d", images.calls, news.calls)
	}
	if len(response.Providers) != 1 || response.Providers[0].Name != "images" {
		t.Fatalf("unexpected providers: %#v", response.Providers)
	}
}

func TestInvalidProviderCategoryDoesNotEnableExecution(t *testing.T) {
	provider := &categoryTestProvider{name: "bad", categories: []string{"everything"}}
	engine := NewEngine(time.Second, provider)
	if engine.SupportsCategory(CategoryFiles) {
		t.Fatal("invalid provider category must not enable files")
	}
}
