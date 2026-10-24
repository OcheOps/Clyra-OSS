// Package mcp exposes Clyra tools over the official Model Context Protocol Go
// SDK. The SDK owns protocol negotiation, Streamable HTTP, JSON-RPC framing,
// content negotiation, notifications, and request-size enforcement.
package mcp

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	protocol "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/oche/diagram/internal/auth"
)

type Tool struct {
	Name             string
	Title            string
	Description      string
	InputSchema      json.RawMessage
	OutputSchema     json.RawMessage
	ChargesManagedAI bool
	Handler          func(context.Context, *ToolContext, json.RawMessage) (any, error)
}

type ToolContext struct {
	User   *auth.SessionUser
	AppURL string
}

type PlanGate interface {
	Allow(user *auth.SessionUser) bool
}

type planGateFunc func(*auth.SessionUser) bool

func (f planGateFunc) Allow(u *auth.SessionUser) bool { return f(u) }

func DefaultPlanGate() PlanGate {
	return planGateFunc(func(u *auth.SessionUser) bool {
		return u != nil && u.EffectivePlan() != auth.PlanFree
	})
}

type AuditSink interface {
	Log(ctx context.Context, event string, fields map[string]any)
}

type Handler struct {
	server   *protocol.Server
	http     http.Handler
	tools    []Tool
	planGate PlanGate
	limiter  DailyLimiter
	logger   *slog.Logger
	audit    AuditSink
	appURL   string
}

type HandlerOpts struct {
	PlanGate PlanGate
	Limiter  DailyLimiter
	Logger   *slog.Logger
	Audit    AuditSink
	AppURL   string
}

const DefaultDailyLimit = 20

func NewHandler(opts HandlerOpts) *Handler {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	gate := opts.PlanGate
	if gate == nil {
		gate = DefaultPlanGate()
	}
	limiter := opts.Limiter
	if limiter == nil {
		limiter = NewMemoryDailyLimiter(DefaultDailyLimit)
	}
	appURL := strings.TrimRight(opts.AppURL, "/")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}

	server := protocol.NewServer(&protocol.Implementation{
		Name:    "clyra",
		Version: "0.3.0",
	}, &protocol.ServerOptions{
		Instructions: "Create editable cloud architecture diagrams with generate_diagram or import Terraform with reverse_from_iac.",
		Logger:       logger,
		Capabilities: &protocol.ServerCapabilities{},
	})
	h := &Handler{
		server: server, planGate: gate, limiter: limiter,
		logger: logger, audit: opts.Audit, appURL: appURL,
	}
	h.http = protocol.NewStreamableHTTPHandler(
		func(*http.Request) *protocol.Server { return server },
		&protocol.StreamableHTTPOptions{
			Stateless:             false,
			JSONResponse:          true,
			Logger:                logger,
			CrossOriginProtection: &http.CrossOriginProtection{},
			MaxRequestBodyBytes:   1 << 20,
			SessionTimeout:        30 * time.Minute,
		},
	)
	return h
}

func (h *Handler) Register(tool Tool) {
	h.tools = append(h.tools, tool)
	destructive := false
	openWorld := tool.ChargesManagedAI
	h.server.AddTool(&protocol.Tool{
		Name:         tool.Name,
		Title:        tool.Title,
		Description:  tool.Description,
		InputSchema:  tool.InputSchema,
		OutputSchema: tool.OutputSchema,
		Annotations: &protocol.ToolAnnotations{
			DestructiveHint: &destructive,
			OpenWorldHint:   &openWorld,
			ReadOnlyHint:    false,
			IdempotentHint:  false,
		},
	}, func(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		return h.callTool(ctx, tool, req)
	})
}

func (h *Handler) WellKnown(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{
		"name": "clyra", "endpoint": "/mcp", "transport": "streamable-http",
		"authentication": "bearer",
	})
}

