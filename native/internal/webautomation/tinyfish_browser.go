package webautomation

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	TinyFishBrowserEndpoint         = "https://api.browser.tinyfish.ai"
	MaxTinyFishBrowserResponseBytes = 256 << 10
	tinyFishBrowserHTTPTimeout      = 75 * time.Second
	tinyFishBrowserHeaderTimeout    = 65 * time.Second
)

type TinyFishBrowser struct {
	apiKey string
	client *http.Client
}

type tinyFishBrowserStartRequest struct {
	URL            string `json:"url"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

type tinyFishBrowserStartResponse struct {
	SessionID string `json:"session_id"`
	CDPURL    string `json:"cdp_url"`
	BaseURL   string `json:"base_url"`
}

func NewTinyFishBrowser(apiKey string) (*TinyFishBrowser, error) {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return nil, fmt.Errorf("%w: TinyFish API key is required", ErrInvalidRequest)
	}
	return &TinyFishBrowser{apiKey: key, client: secureTinyFishBrowserClient()}, nil
}

func newTinyFishBrowserWithHTTPClient(apiKey string, client *http.Client) (*TinyFishBrowser, error) {
	browser, err := NewTinyFishBrowser(apiKey)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("%w: HTTP client is required", ErrInvalidRequest)
	}
	browser.client = client
	return browser, nil
}

func (b *TinyFishBrowser) StartBrowserSession(ctx context.Context, input StartBrowserSessionRequest) (BrowserSession, error) {
	request, err := normalizeStartBrowserSessionRequest(input)
	if err != nil {
		return BrowserSession{}, err
	}
	payload, err := json.Marshal(tinyFishBrowserStartRequest{
		URL:            request.TargetURL,
		TimeoutSeconds: durationSeconds(request.Timeout),
	})
	if err != nil {
		return BrowserSession{}, fmt.Errorf("%w: encode browser session request", ErrInvalidRequest)
	}
	responseBytes, err := b.doJSON(ctx, http.MethodPost, TinyFishBrowserEndpoint, payload, http.StatusCreated, false)
	if err != nil {
		return BrowserSession{}, err
	}
	var response tinyFishBrowserStartResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return BrowserSession{}, fmt.Errorf("%w: browser session response is invalid", ErrAutomationFailed)
	}
	sessionID, err := normalizeProviderReference(response.SessionID, "browser session ID")
	if err != nil {
		return BrowserSession{}, fmt.Errorf("%w: provider returned invalid browser session identity", ErrAutomationFailed)
	}
	cdpURL, err := validateSensitiveSetupURL(response.CDPURL, "wss")
	if err != nil {
		return BrowserSession{}, fmt.Errorf("%w: provider returned invalid browser CDP endpoint", ErrAutomationFailed)
	}
	baseURL, err := validateSensitiveSetupURL(response.BaseURL, "https")
	if err != nil {
		return BrowserSession{}, fmt.Errorf("%w: provider returned invalid browser base endpoint", ErrAutomationFailed)
	}
	return BrowserSession{
		SessionID: sessionID,
		CDPURL:    cdpURL,
		BaseURL:   baseURL,
		TargetURL: request.TargetURL,
		Timeout:   request.Timeout,
	}, nil
}

func (b *TinyFishBrowser) TerminateBrowserSession(ctx context.Context, rawSessionID string) error {
	sessionID, err := normalizeProviderReference(rawSessionID, "browser session ID")
	if err != nil {
		return err
	}
	endpoint := TinyFishBrowserEndpoint + "/" + url.PathEscape(sessionID)
	_, err = b.doJSON(ctx, http.MethodDelete, endpoint, nil, http.StatusNoContent, true)
	return err
}

func (b *TinyFishBrowser) doJSON(ctx context.Context, method, endpoint string, payload []byte, expectedStatus int, terminating bool) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		body = bytes.NewReader(payload)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("%w: create browser provider request", ErrAutomationFailed)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-API-Key", b.apiKey)
	request.Header.Set("User-Agent", "GoreeCloud-Search-Native/1 tinyfish-browser-v1")
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := b.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if terminating {
			return nil, ErrBrowserSessionTerminationUnknown
		}
		return nil, ErrBrowserSessionCreationUnknown
	}
	defer response.Body.Close()
	if response.StatusCode != expectedStatus {
		return nil, mapTinyFishBrowserHTTPError(response.StatusCode, terminating)
	}
	if expectedStatus == http.StatusNoContent {
		return nil, nil
	}
	if contentType := response.Header.Get("Content-Type"); contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		return nil, fmt.Errorf("%w: browser provider returned unsupported content type", ErrAutomationFailed)
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, MaxTinyFishBrowserResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: browser provider response could not be read", ErrAutomationFailed)
	}
	if len(content) > MaxTinyFishBrowserResponseBytes {
		return nil, fmt.Errorf("%w: browser provider response exceeds maximum size", ErrAutomationFailed)
	}
	if len(bytes.TrimSpace(content)) == 0 {
		return nil, fmt.Errorf("%w: browser provider returned an empty response", ErrAutomationFailed)
	}
	return content, nil
}

func mapTinyFishBrowserHTTPError(status int, terminating bool) error {
	if terminating {
		switch status {
		case http.StatusBadRequest:
			return ErrInvalidRequest
		case http.StatusUnauthorized, http.StatusNotFound:
			return ErrBrowserSessionUnavailable
		case http.StatusConflict, http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return ErrBrowserSessionTerminationUnknown
		default:
			return ErrAutomationFailed
		}
	}
	switch status {
	case http.StatusBadRequest:
		return ErrInvalidRequest
	case http.StatusUnauthorized:
		return ErrBrowserSessionUnavailable
	case http.StatusPaymentRequired:
		return ErrBillingRejected
	case http.StatusNotFound:
		return ErrCapabilityUnavailable
	case http.StatusConflict, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return ErrBrowserSessionCreationUnknown
	case http.StatusTooManyRequests:
		return ErrAutomationLimit
	default:
		return ErrAutomationFailed
	}
}

func secureTinyFishBrowserClient() *http.Client {
	endpoint, _ := url.Parse(TinyFishBrowserEndpoint)
	host := endpoint.Hostname()
	transport := &http.Transport{
		Proxy:                 nil,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          4,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       30 * time.Second,
		ResponseHeaderTimeout: tinyFishBrowserHeaderTimeout,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
	}
	transport.DialContext = guardedTinyFishBrowserDialContext(host)
	return &http.Client{
		Transport: transport,
		Timeout:   tinyFishBrowserHTTPTimeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return errors.New("TinyFish Browser redirects are not allowed")
		},
	}
}

func guardedTinyFishBrowserDialContext(expectedHost string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil || strings.TrimSuffix(strings.ToLower(host), ".") != strings.ToLower(expectedHost) {
			return nil, errors.New("TinyFish Browser dial address is invalid")
		}
		addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", expectedHost)
		if err != nil || len(addresses) == 0 {
			return nil, errors.New("TinyFish Browser host could not be resolved")
		}
		dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
		var lastErr error
		for _, candidate := range addresses {
			candidate = candidate.Unmap()
			if !publicAgentAddress(candidate) {
				return nil, errors.New("TinyFish Browser host resolved to a non-public address")
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
		return nil, errors.New("TinyFish Browser host has no usable address")
	}
}
