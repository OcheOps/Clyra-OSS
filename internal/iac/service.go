package iac

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/oche/diagram/internal/abuse"
	"github.com/oche/diagram/internal/iac/graph"
	_ "github.com/oche/diagram/internal/iac/graph/props"
	"github.com/oche/diagram/internal/iac/render"
	"github.com/oche/diagram/internal/provider"
)

// ActivationRecorder is implemented by *activation.Recorder.
// HasKind powers the free-plan one-shot export gate.
type ActivationRecorder interface {
	Record(ctx context.Context, userID string, kind string)
	HasKind(ctx context.Context, userID string, kind string) bool
}

// AuditSink mirrors the contract used elsewhere — local interface to dodge
// an import cycle on internal/audit.
type AuditSink interface {
	Log(ctx context.Context, event string, fields map[string]any)
}

type Service struct {
	registry   *provider.Registry
	guard      *abuse.Guard
	activation ActivationRecorder
	audit      AuditSink
	logger     *slog.Logger
}

// NewService wires up the IaC generation service.
// guard may be nil — abuse checks are skipped in that case.
func NewService(registry *provider.Registry, guard *abuse.Guard, logger *slog.Logger) *Service {
	return &Service{
		registry: registry,
		guard:    guard,
		logger:   logger,
	}
}

// WithActivation wires the recorder so the first successful IaC export per
// user emits `first_iac_export`. No-op if omitted.
func (s *Service) WithActivation(r ActivationRecorder) *Service {
	s.activation = r
	return s
}

// WithAudit attaches the audit sink. Optional.
func (s *Service) WithAudit(a AuditSink) *Service {
	s.audit = a
	return s
}

