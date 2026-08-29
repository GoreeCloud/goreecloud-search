package syncstate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var ErrSyncRetrievalFailed = errors.New("Search history sync retrieval failed")

const (
	maxRetrievalBodyBytes = 1 << 20
	retrievalPageSize     = 256
	maxRetrievalPages     = 1024
	maxRetrievalRecords   = 65536
)

type RetrievalClient struct {
	BaseURL     string
	BearerToken string
	Client      HTTPDoer
}

type retrievalResponse struct {
	Dataset   string     `json:"dataset"`
	Count     int        `json:"count"`
	Records   []Envelope `json:"records"`
	NextAfter string     `json:"nextAfter,omitempty"`
}

func (c RetrievalClient) FetchHistory(ctx context.Context) ([]Envelope, error) {
	capability, ok := searchHistoryCapability()
	token := strings.TrimSpace(c.BearerToken)
	if !ok || !capability.Read || strings.TrimSpace(c.BaseURL) == "" || token == "" || c.Client == nil {
		return nil, ErrSyncRetrievalFailed
	}

	all := make([]Envelope, 0)
	after := ""
	for page := 0; page < maxRetrievalPages; page++ {
		result, err := c.fetchHistoryPage(ctx, token, after)
		if err != nil {
			return nil, err
		}
		if len(all)+len(result.Records) > maxRetrievalRecords {
			return nil, ErrSyncRetrievalFailed
		}
		all = append(all, result.Records...)
		if result.NextAfter == "" {
			return all, nil
		}
		if len(result.Records) == 0 || result.NextAfter != result.Records[len(result.Records)-1].RecordID || (after != "" && result.NextAfter <= after) {
			return nil, ErrSyncRetrievalFailed
		}
		after = result.NextAfter
	}
	return nil, ErrSyncRetrievalFailed
}

func (c RetrievalClient) fetchHistoryPage(ctx context.Context, token, after string) (retrievalResponse, error) {
	if len(after) > maxSyncRecordIDBytes {
		return retrievalResponse{}, ErrSyncRetrievalFailed
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.BaseURL, "/")+"/api/v1/sync/search/history", nil)
	if err != nil {
		return retrievalResponse{}, err
	}
	query := request.URL.Query()
	query.Set("limit", strconv.Itoa(retrievalPageSize))
	if after != "" {
		query.Set("after", after)
	}
	request.URL.RawQuery = query.Encode()
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)

	response, err := c.Client.Do(request)
	if err != nil {
		return retrievalResponse{}, fmt.Errorf("%w: %v", ErrSyncRetrievalFailed, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
		return retrievalResponse{}, fmt.Errorf("%w: status %d", ErrSyncRetrievalFailed, response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxRetrievalBodyBytes+1))
	if err != nil || len(body) > maxRetrievalBodyBytes {
		return retrievalResponse{}, fmt.Errorf("%w: invalid response", ErrSyncRetrievalFailed)
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var result retrievalResponse
	if err := decoder.Decode(&result); err != nil {
		return retrievalResponse{}, fmt.Errorf("%w: invalid response", ErrSyncRetrievalFailed)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return retrievalResponse{}, fmt.Errorf("%w: invalid response", ErrSyncRetrievalFailed)
	}
	if result.Dataset != searchHistoryDataset || result.Count != len(result.Records) || len(result.Records) > retrievalPageSize || len(result.NextAfter) > maxSyncRecordIDBytes {
		return retrievalResponse{}, ErrSyncRetrievalFailed
	}
	previousID := after
	for _, record := range result.Records {
		if !validHistoryEnvelope(record) || (previousID != "" && record.RecordID <= previousID) {
			return retrievalResponse{}, ErrSyncRetrievalFailed
		}
		previousID = record.RecordID
	}
	return result, nil
}
