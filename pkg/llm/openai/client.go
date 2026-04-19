package openai

import (
	"context"
	"fmt"

	goOpenAI "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/llm"
)

const model = "gpt-5-nano-2025-08-07"

type Client struct {
	inner goOpenAI.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		inner: goOpenAI.NewClient(option.WithAPIKey(apiKey)),
	}
}

func (c *Client) buildInput(system string, messages []llm.Message) responses.ResponseInputParam {
	input := make(responses.ResponseInputParam, 0, len(messages)+1)
	if system != "" {
		input = append(input, responses.ResponseInputItemParamOfMessage(system, responses.EasyInputMessageRoleSystem))
	}
	for _, m := range messages {
		role := responses.EasyInputMessageRoleUser
		if m.Role == "assistant" {
			role = responses.EasyInputMessageRoleAssistant
		}
		input = append(input, responses.ResponseInputItemParamOfMessage(m.Content, role))
	}
	return input
}

func (c *Client) Complete(ctx context.Context, system string, messages []llm.Message) (string, error) {
	input := c.buildInput(system, messages)

	resp, err := c.inner.Responses.New(ctx, responses.ResponseNewParams{ //nolint:exhaustruct
		Model: model,
		Input: responses.ResponseNewParamsInputUnion{OfInputItemList: input},
	})
	if err != nil {
		return "", fmt.Errorf("openai.Complete: %w", err)
	}

	text := resp.OutputText()
	if text == "" {
		return "", fmt.Errorf("openai.Complete: empty response")
	}
	return text, nil
}

func (c *Client) CompleteStream(ctx context.Context, system string, messages []llm.Message, cb llm.StreamCallback) error {
	input := c.buildInput(system, messages)

	stream := c.inner.Responses.NewStreaming(ctx, responses.ResponseNewParams{ //nolint:exhaustruct
		Model: model,
		Input: responses.ResponseNewParamsInputUnion{OfInputItemList: input},
	})
	defer stream.Close()

	for stream.Next() {
		event := stream.Current()
		if event.Type == "response.output_text.delta" && event.Delta.OfString != "" {
			if err := cb(event.Delta.OfString); err != nil {
				return err
			}
		}
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("openai.CompleteStream: %w", err)
	}
	return nil
}
