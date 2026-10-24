package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/oche/diagram/internal/cloudsync"
	"github.com/oche/diagram/internal/diagram"
	iacreverse "github.com/oche/diagram/internal/iac/reverse"
	"github.com/oche/diagram/internal/provider"
)

// ToolDeps carries the domain services the tools call. Passed by value —
// each field is a pointer / interface, cheap to copy.
type ToolDeps struct {
	Diagram          *diagram.Service
	Cloud            *cloudsync.Service
	AppURL           string
	AllowedProviders []string // e.g. []{"managed_bedrock"} — restricts MCP-side generation to managed path only, since MCP callers can't pass their own API key safely today
}

// ToolResult is the common envelope every tool returns. Kept minimal so
// LLM clients can reliably parse it without brittle prose parsing.
type ToolResult struct {
	OK        bool   `json:"ok"`
	URL       string `json:"url"`
	DiagramID string `json:"diagram_id"`
	Summary   string `json:"summary"`
	NodeCount int    `json:"node_count"`
	EdgeCount int    `json:"edge_count"`
	Message   string `json:"message"`
}

// =====================================================================
// generate_diagram
// =====================================================================

var generateDiagramSchema = json.RawMessage(`{
  "type": "object",
  "required": ["prompt"],
  "properties": {
    "prompt": {
      "type": "string",
      "minLength": 20,
	      "maxLength": 2000,
      "description": "Natural-language description of the system to design. Examples: 'A 3-tier web app on AWS with ALB, ECS Fargate, RDS Postgres, and S3 media bucket.'"
    },
    "mode": {
      "type": "string",
      "enum": ["solution_architecture", "network_architecture", "enterprise_architecture"],
      "default": "solution_architecture",
      "description": "Which lens the diagram takes. Default is solution_architecture (cloud-native, service-focused)."
    },
    "title": {
      "type": "string",
      "maxLength": 140,
      "description": "Optional project title. Falls back to the first 60 chars of the prompt."
    }
  }
}`)

var toolResultSchema = json.RawMessage(`{
  "type":"object",
  "required":["ok","url","diagram_id","summary","node_count","edge_count","message"],
  "properties":{
    "ok":{"type":"boolean"},
    "url":{"type":"string"},
    "diagram_id":{"type":"string"},
    "summary":{"type":"string"},
    "node_count":{"type":"integer","minimum":0},
    "edge_count":{"type":"integer","minimum":0},
    "message":{"type":"string"}
  }
}`)

type generateInput struct {
	Prompt string `json:"prompt"`
	Mode   string `json:"mode"`
	Title  string `json:"title"`
}

// NewGenerateDiagramTool wires the generate_diagram MCP tool. Calls the
// existing diagram service (same code path the SPA uses — same plan
// checks, same abuse gate, same normalization + salvage + validate),
// then persists the result to the user's project library so the
// returned URL opens cleanly in the editor.
func NewGenerateDiagramTool(deps ToolDeps) Tool {
	return Tool{
		Name:  "generate_diagram",
		Title: "Generate architecture diagram",
		Description: "Turn a plain-English description of a cloud system into an editable " +
			"architecture diagram in Clyra. The diagram is saved to the caller's project " +
			"library and a shareable URL is returned. Users can then edit visually and " +
			"export to Terraform or CloudFormation from the returned link. Best for " +
			"AWS-native designs; GCP + Azure supported.",
		InputSchema:      generateDiagramSchema,
		OutputSchema:     toolResultSchema,
		ChargesManagedAI: true,
		Handler: func(ctx context.Context, tc *ToolContext, args json.RawMessage) (any, error) {
			var in generateInput
			if err := decodeStrict(args, &in); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}
			prompt := strings.TrimSpace(in.Prompt)
			if len(prompt) < 20 {
				return nil, errors.New("prompt too short — need at least 20 chars of description")
			}
			if len(prompt) > 2000 {
				return nil, errors.New("prompt too long — maximum is 2000 characters")
			}
			if len(in.Title) > 140 {
				return nil, errors.New("title too long — maximum is 140 characters")
			}
			mode := in.Mode
			if mode == "" {
				mode = "solution_architecture"
			}
			switch mode {
			case "solution_architecture", "network_architecture", "enterprise_architecture":
			default:
				return nil, errors.New("mode must be solution_architecture, network_architecture, or enterprise_architecture")
			}

			providerName := "managed_bedrock"
			if len(deps.AllowedProviders) > 0 {
				providerName = deps.AllowedProviders[0]
			}

			req := diagram.GenerateRequest{
				Prompt:   prompt,
				Provider: providerName,
				Mode:     mode,
			}
			plan := string(tc.User.EffectivePlan())
			resp, err := deps.Diagram.Generate(ctx, tc.User.UserID, plan, req)
			if err != nil {
				return nil, fmt.Errorf("generation failed: %w", err)
			}

			project, err := saveToLibrary(ctx, deps.Cloud, tc.User.UserID, deriveTitle(in.Title, prompt), resp)
			if err != nil {
				return nil, fmt.Errorf("saved generation but failed to persist to library: %w", err)
			}

			return ToolResult{
				OK:        true,
				URL:       diagramURL(deps.AppURL, project.ID),
				DiagramID: project.ID,
				Summary:   summarize(resp),
				NodeCount: len(resp.Nodes),
				EdgeCount: len(resp.Edges),
				Message: fmt.Sprintf(
					"Generated a %d-node diagram. Open %s to edit and export Terraform.",
					len(resp.Nodes), diagramURL(deps.AppURL, project.ID),
				),
			}, nil
		},
	}
}

