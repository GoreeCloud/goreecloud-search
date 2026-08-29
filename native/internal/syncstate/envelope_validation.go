package syncstate

import "strings"

const (
	searchHistoryDataset       = "search.history"
	searchHistorySchemaVersion = 1
	maxSyncRecordIDBytes       = 512
)

// validHistoryEnvelope enforces Search's negotiated Sync envelope boundary.
// Tombstones deliberately carry no application payload; live records must.
func validHistoryEnvelope(envelope Envelope) bool {
	if envelope.Dataset != searchHistoryDataset ||
		envelope.SchemaVersion != searchHistorySchemaVersion ||
		envelope.RecordID == "" || len(envelope.RecordID) > maxSyncRecordIDBytes ||
		envelope.Revision == 0 || envelope.UpdatedAt.IsZero() ||
		strings.TrimSpace(envelope.OriginDevice) == "" {
		return false
	}
	if envelope.Deleted {
		return envelope.Payload == nil
	}
	return envelope.Payload != nil
}
