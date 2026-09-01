package platformstate

import (
	"regexp"
	"time"
)

const meshEvidenceEnvelopeVersion = "goreecloud.evidence-envelope.v1"

var (
	meshRevisionPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	meshDigestPattern   = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

type MeshEvidenceProducer struct {
	System     string
	Repository string
	Revision   string
	Contract   string
}

type MeshEvidenceSubject struct {
	Kind  string
	ID    string
	Scope string
}

type MeshEvidenceEnvelope struct {
	Version                string
	ID                     string
	Producer               MeshEvidenceProducer
	AuthorityDomain        string
	Subject                MeshEvidenceSubject
	Assertion              string
	Outcome                string
	Source                 string
	ObservedAt             time.Time
	ValidUntil             time.Time
	DataClass              string
	PayloadDigest          string
	ContainsUserContent    bool
	ContainsSecretMaterial bool
}

type MeshEvidenceExpectation struct {
	ProducerSystem  string
	Repository      string
	Contract        string
	AuthorityDomain string
	SubjectKind     string
	SubjectID       string
	Assertion       string
	Source          string
	ObservedAt      time.Time
	ValidUntil      time.Time
}

func ValidateMeshEvidenceEnvelope(
	envelope *MeshEvidenceEnvelope,
	expectation MeshEvidenceExpectation,
	now time.Time,
) bool {
	if envelope == nil || now.IsZero() {
		return false
	}
	if envelope.Version != meshEvidenceEnvelopeVersion ||
		!validBoundedEnvelopeText(envelope.ID, 128) ||
		envelope.Producer.System != expectation.ProducerSystem ||
		envelope.Producer.Repository != expectation.Repository ||
		!meshRevisionPattern.MatchString(envelope.Producer.Revision) ||
		envelope.Producer.Contract != expectation.Contract ||
		envelope.AuthorityDomain != expectation.AuthorityDomain ||
		envelope.Subject.Kind != expectation.SubjectKind ||
		envelope.Subject.ID != expectation.SubjectID ||
		envelope.Assertion != expectation.Assertion ||
		envelope.Source != expectation.Source ||
		!validBoundedEnvelopeText(envelope.Outcome, 256) ||
		!validEvidenceDataClass(envelope.DataClass) ||
		envelope.ContainsUserContent ||
		envelope.ContainsSecretMaterial {
		return false
	}
	if !validBoundedEnvelopeText(envelope.Subject.ID, 240) ||
		!validBoundedEnvelopeText(envelope.Source, 512) {
		return false
	}
	if envelope.Subject.Scope != "" && !validBoundedEnvelopeText(envelope.Subject.Scope, 240) {
		return false
	}
	if envelope.PayloadDigest != "" && !meshDigestPattern.MatchString(envelope.PayloadDigest) {
		return false
	}
	if envelope.ObservedAt.IsZero() || envelope.ValidUntil.IsZero() ||
		envelope.ObservedAt.After(now) || !envelope.ValidUntil.After(envelope.ObservedAt) {
		return false
	}
	if !expectation.ObservedAt.IsZero() && !envelope.ObservedAt.Equal(expectation.ObservedAt) {
		return false
	}
	if !expectation.ValidUntil.IsZero() && !envelope.ValidUntil.Equal(expectation.ValidUntil) {
		return false
	}
	return true
}

func validEvidenceDataClass(value string) bool {
	switch value {
	case "public", "operational", "derived":
		return true
	default:
		return false
	}
}

func validBoundedEnvelopeText(value string, max int) bool {
	if value == "" || len(value) > max {
		return false
	}
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return false
		}
	}
	return true
}
