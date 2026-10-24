package diagram

import (
	"strings"

	"github.com/oche/diagram/internal/provider"
)

// Canonical type format: <family>.<component>
// All lowercase, underscore-separated words, dot-separated family.component.
//
// Families: aws, azure, gcp, k8s, oci, cloudflare, solution, network, enterprise, generic

// knownTypes contains the core cross-mode types. Provider catalog types and
// aliases are generated from components/*.md in catalog_generated.go.
var knownTypes = map[string]string{
	// ── AWS ──────────────────────────────────────────────────────────
	"aws.vpc":              "aws.vpc",
	"aws.subnet":           "aws.subnet",
	"aws.ec2":              "aws.ec2",
	"aws.ecs":              "aws.ecs",
	"aws.eks":              "aws.eks",
	"aws.lambda":           "aws.lambda",
	"aws.alb":              "aws.alb",
	"aws.nlb":              "aws.nlb",
	"aws.elb":              "aws.elb",
	"aws.api_gateway":      "aws.api_gateway",
	"aws.cloudfront":       "aws.cloudfront",
	"aws.s3":               "aws.s3",
	"aws.rds":              "aws.rds",
	"aws.dynamodb":         "aws.dynamodb",
	"aws.elasticache":      "aws.elasticache",
	"aws.route_53":         "aws.route_53",
	"aws.iam":              "aws.iam",
	"aws.waf":              "aws.waf",
	"aws.nat_gateway":      "aws.nat_gateway",
	"aws.internet_gateway": "aws.internet_gateway",
	"aws.sqs":              "aws.sqs",
	"aws.sns":              "aws.sns",
	"aws.eventbridge":      "aws.eventbridge",
	"aws.step_functions":   "aws.step_functions",
	"aws.cloudwatch":       "aws.cloudwatch",
	"aws.secrets_manager":  "aws.secrets_manager",
	"aws.systems_manager":  "aws.systems_manager",
	"aws.ecr":              "aws.ecr",
	"aws.cognito":          "aws.cognito",
	"aws.kinesis":          "aws.kinesis",
	"aws.transit_gateway":  "aws.transit_gateway",
	"aws.direct_connect":   "aws.direct_connect",
	"aws.config":           "aws.config",
	"aws.guardduty":        "aws.guardduty",

	// ── Azure ────────────────────────────────────────────────────────
	"azure.vnet":                "azure.vnet",
	"azure.subnet":              "azure.subnet",
	"azure.vm":                  "azure.vm",
	"azure.aks":                 "azure.aks",
	"azure.app_service":         "azure.app_service",
	"azure.functions":           "azure.functions",
	"azure.application_gateway": "azure.application_gateway",
	"azure.load_balancer":       "azure.load_balancer",
	"azure.front_door":          "azure.front_door",
	"azure.blob_storage":        "azure.blob_storage",
	"azure.sql_database":        "azure.sql_database",
	"azure.cosmos_db":           "azure.cosmos_db",
	"azure.redis":               "azure.redis",
	"azure.dns":                 "azure.dns",
	"azure.key_vault":           "azure.key_vault",
	"azure.monitor":             "azure.monitor",
	"azure.active_directory":    "azure.active_directory",
	"azure.service_bus":         "azure.service_bus",

	// ── GCP ──────────────────────────────────────────────────────────
	"gcp.vpc":                  "gcp.vpc",
	"gcp.compute_engine":       "gcp.compute_engine",
	"gcp.gke":                  "gcp.gke",
	"gcp.cloud_run":            "gcp.cloud_run",
	"gcp.app_engine":           "gcp.app_engine",
	"gcp.cloud_functions":      "gcp.cloud_functions",
	"gcp.cloud_load_balancing": "gcp.cloud_load_balancing",
	"gcp.cloud_cdn":            "gcp.cloud_cdn",
	"gcp.cloud_storage":        "gcp.cloud_storage",
	"gcp.cloud_sql":            "gcp.cloud_sql",
	"gcp.bigquery":             "gcp.bigquery",
	"gcp.memorystore":          "gcp.memorystore",
	"gcp.pubsub":               "gcp.pubsub",
	"gcp.cloud_dns":            "gcp.cloud_dns",
	"gcp.secret_manager":       "gcp.secret_manager",
	"gcp.firestore":            "gcp.firestore",
	"gcp.cloud_armor":          "gcp.cloud_armor",

	// ── Kubernetes ───────────────────────────────────────────────────
	"k8s.cluster":     "k8s.cluster",
	"k8s.node":        "k8s.node",
	"k8s.namespace":   "k8s.namespace",
	"k8s.deployment":  "k8s.deployment",
	"k8s.pod":         "k8s.pod",
	"k8s.service":     "k8s.service",
	"k8s.ingress":     "k8s.ingress",
	"k8s.configmap":   "k8s.configmap",
	"k8s.secret":      "k8s.secret",
	"k8s.statefulset": "k8s.statefulset",
	"k8s.daemonset":   "k8s.daemonset",
	"k8s.job":         "k8s.job",

	// ── OCI ──────────────────────────────────────────────────────────
	"oci.vcn":                 "oci.vcn",
	"oci.compute":             "oci.compute",
	"oci.oke":                 "oci.oke",
	"oci.functions":           "oci.functions",
	"oci.load_balancer":       "oci.load_balancer",
	"oci.object_storage":      "oci.object_storage",
	"oci.autonomous_database": "oci.autonomous_database",
	"oci.dns":                 "oci.dns",
	"oci.vault":               "oci.vault",

	// ── Cloudflare ───────────────────────────────────────────────────
	"cloudflare.dns":           "cloudflare.dns",
	"cloudflare.waf":           "cloudflare.waf",
	"cloudflare.cdn":           "cloudflare.cdn",
	"cloudflare.tunnel":        "cloudflare.tunnel",
	"cloudflare.load_balancer": "cloudflare.load_balancer",
	"cloudflare.r2":            "cloudflare.r2",
	"cloudflare.workers":       "cloudflare.workers",
	"cloudflare.access":        "cloudflare.access",

	// ── Solution Architecture ────────────────────────────────────────
	"solution.user":                "solution.user",
	"solution.browser":             "solution.browser",
	"solution.mobile_app":          "solution.mobile_app",
	"solution.frontend":            "solution.frontend",
	"solution.backend_service":     "solution.backend_service",
	"solution.microservice":        "solution.microservice",
	"solution.worker":              "solution.worker",
	"solution.function":            "solution.function",
	"solution.api_gateway":         "solution.api_gateway",
	"solution.webhook":             "solution.webhook",
	"solution.event_bus":           "solution.event_bus",
	"solution.queue":               "solution.queue",
	"solution.cache":               "solution.cache",
	"solution.relational_database": "solution.relational_database",
	"solution.nosql_database":      "solution.nosql_database",
	"solution.object_storage":      "solution.object_storage",
	"solution.search":              "solution.search",
	"solution.analytics":           "solution.analytics",
	"solution.observability":       "solution.observability",
	"solution.vector_store":        "solution.vector_store",
	"solution.ai_service":          "solution.ai_service",
	"solution.ci_cd":               "solution.ci_cd",
	"solution.registry":            "solution.registry",

	// ── Network Architecture ─────────────────────────────────────────
	"network.internet":         "network.internet",
	"network.wan":              "network.wan",
	"network.lan":              "network.lan",
	"network.vpn":              "network.vpn",
	"network.firewall":         "network.firewall",
	"network.waf":              "network.waf",
	"network.ids_ips":          "network.ids_ips",
	"network.nat":              "network.nat",
	"network.subnet":           "network.subnet",
	"network.private_subnet":   "network.private_subnet",
	"network.public_subnet":    "network.public_subnet",
	"network.route_table":      "network.route_table",
	"network.dns":              "network.dns",
	"network.cdn":              "network.cdn",
	"network.reverse_proxy":    "network.reverse_proxy",
	"network.load_balancer_l4": "network.load_balancer_l4",
	"network.load_balancer_l7": "network.load_balancer_l7",
	"network.bastion":          "network.bastion",
	"network.transit_hub":      "network.transit_hub",
	"network.peering":          "network.peering",
	"network.zero_trust_proxy": "network.zero_trust_proxy",
	"network.office":           "network.office",
	"network.datacenter":       "network.datacenter",
	"network.branch_office":    "network.branch_office",
	"network.on_prem":          "network.on_prem",
	"network.mpls":             "network.mpls",
	"network.siem":             "network.siem",
	"network.file_server":      "network.file_server",
	"network.active_directory": "network.active_directory",
	"network.proxy":            "network.proxy",
	"network.switch":           "network.switch",
	"network.router":           "network.router",
	"network.endpoint":         "network.endpoint",
	"network.ztna":             "network.ztna",

	// ── Enterprise Architecture ──────────────────────────────────────
	"enterprise.employee":            "enterprise.employee",
	"enterprise.customer":            "enterprise.customer",
	"enterprise.admin":               "enterprise.admin",
	"enterprise.partner_system":      "enterprise.partner_system",
	"enterprise.saas_app":            "enterprise.saas_app",
	"enterprise.erp":                 "enterprise.erp",
	"enterprise.crm":                 "enterprise.crm",
	"enterprise.hris":                "enterprise.hris",
	"enterprise.identity_provider":   "enterprise.identity_provider",
	"enterprise.sso":                 "enterprise.sso",
	"enterprise.directory":           "enterprise.directory",
	"enterprise.audit":               "enterprise.audit",
	"enterprise.compliance_boundary": "enterprise.compliance_boundary",
	"enterprise.trust_boundary":      "enterprise.trust_boundary",
	"enterprise.business_unit":       "enterprise.business_unit",
	"enterprise.integration_hub":     "enterprise.integration_hub",
	"enterprise.data_warehouse":      "enterprise.data_warehouse",
	"enterprise.bi_reporting":        "enterprise.bi_reporting",
	"enterprise.etl":                 "enterprise.etl",

	// ── Generic (fallback) ───────────────────────────────────────────
	"generic.server":         "generic.server",
	"generic.database":       "generic.database",
	"generic.queue":          "generic.queue",
	"generic.cache":          "generic.cache",
	"generic.load_balancer":  "generic.load_balancer",
	"generic.cdn":            "generic.cdn",
	"generic.dns":            "generic.dns",
	"generic.firewall":       "generic.firewall",
	"generic.storage":        "generic.storage",
	"generic.object_storage": "generic.object_storage",
	"generic.container":      "generic.container",
	"generic.api":            "generic.api",
	"generic.user":           "generic.user",
	"generic.client":         "generic.client",
	"generic.monitoring":     "generic.monitoring",
	"generic.service":        "generic.service",
	"generic.vpn":            "generic.vpn",
	"generic.internet":       "generic.internet",
	"generic.on_prem":        "generic.on_prem",
	"generic.webhook":        "generic.webhook",
	"generic.browser":        "generic.browser",
	"generic.mobile_app":     "generic.mobile_app",
	"generic.office":         "generic.office",
}

