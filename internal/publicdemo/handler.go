// Package publicdemo serves the unauthenticated landing-page demo:
// a visitor pastes a prompt on useclyra.com/... and gets a real
// diagram back without signing up. Highest-value conversion surface
// on the site, so it's kept deliberately narrow:
//
//   - IP rate-limited at 1/60s so a single visitor can't burn our
//     Bedrock budget.
//   - Prompt hard-capped at 400 chars.
//   - Response truncated to 12 nodes / 20 edges to keep visual noise
//     under control on the hero canvas.
//   - Managed Bedrock only — never accepts user-supplied keys from
//     an unauth endpoint.
//   - Failure returns a canned fallback diagram (503) so the landing
//     never renders a raw error.
//
// Observability leans on structured logs (event=demo_*) rather than
// a DB row — usage_log requires a user_id and adding a schema branch
// for anonymous rows is not worth the coupling at this traffic level.
package publicdemo

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/oche/diagram/internal/diagram"
	"github.com/oche/diagram/internal/provider"
	"github.com/oche/diagram/pkg/response"
)

const (
	// MaxPromptChars is the hard ceiling on the visitor-supplied prompt.
	// 400 is generous enough for a real system description and small
	// enough to bound the token cost per call.
	MaxPromptChars = 400
	// MaxNodes / MaxEdges truncate the returned diagram so the hero
	// canvas stays readable. Values chosen to match the AnimatedHero
	// scenes' complexity.
	MaxNodes = 12
	MaxEdges = 20
	// demoProvider is the only provider allowed from this endpoint.
	demoProvider = "managed_bedrock"
	demoMode     = "solution_architecture"
	// demoUserID is a synthetic id for abuse-guard bookkeeping — the
	// call is anonymous but the service layer wants a non-empty string.
	demoUserID = "public-demo"
	// demoPlan is the plan the service uses to clamp tokens + apply
	// timeouts. Free-tier limits are the right envelope for a public
	// endpoint (managed Bedrock, tight budget).
	demoPlan = "free"
)

// Handler serves POST /api/v1/public/demo/generate.
type Handler struct {
	svc            *diagram.Service
	logger         *slog.Logger
	allowedOrigins []string
}

// New constructs the demo handler. allowedOrigins is the set of Origin/Referer
// hostnames the browser demo is served from (e.g. "useclyra.com",
// "www.useclyra.com"). Requests whose Origin does not match one of these are
// rejected — the demo is intended for the landing page only, not for third-party
// embeds that would fan out under our Bedrock budget on our IP-limit blind spot.
// Pass nil/empty to disable the check (dev, or when you actually want a public
// demo API).
func New(svc *diagram.Service, logger *slog.Logger, allowedOrigins []string) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{svc: svc, logger: logger, allowedOrigins: allowedOrigins}
}

func (h *Handler) allowedOrigin(r *http.Request) bool {
	if len(h.allowedOrigins) == 0 {
		return true
	}
	// Origin is set on browser CORS-relevant requests; Referer covers same-origin
	// XHRs that don't trigger CORS. We accept a match on either.
	sources := []string{r.Header.Get("Origin"), r.Header.Get("Referer")}
	for _, s := range sources {
		if s == "" {
			continue
		}
		for _, allowed := range h.allowedOrigins {
			if strings.Contains(s, allowed) {
				return true
			}
		}
	}
	return false
}

type demoRequest struct {
	Prompt string `json:"prompt"`
	Cloud  string `json:"cloud,omitempty"` // hint — routed as prompt prefix
}

type demoResponse struct {
	Nodes    []provider.Node  `json:"nodes"`
	Edges    []provider.Edge  `json:"edges"`
	Groups   []provider.Group `json:"groups,omitempty"`
	Fallback bool             `json:"fallback,omitempty"` // true when we served the canned diagram
	Message  string           `json:"message,omitempty"`
}

// Generate serves the demo. Rate-limit middleware runs in front of this;
// this handler only sees requests that already passed the per-IP cap.
//
// Open-source build: managed Bedrock is disabled, so the public landing
// demo cannot generate real diagrams anymore (there's no safe way to use
// a BYO key from an unauth endpoint). Serve the canned fallback diagram
// as "demo preview" and point visitors at signup for their own BYO key.
func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	if !h.allowedOrigin(r) {
		h.logger.Info("demo origin rejected",
			"event", "demo_failed", "reason", "origin",
			"origin", r.Header.Get("Origin"), "referer", r.Header.Get("Referer"))
		response.Error(w, http.StatusForbidden, "not available from this origin")
		return
	}
	// Short-circuit: no managed inference available in this build.
	// A self-hoster who re-enables Bedrock can delete this short-circuit
	// (and the surrounding comment) to restore the live demo flow in
	// generateLive below.
	serveFallback(w, "Public demo is paused — sign up and bring your own OpenAI/Anthropic/Gemini key to run live.")
}

