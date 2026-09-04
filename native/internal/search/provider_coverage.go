package search

// ProviderBackedCategories returns the supported Search categories that have at
// least one configured provider with a valid executable path. Unlike
// Engine.SupportsCategory, this helper never treats General's provider-free
// Development fallback as provider coverage.
func ProviderBackedCategories(providers ...Provider) []string {
	covered := make([]string, 0, len(SupportedCategories))
	for _, category := range SupportedCategories {
		for _, provider := range providers {
			if _, ok := validExecutableProvider(provider); !ok {
				continue
			}
			if providerCanExecuteCategory(provider, category) {
				covered = append(covered, category)
				break
			}
	}
	}
	return covered
}