// =====================================================================
// reverse_from_iac
// =====================================================================

var reverseFromIaCSchema = json.RawMessage(`{
  "type": "object",
  "required": ["source"],
  "properties": {
    "source": {
      "type": "string",
      "minLength": 20,
      "maxLength": 50000,
      "description": "Raw Terraform HCL source. Multi-resource is fine (paste an entire .tf file). CloudFormation coming soon."
    },
    "title": {
      "type": "string",
      "maxLength": 140,
      "description": "Optional project title. Defaults to 'Imported Terraform'."
    }
  }
}`)

type reverseInput struct {
	Source string `json:"source"`
	Title  string `json:"title"`
}

// NewReverseFromIaCTool wires the reverse_from_iac MCP tool. Runs the
// same HCL → graph → diagram pipeline the /iac/ingest endpoint uses.
// No provider call, no daily quota (import ≠ generation), just parses
// and maps.
func NewReverseFromIaCTool(deps ToolDeps) Tool {
	return Tool{
		Name:  "reverse_from_iac",
		Title: "Import Terraform as a diagram",
		Description: "Parse existing Terraform HCL and render it as an editable Clyra " +
			"diagram. Great for visualizing existing infrastructure, spotting SPOFs, or " +
			"handing a project off. Returns a URL to open the imported diagram in Clyra.",
		InputSchema:      reverseFromIaCSchema,
		OutputSchema:     toolResultSchema,
		ChargesManagedAI: false,
		Handler: func(ctx context.Context, tc *ToolContext, args json.RawMessage) (any, error) {
			var in reverseInput
			if err := decodeStrict(args, &in); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}
			src := strings.TrimSpace(in.Source)
			if len(src) < 20 {
				return nil, errors.New("source too short — paste at least one full resource block")
			}
			if len(src) > iacreverse.MaxBodyBytes {
				return nil, fmt.Errorf("source too large — max %d bytes", iacreverse.MaxBodyBytes)
			}
			if len(in.Title) > 140 {
				return nil, errors.New("title too long — maximum is 140 characters")
			}

			blocks, _, perr := iacreverse.Parse("import.tf", src)
			if perr != nil {
				return nil, fmt.Errorf("invalid HCL: %w", perr)
			}
			if len(blocks) == 0 {
				return nil, errors.New("no architecture resources found in source")
			}

			result := iacreverse.Build(blocks)
			diag := iacreverse.ToDiagram(result)
			diagram.Normalize(diag)
			diagram.Salvage(diag)
			if ve := diagram.Validate(diag); ve != nil {
				return nil, fmt.Errorf("import produced invalid diagram: %s", ve.Error())
			}

			title := in.Title
			if title == "" {
				title = "Imported Terraform"
			}
			project, err := saveToLibrary(ctx, deps.Cloud, tc.User.UserID, title, diag)
			if err != nil {
				return nil, fmt.Errorf("import succeeded but failed to save: %w", err)
			}

			return ToolResult{
				OK:        true,
				URL:       diagramURL(deps.AppURL, project.ID),
				DiagramID: project.ID,
				Summary:   summarize(diag),
				NodeCount: len(diag.Nodes),
				EdgeCount: len(diag.Edges),
				Message: fmt.Sprintf(
					"Imported %d resources into a diagram. Open %s to edit.",
					len(diag.Nodes), diagramURL(deps.AppURL, project.ID),
				),
			}, nil
		},
	}
}

// =====================================================================
// shared helpers
// =====================================================================

// saveToLibrary persists a diagram as a fresh project in the user's
// library. Reuses cloudsync.Service.Save so per-plan quotas and audit
// events fire consistently with the in-app path.
func saveToLibrary(ctx context.Context, cs *cloudsync.Service, userID, title string, resp any) (*cloudsync.Project, error) {
	return cs.Save(ctx, userID, cloudsync.SaveRequest{
		Title:   title,
		Diagram: resp,
		Source:  cloudsync.SourceMCP,
	})
}

func decodeStrict(raw json.RawMessage, dst any) error {
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func diagramURL(appURL, id string) string {
	return strings.TrimRight(appURL, "/") + "/editor/" + id
}

// deriveTitle uses caller-supplied title when present, otherwise first
// 60 chars of the prompt with whitespace collapsed. Matches docs/mcp.md
// resolution #2.
func deriveTitle(title, prompt string) string {
	if t := strings.TrimSpace(title); t != "" {
		if len(t) > 140 {
			return t[:140]
		}
		return t
	}
	compact := strings.Join(strings.Fields(prompt), " ")
	if len(compact) <= 60 {
		return compact
	}
	return strings.TrimSpace(compact[:60]) + "…"
}

// summarize builds a short human string for Claude to relay: node count,
// edge count, dominant family. Doesn't try to be smart about it — just
// enough for the LLM to acknowledge to the user in one line.
func summarize(v any) string {
	switch d := v.(type) {
	case *diagram.GenerateResponse:
		return summarizeCounts(d.Nodes, d.Edges)
	case *provider.GenerateResponse:
		return summarizeCounts(d.Nodes, d.Edges)
	default:
		return "diagram ready"
	}
}

func summarizeCounts(nodes []provider.Node, edges []provider.Edge) string {
	families := map[string]int{}
	for _, n := range nodes {
		if idx := strings.IndexByte(n.Type, '.'); idx > 0 {
			families[n.Type[:idx]]++
		}
	}
	dominant := "generic"
	max := 0
	for f, c := range families {
		if c > max {
			max = c
			dominant = f
		}
	}
	return fmt.Sprintf("%d nodes, %d edges, %s-focused", len(nodes), len(edges), dominant)
}
