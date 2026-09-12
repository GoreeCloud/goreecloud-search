package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

const (
	TinyFishSearchAdapter  = "tinyfish-search-v1"
	TinyFishSearchEndpoint = "https://api.search.tinyfish.ai"

	maxTinyFishProviderScore = 300
)

type TinyFishSearchConfig struct {
	Name       string
	Endpoint   string
	Categories []string
	APIKey     string
}

type TinyFishSearchProvider struct {
	name       string
	endpoint   *url.URL
	categories []string
	apiKey     string
	client     *http.Client
}

type tinyFishSearchResponse struct {
	Query        string                 `json:"query"`
	Results      []tinyFishSearchResult `json:"results"`
	TotalResults int                    `json:"total_results"`
	Page         int                    `json:"page"`
}

type tinyFishSearchResult struct {
	Position  int    `json:"position"`
	SiteName  string `json:"site_name"`
	Title     string `json:"title"`
	Snippet   string `json:"snippet"`
	URL       string `json:"url"`
	Date      string `json:"date"`
	Publisher string `json:"publisher"`
}

func NewTinyFishSearchProvider(config TinyFishSearchConfig) (*TinyFishSearchProvider, error) {
	endpoint, err := validateTinyFishSearchEndpoint(config.Endpoint)
	if err != nil {
		return nil, err
	}
	categories, err := validateProviderCategories(config.Categories)
	if err != nil {
		return nil, err
	}
	for _, category := range categories {
		if category != searchcore.CategoryGeneral && category != searchcore.CategoryNews {
			return nil, errors.New("TinyFish Search supports only general and news categories")
		}
	}

	name := strings.TrimSpace(config.Name)
	if name == "" || len([]rune(name)) > searchcore.MaxProviderNameRunes || strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return nil, errors.New("provider name is invalid")
	}
	apiKey := strings.TrimSpace(config.APIKey)
	if apiKey == "" {
		return nil, errors.New("TinyFish Search API key is required")
	}

	provider := &TinyFishSearchProvider{
		name:       name,
		endpoint:   endpoint,
		categories: categories,
		apiKey:     apiKey,
	}
	provider.client = secureProviderClient(endpoint)
	return provider, nil
}

func newTinyFishSearchProviderWithClient(config TinyFishSearchConfig, client *http.Client) (*TinyFishSearchProvider, error) {
	provider, err := NewTinyFishSearchProvider(config)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("provider client is required")
	}
	provider.client = client
	return provider, nil
}

func (p *TinyFishSearchProvider) Name() string { return p.name }

func (p *TinyFishSearchProvider) Categories() []string {
	return append([]string(nil), p.categories...)
}

func (p *TinyFishSearchProvider) PublishedAtAuthoritative() bool { return false }

func (p *TinyFishSearchProvider) Search(ctx context.Context, query string) ([]searchcore.Result, error) {
	category := searchcore.CategoryGeneral
	if !containsCategory(p.categories, category) {
		if len(p.categories) != 1 {
			return nil, errors.New("provider requires an explicit search category")
		}
		category = p.categories[0]
	}
	return p.SearchCategory(ctx, query, category)
}

func (p *TinyFishSearchProvider) SearchCategory(ctx context.Context, query, rawCategory string) ([]searchcore.Result, error) {
	category, err := searchcore.ValidateCategory(rawCategory)
	if err != nil || !containsCategory(p.categories, category) {
		return nil, errors.New("provider does not support requested category")
	}
	if category != searchcore.CategoryGeneral && category != searchcore.CategoryNews {
		return nil, errors.New("TinyFish Search category is unsupported")
	}
	query, err = searchcore.ValidateQuery(query)
	if err != nil {
		return nil, err
	}

	requestURL := *p.endpoint
	values := requestURL.Query()
	values.Set("query", query)
	if category == searchcore.CategoryNews {
		values.Set("domain_type", "news")
	} else {
		values.Set("domain_type", "web")
	}
	requestURL.RawQuery = values.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, errors.New("create TinyFish Search request")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "GoreeCloud-Search-Native/1 tinyfish-search-v1")
	request.Header.Set("X-API-Key", p.apiKey)

	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("TinyFish Search request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TinyFish Search returned HTTP %d", response.StatusCode)
	}
	if !isJSONContentType(response.Header.Get("Content-Type")) {
		return nil, errors.New("TinyFish Search returned unsupported content type")
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, MaxProviderResponseBytes+1))
	if err != nil {
		return nil, errors.New("read TinyFish Search response")
	}
	if len(body) > MaxProviderResponseBytes {
		return nil, errors.New("TinyFish Search response exceeds maximum size")
	}

	var decoded tinyFishSearchResponse
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&decoded); err != nil {
		return nil, errors.New("TinyFish Search response is invalid JSON")
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return nil, err
	}

	limit := searchcore.MaxResultsPerProvider + 1
	if len(decoded.Results) > limit {
		decoded.Results = decoded.Results[:limit]
	}
	results := make([]searchcore.Result, 0, len(decoded.Results))
	for _, item := range decoded.Results {
		if len(item.URL) > maxProviderURLBytes {
			continue
		}
		results = append(results, searchcore.Result{
			Title:   boundRunes(item.Title, maxProviderTitleRunes),
			URL:     strings.TrimSpace(item.URL),
			Snippet: boundRunes(item.Snippet, maxProviderSnippetRunes),
			Score:   tinyFishPositionScore(item.Position),
		})
	}
	return results, nil
}

func validateTinyFishSearchEndpoint(raw string) (*url.URL, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = TinyFishSearchEndpoint
	}
	endpoint, err := validateProviderEndpoint(value)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(endpoint.Hostname()) != "api.search.tinyfish.ai" || (endpoint.Path != "" && endpoint.Path != "/") {
		return nil, errors.New("TinyFish Search endpoint must be the official api.search.tinyfish.ai endpoint")
	}
	canonical, err := url.Parse(TinyFishSearchEndpoint)
	if err != nil {
		return nil, errors.New("TinyFish Search endpoint is invalid")
	}
	return canonical, nil
}

func tinyFishPositionScore(position int) int {
	if position <= 0 {
		return 0
	}
	score := maxTinyFishProviderScore + 1 - position
	if score < 1 {
		return 1
	}
	if score > maxTinyFishProviderScore {
		return maxTinyFishProviderScore
	}
	return score
}