// fuzzyAliases maps common variations, shorthand, LLM-likely outputs,
// and legacy colon-separated formats to canonical dot.underscore types.
var fuzzyAliases = map[string]string{
	// ── Legacy colon format (backward compat) ────────────────────────
	// AWS
	"aws:vpc": "aws.vpc", "aws:subnet": "aws.subnet", "aws:ec2": "aws.ec2",
	"aws:ecs": "aws.ecs", "aws:eks": "aws.eks", "aws:lambda": "aws.lambda",
	"aws:alb": "aws.alb", "aws:nlb": "aws.nlb", "aws:elb": "aws.elb",
	"aws:api-gateway": "aws.api_gateway", "aws:cloudfront": "aws.cloudfront",
	"aws:s3": "aws.s3", "aws:rds": "aws.rds", "aws:dynamodb": "aws.dynamodb",
	"aws:elasticache": "aws.elasticache", "aws:route53": "aws.route_53",
	"aws:iam": "aws.iam", "aws:waf": "aws.waf", "aws:nat-gateway": "aws.nat_gateway",
	"aws:internet-gateway": "aws.internet_gateway", "aws:sqs": "aws.sqs",
	"aws:sns": "aws.sns", "aws:eventbridge": "aws.eventbridge",
	"aws:step-functions": "aws.step_functions", "aws:cloudwatch": "aws.cloudwatch",
	"aws:secrets-manager": "aws.secrets_manager", "aws:systems-manager": "aws.systems_manager",
	"aws:ecr": "aws.ecr", "aws:cognito": "aws.cognito", "aws:kinesis": "aws.kinesis",
	// Azure
	"azure:vnet": "azure.vnet", "azure:subnet": "azure.subnet", "azure:vm": "azure.vm",
	"azure:aks": "azure.aks", "azure:app-service": "azure.app_service",
	"azure:functions": "azure.functions", "azure:application-gateway": "azure.application_gateway",
	"azure:load-balancer": "azure.load_balancer", "azure:front-door": "azure.front_door",
	"azure:blob-storage": "azure.blob_storage", "azure:sql-database": "azure.sql_database",
	"azure:cosmos-db": "azure.cosmos_db", "azure:redis-cache": "azure.redis",
	"azure:dns": "azure.dns", "azure:key-vault": "azure.key_vault",
	"azure:monitor": "azure.monitor", "azure:active-directory": "azure.active_directory",
	"azure:service-bus": "azure.service_bus",
	// GCP
	"gcp:vpc": "gcp.vpc", "gcp:compute-engine": "gcp.compute_engine",
	"gcp:gke": "gcp.gke", "gcp:cloud-run": "gcp.cloud_run",
	"gcp:app-engine": "gcp.app_engine", "gcp:cloud-functions": "gcp.cloud_functions",
	"gcp:cloud-load-balancing": "gcp.cloud_load_balancing", "gcp:cloud-cdn": "gcp.cloud_cdn",
	"gcp:cloud-storage": "gcp.cloud_storage", "gcp:cloud-sql": "gcp.cloud_sql",
	"gcp:bigquery": "gcp.bigquery", "gcp:memorystore": "gcp.memorystore",
	"gcp:pub-sub": "gcp.pubsub", "gcp:cloud-dns": "gcp.cloud_dns",
	"gcp:secret-manager": "gcp.secret_manager", "gcp:firestore": "gcp.firestore",
	"gcp:cloud-armor": "gcp.cloud_armor", "gcp:load-balancer": "gcp.cloud_load_balancing",
	// K8s
	"k8s:cluster": "k8s.cluster", "k8s:node": "k8s.node", "k8s:namespace": "k8s.namespace",
	"k8s:deployment": "k8s.deployment", "k8s:pod": "k8s.pod", "k8s:service": "k8s.service",
	"k8s:ingress": "k8s.ingress", "k8s:configmap": "k8s.configmap", "k8s:secret": "k8s.secret",
	"k8s:statefulset": "k8s.statefulset", "k8s:daemonset": "k8s.daemonset", "k8s:job": "k8s.job",
	// OCI
	"oci:vcn": "oci.vcn", "oci:compute": "oci.compute", "oci:oke": "oci.oke",
	"oci:functions": "oci.functions", "oci:load-balancer": "oci.load_balancer",
	"oci:object-storage": "oci.object_storage", "oci:autonomous-database": "oci.autonomous_database",
	"oci:dns": "oci.dns", "oci:vault": "oci.vault",
	// Cloudflare
	"cloudflare:dns": "cloudflare.dns", "cloudflare:waf": "cloudflare.waf",
	"cloudflare:cdn": "cloudflare.cdn", "cloudflare:tunnel": "cloudflare.tunnel",
	"cloudflare:load-balancer": "cloudflare.load_balancer", "cloudflare:r2": "cloudflare.r2",
	"cloudflare:workers": "cloudflare.workers", "cloudflare:access": "cloudflare.access",
	// Generic legacy
	"generic:server": "generic.server", "generic:database": "generic.database",
	"generic:queue": "generic.queue", "generic:cache": "generic.cache",
	"generic:load-balancer": "generic.load_balancer", "generic:cdn": "generic.cdn",
	"generic:dns": "generic.dns", "generic:firewall": "generic.firewall",
	"generic:storage": "generic.storage", "generic:object-storage": "generic.object_storage",
	"generic:container": "generic.container", "generic:api": "generic.api",
	"generic:user": "generic.user", "generic:client": "generic.client",
	"generic:monitoring": "generic.monitoring", "generic:service": "generic.service",
	"generic:vpn": "generic.vpn", "generic:internet": "generic.internet",
	"generic:on-prem": "generic.on_prem", "generic:webhook": "generic.webhook",
	"generic:browser": "generic.browser", "generic:mobile-app": "generic.mobile_app",
	"generic:office": "generic.office",

	// ── AWS shorthand aliases ────────────────────────────────────────
	"lambda": "aws.lambda", "aws lambda": "aws.lambda",
	"ec2": "aws.ec2", "aws ec2": "aws.ec2",
	"s3": "aws.s3", "s3 bucket": "aws.s3", "aws s3": "aws.s3",
	"rds": "aws.rds", "aws rds": "aws.rds",
	"dynamodb": "aws.dynamodb", "dynamo": "aws.dynamodb", "dynamo db": "aws.dynamodb",
	"sqs": "aws.sqs", "aws sqs": "aws.sqs",
	"sns": "aws.sns", "aws sns": "aws.sns",
	"cloudfront": "aws.cloudfront", "aws cloudfront": "aws.cloudfront",
	"api gateway": "aws.api_gateway", "apigateway": "aws.api_gateway",
	"elb": "aws.elb", "alb": "aws.alb", "nlb": "aws.nlb",
	"ecs": "aws.ecs", "eks": "aws.eks",
	"elasticache": "aws.elasticache",
	"route53":     "aws.route_53", "route 53": "aws.route_53",
	"cognito": "aws.cognito", "kinesis": "aws.kinesis",
	"step functions": "aws.step_functions", "stepfunctions": "aws.step_functions",
	"waf": "aws.waf", "vpc": "aws.vpc", "subnet": "aws.subnet",
	"nat gateway": "aws.nat_gateway", "igw": "aws.internet_gateway",
	"internet gateway": "aws.internet_gateway",
	"eventbridge":      "aws.eventbridge", "event bridge": "aws.eventbridge",
	"secrets manager": "aws.secrets_manager", "ssm": "aws.systems_manager",
	"ecr": "aws.ecr", "cloudwatch": "aws.cloudwatch", "iam": "aws.iam",

	// ── Azure shorthand aliases ──────────────────────────────────────
	"virtual machine": "azure.vm", "azure vm": "azure.vm",
	"aks": "azure.aks", "azure aks": "azure.aks",
	"azure functions": "azure.functions",
	"app service":     "azure.app_service", "azure app service": "azure.app_service",
	"cosmos db": "azure.cosmos_db", "cosmosdb": "azure.cosmos_db",
	"service bus": "azure.service_bus", "front door": "azure.front_door",
	"blob storage": "azure.blob_storage",
	"key vault":    "azure.key_vault", "azure ad": "azure.active_directory",
	"vnet": "azure.vnet", "application gateway": "azure.application_gateway",

	// ── GCP shorthand aliases ────────────────────────────────────────
	"cloud run": "gcp.cloud_run", "cloud functions": "gcp.cloud_functions",
	"cloud storage": "gcp.cloud_storage", "gcs": "gcp.cloud_storage",
	"cloud sql": "gcp.cloud_sql", "bigquery": "gcp.bigquery", "big query": "gcp.bigquery",
	"pubsub": "gcp.pubsub", "pub/sub": "gcp.pubsub",
	"gke": "gcp.gke", "compute engine": "gcp.compute_engine",
	"app engine": "gcp.app_engine", "memorystore": "gcp.memorystore",
	"cloud cdn": "gcp.cloud_cdn", "cloud dns": "gcp.cloud_dns",
	"secret manager": "gcp.secret_manager", "firestore": "gcp.firestore",
	"cloud armor": "gcp.cloud_armor", "cloud load balancing": "gcp.cloud_load_balancing",

	// ── K8s shorthand aliases ────────────────────────────────────────
	"deployment": "k8s.deployment", "pod": "k8s.pod",
	"ingress": "k8s.ingress", "configmap": "k8s.configmap", "config map": "k8s.configmap",
	"statefulset": "k8s.statefulset", "daemonset": "k8s.daemonset",
	"cronjob": "k8s.job", "cron job": "k8s.job",
	"namespace":          "k8s.namespace",
	"kubernetes cluster": "k8s.cluster", "kube cluster": "k8s.cluster",

	// ── OCI shorthand aliases ────────────────────────────────────────
	"vcn": "oci.vcn", "oke": "oci.oke",
	"autonomous database": "oci.autonomous_database",

	// ── Cloudflare shorthand aliases ─────────────────────────────────
	"cloudflare": "cloudflare.cdn",
	"cf dns":     "cloudflare.dns", "cf waf": "cloudflare.waf",
	"cf tunnel": "cloudflare.tunnel", "argo tunnel": "cloudflare.tunnel",
	"r2": "cloudflare.r2", "workers": "cloudflare.workers",
	"cf access": "cloudflare.access",

	// ── Solution shorthand aliases ───────────────────────────────────
	"backend service": "solution.backend_service", "backend": "solution.backend_service",
	"microservice":        "solution.microservice",
	"worker":              "solution.worker",
	"function":            "solution.function",
	"event bus":           "solution.event_bus",
	"relational database": "solution.relational_database", "relational db": "solution.relational_database",
	"nosql database": "solution.nosql_database", "nosql db": "solution.nosql_database", "nosql": "solution.nosql_database",
	"object storage": "solution.object_storage",
	"search":         "solution.search", "search engine": "solution.search",
	"elasticsearch": "solution.search", "opensearch": "solution.search",
	"analytics":     "solution.analytics",
	"observability": "solution.observability",
	"vector store":  "solution.vector_store", "vector db": "solution.vector_store",
	"ai service": "solution.ai_service", "ml service": "solution.ai_service",
	"ci cd": "solution.ci_cd", "cicd": "solution.ci_cd", "ci/cd": "solution.ci_cd",
	"registry": "solution.registry", "container registry": "solution.registry",
	"frontend": "solution.frontend",

	// ── Network shorthand aliases ────────────────────────────────────
	"wan": "network.wan", "lan": "network.lan",
	"ids ips": "network.ids_ips", "ids/ips": "network.ids_ips",
	"nat":              "network.nat",
	"private subnet":   "network.private_subnet",
	"public subnet":    "network.public_subnet",
	"route table":      "network.route_table",
	"reverse proxy":    "network.reverse_proxy",
	"l4 load balancer": "network.load_balancer_l4", "l4 lb": "network.load_balancer_l4",
	"l7 load balancer": "network.load_balancer_l7", "l7 lb": "network.load_balancer_l7",
	"bastion": "network.bastion", "bastion host": "network.bastion", "jump box": "network.bastion",
	"transit hub": "network.transit_hub",
	"peering":     "network.peering", "vpc peering": "network.peering",
	"zero trust proxy": "network.zero_trust_proxy", "zero trust": "network.zero_trust_proxy",
	"datacenter": "network.datacenter", "data center": "network.datacenter",
	"branch office": "network.branch_office",
	"mpls":          "network.mpls", "mpls wan": "network.mpls",
	"siem":             "network.siem",
	"file server":      "network.file_server",
	"active directory": "network.active_directory", "ad": "network.active_directory",
	"ztna": "network.ztna", "zero trust network access": "network.ztna",
	"router": "network.router",
	"switch": "network.switch", "network switch": "network.switch",
	"transit gateway": "aws.transit_gateway", "tgw": "aws.transit_gateway",
	"direct connect": "aws.direct_connect",
	"guardduty":      "aws.guardduty", "guard duty": "aws.guardduty",

	// ── Enterprise shorthand aliases ─────────────────────────────────
	"employee":       "enterprise.employee",
	"customer":       "enterprise.customer",
	"admin":          "enterprise.admin",
	"partner system": "enterprise.partner_system",
	"saas app":       "enterprise.saas_app", "saas": "enterprise.saas_app",
	"erp":               "enterprise.erp",
	"crm":               "enterprise.crm",
	"hris":              "enterprise.hris",
	"identity provider": "enterprise.identity_provider", "idp": "enterprise.identity_provider",
	"sso": "enterprise.sso", "single sign on": "enterprise.sso",
	"directory": "enterprise.directory", "ldap": "enterprise.directory",
	"audit": "enterprise.audit", "audit log": "enterprise.audit",
	"compliance boundary": "enterprise.compliance_boundary", "compliance": "enterprise.compliance_boundary",
	"trust boundary":  "enterprise.trust_boundary",
	"business unit":   "enterprise.business_unit",
	"integration hub": "enterprise.integration_hub", "integration": "enterprise.integration_hub",
	"data warehouse": "enterprise.data_warehouse",
	"bi reporting":   "enterprise.bi_reporting", "bi": "enterprise.bi_reporting",
	"etl": "enterprise.etl", "data pipeline": "enterprise.etl",

	// ── Generic shorthand aliases ────────────────────────────────────
	"server": "generic.server", "web server": "generic.server",
	"database": "generic.database", "db": "generic.database",
	"queue": "generic.queue", "message queue": "generic.queue", "mq": "generic.queue",
	"cache": "generic.cache", "redis": "generic.cache", "memcached": "generic.cache",
	"load balancer": "generic.load_balancer", "lb": "generic.load_balancer",
	"cdn":      "generic.cdn",
	"dns":      "generic.dns",
	"firewall": "generic.firewall",
	"storage":  "generic.storage",
	"bucket":   "generic.object_storage", "blob": "generic.object_storage",
	"container": "generic.container", "docker": "generic.container",
	"api": "generic.api", "rest api": "generic.api",
	"user": "generic.user", "users": "generic.user",
	"client":  "generic.client",
	"browser": "generic.browser", "web browser": "generic.browser",
	"mobile app": "generic.mobile_app", "mobile": "generic.mobile_app",
	"monitoring": "generic.monitoring",
	"vpn":        "generic.vpn",
	"internet":   "generic.internet",
	"on prem":    "generic.on_prem", "on premises": "generic.on_prem", "on-premises": "generic.on_prem",
	"office":  "generic.office",
	"webhook": "generic.webhook",
	"service": "generic.service",
}

