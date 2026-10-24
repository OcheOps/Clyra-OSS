package reverse

import (
	"testing"

	"github.com/oche/diagram/internal/iac/graph"
	"github.com/oche/diagram/internal/iac/render"
)

// TestRoundTrip_VPCSubnet verifies the core promise of the spec: a
// diagram → forward render → reverse parse cycle preserves the logical
// structure. This is the "render to TF, paste TF back, get the same
// graph" guarantee that proves the two layers are aligned.
//
// We deliberately don't compare props deeply — TF emits more detail than
// the parser captures (defaults, computed fields), and that's expected.
// What must round-trip is the resource set and the edges between them.
func TestRoundTrip_VPCSubnet(t *testing.T) {
	// IDs and props match the forward renderer's validation rules
	// (snake_case IDs, required-fields schema). The reverse parser doesn't
	// care about either — it accepts whatever the source says.
	original := &graph.ResourceGraph{
		Version: graph.SchemaVersion,
		Cloud:   graph.CloudAWS,
		Region:  "us-east-1",
		Resources: []graph.Resource{
			{
				ID:   "vpc_main",
				Kind: graph.KindVPC,
				Props: map[string]any{
					"cidr_block": "10.0.0.0/16",
				},
			},
			{
				ID:   "subnet_public",
				Kind: graph.KindSubnet,
				Props: map[string]any{
					"vpc":               "vpc_main",
					"cidr_block":        "10.0.1.0/24",
					"availability_zone": "us-east-1a",
				},
				DependsOn: []string{"vpc_main"},
			},
		},
		Edges: []graph.Edge{
			{From: "subnet_public", To: "vpc_main", Relation: graph.RelationReferences},
		},
		Metadata: graph.Metadata{StackName: "rt"},
	}

	tf, err := render.RenderTerraform(original)
	if err != nil {
		t.Fatalf("forward render: %v", err)
	}
	if tf == "" {
		t.Fatal("forward render produced empty output")
	}

	blocks, diags, err := Parse("rt.tf", tf)
	if err != nil {
		t.Fatalf("reverse parse: %v\n--- tf ---\n%s", err, tf)
	}
	if diags.HasErrors() {
		t.Fatalf("parse diagnostics: %s\n--- tf ---\n%s", diags.Error(), tf)
	}
	result := Build(blocks)

	// We must recover at least the vpc + subnet — the renderer may add
	// supporting resources (route tables, IAM roles, etc.) but our two
	// originals must survive.
	gotKinds := map[graph.ResourceKind]int{}
	for _, r := range result.Graph.Resources {
		gotKinds[r.Kind]++
	}
	if gotKinds[graph.KindVPC] < 1 {
		t.Errorf("vpc lost in round trip; got resources: %+v", result.Graph.Resources)
	}
	if gotKinds[graph.KindSubnet] < 1 {
		t.Errorf("subnet lost in round trip; got resources: %+v", result.Graph.Resources)
	}

	// At least one edge from the subnet must point to the vpc — that's
	// the dependency the renderer encodes via vpc_id = aws_vpc.X.id.
	var foundSubnetToVPC bool
	for _, e := range result.Graph.Edges {
		fromIsSubnet := matchKind(result.Graph.Resources, e.From, graph.KindSubnet)
		toIsVPC := matchKind(result.Graph.Resources, e.To, graph.KindVPC)
		if fromIsSubnet && toIsVPC {
			foundSubnetToVPC = true
			break
		}
	}
	if !foundSubnetToVPC {
		t.Errorf("edge subnet→vpc lost in round trip; got edges: %+v", result.Graph.Edges)
	}
}

func matchKind(resources []graph.Resource, id string, want graph.ResourceKind) bool {
	for _, r := range resources {
		if r.ID == id && r.Kind == want {
			return true
		}
	}
	return false
}
