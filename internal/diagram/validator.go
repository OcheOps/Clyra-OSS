package diagram

import (
	"fmt"

	"github.com/oche/diagram/internal/provider"
)

// Validate checks that the provider's output conforms to the diagram schema.
// Returns a ValidationError if any issues are found.
func Validate(resp *provider.GenerateResponse) *ValidationError {
	ve := &ValidationError{}

	if len(resp.Nodes) == 0 {
		ve.Add("diagram must contain at least one node")
	}

	nodeIDs := make(map[string]bool, len(resp.Nodes))
	groupIDs := make(map[string]bool, len(resp.Groups))

	// Validate groups
	for i, g := range resp.Groups {
		if g.ID == "" {
			ve.Add(fmt.Sprintf("group[%d]: missing id", i))
			continue
		}
		if groupIDs[g.ID] {
			ve.Add(fmt.Sprintf("group[%d]: duplicate id %q", i, g.ID))
		}
		groupIDs[g.ID] = true

		if g.Label == "" {
			ve.Add(fmt.Sprintf("group[%d] (%s): missing label", i, g.ID))
		}
	}

	// Validate nodes
	for i, n := range resp.Nodes {
		if n.ID == "" {
			ve.Add(fmt.Sprintf("node[%d]: missing id", i))
			continue
		}
		if nodeIDs[n.ID] {
			ve.Add(fmt.Sprintf("node[%d]: duplicate id %q", i, n.ID))
		}
		nodeIDs[n.ID] = true

		if n.Label == "" {
			ve.Add(fmt.Sprintf("node[%d] (%s): missing label", i, n.ID))
		}
		if n.Type == "" {
			ve.Add(fmt.Sprintf("node[%d] (%s): missing type", i, n.ID))
		}
		if n.Group != "" && !groupIDs[n.Group] {
			ve.Add(fmt.Sprintf("node[%d] (%s): references unknown group %q", i, n.ID, n.Group))
		}
	}

	// Validate edges
	edgeIDs := make(map[string]bool, len(resp.Edges))
	validStyles := map[string]bool{"solid": true, "dashed": true, "dotted": true, "": true}

	for i, e := range resp.Edges {
		if e.ID == "" {
			ve.Add(fmt.Sprintf("edge[%d]: missing id", i))
			continue
		}
		if edgeIDs[e.ID] {
			ve.Add(fmt.Sprintf("edge[%d]: duplicate id %q", i, e.ID))
		}
		edgeIDs[e.ID] = true

		if e.Source == "" {
			ve.Add(fmt.Sprintf("edge[%d] (%s): missing source", i, e.ID))
		} else if !nodeIDs[e.Source] {
			ve.Add(fmt.Sprintf("edge[%d] (%s): source %q references unknown node", i, e.ID, e.Source))
		}

		if e.Target == "" {
			ve.Add(fmt.Sprintf("edge[%d] (%s): missing target", i, e.ID))
		} else if !nodeIDs[e.Target] {
			ve.Add(fmt.Sprintf("edge[%d] (%s): target %q references unknown node", i, e.ID, e.Target))
		}

		if !validStyles[e.Style] {
			ve.Add(fmt.Sprintf("edge[%d] (%s): invalid style %q (must be solid, dashed, or dotted)", i, e.ID, e.Style))
		}
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}
