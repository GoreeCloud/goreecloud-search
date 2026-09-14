package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxPrivacyReferenceVerificationResponseBytes = 16 * 1024

var (
	errPrivacyReferenceHTTPEndpointInvalid       = errors.New("Privacy Shield verification endpoint is invalid")
	errPrivacyReferenceHTTPTransportUnauthenticated = errors.New("Privacy Shield verification transport is not authenticated as GoreeCloud Search")
	errPrivacyReferenceHTTPConsumerMismatch      = errors.New("Privacy Shield verification request consumer does not match authenticated transport identity")
	errPrivacyReferenceHTTPResponseInvalid       = errors.New("Privacy Shield verification response is invalid")
)

// authenticatedPrivacyShieldRoundTripper is the narrow seam where GoreeCloud
// Identity-backed service authentication must eventually enter the Search
// runtime. Merely implementing HTTP transport is not enough: a production
// implementation must authenticate as goreecloud-search and report that exact
// identity here. This client is not wired by default and does not manufacture
// or accept an arbitrary requester/consumer header as authority.
type authenticatedPrivacyShieldRoundTripper interface {
	http.RoundTripper
	AuthenticatedConsumerID() string
}

type privacyReferenceHTTPClient struct {
	endpoint   *url.URL
	client     *http.Client
	consumerID string
}

func newPrivacyReferenceHTTPClient(
	endpoint string,
	transport authenticatedPrivacyShieldRoundTripper,
) (*privacyReferenceHTTPClient, error) {
	parsedEndpoint, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || !validPrivacyReferenceHTTPEndpoint(parsedEndpoint) {
		return nil, errPrivacyReferenceHTTPEndpointInvalid
	}
	if transport == nil {
		return nil, errPrivacyReferenceHTTPTransportUnauthenticated
	}
	consumerID := strings.TrimSpace(transport.AuthenticatedConsumerID())
	if consumerID != searchPrivacyVerificationConsumerID {
		return nil, errPrivacyReferenceHTTPTransportUnauthenticated
	}

	return &privacyReferenceHTTPClient{
		endpoint: parsedEndpoint,
		consumerID: consumerID,
		client: &http.Client{
			Transport: transport,
			Timeout:   5 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

func validPrivacyReferenceHTTPEndpoint(endpoint *url.URL) bool {
	if endpoint == nil || endpoint.Host == "" || endpoint.Path == "" {
		return false
	}
	if endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return false
	}
	if endpoint.Scheme == "https" {
		return true
	}
	if endpoint.Scheme != "http" {
		return false
	}

	hostname := endpoint.Hostname()
	if strings.EqualFold(hostname, "localhost") {
		return true
	}
	ip := net.ParseIP(hostname)
	return ip != nil && ip.IsLoopback()
}

func (c *privacyReferenceHTTPClient) VerifyReference(
	ctx context.Context,
	request privacyReferenceVerificationRequest,
) (privacyReferenceVerificationResponse, error) {
	if c == nil || c.client == nil || c.endpoint == nil {
		return privacyReferenceVerificationResponse{}, errPrivacyReferenceVerificationClientUnavailable
	}
	if request.ConsumerID != c.consumerID {
		return privacyReferenceVerificationResponse{}, errPrivacyReferenceHTTPConsumerMismatch
	}

	body, err := json.Marshal(request)
	if err != nil {
		return privacyReferenceVerificationResponse{}, err
	}
	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.endpoint.String(),
		bytes.NewReader(body),
	)
	if err != nil {
		return privacyReferenceVerificationResponse{}, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("Cache-Control", "no-store")

	response, err := c.client.Do(httpRequest)
	if err != nil {
		return privacyReferenceVerificationResponse{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return privacyReferenceVerificationResponse{}, errPrivacyReferenceHTTPResponseInvalid
	}
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return privacyReferenceVerificationResponse{}, errPrivacyReferenceHTTPResponseInvalid
	}

	encoded, err := io.ReadAll(io.LimitReader(
		response.Body,
		maxPrivacyReferenceVerificationResponseBytes+1,
	))
	if err != nil || len(encoded) > maxPrivacyReferenceVerificationResponseBytes {
		return privacyReferenceVerificationResponse{}, errPrivacyReferenceHTTPResponseInvalid
	}

	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	var decoded privacyReferenceVerificationResponse
	if err := decoder.Decode(&decoded); err != nil {
		return privacyReferenceVerificationResponse{}, errPrivacyReferenceHTTPResponseInvalid
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return privacyReferenceVerificationResponse{}, errPrivacyReferenceHTTPResponseInvalid
	}
	return decoded, nil
}
