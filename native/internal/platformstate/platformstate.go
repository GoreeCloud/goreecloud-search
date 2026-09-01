package platformstate

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const (
	APIVersion    = "1"
	searchScopeID = "goreecloud-search"
)

type SystemStatus struct {
	ID                  string `json:"id"`
	AuthorityRepository string `json:"authority_repository"`
	AuthorityContract   string `json:"authority_contract"`
	SourceIntegration   string `json:"source_integration"`
	RuntimeEvidence     string `json:"runtime_evidence"`
	State               string `json:"state"`
	PositiveClaim       bool   `json:"positive_claim"`
	ProductionAccepted  bool   `json:"production_accepted"`
}

type Snapshot struct {
	SchemaVersion       int            `json:"schema_version"`
	Product             string         `json:"product"`
	Service             string         `json:"service"`
	Scope               string         `json:"scope"`
	Systems             []SystemStatus `json:"systems"`
	ContainsUserContent bool           `json:"contains_user_content"`
	ContainsQueryText   bool           `json:"contains_query_text"`
	CredentialsExposed  bool           `json:"credentials_exposed"`
	ProductionApproved  bool           `json:"production_approved"`
}

type PrivacyShieldRuntimeAcceptance struct {
	SchemaVersion        string
	System               string
	AuthorityDomain      string
	ProductionAcceptance bool
}

type WardveilStatusRecord struct {
	ContractVersion         string
	ScopeKind               string
	ScopeID                 string
	AuthoritySystem         string
	AuthorityControl        string
	AuthorityAuthoritative  bool
	State                   string
	EvidenceStatus          string
	ObservedAt              time.Time
	ValidUntil              time.Time
	ProtectedByWardveil     bool
}

type EverkeepContinuityRecord struct {
	Producer           string
	Scope              string
	Dimension          string
	State              string
	ObservedAt         time.Time
	FreshUntil         time.Time
	RequiredEvidence   bool
	VerificationMethod string
	EvidenceReference  string
}

type RuntimeEvidenceBundle struct {
	PrivacyShield *PrivacyShieldRuntimeAcceptance
	Wardveil      *WardveilStatusRecord
	Everkeep      []EverkeepContinuityRecord
}

type RuntimeEvidenceSource interface {
	Evidence() RuntimeEvidenceBundle
}

type unavailableRuntimeEvidenceSource struct{}

func (unavailableRuntimeEvidenceSource) Evidence() RuntimeEvidenceBundle {
	return RuntimeEvidenceBundle{}
}

func DevelopmentSnapshot() Snapshot {
	return SnapshotWithEvidence(unavailableRuntimeEvidenceSource{}, time.Now())
}

func SnapshotWithEvidence(source RuntimeEvidenceSource, now time.Time) Snapshot {
	snapshot := baseDevelopmentSnapshot()
	if source == nil {
		return snapshot
	}

	evidence := source.Evidence()
	snapshot.Systems[0] = projectPrivacyShield(snapshot.Systems[0], evidence.PrivacyShield)
	snapshot.Systems[1] = projectWardveil(snapshot.Systems[1], evidence.Wardveil, now)
	snapshot.Systems[2] = projectEverkeep(snapshot.Systems[2], evidence.Everkeep, now)
	return snapshot
}

func baseDevelopmentSnapshot() Snapshot {
	return Snapshot{
		SchemaVersion: 1,
		Product:       "GoreeCloud Search",
		Service:       "search",
		Scope:         "native-development-source",
		Systems: []SystemStatus{
			{
				ID:                  "privacy-shield",
				AuthorityRepository: "GoreeCloud/goreecloud-privacy-shield",
				AuthorityContract:   "contracts/privacy-shield.platform-evidence.runtime-acceptance.json",
				SourceIntegration:   "present",
				RuntimeEvidence:     "unavailable",
				State:               "unknown",
			},
			{
				ID:                  "wardveil-security",
				AuthorityRepository: "GoreeCloud/goreecloud-wardveil-security",
				AuthorityContract:   "contracts/wardveil.status.schema.json",
				SourceIntegration:   "present",
				RuntimeEvidence:     "unverified",
				State:               "unknown",
			},
			{
				ID:                  "everkeep",
				AuthorityRepository: "GoreeCloud/goreecloud-everkeep",
				AuthorityContract:   "contracts/continuity.status.schema.json",
				SourceIntegration:   "presentation-boundary-only",
				RuntimeEvidence:     "unavailable",
				State:               "unknown",
			},
		},
		ContainsUserContent: false,
		ContainsQueryText:   false,
		CredentialsExposed:  false,
		ProductionApproved:  false,
	}
}

