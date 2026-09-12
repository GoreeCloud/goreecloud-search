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
	"net/netip"
	"net/url"
	"strings"
	"time"
)

const (
	TinyFishAgentRunAsyncEndpoint = "https://agent.tinyfish.ai/v1/automation/run-async"
	TinyFishAgentRunsBaseURL      = "https://agent.tinyfish.ai/v1/runs/"
	MaxTinyFishAgentResponseBytes = 2 << 20
	MaxTinyFishAgentOutputBytes   = 1 << 20
	defaultTinyFishPollInterval   = time.Second
	maxTinyFishHeaderTimeout      = 20 * time.Second
)

type TinyFishAgentConfig struct {
	APIKey       string
	PollInterval time.Duration
}

type TinyFishAgent struct {
	apiKey       string
	client       *http.Client
	pollInterval time.Duration
}

type tinyFishAgentConfig struct {
	MaxSteps           int `json:"max_steps,omitempty"`
	MaxDurationSeconds int `json:"max_duration_seconds"`
}

type tinyFishAgentRequest struct {
	URL               string              `json:"url"`
	Goal              string              `json:"goal"`
	BrowserProfile    string              `json:"browser_profile"`
	AgentConfig       tinyFishAgentConfig `json:"agent_config"`
	UseProfile        bool                `json:"use_profile,omitempty"`
	ProfileID         string              `json:"profile_id,omitempty"`
	UseVault          bool                `json:"use_vault,omitempty"`
	CredentialItemIDs []string            `json:"credential_item_ids,omitempty"`
}

type tinyFishRun struct {
	RunID  string          `json:"run_id"`
	Status string          `json:"status"`
	Result json.RawMessage `json:"result"`
	Error  json.RawMessage `json:"error"`
}

func NewTinyFishAgent(config TinyFishAgentConfig) (*TinyFishAgent, error) {
	apiKey := strings.TrimSpace(config.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("%w: TinyFish API key is required", ErrInvalidRequest)
	}
	pollInterval := config.PollInterval
	if pollInterval == 0 {
		pollInterval = defaultTinyFishPollInterval
	}
	if pollInterval < 10*time.Millisecond || pollInterval > 30*time.Second {
		return nil, fmt.Errorf("%w: TinyFish poll interval is invalid", ErrInvalidRequest)
	}
	return &TinyFishAgent{apiKey: apiKey, client: secureTinyFishAgentClient(), pollInterval: pollInterval}, nil
}

func newTinyFishAgentWithHTTPClient(config TinyFishAgentConfig, client *http.Client) (*TinyFishAgent, error) {
	agent, err := NewTinyFishAgent(config)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("%w: HTTP client is required", ErrInvalidRequest)
	}
	agent.client = client
	return agent, nil
}

func (a *TinyFishAgent) Run(ctx context.Context, input Request) (Result, error) {
	request, err := normalizeRequest(input)
	if err != nil {
		return Result{}, err
	}
	if request.Capability != CapabilityAgent {
		return Result{}, fmt.Errorf("%w: TinyFish Agent only implements agent capability", ErrCapabilityUnavailable)
	}

	payload, err := json.Marshal(tinyFishAgentRequest{
		URL:            request.URL,
		Goal:           request.Goal,
		BrowserProfile: string(request.BrowserProfile),
		AgentConfig: tinyFishAgentConfig{MaxSteps: request.MaxSteps, MaxDurationSeconds: durationSeconds(request.MaxDuration)},
		UseProfile:        request.UseProfile,
		ProfileID:         request.ProfileID,
		UseVault:          request.UseVault,
		CredentialItemIDs: request.CredentialItemIDs,
	})
	if err != nil {
		return Result{}, fmt.Errorf("%w: encode provider request", ErrInvalidRequest)
	}

	started, err := a.doJSON(ctx, http.MethodPost, TinyFishAgentRunAsyncEndpoint, payload)
	if err != nil {
		return Result{}, err
	}
	var run tinyFishRun
	if err := json.Unmarshal(started, &run); err != nil {
		return Result{}, fmt.Errorf("%w: provider start response is invalid", ErrAutomationFailed)
	}
	runID := strings.TrimSpace(run.RunID)
	if !validRunID(runID) {
		return Result{}, fmt.Errorf("%w: provider returned invalid run identity", ErrAutomationFailed)
	}

	runContext, cancel := context.WithTimeout(ctx, request.MaxDuration+15*time.Second)
	defer cancel()
	for {
		select {
		case <-runContext.Done():
			a.cancelRun(runID)
			if ctx.Err() != nil {
				return Result{}, ctx.Err()
			}
			return Result{}, fmt.Errorf("%w: provider run timed out", ErrAutomationFailed)
		case <-time.After(a.pollInterval):
		}

		stateBytes, stateErr := a.doJSON(runContext, http.MethodGet, TinyFishAgentRunsBaseURL+url.PathEscape(runID), nil)
		if stateErr != nil {
			if runContext.Err() != nil {
				a.cancelRun(runID)
				if ctx.Err() != nil {
					return Result{}, ctx.Err()
				}
				return Result{}, fmt.Errorf("%w: provider run timed out", ErrAutomationFailed)
			}
			return Result{}, stateErr
		}
		if err := json.Unmarshal(stateBytes, &run); err != nil {
			return Result{}, fmt.Errorf("%w: provider run response is invalid", ErrAutomationFailed)
		}
		if strings.TrimSpace(run.RunID) != "" && strings.TrimSpace(run.RunID) != runID {
			return Result{}, fmt.Errorf("%w: provider run identity changed", ErrAutomationFailed)
		}
		switch strings.ToUpper(strings.TrimSpace(run.Status)) {
		case "PENDING", "RUNNING":
			continue
		case "COMPLETED":
			output, normalizeErr := normalizeTinyFishOutput(run.Result)
			if normalizeErr != nil {
				return Result{}, normalizeErr
			}
			return Result{Capability: CapabilityAgent, Output: output}, nil
		case "FAILED", "CANCELLED":
			return Result{}, fmt.Errorf("%w: provider run ended with %s", ErrAutomationFailed, strings.ToLower(strings.TrimSpace(run.Status)))
		default:
			return Result{}, fmt.Errorf("%w: provider returned unknown run status", ErrAutomationFailed)
		}
	}
}

