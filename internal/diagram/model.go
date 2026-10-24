package diagram

import "github.com/oche/diagram/internal/provider"

// GenerateRequest is the client-facing request for diagram generation.
type GenerateRequest struct {
	Prompt   string                  `json:"prompt"   validate:"required,max=10000"`
	Provider string                  `json:"provider" validate:"required,oneof=managed_bedrock openai anthropic gemini"`
	APIKey   string                  `json:"api_key"`
	Mode     string                  `json:"mode"     validate:"omitempty,oneof=solution_architecture network_architecture enterprise_architecture"`
	Config   provider.ProviderConfig `json:"config,omitempty"`
}

// GenerateResponse is the validated, normalized diagram returned to the client.
type GenerateResponse struct {
	Nodes  []provider.Node  `json:"nodes"`
	Edges  []provider.Edge  `json:"edges"`
	Groups []provider.Group `json:"groups"`
}

// ValidationError collects all issues found during schema validation.
type ValidationError struct {
	Errors []string `json:"errors"`
}

func (v *ValidationError) Error() string {
	if len(v.Errors) == 0 {
		return "validation failed"
	}
	return v.Errors[0]
}

func (v *ValidationError) HasErrors() bool {
	return len(v.Errors) > 0
}

func (v *ValidationError) Add(msg string) {
	v.Errors = append(v.Errors, msg)
}
