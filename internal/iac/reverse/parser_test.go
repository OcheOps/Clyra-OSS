package reverse

import (
	"strings"
	"testing"
)

func TestParse_SingleResource(t *testing.T) {
	src := `
resource "aws_s3_bucket" "logs" {
  bucket = "clyra-logs"
  tags = {
    Env = "prod"
  }
}`
	blocks, diags, err := Parse("test.tf", src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if diags.HasErrors() {
		t.Fatalf("unexpected diagnostics: %s", diags.Error())
	}
	if len(blocks) != 1 {
		t.Fatalf("want 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Kind != BlockResource || b.Type != "aws_s3_bucket" || b.Name != "logs" {
		t.Errorf("identity wrong: %+v", b)
	}
	if got := b.Attrs["bucket"]; !strings.Contains(got, `"clyra-logs"`) {
		t.Errorf("bucket attr = %q", got)
	}
	if _, ok := b.Attrs["tags"]; !ok {
		t.Errorf("expected tags attr captured")
	}
}

func TestParse_MultiResourceCrossReference(t *testing.T) {
	// The builder later needs to detect that the subnet references the vpc.
	// We just need the raw attribute text to be preserved here.
	src := `
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "public" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}`
	blocks, _, err := Parse("test.tf", src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("want 2 blocks, got %d", len(blocks))
	}
	subnet := blocks[1]
	if !strings.Contains(subnet.Attrs["vpc_id"], "aws_vpc.main.id") {
		t.Errorf("subnet vpc_id missing reference text: %q", subnet.Attrs["vpc_id"])
	}
}

func TestParse_NestedBlocks(t *testing.T) {
	// Security groups stash ingress/egress in nested blocks. The parser
	// must preserve the source slice of each so reference-scanning works.
	src := `
resource "aws_security_group" "web" {
  name = "web"

  ingress {
    from_port = 443
    to_port   = 443
    protocol  = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  ingress {
    from_port = 80
    to_port   = 80
    protocol  = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}`
	blocks, _, err := Parse("test.tf", src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("want 1 block, got %d", len(blocks))
	}
	ingress := blocks[0].Attrs["ingress"]
	if !strings.Contains(ingress, "aws_security_group.alb.id") {
		t.Errorf("merged ingress text missing the SG reference: %q", ingress)
	}
	if !strings.Contains(ingress, `"0.0.0.0/0"`) {
		t.Errorf("merged ingress text missing the public CIDR: %q", ingress)
	}
}

func TestParse_DataAndModule(t *testing.T) {
	src := `
data "aws_caller_identity" "current" {}

module "networking" {
  source = "./modules/networking"
  cidr   = "10.0.0.0/16"
}

resource "aws_s3_bucket" "app" {
  bucket = "app"
}`
	blocks, _, err := Parse("test.tf", src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(blocks) != 3 {
		t.Fatalf("want 3 blocks, got %d", len(blocks))
	}
	var sawData, sawModule, sawResource bool
	for _, b := range blocks {
		switch b.Kind {
		case BlockData:
			sawData = true
			if b.Type != "aws_caller_identity" || b.Name != "current" {
				t.Errorf("data identity: %+v", b)
			}
		case BlockModule:
			sawModule = true
			if b.Name != "networking" || b.Type != "" {
				t.Errorf("module identity: %+v", b)
			}
		case BlockResource:
			sawResource = true
		}
	}
	if !sawData || !sawModule || !sawResource {
		t.Errorf("missed a kind: data=%v module=%v resource=%v", sawData, sawModule, sawResource)
	}
}

func TestParse_IgnoresNonArchitectureBlocks(t *testing.T) {
	// terraform / provider / variable / output / locals must be silently
	// skipped — none of them describe architecture nodes.
	src := `
terraform {
  required_version = ">= 1.5"
}

provider "aws" {
  region = "us-east-1"
}

variable "env" {
  type    = string
  default = "prod"
}

locals {
  name_prefix = "clyra"
}

output "bucket_arn" {
  value = aws_s3_bucket.logs.arn
}

resource "aws_s3_bucket" "logs" {
  bucket = "clyra-logs"
}`
	blocks, diags, err := Parse("test.tf", src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if diags.HasErrors() {
		t.Fatalf("unexpected diagnostics: %s", diags.Error())
	}
	if len(blocks) != 1 {
		t.Fatalf("want 1 architecture block (just the s3 bucket), got %d", len(blocks))
	}
	if blocks[0].Type != "aws_s3_bucket" {
		t.Errorf("kept the wrong block: %+v", blocks[0])
	}
}

func TestParse_MalformedReturnsError(t *testing.T) {
	src := `resource "aws_s3_bucket" "logs" {`
	_, _, err := Parse("test.tf", src)
	if err == nil {
		t.Fatal("expected error on unterminated resource block")
	}
}
