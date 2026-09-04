package providers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestHTTPProviderExecutesVersionedCategoryRequest(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost {
			t.Fatalf("method = %q", request.Method)
		}
		if request.URL.String() != "https://provider.example/search" {
			t.Fatalf("URL = %q", request.URL.String())
		}
		if request.Header.Get("Authorization") != "Bearer secret-value" {
			t.Fatal("bearer credential was not attached")
		}
		if request.Header.Get("Referer") != "" || request.Header.Get("Cookie") != "" {
			t.Fatal("provider request unexpectedly contains browser state")
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != `{"schema_version":1,"query":"goreecloud search","category":"images"}` {
			t.Fatalf("request body = %s", body)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Body: io.NopCloser(strings.NewReader(`{
				"schema_version":1,
				"results":[{
					"title":"GoreeCloud image",
					"url":"https://example.com/item",
					"snippet":"Native result",
					"score":9999,
					"media":{"kind":"image","thumbnail_url":"https://cdn.example.com/thumb.jpg","content_url":"https://cdn.example.com/full.jpg","mime_type":"image/jpeg","width":1200,"height":800,"alt":"Example"}
				}]
			}`)),
		}, nil
	})}
	provider, err := newHTTPJSONProviderWithClient(HTTPJSONConfig{
		Name:        "Example",
		Endpoint:    "https://provider.example/search",
		Categories:  []string{searchcore.CategoryGeneral, searchcore.CategoryImages},
		BearerToken: "secret-value",
	}, client)
	if err != nil {
		t.Fatal(err)
	}
	results, err := provider.SearchCategory(context.Background(), "goreecloud search", searchcore.CategoryImages)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Title != "GoreeCloud image" || results[0].Media == nil {
		t.Fatalf("unexpected results: %+v", results)
	}
	if results[0].Provider != "" {
		t.Fatal("adapter must not manufacture engine-owned provider attribution")
	}
}

func TestHTTPProviderRejectsUnsupportedResponseContracts(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		contentType string
		body        string
	}{
		{name: "status", status: http.StatusTooManyRequests, contentType: "application/json", body: `{"schema_version":1,"results":[]}`},
		{name: "content type", status: http.StatusOK, contentType: "text/html", body: `<html></html>`},
		{name: "schema", status: http.StatusOK, contentType: "application/json", body: `{"schema_version":2,"results":[]}`},
		{name: "unknown field", status: http.StatusOK, contentType: "application/json", body: `{"schema_version":1,"results":[],"secret":"no"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: test.status,
					Header:     http.Header{"Content-Type": []string{test.contentType}},
					Body:       io.NopCloser(strings.NewReader(test.body)),
				}, nil
			})}
			provider, err := newHTTPJSONProviderWithClient(HTTPJSONConfig{
				Name: "Example", Endpoint: "https://provider.example/search", Categories: []string{searchcore.CategoryGeneral},
			}, client)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := provider.Search(context.Background(), "test"); err == nil {
				t.Fatal("invalid provider response unexpectedly succeeded")
			}
		})
	}
}

func TestHTTPProviderBoundsResponseAndResultFields(t *testing.T) {
	longTitle := strings.Repeat("x", maxProviderTitleRunes+100)
	longSnippet := strings.Repeat("y", maxProviderSnippetRunes+100)
	items := make([]string, 0, searchcore.MaxResultsPerProvider+20)
	for i := 0; i < searchcore.MaxResultsPerProvider+20; i++ {
		items = append(items, `{"title":"`+longTitle+`","url":"https://example.com/item","snippet":"`+longSnippet+`"}`)
	}
	body := `{"schema_version":1,"results":[` + strings.Join(items, ",") + `]}`
	client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})}
	provider, err := newHTTPJSONProviderWithClient(HTTPJSONConfig{
		Name: "Example", Endpoint: "https://provider.example/search", Categories: []string{searchcore.CategoryGeneral},
	}, client)
	if err != nil {
		t.Fatal(err)
	}
	results, err := provider.Search(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != searchcore.MaxResultsPerProvider+1 {
		t.Fatalf("result count = %d", len(results))
	}
	if len([]rune(results[0].Title)) != maxProviderTitleRunes || len([]rune(results[0].Snippet)) != maxProviderSnippetRunes {
		t.Fatal("provider result text was not bounded")
	}
}

func TestHTTPProviderRejectsOversizedResponse(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(strings.Repeat(" ", MaxProviderResponseBytes+1))),
		}, nil
	})}
	provider, err := newHTTPJSONProviderWithClient(HTTPJSONConfig{
		Name: "Example", Endpoint: "https://provider.example/search", Categories: []string{searchcore.CategoryGeneral},
	}, client)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Search(context.Background(), "test"); err == nil || !strings.Contains(err.Error(), "maximum size") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPProviderPublicationAuthorityIsExplicit(t *testing.T) {
	provider, err := NewHTTPJSONProvider(HTTPJSONConfig{
		Name:                     "News authority",
		Endpoint:                 "https://provider.example/search",
		Categories:               []string{searchcore.CategoryNews},
		PublishedAtAuthoritative: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !provider.PublishedAtAuthoritative() {
		t.Fatal("explicit timestamp authority was lost")
	}
	if got := strings.Join(provider.Categories(), ","); got != searchcore.CategoryNews {
		t.Fatalf("categories = %q", got)
	}
}

func TestProviderEndpointSecurityValidation(t *testing.T) {
	for _, endpoint := range []string{
		"http://provider.example/search",
		"https://user:pass@provider.example/search",
		"https://localhost/search",
		"https://127.0.0.1/search",
		"https://10.0.0.1/search",
		"https://provider.example:8443/search",
		"https://provider.example/search?token=secret",
	} {
		if _, err := NewHTTPJSONProvider(HTTPJSONConfig{Name: "Example", Endpoint: endpoint, Categories: []string{searchcore.CategoryGeneral}}); err == nil {
			t.Fatalf("unsafe endpoint %q was accepted", endpoint)
		}
	}
}

func TestHTTPProviderHonorsCancelledContext(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		<-request.Context().Done()
		return nil, request.Context().Err()
	})}
	provider, err := newHTTPJSONProviderWithClient(HTTPJSONConfig{
		Name: "Example", Endpoint: "https://provider.example/search", Categories: []string{searchcore.CategoryGeneral},
	}, client)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if _, err := provider.Search(ctx, "test"); err == nil {
		t.Fatal("cancelled provider request unexpectedly succeeded")
	}
}
