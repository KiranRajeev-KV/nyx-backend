package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var globalService *EmbeddingService

type EmbeddingService struct {
	apiKey string
	model  string
	client *http.Client
}

type HFResponse struct {
	Data []HFEmbedding `json:"data"`
}

type HFEmbedding struct {
	Embedding []float64 `json:"embedding"`
}

func NewEmbeddingService(apiKey string) *EmbeddingService {
	return &EmbeddingService{
		apiKey: apiKey,
		model:  "openai/clip-vit-base-patch32",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func SetGlobalService(svc *EmbeddingService) {
	globalService = svc
}

func GetGlobalService() *EmbeddingService {
	return globalService
}

func (s *EmbeddingService) GetImageEmbedding(imageURL string) ([]float64, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("HuggingFace API key not configured")
	}

	url := fmt.Sprintf("https://router.huggingface.co/models/%s", s.model)

	payload := map[string]interface{}{
		"inputs": map[string]string{
			"image": imageURL,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call HuggingFace API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HuggingFace API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result [][]float64
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	if len(result) == 0 || len(result[0]) == 0 {
		return nil, fmt.Errorf("no embedding returned from API")
	}

	return result[0], nil
}
