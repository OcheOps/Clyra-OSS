package diagram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/oche/diagram/internal/abuse"
	"github.com/oche/diagram/internal/provider"
)

var (
	ErrEmptyPrompt     = errors.New("prompt is required")
	ErrMissingAPIKey   = errors.New("api_key is required for BYO providers")
	ErrMissingProvider = errors.New("provider is required")
)

// ActivationRecorder is implemented by *activation.Recorder. Kept as an
// interface to avoid a hard dependency.
type ActivationRecorder interface {
	Record(ctx context.Context, userID string, kind string)
}

type Service struct {
	registry   *provider.Registry
	guard      *abuse.Guard
	activation ActivationRecorder
	logger     *slog.Logger
}

// NewService creates the diagram generation service.
// guard can be nil — abuse checks will be skipped.
func NewService(registry *provider.Registry, guard *abuse.Guard, logger *slog.Logger) *Service {
	return &Service{
		registry: registry,
		guard:    guard,
		logger:   logger,
	}
}

// WithActivation wires the recorder so the first successful generate per
// user emits `first_diagram`. No-op if omitted.
func (s *Service) WithActivation(r ActivationRecorder) *Service {
	s.activation = r
	return s
}

// Generate performs a non-streaming diagram generation.
// plan is the user's billing/usage plan ("free", "pro", "byo") used for quota enforcement.
func (s *Service) Generate(ctx context.Context, userID, plan string, req GenerateRequest) (*GenerateResponse, error) {
	prompt, err := s.preCheck(ctx, userID, plan, req)
	if err != nil {
		return nil, err
	}
	req.Prompt = prompt

	p, err := s.registry.Get(req.Provider)
	if err != nil {
		return nil, err
	}

	s.logger.Info("generating diagram",
		"user_id", userID,
		"provider", req.Provider,
		"mode", req.Mode,
		"prompt_length", len(req.Prompt),
	)

	// Clamp tokens (paid plans get a higher ceiling)
	maxTokens := req.Config.MaxTokens
	if s.guard != nil {
		maxTokens = s.guard.TokenGuard.ClampTokens(maxTokens, plan)
	}

	sysPrompt := provider.BuildSystemPrompt(req.Mode)

	// Apply provider timeout (plan-aware — paid users get longer headroom).
	callCtx := ctx
	if s.guard != nil {
		if t := s.guard.TimeoutForPlan(plan); t > 0 {
			var cancel context.CancelFunc
			callCtx, cancel = context.WithTimeout(ctx, t)
			defer cancel()
		}
	}

	start := time.Now()
	raw, err := p.GenerateDiagram(callCtx, provider.GenerateRequest{
		Prompt:       req.Prompt,
		SystemPrompt: sysPrompt,
		APIKey:       req.APIKey,
		Config:       provider.ProviderConfig{Model: req.Config.Model, MaxTokens: maxTokens, Temperature: req.Config.Temperature},
	})
	latency := time.Since(start)

	if err != nil {
		s.logger.Error("provider failed", "provider", req.Provider, "error", err, "latency", latency)
		if s.guard != nil {
			s.guard.PostFailure(ctx, userID, req.Provider, latency)
		}
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	Normalize(raw)
	if report := Salvage(raw); report.DroppedEdges > 0 || report.OrphanGroupRef > 0 {
		s.logger.Warn("salvaged diagram",
			"provider", req.Provider,
			"dropped_edges", report.DroppedEdges,
			"orphan_group_refs", report.OrphanGroupRef,
		)
	}

	if ve := Validate(raw); ve != nil {
		s.logger.Warn("provider returned invalid diagram", "provider", req.Provider, "errors", ve.Errors)
		return nil, ve
	}

	if s.guard != nil {
		s.guard.PostSuccess(ctx, userID, req.Provider, len(raw.Nodes), latency)
	}
	if s.activation != nil {
		s.activation.Record(ctx, userID, "first_diagram")
	}

	return &GenerateResponse{Nodes: raw.Nodes, Edges: raw.Edges, Groups: raw.Groups}, nil
}

// Stream sends raw LLM text deltas via the onChunk callback as they arrive.
// plan is the user's billing/usage plan used for quota enforcement.
func (s *Service) Stream(
	ctx context.Context,
	userID, plan string,
	req GenerateRequest,
	onChunk func(provider.StreamChunk),
	onResult func(*GenerateResponse, error),
) {
	prompt, err := s.preCheck(ctx, userID, plan, req)
	if err != nil {
		onResult(nil, err)
		return
	}
	req.Prompt = prompt

	p, err := s.registry.Get(req.Provider)
	if err != nil {
		onResult(nil, err)
		return
	}

	s.logger.Info("streaming diagram",
		"user_id", userID,
		"provider", req.Provider,
		"mode", req.Mode,
		"prompt_length", len(req.Prompt),
	)

	maxTokens := req.Config.MaxTokens
	if s.guard != nil {
		maxTokens = s.guard.TokenGuard.ClampTokens(maxTokens, plan)
	}

	var accumulated []byte
	var stopReason string
	wrappedOnChunk := func(chunk provider.StreamChunk) {
		if chunk.Delta != "" {
			accumulated = append(accumulated, chunk.Delta...)
		}
		if chunk.StopReason != "" {
			stopReason = chunk.StopReason
		}
		onChunk(chunk)
	}

	sysPrompt := provider.BuildSystemPrompt(req.Mode)

	callCtx := ctx
	if s.guard != nil {
		if t := s.guard.TimeoutForPlan(plan); t > 0 {
			var cancel context.CancelFunc
			callCtx, cancel = context.WithTimeout(ctx, t)
			defer cancel()
		}
	}

	start := time.Now()
	err = p.StreamDiagram(callCtx, provider.GenerateRequest{
		Prompt:       req.Prompt,
		SystemPrompt: sysPrompt,
		APIKey:       req.APIKey,
		Config:       provider.ProviderConfig{Model: req.Config.Model, MaxTokens: maxTokens, Temperature: req.Config.Temperature},
	}, wrappedOnChunk)
	latency := time.Since(start)

	if err != nil {
		s.logger.Error("provider stream failed", "provider", req.Provider, "error", err, "latency", latency)
		if s.guard != nil {
			s.guard.PostFailure(ctx, userID, req.Provider, latency)
		}
		onResult(nil, fmt.Errorf("stream failed: %w", err))
		return
	}

	content := extractJSON(string(accumulated))
	var raw provider.GenerateResponse
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		// Truncation (model hit max_tokens / context window) leaves a half-formed
		// JSON object that can't be parsed. Surface a clearer error so the user
		// sees "diagram too large" instead of a parser stack trace.
		if isTruncationStop(stopReason) {
			s.logger.Warn("diagram stream truncated",
				"provider", req.Provider,
				"stop_reason", stopReason,
				"accumulated_bytes", len(accumulated),
			)
			if s.guard != nil {
				s.guard.PostFailure(ctx, userID, req.Provider, latency)
			}
			onResult(nil, fmt.Errorf("%w: diagram exceeded the model output budget — try a more focused prompt", provider.ErrTruncated))
			return
		}
		s.logger.Error("failed to parse streamed JSON",
			"provider", req.Provider,
			"stop_reason", stopReason,
			"error", err,
			"content_length", len(content),
			"content_preview", truncateStr(content, 500),
		)
		if s.guard != nil {
			s.guard.PostFailure(ctx, userID, req.Provider, latency)
		}
		onResult(nil, fmt.Errorf("%w: could not parse streamed diagram JSON: %v", provider.ErrInvalidResponse, err))
		return
	}

	s.logger.Info("parsed streamed diagram",
		"provider", req.Provider,
		"accumulated_bytes", len(accumulated),
		"extracted_bytes", len(content),
		"nodes", len(raw.Nodes),
		"edges", len(raw.Edges),
		"groups", len(raw.Groups),
	)

	Normalize(&raw)
	if report := Salvage(&raw); report.DroppedEdges > 0 || report.OrphanGroupRef > 0 {
		s.logger.Warn("salvaged streamed diagram",
			"provider", req.Provider,
			"dropped_edges", report.DroppedEdges,
			"orphan_group_refs", report.OrphanGroupRef,
		)
	}

	if ve := Validate(&raw); ve != nil {
		s.logger.Warn("streamed diagram failed validation", "provider", req.Provider, "errors", ve.Errors)
		onResult(nil, ve)
		return
	}

	if s.guard != nil {
		s.guard.PostSuccess(ctx, userID, req.Provider, len(raw.Nodes), latency)
	}
	if s.activation != nil {
		s.activation.Record(ctx, userID, "first_diagram")
	}

	onResult(&GenerateResponse{Nodes: raw.Nodes, Edges: raw.Edges, Groups: raw.Groups}, nil)
}

