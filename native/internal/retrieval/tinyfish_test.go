package retrieval

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestTinyFishFetchMapsRequestAndPartialResponse(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost || request.URL.String() != TinyFishFetchEndpoint {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.String())
		}
		if request.Header.Get("X-API-Key") != "test-key" {
			t.Fatal("TinyFish API key was not attached")
		}
		if request.Header.Get("Referer") != "" || request.Header.Get("Cookie") != "" {
			t.Fatal("retrieval request unexpectedly contains browser state")
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != `{"urls":["https://example.com/article","https://example.org/page"],"format":"markdown"}` {
			t.Fatalf("request body = %s", body)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
				"results":[{
					"url":"https://example.com/article#old",
					"final_url":"https://example.com/article#new",
					"title":"Example Article",
					"text":"# Example\n\nExtracted content",
					"future_field":"ignored"
				}],
				"errors":[{"url":"https://example.org/page","error":"upstream detail must not escape"}],
				"future_top_level":"ignored"
			}`)),
		}, nil
	})}
	fetcher, err := newTinyFishFetcherWithHTTPClient(TinyFishConfig{APIKey: "test-key"}, client)
	if err != nil {
		t.Fatal(err)
	}
	response, err := fetcher.Fetch(context.Background(), []string{"https://example.com/article", "https://example.org/page"})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Documents) != 1 || response.Documents[0].Title != "Example Article" {
		t.Fatalf("unexpected documents: %+v", response.Documents)
	}
	if response.Documents[0].URL != "https://example.com/article" || response.Documents[0].FinalURL != "https://example.com/article" {
		t.Fatalf("unexpected normalized URLs: %+v", response.Documents[0])
	}
	if response.Documents[0].Content != "# Example\n\nExtracted content" {
		t.Fatalf("unexpected content: %q", response.Documents[0].Content)
	}
	if len(response.Failures) != 1 || response.Failures[0].Code != FailureFetchFailed {
		t.Fatalf("unexpected failures: %+v", response.Failures)
	}
}

func TestTinyFishFetchRejectsUnsafeTargets(t *testing.T) {
	fetcher, err := NewTinyFishFetcher(TinyFishConfig{APIKey: "test-key"})
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{
		"file:///etc/passwd",
		"http://localhost/private",
		"http://service.local/private",
		"http://127.0.0.1/private",
		"http://10.0.0.1/private",
		"https://user:pass@example.com/private",
		"https://singlelabel/path",
	} {
		if _, err := fetcher.Fetch(context.Background(), []string{target}); err == nil {
			t.Fatalf("unsafe target %q unexpectedly succeeded", target)
		}
	}
}

func TestTinyFishFetchBoundsRequestCardinality(t *testing.T) {
	fetcher, err := NewTinyFishFetcher(TinyFishConfig{APIKey: "test-key"})
	if err != nil {
		t.Fatal(err)
	}
	urls := make([]string, MaxURLsPerRequest+1)
	for i := range urls {
		urls[i] = "https://example.com/item"
	}
	if _, err := fetcher.Fetch(context.Background(), urls); err == nil {
		t.Fatal("oversized URL batch unexpectedly succeeded")
	}
}

func TestTinyFishFetchRejectsUnsupportedHTTPResponse(t *testing.T) {
	for _, test := range []struct {
		name        string
		status      int
		contentType string
		body        string
	}{
		{name: "status", status: http.StatusTooManyRequests, contentType: "application/json", body: `{}`},
		{name: "content type", status: http.StatusOK, contentType: "text/html", body: `<html></html>`},
		{name: "invalid json", status: http.StatusOK, contentType: "application/json", body: `{`},
		{name: "trailing data", status: http.StatusOK, contentType: "application/json", body: `{"results":[]} {}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: test.status,
					Header:     http.Header{"Content-Type": []string{test.contentType}},
					Body:       io.NopCloser(strings.NewReader(test.body)),
				}, nil
			})}
			fetcher, err := newTinyFishFetcherWithHTTPClient(TinyFishConfig{APIKey: "test-key"}, client)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := fetcher.Fetch(context.Background(), []string{"https://example.com"}); err == nil {
				t.Fatal("unsupported response unexpectedly succeeded")
			}
		})
	}
}

func TestTinyFishFetchRejectsOversizedResponse(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(strings.Repeat(" ", MaxTinyFishFetchResponseBytes+1))),
		}, nil
	})}
	fetcher, err := newTinyFishFetcherWithHTTPClient(TinyFishConfig{APIKey: "test-key"}, client)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fetcher.Fetch(context.Background(), []string{"https://example.com"}); err == nil || !strings.Contains(err.Error(), "maximum size") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTinyFishFetchBoundsReturnedText(t *testing.T) {
	longTitle := strings.Repeat("t", maxTinyFishTitleRunes+100)
	longContent := strings.Repeat("c", maxTinyFishContentRunes+100)
	body := `{"results":[{"url":"https://example.com","title":"` + longTitle + `","text":"` + longContent + `"}],"errors":[]}`
	client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})}
	fetcher, err := newTinyFishFetcherWithHTTPClient(TinyFishConfig{APIKey: "test-key"}, client)
	if err != nil {
		t.Fatal(err)
	}
	response, err := fetcher.Fetch(context.Background(), []string{"https://example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Documents) != 1 || len([]rune(response.Documents[0].Title)) != maxTinyFishTitleRunes || len([]rune(response.Documents[0].Content)) != maxTinyFishContentRunes {
		t.Fatal("TinyFish Fetch output was not bounded")
	}
}
