package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Embedder struct {
	apiKey string
	client *http.Client
	model  string
}

type GeminiEmbedRequest struct {
	Content struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"content"`
}

type GeminiEmbedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

func NewEmbedder(apiKey string) *Embedder {
	return &Embedder{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		model: "gemini-embedding-001",
	}
}

func (e *Embedder) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s", e.model, e.apiKey)

	var reqBody GeminiEmbedRequest
	reqBody.Content.Parts = append(reqBody.Content.Parts, struct {
		Text string `json:"text"`
	}{Text: text})

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiEmbedResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	if len(geminiResp.Embedding.Values) == 0 {
		return nil, fmt.Errorf("no embedding values returned from API")
	}

	return geminiResp.Embedding.Values, nil
}
