package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"

type Anthropic struct {
	client *http.Client
}

func NewAnthropic() *Anthropic {
	return &Anthropic{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *Anthropic) Name() string  { return "anthropic" }
func (a *Anthropic) Managed() bool { return false }

func (a *Anthropic) GenerateDiagram(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	model, maxTokens, temperature := applyAnthropicDefaults(req.Config)

	body := anthropicRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    req.SystemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: req.Prompt},
		},
		Temperature: temperature,
		Stream:      false,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	setAnthropicHeaders(httpReq, req.APIKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderFailed, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d: %s", ErrProviderFailed, resp.StatusCode, truncate(string(respBody), 200))
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return nil, fmt.Errorf("%w: malformed response: %v", ErrInvalidResponse, err)
	}

	content := anthropicResp.extractText()
	if content == "" {
		return nil, fmt.Errorf("%w: no text content in response", ErrInvalidResponse)
	}

	content = stripCodeFences(content)

	var diagram GenerateResponse
	if err := json.Unmarshal([]byte(content), &diagram); err != nil {
		return nil, fmt.Errorf("%w: could not parse diagram JSON: %v", ErrInvalidResponse, err)
	}

	return &diagram, nil
}

func (a *Anthropic) GenerateText(ctx context.Context, req GenerateRequest) (string, error) {
	model, maxTokens, temperature := applyAnthropicDefaults(req.Config)

	body := anthropicRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    req.SystemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: req.Prompt},
		},
		Temperature: temperature,
		Stream:      false,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	setAnthropicHeaders(httpReq, req.APIKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrProviderFailed, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: status %d: %s", ErrProviderFailed, resp.StatusCode, truncate(string(respBody), 200))
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return "", fmt.Errorf("%w: malformed response: %v", ErrInvalidResponse, err)
	}

	content := anthropicResp.extractText()
	if content == "" {
		return "", fmt.Errorf("%w: no text content in response", ErrInvalidResponse)
	}
	return content, nil
}

func (a *Anthropic) GenerateStructured(ctx context.Context, req GenerateRequest, schema string) (string, error) {
	req.Prompt = req.Prompt + "\n\nJSON schema hint:\n" + schema
	return a.GenerateText(ctx, req)
}

func (a *Anthropic) StreamDiagram(ctx context.Context, req GenerateRequest, onChunk func(StreamChunk)) error {
	model, maxTokens, temperature := applyAnthropicDefaults(req.Config)

	body := anthropicRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    req.SystemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: req.Prompt},
		},
		Temperature: temperature,
		Stream:      true,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	setAnthropicHeaders(httpReq, req.APIKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProviderFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("%w: status %d: %s", ErrProviderFailed, resp.StatusCode, truncate(string(respBody), 200))
	}

	// Anthropic streams SSE events with types: content_block_delta, message_stop, etc.
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var event anthropicStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		switch event.Type {
		case "content_block_delta":
			if event.Delta.Type == "text_delta" && event.Delta.Text != "" {
				onChunk(StreamChunk{Delta: event.Delta.Text})
			}
		case "message_stop":
			onChunk(StreamChunk{Done: true})
			return nil
		case "error":
			onChunk(StreamChunk{Error: "provider stream error", Done: true})
			return fmt.Errorf("%w: stream error from provider", ErrProviderFailed)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%w: stream read error: %v", ErrProviderFailed, err)
	}

	onChunk(StreamChunk{Done: true})
	return nil
}

func applyAnthropicDefaults(cfg ProviderConfig) (model string, maxTokens int, temperature float64) {
	model = cfg.Model
	if model == "" {
		// Sonnet 4.6 stays the default for the BYO Anthropic path — BYO
		// users pay per-token on their own key, so silently defaulting
		// them to Opus would inflate their bill without consent. If they
		// want Opus they can pass ProviderConfig.Model per request.
		// Managed (Bedrock / Pro) users get Opus by default — that's on
		// our tab and they're paying for the best.
		model = "claude-sonnet-4-6"
	}
	maxTokens = cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}
	temperature = cfg.Temperature
	if temperature == 0 {
		temperature = 0.3
	}
	return
}

func setAnthropicHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
}

// Anthropic API types

type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	System      string             `json:"system,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
	Temperature float64            `json:"temperature"`
	Stream      bool               `json:"stream"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (r *anthropicResponse) extractText() string {
	for _, c := range r.Content {
		if c.Type == "text" {
			return c.Text
		}
	}
	return ""
}

// Streaming types

type anthropicStreamEvent struct {
	Type  string               `json:"type"`
	Delta anthropicStreamDelta `json:"delta,omitempty"`
}

type anthropicStreamDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// stripCodeFences removes markdown code fences that the model might wrap around JSON.
func stripCodeFences(s string) string {
	trimmed := []byte(s)

	if len(trimmed) > 3 && string(trimmed[:3]) == "```" {
		for i := 3; i < len(trimmed); i++ {
			if trimmed[i] == '\n' {
				trimmed = trimmed[i+1:]
				break
			}
		}
	}

	if len(trimmed) > 3 && string(trimmed[len(trimmed)-3:]) == "```" {
		trimmed = trimmed[:len(trimmed)-3]
	}

	if len(trimmed) > 4 && string(trimmed[len(trimmed)-4:]) == "```\n" {
		trimmed = trimmed[:len(trimmed)-4]
	}

	return string(bytes.TrimSpace(trimmed))
}
