package syncstate

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"
)

func TestSignedHistoryTombstoneIsPayloadFree(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	envelope, proof, err := SignedHistoryTombstone("history-1", 3, time.Unix(100, 0).UTC(), DeviceIdentity{
		DeviceID: "device-a", PublicKey: publicKey, PrivateKey: privateKey,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !envelope.Deleted || envelope.Payload != nil || envelope.Dataset != "search.history" {
		t.Fatalf("unexpected tombstone: %+v", envelope)
	}
	if proof.DeviceID != "device-a" || proof.Signature == "" {
		t.Fatalf("unexpected proof: %+v", proof)
	}
}
