package webui

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
)

type glazeCSSClosureManifest struct {
	SchemaVersion      int               `json:"schema_version"`
	AuthorityRevision string            `json:"authority_revision"`
	Root              string            `json:"root"`
	Complete          bool              `json:"complete"`
	Active            bool              `json:"active"`
	StagedFiles       map[string]string `json:"staged_files"`
	PendingLeafImports []string          `json:"pending_leaf_imports"`
}

func gitBlobSHA(content []byte) string {
	header := []byte(fmt.Sprintf("blob %d\x00", len(content)))
	sum := sha1.Sum(append(header, content...))
	return hex.EncodeToString(sum[:])
}

func TestGlazeCanonicalCSSStageOneIsExactAndInactive(t *testing.T) {
	var manifest glazeCSSClosureManifest
	if err := json.Unmarshal([]byte(mustAsset("assets/glaze/canonical/css/closure.json")), &manifest); err != nil {
		t.Fatalf("decode canonical CSS closure manifest: %v", err)
	}
	if manifest.SchemaVersion != 1 || manifest.AuthorityRevision != "84cb3db4884042f0fa25ed6d475a127fb110f596" {
		t.Fatalf("unexpected canonical CSS authority: schema=%d revision=%q", manifest.SchemaVersion, manifest.AuthorityRevision)
	}
	if manifest.Root != "glaze-v1.4.0.css" {
		t.Fatalf("canonical CSS root = %q", manifest.Root)
	}
	if manifest.Complete || manifest.Active {
		t.Fatal("incomplete canonical CSS closure must remain inactive")
	}
	if len(manifest.PendingLeafImports) == 0 {
		t.Fatal("stage one must declare unresolved leaf imports")
	}

	for name, wantSHA := range manifest.StagedFiles {
		content, err := assets.ReadFile("assets/glaze/canonical/css/" + name)
		if err != nil {
			t.Fatalf("read staged canonical Glaze file %q: %v", name, err)
		}
		if got := gitBlobSHA(content); got != wantSHA {
			t.Fatalf("canonical Glaze file %q blob SHA = %s, want %s", name, got, wantSHA)
		}
	}
}
