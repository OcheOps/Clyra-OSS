package render

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/oche/diagram/internal/iac/graph"
	"github.com/oche/diagram/internal/iac/graph/props"
)

func RenderTerraform(g *graph.ResourceGraph) (string, error) {
	if err := graph.Validate(g); err != nil {
		return "", err
	}
	ordered, err := TopoSort(g)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(renderHeader(g))
	b.WriteString(renderVariables(g))

	for _, r := range ordered {
		snippet, err := renderResource(r)
		if err != nil {
			return "", fmt.Errorf("render %s (%s): %w", r.ID, r.Kind, err)
		}
		b.WriteString(snippet)
		if !strings.HasSuffix(snippet, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(renderOutputs(g))
	return b.String(), nil
}

func renderHeader(g *graph.ResourceGraph) string {
	return fmt.Sprintf(`# %s
# %s
# Starter / scaffold code - review before applying to production.

terraform {
  required_version = %q
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = %q
    }
  }
}

provider "aws" {
  region                      = var.region
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  skip_metadata_api_check     = true
  skip_region_validation      = true
}

`, g.Metadata.StackName, g.Metadata.Description, TerraformMinVersion, AWSProviderVersion)
}

func renderVariables(g *graph.ResourceGraph) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`variable "region" {
  description = "AWS region to deploy into."
  type        = string
  default     = %q
}

variable "db_password" {
  description = "Initial DB password (rotate after apply)."
  type        = string
  sensitive   = true
}

`, g.Region))

	params := append([]graph.Parameter(nil), g.Parameters...)
	sort.Slice(params, func(i, j int) bool { return params[i].Name < params[j].Name })
	for _, p := range params {
		if p.Name == "db_password" {
			continue
		}
		b.WriteString(fmt.Sprintf("variable %q {\n", p.Name))
		if p.Description != "" {
			b.WriteString(fmt.Sprintf("  description = %q\n", p.Description))
		}
		b.WriteString(fmt.Sprintf("  type        = %s\n", hclType(p.Type)))
		if p.Default != nil {
			b.WriteString(fmt.Sprintf("  default     = %s\n", hclValue(p.Default)))
		}
		b.WriteString("}\n\n")
	}
	return b.String()
}

func renderOutputs(g *graph.ResourceGraph) string {
	if len(g.Outputs) == 0 {
		return ""
	}
	var b strings.Builder
	for _, o := range g.Outputs {
		tfType := terraformResourceType(o.From, g)
		if tfType == "" {
			continue
		}
		b.WriteString(fmt.Sprintf(`output %q {
  description = %q
  value       = %s.%s.%s
}

`, o.Name, o.Description, tfType, o.From, terraformAttr(o.Attr)))
	}
	return b.String()
}

