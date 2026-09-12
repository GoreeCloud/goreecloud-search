package webautomation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestClassifyTinyFishCompletedOutcomeRecognizesStructuredFailure(t *testing.T) {
	tests := []struct {
		name string
		raw  json.RawMessage
		want error
	}{
		{name: "goal success false", raw: json.RawMessage(`{"goal_succeeded":false}`), want: ErrGoalFailed},
		{name: "nested success false", raw: json.RawMessage(`{"result":{"success":false}}`), want: ErrGoalFailed},
		{name: "blocked status", raw: json.RawMessage(`{"status":"blocked"}`), want: ErrSiteBlocked},
		{name: "captcha outcome", raw: json.RawMessage(`{"outcome":"captcha"}`), want: ErrSiteBlocked},
		{name: "authentication required", raw: json.RawMessage(`{"result_status":"authentication-required"}`), want: ErrGoalFailed},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := classifyTinyFishCompletedOutcome(test.raw); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestClassifyTinyFishCompletedOutcomeRecognizesUnambiguousTextFailure(t *testing.T) {
	tests := []struct {
		name string
		text string
		want error
	}{
		{name: "missing credentials", text: "The task cannot be completed because no credentials are configured for this account.", want: ErrGoalFailed},
		{name: "missing saved credentials", text: "No saved credentials are available, so I cannot complete this task.", want: ErrGoalFailed},
		{name: "unauthenticated github", text: "The browser session is not authenticated to GitHub.", want: ErrGoalFailed},
		{name: "captcha", text: "The page presented a CAPTCHA, so the requested navigation could not continue.", want: ErrSiteBlocked},
		{name: "access denied", text: "The site returned access denied before the target page loaded.", want: ErrSiteBlocked},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := json.Marshal(test.text)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if got := classifyTinyFishCompletedOutcome(raw); !errors.Is(got, test.want) {
				t.Fatalf("error = %v, want %v", got, test.want)
			}
		})
	}
}

func TestClassifyTinyFishCompletedOutcomeDoesNotRejectSuccessfulDiagnostics(t *testing.T) {
	tests := []json.RawMessage{
		json.RawMessage(`{"success":true,"status":"completed"}`),
		json.RawMessage(`{"goal_succeeded":true,"result":"No CAPTCHA or access-denied page was present. The task completed successfully."}`),
		json.RawMessage(`"No CAPTCHA or access-denied page was present. The task completed successfully."`),
	}
	for index, raw := range tests {
		if err := classifyTinyFishCompletedOutcome(raw); err != nil {
			t.Fatalf("case %d: unexpected failure: %v", index, err)
		}
	}
}

func TestTinyFishAgentRejectsCompletedRunThatDescribesGoalFailure(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method == http.MethodPost {
			return jsonResponse(http.StatusOK, `{"run_id":"run_completed_failed","status":"PENDING"}`), nil
		}
		return jsonResponse(http.StatusOK, `{"run_id":"run_completed_failed","status":"COMPLETED","result":"The task cannot be completed because no credentials are configured for this account."}`), nil
	})}
	agent, err := newTinyFishAgentWithHTTPClient(TinyFishAgentConfig{APIKey: "key", PollInterval: 10 * time.Millisecond}, client)
	if err != nil {
		t.Fatalf("construct agent: %v", err)
	}
	_, err = agent.Run(context.Background(), Request{URL: "https://example.com", Goal: "complete authenticated task", Purpose: "test"})
	if !errors.Is(err, ErrGoalFailed) {
		t.Fatalf("error = %v, want %v", err, ErrGoalFailed)
	}
}
