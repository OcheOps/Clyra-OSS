package diagram

import (
	"testing"

	"github.com/oche/diagram/internal/provider"
)

func TestNormalizeType_DirectMatch(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"aws.ec2", "aws.ec2"},
		{"aws.lambda", "aws.lambda"},
		{"gcp.cloud_run", "gcp.cloud_run"},
		{"azure.aks", "azure.aks"},
		{"k8s.deployment", "k8s.deployment"},
		{"oci.compute", "oci.compute"},
		{"cloudflare.workers", "cloudflare.workers"},
		{"generic.database", "generic.database"},
		{"solution.microservice", "solution.microservice"},
		{"network.firewall", "network.firewall"},
		{"enterprise.erp", "enterprise.erp"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeType(tt.input)
			if got != tt.want {
				t.Errorf("normalizeType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeType_GeneratedCatalog(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"aws.bedrock_agents", "aws.bedrock_agents"},
		{"Cloud Deploy", "gcp.cloud_deploy"},
		{"Cognitive Search", "azure.ai_search"},
		{"Vertical Pod Autoscaler", "k8s.vpa"},
	}
	for _, tt := range tests {
		if got := normalizeType(tt.input); got != tt.want {
			t.Errorf("normalizeType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeType_CaseInsensitive(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"AWS.EC2", "aws.ec2"},
		{"AWS.Lambda", "aws.lambda"},
		{"GCP.Cloud_Run", "gcp.cloud_run"},
		{"K8S.Deployment", "k8s.deployment"},
		{"Solution.Microservice", "solution.microservice"},
		{"Network.Firewall", "network.firewall"},
		{"Enterprise.ERP", "enterprise.erp"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeType(tt.input)
			if got != tt.want {
				t.Errorf("normalizeType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeType_LegacyColonFormat(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// All legacy colon-separated types should map to new dot format
		{"aws:ec2", "aws.ec2"},
		{"aws:lambda", "aws.lambda"},
		{"aws:api-gateway", "aws.api_gateway"},
		{"aws:step-functions", "aws.step_functions"},
		{"aws:nat-gateway", "aws.nat_gateway"},
		{"gcp:cloud-run", "gcp.cloud_run"},
		{"gcp:pub-sub", "gcp.pubsub"},
		{"azure:app-service", "azure.app_service"},
		{"azure:cosmos-db", "azure.cosmos_db"},
		{"k8s:deployment", "k8s.deployment"},
		{"k8s:statefulset", "k8s.statefulset"},
		{"oci:object-storage", "oci.object_storage"},
		{"cloudflare:load-balancer", "cloudflare.load_balancer"},
		{"generic:load-balancer", "generic.load_balancer"},
		{"generic:mobile-app", "generic.mobile_app"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeType(tt.input)
			if got != tt.want {
				t.Errorf("normalizeType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeType_FuzzyAliases(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// AWS
		{"lambda", "aws.lambda"},
		{"s3", "aws.s3"},
		{"dynamodb", "aws.dynamodb"},
		{"api gateway", "aws.api_gateway"},
		{"igw", "aws.internet_gateway"},
		{"ssm", "aws.systems_manager"},

		// GCP
		{"cloud run", "gcp.cloud_run"},
		{"gcs", "gcp.cloud_storage"},
		{"bigquery", "gcp.bigquery"},
		{"pubsub", "gcp.pubsub"},
		{"pub/sub", "gcp.pubsub"},

		// Azure
		{"cosmos db", "azure.cosmos_db"},
		{"cosmosdb", "azure.cosmos_db"},
		{"azure ad", "azure.active_directory"},
		{"key vault", "azure.key_vault"},

		// K8s
		{"deployment", "k8s.deployment"},
		{"pod", "k8s.pod"},
		{"configmap", "k8s.configmap"},
		{"cronjob", "k8s.job"},

		// OCI
		{"vcn", "oci.vcn"},
		{"oke", "oci.oke"},

		// Cloudflare
		{"cloudflare", "cloudflare.cdn"},
		{"argo tunnel", "cloudflare.tunnel"},
		{"r2", "cloudflare.r2"},

		// Solution
		{"microservice", "solution.microservice"},
		{"backend", "solution.backend_service"},
		{"nosql", "solution.nosql_database"},
		{"vector store", "solution.vector_store"},
		{"ci/cd", "solution.ci_cd"},
		{"elasticsearch", "solution.search"},
		{"frontend", "solution.frontend"},

		// Network
		{"bastion", "network.bastion"},
		{"jump box", "network.bastion"},
		{"l7 load balancer", "network.load_balancer_l7"},
		{"transit gateway", "aws.transit_gateway"},
		{"zero trust", "network.zero_trust_proxy"},
		{"datacenter", "network.datacenter"},
		{"private subnet", "network.private_subnet"},

		// Enterprise
		{"erp", "enterprise.erp"},
		{"crm", "enterprise.crm"},
		{"sso", "enterprise.sso"},
		{"idp", "enterprise.identity_provider"},
		{"ldap", "enterprise.directory"},
		{"etl", "enterprise.etl"},
		{"data pipeline", "enterprise.etl"},
		{"saas", "enterprise.saas_app"},
		{"bi", "enterprise.bi_reporting"},

		// Generic
		{"server", "generic.server"},
		{"database", "generic.database"},
		{"db", "generic.database"},
		{"redis", "generic.cache"},
		{"load balancer", "generic.load_balancer"},
		{"docker", "generic.container"},
		{"browser", "generic.browser"},
		{"mobile app", "generic.mobile_app"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeType(tt.input)
			if got != tt.want {
				t.Errorf("normalizeType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeType_HyphenToSpaceFallback(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"cloud-run", "gcp.cloud_run"},
		{"cloud-functions", "gcp.cloud_functions"},
		{"cloud-storage", "gcp.cloud_storage"},
		{"app-service", "azure.app_service"},
		{"nat-gateway", "aws.nat_gateway"},
		{"load-balancer", "generic.load_balancer"},
		{"mobile-app", "generic.mobile_app"},
		{"on-prem", "generic.on_prem"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeType(tt.input)
			if got != tt.want {
				t.Errorf("normalizeType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeType_UnknownWithPrefix(t *testing.T) {
	// Unknown types with a dot or colon keep their prefix in canonical format
	got := normalizeType("custom:my-thing")
	if got != "custom.my_thing" {
		t.Errorf("expected custom.my_thing, got %q", got)
	}

	got = normalizeType("custom.my_thing")
	if got != "custom.my_thing" {
		t.Errorf("expected custom.my_thing, got %q", got)
	}
}

func TestNormalizeType_UnknownFallsBackToGeneric(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"something random", "generic.something_random"},
		{"FooBar", "generic.foobar"},
		{"my cool service", "generic.my_cool_service"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeType(tt.input)
			if got != tt.want {
				t.Errorf("normalizeType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalize_FullResponse(t *testing.T) {
	resp := &provider.GenerateResponse{
		Nodes: []provider.Node{
			{ID: "1", Label: "Web", Type: "EC2"},
			{ID: "2", Label: "DB", Type: "dynamodb"},
			{ID: "3", Label: "Cache", Type: "Redis"},
			{ID: "4", Label: "LB", Type: "aws:alb"},
			{ID: "5", Label: "Queue", Type: "SQS"},
			{ID: "6", Label: "K8s", Type: "deployment"},
			{ID: "7", Label: "ERP", Type: "erp"},
			{ID: "8", Label: "FW", Type: "network.firewall"},
		},
		Edges: []provider.Edge{
			{ID: "e1", Source: "1", Target: "2", Style: ""},
			{ID: "e2", Source: "1", Target: "3", Style: "dashed"},
		},
		Groups: nil,
	}

	Normalize(resp)

	expected := map[string]string{
		"1": "aws.ec2",
		"2": "aws.dynamodb",
		"3": "generic.cache",
		"4": "aws.alb",
		"5": "aws.sqs",
		"6": "k8s.deployment",
		"7": "enterprise.erp",
		"8": "network.firewall",
	}

	for _, n := range resp.Nodes {
		want, ok := expected[n.ID]
		if !ok {
			t.Errorf("unexpected node id %q", n.ID)
			continue
		}
		if n.Type != want {
			t.Errorf("node %q: type = %q, want %q", n.ID, n.Type, want)
		}
	}

	if resp.Edges[0].Style != "solid" {
		t.Errorf("empty edge style should default to solid, got %q", resp.Edges[0].Style)
	}
	if resp.Edges[1].Style != "dashed" {
		t.Errorf("explicit edge style should be preserved, got %q", resp.Edges[1].Style)
	}
	if resp.Groups == nil {
		t.Error("groups should not be nil after normalization")
	}
}

func TestNormalize_PreservesExistingGroups(t *testing.T) {
	resp := &provider.GenerateResponse{
		Nodes: []provider.Node{
			{ID: "1", Label: "Web", Type: "aws.ec2"},
		},
		Edges:  []provider.Edge{},
		Groups: []provider.Group{{ID: "g1", Label: "VPC", Type: "vpc"}},
	}

	Normalize(resp)

	if len(resp.Groups) != 1 || resp.Groups[0].ID != "g1" {
		t.Error("existing groups should be preserved")
	}
}
