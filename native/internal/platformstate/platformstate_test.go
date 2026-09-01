package platformstate

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDevelopmentSnapshotDoesNotManufactureAcceptance(t *testing.T) {
	snapshot := DevelopmentSnapshot()
	if snapshot.ProductionApproved {
		t.Fatal("development snapshot must not claim production approval")
	}
	if snapshot.ContainsUserContent || snapshot.ContainsQueryText || snapshot.CredentialsExposed {
		t.Fatal("platform status must remain minimized and non-sensitive")
	}
	if len(snapshot.Systems) != 3 {
		t.Fatalf("systems = %d, want 3", len(snapshot.Systems))
	}
	for _, system := range snapshot.Systems {
		if system.PositiveClaim || system.ProductionAccepted {
			t.Fatalf("%s manufactured a positive/production claim", system.ID)
		}
		if system.State != "unknown" {
			t.Fatalf("%s state = %q, want unknown without runtime evidence", system.ID, system.State)
		}
	}
}

func TestPlatformStatusHandlerIsSanitizedAndVersioned(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/platform/status?query=private-search", nil)
	response := httptest.NewRecorder()
	Handler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Header().Get("X-GoreeCloud-API-Version") != APIVersion {
		t.Fatalf("missing API version header")
	}
	body := response.Body.String()
	for _, required := range []string{
		`"id":"privacy-shield"`,
		`"id":"wardveil-security"`,
		`"id":"everkeep"`,
		`"runtime_evidence":"unavailable"`,
		`"runtime_evidence":"unverified"`,
		`"positive_claim":false`,
		`"production_accepted":false`,
		`"production_approved":false`,
		`"contains_query_text":false`,
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("platform response missing %s: %s", required, body)
		}
	}
	if strings.Contains(body, "private-search") {
		t.Fatalf("platform status echoed query input: %s", body)
	}
}
