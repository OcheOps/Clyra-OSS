package reverse

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/oche/diagram/internal/iac/graph"
)

// TestE2E_RealisticThreeTier exercises the full Parse → Build → ToDiagram
// pipeline against a hand-written-but-realistic three-tier web app TF
// definition. Not a unit test — a "show me what comes out for real input"
// driver. Run with `-v` to see the rendered diagram and warnings dumped.
//
// The fixture is shaped like what you'd get from a typical mid-size team:
//   - VPC with public + private subnets
//   - ALB → ECS Fargate cluster with task def
//   - RDS Postgres with subnet group + SG
//   - ElastiCache Redis
//   - S3 bucket for uploads
//   - CloudFront distribution
//   - One module reference (networking)
//   - Couple of unsupported types (KMS, S3 SSE config)
func TestE2E_RealisticThreeTier(t *testing.T) {
	src := `
# Networking sourced from a shared module — modules stay opaque in MVP.
module "networking" {
  source = "git::https://github.com/example/tf-modules.git//networking?ref=v2.1"
  cidr   = "10.0.0.0/16"
}

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  tags = { Name = "clyra-prod" }
}

resource "aws_subnet" "public_a" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-1a"
}

resource "aws_subnet" "public_b" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-east-1b"
}

resource "aws_subnet" "private_a" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.11.0/24"
  availability_zone = "us-east-1a"
}

resource "aws_subnet" "private_b" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.12.0/24"
  availability_zone = "us-east-1b"
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
}

resource "aws_nat_gateway" "main" {
  subnet_id = aws_subnet.public_a.id
}

# Load balancer + SGs
resource "aws_security_group" "alb" {
  vpc_id = aws_vpc.main.id
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "app" {
  vpc_id = aws_vpc.main.id
  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }
}

resource "aws_security_group" "db" {
  vpc_id = aws_vpc.main.id
  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.app.id]
  }
}

resource "aws_alb" "web" {
  subnets         = [aws_subnet.public_a.id, aws_subnet.public_b.id]
  security_groups = [aws_security_group.alb.id]
}

resource "aws_target_group" "web" {
  vpc_id      = aws_vpc.main.id
  port        = 8080
  protocol    = "HTTP"
  target_type = "ip"
}

resource "aws_listener" "https" {
  load_balancer_arn = aws_alb.web.arn
  port              = 443
  protocol          = "HTTPS"
  default_action {
    target_group_arn = aws_target_group.web.arn
    type             = "forward"
  }
}

# ECS Fargate
resource "aws_ecs_cluster" "main" {
  name = "clyra-prod"
}

resource "aws_ecs_task_definition" "api" {
  family                   = "api"
  cpu                      = "512"
  memory                   = "1024"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
}

resource "aws_ecs_service" "api" {
  name            = "api"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = 2
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = [aws_subnet.private_a.id, aws_subnet.private_b.id]
    security_groups = [aws_security_group.app.id]
  }

  load_balancers {
    target_group_arn = aws_target_group.web.arn
    container_name   = "api"
    container_port   = 8080
  }
}

# Database
resource "aws_rds_subnet_group" "main" {
  name       = "main"
  subnet_ids = [aws_subnet.private_a.id, aws_subnet.private_b.id]
}

resource "aws_rds_instance" "main" {
  identifier             = "clyra-prod"
  engine                 = "postgres"
  engine_version         = "15.4"
  instance_class         = "db.t4g.small"
  allocated_storage      = 50
  db_subnet_group_name   = aws_rds_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.db.id]
  multi_az               = true
}

# Cache
resource "aws_elasticache_subnet_group" "main" {
  name       = "main"
  subnet_ids = [aws_subnet.private_a.id, aws_subnet.private_b.id]
}

resource "aws_elasticache_cluster" "main" {
  cluster_id           = "clyra-prod"
  engine               = "redis"
  node_type            = "cache.t4g.micro"
  num_cache_nodes      = 1
  subnet_group_name    = aws_elasticache_subnet_group.main.name
  security_group_ids   = [aws_security_group.db.id]
}

# Object storage + CDN
resource "aws_s3_bucket" "uploads" {
  bucket = "clyra-prod-uploads"
}

# Unsupported in MVP — KMS doesn't have a Kind yet
resource "aws_kms_key" "uploads" {
  description = "Encrypt user uploads"
}

# Also unsupported
resource "aws_s3_bucket_server_side_encryption_configuration" "uploads" {
  bucket = aws_s3_bucket.uploads.id
  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.uploads.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

resource "aws_cloudfront_distribution" "uploads" {
  enabled = true
  origin {
    domain_name = aws_s3_bucket.uploads.bucket_regional_domain_name
    origin_id   = "uploads"
  }
}`

	blocks, diags, err := Parse("real.tf", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if diags.HasErrors() {
		t.Logf("parse diagnostics: %s", diags.Error())
	}
	t.Logf("parsed %d top-level blocks", len(blocks))

	br := Build(blocks)
	t.Logf("built graph: %d resources, %d edges, %d modules, %d unsupported kinds",
		len(br.Graph.Resources), len(br.Graph.Edges), len(br.Modules), len(br.UnsupportedKinds))

	// Dump resource set so we can eyeball coverage.
	knownCount := 0
	unknownCount := 0
	known := graph.Known()
	for _, r := range br.Graph.Resources {
		if known[r.Kind] {
			knownCount++
		} else {
			unknownCount++
		}
	}
	t.Logf("kind split: %d known, %d unsupported (placeholders)", knownCount, unknownCount)
	t.Logf("unsupported HCL types: %v", br.UnsupportedKinds)
	t.Logf("warnings:")
	for _, w := range br.Warnings {
		t.Logf("  · %s", w)
	}

	// Spot-check the critical edges that prove edge inference works on a
	// realistic shape, not just toy fixtures.
	type expectedEdge struct {
		fromContains string
		toContains   string
		desc         string
	}
	expected := []expectedEdge{
		{"subnet-public_a", "vpc-main", "public_a → vpc"},
		{"subnet-private_a", "vpc-main", "private_a → vpc"},
		{"security_group-app", "security_group-alb", "app SG ingress refs ALB SG"},
		{"security_group-db", "security_group-app", "db SG ingress refs app SG"},
		{"alb-web", "subnet-public_a", "ALB attached to public_a"},
		{"alb-web", "security_group-alb", "ALB → its SG"},
		{"listener-https", "alb-web", "listener → ALB"},
		{"listener-https", "target_group-web", "listener default_action → target group"},
		{"ecs_service-api", "ecs_cluster-main", "service → cluster"},
		{"ecs_service-api", "ecs_task_definition-api", "service → task def"},
		{"ecs_service-api", "target_group-web", "service load_balancers → target group"},
		{"ecs_service-api", "subnet-private_a", "service network_config → private subnet"},
		{"ecs_service-api", "security_group-app", "service → app SG"},
		{"rds_subnet_group-main", "subnet-private_a", "rds subnet group → private subnet"},
		{"rds_instance-main", "rds_subnet_group-main", "rds → its subnet group"},
		{"rds_instance-main", "security_group-db", "rds → db SG"},
		{"elasticache_subnet_group-main", "subnet-private_a", "elasticache subnet group → private subnet"},
		{"elasticache_cluster-main", "elasticache_subnet_group-main", "elasticache → subnet group"},
		{"cloudfront_distribution-uploads", "s3_bucket-uploads", "cloudfront origin → s3"},
	}

	edgeSet := map[string]bool{}
	for _, e := range br.Graph.Edges {
		edgeSet[e.From+"|"+e.To] = true
	}
	missing := 0
	for _, ee := range expected {
		if !edgeSet[ee.fromContains+"|"+ee.toContains] {
			t.Errorf("missing edge: %s (expected %s → %s)", ee.desc, ee.fromContains, ee.toContains)
			missing++
		}
	}
	if missing == 0 {
		t.Logf("all %d critical edges present", len(expected))
	} else {
		t.Logf("%d / %d critical edges missing — full edge list follows:", missing, len(expected))
		// Print every edge so the failure is debuggable in CI logs.
		for _, e := range br.Graph.Edges {
			t.Logf("  %s → %s", e.From, e.To)
		}
	}

	// Run the diagram converter and dump a node-type summary.
	d := ToDiagram(br)
	typeCounts := map[string]int{}
	for _, n := range d.Nodes {
		typeCounts[n.Type]++
	}
	t.Logf("diagram: %d nodes, %d edges, %d groups", len(d.Nodes), len(d.Edges), len(d.Groups))
	t.Logf("node types:")
	for tp, n := range typeCounts {
		t.Logf("  %-30s %d", tp, n)
	}

	// Compact JSON for spot-inspection. Pretty-print just the first 1KB so
	// the test log isn't a wall of JSON.
	buf, _ := json.MarshalIndent(d, "", "  ")
	preview := string(buf)
	if len(preview) > 1500 {
		preview = preview[:1500] + "\n... (truncated)"
	}
	t.Logf("diagram JSON preview:\n%s", preview)

	// Hard assertions on totals we already counted.
	if len(br.Graph.Resources) < 20 {
		t.Errorf("expected at least 20 resources, got %d", len(br.Graph.Resources))
	}
	if len(br.Modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(br.Modules))
	}
	// 2 unsupported kinds: aws_kms_key and aws_s3_bucket_server_side_encryption_configuration.
	if len(br.UnsupportedKinds) != 2 {
		t.Errorf("expected 2 unsupported kinds, got %d: %v",
			len(br.UnsupportedKinds), br.UnsupportedKinds)
	}
	if !strings.Contains(strings.Join(br.UnsupportedKinds, ","), "aws_kms_key") {
		t.Errorf("kms_key not flagged as unsupported")
	}

	// One module group + module-interior node should be present.
	if len(d.Groups) != 1 {
		t.Errorf("expected 1 group (module networking), got %d", len(d.Groups))
	}

	// Print a final summary for the human reader.
	fmt.Println(strings.Repeat("=", 64))
	fmt.Printf("E2E summary: %d resources / %d edges / %d modules / %d unsupported\n",
		len(br.Graph.Resources), len(br.Graph.Edges), len(br.Modules), len(br.UnsupportedKinds))
	fmt.Printf("Diagram nodes by type:\n")
	for tp, n := range typeCounts {
		fmt.Printf("  %-40s %d\n", tp, n)
	}
	fmt.Println(strings.Repeat("=", 64))
}
