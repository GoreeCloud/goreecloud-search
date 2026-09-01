package platformstate

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type staticEvidenceSource struct {
	bundle RuntimeEvidenceBundle
}

func (source staticEvidenceSource) Evidence() RuntimeEvidenceBundle {
	return source.bundle
}

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

func TestPrivacyShieldTransportAcceptanceDoesNotCreateAuthorization(t *testing.T) {
	now := time.Date(2026, time.September, 1, 17, 0, 0, 0, time.UTC)
	snapshot := SnapshotWithEvidence(staticEvidenceSource{bundle: RuntimeEvidenceBundle{
		PrivacyShield: &PrivacyShieldRuntimeAcceptance{
			SchemaVersion:        "1.4.0",
			System:               "privacy-shield",
			AuthorityDomain:      "privacy",
			ProductionAcceptance: true,
		},
	}}, now)
	privacy := snapshot.Systems[0]
	if privacy.RuntimeEvidence != "transport-accepted-application-unverified" {
		t.Fatalf("privacy runtime evidence = %q", privacy.RuntimeEvidence)
	}
	if privacy.State != "unknown" || privacy.PositiveClaim || privacy.ProductionAccepted {
		t.Fatal("Privacy Shield transport acceptance must not become Search authorization")
	}
}

func TestWardveilProtectedClaimRequiresCurrentAuthoritativeEvidence(t *testing.T) {
	now := time.Date(2026, time.September, 1, 17, 0, 0, 0, time.UTC)
	snapshot := SnapshotWithEvidence(staticEvidenceSource{bundle: RuntimeEvidenceBundle{
		Wardveil: &WardveilStatusRecord{
			ContractVersion:        "0.1.0",
			ScopeKind:              "application",
			ScopeID:                searchScopeID,
			AuthoritySystem:        "wardveil-security",
			AuthorityControl:       "application-runtime",
			AuthorityAuthoritative: true,
			State:                  "protected",
			EvidenceStatus:         "current",
			ObservedAt:             now.Add(-time.Minute),
			ValidUntil:             now.Add(5 * time.Minute),
			ProtectedByWardveil:    true,
		},
	}}, now)
	wardveil := snapshot.Systems[1]
	if wardveil.RuntimeEvidence != "current" || wardveil.State != "protected" || !wardveil.PositiveClaim {
		t.Fatalf("unexpected Wardveil projection: %+v", wardveil)
	}
	if wardveil.ProductionAccepted {
		t.Fatal("Wardveil status alone must not create Search production acceptance")
	}
}

func TestExpiredWardveilProtectedClaimFailsClosed(t *testing.T) {
	now := time.Date(2026, time.September, 1, 17, 0, 0, 0, time.UTC)
	snapshot := SnapshotWithEvidence(staticEvidenceSource{bundle: RuntimeEvidenceBundle{
		Wardveil: &WardveilStatusRecord{
			ContractVersion:        "0.1.0",
			ScopeKind:              "service",
			ScopeID:                searchScopeID,
			AuthoritySystem:        "wardveil-security",
			AuthorityControl:       "service-runtime",
			AuthorityAuthoritative: true,
			State:                  "protected",
			EvidenceStatus:         "current",
			ObservedAt:             now.Add(-time.Minute),
			ValidUntil:             now.Add(-time.Second),
			ProtectedByWardveil:    true,
		},
	}}, now)
	wardveil := snapshot.Systems[1]
	if wardveil.RuntimeEvidence != "stale" || wardveil.State != "unknown" || wardveil.PositiveClaim {
		t.Fatalf("expired Wardveil evidence did not fail closed: %+v", wardveil)
	}
}

func TestEverkeepRequiresCompleteFreshContinuitySet(t *testing.T) {
	now := time.Date(2026, time.September, 1, 17, 0, 0, 0, time.UTC)
	record := func(dimension string) EverkeepContinuityRecord {
		return EverkeepContinuityRecord{
			Producer:           "everkeep",
			Scope:              searchScopeID,
			Dimension:          dimension,
			State:              "ready",
			ObservedAt:         now.Add(-time.Minute),
			FreshUntil:         now.Add(10 * time.Minute),
			RequiredEvidence:   true,
			VerificationMethod: "accepted-runtime-evidence",
			EvidenceReference:  "everkeep:evidence:" + dimension,
		}
	}
	snapshot := SnapshotWithEvidence(staticEvidenceSource{bundle: RuntimeEvidenceBundle{
		Everkeep: []EverkeepContinuityRecord{
			record("backup_coverage"),
			record("restore_capability"),
			record("recovery_freshness"),
		},
	}}, now)
	everkeep := snapshot.Systems[2]
	if everkeep.RuntimeEvidence != "current" || everkeep.State != "ready" || !everkeep.PositiveClaim {
		t.Fatalf("unexpected Everkeep projection: %+v", everkeep)
	}
	if everkeep.ProductionAccepted {
		t.Fatal("Everkeep continuity status alone must not create Search production acceptance")
	}
}

func TestEverkeepDuplicateDimensionFailsClosed(t *testing.T) {
	now := time.Date(2026, time.September, 1, 17, 0, 0, 0, time.UTC)
	record := EverkeepContinuityRecord{
		Producer:           "everkeep",
		Scope:              searchScopeID,
		Dimension:          "backup_coverage",
		State:              "ready",
		ObservedAt:         now.Add(-time.Minute),
		FreshUntil:         now.Add(10 * time.Minute),
		VerificationMethod: "accepted-runtime-evidence",
	}
	snapshot := SnapshotWithEvidence(staticEvidenceSource{bundle: RuntimeEvidenceBundle{
		Everkeep: []EverkeepContinuityRecord{record, record},
	}}, now)
	everkeep := snapshot.Systems[2]
	if everkeep.RuntimeEvidence != "unverified" || everkeep.State != "unknown" || everkeep.PositiveClaim {
		t.Fatalf("duplicate Everkeep dimension did not fail closed: %+v", everkeep)
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
