package graph

import "encoding/json"

type CDKContext struct {
	Version   string     `json:"version"`
	Cloud     Cloud      `json:"cloud"`
	Region    string     `json:"region"`
	Resources []Resource `json:"resources"`
	Edges     []Edge     `json:"edges,omitempty"`
	Outputs   []Output   `json:"outputs,omitempty"`
	Metadata  Metadata   `json:"metadata"`
}

func ExportCDKContext(g *ResourceGraph) ([]byte, error) {
	ctx := CDKContext{
		Version:   g.Version,
		Cloud:     g.Cloud,
		Region:    g.Region,
		Resources: g.Resources,
		Edges:     g.Edges,
		Outputs:   g.Outputs,
		Metadata:  g.Metadata,
	}
	return json.MarshalIndent(ctx, "", "  ")
}
