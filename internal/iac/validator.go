package iac

import "errors"

const (
	// MaxDiagramNodes caps the diagram size passed to IaC generation.
	// Matches product constraint: starter scaffolds, not full deployments.
	MaxDiagramNodes = 25
	// MinDiagramNodes ensures we have something to scaffold.
	MinDiagramNodes = 1
)

var (
	ErrModeNotSupported  = errors.New("IaC generation is only available in solution_architecture mode")
	ErrDiagramEmpty      = errors.New("diagram has no nodes")
	ErrDiagramTooLarge   = errors.New("diagram exceeds 25-node limit for IaC generation")
	ErrFormatUnsupported = errors.New("unsupported format — use terraform or cloudformation")
	ErrCloudUnsupported  = errors.New("unsupported cloud — use aws, gcp, or azure")
	ErrMissingProvider   = errors.New("provider is required")
	ErrMissingAPIKey     = errors.New("api_key is required for BYO providers")
	ErrFreePlanIaCUsed   = errors.New("free plan includes one IaC export — upgrade to Pro or BYO for unlimited")
)

// ValidateInput runs request-level checks before any provider call.
func ValidateInput(req GenerateRequest) error {
	if req.Mode != "solution_architecture" {
		return ErrModeNotSupported
	}

	n := len(req.Diagram.Nodes)
	if n < MinDiagramNodes {
		return ErrDiagramEmpty
	}
	if n > MaxDiagramNodes {
		return ErrDiagramTooLarge
	}

	switch req.Format {
	case FormatTerraform, FormatCloudFormation:
	default:
		return ErrFormatUnsupported
	}

	switch req.Cloud {
	case CloudAWS, CloudGCP, CloudAzure:
	default:
		return ErrCloudUnsupported
	}

	// CloudFormation is AWS-only — reject mismatch up front.
	if req.Format == FormatCloudFormation && req.Cloud != CloudAWS {
		return ErrCloudUnsupported
	}

	if req.Provider == "" {
		return ErrMissingProvider
	}
	return nil
}

// LanguageHint returns the syntax-highlighting language for a given format.
func LanguageHint(f Format) string {
	switch f {
	case FormatTerraform:
		return "hcl"
	case FormatCloudFormation:
		return "yaml"
	default:
		return "text"
	}
}
