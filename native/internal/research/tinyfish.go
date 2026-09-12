package research

import (
	"bufio"
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
	"unicode/utf8"
)

const (
	TinyFishResearchEndpoint        = "https://agent.tinyfish.ai/v1/automation/run-research"
	MaxTinyFishResearchStreamBytes  = 16 << 20
	MaxTinyFishResearchEventBytes   = 4 << 20
	maxTinyFishReportRunes           = 1_000_000
	maxTinyFishCitationCount         = 128
	maxTinyFishCitationTitleRunes    = 2048
	maxTinyFishCitationURLBytes      = 8192
	maxTinyFishTerminationRunes      = 128
	maxTinyFishResearchRunDuration   = 15 * time.Minute
	maxTinyFishResearchHeaderTimeout = 20 * time.Second
)

var (
	ErrInvalidRequest  = errors.New("research request is invalid")
	ErrUnavailable     = errors.New("research provider is unavailable")
	ErrInvalidResponse = errors.New("research provider response is invalid")
)

type TinyFishConfig struct {
	APIKey string
}

type TinyFishResearcher struct {
	apiKey string
	client *http.Client
}

type tinyFishRequest struct {
	Query          string `json:"query"`
	Mode           string `json:"mode"`
	Stream         bool   `json:"stream"`
	OutputLanguage string `json:"output_language,omitempty"`
}

func NewTinyFishResearcher(config TinyFishConfig) (*TinyFishResearcher, error) {
	apiKey := strings.TrimSpace(config.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("%w: TinyFish API key is required", ErrInvalidRequest)
	}
	return &TinyFishResearcher{apiKey: apiKey, client: secureTinyFishResearchClient()}, nil
}

func newTinyFishResearcherWithHTTPClient(config TinyFishConfig, client *http.Client) (*TinyFishResearcher, error) {
	researcher, err := NewTinyFishResearcher(config)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("%w: HTTP client is required", ErrInvalidRequest)
	}
	researcher.client = client
	return researcher, nil
}

func (r *TinyFishResearcher) Research(ctx context.Context, input Request) (Result, error) {
	request, err := normalizeRequest(input)
	if err != nil {
		return Result{}, err
	}

	providerContext, cancel := context.WithTimeout(ctx, maxTinyFishResearchRunDuration)
	defer cancel()

	payload, err := json.Marshal(tinyFishRequest{
		Query:          request.Query,
		Mode:           "standard",
		Stream:         true,
		OutputLanguage: request.OutputLanguage,
	})
	if err != nil {
		return Result{}, fmt.Errorf("%w: encode request", ErrInvalidRequest)
	}

	httpRequest, err := http.NewRequestWithContext(
		providerContext,
		http.MethodPost,
		TinyFishResearchEndpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return Result{}, fmt.Errorf("%w: create request", ErrInvalidRequest)
	}
	httpRequest.Header.Set("Accept", "text/event-stream")
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("User-Agent", "GoreeCloud-Search-Native/1 tinyfish-research-v1")
	httpRequest.Header.Set("X-API-Key", r.apiKey)

	response, err := r.client.Do(httpRequest)
	if err != nil {
		if errors.Is(providerContext.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			return Result{}, fmt.Errorf("%w: request timed out", ErrUnavailable)
		}
		if errors.Is(providerContext.Err(), context.Canceled) || errors.Is(err, context.Canceled) {
			return Result{}, context.Canceled
		}
		return Result{}, fmt.Errorf("%w: request failed", ErrUnavailable)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Result{}, fmt.Errorf("%w: HTTP %d", ErrUnavailable, response.StatusCode)
	}
	if !isEventStreamContentType(response.Header.Get("Content-Type")) {
		return Result{}, fmt.Errorf("%w: unsupported content type", ErrInvalidResponse)
	}

	return parseTinyFishEventStream(response.Body)
}

func normalizeRequest(input Request) (Request, error) {
	query := strings.TrimSpace(input.Query)
	purpose := strings.TrimSpace(input.Purpose)
	language := strings.TrimSpace(input.OutputLanguage)
	if query == "" || utf8.RuneCountInString(query) > MaxQueryRunes {
		return Request{}, fmt.Errorf("%w: query is required and must be at most %d characters", ErrInvalidRequest, MaxQueryRunes)
	}
	if purpose == "" || utf8.RuneCountInString(purpose) > MaxPurposeRunes {
		return Request{}, fmt.Errorf("%w: purpose is required and must be at most %d characters", ErrInvalidRequest, MaxPurposeRunes)
	}
	if utf8.RuneCountInString(language) > MaxOutputLanguageRunes || strings.ContainsAny(language, "\r\n\x00") {
		return Request{}, fmt.Errorf("%w: output language is invalid", ErrInvalidRequest)
	}
	return Request{Query: query, Purpose: purpose, OutputLanguage: language}, nil
}

