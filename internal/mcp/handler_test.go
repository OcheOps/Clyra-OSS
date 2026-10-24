package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oche/diagram/internal/auth"
)

type countingLimiter struct{ calls int }

func (l *countingLimiter) Allow(context.Context, string) (bool, int, time.Time) {
	l.calls++
	return true, 19, time.Now().Add(24 * time.Hour)
}

func TestStreamableHTTPListsAndCallsTools(t *testing.T) {
	limiter := &countingLimiter{}
	h := NewHandler(HandlerOpts{Limiter: limiter, AppURL: "https://example.test"})
	makeTool := func(name string, managed bool) Tool {
		return Tool{
			Name:             name,
			Title:            name,
			Description:      "test tool",
			InputSchema:      json.RawMessage(`{"type":"object"}`),
			OutputSchema:     json.RawMessage(`{"type":"object"}`),
			ChargesManagedAI: managed,
			Handler: func(context.Context, *ToolContext, json.RawMessage) (any, error) {
				return map[string]any{"ok": true}, nil
			},
		}
	}
	h.Register(makeTool("local_tool", false))
	h.Register(makeTool("managed_tool", true))

	user := &auth.SessionUser{UserID: "user-1", Plan: auth.PlanPro}
	request := func(body string, sessionID string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "http://example.test/mcp", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		if sessionID != "" {
			req.Header.Set("Mcp-Session-Id", sessionID)
		}
		recorder := httptest.NewRecorder()
		h.ServeHTTP(recorder, req.WithContext(auth.NewContext(req.Context(), user)))
		return recorder
	}

	initialized := request(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`, "")
	if initialized.Code != http.StatusOK {
		t.Fatalf("initialize status=%d body=%s", initialized.Code, initialized.Body.String())
	}
	sessionID := initialized.Header().Get("Mcp-Session-Id")
	if sessionID == "" {
		t.Fatal("initialize omitted MCP session id")
	}

	listed := request(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`, sessionID)
	if listed.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
	var listEnvelope struct {
		Result struct {
			Tools []json.RawMessage `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(listed.Body.Bytes(), &listEnvelope); err != nil {
		t.Fatalf("decode tool list: %v", err)
	}
	if len(listEnvelope.Result.Tools) != 2 {
		t.Fatalf("got %d tools, want 2", len(listEnvelope.Result.Tools))
	}

	call := func(id int, name string) struct {
		Result struct {
			IsError           bool           `json:"isError"`
			StructuredContent map[string]any `json:"structuredContent"`
		} `json:"result"`
	} {
		response := request(`{"jsonrpc":"2.0","id":`+fmt.Sprint(id)+`,"method":"tools/call","params":{"name":"`+name+`","arguments":{}}}`, sessionID)
		if response.Code != http.StatusOK {
			t.Fatalf("call %s status=%d body=%s", name, response.Code, response.Body.String())
		}
		var envelope struct {
			Result struct {
				IsError           bool           `json:"isError"`
				StructuredContent map[string]any `json:"structuredContent"`
			} `json:"result"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("decode %s result: %v", name, err)
		}
		return envelope
	}

	local := call(3, "local_tool")
	if local.Result.IsError {
		t.Fatalf("local tool returned error: %+v", local)
	}
	if limiter.calls != 0 {
		t.Fatalf("local tool consumed managed quota: %d calls", limiter.calls)
	}

	managed := call(4, "managed_tool")
	if managed.Result.IsError {
		t.Fatalf("managed tool returned error: %+v", managed)
	}
	if limiter.calls != 1 {
		t.Fatalf("managed tool quota calls = %d, want 1", limiter.calls)
	}
	if managed.Result.StructuredContent == nil {
		t.Fatal("managed tool omitted structured content")
	}
}

func TestToolFailuresUseMCPErrorResult(t *testing.T) {
	h := NewHandler(HandlerOpts{})
	tool := Tool{Name: "paid", ChargesManagedAI: true}
	ctx := auth.NewContext(context.Background(), &auth.SessionUser{UserID: "free-user", Plan: auth.PlanFree})
	result, err := h.callTool(ctx, tool, nil)
	if err != nil {
		t.Fatalf("callTool returned protocol error: %v", err)
	}
	if !result.IsError || len(result.Content) == 0 {
		t.Fatalf("expected MCP tool error result, got %+v", result)
	}
}