func (a *TinyFishAgent) doJSON(ctx context.Context, method, endpoint string, payload []byte) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		body = bytes.NewReader(payload)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("%w: create provider request", ErrAutomationFailed)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-API-Key", a.apiKey)
	request.Header.Set("User-Agent", "GoreeCloud-Search-Native/1 tinyfish-agent-v1")
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := a.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("%w: provider request failed", ErrAutomationFailed)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: provider returned HTTP %d", ErrAutomationFailed, response.StatusCode)
	}
	if contentType := response.Header.Get("Content-Type"); contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		return nil, fmt.Errorf("%w: provider returned unsupported content type", ErrAutomationFailed)
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, MaxTinyFishAgentResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: provider response could not be read", ErrAutomationFailed)
	}
	if len(content) > MaxTinyFishAgentResponseBytes {
		return nil, fmt.Errorf("%w: provider response exceeds maximum size", ErrAutomationFailed)
	}
	return content, nil
}

func (a *TinyFishAgent) cancelRun(runID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = a.doJSON(ctx, http.MethodPost, TinyFishAgentRunsBaseURL+url.PathEscape(runID)+"/cancel", []byte("{}"))
}

func normalizeTinyFishOutput(raw json.RawMessage) (string, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return "", nil
	}
	if len(trimmed) > MaxTinyFishAgentOutputBytes {
		return "", fmt.Errorf("%w: provider output exceeds maximum size", ErrAutomationFailed)
	}
	var text string
	if err := json.Unmarshal(trimmed, &text); err == nil {
		return text, nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	if err := decoder.Decode(&value); err != nil {
		return "", fmt.Errorf("%w: provider output is invalid JSON", ErrAutomationFailed)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return "", fmt.Errorf("%w: provider output contains trailing data", ErrAutomationFailed)
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, trimmed); err != nil {
		return "", fmt.Errorf("%w: provider output is invalid JSON", ErrAutomationFailed)
	}
	return compact.String(), nil
}

func durationSeconds(duration time.Duration) int {
	return int((duration + time.Second - 1) / time.Second)
}

func validRunID(value string) bool {
	if value == "" || len(value) > 256 {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' {
			continue
		}
		return false
	}
	return true
}

func secureTinyFishAgentClient() *http.Client {
	endpoint, _ := url.Parse(TinyFishAgentRunAsyncEndpoint)
	host := endpoint.Hostname()
	transport := &http.Transport{Proxy: nil, ForceAttemptHTTP2: true, MaxIdleConns: 4, MaxIdleConnsPerHost: 2, IdleConnTimeout: 30 * time.Second, ResponseHeaderTimeout: maxTinyFishHeaderTimeout, TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
	transport.DialContext = guardedTinyFishAgentDialContext(host)
	return &http.Client{Transport: transport, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return errors.New("TinyFish Agent redirects are not allowed") }}
}

func guardedTinyFishAgentDialContext(expectedHost string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil || strings.TrimSuffix(strings.ToLower(host), ".") != strings.ToLower(expectedHost) {
			return nil, errors.New("TinyFish Agent dial address is invalid")
		}
		addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", expectedHost)
		if err != nil || len(addresses) == 0 {
			return nil, errors.New("TinyFish Agent host could not be resolved")
		}
		dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
		var lastErr error
		for _, candidate := range addresses {
			candidate = candidate.Unmap()
			if !publicAgentAddress(candidate) {
				return nil, errors.New("TinyFish Agent host resolved to a non-public address")
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
		return nil, errors.New("TinyFish Agent host has no usable address")
	}
}

func publicAgentAddress(address netip.Addr) bool {
	address = address.Unmap()
	return address.IsValid() && !address.IsUnspecified() && !address.IsLoopback() && !address.IsPrivate() && !address.IsLinkLocalUnicast() && !address.IsLinkLocalMulticast() && !address.IsMulticast()
}
