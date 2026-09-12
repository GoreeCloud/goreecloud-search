package retrieval

import "context"

const MaxURLsPerRequest = 10

// Document is GoreeCloud's bounded retrieval representation. Provider-specific
// response metadata must be normalized before it reaches callers.
type Document struct {
	URL      string `json:"url"`
	FinalURL string `json:"final_url,omitempty"`
	Title    string `json:"title,omitempty"`
	Content  string `json:"content"`
}

// Failure intentionally exposes a stable GoreeCloud code instead of raw
// provider diagnostics, which may contain provider-internal or target details.
type Failure struct {
	URL  string `json:"url"`
	Code string `json:"code"`
}

type Response struct {
	Documents []Document `json:"documents"`
	Failures  []Failure  `json:"failures,omitempty"`
}

type Fetcher interface {
	Fetch(context.Context, []string) (Response, error)
}
