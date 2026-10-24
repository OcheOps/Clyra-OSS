package reverse

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oche/diagram/internal/auth"
)

// withUser puts a session user into a request context — bypasses the real
// auth middleware so we can exercise the handler logic directly.
func withUser(r *http.Request, plan auth.Plan) *http.Request {
	su := &auth.SessionUser{
		UserID: "test-user",
		Email:  "t@example.com",
		Role:   auth.RoleUser,
		Plan:   plan,
	}
	return r.WithContext(auth.NewContext(context.Background(), su))
}

func TestIngestHandler_HappyPath(t *testing.T) {
	// Nil guard skips quota/circuit-breaker — that's exercised separately.
	h := NewHandler(nil)

	body, _ := json.Marshal(IngestRequest{
		Source: `
resource "aws_vpc" "main" { cidr_block = "10.0.0.0/16" }
resource "aws_subnet" "public" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}`,
	})
	req := withUser(httptest.NewRequest(http.MethodPost, "/api/v1/iac/ingest", bytes.NewReader(body)), auth.PlanFree)
	w := httptest.NewRecorder()
	h.Ingest(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	// Response is wrapped by pkg/response — it lives under "data".
	var envelope struct {
		Data IngestResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("response decode: %v; body: %s", err, w.Body.String())
	}
	// Diagram came back as map[string]any (since we typed it as `any`).
	diag, ok := envelope.Data.Diagram.(map[string]any)
	if !ok {
		t.Fatalf("diagram not an object: %T", envelope.Data.Diagram)
	}
	nodes, _ := diag["nodes"].([]any)
	if len(nodes) != 2 {
		t.Errorf("want 2 nodes, got %d", len(nodes))
	}
	edges, _ := diag["edges"].([]any)
	if len(edges) != 1 {
		t.Errorf("want 1 edge, got %d", len(edges))
	}
}

func TestIngestHandler_Unauthenticated(t *testing.T) {
	h := NewHandler(nil)
	body, _ := json.Marshal(IngestRequest{Source: `resource "aws_vpc" "a" { cidr_block = "10.0.0.0/16" }`})
	// No withUser — context has no SessionUser.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/iac/ingest", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Ingest(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestIngestHandler_BodyTooLarge(t *testing.T) {
	h := NewHandler(nil)
	// Build a payload that crosses the 50KB cap. Repeat-pad past the limit.
	pad := strings.Repeat("a", MaxBodyBytes+200)
	body, _ := json.Marshal(IngestRequest{Source: pad})
	req := withUser(httptest.NewRequest(http.MethodPost, "/api/v1/iac/ingest", bytes.NewReader(body)), auth.PlanFree)
	w := httptest.NewRecorder()
	h.Ingest(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", w.Code)
	}
}

func TestIngestHandler_InvalidHCL(t *testing.T) {
	h := NewHandler(nil)
	body, _ := json.Marshal(IngestRequest{Source: `resource "aws_vpc" "main" {`}) // unterminated
	req := withUser(httptest.NewRequest(http.MethodPost, "/api/v1/iac/ingest", bytes.NewReader(body)), auth.PlanFree)
	w := httptest.NewRecorder()
	h.Ingest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 on invalid HCL", w.Code)
	}
}

func TestIngestHandler_EmptySource(t *testing.T) {
	h := NewHandler(nil)
	body, _ := json.Marshal(IngestRequest{Source: ""})
	req := withUser(httptest.NewRequest(http.MethodPost, "/api/v1/iac/ingest", bytes.NewReader(body)), auth.PlanFree)
	w := httptest.NewRecorder()
	h.Ingest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 on empty source", w.Code)
	}
}

func TestIngestHandler_NoResourcesFound(t *testing.T) {
	// HCL with only provider + variable blocks — parses cleanly, no
	// resources, should 400 not 200 with an empty diagram.
	h := NewHandler(nil)
	body, _ := json.Marshal(IngestRequest{Source: `
terraform { required_version = ">= 1.5" }
provider "aws" { region = "us-east-1" }
variable "env" { type = string }`})
	req := withUser(httptest.NewRequest(http.MethodPost, "/api/v1/iac/ingest", bytes.NewReader(body)), auth.PlanFree)
	w := httptest.NewRecorder()
	h.Ingest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 when no architecture blocks", w.Code)
	}
}
