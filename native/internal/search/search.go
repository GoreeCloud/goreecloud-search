package search

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

const MaxQueryRunes = 512

const (
	CategoryGeneral = "general"
	CategoryImages  = "images"
	CategoryVideos  = "videos"
	CategoryNews    = "news"
	CategoryFiles   = "files"
)

var SupportedCategories = []string{CategoryGeneral, CategoryImages, CategoryVideos, CategoryNews, CategoryFiles}

const (
	ProviderStateAvailable   = "available"
	ProviderStateUnavailable = "unavailable"
	ProviderCodeUnavailable  = "provider_unavailable"
	ProviderCodeTimeout      = "provider_timeout"
)

type Result struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	Snippet  string `json:"snippet,omitempty"`
	Provider string `json:"provider"`
	Score    int    `json:"score"`
}

type Provider interface {
	Name() string
	Search(context.Context, string) ([]Result, error)
}

// CategoryProvider extends the base provider contract with explicit category
// capabilities. Providers that do not implement this interface remain General-
// only for compatibility and may not be invoked for another category.
type CategoryProvider interface {
	Provider
	Categories() []string
}

type ProviderDefinition struct {
	Name       string   `json:"name"`
	Categories []string `json:"categories"`
	Legacy     bool     `json:"legacy_general_only"`
}

type ProviderStatus struct {
	Name  string `json:"name"`
	State string `json:"state"`
	Code  string `json:"code,omitempty"`
	Count int    `json:"count"`
}

type Response struct {
	Query     string           `json:"query"`
	Category  string           `json:"category"`
	Results   []Result         `json:"results"`
	Providers []ProviderStatus `json:"providers"`
	Degraded  bool             `json:"degraded"`
}

type Engine struct {
	providers []Provider
	timeout   time.Duration
}

func NewEngine(timeout time.Duration, providers ...Provider) *Engine {
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	return &Engine{providers: append([]Provider(nil), providers...), timeout: timeout}
}

// ProviderDefinitions returns a sanitized, deterministic view of configured
// provider identity and category capabilities. It does not expose credentials,
// endpoint configuration, runtime errors, request state, or mutable controls.
func (e *Engine) ProviderDefinitions() []ProviderDefinition {
	definitions := make([]ProviderDefinition, 0, len(e.providers))
	for _, provider := range e.providers {
		if provider == nil {
			continue
		}
		definition := ProviderDefinition{Name: strings.TrimSpace(provider.Name())}
		categorized, ok := provider.(CategoryProvider)
		if !ok {
			definition.Categories = []string{CategoryGeneral}
			definition.Legacy = true
			definitions = append(definitions, definition)
			continue
		}
		seen := map[string]bool{}
		for _, rawCategory := range categorized.Categories() {
			category, err := ValidateCategory(rawCategory)
			if err == nil && !seen[category] {
				definition.Categories = append(definition.Categories, category)
				seen[category] = true
			}
		}
		sort.Strings(definition.Categories)
		definitions = append(definitions, definition)
	}
	sort.Slice(definitions, func(i, j int) bool {
		if definitions[i].Name == definitions[j].Name {
			return strings.Join(definitions[i].Categories, ",") < strings.Join(definitions[j].Categories, ",")
		}
		return definitions[i].Name < definitions[j].Name
	})
	return definitions
}

func ValidateQuery(raw string) (string, error) {
	query := strings.TrimSpace(raw)
	if query == "" {
		return "", errors.New("query is required")
	}
	if len([]rune(query)) > MaxQueryRunes {
		return "", errors.New("query exceeds maximum length")
	}
	return query, nil
}

func ValidateCategory(raw string) (string, error) {
	category := strings.ToLower(strings.TrimSpace(raw))
	if category == "" {
		return CategoryGeneral, nil
	}
	for _, allowed := range SupportedCategories {
		if category == allowed {
			return category, nil
		}
	}
	return "", errors.New("unsupported search category")
}

// SupportsCategory reports whether the native engine has a usable execution
// path for a validated category. General remains implemented even with no
// configured providers so the development shell retains its bounded empty
// result behavior. Additional categories require an executable provider path.
func (e *Engine) SupportsCategory(category string) bool {
	category, err := ValidateCategory(category)
	if err != nil {
		return false
	}
	if category == CategoryGeneral {
		return true
	}
	for _, provider := range e.providers {
		if providerCanExecuteCategory(provider, category) {
			return true
		}
	}
	return false
}