func renderResource(r graph.Resource) (string, error) {
	tmpl, ok := tfTemplates[r.Kind]
	if !ok {
		return "", fmt.Errorf("no terraform template for kind %s", r.Kind)
	}
	data, err := buildTemplateData(r)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func terraformResourceType(resourceID string, g *graph.ResourceGraph) string {
	for _, r := range g.Resources {
		if r.ID != resourceID {
			continue
		}
		switch r.Kind {
		case graph.KindALB:
			return "aws_lb"
		case graph.KindTargetGroup:
			return "aws_lb_target_group"
		case graph.KindListener:
			return "aws_lb_listener"
		case graph.KindRDSInstance:
			return "aws_db_instance"
		case graph.KindRDSSubnetGroup:
			return "aws_db_subnet_group"
		case graph.KindElastiCacheCluster:
			return "aws_elasticache_cluster"
		case graph.KindElastiCacheSubnetGroup:
			return "aws_elasticache_subnet_group"
		case graph.KindLambdaFunction:
			return "aws_lambda_function"
		case graph.KindLambdaPermission:
			return "aws_lambda_permission"
		case graph.KindAPIGatewayV2API:
			return "aws_apigatewayv2_api"
		case graph.KindAPIGatewayV2Route:
			return "aws_apigatewayv2_route"
		case graph.KindAPIGatewayV2Integration:
			return "aws_apigatewayv2_integration"
		case graph.KindDynamoDBTable:
			return "aws_dynamodb_table"
		case graph.KindS3Bucket:
			return "aws_s3_bucket"
		case graph.KindOpenSearchDomain:
			return "aws_opensearch_domain"
		case graph.KindEC2Instance:
			return "aws_instance"
		case graph.KindCloudFrontDist:
			return "aws_cloudfront_distribution"
		default:
			return string(r.Kind)
		}
	}
	return ""
}

func terraformAttr(attr string) string {
	switch attr {
	case "dns_name":
		return "dns_name"
	case "endpoint":
		return "address"
	case "api_endpoint":
		return "api_endpoint"
	default:
		return attr
	}
}

func hclType(t string) string {
	switch t {
	case "number":
		return "number"
	case "bool":
		return "bool"
	default:
		return "string"
	}
}

func hclValue(v any) string {
	switch x := v.(type) {
	case string:
		return fmt.Sprintf("%q", x)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		return fmt.Sprintf("%v", x)
	case int:
		return fmt.Sprintf("%d", x)
	case []any:
		parts := make([]string, len(x))
		for i, e := range x {
			parts[i] = hclValue(e)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return fmt.Sprintf("%q", fmt.Sprint(v))
	}
}

func buildTemplateData(r graph.Resource) (any, error) {
	switch r.Kind {
	case graph.KindVPC:
		p, err := props.AsVPC(r.Props)
		return tfVPCData{R: r, P: p}, err
	case graph.KindSubnet:
		p, err := props.AsSubnet(r.Props)
		return tfSubnetData{R: r, P: p}, err
	case graph.KindSecurityGroup:
		p, err := props.AsSecurityGroup(r.Props)
		return tfSecurityGroupData{R: r, P: p}, err
	case graph.KindALB:
		p, err := props.AsALB(r.Props)
		return tfALBData{R: r, P: p}, err
	case graph.KindTargetGroup:
		p, err := props.AsTargetGroup(r.Props)
		return tfTargetGroupData{R: r, P: p}, err
	case graph.KindListener:
		p, err := props.AsListener(r.Props)
		return tfListenerData{R: r, P: p}, err
	case graph.KindECSCluster:
		return tfSimpleData{R: r}, nil
	case graph.KindECSTaskDefinition:
		p, err := props.AsECSTaskDefinition(r.Props)
		return tfTaskData{R: r, P: p}, err
	case graph.KindECSService:
		p, err := props.AsECSService(r.Props)
		return tfECSServiceData{R: r, P: p}, err
	case graph.KindRDSSubnetGroup:
		p, err := props.AsRDSSubnetGroup(r.Props)
		return tfRDSSubnetGroupData{R: r, P: p}, err
	case graph.KindRDSInstance:
		p, err := props.AsRDSInstance(r.Props)
		return tfRDSData{R: r, P: p}, err
	case graph.KindElastiCacheSubnetGroup:
		p, err := props.AsElastiCacheSubnetGroup(r.Props)
		return tfElastiCacheSubnetGroupData{R: r, P: p}, err
	case graph.KindElastiCacheCluster:
		p, err := props.AsElastiCacheCluster(r.Props)
		return tfElastiCacheClusterData{R: r, P: p}, err
	case graph.KindLambdaFunction:
		p, err := props.AsLambdaFunction(r.Props)
		return tfLambdaFunctionData{R: r, P: p}, err
	case graph.KindLambdaPermission:
		p, err := props.AsLambdaPermission(r.Props)
		return tfLambdaPermissionData{R: r, P: p}, err
	case graph.KindAPIGatewayV2API:
		p, err := props.AsAPIGatewayV2API(r.Props)
		return tfAPIGatewayAPIData{R: r, P: p}, err
	case graph.KindAPIGatewayV2Integration:
		p, err := props.AsAPIGatewayV2Integration(r.Props)
		return tfAPIGatewayIntegrationData{R: r, P: p}, err
	case graph.KindAPIGatewayV2Route:
		p, err := props.AsAPIGatewayV2Route(r.Props)
		return tfAPIGatewayRouteData{R: r, P: p}, err
	case graph.KindDynamoDBTable:
		p, err := props.AsDynamoDBTable(r.Props)
		return tfDynamoDBData{R: r, P: p}, err
	case graph.KindS3Bucket:
		p, err := props.AsS3Bucket(r.Props)
		return tfS3BucketData{R: r, P: p}, err
	case graph.KindOpenSearchDomain:
		p, err := props.AsOpenSearchDomain(r.Props)
		return tfOpenSearchData{R: r, P: p}, err
	case graph.KindEC2Instance:
		p, err := props.AsEC2Instance(r.Props)
		return tfEC2Data{R: r, P: p}, err
	case graph.KindCloudWatchLogGroup:
		p, err := props.AsCloudWatchLogGroup(r.Props)
		return tfCloudWatchLogGroupData{R: r, P: p}, err
	case graph.KindCloudFrontDist:
		p, err := props.AsCloudFrontDist(r.Props)
		return tfCloudFrontData{R: r, P: p}, err
	default:
		return nil, fmt.Errorf("unhandled kind %s", r.Kind)
	}
}

type tfSimpleData struct{ R graph.Resource }
type tfVPCData struct {
	R graph.Resource
	P *props.VPC
}
type tfSubnetData struct {
	R graph.Resource
	P *props.Subnet
}
type tfSecurityGroupData struct {
	R graph.Resource
	P *props.SecurityGroup
}
type tfALBData struct {
	R graph.Resource
	P *props.ALB
}
type tfTargetGroupData struct {
	R graph.Resource
	P *props.TargetGroup
}
type tfListenerData struct {
	R graph.Resource
	P *props.Listener
}
type tfTaskData struct {
	R graph.Resource
	P *props.ECSTaskDefinition
}
type tfECSServiceData struct {
	R graph.Resource
	P *props.ECSService
}
type tfRDSSubnetGroupData struct {
	R graph.Resource
	P *props.RDSSubnetGroup
}
type tfElastiCacheSubnetGroupData struct {
	R graph.Resource
	P *props.ElastiCacheSubnetGroup
}
type tfElastiCacheClusterData struct {
	R graph.Resource
	P *props.ElastiCacheCluster
}
type tfLambdaFunctionData struct {
	R graph.Resource
	P *props.LambdaFunction
}
type tfLambdaPermissionData struct {
	R graph.Resource
	P *props.LambdaPermission
}
type tfAPIGatewayAPIData struct {
	R graph.Resource
	P *props.APIGatewayV2API
}
type tfAPIGatewayIntegrationData struct {
	R graph.Resource
	P *props.APIGatewayV2Integration
}
type tfAPIGatewayRouteData struct {
	R graph.Resource
	P *props.APIGatewayV2Route
}
type tfDynamoDBData struct {
	R graph.Resource
	P *props.DynamoDBTable
}
type tfS3BucketData struct {
	R graph.Resource
	P *props.S3Bucket
}
type tfOpenSearchData struct {
	R graph.Resource
	P *props.OpenSearchDomain
}
type tfEC2Data struct {
	R graph.Resource
	P *props.EC2Instance
}
type tfCloudWatchLogGroupData struct {
	R graph.Resource
	P *props.CloudWatchLogGroup
}
type tfRDSData struct {
	R graph.Resource
	P *props.RDSInstance
}
type tfCloudFrontData struct {
	R graph.Resource
	P *props.CloudFrontDistribution
}

var tfTemplates = map[graph.ResourceKind]*template.Template{}

func mustTemplate(name, body string) *template.Template {
	return template.Must(template.New(name).Funcs(template.FuncMap{
		"hcl":       hclValue,
		"tfSubnets": tfSubnetList,
		"tfSGs":     tfSecurityGroupList,
		"tfAliases": tfAliasList,
	}).Parse(body))
}

func tfSubnetList(ids []string) string {
	return tfRefList("aws_subnet", "id", ids)
}

func tfSecurityGroupList(ids []string) string {
	return tfRefList("aws_security_group", "id", ids)
}

func tfAliasList(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = hclValue(v)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func tfRefList(resourceType, attr string, ids []string) string {
	if len(ids) == 0 {
		return "[]"
	}
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprintf("%s.%s.%s", resourceType, id, attr)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
