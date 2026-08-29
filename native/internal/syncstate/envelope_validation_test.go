package syncstate

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFetchHistoryRejectsNonconformingEnvelopeShape(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "unnegotiated schema",
			body: `{"dataset":"search.history","count":1,"records":[{"dataset":"search.history","schemaVersion":2,"recordId":"history-1","revision":1,"updatedAt":"2026-08-28T20:00:00Z","originDevice":"device-1","deleted":false,"payload":{"query":"one"}}]}`,
		},
		{
			name: "tombstone retains payload",
			body: `{"dataset":"search.history","count":1,"records":[{"dataset":"search.history","schemaVersion":1,"recordId":"history-1","revision":1,"updatedAt":"2026-08-28T20:00:00Z","originDevice":"device-1","deleted":true,"payload":{"query":"must-not-survive"}}]}`,
		},
		{
			name: "live record missing payload",
			body: `{"dataset":"search.history","count":1,"records":[{"dataset":"search.history","schemaVersion":1,"recordId":"history-1","revision":1,"updatedAt":"2026-08-28T20:00:00Z","originDevice":"device-1","deleted":false}]}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := RetrievalClient{
				BaseURL: "https://sync.goreecloud.test", BearerToken: "session-token",
				Client: retrievalDoerFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(test.body)), Header: make(http.Header)}, nil
				}),
			}
			if _, err := client.FetchHistory(context.Background()); err == nil {
				t.Fatal("nonconforming retrieval envelope must fail closed")
			}
		})
	}
}

func TestFetchHistoryAcceptsPayloadFreeTombstone(t *testing.T) {
	body := `{"dataset":"search.history","count":1,"records":[{"dataset":"search.history","schemaVersion":1,"recordId":"history-1","revision":2,"updatedAt":"2026-08-28T20:00:00Z","originDevice":"device-1","deleted":true}]}`
	client := RetrievalClient{
		BaseURL: "https://sync.goreecloud.test", BearerToken: "session-token",
		Client: retrievalDoerFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
		}),
	}
	records, err := client.FetchHistory(context.Background())
	if err != nil {
		t.Fatalf("FetchHistory: %v", err)
	}
	if len(records) != 1 || !records[0].Deleted || records[0].Payload != nil {
		t.Fatalf("unexpected tombstone: %+v", records)
	}
}
