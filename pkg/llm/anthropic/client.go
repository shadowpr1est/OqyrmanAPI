package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	return &Client{apiKey: apiKey, httpClient: &http.Client{Timeout: 60 * time.Second}}
}

func (c *Client) newRequest(ctx context.Context, system string, messages []llm.Message, stream bool) (*http.Request, error) {
	msgs := make([]map[string]string, len(messages))
	for i, m := range messages {
		msgs[i] = map[string]string{"role": m.Role, "content": m.Content}
	}

	payload := map[string]any{
		"model":      model,
		"max_tokens": 1024,
		"system":     system,
		"messages":   msgs,
	}
	if stream {
		payload["stream"] = true
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("anthropic marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("anthropic create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	return req, nil
}

func (c *Client) Complete(ctx context.Context, system string, messages []llm.Message) (string, error) {
	req, err := c.newRequest(ctx, system, messages, false)
	if err != nil {
		return "", err
	}

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

func (c *Client) CompleteStream(ctx context.Context, system string, messages []llm.Message, cb llm.StreamCallback) error {
	req, err := c.newRequest(ctx, system, messages, true)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("anthropic.CompleteStream do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error.Message != "" {
			return fmt.Errorf("anthropic error: %s", errResp.Error.Message)
		}
		return fmt.Errorf("anthropic.CompleteStream: status %d", resp.StatusCode)
	}

	// Anthropic SSE: event lines + data lines
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var event struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" && event.Delta.Text != "" {
			if err := cb(event.Delta.Text); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}
