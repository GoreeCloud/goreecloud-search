package search

import "testing"

func TestSuggestQueryCorrectionRequiresIndependentProviderAgreement(t *testing.T) {
	query := "goreecluod search"
	oneProvider := []Result{
		{Title: "GoreeCloud Search", Provider: "one"},
		{Title: "GoreeCloud Search architecture", Provider: "one"},
	}
	if got := suggestQueryCorrection(query, oneProvider); got != "" {
		t.Fatalf("one provider must not authoritatively correct a query, got %q", got)
	}

	twoProviders := append(oneProvider, Result{Title: "GoreeCloud private search", Provider: "two"})
	if got := suggestQueryCorrection(query, twoProviders); got != "goreecloud search" {
		t.Fatalf("cross-provider title agreement should support a correction, got %q", got)
	}
}

func TestSuggestQueryCorrectionDoesNotRewriteExactToken(t *testing.T) {
	results := []Result{
		{Title: "Goreecluod Search project", Provider: "one"},
		{Title: "GoreeCloud Search", Provider: "two"},
		{Title: "GoreeCloud Search docs", Provider: "three"},
	}
	if got := suggestQueryCorrection("goreecluod search", results); got != "" {
		t.Fatalf("an exact title occurrence should suppress correction authority, got %q", got)
	}
}

func TestSuggestQueryCorrectionRejectsAmbiguousEqualSupport(t *testing.T) {
	results := []Result{
		{Title: "GoreeCloud Search", Provider: "one"},
		{Title: "GoreeCloud Docs", Provider: "two"},
		{Title: "Goreecload Search", Provider: "three"},
		{Title: "Goreecload Docs", Provider: "four"},
	}
	if got := suggestQueryCorrection("goreecluod search", results); got != "" {
		t.Fatalf("equal cross-provider support for two corrections should remain ambiguous, got %q", got)
	}
}

func TestSuggestQueryCorrectionPreservesOperatorsAndQuotedText(t *testing.T) {
	results := []Result{
		{Title: "GoreeCloud Private Search", Provider: "one"},
		{Title: "GoreeCloud Search privacy", Provider: "two"},
	}
	query := `site:docs.example goreecluod "private search" filetype:pdf`
	want := `site:docs.example goreecloud "private search" filetype:pdf`
	if got := suggestQueryCorrection(query, results); got != want {
		t.Fatalf("suggestion=%q, want %q", got, want)
	}
}

func TestSuggestQueryCorrectionDoesNotRewriteDomainTarget(t *testing.T) {
	results := []Result{
		{Title: "Example Search", Provider: "one"},
		{Title: "Example Search docs", Provider: "two"},
	}
	if got := suggestQueryCorrection("example.com", results); got != "" {
		t.Fatalf("navigational domains must not be rewritten, got %q", got)
	}
}

func TestSuggestQueryCorrectionAllowsAtMostOneCorrectedToken(t *testing.T) {
	results := []Result{
		{Title: "GoreeCloud Privacy", Provider: "one"},
		{Title: "GoreeCloud Privacy architecture", Provider: "two"},
	}
	if got := suggestQueryCorrection("goreecluod pirvacy", results); got != "" {
		t.Fatalf("multiple corrections should fail conservative suggestion policy, got %q", got)
	}
}