func (e *Engine) Search(ctx context.Context, raw string) (Response, error) {
	return e.SearchCategory(ctx, raw, CategoryGeneral)
}

func (e *Engine) SearchCategory(ctx context.Context, raw, rawCategory string) (Response, error) {
	query, err := ValidateQuery(raw)
	if err != nil {
		return Response{}, err
	}
	category, err := ValidateCategory(rawCategory)
	if err != nil {
		return Response{}, err
	}
	if !e.SupportsCategory(category) {
		return Response{}, errors.New("search category is not implemented in the native provider layer")
	}

	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	type providerResult struct {
		status  ProviderStatus
		results []Result
	}

	selected := make([]Provider, 0, len(e.providers))
	for _, provider := range e.providers {
		if providerCanExecuteCategory(provider, category) {
			selected = append(selected, provider)
		}
	}

	ch := make(chan providerResult, len(selected))
	var wg sync.WaitGroup
	for _, provider := range selected {
		provider := provider
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, searchErr := executeProviderSearch(ctx, provider, query, category)
			status := ProviderStatus{Name: provider.Name(), State: ProviderStateAvailable, Count: len(items)}
			if searchErr != nil {
				status.State = ProviderStateUnavailable
				status.Code = classifyProviderFailure(searchErr)
				status.Count = 0
				items = nil
			}
			for i := range items {
				items[i].Provider = provider.Name()
			}
			ch <- providerResult{status: status, results: items}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	response := Response{Query: query, Category: category, Results: []Result{}, Providers: []ProviderStatus{}}
	bestByURL := map[string]Result{}
	for item := range ch {
		response.Providers = append(response.Providers, item.status)
		if item.status.State != ProviderStateAvailable {
			response.Degraded = true
		}
		for _, result := range item.results {
			normalized, ok := normalizeResultURL(result.URL)
			if !ok {
				continue
			}
			result.URL = normalized
			current, exists := bestByURL[normalized]
			if !exists || resultBetterThan(result, current) {
				bestByURL[normalized] = result
			}
		}
	}

	for _, result := range bestByURL {
		response.Results = append(response.Results, result)
	}
	sort.Slice(response.Results, func(i, j int) bool {
		if response.Results[i].Score == response.Results[j].Score {
			if response.Results[i].URL == response.Results[j].URL {
				return response.Results[i].Provider < response.Results[j].Provider
			}
			return response.Results[i].URL < response.Results[j].URL
		}
		return response.Results[i].Score > response.Results[j].Score
	})
	sort.Slice(response.Providers, func(i, j int) bool { return response.Providers[i].Name < response.Providers[j].Name })
	return response, nil
}

func providerSupportsCategory(provider Provider, category string) bool {
	if provider == nil {
		return false
	}
	category, err := ValidateCategory(category)
	if err != nil {
		return false
	}
	categorized, ok := provider.(CategoryProvider)
	if !ok {
		return category == CategoryGeneral
	}
	for _, candidate := range categorized.Categories() {
		validated, err := ValidateCategory(candidate)
		if err == nil && validated == category {
			return true
		}
	}
	return false
}

func providerCanExecuteCategory(provider Provider, category string) bool {
	if !providerSupportsCategory(provider, category) {
		return false
	}
	if category == CategoryGeneral {
		return true
	}
	if _, ok := provider.(CategorySearcher); ok {
		return true
	}
	categorized, ok := provider.(CategoryProvider)
	if !ok {
		return false
	}
	valid := map[string]bool{}
	for _, raw := range categorized.Categories() {
		candidate, err := ValidateCategory(raw)
		if err == nil {
			valid[candidate] = true
		}
	}
	return len(valid) == 1 && valid[category]
}

func executeProviderSearch(ctx context.Context, provider Provider, query, category string) ([]Result, error) {
	if categorized, ok := provider.(CategorySearcher); ok {
		return categorized.SearchCategory(ctx, query, category)
	}
	return provider.Search(ctx, query)
}

func resultBetterThan(candidate, current Result) bool {
	if candidate.Score != current.Score {
		return candidate.Score > current.Score
	}
	if candidate.Provider != current.Provider {
		return candidate.Provider < current.Provider
	}
	if candidate.Title != current.Title {
		return candidate.Title < current.Title
	}
	return candidate.Snippet < current.Snippet
}

func classifyProviderFailure(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return ProviderCodeTimeout
	}
	return ProviderCodeUnavailable
}

func normalizeResultURL(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", false
	}
	parsed.Fragment = ""
	return parsed.String(), true
}