func parseTinyFishEventStream(reader io.Reader) (Result, error) {
	limited := &io.LimitedReader{R: reader, N: MaxTinyFishResearchStreamBytes + 1}
	scanner := bufio.NewScanner(limited)
	scanner.Buffer(make([]byte, 64*1024), MaxTinyFishResearchEventBytes)

	var eventName string
	var dataLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if result, terminal, err := consumeTinyFishEvent(eventName, dataLines); terminal || err != nil {
				return result, err
			}
			eventName = ""
			dataLines = dataLines[:0]
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		field, value, found := strings.Cut(line, ":")
		if !found {
			field = line
			value = ""
		}
		value = strings.TrimPrefix(value, " ")
		switch field {
		case "event":
			eventName = strings.TrimSpace(value)
		case "data":
			dataLines = append(dataLines, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return Result{}, fmt.Errorf("%w: event stream could not be read", ErrInvalidResponse)
	}
	if limited.N <= 0 {
		return Result{}, fmt.Errorf("%w: event stream exceeds maximum size", ErrInvalidResponse)
	}
	if len(dataLines) > 0 || eventName != "" {
		if result, terminal, err := consumeTinyFishEvent(eventName, dataLines); terminal || err != nil {
			return result, err
		}
	}
	return Result{}, fmt.Errorf("%w: final_result event was not received", ErrInvalidResponse)
}

func consumeTinyFishEvent(eventName string, dataLines []string) (Result, bool, error) {
	if len(dataLines) == 0 {
		return Result{}, false, nil
	}
	data := strings.TrimSpace(strings.Join(dataLines, "\n"))
	if data == "" || data == "[DONE]" {
		return Result{}, false, nil
	}
	if len(data) > MaxTinyFishResearchEventBytes {
		return Result{}, true, fmt.Errorf("%w: event exceeds maximum size", ErrInvalidResponse)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return Result{}, true, fmt.Errorf("%w: event data is not valid JSON", ErrInvalidResponse)
	}

	kind := strings.TrimSpace(eventName)
	if kind == "" {
		kind = firstString(payload, "type", "event")
	}
	if kind != "final_result" {
		return Result{}, false, nil
	}

	result, err := normalizeFinalResult(payload)
	if err != nil {
		return Result{}, true, err
	}
	return result, true, nil
}

func normalizeFinalResult(payload map[string]any) (Result, error) {
	body := payload
	if nested, ok := payload["data"].(map[string]any); ok && firstString(payload, "report") == "" {
		body = nested
	}

	report := boundRunes(firstString(body, "report"), maxTinyFishReportRunes)
	if report == "" {
		return Result{}, fmt.Errorf("%w: final_result contains no report", ErrInvalidResponse)
	}
	termination := boundRunes(firstString(body, "termination_reason", "terminationReason"), maxTinyFishTerminationRunes)
	citations := normalizeCitations(body["citations"])
	return Result{Report: report, Citations: citations, TerminationReason: termination}, nil
}

func normalizeCitations(raw any) []Citation {
	items, ok := raw.([]any)
	if !ok || len(items) == 0 {
		return nil
	}
	result := make([]Citation, 0, min(len(items), maxTinyFishCitationCount))
	seen := map[string]bool{}
	for _, item := range items {
		if len(result) >= maxTinyFishCitationCount {
			break
		}
		var title, rawURL string
		switch value := item.(type) {
		case string:
			rawURL = value
		case map[string]any:
			title = firstString(value, "title", "source_title", "name")
			rawURL = firstString(value, "url", "source_url", "link")
		}
		normalizedURL, err := normalizeCitationURL(rawURL)
		if err != nil || seen[normalizedURL] {
			continue
		}
		seen[normalizedURL] = true
		result = append(result, Citation{
			Title: boundRunes(title, maxTinyFishCitationTitleRunes),
			URL:   normalizedURL,
		})
	}
	return result
}

func normalizeCitationURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" || len(value) > maxTinyFishCitationURLBytes {
		return "", errors.New("citation URL is invalid")
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", errors.New("citation URL is invalid")
	}
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || !strings.Contains(host, ".") {
		return "", errors.New("citation URL host is not allowed")
	}
	if ip, err := netip.ParseAddr(host); err == nil && !publicResearchAddress(ip) {
		return "", errors.New("citation URL address is not public")
	}
	parsed.Fragment = ""
	return parsed.String(), nil
}

func firstString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key].(string); ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func boundRunes(raw string, maximum int) string {
	value := strings.TrimSpace(raw)
	runes := []rune(value)
	if len(runes) > maximum {
		return string(runes[:maximum])
	}
	return value
}

func isEventStreamContentType(raw string) bool {
	mediaType, _, err := mime.ParseMediaType(raw)
	return err == nil && mediaType == "text/event-stream"
}

func secureTinyFishResearchClient() *http.Client {
	endpoint, _ := url.Parse(TinyFishResearchEndpoint)
	host := endpoint.Hostname()
	transport := &http.Transport{
		Proxy:                 nil,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          4,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       30 * time.Second,
		ResponseHeaderTimeout: maxTinyFishResearchHeaderTimeout,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
	}
	transport.DialContext = guardedTinyFishResearchDialContext(host)
	return &http.Client{
		Transport: transport,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return errors.New("TinyFish Research redirects are not allowed")
		},
	}
}

func guardedTinyFishResearchDialContext(expectedHost string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil || strings.TrimSuffix(strings.ToLower(host), ".") != strings.ToLower(expectedHost) {
			return nil, errors.New("TinyFish Research dial address is invalid")
		}
		addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", expectedHost)
		if err != nil || len(addresses) == 0 {
			return nil, errors.New("TinyFish Research host could not be resolved")
		}
		dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
		var lastErr error
		for _, candidate := range addresses {
			candidate = candidate.Unmap()
			if !publicResearchAddress(candidate) {
				return nil, errors.New("TinyFish Research host resolved to a non-public address")
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
		return nil, errors.New("TinyFish Research host has no usable address")
	}
}

func publicResearchAddress(address netip.Addr) bool {
	address = address.Unmap()
	return address.IsValid() && !address.IsUnspecified() && !address.IsLoopback() && !address.IsPrivate() &&
		!address.IsLinkLocalUnicast() && !address.IsLinkLocalMulticast() && !address.IsMulticast()
}
