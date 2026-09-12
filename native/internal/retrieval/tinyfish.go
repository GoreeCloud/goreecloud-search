package retrieval

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
)

const (
	TinyFishFetchEndpoint         = "https://api.fetch.tinyfish.ai"
	MaxTinyFishFetchResponseBytes = 8 << 20
	maxTinyFishTitleRunes         = 2048
	maxTinyFishContentRunes       = 500000
	maxTinyFishURLBytes           = 8192
)

const (
	FailureFetchFailed = "fetch_failed"
)

type TinyFishConfig struct {
	APIKey string
}

type TinyFishFetcher struct {
	apiKey string
	client *http.Client
}

type tinyFishRequest struct {
	URLs   []string `json:"urls"`
	Format string   `json:"format"`
}

type tinyFishResponse struct {
	Results []tinyFishResult `json:"results"`
	Errors  []tinyFishError  `json:"errors"`
}

type tinyFishResult struct {
	URL      string `json:"url"`
	FinalURL string `json:"final_url"`
	Title    string `json:"title"`
	Text     string `json:"text"`
}

type tinyFishError struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

func NewTinyFishFetcher(config TinyFishConfig) (*TinyFishFetcher, error) {
	apiKey := strings.TrimSpace(config.APIKey)
	if apiKey == "" {
		return nil, errors.New("TinyFish Fetch API key is required")
	}
	return &TinyFishFetcher{apiKey: apiKey, client: secureTinyFishClient()}, nil
}

func newTinyFishFetcherWithHTTPClient(config TinyFishConfig, client *http.Client) (*TinyFishFetcher, error) {
	fetcher, err := NewTinyFishFetcher(config)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("TinyFish Fetch HTTP client is required")
	}
	fetcher.client = client
	return fetcher, nil
}

func (f *TinyFishFetcher) Fetch(ctx context.Context, rawURLs []string) (Response, error) {
	if len(rawURLs) == 0 {
		return Response{}, errors.New("at least one retrieval URL is required")
	}
	if len(rawURLs) > MaxURLsPerRequest {
		return Response{}, fmt.Errorf("retrieval request exceeds %d URLs", MaxURLsPerRequest)
	}

	urls := make([]string, 0, len(rawURLs))
	seen := map[string]bool{}
	for _, raw := range rawURLs {
		normalized, err := normalizeRetrievalURL(raw)
		if err != nil {
			return Response{}, err
		}
		if !seen[normalized] {
			seen[normalized] = true
			urls = append(urls, normalized)
		}
	}
	if len(urls) == 0 {
		return Response{}, errors.New("retrieval request contains no usable URLs")
	}

	payload, err := json.Marshal(tinyFishRequest{URLs: urls, Format: "markdown"})
	if err != nil {
		return Response{}, errors.New("encode TinyFish Fetch request")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, TinyFishFetchEndpoint, bytes.NewReader(payload))
	if err != nil {
		return Response{}, errors.New("create TinyFish Fetch request")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "GoreeCloud-Search-Native/1 tinyfish-fetch-v1")
	request.Header.Set("X-API-Key", f.apiKey)

	response, err := f.client.Do(request)
	if err != nil {
		return Response{}, fmt.Errorf("TinyFish Fetch request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("TinyFish Fetch returned HTTP %d", response.StatusCode)
	}
	if !isJSONContentType(response.Header.Get("Content-Type")) {
		return Response{}, errors.New("TinyFish Fetch returned unsupported content type")
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, MaxTinyFishFetchResponseBytes+1))
	if err != nil {
		return Response{}, errors.New("read TinyFish Fetch response")
	}
	if len(body) > MaxTinyFishFetchResponseBytes {
		return Response{}, errors.New("TinyFish Fetch response exceeds maximum size")
	}

	var decoded tinyFishResponse
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&decoded); err != nil {
		return Response{}, errors.New("TinyFish Fetch response is invalid JSON")
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return Response{}, err
	}

	result := Response{Documents: []Document{}}
	for _, item := range decoded.Results {
		if len(item.URL) > maxTinyFishURLBytes || len(item.FinalURL) > maxTinyFishURLBytes {
			continue
		}
		requested, err := normalizeRetrievalURL(item.URL)
		if err != nil {
			continue
		}
		finalURL := ""
		if strings.TrimSpace(item.FinalURL) != "" {
			finalURL, err = normalizeRetrievalURL(item.FinalURL)
			if err != nil {
				continue
			}
		}
		result.Documents = append(result.Documents, Document{
			URL:      requested,
			FinalURL: finalURL,
			Title:    boundRunes(item.Title, maxTinyFishTitleRunes),
			Content:  boundRunes(item.Text, maxTinyFishContentRunes),
		})
		if len(result.Documents) >= len(urls) {
			break
		}
	}
	for _, item := range decoded.Errors {
		if len(item.URL) > maxTinyFishURLBytes {
			continue
		}
		normalized, err := normalizeRetrievalURL(item.URL)
		if err != nil {
			continue
		}
		result.Failures = append(result.Failures, Failure{URL: normalized, Code: FailureFetchFailed})
		if len(result.Failures) >= len(urls) {
			break
		}
	}
	return result, nil
}

func normalizeRetrievalURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" || len(value) > maxTinyFishURLBytes {
		return "", errors.New("retrieval URL is invalid")
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", errors.New("retrieval URL must be HTTP or HTTPS without embedded credentials")
	}
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || !strings.Contains(host, ".") {
		return "", errors.New("retrieval URL host is not allowed")
	}
	if ip, err := netip.ParseAddr(host); err == nil && !publicAddress(ip) {
		return "", errors.New("retrieval URL address is not public")
	}
	parsed.Fragment = ""
	return parsed.String(), nil
}

func secureTinyFishClient() *http.Client {
	endpoint, _ := url.Parse(TinyFishFetchEndpoint)
	host := endpoint.Hostname()
	transport := &http.Transport{
		Proxy:               nil,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        8,
		MaxIdleConnsPerHost: 4,
		IdleConnTimeout:     30 * time.Second,
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
	}
	transport.DialContext = guardedTinyFishDialContext(host)
	return &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return errors.New("TinyFish Fetch redirects are not allowed")
		},
	}
}

func guardedTinyFishDialContext(expectedHost string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil || strings.TrimSuffix(strings.ToLower(host), ".") != strings.ToLower(expectedHost) {
			return nil, errors.New("TinyFish Fetch dial address is invalid")
		}
		addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", expectedHost)
		if err != nil || len(addresses) == 0 {
			return nil, errors.New("TinyFish Fetch host could not be resolved")
		}
		dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
		var lastErr error
		for _, candidate := range addresses {
			candidate = candidate.Unmap()
			if !publicAddress(candidate) {
				return nil, errors.New("TinyFish Fetch host resolved to a non-public address")
			}
			connection, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(candidate.String(), port))
			if dialErr == nil {
				return connection, nil
			}
			lastErr = dialErr
		}
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, errors.New("TinyFish Fetch host has no usable address")
	}
}

func publicAddress(address netip.Addr) bool {
	address = address.Unmap()
	return address.IsValid() && !address.IsUnspecified() && !address.IsLoopback() && !address.IsPrivate() &&
		!address.IsLinkLocalUnicast() && !address.IsLinkLocalMulticast() && !address.IsMulticast()
}

func isJSONContentType(raw string) bool {
	mediaType, _, err := mime.ParseMediaType(raw)
	return err == nil && (mediaType == "application/json" || strings.HasSuffix(mediaType, "+json"))
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("TinyFish Fetch response contains trailing data")
	}
	return nil
}

func boundRunes(raw string, maximum int) string {
	value := strings.TrimSpace(raw)
	runes := []rune(value)
	if len(runes) > maximum {
		return string(runes[:maximum])
	}
	return value
}