// Normalize takes a provider response and normalizes all node types
// to canonical <family>.<component> format with underscore separation.
func Normalize(resp *provider.GenerateResponse) {
	for i := range resp.Nodes {
		resp.Nodes[i].Type = normalizeType(resp.Nodes[i].Type)
	}

	// Default empty edge styles to "solid"
	for i := range resp.Edges {
		if resp.Edges[i].Style == "" {
			resp.Edges[i].Style = "solid"
		}
	}

	// Ensure groups is never nil
	if resp.Groups == nil {
		resp.Groups = []provider.Group{}
	}
}

// SalvageReport summarizes what Salvage stripped out — surface via logs so
// we can spot patterns (e.g. one model consistently emitting phantom edges)
// without blowing up the actual generation.
type SalvageReport struct {
	DroppedEdges   int
	OrphanGroupRef int
}

// Salvage prunes soft-broken pieces of a generated diagram so a single bad
// edge from the LLM doesn't nuke an otherwise valid response. Anything truly
// malformed (no nodes, missing IDs, duplicate IDs) still trips Validate.
//
// Specifically it:
//   - drops edges whose source or target is empty or references an unknown node
//   - clears node.Group when it points at a non-existent group
//
// Returns a report so callers can log how much was salvaged. Always call this
// AFTER Normalize and BEFORE Validate.
func Salvage(resp *provider.GenerateResponse) SalvageReport {
	var report SalvageReport

	nodeIDs := make(map[string]bool, len(resp.Nodes))
	for _, n := range resp.Nodes {
		if n.ID != "" {
			nodeIDs[n.ID] = true
		}
	}
	groupIDs := make(map[string]bool, len(resp.Groups))
	for _, g := range resp.Groups {
		if g.ID != "" {
			groupIDs[g.ID] = true
		}
	}

	if len(resp.Edges) > 0 {
		kept := resp.Edges[:0]
		for _, e := range resp.Edges {
			if e.Source == "" || e.Target == "" || !nodeIDs[e.Source] || !nodeIDs[e.Target] {
				report.DroppedEdges++
				continue
			}
			kept = append(kept, e)
		}
		resp.Edges = kept
	}

	for i, n := range resp.Nodes {
		if n.Group != "" && !groupIDs[n.Group] {
			resp.Nodes[i].Group = ""
			report.OrphanGroupRef++
		}
	}

	return report
}

