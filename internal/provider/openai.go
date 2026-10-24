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

const openaiAPIURL = "https://api.openai.com/v1/chat/completions"

type OpenAI struct {
	client *http.Client
}

func NewOpenAI() *OpenAI {
	return &OpenAI{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (o *OpenAI) Name() string  { return "openai" }
func (o *OpenAI) Managed() bool { return false }

func (o *OpenAI) GenerateDiagram(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	model, maxTokens, temperature := applyOpenAIDefaults(req.Config)

	body := openaiRequest{
		Model: model,
		Messages: []openaiMessage{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.Prompt},
		},
		MaxTokens:      maxTokens,
		Temperature:    temperature,
		Stream:         false,
		ResponseFormat: &openaiResponseFormat{Type: "json_object"},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openaiAPIURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)

	resp, err := o.client.Do(httpReq)
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

	var openaiResp openaiResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return nil, fmt.Errorf("%w: malformed response: %v", ErrInvalidResponse, err)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices returned", ErrInvalidResponse)
	}

	content := openaiResp.Choices[0].Message.Content

	var diagram GenerateResponse
	if err := json.Unmarshal([]byte(content), &diagram); err != nil {
		return nil, fmt.Errorf("%w: could not parse diagram JSON: %v", ErrInvalidResponse, err)
	}

	return &diagram, nil
}

func (o *OpenAI) GenerateText(ctx context.Context, req GenerateRequest) (string, error) {
	model, maxTokens, temperature := applyOpenAIDefaults(req.Config)

	body := openaiRequest{
		Model: model,
		Messages: []openaiMessage{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.Prompt},
		},
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Stream:      false,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openaiAPIURL, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)

	resp, err := o.client.Do(httpReq)
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

	var openaiResp openaiResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return "", fmt.Errorf("%w: malformed response: %v", ErrInvalidResponse, err)
	}

	if len(openaiResp.Choices) == 0 {
		return "", fmt.Errorf("%w: no choices returned", ErrInvalidResponse)
	}

	return openaiResp.Choices[0].Message.Content, nil
}

func (o *OpenAI) GenerateStructured(ctx context.Context, req GenerateRequest, schema string) (string, error) {
	model, maxTokens, temperature := applyOpenAIDefaults(req.Config)
	body := openaiRequest{
		Model: model,
		Messages: []openaiMessage{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.Prompt + "\n\nJSON schema hint:\n" + schema},
		},
		MaxTokens:      maxTokens,
		Temperature:    temperature,
		Stream:         false,
		ResponseFormat: &openaiResponseFormat{Type: "json_object"},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openaiAPIURL, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)

	resp, err := o.client.Do(httpReq)
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

	var openaiResp openaiResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return "", fmt.Errorf("%w: malformed response: %v", ErrInvalidResponse, err)
	}
	if len(openaiResp.Choices) == 0 {
		return "", fmt.Errorf("%w: no choices returned", ErrInvalidResponse)
	}
	return openaiResp.Choices[0].Message.Content, nil
}

func (o *OpenAI) StreamDiagram(ctx context.Context, req GenerateRequest, onChunk func(StreamChunk)) error {
	model, maxTokens, temperature := applyOpenAIDefaults(req.Config)

	body := openaiRequest{
		Model: model,
		Messages: []openaiMessage{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.Prompt},
		},
		MaxTokens:      maxTokens,
		Temperature:    temperature,
		Stream:         true,
		ResponseFormat: &openaiResponseFormat{Type: "json_object"},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openaiAPIURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProviderFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("%w: status %d: %s", ErrProviderFailed, resp.StatusCode, truncate(string(respBody), 200))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		if data == "[DONE]" {
			onChunk(StreamChunk{Done: true})
			return nil
		}

		var chunk openaiStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				onChunk(StreamChunk{Delta: delta})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%w: stream read error: %v", ErrProviderFailed, err)
	}

	onChunk(StreamChunk{Done: true})
	return nil
}

func applyOpenAIDefaults(cfg ProviderConfig) (model string, maxTokens int, temperature float64) {
	model = cfg.Model
	if model == "" {
		model = "gpt-4o"
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

// OpenAI API types

type openaiRequest struct {
	Model          string                `json:"model"`
	Messages       []openaiMessage       `json:"messages"`
	MaxTokens      int                   `json:"max_tokens"`
	Temperature    float64               `json:"temperature"`
	Stream         bool                  `json:"stream"`
	ResponseFormat *openaiResponseFormat `json:"response_format,omitempty"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponseFormat struct {
	Type string `json:"type"`
}

type openaiResponse struct {
	Choices []openaiChoice `json:"choices"`
}

type openaiChoice struct {
	Message openaiMessage `json:"message"`
}

type openaiStreamChunk struct {
	Choices []openaiStreamChoice `json:"choices"`
}

type openaiStreamChoice struct {
	Delta openaiMessage `json:"delta"`
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
