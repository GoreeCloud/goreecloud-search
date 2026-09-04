package buildinfo

import (
	"runtime/debug"
	"testing"
)

func boolValue(value bool) *bool { return &value }

func TestFromSettingsAcceptsCanonicalCleanRevision(t *testing.T) {
	got := fromSettings([]debug.BuildSetting{
		{Key: "vcs.revision", Value: "0123456789abcdef0123456789abcdef01234567"},
		{Key: "vcs.modified", Value: "false"},
		{Key: "GOARCH", Value: "amd64"},
		{Key: "vcs.time", Value: "2026-09-04T00:00:00Z"},
	})
	if !got.ProvenanceAvailable {
		t.Fatal("canonical build provenance unexpectedly unavailable")
	}
	if got.SourceRevision != "0123456789abcdef0123456789abcdef01234567" {
		t.Fatalf("source revision = %q", got.SourceRevision)
	}
	if got.SourceModified == nil || *got.SourceModified {
		t.Fatalf("source modified = %v, want false", got.SourceModified)
	}
}

func TestFromSettingsPreservesModifiedState(t *testing.T) {
	got := fromSettings([]debug.BuildSetting{
		{Key: "vcs.revision", Value: "fedcba9876543210fedcba9876543210fedcba98"},
		{Key: "vcs.modified", Value: "true"},
	})
	if !got.ProvenanceAvailable || got.SourceModified == nil || !*got.SourceModified {
		t.Fatalf("modified provenance = %+v", got)
	}
}

func TestFromSettingsFailsClosedWhenMetadataIsMissingOrMalformed(t *testing.T) {
	tests := [][]debug.BuildSetting{
		nil,
		{{Key: "vcs.revision", Value: "0123456789abcdef0123456789abcdef01234567"}},
		{{Key: "vcs.modified", Value: "false"}},
		{
			{Key: "vcs.revision", Value: "not-a-git-revision"},
			{Key: "vcs.modified", Value: "false"},
		},
		{
			{Key: "vcs.revision", Value: "0123456789ABCDEF0123456789ABCDEF01234567"},
			{Key: "vcs.modified", Value: "false"},
		},
		{
			{Key: "vcs.revision", Value: "0123456789abcdef0123456789abcdef01234567"},
			{Key: "vcs.modified", Value: "unknown"},
		},
	}
	for _, settings := range tests {
		got := fromSettings(settings)
		if got.ProvenanceAvailable || got.SourceRevision != "" || got.SourceModified != nil {
			t.Fatalf("invalid provenance unexpectedly exposed: %+v", got)
		}
	}
}
