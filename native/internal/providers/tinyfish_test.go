package providers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

func TestTinyFishSearchMapsGeneralRequestAndResults(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet {
			t.Fatalf("method = %q", request.Method)
		}
		if request.URL.Scheme != "https" || request.URL.Host != "api.search.tinyfish.ai" || request.URL.Path != "" {
			t.Fatalf("unexpected endpoint: %s", request.URL.String())
		}
		if got := request.URL.Query().Get("query"); got != "native search" {
			t.Fatalf("query = %q", got)
		}
		if got := request.URL.Query().Get("domain_type"); got != "web" {
			t.Fatalf("domain_type = %q", got)
		}
		if got := request.Header.Get("X-API-Key"); got != "test-key" {
			t.Fatalf("X-API-Key = %q", got)
		}
		if got := request.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("Accept = %q", got)
		}
		body := `{"query":"native search","results":[{"position":1,"site_name":"Example","title":"Example title","snippet":"Example snippet","url":"https://example.com/result","future_field":"ignored"}],"total_results":1,"page":0,"future_top_level":"ignored"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})}
	provider, err := newTinyFishSearchProviderWithClient(TinyFishSearchConfig{
		Name:       "TinyFish",
		Categories: []string{searchcore.CategoryGeneral, searchcore.CategoryNews},
		APIKey:     "test-key",
	}, client)
	if err != nil {
		t.Fatal(err)
	}
	results, err := provider.SearchCategory(context.Background(), "native search", searchcore.CategoryGeneral)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("result count = %d", len(results))
	}
	if results[0].Title != "Example title" || results[0].URL != "https://example.com/result" || results[0].Snippet != "Example snippet" {
		t.Fatalf("unexpected result: %+v", results[0])
	}
	if results[0].Score != 300 {
		t.Fatalf("score = %d", results[0].Score)
	}
	if results[0].PublishedAt != nil {
		t.Fatal("TinyFish date unexpectedly became an authoritative publication timestamp")
	}
}

func TestTinyFishSearchMapsNewsCategory(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if got := request.URL.Query().Get("domain_type"); got != "news" {
			t.Fatalf("domain_type = %q", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"query":"updates","results":[],"total_results":0,"page":0}`)),
		}, nil
	})}
	provider, err := newTinyFishSearchProviderWithClient(TinyFishSearchConfig{
		Name:       "TinyFish News",
		Categories: []string{searchcore.CategoryNews},
		APIKey:     "test-key",
	}, client)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.SearchCategory(context.Background(), "updates", searchcore.CategoryNews); err != nil {
		t.Fatal(err)
	}
}

func TestTinyFishSearchRejectsUnsafeConfiguration(t *testing.T) {
	tests := []TinyFishSearchConfig{
		{Name: "TinyFish", Endpoint: "https://example.com", Categories: []string{searchcore.CategoryGeneral}, APIKey: "secret"},
		{Name: "TinyFish", Categories: []string{searchcore.CategoryImages}, APIKey: "secret"},
		{Name: "TinyFish", Categories: []string{searchcore.CategoryGeneral}, APIKey: ""},
	}
	for _, config := range tests {
		if _, err := NewTinyFishSearchProvider(config); err == nil {
			t.Fatalf("unsafe config unexpectedly succeeded: %+v", config)
		}
	}
}

func TestTinyFishSearchRejectsUnsupportedRuntimeCategory(t *testing.T) {
	provider, err := NewTinyFishSearchProvider(TinyFishSearchConfig{
		Name:       "TinyFish",
		Categories: []string{searchcore.CategoryGeneral},
		APIKey:     "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.SearchCategory(context.Background(), "cats", searchcore.CategoryImages); err == nil {
		t.Fatal("unsupported category unexpectedly succeeded")
	}
}

func TestTinyFishConfigRequiresEnvironmentCredentialAndNonAuthoritativeTime(t *testing.T) {
	valid := `{"schema_version":1,"providers":[{"name":"TinyFish","adapter":"tinyfish-search-v1","endpoint":"https://api.search.tinyfish.ai","categories":["general","news"],"credential_env":"TINYFISH_API_KEY"}]}`
	configured, err := loadConfigBytes([]byte(valid), func(name string) (string, bool) {
		if name == "TINYFISH_API_KEY" {
			return "secret", true
		}
		return "", false
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(configured) != 1 || configured[0].Name() != "TinyFish" {
		t.Fatalf("unexpected providers: %+v", configured)
	}
	categorized, ok := configured[0].(searchcore.CategoryProvider)
	if !ok || strings.Join(categorized.Categories(), ",") != "general,news" {
		t.Fatalf("unexpected TinyFish categories")
	}

	missingCredential := `{"schema_version":1,"providers":[{"name":"TinyFish","adapter":"tinyfish-search-v1","categories":["general"]}]}`
	if _, err := loadConfigBytes([]byte(missingCredential), func(string) (string, bool) { return "", false }); err == nil {
		t.Fatal("TinyFish config without credential_env unexpectedly succeeded")
	}

	authoritativeTimestamp := `{"schema_version":1,"providers":[{"name":"TinyFish","adapter":"tinyfish-search-v1","categories":["news"],"credential_env":"TINYFISH_API_KEY","published_at_authoritative":true}]}`
	if _, err := loadConfigBytes([]byte(authoritativeTimestamp), func(string) (string, bool) { return "secret", true }); err == nil {
		t.Fatal("TinyFish authoritative timestamp config unexpectedly succeeded")
	}
}

func TestTinyFishPositionScoreIsBounded(t *testing.T) {
	for position, want := range map[int]int{0: 0, 1: 300, 2: 299, 300: 1, 301: 1, 1000: 1} {
		if got := tinyFishPositionScore(position); got != want {
			t.Fatalf("position %d score = %d, want %d", position, got, want)
		}
	}
}
