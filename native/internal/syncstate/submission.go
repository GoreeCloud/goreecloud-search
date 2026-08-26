package syncstate

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrInvalidDeviceIdentity = errors.New("invalid Search sync device identity")
	ErrSyncSubmissionFailed  = errors.New("Search history sync submission failed")
)

type Envelope struct {
	Dataset       string         `json:"dataset"`
	SchemaVersion int            `json:"schemaVersion"`
	RecordID      string         `json:"recordId"`
	Revision      uint64         `json:"revision"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	OriginDevice  string         `json:"originDevice"`
	Deleted       bool           `json:"deleted"`
	Payload       map[string]any `json:"payload,omitempty"`
}

type RecordProof struct {
	DeviceID  string `json:"deviceId"`
	PublicKey string `json:"publicKey"`
	Signature string `json:"signature"`
}

type DeviceIdentity struct {
	DeviceID   string
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

func SignedHistoryEnvelope(record Record, revision uint64, updatedAt time.Time, identity DeviceIdentity) (Envelope, RecordProof, error) {
	if record.Dataset != "search.history" || record.SchemaVersion < 1 || record.RecordID == "" || record.Payload == nil || revision == 0 || updatedAt.IsZero() {
		return Envelope{}, RecordProof{}, ErrInvalidSyncRecord
	}
	if strings.TrimSpace(identity.DeviceID) == "" || len(identity.PublicKey) != ed25519.PublicKeySize || len(identity.PrivateKey) != ed25519.PrivateKeySize {
		return Envelope{}, RecordProof{}, ErrInvalidDeviceIdentity
	}
	envelope := Envelope{
		Dataset: record.Dataset, SchemaVersion: record.SchemaVersion, RecordID: record.RecordID,
		Revision: revision, UpdatedAt: updatedAt.UTC(), OriginDevice: identity.DeviceID,
		Payload: clonePayload(record.Payload),
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

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type SubmissionClient struct {
	BaseURL     string
	BearerToken string
	Client      HTTPDoer
}

func (c SubmissionClient) SubmitHistory(ctx context.Context, envelope Envelope, proof RecordProof) error {
	if strings.TrimSpace(c.BaseURL) == "" || c.Client == nil {
		return ErrSyncSubmissionFailed
	}
	body, err := json.Marshal(struct {
		Record Envelope    `json:"record"`
		Proof  RecordProof `json:"proof"`
	}{Record: envelope, Proof: proof})
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/api/v1/sync/search/history", bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if token := strings.TrimSpace(c.BearerToken); token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := c.Client.Do(request)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSyncSubmissionFailed, err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
	if response.StatusCode != http.StatusAccepted {
		return fmt.Errorf("%w: status %d", ErrSyncSubmissionFailed, response.StatusCode)
	}
	return nil
}

func proofMessage(record Envelope) ([]byte, error) {
	payload, err := json.Marshal(record.Payload)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(payload)
	return []byte(fmt.Sprintf(
		"GC-SYNC-RECORD/1\n%s\n%d\n%s\n%d\n%s\n%s\n%t\n%s",
		record.Dataset,
		record.SchemaVersion,
		record.RecordID,
		record.Revision,
		record.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		record.OriginDevice,
		record.Deleted,
		hex.EncodeToString(digest[:]),
	)), nil
}

func clonePayload(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
