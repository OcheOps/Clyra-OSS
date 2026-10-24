package publicdemo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oche/diagram/internal/diagram"
	"github.com/oche/diagram/internal/provider"
)

// After the open-source wind-down the public demo always serves the canned
// fallback (no managed inference available). The validation paths exercised
// here are kept in generateLive for a future self-hoster who re-enables
// Bedrock, and we assert the live-flow validations against that function
// directly so the invariants don't rot.

func TestGenerate_AlwaysServesFallback(t *testing.T) {
	h := New(nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/demo/generate",
		strings.NewReader(`{"prompt":"anything"}`))
	rec := httptest.NewRecorder()
	h.Generate(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("post-shutdown Generate should return 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "fallback") &&
		!strings.Contains(rec.Body.String(), "paused") {
		t.Fatalf("expected fallback payload, got %q", rec.Body.String())
	}
}

func TestGenerateLive_RejectsEmptyPrompt(t *testing.T) {
	h := New(nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/demo/generate",
		strings.NewReader(`{"prompt":""}`))
	rec := httptest.NewRecorder()
	h.generateLive(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty prompt: status = %d, want 400", rec.Code)
	}
}

func TestGenerateLive_RejectsPromptTooLong(t *testing.T) {
	h := New(nil, nil, nil)
	long := strings.Repeat("a", MaxPromptChars+1)
	body := `{"prompt":"` + long + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/demo/generate",
		strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.generateLive(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("long prompt: status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "400") && !strings.Contains(rec.Body.String(), "long") {
		t.Fatalf("expected reason in response body, got %q", rec.Body.String())
	}
}

func TestTruncate_CapsNodesAndDropsDanglingEdges(t *testing.T) {
	nodes := make([]provider.Node, MaxNodes+5)
	for i := range nodes {
		nodes[i] = provider.Node{ID: idFor(i), Label: "n"}
	}
	edges := []provider.Edge{
		{ID: "e1", Source: idFor(0), Target: idFor(1)},          // survives
		{ID: "e2", Source: idFor(0), Target: idFor(MaxNodes + 2)}, // dropped — target cut
		{ID: "e3", Source: idFor(MaxNodes + 3), Target: idFor(1)}, // dropped — source cut
	}
	out := truncate(&diagram.GenerateResponse{Nodes: nodes, Edges: edges})
	if len(out.Nodes) != MaxNodes {
		t.Fatalf("nodes = %d, want %d", len(out.Nodes), MaxNodes)
	}
	if len(out.Edges) != 1 || out.Edges[0].ID != "e1" {
		t.Fatalf("edges = %+v, want only e1", out.Edges)
	}
}

func TestTruncate_CapsEdges(t *testing.T) {
	// Both endpoints survive; only the count-cap should apply.
	nodes := []provider.Node{{ID: "a"}, {ID: "b"}}
	edges := make([]provider.Edge, MaxEdges+5)
	for i := range edges {
		edges[i] = provider.Edge{ID: idFor(i), Source: "a", Target: "b"}
	}
	out := truncate(&diagram.GenerateResponse{Nodes: nodes, Edges: edges})
	if len(out.Edges) != MaxEdges {
		t.Fatalf("edges = %d, want %d", len(out.Edges), MaxEdges)
	}
}

func TestGenerateLive_MalformedJSON400(t *testing.T) {
	h := New(nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/demo/generate",
		strings.NewReader(`{malformed`))
	rec := httptest.NewRecorder()
	h.generateLive(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed json: status = %d, want 400", rec.Code)
	}
}

func TestServeFallback_ShapeAndStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	serveFallback(rec, "test")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("fallback status = %d, want 503", rec.Code)
	}
	var envelope struct {
		Data demoResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("fallback body not JSON: %v", err)
	}
	got := envelope.Data
	if !got.Fallback {
		t.Fatal("fallback flag not set")
	}
	if len(got.Nodes) == 0 || len(got.Edges) == 0 {
		t.Fatal("fallback should carry the canned 3-tier diagram")
	}
}

func idFor(i int) string {
	return "n" + string(rune('0'+i%10)) + string(rune('a'+i/10))
}
