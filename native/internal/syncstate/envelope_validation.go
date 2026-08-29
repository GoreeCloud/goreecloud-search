package syncstate

import "strings"

const (
	searchHistoryDataset = "search.history"
	maxSyncRecordIDBytes = 512
)

func searchHistoryCapability() (Capability, bool) {
	for _, capability := range capabilities {
		if capability.Dataset == searchHistoryDataset {
			return capability, true
		}
	}
	return Capability{}, false
}

// validHistoryEnvelope validates the direction-neutral negotiated envelope
// shape. Read/write/delete permissions are enforced at the operation boundary.
func validHistoryEnvelope(envelope Envelope) bool {
	capability, ok := searchHistoryCapability()
	if !ok || envelope.Dataset != capability.Dataset ||
		envelope.SchemaVersion != capability.SchemaVersion ||
		envelope.RecordID == "" || len(envelope.RecordID) > maxSyncRecordIDBytes ||
		envelope.Revision == 0 || envelope.UpdatedAt.IsZero() ||
		strings.TrimSpace(envelope.OriginDevice) == "" {
		return false
	}
	// Privacy Shield data minimization: tombstones carry no application payload;
	// live records carry application state.
	if envelope.Deleted {
		return envelope.Payload == nil
	}
	return envelope.Payload != nil
}

func canSubmitHistoryEnvelope(envelope Envelope) bool {
	capability, ok := searchHistoryCapability()
	if !ok || !validHistoryEnvelope(envelope) {
		return false
	}
	if envelope.Deleted {
		return capability.Delete
	}
	return capability.Write
}
