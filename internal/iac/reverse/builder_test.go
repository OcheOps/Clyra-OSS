package reverse

import (
	"strings"
	"testing"

	"github.com/oche/diagram/internal/iac/graph"
)

// parseAndBuild is the test helper for both layers chained — we don't ship a
// public wrapper because callers want different error handling, but every
// test wants the convenience.
func parseAndBuild(t *testing.T, src string) *BuildResult {
	t.Helper()
	blocks, _, err := Parse("test.tf", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return Build(blocks)
}

func TestBuild_VPCSubnetRefInferred(t *testing.T) {
	src := `
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "public" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}`
	res := parseAndBuild(t, src)
	if len(res.Graph.Resources) != 2 {
		t.Fatalf("want 2 resources, got %d", len(res.Graph.Resources))
	}
	if len(res.Graph.Edges) != 1 {
		t.Fatalf("want 1 edge, got %d: %+v", len(res.Graph.Edges), res.Graph.Edges)
	}
	e := res.Graph.Edges[0]
	if e.From != "subnet-public" || e.To != "vpc-main" {
		t.Errorf("edge direction wrong: %+v", e)
	}
	if e.Relation != graph.RelationReferences {
		t.Errorf("edge relation = %q, want references", e.Relation)
	}
}

func TestBuild_MultipleRefsDeduped(t *testing.T) {
	// subnet_ids = [aws_subnet.a.id, aws_subnet.a.arn] is two regex matches,
	// one edge. Also tests that distinct targets still emit distinct edges.
	src := `
resource "aws_subnet" "a" {
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "b" {
  cidr_block = "10.0.2.0/24"
}

resource "aws_rds_subnet_group" "db" {
  subnet_ids = [aws_subnet.a.id, aws_subnet.a.arn, aws_subnet.b.id]
}`
	res := parseAndBuild(t, src)
	// Expect 2 edges: db → a, db → b. The duplicate ref to a (id + arn) collapses.
	if len(res.Graph.Edges) != 2 {
		t.Fatalf("want 2 dedup'd edges, got %d: %+v", len(res.Graph.Edges), res.Graph.Edges)
	}
	for _, e := range res.Graph.Edges {
		if e.From != "rds_subnet_group-db" {
			t.Errorf("unexpected From in edge: %+v", e)
		}
	}
}

func TestBuild_SecurityGroupIngressReferenceFromNestedBlock(t *testing.T) {
	// SG references inside ingress {} must be picked up — the parser preserves
	// the nested-block source text as a single Attrs["ingress"] value.
	src := `
resource "aws_security_group" "alb" {
  name = "alb"
}

resource "aws_security_group" "web" {
  name = "web"

  ingress {
    from_port       = 443
    to_port         = 443
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }
}`
	res := parseAndBuild(t, src)
	if len(res.Graph.Edges) != 1 {
		t.Fatalf("want 1 edge (web → alb via ingress), got %d", len(res.Graph.Edges))
	}
	e := res.Graph.Edges[0]
	if e.From != "security_group-web" || e.To != "security_group-alb" {
		t.Errorf("ingress reference not picked up correctly: %+v", e)
	}
}

func TestBuild_UnknownResourceKeptAsPlaceholder(t *testing.T) {
	src := `
resource "aws_s3_bucket" "app" {
  bucket = "app"
}

resource "aws_kms_key" "data" {
  description = "for app bucket"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "app" {
  bucket = aws_s3_bucket.app.id
  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.data.arn
      sse_algorithm     = "aws:kms"
    }
  }
}`
	res := parseAndBuild(t, src)

	// All 3 resources present even though kms_key + sse_config aren't known kinds.
	if len(res.Graph.Resources) != 3 {
		t.Fatalf("want 3 resources kept, got %d", len(res.Graph.Resources))
	}
	// Two unsupported kinds, sorted.
	want := []string{"aws_kms_key", "aws_s3_bucket_server_side_encryption_configuration"}
	if !equalStrSlice(res.UnsupportedKinds, want) {
		t.Errorf("UnsupportedKinds = %v, want %v", res.UnsupportedKinds, want)
	}
	// Edges still drawn between unknown and known resources.
	if len(res.Graph.Edges) < 2 {
		t.Errorf("want at least 2 edges connecting placeholder → known, got %d", len(res.Graph.Edges))
	}
	// Warnings mention both modules count (0 → not present) and unsupported list.
	foundUnsup := false
	for _, w := range res.Warnings {
		if strings.Contains(w, "unsupported") && strings.Contains(w, "aws_kms_key") {
			foundUnsup = true
		}
	}
	if !foundUnsup {
		t.Errorf("missing warning about unsupported kinds: %v", res.Warnings)
	}
}

func TestBuild_ModuleBlockBecomesGroupRef(t *testing.T) {
	src := `
module "networking" {
  source = "./modules/networking"
  cidr   = "10.0.0.0/16"
}

resource "aws_rds_instance" "main" {
  subnet_group_name = module.networking.private_subnet_group
}`
	res := parseAndBuild(t, src)
	if len(res.Modules) != 1 {
		t.Fatalf("want 1 module, got %d", len(res.Modules))
	}
	if res.Modules[0].Name != "networking" {
		t.Errorf("module name = %q", res.Modules[0].Name)
	}
	if res.Modules[0].Source != "./modules/networking" {
		t.Errorf("module source = %q (quotes not stripped?)", res.Modules[0].Source)
	}
	// Edge from rds → module
	if len(res.Graph.Edges) != 1 {
		t.Fatalf("want 1 edge (rds → module), got %d", len(res.Graph.Edges))
	}
	e := res.Graph.Edges[0]
	if e.From != "rds_instance-main" || e.To != "module.networking" {
		t.Errorf("module ref edge: %+v", e)
	}
	// Warning about modules.
	foundMod := false
	for _, w := range res.Warnings {
		if strings.Contains(w, "module") {
			foundMod = true
		}
	}
	if !foundMod {
		t.Errorf("missing module warning: %v", res.Warnings)
	}
}

func TestBuild_TenKindCoverage(t *testing.T) {
	// 10 of the 30 supported AWS kinds wired together in a realistic
	// three-tier shape. Validates that the same `Type == Kind` mapping
	// holds for the core set we'll demo with.
	src := `
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "public" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
}

resource "aws_security_group" "alb" {
  vpc_id = aws_vpc.main.id
}

resource "aws_alb" "web" {
  subnets         = [aws_subnet.public.id]
  security_groups = [aws_security_group.alb.id]
}

resource "aws_target_group" "web" {
  vpc_id = aws_vpc.main.id
}

resource "aws_listener" "web" {
  load_balancer_arn = aws_alb.web.arn
  default_action {
    target_group_arn = aws_target_group.web.arn
  }
}

resource "aws_ecs_cluster" "main" {}

resource "aws_ecs_task_definition" "app" {
  family = "app"
}

resource "aws_ecs_service" "app" {
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  load_balancers {
    target_group_arn = aws_target_group.web.arn
  }
}`
	res := parseAndBuild(t, src)
	if len(res.Graph.Resources) != 10 {
		t.Fatalf("want 10 resources, got %d", len(res.Graph.Resources))
	}
	if len(res.UnsupportedKinds) != 0 {
		t.Errorf("unexpected unsupported kinds: %v", res.UnsupportedKinds)
	}
	// All ten kinds should be Known().
	known := graph.Known()
	for _, r := range res.Graph.Resources {
		if !known[r.Kind] {
			t.Errorf("resource %s has unrecognized kind %q", r.ID, r.Kind)
		}
	}
	// Spot-check a few expected edges.
	mustEdge := func(from, to string) {
		for _, e := range res.Graph.Edges {
			if e.From == from && e.To == to {
				return
			}
		}
		t.Errorf("missing expected edge %s → %s; got %+v", from, to, res.Graph.Edges)
	}
	mustEdge("subnet-public", "vpc-main")
	mustEdge("internet_gateway-main", "vpc-main")
	mustEdge("ecs_service-app", "ecs_cluster-main")
	mustEdge("ecs_service-app", "target_group-web")
	mustEdge("listener-web", "alb-web")
}

func TestBuild_NoSelfReferenceEdges(t *testing.T) {
	// `tags = { Owner = aws_iam_role.this.arn }` inside aws_iam_role.this
	// must not produce a self-edge.
	src := `
resource "aws_iam_role" "this" {
  name = "this"
  tags = {
    Owner = aws_iam_role.this.arn
  }
}`
	res := parseAndBuild(t, src)
	for _, e := range res.Graph.Edges {
		if e.From == e.To {
			t.Errorf("self-edge emitted: %+v", e)
		}
	}
}

func TestBuild_DataBlockNotResolved(t *testing.T) {
	// Refs to a data block must not create edges — the architectural
	// dependency is on the consumer, not the lookup.
	src := `
data "aws_ami" "base" {
  most_recent = true
}

resource "aws_ec2_instance" "web" {
  ami = data.aws_ami.base.id
}`
	res := parseAndBuild(t, src)
	if len(res.Graph.Resources) != 1 {
		t.Fatalf("want only the ec2 resource, got %d", len(res.Graph.Resources))
	}
	if len(res.Graph.Edges) != 0 {
		t.Errorf("want no edges from data refs, got %+v", res.Graph.Edges)
	}
}

func equalStrSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
