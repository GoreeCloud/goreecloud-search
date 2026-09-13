package webui

import (
	"encoding/json"
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
		CurrentSourceProjection              string `json:"current_source_projection"`
		CanonicalEntrypointVendored         bool   `json:"canonical_entrypoint_vendored"`
		CanonicalRuntimeVendored            bool   `json:"canonical_runtime_vendored"`
		ProductionAccepted                  bool   `json:"production_accepted"`
		ManualDeviceQualificationComplete   bool   `json:"manual_device_qualification_complete"`
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
		t.Fatal("Glaze authority lifecycle is not current Stable consumer authority")
	}
	if manifest.AuthorityRepository != "GoreeCloud/goreecloud-glaze-ui" || manifest.AuthorityRevision != "84cb3db4884042f0fa25ed6d475a127fb110f596" {
		t.Fatalf("Glaze authority revision mismatch: %s@%s", manifest.AuthorityRepository, manifest.AuthorityRevision)
	}
	if manifest.RollbackBaseline != "1.3.0" {
		t.Fatalf("rollback baseline = %q, want 1.3.0", manifest.RollbackBaseline)
	}

	expected := map[string][2]string{
		"web": {"css/glaze-v1.4.0.css", "d48a9bc317090d152799769271de0fb4325494c4"},
		"runtime": {"js/glaze-v1.4.0.mjs", "a39e7f7209533f5c5b4fe78ce496307fc7850bfc"},
		"optical": {"js/glaze-v1.4-optical-engine.mjs", "96590bedcdb581a3c5c41e548c5866c7b6cd80a4"},
	}
	actual := map[string][2]string{
		"web": {manifest.Entrypoints.Web.Path, manifest.Entrypoints.Web.GitBlobSHA},
		"runtime": {manifest.Entrypoints.Runtime.Path, manifest.Entrypoints.Runtime.GitBlobSHA},
		"optical": {manifest.Entrypoints.OpticalEngine.Path, manifest.Entrypoints.OpticalEngine.GitBlobSHA},
	}
	for label, want := range expected {
		got := actual[label]
		if got != want {
			t.Fatalf("%s authority = %q @ %q, want %q @ %q", label, got[0], got[1], want[0], want[1])
		}
	}

	if manifest.ConsumerBoundary.CurrentSourceProjection != "assets/glaze-v1.4.css" {
		t.Fatalf("source projection = %q", manifest.ConsumerBoundary.CurrentSourceProjection)
	}
	if manifest.ConsumerBoundary.CanonicalEntrypointVendored || manifest.ConsumerBoundary.CanonicalRuntimeVendored || manifest.ConsumerBoundary.ProductionAccepted || manifest.ConsumerBoundary.ManualDeviceQualificationComplete {
		t.Fatal("Glaze authority manifest overstates Search adoption or acceptance")
	}

	projection := mustAsset(manifest.ConsumerBoundary.CurrentSourceProjection)
	for _, required := range []string{
		"GLAZE UI V1.4 / 1.4.0 consumer source projection",
		manifest.AuthorityRevision,
		"--glz14-frost-strength",
		"--glz14-memory-tint-influence",
		"prefers-reduced-transparency",
		"forced-colors",
	} {
		if !strings.Contains(projection, required) {
			t.Fatalf("current Glaze source projection is missing authority marker %q", required)
		}
	}
}
