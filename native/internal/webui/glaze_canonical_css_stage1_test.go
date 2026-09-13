package webui

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
)

type glazeCSSClosureManifest struct {
	SchemaVersion       int               `json:"schema_version"`
	AuthorityRevision  string            `json:"authority_revision"`
	Root               string            `json:"root"`
	Complete           bool              `json:"complete"`
	Active             bool              `json:"active"`
	StagedFiles        map[string]string `json:"staged_files"`
	PendingLeafImports []string          `json:"pending_leaf_imports"`
}

var canonicalCSSImport = regexp.MustCompile(`@import\s+url\(["']?\./([^"')]+)["']?\)`)

func gitBlobSHA(content []byte) string {
	header := []byte(fmt.Sprintf("blob %d\x00", len(content)))
	sum := sha1.Sum(append(header, content...))
	return hex.EncodeToString(sum[:])
}

func TestGlazeCanonicalCSSClosureIsExactAndFailClosed(t *testing.T) {
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
	if _, ok := manifest.StagedFiles[manifest.Root]; !ok {
		t.Fatalf("canonical CSS root %q is not hash-bound in staged_files", manifest.Root)
	}
	if manifest.Active && !manifest.Complete {
		t.Fatal("canonical CSS closure cannot be active before it is complete")
	}
	if manifest.Complete && len(manifest.PendingLeafImports) != 0 {
		t.Fatalf("complete canonical CSS closure still declares %d pending imports", len(manifest.PendingLeafImports))
	}
	if !manifest.Complete && manifest.Active {
		t.Fatal("incomplete canonical CSS closure must remain inactive")
	}

	pending := make(map[string]struct{}, len(manifest.PendingLeafImports))
	for _, name := range manifest.PendingLeafImports {
		if name == "" {
			t.Fatal("canonical CSS closure contains a blank pending import")
		}
		if _, exists := pending[name]; exists {
			t.Fatalf("canonical CSS closure contains duplicate pending import %q", name)
		}
		if _, staged := manifest.StagedFiles[name]; staged {
			t.Fatalf("canonical CSS file %q cannot be both staged and pending", name)
		}
		pending[name] = struct{}{}
	}

	referencedPending := make(map[string]struct{}, len(pending))
	for name, wantSHA := range manifest.StagedFiles {
		content, err := assets.ReadFile("assets/glaze/canonical/css/" + name)
		if err != nil {
			t.Fatalf("read staged canonical Glaze file %q: %v", name, err)
		}
		if got := gitBlobSHA(content); got != wantSHA {
			t.Fatalf("canonical Glaze file %q blob SHA = %s, want %s", name, got, wantSHA)
		}

		for _, match := range canonicalCSSImport.FindAllSubmatch(content, -1) {
			dependency := string(match[1])
			if _, staged := manifest.StagedFiles[dependency]; staged {
				continue
			}
			if _, declared := pending[dependency]; !declared {
				t.Fatalf("canonical Glaze file %q imports unstaged dependency %q that is not declared pending", name, dependency)
			}
			referencedPending[dependency] = struct{}{}
		}
	}

	for name := range pending {
		if _, referenced := referencedPending[name]; !referenced {
			t.Fatalf("canonical CSS pending import %q is not referenced by any staged canonical file", name)
		}
	}

	if !manifest.Complete && len(pending) == 0 {
		t.Fatal("incomplete canonical CSS closure must declare unresolved imports")
	}
}
