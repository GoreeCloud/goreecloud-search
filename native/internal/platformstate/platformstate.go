package platformstate

import (
	"encoding/json"
	"log"
	"net/http"
)

const APIVersion = "1"

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

func DevelopmentSnapshot() Snapshot {
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
				PositiveClaim:       false,
				ProductionAccepted:  false,
			},
			{
				ID:                  "wardveil-security",
				AuthorityRepository: "GoreeCloud/goreecloud-wardveil-security",
				AuthorityContract:   "contracts/wardveil.status.schema.json",
				SourceIntegration:   "present",
				RuntimeEvidence:     "unverified",
				State:               "unknown",
				PositiveClaim:       false,
				ProductionAccepted:  false,
			},
			{
				ID:                  "everkeep",
				AuthorityRepository: "GoreeCloud/goreecloud-everkeep",
				AuthorityContract:   "contracts/continuity.status.schema.json",
				SourceIntegration:   "presentation-boundary-only",
				RuntimeEvidence:     "unavailable",
				State:               "unknown",
				PositiveClaim:       false,
				ProductionAccepted:  false,
			},
		},
		ContainsUserContent: false,
		ContainsQueryText:   false,
		CredentialsExposed:  false,
		ProductionApproved:  false,
	}
}

func Handler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-GoreeCloud-API-Version", APIVersion)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(DevelopmentSnapshot()); err != nil {
		log.Printf("encode platform status response: %v", err)
	}
}
