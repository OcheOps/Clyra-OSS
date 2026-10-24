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

const geminiAPIBaseURL = "https://generativelanguage.googleapis.com/v1beta/models"

type Gemini struct {
	client *http.Client
}

func NewGemini() *Gemini {
	return &Gemini{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (g *Gemini) Name() string  { return "gemini" }
func (g *Gemini) Managed() bool { return false }

func (g *Gemini) GenerateDiagram(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	model, maxTokens, temperature := applyGeminiDefaults(req.Config)

	body := geminiRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: req.SystemPrompt + "\n\n" + req.Prompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			MaxOutputTokens:  maxTokens,
			Temperature:      temperature,
			ResponseMimeType: "application/json",
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	// The key goes in x-goog-api-key, NOT the URL query, so a network
	// error (which Go stringifies with the full URL) can't leak the
	// caller's BYO Gemini key through our error wrapping into Sentry.
	url := fmt.Sprintf("%s/%s:generateContent", geminiAPIBaseURL, model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", req.APIKey)

	resp, err := g.client.Do(httpReq)
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

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("%w: malformed response: %v", ErrInvalidResponse, err)
	}

	content := geminiResp.extractText()
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

func (g *Gemini) GenerateText(ctx context.Context, req GenerateRequest) (string, error) {
	model, maxTokens, temperature := applyGeminiDefaults(req.Config)

	body := geminiRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: req.SystemPrompt + "\n\n" + req.Prompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			MaxOutputTokens: maxTokens,
			Temperature:     temperature,
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent", geminiAPIBaseURL, model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", req.APIKey)

	resp, err := g.client.Do(httpReq)
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

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", fmt.Errorf("%w: malformed response: %v", ErrInvalidResponse, err)
	}

	content := geminiResp.extractText()
	if content == "" {
		return "", fmt.Errorf("%w: no text content in response", ErrInvalidResponse)
	}
	return content, nil
}

func (g *Gemini) GenerateStructured(ctx context.Context, req GenerateRequest, schema string) (string, error) {
	model, maxTokens, temperature := applyGeminiDefaults(req.Config)

	body := geminiRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: req.SystemPrompt + "\n\n" + req.Prompt + "\n\nJSON schema hint:\n" + schema},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			MaxOutputTokens:  maxTokens,
			Temperature:      temperature,
			ResponseMimeType: "application/json",
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent", geminiAPIBaseURL, model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", req.APIKey)

	resp, err := g.client.Do(httpReq)
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

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", fmt.Errorf("%w: malformed response: %v", ErrInvalidResponse, err)
	}
	content := geminiResp.extractText()
	if content == "" {
		return "", fmt.Errorf("%w: no text content in response", ErrInvalidResponse)
	}
	return content, nil
}

func (g *Gemini) StreamDiagram(ctx context.Context, req GenerateRequest, onChunk func(StreamChunk)) error {
	model, maxTokens, temperature := applyGeminiDefaults(req.Config)

	body := geminiRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: req.SystemPrompt + "\n\n" + req.Prompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			MaxOutputTokens:  maxTokens,
			Temperature:      temperature,
			ResponseMimeType: "application/json",
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:streamGenerateContent?alt=sse", geminiAPIBaseURL, model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", req.APIKey)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProviderFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("%w: status %d: %s", ErrProviderFailed, resp.StatusCode, truncate(string(respBody), 200))
	}

	// Gemini streams SSE with data: prefixed JSON lines containing partial candidates.
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var chunk geminiStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		// Extract text from candidates
		for _, candidate := range chunk.Candidates {
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					onChunk(StreamChunk{Delta: part.Text})
				}
			}
			// Check for finish reason on the last candidate
			if candidate.FinishReason != "" && candidate.FinishReason != "STOP" {
				if candidate.FinishReason == "SAFETY" {
					onChunk(StreamChunk{Error: "content blocked by safety filters", Done: true})
					return fmt.Errorf("%w: content blocked by safety filters", ErrProviderFailed)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%w: stream read error: %v", ErrProviderFailed, err)
	}

	onChunk(StreamChunk{Done: true})
	return nil
}

func applyGeminiDefaults(cfg ProviderConfig) (model string, maxTokens int, temperature float64) {
	model = cfg.Model
	if model == "" {
		model = "gemini-2.0-flash"
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

// Gemini API types

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens  int     `json:"maxOutputTokens"`
	Temperature      float64 `json:"temperature"`
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason,omitempty"`
}

func (r *geminiResponse) extractText() string {
	for _, c := range r.Candidates {
		for _, p := range c.Content.Parts {
			if p.Text != "" {
				return p.Text
			}
		}
	}
	return ""
}

// Streaming types

type geminiStreamChunk struct {
	Candidates []geminiCandidate `json:"candidates"`
}
