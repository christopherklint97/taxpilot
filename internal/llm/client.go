package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// ProviderPreferences controls OpenRouter provider routing options.
type ProviderPreferences struct {
	ZDR bool `json:"zdr,omitempty"` // Zero Data Retention — route only to endpoints that don't store data
}

// ChatRequest is the request payload for OpenRouter.
type ChatRequest struct {
	Model       string               `json:"model"`
	Messages    []Message            `json:"messages"`
	MaxTokens   int                  `json:"max_tokens,omitempty"`
	Temperature float64              `json:"temperature,omitempty"`
	Provider    *ProviderPreferences `json:"provider,omitempty"`
}

// Choice represents a single completion choice.
type Choice struct {
	Message Message `json:"message"`
}

// Usage tracks token usage.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResponse is the response from OpenRouter.
type ChatResponse struct {
	Choices []Choice  `json:"choices"`
	Usage   Usage     `json:"usage"`
	Error   *APIError `json:"error,omitempty"`
}

// APIError represents an API error response.
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    int    `json:"code"`
}

// Client communicates with the OpenRouter API.
type Client struct {
	apiKey     string
	baseURL    string
	model      string
	zdr        bool // zero data retention
	httpClient *http.Client
}

// DefaultModel is the default model to use via OpenRouter.
const DefaultModel = "anthropic/claude-sonnet-4.6"

// NewClient creates a new OpenRouter client.
// Reads OPENROUTER_API_KEY from environment if apiKey is empty.
// Reads TAXPILOT_MODEL from environment to override the default model.
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY not set — set it in your environment or pass it directly")
	}
	model := DefaultModel
	if envModel := os.Getenv("TAXPILOT_MODEL"); envModel != "" {
		model = envModel
	}
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1/chat/completions",
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// SetZDR enables or disables Zero Data Retention routing.
func (c *Client) SetZDR(enabled bool) {
	c.zdr = enabled
}

// SetModel changes the model used for completions.
func (c *Client) SetModel(model string) {
	c.model = model
}

// Chat sends a chat completion request and returns the assistant's response text.
func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	return c.ChatWithOptions(ctx, messages, 1024, 0.3)
}

// ChatWithOptions sends a chat completion with custom max_tokens and temperature.
func (c *Client) ChatWithOptions(ctx context.Context, messages []Message, maxTokens int, temperature float64) (string, error) {
	req := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}
	if c.zdr {
		req.Provider = &ProviderPreferences{ZDR: true}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/taxpilot")
	httpReq.Header.Set("X-Title", "TaxPilot")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// StreamChunk represents one piece of a streaming response.
type StreamChunk struct {
	Text string // delta text content
	Err  error  // non-nil on error or end of stream
	Done bool   // true when stream is complete
}

// streamDelta is the SSE delta object inside a streaming chunk.
type streamDelta struct {
	Content string `json:"content"`
}

// streamChoice is a single choice in a streaming response.
type streamChoice struct {
	Delta        streamDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

// streamEvent is one SSE data payload from OpenRouter.
type streamEvent struct {
	Choices []streamChoice `json:"choices"`
	Error   *APIError      `json:"error,omitempty"`
}

// ChatStream sends a streaming chat completion request and returns a channel
// of text chunks. The channel is closed when the stream ends. The caller
// should read from the channel until it's closed.
func (c *Client) ChatStream(ctx context.Context, messages []Message) (<-chan StreamChunk, error) {
	return c.ChatStreamWithOptions(ctx, messages, 1024, 0.3)
}

// ChatStreamWithOptions sends a streaming chat completion with custom settings.
func (c *Client) ChatStreamWithOptions(ctx context.Context, messages []Message, maxTokens int, temperature float64) (<-chan StreamChunk, error) {
	type streamRequest struct {
		ChatRequest
		Stream bool `json:"stream"`
	}

	req := streamRequest{
		ChatRequest: ChatRequest{
			Model:       c.model,
			Messages:    messages,
			MaxTokens:   maxTokens,
			Temperature: temperature,
		},
		Stream: true,
	}
	if c.zdr {
		req.Provider = &ProviderPreferences{ZDR: true}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/taxpilot")
	httpReq.Header.Set("X-Title", "TaxPilot")

	// Use a client without timeout — streaming keeps the connection open.
	// The context controls cancellation.
	streamClient := &http.Client{}
	resp, err := streamClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	ch := make(chan StreamChunk, 8)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				ch <- StreamChunk{Done: true}
				return
			}
			var event streamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue // skip malformed chunks
			}
			if event.Error != nil {
				ch <- StreamChunk{Err: fmt.Errorf("API error: %s", event.Error.Message)}
				return
			}
			if len(event.Choices) > 0 {
				delta := event.Choices[0].Delta.Content
				if delta != "" {
					ch <- StreamChunk{Text: delta}
				}
				if event.Choices[0].FinishReason != nil {
					ch <- StreamChunk{Done: true}
					return
				}
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- StreamChunk{Err: fmt.Errorf("read stream: %w", err)}
		}
	}()

	return ch, nil
}
