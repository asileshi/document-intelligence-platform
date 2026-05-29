package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type QdrantClient struct {
	baseURL      string
	collection   string
	client       *http.Client
	embeddingDim int
}

func NewQdrantClient(baseURL string, collection string, embeddingDim int) *QdrantClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "http://qdrant:6333"
	}
	if collection == "" {
		collection = "documents"
	}
	if embeddingDim <= 0 {
		embeddingDim = 8
	}
	return &QdrantClient{
		baseURL:      baseURL,
		collection:   collection,
		embeddingDim: embeddingDim,
		client:       &http.Client{Timeout: 15 * time.Second},
	}
}

func (q *QdrantClient) EnsureCollection(ctx context.Context) error {
	// Idempotent create with vector size.
	url := fmt.Sprintf("%s/collections/%s", q.baseURL, q.collection)
	createBody := map[string]any{
		"vectors": map[string]any{
			"size":     q.embeddingDim,
			"distance": "Cosine",
		},
	}

	b, _ := json.Marshal(createBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	// Qdrant returns 409 when the collection already exists.
	if resp.StatusCode == 409 {
		return nil
	}
	return fmt.Errorf("ensure collection failed: status=%d", resp.StatusCode)
}

type QdrantPoint struct {
	ID      string         `json:"id"`
	Vector  []float32      `json:"vector"`
	Payload map[string]any `json:"payload"`
}

type upsertPointsRequest struct {
	Points []QdrantPoint `json:"points"`
}

func (q *QdrantClient) UpsertPoints(ctx context.Context, points []QdrantPoint) error {
	if len(points) == 0 {
		return nil
	}
	url := fmt.Sprintf("%s/collections/%s/points?wait=true", q.baseURL, q.collection)
	body := upsertPointsRequest{Points: points}
	b, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := q.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("qdrant upsert failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(respBody)))
}

func deterministicPointID(jobID string, chunkIndex int) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", jobID, chunkIndex)))
	return hex.EncodeToString(sum[:16])
}
