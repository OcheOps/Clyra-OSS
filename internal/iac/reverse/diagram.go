package reverse

import (
	"strings"

	"github.com/oche/diagram/internal/iac/graph"
	"github.com/oche/diagram/internal/provider"
)

// ToDiagram converts a BuildResult into the same provider.GenerateResponse
// shape Clyra's frontend already loads via loadDiagram(). This is the
// reverse pipeline's final stage — once the response leaves here it flows
// through the same Normalize → Salvage → Validate path AI-generated
// diagrams do, so identical post-processing applies (dangling edges
// dropped, types canonicalized, etc.).
//
// Mapping rules:
//
//   - Known kinds (`aws_vpc`, `aws_subnet`, …) become nodes with type
//     "aws.vpc", "aws.subnet", … (replace the first underscore with a
//     dot — matches the canonical type format the icon catalog uses).
//   - Unsupported kinds become "generic.service" placeholder nodes with
//     the original HCL type stashed in metadata["hcl_type"]. The UI can
//     surface this in the node tooltip so users see what was approximated.
//   - Modules dual-emit: a Group (the labeled box) PLUS a single interior
//     node typed "generic.container" so edges to `module.x` have a real
//     node to land on. The interior node belongs to the group; visually
//     you get a wrapping rectangle with one labeled child.
//   - Edges keep their from/to resource IDs and pick up a synthetic edge
//     ID + default "solid" style — the existing Normalize() pass would
//     supply that anyway, but emitting it here keeps responses self-
//     contained.
func ToDiagram(b *BuildResult) *provider.GenerateResponse {
	if b == nil || b.Graph == nil {
		return &provider.GenerateResponse{}
	}

	nodes := make([]provider.Node, 0, len(b.Graph.Resources)+len(b.Modules))
	groups := make([]provider.Group, 0, len(b.Modules))

	for _, r := range b.Graph.Resources {
		nodes = append(nodes, resourceToNode(r))
	}

	for _, m := range b.Modules {
		groupID := "g-module-" + m.Name
		groups = append(groups, provider.Group{
			ID:    groupID,
			Label: "module: " + m.Name,
			// "boundary" is the most neutral group kind in our schema —
			// modules aren't a VPC or a subnet, just a logical wrapper.
			Type: "boundary",
		})
		// Interior placeholder node so edges target something real and the
		// group has a child to wrap. Without this the box would be empty
		// and dagre layout has no anchor to position it against.
		nodes = append(nodes, provider.Node{
			ID:    "module." + m.Name,
			Label: m.Name,
			Type:  "generic.container",
			Group: groupID,
			Metadata: map[string]any{
				"module_source": m.Source,
			},
		})
	}

	edges := make([]provider.Edge, 0, len(b.Graph.Edges))
	for _, e := range b.Graph.Edges {
		edges = append(edges, provider.Edge{
			ID:     "e-" + e.From + "-" + e.To,
			Source: e.From,
			Target: e.To,
			Style:  "solid",
		})
	}

	return &provider.GenerateResponse{
		Nodes:  nodes,
		Edges:  edges,
		Groups: groups,
	}
}

// resourceToNode picks the canonical diagram type for a single resource and
// builds the node. Knowing happens via graph.Known() — anything not in the
// set is downgraded to generic.service with the original HCL type preserved
// in metadata for the UI to display.
func resourceToNode(r graph.Resource) provider.Node {
	known := graph.Known()
	rawKind := string(r.Kind)

	var nodeType string
	metadata := map[string]any{}

	if known[r.Kind] {
		nodeType = hclKindToCanonicalType(rawKind)
	} else {
		nodeType = "generic.service"
		metadata["hcl_type"] = rawKind
	}

	// Label is the suffix of the resource ID (after the "<kind-short>-"
	// prefix newID() applied). Falls back to the full ID if the prefix is
	// missing — defensive but shouldn't happen in practice.
	label := labelFromID(r.ID)

	if len(metadata) == 0 {
		metadata = nil
	}
	return provider.Node{
		ID:       r.ID,
		Label:    label,
		Type:     nodeType,
		Metadata: metadata,
	}
}

// hclKindToCanonicalType maps "aws_vpc" → "aws.vpc". Specifically it
// replaces the first underscore with a dot — everything after the first
// underscore is the component identifier with its own internal underscores
// preserved ("aws_api_gateway" → "aws.api_gateway"). Matches what's in
// internal/diagram/normalizer.go's knownTypes table.
func hclKindToCanonicalType(hclKind string) string {
	idx := strings.IndexByte(hclKind, '_')
	if idx < 0 {
		return hclKind
	}
	return hclKind[:idx] + "." + hclKind[idx+1:]
}

// labelFromID strips the kind-short prefix newID() applied so the visible
// node label is the user's original TF name. "vpc-main" → "main".
func labelFromID(id string) string {
	if idx := strings.IndexByte(id, '-'); idx > 0 && idx < len(id)-1 {
		return id[idx+1:]
	}
	return id
}
