package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shadowpr1est/OqyrmanAPI/pkg/llm"
)

const (
	apiURL = "https://api.anthropic.com/v1/messages"
	// FIX: было claude-sonnet-4-20250514 — устаревший идентификатор модели.
	// Актуальная модель на март 2026: claude-sonnet-4-6
	model = "claude-sonnet-4-6"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, httpClient: &http.Client{}}
}

func (c *Client) Complete(ctx context.Context, system string, messages []llm.Message) (string, error) {
	msgs := make([]map[string]string, len(messages))
	for i, m := range messages {
		msgs[i] = map[string]string{"role": m.Role, "content": m.Content}
	}

	body, _ := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": 1024,
		"system":     system,
		"messages":   msgs,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("anthropic.Complete create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic.Complete do request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("anthropic.Complete decode response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("anthropic error: %s", result.Error.Message)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("anthropic.Complete: empty response")
	}
	return result.Content[0].Text, nil
}
