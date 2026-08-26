package syncstate

import "testing"

func TestCapabilitiesExposeStableSearchDatasets(t *testing.T) {
	got := Capabilities()
	if len(got) != 3 {
		t.Fatalf("capability count = %d, want 3", len(got))
	}
	want := []string{"search.preferences", "search.history", "search.sources"}
	for i, dataset := range want {
		if got[i].Dataset != dataset {
			t.Fatalf("capability[%d] = %q, want %q", i, got[i].Dataset, dataset)
		}
		if got[i].Application != "search" || got[i].SchemaVersion != 1 {
			t.Fatalf("unexpected capability metadata: %+v", got[i])
		}
	}

	got[0].Dataset = "mutated"
	if Capabilities()[0].Dataset != "search.preferences" {
		t.Fatal("Capabilities returned shared mutable storage")
	}
}
