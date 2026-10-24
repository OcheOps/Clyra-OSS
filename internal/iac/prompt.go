package iac

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oche/diagram/internal/iac/graph"
)

const sharedRules = `You are generating STARTER / SCAFFOLD infrastructure code.
This is not a production deployment — it is a starting point that an engineer will review and extend.

Hard constraints:
- Maximum 25 resources total. If the diagram implies more, pick the essential ones.
- No complex IAM policies. Use managed policies or simple role references only.
- No advanced networking (no peering, transit gateways, private endpoints) unless the diagram shows them explicitly.
- No secrets, no credentials, no account IDs. Use placeholders like <YOUR_ACCOUNT_ID>.
- Prefer defaults. Minimize required inputs.
- Add concise comments above each resource explaining its role.
- Header comment at the top noting this is starter / scaffold code, not production.

Output format:
- Return ONLY the code. No prose outside comments. No markdown fences.`

// BuildTerraformPrompt constructs the system prompt for Terraform generation.
func BuildTerraformPrompt(cloud Cloud) string {
	var b strings.Builder
	b.WriteString("You are a senior cloud engineer generating Terraform (HCL2).\n\n")
	b.WriteString(sharedRules)
	b.WriteString("\n\nTerraform-specific rules:\n")
	b.WriteString("- Use the official provider for the target cloud.\n")
	b.WriteString("- Use terraform >= 1.5.\n")
	b.WriteString("- Declare a provider block with region as a variable.\n")
	b.WriteString("- Use variables for anything the user would want to tweak.\n")
	b.WriteString("- Use resource blocks, not modules.\n")

	switch cloud {
	case CloudAWS:
		b.WriteString("- Target provider: hashicorp/aws ~> 5.0. Default region us-east-1.\n")
	case CloudGCP:
		b.WriteString("- Target provider: hashicorp/google ~> 5.0. Require project_id variable.\n")
	case CloudAzure:
		b.WriteString("- Target provider: hashicorp/azurerm ~> 3.0.\n")
	}

	return b.String()
}

// BuildCloudFormationPrompt constructs the system prompt for CloudFormation generation.
func BuildCloudFormationPrompt(cloud Cloud) string {
	var b strings.Builder
	b.WriteString("You are a senior cloud engineer generating an AWS CloudFormation template (YAML).\n\n")
	b.WriteString(sharedRules)
	b.WriteString("\n\nCloudFormation-specific rules:\n")
	b.WriteString("- AWSTemplateFormatVersion: '2010-09-09'\n")
	b.WriteString("- Include Description at top.\n")
	b.WriteString("- Use Parameters for region and any tweakable values.\n")
	b.WriteString("- Use Resources with clear logical names.\n")
	b.WriteString("- Use Outputs for important resource references (URLs, ARNs).\n")
	b.WriteString("- CloudFormation is AWS-only regardless of the `cloud` input.\n")
	return b.String()
}

// BuildUserPrompt converts the diagram into a compact instruction for the LLM.
func BuildUserPrompt(diagram DiagramInput, format Format, cloud Cloud) (string, error) {
	payload, err := json.Marshal(diagram)
	if err != nil {
		return "", fmt.Errorf("marshaling diagram: %w", err)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Generate %s infrastructure code for the following architecture diagram.\n", format)
	fmt.Fprintf(&b, "Target cloud: %s.\n\n", cloud)
	b.WriteString("Diagram (JSON):\n")
	b.Write(payload)
	b.WriteString("\n\nEach node represents a cloud resource. Each edge represents a connection or dependency. ")
	b.WriteString("Groups represent VPCs, subnets, or logical containers.\n\n")
	b.WriteString("Map the node types to concrete resources. Preserve the intent of edges (e.g. a load balancer in front of compute, a database attached to compute).\n")
	return b.String(), nil
}

// BuildExplainPrompt asks the model to summarize generated IaC in plain English.
func BuildExplainPrompt(code string, format Format) string {
	var b strings.Builder
	b.WriteString("You are explaining infrastructure code to an engineer who wants a quick overview.\n")
	b.WriteString("Return 3-6 short bullet points covering: what this deploys, main resources, key connections, and anything the reader should review before applying.\n")
	b.WriteString("No fluff. No marketing. No markdown fences. Plain text bullets starting with '- '.\n\n")
	fmt.Fprintf(&b, "Code format: %s\n\n", format)
	b.WriteString("Code:\n")
	b.WriteString(code)
	return b.String()
}

func BuildStructuredGraphSystemPrompt() string {
	var b strings.Builder
	b.WriteString("You are a senior AWS platform engineer emitting a validated ResourceGraph JSON document.\n")
	b.WriteString("Target: starter/scaffold infrastructure for AWS only.\n")
	b.WriteString("Do not emit Terraform, CloudFormation, markdown, or prose.\n")
	b.WriteString("Return JSON only and follow the schema exactly.\n")
	b.WriteString("Use only supported resource kinds for v1: aws_vpc, aws_subnet, aws_security_group, aws_alb, aws_target_group, aws_listener, aws_ecs_cluster, aws_ecs_service, aws_ecs_task_definition, aws_rds_instance, aws_rds_subnet_group, aws_cloudfront_distribution.\n")
	b.WriteString("Choose safe starter defaults, wire dependencies correctly, and keep output under 25 resources.\n")
	return b.String()
}

func BuildStructuredGraphPrompt(diagram DiagramInput) (string, error) {
	payload, err := json.Marshal(diagram)
	if err != nil {
		return "", fmt.Errorf("marshaling diagram: %w", err)
	}

	var b strings.Builder
	b.WriteString("Emit a ResourceGraph JSON document for this AWS architecture.\n")
	b.WriteString(`Use:
- version = "1"
- cloud = "aws"
- region = "us-east-1"
- metadata.starter_only = true
- include a db_password string parameter if a database exists
- outputs for ALB DNS and DB endpoint when present

ResourceGraph shape:
{
  "version": "1",
  "cloud": "aws",
  "region": "us-east-1",
  "resources": [{"id":"","kind":"","props":{},"depends_on":[]}],
  "edges": [{"from":"","to":"","relation":"attaches_to|routes_to|references|security_group_ingress"}],
  "parameters": [{"name":"","type":"string|number|bool","description":""}],
  "outputs": [{"name":"","from":"","attr":"","description":""}],
  "metadata": {"stack_name":"","description":"","starter_only":true}
}
` + "\n")
	b.WriteString("Diagram JSON:\n")
	b.Write(payload)
	b.WriteString("\n")
	return b.String(), nil
}

func BuildStructuredGraphRepairPrompt(diagram DiagramInput, errs []string) (string, error) {
	base, err := BuildStructuredGraphPrompt(diagram)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(base)
	b.WriteString("\nThe previous graph was invalid. Fix these validation errors and return a full corrected JSON graph only:\n")
	for _, e := range errs {
		b.WriteString("- ")
		b.WriteString(e)
		b.WriteString("\n")
	}
	return b.String(), nil
}

func StructuredGraphSchema() string {
	schema := graph.ResourceGraph{
		Version: "1",
		Cloud:   graph.CloudAWS,
	}
	raw, _ := json.Marshal(schema)
	return string(raw)
}
