package webautomation

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestTinyFishBrowserMapsExactCreateRequestAndResponse(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost || request.URL.String() != TinyFishBrowserEndpoint {
			t.Fatalf("unexpected create request: %s %s", request.Method, request.URL)
		}
		if request.Header.Get("X-API-Key") != "browser-secret" {
			t.Fatal("missing TinyFish Browser API key")
		}
		if request.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("content type = %q", request.Header.Get("Content-Type"))
		}
		var payload map[string]any
		decoder := json.NewDecoder(request.Body)
		if err := decoder.Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(payload) != 2 {
			t.Fatalf("create payload must contain only documented fields: %#v", payload)
		}
		if payload["url"] != "https://example.com/path" {
			t.Fatalf("url = %#v", payload["url"])
		}
		if payload["timeout_seconds"] != float64(120) {
			t.Fatalf("timeout_seconds = %#v", payload["timeout_seconds"])
		}
		for _, forbidden := range []string{"purpose", "browser_profile", "use_profile", "profile_id", "use_vault", "credential_item_ids", "proxy_config"} {
			if _, exists := payload[forbidden]; exists {
				t.Fatalf("undocumented field %q crossed provider boundary: %#v", forbidden, payload)
			}
		}
		return &http.Response{
			StatusCode: http.StatusCreated,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
  "session_id": "browser_session_1",
  "cdp_url": "wss://browser.example.test/cdp",
  "base_url": "https://browser.example.test"
}`)),
		}, nil
	})}
	browser, err := newTinyFishBrowserWithHTTPClient("browser-secret", client)
	if err != nil {
		t.Fatal(err)
	}
	session, err := browser.StartBrowserSession(context.Background(), StartBrowserSessionRequest{
		TargetURL: "https://example.com/path#fragment",
		Purpose:   "local-only authorization purpose",
		Timeout:   2 * time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	if session.SessionID != "browser_session_1" || session.CDPURL != "wss://browser.example.test/cdp" || session.BaseURL != "https://browser.example.test" {
		t.Fatalf("unexpected session: %#v", session)
	}
	if session.TargetURL != "https://example.com/path" || session.Timeout != 2*time.Minute {
		t.Fatalf("unexpected normalized session: %#v", session)
	}
}

func TestTinyFishBrowserMapsExactTerminateRequest(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodDelete || request.URL.String() != TinyFishBrowserEndpoint+"/browser_session_1" {
			t.Fatalf("unexpected terminate request: %s %s", request.Method, request.URL)
		}
		if request.Header.Get("X-API-Key") != "browser-secret" {
			t.Fatal("missing TinyFish Browser API key")
		}
		return &http.Response{StatusCode: http.StatusNoContent, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	browser, err := newTinyFishBrowserWithHTTPClient("browser-secret", client)
	if err != nil {
		t.Fatal(err)
	}
	if err := browser.TerminateBrowserSession(context.Background(), "browser_session_1"); err != nil {
		t.Fatal(err)
	}
}

func TestTinyFishBrowserCreateHTTPErrorMapping(t *testing.T) {
	tests := []struct {
		status int
		want   error
	}{
		{http.StatusBadRequest, ErrInvalidRequest},
		{http.StatusPaymentRequired, ErrBillingRejected},
		{http.StatusNotFound, ErrCapabilityUnavailable},
		{http.StatusConflict, ErrBrowserSessionCreationUnknown},
		{http.StatusTooManyRequests, ErrAutomationLimit},
		{http.StatusServiceUnavailable, ErrBrowserSessionCreationUnknown},
		{http.StatusGatewayTimeout, ErrBrowserSessionCreationUnknown},
	}
	for _, test := range tests {
		t.Run(http.StatusText(test.status), func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: test.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"error":"redacted"}`))}, nil
			})}
			browser, err := newTinyFishBrowserWithHTTPClient("browser-secret", client)
			if err != nil {
				t.Fatal(err)
			}
			_, err = browser.StartBrowserSession(context.Background(), StartBrowserSessionRequest{TargetURL: "https://example.com", Purpose: "test browser provider mapping"})
			if !errors.Is(err, test.want) {
				t.Fatalf("status %d: want %v, got %v", test.status, test.want, err)
			}
		})
	}
}

func TestTinyFishBrowserTerminateUnconfirmedMapping(t *testing.T) {
	for _, status := range []int{http.StatusConflict, http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusServiceUnavailable, http.StatusGatewayTimeout} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}, nil
			})}
			browser, err := newTinyFishBrowserWithHTTPClient("browser-secret", client)
			if err != nil {
				t.Fatal(err)
			}
			if err := browser.TerminateBrowserSession(context.Background(), "browser_session_1"); !errors.Is(err, ErrBrowserSessionTerminationUnknown) {
				t.Fatalf("status %d: expected unconfirmed termination, got %v", status, err)
			}
		})
	}
}

func TestTinyFishBrowserTransportErrorsPreserveUncertainOutcome(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network interrupted after request write")
	})}
	browser, err := newTinyFishBrowserWithHTTPClient("browser-secret", client)
	if err != nil {
		t.Fatal(err)
	}
	_, err = browser.StartBrowserSession(context.Background(), StartBrowserSessionRequest{TargetURL: "https://example.com", Purpose: "test uncertain create"})
	if !errors.Is(err, ErrBrowserSessionCreationUnknown) {
		t.Fatalf("expected uncertain creation, got %v", err)
	}
	if err := browser.TerminateBrowserSession(context.Background(), "browser_session_1"); !errors.Is(err, ErrBrowserSessionTerminationUnknown) {
		t.Fatalf("expected uncertain termination, got %v", err)
	}
}

func TestTinyFishBrowserRejectsInvalidProviderEndpoints(t *testing.T) {
	for name, response := range map[string]string{
		"invalid cdp scheme": `{"session_id":"browser_1","cdp_url":"https://browser.example.test/cdp","base_url":"https://browser.example.test"}`,
		"invalid base scheme": `{"session_id":"browser_1","cdp_url":"wss://browser.example.test/cdp","base_url":"http://browser.example.test"}`,
	} {
		t.Run(name, func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusCreated, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(response))}, nil
			})}
			browser, err := newTinyFishBrowserWithHTTPClient("browser-secret", client)
			if err != nil {
				t.Fatal(err)
			}
			_, err = browser.StartBrowserSession(context.Background(), StartBrowserSessionRequest{TargetURL: "https://example.com", Purpose: "validate provider endpoints"})
			if !errors.Is(err, ErrAutomationFailed) {
				t.Fatalf("expected provider validation failure, got %v", err)
			}
		})
	}
}
