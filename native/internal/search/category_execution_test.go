package search

import (
	"context"
	"testing"
	"time"
)

type fixedCategoryProvider struct {
	name     string
	category string
}

func (p fixedCategoryProvider) Name() string { return p.name }
func (p fixedCategoryProvider) Categories() []string { return []string{p.category} }
func (p fixedCategoryProvider) Search(context.Context, string) ([]Result, error) {
	return []Result{{Title: "fixed", URL: "https://example.com/fixed", Score: 1}}, nil
}

type declaredMultiCategoryProvider struct{}

func (declaredMultiCategoryProvider) Name() string { return "declared-multi" }
func (declaredMultiCategoryProvider) Categories() []string {
	return []string{CategoryImages, CategoryNews}
}
func (declaredMultiCategoryProvider) Search(context.Context, string) ([]Result, error) {
	return []Result{{Title: "ambiguous", URL: "https://example.com/ambiguous", Score: 1}}, nil
}

type categoryAwareProvider struct {
	seenCategory string
}

func (p *categoryAwareProvider) Name() string { return "category-aware" }
func (p *categoryAwareProvider) Categories() []string {
	return []string{CategoryImages, CategoryNews}
}
func (p *categoryAwareProvider) Search(context.Context, string) ([]Result, error) {
	return nil, nil
}
func (p *categoryAwareProvider) SearchCategory(_ context.Context, _ string, category string) ([]Result, error) {
	p.seenCategory = category
	return []Result{{Title: category, URL: "https://example.com/" + category, Score: 2}}, nil
}

func TestSinglePurposeCategoryProviderRemainsExecutable(t *testing.T) {
	engine := NewEngine(time.Second, fixedCategoryProvider{name: "images-only", category: CategoryImages})
	if !engine.SupportsCategory(CategoryImages) {
		t.Fatal("expected a single-purpose category provider to remain executable")
	}
	response, err := engine.SearchCategory(context.Background(), "query", CategoryImages)
	if err != nil {
		t.Fatalf("SearchCategory returned error: %v", err)
	}
	if len(response.Results) != 1 || response.Results[0].Provider != "images-only" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestMultiCategoryDeclarationRequiresCategoryAwareExecution(t *testing.T) {
	engine := NewEngine(time.Second, declaredMultiCategoryProvider{})
	if engine.SupportsCategory(CategoryImages) || engine.SupportsCategory(CategoryNews) {
		t.Fatal("multi-category declaration without category-aware execution must fail closed")
	}
	if _, err := engine.SearchCategory(context.Background(), "query", CategoryImages); err == nil {
		t.Fatal("expected unimplemented category error")
	}
}

func TestCategoryAwareProviderReceivesRequestedCategory(t *testing.T) {
	provider := &categoryAwareProvider{}
	engine := NewEngine(time.Second, provider)
	response, err := engine.SearchCategory(context.Background(), "query", CategoryNews)
	if err != nil {
		t.Fatalf("SearchCategory returned error: %v", err)
	}
	if provider.seenCategory != CategoryNews {
		t.Fatalf("provider received %q, want %q", provider.seenCategory, CategoryNews)
	}
	if len(response.Results) != 1 || response.Results[0].Title != CategoryNews {
		t.Fatalf("unexpected response: %#v", response)
	}
}
