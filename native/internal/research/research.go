package research

import "context"

const (
	MaxQueryRunes          = 4096
	MaxPurposeRunes        = 512
	MaxOutputLanguageRunes = 64
)

// Request is GoreeCloud's provider-neutral research request. Purpose is a local
// authorization/audit binding and must not be forwarded to an external provider
// unless a separately reviewed provider contract explicitly requires it.
type Request struct {
	Query          string `json:"query"`
	Purpose        string `json:"purpose"`
	OutputLanguage string `json:"output_language,omitempty"`
}

// Citation is the bounded source identity that GoreeCloud exposes from a
// provider research result. Provider-specific citation metadata is intentionally
// not part of the stable contract.
type Citation struct {
	Title string `json:"title,omitempty"`
	URL   string `json:"url"`
}

// Result is GoreeCloud's normalized research representation.
type Result struct {
	Report            string     `json:"report"`
	Citations         []Citation `json:"citations,omitempty"`
	TerminationReason string     `json:"termination_reason,omitempty"`
}

// Researcher is implemented by external research providers behind GoreeCloud
// Search. Callers depend on this interface rather than provider-specific APIs.
type Researcher interface {
	Research(context.Context, Request) (Result, error)
}