// Generate produces IaC code from a diagram.
// plan is the user's billing/usage plan ("free", "pro", "byo") for quota enforcement.
func (s *Service) Generate(ctx context.Context, userID, plan string, req GenerateRequest) (*GenerateResponse, error) {
	if err := ValidateInput(req); err != nil {
		return nil, err
	}

	// Free-plan one-shot export gate. Lifetime-unique, keyed on the same
	// `first_iac_export` activation marker we already record on success —
	// so the first successful export both satisfies and locks the gate.
	// Paid plans skip this entirely. If the activation lookup errors out,
	// HasKind returns false so we fail open (better to grant an extra
	// export than lock a paying-trajectory user).
	if plan == "free" && s.activation != nil {
		if s.activation.HasKind(ctx, userID, "first_iac_export") {
			return nil, ErrFreePlanIaCUsed
		}
	}

	p, err := s.registry.Get(req.Provider)
	if err != nil {
		return nil, err
	}
	isManaged := p.Managed()

	if !isManaged && req.APIKey == "" {
		return nil, ErrMissingAPIKey
	}

	// Pre-checks (abuse guard: circuit breaker, daily quota, prompt limits).
	// We synthesize a short "prompt" string for the guard's validator — the real
	// prompt is built below from the diagram structure.
	if s.guard != nil {
		synthetic := fmt.Sprintf("iac:%s:%s:nodes=%d", req.Format, req.Cloud, len(req.Diagram.Nodes))
		if _, err := s.guard.PreCheck(ctx, userID, plan, synthetic); err != nil {
			return nil, err
		}
	}

	systemPrompt := pickSystemPrompt(req.Format, req.Cloud)
	userPrompt, err := BuildUserPrompt(req.Diagram, req.Format, req.Cloud)
	if err != nil {
		return nil, err
	}

	// Clamp tokens via guard (paid plans get a higher ceiling).
	maxTokens := 4096
	if s.guard != nil {
		maxTokens = s.guard.TokenGuard.ClampTokens(maxTokens, plan)
	}

	callCtx := ctx
	if s.guard != nil {
		if t := s.guard.TimeoutForPlan(plan); t > 0 {
			var cancel context.CancelFunc
			callCtx, cancel = context.WithTimeout(ctx, t)
			defer cancel()
		}
	}

	s.logger.Info("generating IaC",
		"user_id", userID,
		"provider", req.Provider,
		"format", req.Format,
		"cloud", req.Cloud,
		"node_count", len(req.Diagram.Nodes),
	)

	if req.Cloud == CloudAWS {
		return s.generateStructured(callCtx, ctx, p, userID, req, maxTokens)
	}

	start := time.Now()
	raw, err := p.GenerateText(callCtx, provider.GenerateRequest{
		Prompt:       userPrompt,
		SystemPrompt: systemPrompt,
		APIKey:       req.APIKey,
		Config:       provider.ProviderConfig{MaxTokens: maxTokens, Temperature: 0.2},
	})
	latency := time.Since(start)

	if err != nil {
		s.logger.Error("iac provider failed", "provider", req.Provider, "error", err, "latency", latency)
		if s.guard != nil {
			s.guard.PostFailure(ctx, userID, req.Provider, latency)
		}
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	code := stripCodeFences(raw)
	code = strings.TrimSpace(code)

	if code == "" {
		return nil, errors.New("provider returned empty code")
	}

	warnings := collectWarnings(code, req.Format)

	if s.guard != nil {
		// Count resources as "nodes" for quota purposes — bounded by MaxDiagramNodes.
		s.guard.PostSuccess(ctx, userID, req.Provider, len(req.Diagram.Nodes), latency)
	}
	if s.activation != nil {
		s.activation.Record(ctx, userID, "first_iac_export")
	}
	if s.audit != nil {
		s.audit.Log(ctx, "iac.exported", map[string]any{
			"actor_user_id": userID,
			"format":        string(req.Format),
			"cloud":         string(req.Cloud),
			"plan":          plan,
		})
	}

	resp := &GenerateResponse{
		Code:     code,
		Format:   req.Format,
		Language: LanguageHint(req.Format),
		Warnings: warnings,
	}
	if req.Cloud != CloudAWS {
		resp.Warnings = append(resp.Warnings, "structured generator is AWS-only in v1; this output used the legacy prompt-only path")
	}

	if req.Explain {
		explanation, exErr := s.explainCode(callCtx, p, req, code)
		if exErr != nil {
			// Don't fail the whole request — explanation is best-effort.
			s.logger.Warn("iac explain failed, returning code without explanation", "error", exErr)
		} else {
			resp.Explanation = explanation
		}
	}

	return resp, nil
}

func (s *Service) generateStructured(callCtx, originalCtx context.Context, p provider.Provider, userID string, req GenerateRequest, maxTokens int) (*GenerateResponse, error) {
	systemPrompt := BuildStructuredGraphSystemPrompt()
	userPrompt, err := BuildStructuredGraphPrompt(req.Diagram)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	raw, err := p.GenerateStructured(callCtx, provider.GenerateRequest{
		Prompt:       userPrompt,
		SystemPrompt: systemPrompt,
		APIKey:       req.APIKey,
		Config:       provider.ProviderConfig{MaxTokens: maxTokens, Temperature: 0.1},
	}, StructuredGraphSchema())
	latency := time.Since(start)
	if err != nil {
		s.logger.Error("structured iac provider failed", "provider", req.Provider, "error", err, "latency", latency)
		if s.guard != nil {
			s.guard.PostFailure(originalCtx, userID, req.Provider, latency)
		}
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	graphDoc, parseErr := parseGraph(raw)
	if parseErr != nil {
		s.logger.Warn("structured graph parse failed, attempting repair",
			"provider", req.Provider, "cloud", req.Cloud, "error", parseErr)
		var repairErr error
		graphDoc, repairErr = s.repairGraph(callCtx, p, req, raw, parseErr)
		if repairErr != nil {
			s.logger.Error("structured graph repair failed",
				"provider", req.Provider, "cloud", req.Cloud,
				"parse_error", parseErr, "error", repairErr)
			if s.guard != nil {
				s.guard.PostFailure(originalCtx, userID, req.Provider, latency)
			}
			return nil, repairErr
		}
	}

	var code string
	switch req.Format {
	case FormatCloudFormation:
		code, err = render.RenderCloudFormation(graphDoc)
	default:
		code, err = render.RenderTerraform(graphDoc)
	}
	if err != nil {
		s.logger.Error("iac render failed",
			"provider", req.Provider, "cloud", req.Cloud, "format", req.Format, "error", err)
		if s.guard != nil {
			s.guard.PostFailure(originalCtx, userID, req.Provider, latency)
		}
		return nil, fmt.Errorf("render failed: %w", err)
	}

	if s.guard != nil {
		s.guard.PostSuccess(originalCtx, userID, req.Provider, len(req.Diagram.Nodes), latency)
	}
	if s.activation != nil {
		s.activation.Record(originalCtx, userID, "first_iac_export")
	}
	if s.audit != nil {
		s.audit.Log(originalCtx, "iac.exported", map[string]any{
			"actor_user_id": userID,
			"format":        string(req.Format),
			"cloud":         string(req.Cloud),
			"path":          "structured",
		})
	}

	resp := &GenerateResponse{
		Code:     code,
		Format:   req.Format,
		Language: LanguageHint(req.Format),
		Warnings: collectWarnings(code, req.Format),
	}
	if req.Explain {
		explanation, exErr := s.explainCode(callCtx, p, req, code)
		if exErr != nil {
			s.logger.Warn("iac explain failed, returning code without explanation", "error", exErr)
		} else {
			resp.Explanation = explanation
		}
	}
	return resp, nil
}

func parseGraph(raw string) (*graph.ResourceGraph, error) {
	clean := stripCodeFences(raw)
	var doc graph.ResourceGraph
	if err := json.Unmarshal([]byte(clean), &doc); err != nil {
		return nil, fmt.Errorf("%w: could not parse resource graph JSON: %v", provider.ErrInvalidResponse, err)
	}
	if err := graph.Validate(&doc); err != nil {
		return nil, fmt.Errorf("%w: %v", provider.ErrInvalidResponse, err)
	}
	return &doc, nil
}

func (s *Service) repairGraph(ctx context.Context, p provider.Provider, req GenerateRequest, raw string, parseErr error) (*graph.ResourceGraph, error) {
	validationErrors := []string{parseErr.Error()}
	var ve *graph.ValidationError
	if errors.As(parseErr, &ve) {
		validationErrors = ve.Errors
	}
	prompt, err := BuildStructuredGraphRepairPrompt(req.Diagram, validationErrors)
	if err != nil {
		return nil, err
	}
	repaired, err := p.GenerateStructured(ctx, provider.GenerateRequest{
		Prompt:       prompt,
		SystemPrompt: BuildStructuredGraphSystemPrompt(),
		APIKey:       req.APIKey,
		Config:       provider.ProviderConfig{MaxTokens: 4096, Temperature: 0.1},
	}, StructuredGraphSchema())
	if err != nil {
		return nil, fmt.Errorf("graph repair failed: %w", err)
	}
	return parseGraph(repaired)
}

func (s *Service) explainCode(ctx context.Context, p provider.Provider, req GenerateRequest, code string) (string, error) {
	prompt := BuildExplainPrompt(code, req.Format)
	out, err := p.GenerateText(ctx, provider.GenerateRequest{
		Prompt:       prompt,
		SystemPrompt: "You are a concise cloud engineer. Respond with plain-text bullets only.",
		APIKey:       req.APIKey,
		Config:       provider.ProviderConfig{MaxTokens: 600, Temperature: 0.2},
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stripCodeFences(out)), nil
}

func pickSystemPrompt(format Format, cloud Cloud) string {
	switch format {
	case FormatCloudFormation:
		return BuildCloudFormationPrompt(cloud)
	default:
		return BuildTerraformPrompt(cloud)
	}
}

// collectWarnings runs light sanity checks on generated code.
// We keep these cheap and string-based — deep validation is out of scope.
func collectWarnings(code string, format Format) []string {
	var warnings []string
	if len(code) < 80 {
		warnings = append(warnings, "generated code is very short — review carefully before use")
	}
	lower := strings.ToLower(code)
	switch format {
	case FormatTerraform:
		if !strings.Contains(lower, "resource") {
			warnings = append(warnings, "no `resource` blocks found — output may be incomplete")
		}
	case FormatCloudFormation:
		if !strings.Contains(lower, "resources:") {
			warnings = append(warnings, "no Resources section found — output may be incomplete")
		}
	}
	return warnings
}

// stripCodeFences removes markdown code fences and <think>…</think> blocks
// that some models emit. Mirrors the diagram package but kept local to avoid
// cross-package coupling.
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
