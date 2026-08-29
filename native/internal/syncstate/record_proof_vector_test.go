package syncstate

import (
	"testing"
	"time"
)

func TestRecordProofCanonicalVector(t *testing.T) {
	updatedAt, err := time.Parse(time.RFC3339Nano, "2026-08-28T21:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	envelope := Envelope{
		Dataset: "protocol.vector", SchemaVersion: 1, RecordID: "record-vector-1", Revision: 7,
		UpdatedAt: updatedAt, OriginDevice: "device-vector", Deleted: false,
		Payload: map[string]any{"query": "goreecloud", "scope": "general"},
	}
	proof := RecordProof{
		DeviceID: "device-vector",
		PublicKey: "A6EHv_POEL4dcN0Y50vAmWfk1jCbpQ1fHdyGZBJVMbg",
		Signature: "tnC7VLxc3KABo_kCVMzX1GBbq139gnO276fpvWdPhx1burKdMG9nCfbY3AJW0hnHVsrJRwUSww-rIWBHgsapAw",
	}
	const expectedMessage = "GC-SYNC-RECORD/1\nprotocol.vector\n1\nrecord-vector-1\n7\n2026-08-28T21:00:00Z\ndevice-vector\nfalse\nffb813686fb3ae7233ff7516949e0280523ab84efc6f17e89a9f8d12cb4b4e11"

	message, err := proofMessage(envelope)
	if err != nil {
		t.Fatalf("proofMessage: %v", err)
	}
	if string(message) != expectedMessage {
		t.Fatalf("canonical proof message drifted:\n%s", message)
	}
	if !validRecordProof(envelope, proof) {
		t.Fatal("canonical goreecloud-sync record-proof vector must verify")
	}

	tampered := envelope
	tampered.RecordID = "record-vector-tampered"
	if validRecordProof(tampered, proof) {
		t.Fatal("canonical proof must not verify a mutated record")
	}
}
