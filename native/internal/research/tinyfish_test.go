package research

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func response(status int, contentType, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{contentType}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestTinyFishResearchRequestAndFinalResult(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost || request.URL.String() != TinyFishResearchEndpoint {
			t.Fatalf("unexpected request target: %s %s", request.Method, request.URL)
		}
		if request.Header.Get("X-API-Key") != "secret-key" {
			t.Fatal("TinyFish API key was not sent in X-API-Key")
		}
		if request.Header.Get("Accept") != "text/event-stream" {
			t.Fatalf("unexpected Accept header: %q", request.Header.Get("Accept"))
		}
		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["query"] != "compare storage systems" || payload["mode"] != "standard" || payload["stream"] != true {
			t.Fatalf("unexpected request payload: %#v", payload)
		}
		if payload["output_language"] != "en" {
			t.Fatalf("unexpected output language: %#v", payload["output_language"])
		}
		if _, found := payload["purpose"]; found {
			t.Fatal("local purpose binding must not be sent to TinyFish")
		}
		body := "event: research_progress\ndata: {\"message\":\"working\"}\n\n" +
			"event: final_result\ndata: {\"report\":\"Bounded report\",\"citations\":[{\"title\":\"Source A\",\"url\":\"https://example.com/a#section\"},{\"url\":\"http://127.0.0.1/private\"}],\"termination_reason\":\"completed\"}\n\n"
		return response(http.StatusOK, "text/event-stream; charset=utf-8", body), nil
	})}

	researcher, err := newTinyFishResearcherWithHTTPClient(TinyFishConfig{APIKey: "secret-key"}, client)
	if err != nil {
		t.Fatalf("construct researcher: %v", err)
	}
	result, err := researcher.Research(context.Background(), Request{
		Query:          " compare storage systems ",
		Purpose:        " product research ",
		OutputLanguage: "en",
	})
	if err != nil {
		t.Fatalf("research: %v", err)
	}
	if result.Report != "Bounded report" || result.TerminationReason != "completed" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if len(result.Citations) != 1 || result.Citations[0].URL != "https://example.com/a" || result.Citations[0].Title != "Source A" {
		t.Fatalf("unexpected citations: %#v", result.Citations)
	}
}

func TestTinyFishResearchAcceptsDataEventTypeAndNestedPayload(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		body := "data: {\"type\":\"final_result\",\"data\":{\"report\":\"Nested report\",\"citations\":[\"https://example.org/source\"]}}\n\n"
		return response(http.StatusOK, "text/event-stream", body), nil
	})}
	researcher, err := newTinyFishResearcherWithHTTPClient(TinyFishConfig{APIKey: "key"}, client)
	if err != nil {
		t.Fatalf("construct researcher: %v", err)
	}
	result, err := researcher.Research(context.Background(), Request{Query: "query", Purpose: "research"})
	if err != nil {
		t.Fatalf("research: %v", err)
	}
	if result.Report != "Nested report" || len(result.Citations) != 1 || result.Citations[0].URL != "https://example.org/source" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestTinyFishResearchRejectsInvalidRequestsBeforeNetwork(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("network should not be reached")
	})}
	researcher, err := newTinyFishResearcherWithHTTPClient(TinyFishConfig{APIKey: "key"}, client)
	if err != nil {
		t.Fatalf("construct researcher: %v", err)
	}
	for _, request := range []Request{
		{Query: "", Purpose: "purpose"},
		{Query: "query", Purpose: ""},
		{Query: "query", Purpose: "purpose", OutputLanguage: "en\nmalformed"},
	} {
		if _, err := researcher.Research(context.Background(), request); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("expected invalid request, got %v", err)
		}
	}
	if calls != 0 {
		t.Fatalf("invalid request reached network %d times", calls)
	}
}

func TestTinyFishResearchSanitizesProviderHTTPFailure(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return response(http.StatusTooManyRequests, "application/json", `{"error":"sensitive provider detail"}`), nil
	})}
	researcher, err := newTinyFishResearcherWithHTTPClient(TinyFishConfig{APIKey: "key"}, client)
	if err != nil {
		t.Fatalf("construct researcher: %v", err)
	}
	_, err = researcher.Research(context.Background(), Request{Query: "query", Purpose: "research"})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected provider unavailable, got %v", err)
	}
	if strings.Contains(err.Error(), "sensitive provider detail") {
		t.Fatal("provider response body leaked through error")
	}
}

func TestTinyFishResearchRequiresEventStreamAndFinalResult(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{name: "content type", contentType: "application/json", body: `{}`},
		{name: "missing final", contentType: "text/event-stream", body: "event: research_progress\ndata: {\"message\":\"working\"}\n\n"},
		{name: "invalid json", contentType: "text/event-stream", body: "event: final_result\ndata: {not-json}\n\n"},
		{name: "missing report", contentType: "text/event-stream", body: "event: final_result\ndata: {\"citations\":[]}\n\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
				return response(http.StatusOK, test.contentType, test.body), nil
			})}
			researcher, err := newTinyFishResearcherWithHTTPClient(TinyFishConfig{APIKey: "key"}, client)
			if err != nil {
				t.Fatalf("construct researcher: %v", err)
			}
			_, err = researcher.Research(context.Background(), Request{Query: "query", Purpose: "research"})
			if !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("expected invalid response, got %v", err)
			}
		})
	}
}

func TestTinyFishResearchRejectsOversizedEvent(t *testing.T) {
	body := "event: final_result\ndata: " + strings.Repeat("x", MaxTinyFishResearchEventBytes+1) + "\n\n"
	client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return response(http.StatusOK, "text/event-stream", body), nil
	})}
	researcher, err := newTinyFishResearcherWithHTTPClient(TinyFishConfig{APIKey: "key"}, client)
	if err != nil {
		t.Fatalf("construct researcher: %v", err)
	}
	_, err = researcher.Research(context.Background(), Request{Query: "query", Purpose: "research"})
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("expected invalid response, got %v", err)
	}
}

func TestTinyFishResearchConstructorRequiresCredential(t *testing.T) {
	if _, err := NewTinyFishResearcher(TinyFishConfig{}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid request for missing API key, got %v", err)
	}
}
