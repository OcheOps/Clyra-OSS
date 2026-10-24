package iac

import "github.com/oche/diagram/internal/provider"

// Format identifies the IaC output language.
type Format string

const (
	FormatTerraform      Format = "terraform"
	FormatCloudFormation Format = "cloudformation"
)

// Cloud identifies the target cloud. Currently informational — drives prompt.
type Cloud string

const (
	CloudAWS   Cloud = "aws"
	CloudGCP   Cloud = "gcp"
	CloudAzure Cloud = "azure"
)

// GenerateRequest is the client-facing payload for IaC generation.
type GenerateRequest struct {
	Diagram  DiagramInput `json:"diagram"  validate:"required"`
	Mode     string       `json:"mode"     validate:"required,oneof=solution_architecture"`
	Cloud    Cloud        `json:"cloud"    validate:"required,oneof=aws gcp azure"`
	Format   Format       `json:"format"   validate:"required,oneof=terraform cloudformation"`
	Provider string       `json:"provider" validate:"required,oneof=managed_bedrock openai anthropic gemini"`
	APIKey   string       `json:"api_key,omitempty"`
	Params   RenderParams `json:"params,omitempty"`
	Explain  bool         `json:"explain,omitempty"`
}

type RenderParams struct {
	Region          string `json:"region,omitempty"`
	DBEngineVersion string `json:"db_engine_version,omitempty"`
	InstanceSize    string `json:"instance_size,omitempty"`
	NodeRuntime     string `json:"node_runtime,omitempty"`
}

// DiagramInput mirrors the shape the diagram service returns.
type DiagramInput struct {
	Nodes  []provider.Node  `json:"nodes"`
	Edges  []provider.Edge  `json:"edges"`
	Groups []provider.Group `json:"groups,omitempty"`
}

// GenerateResponse is returned to the client on success.
type GenerateResponse struct {
	Code        string   `json:"code"`
	Format      Format   `json:"format"`
	Language    string   `json:"language"` // for syntax highlighting hint
	Warnings    []string `json:"warnings,omitempty"`
	Explanation string   `json:"explanation,omitempty"`
}
