package providers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

const (
	ProviderContractVersion = 1
	MaxProviderResponseBytes = 4 << 20
	maxProviderTitleRunes    = 2048
	maxProviderSnippetRunes  = 8192
	maxProviderURLBytes      = 8192
)

var blockedProviderPrefixes = mustProviderPrefixes(
	"0.0.0.0/8",
	"10.0.0.0/8",
	"100.64.0.0/10",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"192.168.0.0/16",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"::/128",
	"::1/128",
	"fc00::/7",
	"fe80::/10",
	"ff00::/8",
	"2001:db8::/32",
)

type HTTPJSONConfig struct {
	Name                     string
	Endpoint                 string
	Categories               []string
	BearerToken              string
	PublishedAtAuthoritative bool
}

type HTTPJSONProvider struct {
	name                     string
	endpoint                 *url.URL
	categories               []string
	bearerToken              string
	publishedAtAuthoritative bool
	client                   *http.Client
}

type providerRequest struct {
	SchemaVersion int    `json:"schema_version"`
	Query         string `json:"query"`
	Category      string `json:"category"`
}

type providerResponse struct {
	SchemaVersion int          `json:"schema_version"`
	Results       []wireResult `json:"results"`
}

type wireResult struct {
	Title       string            `json:"title"`
	URL         string            `json:"url"`
	Snippet     string            `json:"snippet,omitempty"`
	Score       int               `json:"score,omitempty"`
	PublishedAt *time.Time        `json:"published_at,omitempty"`
	Media       *searchcore.Media `json:"media,omitempty"`
}

func NewHTTPJSONProvider(config HTTPJSONConfig) (*HTTPJSONProvider, error) {
	endpoint, err := validateProviderEndpoint(config.Endpoint)
	if err != nil {
		return nil, err
	}
	categories, err := validateProviderCategories(config.Categories)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(config.Name)
	if name == "" || len([]rune(name)) > searchcore.MaxProviderNameRunes {
		return nil, errors.New("provider name is invalid")
	}
	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return nil, errors.New("provider name contains a control character")
		}
	}

	provider := &HTTPJSONProvider{
		name:                     name,
		endpoint:                 endpoint,
		categories:               categories,
		bearerToken:              strings.TrimSpace(config.BearerToken),
		publishedAtAuthoritative: config.PublishedAtAuthoritative,
	}
	provider.client = secureProviderClient(endpoint)
	return provider, nil
}

func newHTTPJSONProviderWithClient(config HTTPJSONConfig, client *http.Client) (*HTTPJSONProvider, error) {
	provider, err := NewHTTPJSONProvider(config)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("provider client is required")
	}
	provider.client = client
	return provider, nil
}

func (p *HTTPJSONProvider) Name() string { return p.name }

func (p *HTTPJSONProvider) Categories() []string {
	return append([]string(nil), p.categories...)
}

func (p *HTTPJSONProvider) PublishedAtAuthoritative() bool {
	return p.publishedAtAuthoritative
}

func (p *HTTPJSONProvider) Search(ctx context.Context, query string) ([]searchcore.Result, error) {
	category := searchcore.CategoryGeneral
	if !containsCategory(p.categories, category) {
		if len(p.categories) != 1 {
			return nil, errors.New("provider requires an explicit search category")
		}
		category = p.categories[0]
	}
	return p.SearchCategory(ctx, query, category)
}

func (p *HTTPJSONProvider) SearchCategory(ctx context.Context, query, rawCategory string) ([]searchcore.Result, error) {
	category, err := searchcore.ValidateCategory(rawCategory)
	if err != nil || !containsCategory(p.categories, category) {
		return nil, errors.New("provider does not support requested category")
	}
	query, err = searchcore.ValidateQuery(query)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(providerRequest{
		SchemaVersion: ProviderContractVersion,
		Query:         query,
		Category:      category,
	})
	if err != nil {
		return nil, errors.New("encode provider request")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint.String(), bytes.NewReader(payload))
	if err != nil {
		return nil, errors.New("create provider request")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "GoreeCloud-Search-Native/1 provider-contract/1")
	if p.bearerToken != "" {
		request.Header.Set("Authorization", "Bearer "+p.bearerToken)
	}

	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("provider request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider returned HTTP %d", response.StatusCode)
	}
	if !isJSONContentType(response.Header.Get("Content-Type")) {
		return nil, errors.New("provider returned unsupported content type")
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, MaxProviderResponseBytes+1))
	if err != nil {
		return nil, errors.New("read provider response")
	}
	if len(body) > MaxProviderResponseBytes {
		return nil, errors.New("provider response exceeds maximum size")
	}

	var decoded providerResponse
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return nil, errors.New("provider response is invalid JSON")
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return nil, err
	}
	if decoded.SchemaVersion != ProviderContractVersion {
		return nil, errors.New("provider response schema version is unsupported")
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
			Title:       boundRunes(item.Title, maxProviderTitleRunes),
			URL:         strings.TrimSpace(item.URL),
			Snippet:     boundRunes(item.Snippet, maxProviderSnippetRunes),
			Score:       item.Score,
			PublishedAt: item.PublishedAt,
			Media:       item.Media,
		})
	}
	return results, nil
}

