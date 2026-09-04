package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	searchcore "github.com/GoreeCloud/goreecloud-search/native/internal/search"
)

// TestNativeProviderRuntimeExecutesEverySearchCategory is a deterministic
// source-level acceptance check for the GoreeCloud-owned provider contract.
// It deliberately uses an injected transport instead of a live third-party
// provider so this test cannot be mistaken for provider selection, provider
// approval, credential validation, or production network acceptance.
func TestNativeProviderRuntimeExecutesEverySearchCategory(t *testing.T) {
	published := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var input providerRequest
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			return nil, err
		}
		if input.SchemaVersion != ProviderContractVersion || input.Query != "goreecloud acceptance" {
			return nil, fmt.Errorf("unexpected request contract: %+v", input)
		}

		result := wireResult{
			Title:       "GoreeCloud " + input.Category + " acceptance",
			URL:         "https://results.example/" + input.Category,
			Snippet:     "Deterministic native provider result for " + input.Category,
			Score:       50,
			PublishedAt: &published,
		}
		if input.Category == searchcore.CategoryImages {
			result.Media = &searchcore.Media{
				Kind:         searchcore.MediaKindImage,
				ThumbnailURL: "https://media.example/thumb.jpg#ignored",
				ContentURL:   "https://media.example/full.jpg#ignored",
				MIMEType:     "IMAGE/JPEG",
				Width:        1600,
				Height:       900,
				Alt:          "  GoreeCloud acceptance image  ",
			}
		}
		payload, err := json.Marshal(providerResponse{
			SchemaVersion: ProviderContractVersion,
			Results:       []wireResult{result},
		})
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(string(payload))),
		}, nil
	})

	provider, err := newHTTPJSONProviderWithClient(HTTPJSONConfig{
		Name:                     "Deterministic acceptance provider",
		Endpoint:                 "https://provider.example/search",
		Categories:               append([]string(nil), searchcore.SupportedCategories...),
		PublishedAtAuthoritative: true,
	}, &http.Client{Transport: transport})
	if err != nil {
		t.Fatal(err)
	}
	engine := searchcore.NewEngine(2*time.Second, provider)

	for _, category := range searchcore.SupportedCategories {
		t.Run(category, func(t *testing.T) {
			if !engine.SupportsCategory(category) {
				t.Fatalf("native engine does not report %q executable", category)
			}
			response, err := engine.SearchCategory(context.Background(), "goreecloud acceptance", category)
			if err != nil {
				t.Fatal(err)
			}
			if response.Category != category {
				t.Fatalf("response category = %q, want %q", response.Category, category)
			}
			if response.Degraded {
				t.Fatalf("deterministic %q response unexpectedly degraded", category)
			}
			if len(response.Providers) != 1 || response.Providers[0].State != searchcore.ProviderStateAvailable || response.Providers[0].Count != 1 {
				t.Fatalf("unexpected provider state for %q: %+v", category, response.Providers)
			}
			if len(response.Results) != 1 {
				t.Fatalf("result count for %q = %d", category, len(response.Results))
			}
			result := response.Results[0]
			if result.Provider != provider.Name() || result.URL != "https://results.example/"+category {
				t.Fatalf("unexpected normalized result for %q: %+v", category, result)
			}
			if result.PublishedAt == nil || result.PublishedAtSource != provider.Name() {
				t.Fatalf("authoritative publication metadata was not preserved for %q: %+v", category, result)
			}

			if category == searchcore.CategoryImages {
				if result.Media == nil {
					t.Fatal("image category lost media metadata")
				}
				if result.Media.Kind != searchcore.MediaKindImage ||
					result.Media.ThumbnailURL != "https://media.example/thumb.jpg" ||
					result.Media.ContentURL != "https://media.example/full.jpg" ||
					result.Media.MIMEType != "image/jpeg" ||
					result.Media.Alt != "GoreeCloud acceptance image" {
					t.Fatalf("image media was not normalized: %+v", result.Media)
				}
			} else if result.Media != nil {
				t.Fatalf("non-image category unexpectedly gained media metadata: %+v", result.Media)
			}
		})
	}
}

func TestNativeProviderRuntimeKeepsUnconfiguredCategoriesFailClosed(t *testing.T) {
	provider, err := newHTTPJSONProviderWithClient(HTTPJSONConfig{
		Name:       "General only",
		Endpoint:   "https://provider.example/search",
		Categories: []string{searchcore.CategoryGeneral},
	}, &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		t.Fatal("unconfigured specialized category unexpectedly reached provider transport")
		return nil, nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	engine := searchcore.NewEngine(time.Second, provider)
	for _, category := range []string{
		searchcore.CategoryImages,
		searchcore.CategoryVideos,
		searchcore.CategoryNews,
		searchcore.CategoryFiles,
	} {
		if engine.SupportsCategory(category) {
			t.Fatalf("unconfigured category %q unexpectedly executable", category)
		}
		if _, err := engine.SearchCategory(context.Background(), "test", category); err == nil {
			t.Fatalf("unconfigured category %q unexpectedly searched", category)
		}
	}
}
