package syncstate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var ErrSyncRetrievalFailed = errors.New("Search history sync retrieval failed")

const maxRetrievalBodyBytes = 1 << 20

type RetrievalClient struct {
	BaseURL     string
	BearerToken string
	Client      HTTPDoer
}

type retrievalResponse struct {
	Dataset string     `json:"dataset"`
	Count   int        `json:"count"`
	Records []Envelope `json:"records"`
}

func (c RetrievalClient) FetchHistory(ctx context.Context) ([]Envelope, error) {
	if strings.TrimSpace(c.BaseURL) == "" || c.Client == nil {
		return nil, ErrSyncRetrievalFailed
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.BaseURL, "/")+"/api/v1/sync/search/history", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	if token := strings.TrimSpace(c.BearerToken); token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := c.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSyncRetrievalFailed, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
		return nil, fmt.Errorf("%w: status %d", ErrSyncRetrievalFailed, response.StatusCode)
	}

	decoder := json.NewDecoder(io.LimitReader(response.Body, maxRetrievalBodyBytes))
	decoder.DisallowUnknownFields()
	var result retrievalResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("%w: invalid response", ErrSyncRetrievalFailed)
	}
	if result.Dataset != "search.history" || result.Count != len(result.Records) {
		return nil, ErrSyncRetrievalFailed
	}
	for _, record := range result.Records {
		if record.Dataset != "search.history" || record.SchemaVersion < 1 || record.RecordID == "" || record.Revision == 0 || record.UpdatedAt.IsZero() || strings.TrimSpace(record.OriginDevice) == "" {
			return nil, ErrSyncRetrievalFailed
		}
	}
	return result.Records, nil
}
