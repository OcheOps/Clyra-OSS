package graph

const SchemaVersion = "1"

type Cloud string

const CloudAWS Cloud = "aws"

type ResourceGraph struct {
	Version    string      `json:"version"`
	Cloud      Cloud       `json:"cloud"`
	Region     string      `json:"region"`
	Resources  []Resource  `json:"resources"`
	Edges      []Edge      `json:"edges,omitempty"`
	Parameters []Parameter `json:"parameters,omitempty"`
	Outputs    []Output    `json:"outputs,omitempty"`
	Metadata   Metadata    `json:"metadata"`
}

type Resource struct {
	ID        string            `json:"id"`
	Kind      ResourceKind      `json:"kind"`
	Props     map[string]any    `json:"props,omitempty"`
	DependsOn []string          `json:"depends_on,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
}

type Edge struct {
	From     string   `json:"from"`
	To       string   `json:"to"`
	Relation Relation `json:"relation"`
}

type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Default     any    `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

type Output struct {
	Name        string `json:"name"`
	From        string `json:"from"`
	Attr        string `json:"attr"`
	Description string `json:"description,omitempty"`
}

type Metadata struct {
	StackName   string `json:"stack_name"`
	Description string `json:"description"`
	StarterOnly bool   `json:"starter_only"`
}

type Relation string

const (
	RelationAttachesTo           Relation = "attaches_to"
	RelationRoutesTo             Relation = "routes_to"
	RelationReferences           Relation = "references"
	RelationSecurityGroupIngress Relation = "security_group_ingress"
)