// preCheck runs basic validation + abuse guard pre-checks.
func (s *Service) preCheck(ctx context.Context, userID, plan string, req GenerateRequest) (string, error) {
	if req.Prompt == "" {
		return "", ErrEmptyPrompt
	}
	if req.Provider == "" {
		return "", ErrMissingProvider
	}
	if !s.registry.IsManaged(req.Provider) && req.APIKey == "" {
		return "", ErrMissingAPIKey
	}

	// Run abuse guard pre-checks (prompt validation, circuit breaker, quota)
	if s.guard != nil {
		cleaned, err := s.guard.PreCheck(ctx, userID, plan, req.Prompt)
		if err != nil {
			return "", err
		}
		return cleaned, nil
	}

	return req.Prompt, nil
}

// isTruncationStop reports whether a provider stop reason indicates the model
// ran out of output budget rather than finishing cleanly.
func isTruncationStop(reason string) bool {
	switch reason {
	case "max_tokens", "model_context_window_exceeded", "length":
		return true
	}
	return false
}

// --- JSON extraction helpers (unchanged) ---

func extractJSON(s string) string {
	content := stripCodeFences(s)

	firstBrace := strings.Index(content, "{")
	lastBrace := strings.LastIndex(content, "}")
	if firstBrace >= 0 && lastBrace > firstBrace {
		candidate := content[firstBrace : lastBrace+1]
		if strings.Contains(candidate, `"nodes"`) {
			return candidate
		}
	}

	firstBrace = strings.Index(s, "{")
	lastBrace = strings.LastIndex(s, "}")
	if firstBrace >= 0 && lastBrace > firstBrace {
		candidate := s[firstBrace : lastBrace+1]
		if strings.Contains(candidate, `"nodes"`) {
			return candidate
		}
	}

	return strings.TrimSpace(content)
}

func stripCodeFences(s string) string {
	content := strings.TrimSpace(s)

	for {
		start := strings.Index(content, "<think>")
		if start == -1 {
			break
		}
		end := strings.Index(content, "</think>")
		if end == -1 {
			content = strings.TrimSpace(content[:start])
			break
		}
		content = strings.TrimSpace(content[:start] + content[end+len("</think>"):])
	}

	if strings.HasPrefix(content, "```") {
		nl := strings.Index(content, "\n")
		if nl >= 0 {
			content = content[nl+1:]
		}
	}

	content = strings.TrimSpace(content)
	if strings.HasSuffix(content, "```") {
		content = strings.TrimSpace(content[:len(content)-3])
	}

	return strings.TrimSpace(content)
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
