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

type ProviderStatus struct {
	Name  string `json:"name"`
	Error string `json:"error,omitempty"`
	Count int    `json:"count"`
}

type Response struct {
	Query     string           `json:"query"`
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

func (e *Engine) Search(ctx context.Context, raw string) (Response, error) {
	query, err := ValidateQuery(raw)
	if err != nil {
		return Response{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	type providerResult struct {
		status  ProviderStatus
		results []Result
	}

	ch := make(chan providerResult, len(e.providers))
	var wg sync.WaitGroup
	for _, provider := range e.providers {
		provider := provider
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, searchErr := provider.Search(ctx, query)
			status := ProviderStatus{Name: provider.Name(), Count: len(items)}
			if searchErr != nil {
				status.Error = searchErr.Error()
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

	response := Response{Query: query, Results: []Result{}, Providers: []ProviderStatus{}}
	seen := map[string]struct{}{}
	for item := range ch {
		response.Providers = append(response.Providers, item.status)
		if item.status.Error != "" {
			response.Degraded = true
		}
		for _, result := range item.results {
			normalized, ok := normalizeResultURL(result.URL)
			if !ok {
				continue
			}
			if _, exists := seen[normalized]; exists {
				continue
			}
			seen[normalized] = struct{}{}
			result.URL = normalized
			response.Results = append(response.Results, result)
		}
	}

	sort.SliceStable(response.Results, func(i, j int) bool {
		if response.Results[i].Score == response.Results[j].Score {
			return response.Results[i].URL < response.Results[j].URL
		}
		return response.Results[i].Score > response.Results[j].Score
	})
	sort.Slice(response.Providers, func(i, j int) bool { return response.Providers[i].Name < response.Providers[j].Name })
	return response, nil
}

func normalizeResultURL(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", false
	}
	parsed.Fragment = ""
	return parsed.String(), true
}
