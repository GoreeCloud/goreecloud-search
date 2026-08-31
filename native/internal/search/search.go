package search

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	MaxQueryRunes         = 512
	MaxProviderNameRunes  = 128
	MaxResultsPerProvider = 512
)

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
	Title              string     `json:"title"`
	URL                string     `json:"url"`
	Snippet            string     `json:"snippet,omitempty"`
	Provider           string     `json:"provider"`
	Score              int        `json:"score"`
	SourceCount        int        `json:"source_count,omitempty"`
	Sources            []string   `json:"sources,omitempty"`
	PublishedAt        *time.Time `json:"published_at,omitempty"`
	PublishedAtSource  string     `json:"published_at_source,omitempty"`
	publishedAtTrusted bool
	recencyBonus       int
}

type Provider interface {
	Name() string
	Search(context.Context, string) ([]Result, error)
}

// PublishedAtProvider is an explicit metadata-authority opt-in. Implementations
// must return true only when Result.PublishedAt is copied from a trustworthy
// upstream publication/update field with publication semantics. Search must not
// infer this authority from snippets, URLs, crawl time, or provider score.
type PublishedAtProvider interface {
	Provider
	PublishedAtAuthoritative() bool
}

// CategoryProvider extends the base provider contract with explicit category
// capabilities. Providers that do not implement this interface remain General-
// only for compatibility and may not be invoked for another category.
type CategoryProvider interface {
	Provider
	Categories() []string
}

type ProviderDefinition struct {
	Name                     string   `json:"name"`
	Categories               []string `json:"categories"`
	Legacy                   bool     `json:"legacy_general_only"`
	PublishedAtAuthoritative bool     `json:"published_at_authoritative"`
}

type ProviderStatus struct {
	Name      string `json:"name"`
	State     string `json:"state"`
	Code      string `json:"code,omitempty"`
	Count     int    `json:"count"`
	Truncated bool   `json:"truncated,omitempty"`
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
// provider identity and category/metadata capabilities. It does not expose
// credentials, endpoint configuration, runtime errors, request state, or
// mutable controls.
func (e *Engine) ProviderDefinitions() []ProviderDefinition {
	definitions := make([]ProviderDefinition, 0, len(e.providers))
	for _, provider := range e.providers {
		if provider == nil {
			continue
		}
		name, ok := normalizeProviderName(provider.Name())
		if !ok {
			continue
		}
		definition := ProviderDefinition{Name: name}
		if timestampProvider, supportsTimestamps := provider.(PublishedAtProvider); supportsTimestamps {
			definition.PublishedAtAuthoritative = timestampProvider.PublishedAtAuthoritative()
		}
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

func normalizeProviderName(raw string) (string, bool) {
	name := strings.TrimSpace(raw)
	if name == "" || len([]rune(name)) > MaxProviderNameRunes {
		return "", false
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", false
		}
	}
	return name, true
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
		if _, ok := validExecutableProvider(provider); ok && providerCanExecuteCategory(provider, category) {
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
	requestNow := time.Now().UTC()
	intent := parseQueryIntent(query)

	type providerResult struct {
		index   int
		status  ProviderStatus
		results []Result
	}

	type selectedProvider struct {
		provider                 Provider
		name                     string
		publishedAtAuthoritative bool
	}
	selected := make([]selectedProvider, 0, len(e.providers))
	for _, provider := range e.providers {
		name, ok := validExecutableProvider(provider)
		if ok && providerCanExecuteCategory(provider, category) {
			authoritative := false
			if timestampProvider, supportsTimestamps := provider.(PublishedAtProvider); supportsTimestamps {
				authoritative = timestampProvider.PublishedAtAuthoritative()
			}
			selected = append(selected, selectedProvider{provider: provider, name: name, publishedAtAuthoritative: authoritative})
		}
	}

	// One slot per selected provider prevents a late provider completion from
	// blocking after this request has already returned at its deadline.
	ch := make(chan providerResult, len(selected))
	for index, selectedProvider := range selected {
		index, selectedProvider := index, selectedProvider
		go func() {
			items, searchErr := executeProviderSearch(ctx, selectedProvider.provider, query, category)
			status := ProviderStatus{Name: selectedProvider.name, State: ProviderStateAvailable}
			if searchErr != nil {
				status.State = ProviderStateUnavailable
				status.Code = classifyProviderFailure(searchErr)
				items = nil
			} else {
				items, status.Truncated = boundProviderResults(items)
				status.Count = len(items)
			}
			for i := range items {
				items[i].Provider = selectedProvider.name
				publishedAt, trusted := normalizeAuthoritativePublishedAt(items[i].PublishedAt, selectedProvider.publishedAtAuthoritative, requestNow)
				items[i].PublishedAt = publishedAt
				items[i].PublishedAtSource = ""
				items[i].publishedAtTrusted = trusted
				if trusted {
					items[i].PublishedAtSource = selectedProvider.name
				}
			}
			ch <- providerResult{index: index, status: status, results: items}
		}()
	}

	response := Response{Query: query, Category: category, Results: []Result{}, Providers: []ProviderStatus{}}
	candidates := make([]Result, 0)
	resolved := make([]bool, len(selected))
	remaining := len(selected)

	consume := func(item providerResult) {
		if item.index < 0 || item.index >= len(selected) || resolved[item.index] {
			return
		}
		resolved[item.index] = true
		remaining--
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
			result.recencyBonus = freshnessScore(intent, category, result, requestNow)
			candidates = append(candidates, result)
		}
	}

	for remaining > 0 {
		select {
		case item := <-ch:
			consume(item)
		case <-ctx.Done():
			for index, provider := range selected {
				if resolved[index] {
					continue
				}
				resolved[index] = true
				remaining--
				response.Providers = append(response.Providers, ProviderStatus{
					Name: provider.name, State: ProviderStateUnavailable, Code: ProviderCodeTimeout,
				})
				response.Degraded = true
			}
		}
	}

	response.Results = rankResults(query, candidates)
	sort.Slice(response.Providers, func(i, j int) bool { return response.Providers[i].Name < response.Providers[j].Name })
	return response, nil
}

func boundProviderResults(items []Result) ([]Result, bool) {
	if len(items) <= MaxResultsPerProvider {
		return items, false
	}
	bounded := append([]Result(nil), items[:MaxResultsPerProvider]...)
	return bounded, true
}

func validExecutableProvider(provider Provider) (string, bool) {
	if provider == nil {
		return "", false
	}
	return normalizeProviderName(provider.Name())
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
	if err != nil || parsed.Host == "" || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", false
	}
	parsed.Fragment = ""
	return parsed.String(), true
}