func (h *Handler) ListToolsHTTP(w http.ResponseWriter, _ *http.Request) {
	type publicTool struct {
		Name         string          `json:"name"`
		Title        string          `json:"title,omitempty"`
		Description  string          `json:"description"`
		InputSchema  json.RawMessage `json:"inputSchema"`
		OutputSchema json.RawMessage `json:"outputSchema,omitempty"`
	}
	out := make([]publicTool, 0, len(h.tools))
	for _, tool := range h.tools {
		out = append(out, publicTool{tool.Name, tool.Title, tool.Description, tool.InputSchema, tool.OutputSchema})
	}
	writeJSON(w, map[string]any{"tools": out})
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.http.ServeHTTP(w, r)
}

func (h *Handler) callTool(ctx context.Context, tool Tool, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		h.auditCall(ctx, nil, tool.Name, false, 0, "unauthorized", "")
		return toolError("unauthorized — supply a Clyra personal access token"), nil
	}
	if !h.planGate.Allow(user) {
		h.auditCall(ctx, user, tool.Name, false, 0, "payment_required", "")
		return toolError("Clyra MCP requires a Pro or BYO plan — upgrade at " + h.appURL + "/billing"), nil
	}
	if tool.ChargesManagedAI {
		allowed, remaining, resetAt := h.limiter.Allow(ctx, user.UserID)
		if !allowed {
			h.auditCall(ctx, user, tool.Name, false, 0, "rate_limited", "")
			return toolError("daily managed-AI quota exhausted — resets at " + resetAt.Format(time.RFC3339)), nil
		}
		if remaining >= 0 && remaining <= 3 {
			h.logger.Info("mcp quota nearing limit", "user_id", user.UserID, "tool", tool.Name, "remaining", remaining)
		}
	}

	var args json.RawMessage
	if req != nil && req.Params != nil {
		args = req.Params.Arguments
	}
	start := time.Now()
	result, err := tool.Handler(ctx, &ToolContext{User: user, AppURL: h.appURL}, args)
	duration := time.Since(start)
	if err != nil {
		h.logger.Warn("mcp.tool_called", "tool", tool.Name, "user_id", user.UserID,
			"duration_ms", duration.Milliseconds(), "ok", false, "error", err.Error())
		h.auditCall(ctx, user, tool.Name, false, duration, "tool_error", "")
		return toolError(err.Error()), nil
	}

	payload, err := json.Marshal(result)
	if err != nil {
		h.auditCall(ctx, user, tool.Name, false, duration, "encoding_error", "")
		return toolError("failed to encode tool result"), nil
	}
	h.logger.Info("mcp.tool_called", "tool", tool.Name, "user_id", user.UserID,
		"duration_ms", duration.Milliseconds(), "ok", true)
	diagramID := ""
	if toolResult, ok := result.(ToolResult); ok {
		diagramID = toolResult.DiagramID
	}
	h.auditCall(ctx, user, tool.Name, true, duration, "", diagramID)
	return &protocol.CallToolResult{
		Content:           []protocol.Content{&protocol.TextContent{Text: string(payload)}},
		StructuredContent: result,
	}, nil
}

func toolError(message string) *protocol.CallToolResult {
	return &protocol.CallToolResult{
		Content: []protocol.Content{&protocol.TextContent{Text: message}},
		IsError: true,
	}
}

func (h *Handler) auditCall(ctx context.Context, user *auth.SessionUser, tool string, ok bool, duration time.Duration, reason, diagramID string) {
	if h.audit == nil {
		return
	}
	fields := map[string]any{
		"tool": tool, "ok": ok, "duration_ms": duration.Milliseconds(),
	}
	if user != nil {
		fields["actor_user_id"] = user.UserID
		fields["plan"] = string(user.EffectivePlan())
	}
	if reason != "" {
		fields["reason"] = reason
	}
	if diagramID != "" {
		fields["project_id"] = diagramID
	}
	h.audit.Log(ctx, "mcp.tool_called", fields)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Default().Warn("mcp: encode discovery response", "error", err)
	}
}
