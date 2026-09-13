package webui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type glazeAuthorityManifest struct {
	SchemaVersion       int    `json:"schema_version"`
	Version             string `json:"version"`
	Lifecycle           string `json:"lifecycle"`
	ConsumerEligible    bool   `json:"consumer_eligible"`
	AuthorityRepository string `json:"authority_repository"`
	AuthorityRevision   string `json:"authority_revision"`
	RollbackBaseline    string `json:"rollback_baseline"`
	Entrypoints         struct {
		Web struct {
			Path       string `json:"path"`
			GitBlobSHA string `json:"git_blob_sha"`
		} `json:"web"`
		Runtime struct {
			Path       string `json:"path"`
			GitBlobSHA string `json:"git_blob_sha"`
		} `json:"runtime"`
		OpticalEngine struct {
			Path       string `json:"path"`
			GitBlobSHA string `json:"git_blob_sha"`
		} `json:"optical_engine"`
	} `json:"entrypoints"`
	ConsumerBoundary struct {
		CanonicalEntrypointVendored         bool `json:"canonical_entrypoint_vendored"`
		CanonicalRuntimeVendored            bool `json:"canonical_runtime_vendored"`
		ProductionAccepted                  bool `json:"production_accepted"`
		ManualDeviceQualificationComplete   bool `json:"manual_device_qualification_complete"`
	} `json:"consumer_boundary"`
}

func TestGlazeV14AuthorityManifestBindsExactStableRelease(t *testing.T) {
	var manifest glazeAuthorityManifest
	if err := json.Unmarshal([]byte(mustAsset("assets/glaze-v1.4-authority.json")), &manifest); err != nil {
		t.Fatalf("decode Glaze authority manifest: %v", err)
	}

	if manifest.SchemaVersion != 1 || manifest.Version != glazeVersion {
		t.Fatalf("Glaze authority manifest version mismatch: schema=%d version=%q", manifest.SchemaVersion, manifest.Version)
	}
	if manifest.Lifecycle != "stable" || !manifest.ConsumerEligible {
		t.Fatalf("Glaze authority lifecycle is not current Stable consumer authority")
	}
	if manifest.AuthorityRepository != "GoreeCloud/goreecloud-glaze-ui" || manifest.AuthorityRevision != "84cb3db4884042f0fa25ed6d475a127fb110f596" {
		t.Fatalf("Glaze authority revision mismatch: %s@%s", manifest.AuthorityRepository, manifest.AuthorityRevision)
	}
	if manifest.RollbackBaseline != "1.3.0" {
		t.Fatalf("rollback baseline = %q, want 1.3.0", manifest.RollbackBaseline)
	}

	for label, got := range map[string]string{
		"web path": manifest.Entrypoints.Web.Path,
		"web blob": manifest.Entrypoints.Web.GitBlobSHA,
		"runtime path": manifest.Entrypoints.Runtime.Path,
		"runtime blob": manifest.Entrypoints.Runtime.GitBlobSHA,
		"optical path": manifest.Entrypoints.OpticalEngine.Path,
		"optical blob": manifest.Entrypoints.OpticalEngine.GitBlobSHA,
	} {
		if strings.TrimSpace(got) == "" {
			t.Fatalf("Glaze authority manifest missing %s", label)
		}
	}

	if manifest.ConsumerBoundary.CanonicalEntrypointVendored || manifest.ConsumerBoundary.CanonicalRuntimeVendored || manifest.ConsumerBoundary.ProductionAccepted || manifest.ConsumerBoundary.ManualDeviceQualificationComplete {
		t.Fatal("Glaze authority manifest overstates Search adoption or acceptance")
	}
}

func TestGlazeAuthorityEndpointServesLocalMachineReadableProvenance(t *testing.T) {
	recorder := httptest.NewRecorder()
	GlazeAuthority(recorder, httptest.NewRequest(http.MethodGet, "/assets/glaze-v1.4-authority.json", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	if !strings.Contains(recorder.Body.String(), `"authority_revision": "84cb3db4884042f0fa25ed6d475a127fb110f596"`) {
		t.Fatal("authority endpoint did not expose exact reviewed Glaze release revision")
	}
}
