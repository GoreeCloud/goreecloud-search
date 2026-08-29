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

// validHistoryEnvelope enforces Search's negotiated Sync envelope boundary.
// Tombstones deliberately carry no application payload; live records must.
func validHistoryEnvelope(envelope Envelope) bool {
	capability, ok := searchHistoryCapability()
	if !ok || envelope.Dataset != capability.Dataset ||
		envelope.SchemaVersion != capability.SchemaVersion ||
		envelope.RecordID == "" || len(envelope.RecordID) > maxSyncRecordIDBytes ||
		envelope.Revision == 0 || envelope.UpdatedAt.IsZero() ||
		strings.TrimSpace(envelope.OriginDevice) == "" {
		return false
	}
	if envelope.Deleted {
		return capability.Delete && envelope.Payload == nil
	}
	return capability.Write && envelope.Payload != nil
}
