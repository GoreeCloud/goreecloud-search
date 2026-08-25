package search

import "context"

// CategorySearcher extends CategoryProvider with an explicit category-aware
// execution path. Providers may still implement CategoryProvider alone when
// they represent one fixed category, but a provider declaring multiple
// categories must implement this interface before specialized requests can be
// routed through it.
type CategorySearcher interface {
	CategoryProvider
	SearchCategory(context.Context, string, string) ([]Result, error)
}