func projectPrivacyShield(status SystemStatus, evidence *PrivacyShieldRuntimeAcceptance) SystemStatus {
	if evidence == nil {
		return status
	}
	if evidence.SchemaVersion != "1.4.0" || evidence.System != "privacy-shield" || evidence.AuthorityDomain != "privacy" {
		status.RuntimeEvidence = "unverified"
		return status
	}
	if evidence.ProductionAcceptance {
		// This contract governs Privacy Shield evidence delivery, not Search authorization.
		// Its own boundary says transport validity never creates privacy authorization.
		status.RuntimeEvidence = "transport-accepted-application-unverified"
		return status
	}
	status.RuntimeEvidence = "transport-not-production-accepted"
	return status
}

func projectWardveil(status SystemStatus, evidence *WardveilStatusRecord, now time.Time) SystemStatus {
	if evidence == nil {
		return status
	}
	if evidence.ContractVersion != "0.1.0" ||
		(evidence.ScopeKind != "application" && evidence.ScopeKind != "service") ||
		evidence.ScopeID != searchScopeID || evidence.AuthoritySystem == "" ||
		evidence.AuthorityControl == "" || evidence.ObservedAt.IsZero() || evidence.ObservedAt.After(now) {
		status.RuntimeEvidence = "unverified"
		return status
	}
	if evidence.EvidenceStatus != "current" {
		status.RuntimeEvidence = normalizedEvidenceStatus(evidence.EvidenceStatus)
		return status
	}
	if evidence.ProtectedByWardveil != (evidence.State == "protected") {
		status.RuntimeEvidence = "unverified"
		return status
	}
	if evidence.State == "protected" {
		if !evidence.AuthorityAuthoritative || evidence.ValidUntil.IsZero() {
			status.RuntimeEvidence = "unverified"
			return status
		}
		if !evidence.ValidUntil.After(now) {
			status.RuntimeEvidence = "stale"
			return status
		}
		status.RuntimeEvidence = "current"
		status.State = "protected"
		status.PositiveClaim = true
		return status
	}

	status.RuntimeEvidence = "current"
	status.State = normalizedWardveilState(evidence.State)
	return status
}

func normalizedEvidenceStatus(value string) string {
	switch value {
	case "stale", "unavailable", "unverified":
		return value
	default:
		return "unverified"
	}
}

func normalizedWardveilState(state string) string {
	switch state {
	case "attention", "degraded", "unknown", "not_applicable":
		return state
	default:
		return "unknown"
	}
}

func projectEverkeep(status SystemStatus, evidence []EverkeepContinuityRecord, now time.Time) SystemStatus {
	if len(evidence) == 0 {
		return status
	}

	required := map[string]bool{
		"backup_coverage":    false,
		"restore_capability": false,
		"recovery_freshness": false,
	}
	seen := make(map[string]bool, len(required))
	for _, record := range evidence {
		if _, relevant := required[record.Dimension]; !relevant {
			continue
		}
		if seen[record.Dimension] {
			status.RuntimeEvidence = "unverified"
			return status
		}
		seen[record.Dimension] = true
		if record.Producer == "" || record.Scope != searchScopeID || record.ObservedAt.IsZero() ||
			record.ObservedAt.After(now) || record.VerificationMethod == "" {
			status.RuntimeEvidence = "unverified"
			return status
		}
		if record.RequiredEvidence && record.EvidenceReference == "" {
			status.RuntimeEvidence = "unverified"
			return status
		}
		if record.State == "ready" && record.FreshUntil.After(now) {
			required[record.Dimension] = true
		}
	}

	for _, ready := range required {
		if !ready {
			status.RuntimeEvidence = "partial"
			return status
		}
	}
	status.RuntimeEvidence = "current"
	status.State = "ready"
	status.PositiveClaim = true
	return status
}

func Handler(w http.ResponseWriter, r *http.Request) {
	HandlerWithSource(unavailableRuntimeEvidenceSource{})(w, r)
}

func HandlerWithSource(source RuntimeEvidenceSource) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-GoreeCloud-API-Version", APIVersion)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(SnapshotWithEvidence(source, time.Now())); err != nil {
			log.Printf("encode platform status response: %v", err)
		}
	}
}
