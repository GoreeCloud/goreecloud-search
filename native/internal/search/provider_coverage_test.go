package search

import (
	"context"
	"reflect"
	"testing"
)

type coverageLegacyProvider struct{}

func (coverageLegacyProvider) Name() string { return "legacy" }
func (coverageLegacyProvider) Search(context.Context, string) ([]Result, error) {
	return nil, nil
}

type coverageCategoryProvider struct {
	name       string
	categories []string
}

func (p coverageCategoryProvider) Name() string { return p.name }
func (p coverageCategoryProvider) Search(context.Context, string) ([]Result, error) {
	return nil, nil
}
func (p coverageCategoryProvider) Categories() []string {
	return append([]string(nil), p.categories...)
}

type coverageCategorySearcher struct {
	coverageCategoryProvider
}

func (p coverageCategorySearcher) SearchCategory(context.Context, string, string) ([]Result, error) {
	return nil, nil
}

func TestProviderBackedCategoriesExcludesDevelopmentFallback(t *testing.T) {
	if got := ProviderBackedCategories(); len(got) != 0 {
		t.Fatalf("provider-backed categories = %v, want none", got)
	}
}

func TestProviderBackedCategoriesReflectsExecutableProviderPaths(t *testing.T) {
	tests := []struct {
		name      string
		providers []Provider
		want      []string
	}{
		{
			name:      "legacy provider remains general only",
			providers: []Provider{coverageLegacyProvider{}},
			want:      []string{CategoryGeneral},
		},
		{
			name: "single specialized category can use base search",
			providers: []Provider{coverageCategoryProvider{
				name:       "images",
				categories: []string{CategoryImages},
			}},
			want: []string{CategoryImages},
		},
		{
			name: "multi category declaration without category search is not executable for specialized category",
			providers: []Provider{coverageCategoryProvider{
				name:       "declared-only",
				categories: []string{CategoryGeneral, CategoryImages},
			}},
			want: []string{CategoryGeneral},
		},
		{
			name: "category searcher covers every declared category",
			providers: []Provider{coverageCategorySearcher{coverageCategoryProvider{
				name:       "all",
				categories: append([]string(nil), SupportedCategories...),
			}}},
			want: append([]string(nil), SupportedCategories...),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ProviderBackedCategories(test.providers...); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("provider-backed categories = %v, want %v", got, test.want)
			}
		})
	}
}
