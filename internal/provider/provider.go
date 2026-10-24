package provider

import (
	"context"
	"errors"
)

var (
	ErrProviderNotFound = errors.New("unknown provider")
	ErrProviderFailed   = errors.New("provider request failed")
	ErrInvalidResponse  = errors.New("provider returned invalid response")
	// ErrTruncated means the model stopped before finishing — typically because
	// the output ran past the configured max_tokens budget. Surfaced to the
	// user as a friendlier "diagram too large" error rather than a JSON parse
	// failure.
	ErrTruncated = errors.New("provider response truncated")
)

// Provider defines the contract for LLM-based generation.
type Provider interface {
	// GenerateDiagram performs a non-streaming request and returns the full response.
	GenerateDiagram(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)

	// StreamDiagram sends partial text chunks to the callback as they arrive from the LLM.
	// The callback receives raw text deltas — the caller is responsible for accumulating
	// and parsing the final JSON. An empty delta signals the stream is complete.
	StreamDiagram(ctx context.Context, req GenerateRequest, onChunk func(StreamChunk)) error

	// GenerateText performs a non-streaming, free-form text generation (no JSON parsing).
	// Used for IaC code generation and explanations where the output is not diagram JSON.
	GenerateText(ctx context.Context, req GenerateRequest) (string, error)

	// GenerateStructured returns a JSON string following the caller-provided schema.
	GenerateStructured(ctx context.Context, req GenerateRequest, schema string) (string, error)

	// Managed returns true if this provider uses server-side credentials
	// (no user-supplied API key required).
	Managed() bool

	Name() string
}

// GenerateRequest contains everything the provider needs to generate a diagram.
type GenerateRequest struct {
	Prompt       string         `json:"prompt"`
	SystemPrompt string         `json:"-"` // Mode-aware system instructions; built by service layer
	APIKey       string         `json:"-"` // Never serialized; empty for managed providers
	Config       ProviderConfig `json:"config,omitempty"`
}

// ProviderConfig holds optional tuning parameters for the LLM call.
type ProviderConfig struct {
	Model       string  `json:"model,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// StreamChunk represents a single chunk in a streaming response.
type StreamChunk struct {
	Delta      string `json:"delta"`                 // Partial text content
	Done       bool   `json:"done"`                  // True when stream is complete
	Error      string `json:"error,omitempty"`       // Non-empty if the stream errored
	StopReason string `json:"stop_reason,omitempty"` // Provider stop reason (e.g. "max_tokens"); only set on the final Done chunk
}

// GenerateResponse is the raw structured output from the provider.
type GenerateResponse struct {
	Nodes  []Node  `json:"nodes"`
	Edges  []Edge  `json:"edges"`
	Groups []Group `json:"groups,omitempty"`
}

type Node struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Group string `json:"group,omitempty"`
	// Metadata is passthrough — we don't read it on the backend. Typed as
	// map[string]any (was map[string]string) because models occasionally
	// emit non-string values ("metadata": {"managed": true, "port": 443})
	// which used to blow up the whole generate stream with:
	//   cannot unmarshal bool into Go struct field Node.nodes.metadata
	// The frontend either ignores it or renders as-is; anything the model
	// wants to carry through is fine.
	Metadata map[string]any `json:"metadata,omitempty"`
}

type Edge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label,omitempty"`
	Style  string `json:"style,omitempty"` // "solid", "dashed", "dotted"
}

type Group struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type,omitempty"` // e.g., "vpc", "subnet", "region"
}
