package syncstate

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type doerFunc func(*http.Request) (*http.Response, error)

func (f doerFunc) Do(request *http.Request) (*http.Response, error) { return f(request) }

func TestSignedHistoryEnvelopeAndSubmission(t *testing.T) {
	capability, ok := searchHistoryCapability()
	if !ok {
		t.Fatal("search.history capability missing")
	}
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	history, err := NewHistoryRecord("history-1", "goreecloud", "general", time.Unix(100, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	envelope, proof, err := SignedHistoryEnvelope(history, 2, time.Unix(101, 0).UTC(), DeviceIdentity{
		DeviceID: "device-a", PublicKey: publicKey, PrivateKey: privateKey,
	})
	if err != nil {
		t.Fatal(err)
	}
	if envelope.Dataset != capability.Dataset || envelope.SchemaVersion != capability.SchemaVersion || envelope.Revision != 2 || proof.DeviceID != "device-a" || proof.Signature == "" {
		t.Fatalf("unexpected signed envelope: envelope=%+v proof=%+v", envelope, proof)
	}

	client := SubmissionClient{
		BaseURL:     "https://sync.internal",
		BearerToken: "session-token",
		Client: doerFunc(func(request *http.Request) (*http.Response, error) {
			if request.URL.Path != "/api/v1/sync/search/history" || request.Method != http.MethodPost {
				t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
			}
			if got := request.Header.Get("Authorization"); got != "Bearer session-token" {
				t.Fatalf("Authorization = %q", got)
			}
			var payload struct {
				Record Envelope    `json:"record"`
				Proof  RecordProof `json:"proof"`
			}
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if payload.Record.RecordID != "history-1" || payload.Proof.Signature == "" {
				t.Fatalf("unexpected payload: %+v", payload)
			}
			return &http.Response{
				StatusCode: http.StatusAccepted,
				Body:       io.NopCloser(strings.NewReader(`{"accepted":true}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}
	if err := client.SubmitHistory(context.Background(), envelope, proof); err != nil {
		t.Fatalf("SubmitHistory: %v", err)
	}
}

func TestSignedHistoryEnvelopeRejectsOversizedRecordID(t *testing.T) {
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	history, err := NewHistoryRecord(strings.Repeat("r", maxSyncRecordIDBytes+1), "goreecloud", "general", time.Unix(100, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := SignedHistoryEnvelope(history, 1, time.Unix(101, 0).UTC(), DeviceIdentity{
		DeviceID: "device-a", PublicKey: publicKey, PrivateKey: privateKey,
	}); err == nil {
		t.Fatal("oversized record ID must fail before signing")
	}
}

func TestSignedHistoryEnvelopeRejectsUnnegotiatedSchema(t *testing.T) {
	capability, ok := searchHistoryCapability()
	if !ok {
		t.Fatal("search.history capability missing")
	}
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	record := Record{
		Dataset: capability.Dataset, SchemaVersion: capability.SchemaVersion + 1,
		RecordID: "history-1", Payload: map[string]any{"query": "goreecloud"},
	}
	if _, _, err := SignedHistoryEnvelope(record, 1, time.Unix(101, 0).UTC(), DeviceIdentity{
		DeviceID: "device-a", PublicKey: publicKey, PrivateKey: privateKey,
	}); err == nil {
		t.Fatal("unnegotiated schema must fail before signing")
	}
}

func TestSubmitHistoryRequiresBearerBeforeTransport(t *testing.T) {
	called := false
	client := SubmissionClient{
		BaseURL: "https://sync.internal",
		Client: doerFunc(func(*http.Request) (*http.Response, error) {
			called = true
			return nil, nil
		}),
	}
	if err := client.SubmitHistory(context.Background(), Envelope{RecordID: "history-1"}, RecordProof{}); err == nil {
		t.Fatal("missing bearer session must fail closed")
	}
	if called {
		t.Fatal("transport must not be called without an authenticated Sync session")
	}
}

func TestSubmitHistoryRejectsNonconformingEnvelopeBeforeTransport(t *testing.T) {
	capability, ok := searchHistoryCapability()
	if !ok {
		t.Fatal("search.history capability missing")
	}
	base := Envelope{
		Dataset: capability.Dataset, SchemaVersion: capability.SchemaVersion,
		RecordID: "history-1", Revision: 1, UpdatedAt: time.Unix(101, 0).UTC(),
		OriginDevice: "device-a", Payload: map[string]any{"query": "goreecloud"},
	}
	tests := []struct {
		name   string
		mutate func(*Envelope)
	}{
		{name: "unnegotiated schema", mutate: func(envelope *Envelope) { envelope.SchemaVersion++ }},
		{name: "tombstone payload", mutate: func(envelope *Envelope) { envelope.Deleted = true }},
		{name: "live record without payload", mutate: func(envelope *Envelope) { envelope.Payload = nil }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			envelope := base
			test.mutate(&envelope)
			called := false
			client := SubmissionClient{
				BaseURL: "https://sync.internal", BearerToken: "session-token",
				Client: doerFunc(func(*http.Request) (*http.Response, error) {
					called = true
					return nil, nil
				}),
			}
			if err := client.SubmitHistory(context.Background(), envelope, RecordProof{}); err == nil {
				t.Fatal("nonconforming envelope must fail closed")
			}
			if called {
				t.Fatal("transport must not receive a nonconforming envelope")
			}
		})
	}
}

func TestSubmitHistoryRejectsInvalidProofBeforeTransport(t *testing.T) {
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	history, err := NewHistoryRecord("history-proof", "goreecloud", "general", time.Unix(100, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	baseEnvelope, baseProof, err := SignedHistoryEnvelope(history, 1, time.Unix(101, 0).UTC(), DeviceIdentity{
		DeviceID: "device-a", PublicKey: publicKey, PrivateKey: privateKey,
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		mutate func(*Envelope, *RecordProof)
	}{
		{name: "device mismatch", mutate: func(_ *Envelope, proof *RecordProof) { proof.DeviceID = "device-b" }},
		{name: "malformed public key", mutate: func(_ *Envelope, proof *RecordProof) { proof.PublicKey = "!" }},
		{name: "malformed signature", mutate: func(_ *Envelope, proof *RecordProof) { proof.Signature = "!" }},
		{name: "invalid signature", mutate: func(_ *Envelope, proof *RecordProof) {
			proof.Signature = base64.RawURLEncoding.EncodeToString(make([]byte, ed25519.SignatureSize))
		}},
		{name: "record changed after signing", mutate: func(envelope *Envelope, _ *RecordProof) { envelope.Revision++ }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			envelope := baseEnvelope
			proof := baseProof
			test.mutate(&envelope, &proof)
			called := false
			client := SubmissionClient{
				BaseURL: "https://sync.internal", BearerToken: "session-token",
				Client: doerFunc(func(*http.Request) (*http.Response, error) {
					called = true
					return nil, nil
				}),
			}
			if err := client.SubmitHistory(context.Background(), envelope, proof); err == nil {
				t.Fatal("invalid record proof must fail closed")
			}
			if called {
				t.Fatal("transport must not receive an invalid record proof")
			}
		})
	}
}

func TestSubmissionDoesNotCarryAuthoritativePolicyFields(t *testing.T) {
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	history, _ := NewHistoryRecord("history-2", "privacy", "general", time.Unix(200, 0).UTC())
	envelope, proof, _ := SignedHistoryEnvelope(history, 1, time.Unix(201, 0).UTC(), DeviceIdentity{
		DeviceID: "device-b", PublicKey: publicKey, PrivateKey: privateKey,
	})

	client := SubmissionClient{
		BaseURL:     "https://sync.internal",
		BearerToken: "session-token",
		Client: doerFunc(func(request *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatal(err)
			}
			text := string(body)
			for _, forbidden := range []string{"consentGranted", "purposeAllowed", "trusted", "evidenceId"} {
				if strings.Contains(text, forbidden) {
					t.Fatalf("request contains authoritative policy field %q: %s", forbidden, text)
				}
			}
			return &http.Response{StatusCode: http.StatusAccepted, Body: io.NopCloser(strings.NewReader("{}")), Header: make(http.Header)}, nil
		}),
	}
	if err := client.SubmitHistory(context.Background(), envelope, proof); err != nil {
		t.Fatal(err)
	}
}
