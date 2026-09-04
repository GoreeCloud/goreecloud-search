package buildinfo

import (
	"regexp"
	"runtime/debug"
	"strconv"
)

var canonicalGitRevision = regexp.MustCompile(`^[0-9a-f]{40}$`)

// Provenance is the intentionally minimized runtime build identity safe to
// expose through GoreeCloud Search status. It contains no build paths, module
// cache information, environment data, credentials, or other BuildInfo fields.
type Provenance struct {
	ProvenanceAvailable bool  `json:"provenance_available"`
	SourceRevision      string `json:"source_revision,omitempty"`
	SourceModified      *bool  `json:"source_modified,omitempty"`
}

// Current returns the exact Git revision embedded by the Go toolchain when it
// is present and structurally valid. Missing or malformed VCS metadata remains
// explicitly unavailable rather than inventing a source identity.
func Current() Provenance {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return Provenance{}
	}
	return fromSettings(info.Settings)
}

func fromSettings(settings []debug.BuildSetting) Provenance {
	var revision string
	var modifiedValue string
	for _, setting := range settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modifiedValue = setting.Value
		}
	}
	if !canonicalGitRevision.MatchString(revision) {
		return Provenance{}
	}
	modified, err := strconv.ParseBool(modifiedValue)
	if err != nil {
		return Provenance{}
	}
	return Provenance{
		ProvenanceAvailable: true,
		SourceRevision:      revision,
		SourceModified:      &modified,
	}
}