func normalizeType(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))

	// Normalize whitespace
	lower = strings.Join(strings.Fields(lower), " ")

	// Direct match in canonical types (already in dot.underscore format)
	if canonical, ok := knownTypes[lower]; ok {
		return canonical
	}
	if _, ok := generatedCatalogTypes[lower]; ok {
		return lower
	}

	// Try fuzzy aliases (handles legacy colon format, shorthand, etc.)
	if canonical, ok := fuzzyAliases[lower]; ok {
		return canonical
	}
	if canonical, ok := generatedCatalogAliases[lower]; ok {
		return canonical
	}

	// Convert hyphens to underscores and colons to dots for another lookup pass
	normalized := strings.ReplaceAll(lower, "-", "_")
	normalized = strings.ReplaceAll(normalized, ":", ".")
	normalized = strings.ReplaceAll(normalized, " ", "_")

	if canonical, ok := knownTypes[normalized]; ok {
		return canonical
	}
	if _, ok := generatedCatalogTypes[normalized]; ok {
		return normalized
	}

	// Try alias with hyphens→spaces (catches "cloud-run" → "cloud run" → gcp.cloud_run)
	withSpaces := strings.ReplaceAll(lower, "-", " ")
	withSpaces = strings.ReplaceAll(withSpaces, "_", " ")
	if canonical, ok := fuzzyAliases[withSpaces]; ok {
		return canonical
	}
	if canonical, ok := generatedCatalogAliases[withSpaces]; ok {
		return canonical
	}

	// If it already has a family prefix (contains "." or ":"), keep it in canonical format
	if strings.Contains(lower, ".") || strings.Contains(lower, ":") {
		result := strings.ReplaceAll(lower, ":", ".")
		result = strings.ReplaceAll(result, "-", "_")
		result = strings.ReplaceAll(result, " ", "_")
		return result
	}

	// Fall back to generic.<name>
	fallback := strings.ReplaceAll(lower, " ", "_")
	fallback = strings.ReplaceAll(fallback, "-", "_")
	return "generic." + fallback
}
