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

type Describer struct {
	apiKey string
	client *http.Client
	model  string
}

type OpenRouterRequest struct {
	Model    string              `json:"model"`
	Messages []OpenRouterMessage `json:"messages"`
}

type OpenRouterMessage struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

type OpenRouterResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Message struct {
	Content string `json:"content"`
}

func NewDescriber(apiKey string) *Describer {
	return &Describer{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		model: "qwen/qwen-2.5-vl-7b-instruct",
	}
}

func (d *Describer) GenerateAIDesc(ctx context.Context, imageURL, name, description, location string) (string, error) {
	prompt := fmt.Sprintf(`Analyze this lost/found item image and provide a detailed description that would help identify the item's owner or match it with lost items.

Item Details:
- Name: %s
- Description: %s  
- Location: %s

Please provide a comprehensive description focusing on:
1. Visual details visible in the image (brand, color, style, condition, distinguishing marks)
2. Any text or serial numbers visible
3. The type and category of item
4. Any other details that could help identify or match this item

Be specific and detailed. If the image is unclear or the item is not visible, note that as well.`, name, description, location)

	reqBody := OpenRouterRequest{
		Model: d.model,
		Messages: []OpenRouterMessage{
			{
				Role: "user",
				Content: []Content{
					{
						Type: "text",
						Text: prompt,
					},
					{
						Type:     "image_url",
						ImageURL: imageURL,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiKey))
	req.Header.Set("HTTP-Referer", "https://nyx-backend.local")
	req.Header.Set("X-Title", "Nyx Backend")

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var openResp OpenRouterResponse
	if err := json.Unmarshal(body, &openResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	if len(openResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from API")
	}

	return openResp.Choices[0].Message.Content, nil
}
