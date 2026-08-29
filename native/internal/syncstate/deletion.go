package syncstate

import (
	"crypto/ed25519"
	"encoding/base64"
	"strings"
	"time"
)

// SignedHistoryTombstone creates a payload-free deletion envelope so removed
// Search history can converge without retaining the deleted query payload.
func SignedHistoryTombstone(recordID string, revision uint64, updatedAt time.Time, identity DeviceIdentity) (Envelope, RecordProof, error) {
	recordID = strings.TrimSpace(recordID)
	capability, ok := searchHistoryCapability()
	if !ok || !capability.Delete || recordID == "" || len(recordID) > maxSyncRecordIDBytes || revision == 0 || updatedAt.IsZero() {
		return Envelope{}, RecordProof{}, ErrInvalidSyncRecord
	}
	if strings.TrimSpace(identity.DeviceID) == "" || len(identity.PublicKey) != ed25519.PublicKeySize || len(identity.PrivateKey) != ed25519.PrivateKeySize {
		return Envelope{}, RecordProof{}, ErrInvalidDeviceIdentity
	}
	envelope := Envelope{
		Dataset: capability.Dataset, SchemaVersion: capability.SchemaVersion, RecordID: recordID,
		Revision: revision, UpdatedAt: updatedAt.UTC(), OriginDevice: identity.DeviceID,
		Deleted: true,
	}
	if !validHistoryEnvelope(envelope) {
		return Envelope{}, RecordProof{}, ErrInvalidSyncRecord
	}
	message, err := proofMessage(envelope)
	if err != nil {
		return Envelope{}, RecordProof{}, err
	}
	proof := RecordProof{
		DeviceID: identity.DeviceID,
		PublicKey: base64.RawURLEncoding.EncodeToString(identity.PublicKey),
		Signature: base64.RawURLEncoding.EncodeToString(ed25519.Sign(identity.PrivateKey, message)),
	}
	return envelope, proof, nil
}
