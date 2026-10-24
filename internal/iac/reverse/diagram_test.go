package reverse

import (
	"testing"
)

func TestToDiagram_KnownKindMapping(t *testing.T) {
	src := `
resource "aws_vpc" "main" { cidr_block = "10.0.0.0/16" }
resource "aws_api_gateway_rest_api" "v1" { name = "v1" }
resource "aws_s3_bucket" "logs" { bucket = "logs" }`
	// Note: aws_api_gateway_rest_api isn't in graph.Known() — that's
	// intentional, it'll show how the converter handles a near-miss.
	res := parseAndBuild(t, src)
	d := ToDiagram(res)

	// 3 nodes total.
	if len(d.Nodes) != 3 {
		t.Fatalf("want 3 nodes, got %d", len(d.Nodes))
	}

	byID := map[string]string{}
	for _, n := range d.Nodes {
		byID[n.ID] = n.Type
	}

	if byID["vpc-main"] != "aws.vpc" {
		t.Errorf("vpc type = %q, want aws.vpc", byID["vpc-main"])
	}
	if byID["s3_bucket-logs"] != "aws.s3_bucket" {
		t.Errorf("s3 bucket type = %q, want aws.s3_bucket", byID["s3_bucket-logs"])
	}
	if byID["api_gateway_rest_api-v1"] != "generic.service" {
		t.Errorf("unsupported resource should fall back to generic.service, got %q", byID["api_gateway_rest_api-v1"])
	}
	// And metadata on the fallback preserves the HCL type.
	for _, n := range d.Nodes {
		if n.ID == "api_gateway_rest_api-v1" {
			if n.Metadata["hcl_type"] != "aws_api_gateway_rest_api" {
				t.Errorf("fallback node lost original HCL type: %v", n.Metadata)
			}
		}
	}
}

func TestToDiagram_LabelStripsKindPrefix(t *testing.T) {
	src := `resource "aws_vpc" "production" { cidr_block = "10.0.0.0/16" }`
	d := ToDiagram(parseAndBuild(t, src))
	if d.Nodes[0].Label != "production" {
		t.Errorf("label = %q, want 'production'", d.Nodes[0].Label)
	}
}

func TestToDiagram_EdgeConversion(t *testing.T) {
	src := `
resource "aws_vpc" "main" { cidr_block = "10.0.0.0/16" }
resource "aws_subnet" "a" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}`
	d := ToDiagram(parseAndBuild(t, src))
	if len(d.Edges) != 1 {
		t.Fatalf("want 1 edge, got %d", len(d.Edges))
	}
	e := d.Edges[0]
	if e.Source != "subnet-a" || e.Target != "vpc-main" {
		t.Errorf("edge wrong: %+v", e)
	}
	if e.ID == "" {
		t.Errorf("synthetic edge ID missing")
	}
	if e.Style != "solid" {
		t.Errorf("edge style = %q, want solid", e.Style)
	}
}

func TestToDiagram_ModuleDualEmits(t *testing.T) {
	// One module → one group + one interior placeholder node + edges to it.
	src := `
module "networking" {
  source = "./modules/networking"
}

resource "aws_rds_instance" "main" {
  subnet_group_name = module.networking.private_subnet_group
}`
	d := ToDiagram(parseAndBuild(t, src))

	if len(d.Groups) != 1 {
		t.Fatalf("want 1 group, got %d", len(d.Groups))
	}
	g := d.Groups[0]
	if g.ID != "g-module-networking" || g.Label != "module: networking" {
		t.Errorf("group wrong: %+v", g)
	}
	if g.Type != "boundary" {
		t.Errorf("group type = %q, want boundary", g.Type)
	}

	// Two nodes total: the rds + the module-interior placeholder.
	if len(d.Nodes) != 2 {
		t.Fatalf("want 2 nodes, got %d", len(d.Nodes))
	}
	var moduleNode *struct {
		Group string
		Type  string
	}
	for _, n := range d.Nodes {
		if n.ID == "module.networking" {
			moduleNode = &struct {
				Group string
				Type  string
			}{Group: n.Group, Type: n.Type}
		}
	}
	if moduleNode == nil {
		t.Fatal("interior module node missing")
	}
	if moduleNode.Group != "g-module-networking" {
		t.Errorf("module node not assigned to its group: %+v", moduleNode)
	}
	if moduleNode.Type != "generic.container" {
		t.Errorf("module node type = %q, want generic.container", moduleNode.Type)
	}

	// Edge from rds to module
	if len(d.Edges) != 1 {
		t.Fatalf("want 1 edge, got %d", len(d.Edges))
	}
	if d.Edges[0].Target != "module.networking" {
		t.Errorf("module edge target wrong: %+v", d.Edges[0])
	}
}

func TestToDiagram_NilSafe(t *testing.T) {
	if d := ToDiagram(nil); d == nil {
		t.Fatal("nil input should produce empty response, not nil")
	}
	if d := ToDiagram(&BuildResult{}); d == nil {
		t.Fatal("empty build result should produce empty response")
	}
}

func TestHclKindToCanonical(t *testing.T) {
	cases := map[string]string{
		"aws_vpc":                 "aws.vpc",
		"aws_api_gateway":         "aws.api_gateway",
		"aws_apigatewayv2_route":  "aws.apigatewayv2_route",
		"aws":                     "aws", // no underscore — pass-through
	}
	for in, want := range cases {
		if got := hclKindToCanonicalType(in); got != want {
			t.Errorf("%q → %q, want %q", in, got, want)
		}
	}
}
