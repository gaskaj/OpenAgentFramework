package claude

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client wraps the Anthropic SDK for Claude API interactions.
type Client struct {
	sdk       anthropic.Client
	model     string
	maxTokens int
}

// NewClient creates a new Claude API client.
func NewClient(apiKey, model string, maxTokens int) *Client {
	sdk := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Client{
		sdk:       sdk,
		model:     model,
		maxTokens: maxTokens,
	}
}

// SendMessage sends a single message to Claude and returns the text response.
func (c *Client) SendMessage(ctx context.Context, system string, messages []anthropic.MessageParam) (*anthropic.Message, error) {
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: int64(c.maxTokens),
		Messages:  messages,
	}
	if system != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: system},
		}
	}

	msg, err := c.sdk.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("claude API call: %w", err)
	}
	return msg, nil
}

// SendMessageWithTools sends a message with tool definitions and returns the response.
func (c *Client) SendMessageWithTools(ctx context.Context, system string, messages []anthropic.MessageParam, tools []anthropic.ToolUnionParam) (*anthropic.Message, error) {
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: int64(c.maxTokens),
		Messages:  messages,
		Tools:     tools,
	}
	if system != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: system},
		}
	}

	msg, err := c.sdk.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("claude API call with tools: %w", err)
	}
	return msg, nil
}

// ExtractText extracts the text content from a Claude message response.
func ExtractText(msg *anthropic.Message) string {
	for _, block := range msg.Content {
		if block.Type == "text" {
			return block.Text
		}
	}
	return ""
}
