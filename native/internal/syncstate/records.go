package syncstate

import (
	"errors"
	"strings"
	"time"

	"github.com/GoreeCloud/goreecloud-search/native/internal/preferences"
)

var (
	ErrInvalidSyncRecord       = errors.New("invalid Search sync record")
	ErrPreferenceNotSyncable   = errors.New("preference is not account-syncable")
)

// Record is Search's application-owned payload representation before GoreeCloud
// Sync adds transport metadata such as device, revision and conflict state.
type Record struct {
	Dataset       string         `json:"dataset"`
	SchemaVersion int            `json:"schemaVersion"`
	RecordID      string         `json:"recordId"`
	Payload       map[string]any `json:"payload"`
}

// NewHistoryRecord serializes the minimum useful search-history state. Search
// result bodies, provider credentials and request headers are intentionally not
// included in synchronized history records.
func NewHistoryRecord(recordID, query, category string, executedAt time.Time) (Record, error) {
	recordID = strings.TrimSpace(recordID)
	query = strings.TrimSpace(query)
	category = strings.TrimSpace(category)
	if recordID == "" || query == "" || category == "" || executedAt.IsZero() {
		return Record{}, ErrInvalidSyncRecord
	}
	return Record{
		Dataset: "search.history", SchemaVersion: 1, RecordID: recordID,
		Payload: map[string]any{
			"query": query,
			"category": category,
			"executedAt": executedAt.UTC().Format(time.RFC3339Nano),
		},
	}, nil
}

// NewPreferenceRecord fails closed for local and deployment-scoped settings.
// A preference becomes syncable only after Search explicitly declares it
// account-scoped in its preference contract.
func NewPreferenceRecord(key string, value any) (Record, error) {
	key = strings.TrimSpace(key)
	definition, ok := preferences.Find(key)
	if !ok || definition.Scope != preferences.ScopeAccount {
		return Record{}, ErrPreferenceNotSyncable
	}
	return Record{
		Dataset: "search.preferences", SchemaVersion: 1, RecordID: key,
		Payload: map[string]any{"key": key, "value": value},
	}, nil
}