func validateProviderEndpoint(raw string) (*url.URL, error) {
	endpoint, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || endpoint.Scheme != "https" || endpoint.Host == "" || endpoint.User != nil {
		return nil, errors.New("provider endpoint must be an HTTPS URL without credentials")
	}
	if endpoint.Fragment != "" || endpoint.RawQuery != "" {
		return nil, errors.New("provider endpoint must not contain a query or fragment")
	}
	if port := endpoint.Port(); port != "" && port != "443" {
		return nil, errors.New("provider endpoint uses an unsupported port")
	}
	host := strings.TrimSuffix(strings.ToLower(endpoint.Hostname()), ".")
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") {
		return nil, errors.New("provider endpoint host is not allowed")
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		if !providerAddressAllowed(ip) {
			return nil, errors.New("provider endpoint address is not public")
		}
	} else if !strings.Contains(host, ".") {
		return nil, errors.New("provider endpoint must use a fully qualified host")
	}
	return endpoint, nil
}

func validateProviderCategories(raw []string) ([]string, error) {
	if len(raw) == 0 {
		return nil, errors.New("provider must declare at least one category")
	}
	seen := map[string]bool{}
	categories := make([]string, 0, len(raw))
	for _, value := range raw {
		category, err := searchcore.ValidateCategory(value)
		if err != nil {
			return nil, errors.New("provider declares an unsupported category")
		}
		if !seen[category] {
			seen[category] = true
			categories = append(categories, category)
		}
	}
	return categories, nil
}

func secureProviderClient(endpoint *url.URL) *http.Client {
	host := strings.TrimSuffix(strings.ToLower(endpoint.Hostname()), ".")
	transport := &http.Transport{
		Proxy:               nil,
		ForceAttemptHTTP2:   true,
		DisableCompression:  false,
		MaxIdleConns:        16,
		MaxIdleConnsPerHost: 4,
		IdleConnTimeout:     30 * time.Second,
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
	}
	transport.DialContext = guardedProviderDialContext(host)
	return &http.Client{
		Transport: transport,
		Timeout:   7 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return errors.New("provider redirects are not allowed")
		},
	}
}

func guardedProviderDialContext(expectedHost string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, errors.New("provider dial address is invalid")
		}
		normalizedHost := strings.TrimSuffix(strings.ToLower(host), ".")
		if normalizedHost != expectedHost {
			return nil, errors.New("provider dial host changed unexpectedly")
		}

		addresses, err := resolveProviderAddresses(ctx, normalizedHost)
		if err != nil {
			return nil, err
		}
		dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
		var lastErr error
		for _, ip := range addresses {
			connection, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if dialErr == nil {
				return connection, nil
			}
			lastErr = dialErr
		}
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, errors.New("provider host has no usable addresses")
	}
}

func resolveProviderAddresses(ctx context.Context, host string) ([]netip.Addr, error) {
	if ip, err := netip.ParseAddr(host); err == nil {
		if !providerAddressAllowed(ip) {
			return nil, errors.New("provider address is not public")
		}
		return []netip.Addr{ip.Unmap()}, nil
	}
	addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil || len(addresses) == 0 {
		return nil, errors.New("provider host could not be resolved")
	}
	allowed := make([]netip.Addr, 0, len(addresses))
	for _, address := range addresses {
		address = address.Unmap()
		if !providerAddressAllowed(address) {
			return nil, errors.New("provider host resolved to a non-public address")
		}
		allowed = append(allowed, address)
	}
	return allowed, nil
}

func providerAddressAllowed(address netip.Addr) bool {
	address = address.Unmap()
	if !address.IsValid() || address.IsUnspecified() || address.IsLoopback() || address.IsPrivate() ||
		address.IsLinkLocalUnicast() || address.IsLinkLocalMulticast() || address.IsMulticast() {
		return false
	}
	for _, prefix := range blockedProviderPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}

func mustProviderPrefixes(values ...string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefixes = append(prefixes, netip.MustParsePrefix(value))
	}
	return prefixes
}

func isJSONContentType(raw string) bool {
	mediaType, _, err := mime.ParseMediaType(raw)
	if err != nil {
		return false
	}
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("provider response contains trailing data")
	}
	return nil
}

func containsCategory(categories []string, category string) bool {
	for _, value := range categories {
		if value == category {
			return true
		}
	}
	return false
}

func boundRunes(raw string, maximum int) string {
	value := strings.TrimSpace(raw)
	runes := []rune(value)
	if len(runes) > maximum {
		return string(runes[:maximum])
	}
	return value
}
