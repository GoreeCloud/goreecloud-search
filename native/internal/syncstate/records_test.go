package syncstate

import (
	"errors"
	"testing"
	"time"
)

func TestHistoryRecordMinimizesSynchronizedPayload(t *testing.T) {
	when := time.Unix(100, 200).UTC()
	record, err := NewHistoryRecord("history-1", " goreecloud browser ", "general", when)
	if err != nil {
		t.Fatalf("NewHistoryRecord: %v", err)
	}
	if record.Dataset != "search.history" || record.SchemaVersion != 1 || record.RecordID != "history-1" {
		t.Fatalf("unexpected record metadata: %+v", record)
	}
	if len(record.Payload) != 3 {
		t.Fatalf("payload fields = %d, want 3: %+v", len(record.Payload), record.Payload)
	}
	if record.Payload["query"] != "goreecloud browser" || record.Payload["category"] != "general" {
		t.Fatalf("unexpected history payload: %+v", record.Payload)
	}
	for _, forbidden := range []string{"results", "credentials", "headers", "ip", "userAgent"} {
		if _, ok := record.Payload[forbidden]; ok {
			t.Fatalf("history payload unexpectedly includes %q", forbidden)
		}
	}
}

func TestLocalPreferenceFailsClosedForSync(t *testing.T) {
	_, err := NewPreferenceRecord("appearance.theme", "dark")
	if !errors.Is(err, ErrPreferenceNotSyncable) {
		t.Fatalf("error = %v, want %v", err, ErrPreferenceNotSyncable)
	}
}

func TestUnknownPreferenceFailsClosedForSync(t *testing.T) {
	_, err := NewPreferenceRecord("privacy.unknown", true)
	if !errors.Is(err, ErrPreferenceNotSyncable) {
		t.Fatalf("error = %v, want %v", err, ErrPreferenceNotSyncable)
	}
}