// generateLive is the pre-shutdown flow that calls through to the diagram
// service with the managed provider. Kept as a package-private function so
// the code path isn't lost — not called today because the public Generate
// handler short-circuits to the canned fallback.
func (h *Handler) generateLive(w http.ResponseWriter, r *http.Request) { //nolint:unused
	var body demoRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.logger.Info("demo bad request", "event", "demo_failed", "reason", "decode")
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	prompt := strings.TrimSpace(body.Prompt)
	if prompt == "" {
		response.Error(w, http.StatusBadRequest, "prompt is required")
		return
	}
	if len(prompt) > MaxPromptChars {
		h.logger.Info("demo prompt too long",
			"event", "demo_failed", "reason", "prompt_too_long", "len", len(prompt))
		response.Error(w, http.StatusBadRequest,
			"prompt is too long — keep it under 400 characters")
		return
	}

	// Weave the cloud hint into the prompt without letting a caller
	// inject arbitrary system-prompt-style text. The service prompt
	// normalizer will strip anything it doesn't recognize.
	if cloud := strings.ToLower(strings.TrimSpace(body.Cloud)); cloud != "" {
		switch cloud {
		case "aws", "gcp", "azure":
			prompt = "On " + strings.ToUpper(cloud) + ": " + prompt
		}
	}

	start := time.Now()
	req := diagram.GenerateRequest{
		Prompt:   prompt,
		Provider: demoProvider,
		Mode:     demoMode,
	}
	resp, err := h.svc.Generate(r.Context(), demoUserID, demoPlan, req)
	latency := time.Since(start)

	if err != nil {
		h.logger.Warn("demo upstream failed",
			"event", "demo_failed", "reason", "upstream",
			"error", err.Error(), "latency_ms", latency.Milliseconds())
		serveFallback(w, "our generator is having a moment — here's a canned example")
		return
	}

	// Truncate large responses so the hero canvas stays legible.
	truncated := truncate(resp)

	h.logger.Info("demo served",
		"event", "demo_served",
		"nodes", len(truncated.Nodes), "edges", len(truncated.Edges),
		"latency_ms", latency.Milliseconds())
	response.JSON(w, http.StatusOK, truncated)
}

func truncate(resp *diagram.GenerateResponse) demoResponse {
	nodes := resp.Nodes
	if len(nodes) > MaxNodes {
		nodes = nodes[:MaxNodes]
	}
	// Only keep edges whose endpoints survived the node cap.
	keep := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		keep[n.ID] = true
	}
	edges := make([]provider.Edge, 0, len(resp.Edges))
	for _, e := range resp.Edges {
		if !keep[e.Source] || !keep[e.Target] {
			continue
		}
		edges = append(edges, e)
		if len(edges) >= MaxEdges {
			break
		}
	}
	return demoResponse{Nodes: nodes, Edges: edges, Groups: resp.Groups}
}

// serveFallback returns a small, hand-authored 3-tier diagram. Same
// shape as a real response so the frontend renders it identically.
// 503 signals to a downstream monitor that upstream failed even though
// the visitor still sees something.
func serveFallback(w http.ResponseWriter, msg string) {
	fallback := demoResponse{
		Nodes: []provider.Node{
			{ID: "route53", Type: "aws.route_53", Label: "Route 53"},
			{ID: "cloudfront", Type: "aws.cloudfront", Label: "CloudFront"},
			{ID: "alb", Type: "aws.alb", Label: "ALB"},
			{ID: "ecs", Type: "aws.ecs", Label: "ECS Fargate"},
			{ID: "rds", Type: "aws.rds", Label: "RDS Postgres"},
			{ID: "s3", Type: "aws.s3", Label: "S3"},
		},
		Edges: []provider.Edge{
			{ID: "e1", Source: "route53", Target: "cloudfront"},
			{ID: "e2", Source: "cloudfront", Target: "alb"},
			{ID: "e3", Source: "alb", Target: "ecs"},
			{ID: "e4", Source: "ecs", Target: "rds"},
			{ID: "e5", Source: "ecs", Target: "s3"},
		},
		Fallback: true,
		Message:  msg,
	}
	response.JSON(w, http.StatusServiceUnavailable, fallback)
}
