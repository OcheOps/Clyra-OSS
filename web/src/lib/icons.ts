import type { TypeFamily } from './constants';
import { COMPONENTS } from './generated/componentCatalog';
import { inferLucideIcon, EXTENDED_FAMILY_COLORS } from './iconHeuristics';

export interface IconDef {
  label: string;
  color: string;
  icon: string; // Lucide icon name
}

const FAMILY_COLORS: Record<TypeFamily, string> = {
  aws: '#FF9900',
  gcp: '#4285F4',
  azure: '#0078D4',
  k8s: '#326CE5',
  oci: '#F80000',
  cloudflare: '#F6821F',
  solution: '#10B981',
  network: '#6366F1',
  enterprise: '#8B5CF6',
  generic: '#8892a8',
};

// Map generated-catalog provider codes → FAMILY_COLORS / EXTENDED_FAMILY_COLORS.
// The md files use `net` / `ent` which maps to `network` / `enterprise`.
function colorFor(provider: string): string {
  if (provider in FAMILY_COLORS) return FAMILY_COLORS[provider as TypeFamily];
  if (provider === 'net') return FAMILY_COLORS.network;
  if (provider === 'ent') return FAMILY_COLORS.enterprise;
  return EXTENDED_FAMILY_COLORS[provider] ?? FAMILY_COLORS.generic;
}

const ICON_REGISTRY: Record<string, IconDef> = {
  // ── AWS (40+ services) ──────────────────────────────────────────
  'aws.vpc':                { label: 'VPC',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Networking-Content-Delivery/Virtual-Private-Cloud.svg' },
  'aws.subnet':             { label: 'Subnet',           color: FAMILY_COLORS.aws, icon: 'grid-2x2' },
  'aws.ec2':                { label: 'EC2',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Compute/EC2.svg' },
  'aws.ecs':                { label: 'ECS',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Containers/Elastic-Container-Service.svg' },
  'aws.eks':                { label: 'EKS',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Containers/Elastic-Kubernetes-Service.svg' },
  'aws.lambda':             { label: 'Lambda',           color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Compute/Lambda.svg' },
  'aws.fargate':            { label: 'Fargate',          color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Compute/Fargate.svg' },
  'aws.alb':                { label: 'ALB',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Networking-Content-Delivery/Elastic-Load-Balancing.svg' },
  'aws.nlb':                { label: 'NLB',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Networking-Content-Delivery/Elastic-Load-Balancing.svg' },
  'aws.elb':                { label: 'ELB',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Networking-Content-Delivery/Elastic-Load-Balancing.svg' },
  'aws.api_gateway':        { label: 'API Gateway',      color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/App-Integration/API-Gateway.svg' },
  'aws.cloudfront':         { label: 'CloudFront',       color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Networking-Content-Delivery/CloudFront.svg' },
  'aws.s3':                 { label: 'S3',               color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Storage/Simple-Storage-Service.svg' },
  'aws.rds':                { label: 'RDS',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Database/RDS.svg' },
  'aws.aurora':             { label: 'Aurora',           color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Database/Aurora.svg' },
  'aws.dynamodb':           { label: 'DynamoDB',         color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Database/DynamoDB.svg' },
  'aws.elasticache':        { label: 'ElastiCache',      color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Database/ElastiCache.svg' },
  'aws.redshift':           { label: 'Redshift',         color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Analytics/Redshift.svg' },
  'aws.route_53':           { label: 'Route 53',         color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Networking-Content-Delivery/Route-53.svg' },
  'aws.iam':                { label: 'IAM',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Security-Identity-Compliance/Identity-and-Access-Management.svg' },
  'aws.waf':                { label: 'WAF',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Security-Identity-Compliance/WAF.svg' },
  'aws.shield':             { label: 'Shield',           color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Security-Identity-Compliance/Shield.svg' },
  'aws.nat_gateway':        { label: 'NAT Gateway',      color: FAMILY_COLORS.aws, icon: 'arrow-right-left' },
  'aws.internet_gateway':   { label: 'Internet GW',      color: FAMILY_COLORS.aws, icon: 'globe' },
  'aws.sqs':                { label: 'SQS',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/App-Integration/Simple-Queue-Service.svg' },
  'aws.sns':                { label: 'SNS',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/App-Integration/Simple-Notification-Service.svg' },
  'aws.eventbridge':        { label: 'EventBridge',      color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/App-Integration/EventBridge.svg' },
  'aws.step_functions':     { label: 'Step Functions',    color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/App-Integration/Step-Functions.svg' },
  'aws.cloudwatch':         { label: 'CloudWatch',       color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Management-Governance/CloudWatch.svg' },
  'aws.cloudtrail':         { label: 'CloudTrail',       color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Management-Governance/CloudTrail.svg' },
  'aws.secrets_manager':    { label: 'Secrets Manager',   color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Security-Identity-Compliance/Secrets-Manager.svg' },
  'aws.systems_manager':    { label: 'Systems Manager',   color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Management-Governance/Systems-Manager.svg' },
  'aws.ecr':                { label: 'ECR',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Containers/Elastic-Container-Registry.svg' },
  'aws.cognito':            { label: 'Cognito',          color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Security-Identity-Compliance/Cognito.svg' },
  'aws.kinesis':            { label: 'Kinesis',          color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Analytics/Kinesis.svg' },
  'aws.transit_gateway':    { label: 'Transit GW',       color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Networking-Content-Delivery/Transit-Gateway.svg' },
  'aws.direct_connect':     { label: 'Direct Connect',   color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Networking-Content-Delivery/Direct-Connect.svg' },
  'aws.config':             { label: 'Config',           color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Management-Governance/Config.svg' },
  'aws.guardduty':          { label: 'GuardDuty',        color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Security-Identity-Compliance/GuardDuty.svg' },
  'aws.ground_station':     { label: 'Ground Station',   color: FAMILY_COLORS.aws, icon: 'satellite-dish' },
  'aws.sagemaker':          { label: 'SageMaker',        color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Machine-Learning/SageMaker.svg' },
  'aws.bedrock':            { label: 'Bedrock',          color: FAMILY_COLORS.aws, icon: 'bot' },
  'aws.opensearch':         { label: 'OpenSearch',       color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Analytics/OpenSearch-Service.svg' },
  'aws.msk':                { label: 'MSK',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Analytics/Managed-Streaming-for-Apache-Kafka.svg' },
  'aws.efs':                { label: 'EFS',              color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Storage/EFS.svg' },
  'aws.backup':             { label: 'Backup',           color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Storage/Backup.svg' },
  'aws.codepipeline':       { label: 'CodePipeline',     color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Developer-Tools/CodePipeline.svg' },
  'aws.codebuild':          { label: 'CodeBuild',        color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Developer-Tools/CodeBuild.svg' },
  'aws.elastic_beanstalk':  { label: 'Elastic Beanstalk',color: FAMILY_COLORS.aws, icon: 'cdn:AWS/svg/Compute/Elastic-Beanstalk.svg' },
  'aws.app_runner':         { label: 'App Runner',       color: FAMILY_COLORS.aws, icon: 'play' },

  // ── GCP (30+ services) ──────────────────────────────────────────
  'gcp.vpc':                  { label: 'VPC',                color: FAMILY_COLORS.gcp, icon: 'square-dashed' },
  'gcp.compute_engine':       { label: 'Compute Engine',     color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Compute-Engine.svg' },
  'gcp.gke':                  { label: 'GKE',                color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Google-Kubernetes-Engine.svg' },
  'gcp.cloud_run':            { label: 'Cloud Run',          color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Cloud-Run.svg' },
  'gcp.app_engine':           { label: 'App Engine',         color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/App-Engine.svg' },
  'gcp.cloud_functions':      { label: 'Cloud Functions',    color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Cloud-Functions.svg' },
  'gcp.cloud_load_balancing': { label: 'Cloud LB',           color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Cloud-Load-Balancing.svg' },
  'gcp.cloud_cdn':            { label: 'Cloud CDN',          color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Cloud-CDN.svg' },
  'gcp.cloud_storage':        { label: 'Cloud Storage',      color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Cloud-Storage.svg' },
  'gcp.cloud_sql':            { label: 'Cloud SQL',          color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Cloud-SQL.svg' },
  'gcp.bigquery':             { label: 'BigQuery',           color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/BigQuery.svg' },
  'gcp.bigtable':             { label: 'Bigtable',           color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Bigtable.svg' },
  'gcp.memorystore':          { label: 'Memorystore',        color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Memorystore.svg' },
  'gcp.pubsub':                { label: 'Pub/Sub',            color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/PubSub.svg' },
  'gcp.cloud_dns':            { label: 'Cloud DNS',          color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Cloud-DNS.svg' },
  'gcp.secret_manager':       { label: 'Secret Manager',     color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Secret-Manager.svg' },
  'gcp.firestore':            { label: 'Firestore',          color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Firestore.svg' },
  'gcp.cloud_armor':          { label: 'Cloud Armor',        color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Cloud-Armor.svg' },
  'gcp.vertex_ai':            { label: 'Vertex AI',          color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Vertex-AI.svg' },
  'gcp.cloud_build':          { label: 'Cloud Build',        color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Cloud-Build.svg' },
  'gcp.artifact_registry':    { label: 'Artifact Registry',  color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Artifact-Registry.svg' },
  'gcp.spanner':              { label: 'Spanner',            color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Cloud-Spanner.svg' },
  'gcp.dataflow':             { label: 'Dataflow',           color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Dataflow.svg' },
  'gcp.cloud_nat':            { label: 'Cloud NAT',          color: FAMILY_COLORS.gcp, icon: 'cdn:GCP/svg/Cloud-NAT.svg' },

  // ── Azure (30+ services) ────────────────────────────────────────
  'azure.vnet':                { label: 'VNet',               color: FAMILY_COLORS.azure, icon: 'square-dashed' },
  'azure.subnet':              { label: 'Subnet',             color: FAMILY_COLORS.azure, icon: 'grid-2x2' },
  'azure.vm':                  { label: 'Virtual Machine',    color: FAMILY_COLORS.azure, icon: 'iconify:logos:microsoft-azure' },
  'azure.aks':                 { label: 'AKS',                color: FAMILY_COLORS.azure, icon: 'iconify:logos:microsoft-azure' },
  'azure.app_service':         { label: 'App Service',        color: FAMILY_COLORS.azure, icon: 'iconify:logos:microsoft-azure' },
  'azure.functions':           { label: 'Functions',           color: FAMILY_COLORS.azure, icon: 'iconify:logos:microsoft-azure' },
  'azure.container_instances': { label: 'Container Instances', color: FAMILY_COLORS.azure, icon: 'container' },
  'azure.application_gateway': { label: 'App Gateway',        color: FAMILY_COLORS.azure, icon: 'door-open' },
  'azure.load_balancer':       { label: 'Load Balancer',      color: FAMILY_COLORS.azure, icon: 'split' },
  'azure.front_door':          { label: 'Front Door',         color: FAMILY_COLORS.azure, icon: 'door-open' },
  'azure.blob_storage':        { label: 'Blob Storage',       color: FAMILY_COLORS.azure, icon: 'iconify:logos:microsoft-azure' },
  'azure.sql_database':        { label: 'SQL Database',       color: FAMILY_COLORS.azure, icon: 'iconify:logos:microsoft-azure' },
  'azure.cosmos_db':           { label: 'Cosmos DB',          color: FAMILY_COLORS.azure, icon: 'iconify:logos:microsoft-azure' },
  'azure.redis':               { label: 'Redis Cache',        color: FAMILY_COLORS.azure, icon: 'memory-stick' },
  'azure.dns':                 { label: 'DNS',                color: FAMILY_COLORS.azure, icon: 'signpost' },
  'azure.key_vault':           { label: 'Key Vault',          color: FAMILY_COLORS.azure, icon: 'iconify:logos:microsoft-azure' },
  'azure.monitor':             { label: 'Monitor',            color: FAMILY_COLORS.azure, icon: 'activity' },
  'azure.active_directory':    { label: 'Entra ID',           color: FAMILY_COLORS.azure, icon: 'user-check' },
  'azure.service_bus':         { label: 'Service Bus',        color: FAMILY_COLORS.azure, icon: 'bus' },
  'azure.event_hub':           { label: 'Event Hub',          color: FAMILY_COLORS.azure, icon: 'radio-tower' },
  'azure.synapse':             { label: 'Synapse',            color: FAMILY_COLORS.azure, icon: 'bar-chart-3' },
  'azure.data_factory':        { label: 'Data Factory',       color: FAMILY_COLORS.azure, icon: 'workflow' },
  'azure.devops':              { label: 'DevOps',             color: FAMILY_COLORS.azure, icon: 'git-branch' },
  'azure.firewall':            { label: 'Firewall',           color: FAMILY_COLORS.azure, icon: 'shield' },
  'azure.bastion':             { label: 'Bastion',            color: FAMILY_COLORS.azure, icon: 'castle' },
  'azure.vpn_gateway':         { label: 'VPN Gateway',        color: FAMILY_COLORS.azure, icon: 'lock' },

  // ── Kubernetes ───────────────────────────────────────────────────
  'k8s.cluster':     { label: 'Cluster',     color: FAMILY_COLORS.k8s, icon: 'iconify:logos:kubernetes' },
  'k8s.node':        { label: 'Node',        color: FAMILY_COLORS.k8s, icon: 'iconify:logos:kubernetes' },
  'k8s.namespace':   { label: 'Namespace',   color: FAMILY_COLORS.k8s, icon: 'iconify:logos:kubernetes' },
  'k8s.deployment':  { label: 'Deployment',  color: FAMILY_COLORS.k8s, icon: 'iconify:logos:kubernetes' },
  'k8s.pod':         { label: 'Pod',         color: FAMILY_COLORS.k8s, icon: 'iconify:logos:kubernetes' },
  'k8s.service':     { label: 'Service',     color: FAMILY_COLORS.k8s, icon: 'iconify:logos:kubernetes' },
  'k8s.ingress':     { label: 'Ingress',     color: FAMILY_COLORS.k8s, icon: 'iconify:logos:kubernetes' },
  'k8s.configmap':   { label: 'ConfigMap',   color: FAMILY_COLORS.k8s, icon: 'file-cog' },
  'k8s.secret':      { label: 'Secret',      color: FAMILY_COLORS.k8s, icon: 'lock-keyhole' },
  'k8s.statefulset': { label: 'StatefulSet', color: FAMILY_COLORS.k8s, icon: 'iconify:logos:kubernetes' },
  'k8s.daemonset':   { label: 'DaemonSet',   color: FAMILY_COLORS.k8s, icon: 'copy' },
  'k8s.job':         { label: 'Job',         color: FAMILY_COLORS.k8s, icon: 'timer' },
  'k8s.hpa':         { label: 'HPA',         color: FAMILY_COLORS.k8s, icon: 'scaling' },
  'k8s.pvc':         { label: 'PVC',         color: FAMILY_COLORS.k8s, icon: 'hard-drive' },

  // ── OCI ──────────────────────────────────────────────────────────
  'oci.vcn':                 { label: 'VCN',            color: FAMILY_COLORS.oci, icon: 'square-dashed' },
  'oci.compute':             { label: 'Compute',        color: FAMILY_COLORS.oci, icon: 'server' },
  'oci.oke':                 { label: 'OKE',            color: FAMILY_COLORS.oci, icon: 'hexagon' },
  'oci.functions':           { label: 'Functions',       color: FAMILY_COLORS.oci, icon: 'zap' },
  'oci.load_balancer':       { label: 'Load Balancer',  color: FAMILY_COLORS.oci, icon: 'split' },
  'oci.object_storage':      { label: 'Object Storage', color: FAMILY_COLORS.oci, icon: 'bucket' },
  'oci.autonomous_database': { label: 'Autonomous DB',  color: FAMILY_COLORS.oci, icon: 'database' },
  'oci.dns':                 { label: 'DNS',            color: FAMILY_COLORS.oci, icon: 'signpost' },
  'oci.vault':               { label: 'Vault',          color: FAMILY_COLORS.oci, icon: 'lock-keyhole' },

  // ── Cloudflare ──────────────────────────────────────────────────
  'cloudflare.dns':           { label: 'DNS',           color: FAMILY_COLORS.cloudflare, icon: 'iconify:logos:cloudflare' },
  'cloudflare.waf':           { label: 'WAF',           color: FAMILY_COLORS.cloudflare, icon: 'iconify:logos:cloudflare' },
  'cloudflare.cdn':           { label: 'CDN',           color: FAMILY_COLORS.cloudflare, icon: 'iconify:logos:cloudflare' },
  'cloudflare.tunnel':        { label: 'Tunnel',        color: FAMILY_COLORS.cloudflare, icon: 'iconify:logos:cloudflare' },
  'cloudflare.load_balancer': { label: 'Load Balancer', color: FAMILY_COLORS.cloudflare, icon: 'iconify:logos:cloudflare' },
  'cloudflare.r2':            { label: 'R2',            color: FAMILY_COLORS.cloudflare, icon: 'iconify:logos:cloudflare' },
  'cloudflare.workers':       { label: 'Workers',       color: FAMILY_COLORS.cloudflare, icon: 'iconify:logos:cloudflare-workers' },
  'cloudflare.access':        { label: 'Access',        color: FAMILY_COLORS.cloudflare, icon: 'iconify:logos:cloudflare' },
  'cloudflare.pages':         { label: 'Pages',         color: FAMILY_COLORS.cloudflare, icon: 'iconify:logos:cloudflare' },

  // ── Solution Architecture ───────────────────────────────────────
  'solution.user':                { label: 'User',            color: FAMILY_COLORS.solution, icon: 'user' },
  'solution.browser':             { label: 'Browser',         color: FAMILY_COLORS.solution, icon: 'globe' },
  'solution.mobile_app':          { label: 'Mobile App',      color: FAMILY_COLORS.solution, icon: 'smartphone' },
  'solution.frontend':            { label: 'Frontend',        color: FAMILY_COLORS.solution, icon: 'layout' },
  'solution.backend_service':     { label: 'Backend Service', color: FAMILY_COLORS.solution, icon: 'server' },
  'solution.microservice':        { label: 'Microservice',    color: FAMILY_COLORS.solution, icon: 'boxes' },
  'solution.worker':              { label: 'Worker',          color: FAMILY_COLORS.solution, icon: 'cog' },
  'solution.function':            { label: 'Function',        color: FAMILY_COLORS.solution, icon: 'zap' },
  'solution.api_gateway':         { label: 'API Gateway',     color: FAMILY_COLORS.solution, icon: 'door-open' },
  'solution.webhook':             { label: 'Webhook',         color: FAMILY_COLORS.solution, icon: 'webhook' },
  'solution.event_bus':           { label: 'Event Bus',       color: FAMILY_COLORS.solution, icon: 'bus' },
  'solution.queue':               { label: 'Queue',           color: FAMILY_COLORS.solution, icon: 'mail' },
  'solution.cache':               { label: 'Cache',           color: FAMILY_COLORS.solution, icon: 'memory-stick' },
  'solution.relational_database': { label: 'Relational DB',   color: FAMILY_COLORS.solution, icon: 'database' },
  'solution.nosql_database':      { label: 'NoSQL DB',        color: FAMILY_COLORS.solution, icon: 'table' },
  'solution.object_storage':      { label: 'Object Storage',  color: FAMILY_COLORS.solution, icon: 'bucket' },
  'solution.search':              { label: 'Search',          color: FAMILY_COLORS.solution, icon: 'search' },
  'solution.analytics':           { label: 'Analytics',       color: FAMILY_COLORS.solution, icon: 'bar-chart-3' },
  'solution.observability':       { label: 'Observability',   color: FAMILY_COLORS.solution, icon: 'activity' },
  'solution.vector_store':        { label: 'Vector Store',    color: FAMILY_COLORS.solution, icon: 'brain' },
  'solution.ai_service':          { label: 'AI Service',      color: FAMILY_COLORS.solution, icon: 'bot' },
  'solution.ci_cd':               { label: 'CI/CD',           color: FAMILY_COLORS.solution, icon: 'git-branch' },
  'solution.registry':            { label: 'Registry',        color: FAMILY_COLORS.solution, icon: 'package' },

  // ── Network Architecture (Cisco-style) ──────────────────────────
  'network.internet':          { label: 'Internet',         color: FAMILY_COLORS.network, icon: 'globe' },
  'network.wan':               { label: 'WAN',              color: FAMILY_COLORS.network, icon: 'cloud' },
  'network.lan':               { label: 'LAN',              color: FAMILY_COLORS.network, icon: 'network' },
  'network.vpn':               { label: 'VPN',              color: FAMILY_COLORS.network, icon: 'lock' },
  'network.firewall':          { label: 'Firewall',         color: FAMILY_COLORS.network, icon: 'shield' },
  'network.waf':               { label: 'WAF',              color: FAMILY_COLORS.network, icon: 'shield-check' },
  'network.ids_ips':           { label: 'IDS/IPS',          color: FAMILY_COLORS.network, icon: 'scan-eye' },
  'network.nat':               { label: 'NAT',              color: FAMILY_COLORS.network, icon: 'arrow-right-left' },
  'network.subnet':            { label: 'Subnet',           color: FAMILY_COLORS.network, icon: 'grid-2x2' },
  'network.private_subnet':    { label: 'Private Subnet',   color: FAMILY_COLORS.network, icon: 'lock' },
  'network.public_subnet':     { label: 'Public Subnet',    color: FAMILY_COLORS.network, icon: 'globe' },
  'network.route_table':       { label: 'Route Table',      color: FAMILY_COLORS.network, icon: 'route' },
  'network.dns':               { label: 'DNS',              color: FAMILY_COLORS.network, icon: 'signpost' },
  'network.cdn':               { label: 'CDN',              color: FAMILY_COLORS.network, icon: 'globe' },
  'network.reverse_proxy':     { label: 'Reverse Proxy',    color: FAMILY_COLORS.network, icon: 'arrow-left-right' },
  'network.load_balancer_l4':  { label: 'L4 LB',            color: FAMILY_COLORS.network, icon: 'split' },
  'network.load_balancer_l7':  { label: 'L7 LB',            color: FAMILY_COLORS.network, icon: 'split' },
  'network.bastion':           { label: 'Bastion',          color: FAMILY_COLORS.network, icon: 'castle' },
  'network.transit_hub':       { label: 'Transit Hub',      color: FAMILY_COLORS.network, icon: 'network' },
  'network.peering':           { label: 'Peering',          color: FAMILY_COLORS.network, icon: 'link' },
  'network.zero_trust_proxy':  { label: 'Zero Trust',       color: FAMILY_COLORS.network, icon: 'fingerprint' },
  'network.office':            { label: 'Office',           color: FAMILY_COLORS.network, icon: 'building-2' },
  'network.datacenter':        { label: 'Datacenter',       color: FAMILY_COLORS.network, icon: 'server' },
  'network.branch_office':     { label: 'Branch Office',    color: FAMILY_COLORS.network, icon: 'building' },
  'network.on_prem':           { label: 'On-Prem',          color: FAMILY_COLORS.network, icon: 'building-2' },
  'network.mpls':              { label: 'MPLS',             color: FAMILY_COLORS.network, icon: 'cable' },
  'network.siem':              { label: 'SIEM',             color: FAMILY_COLORS.network, icon: 'scan-eye' },
  'network.file_server':       { label: 'File Server',      color: FAMILY_COLORS.network, icon: 'hard-drive' },
  'network.active_directory':  { label: 'Active Directory', color: FAMILY_COLORS.network, icon: 'user-check' },
  'network.proxy':             { label: 'Proxy',            color: FAMILY_COLORS.network, icon: 'arrow-left-right' },
  'network.switch':            { label: 'Switch',           color: FAMILY_COLORS.network, icon: 'git-fork' },
  'network.router':            { label: 'Router',           color: FAMILY_COLORS.network, icon: 'router' },
  'network.endpoint':          { label: 'Endpoint',         color: FAMILY_COLORS.network, icon: 'monitor' },
  'network.ztna':              { label: 'ZTNA',             color: FAMILY_COLORS.network, icon: 'fingerprint' },
  'network.access_point':      { label: 'Access Point',     color: FAMILY_COLORS.network, icon: 'wifi' },

  // ── Enterprise Architecture ─────────────────────────────────────
  'enterprise.employee':            { label: 'Employee',           color: FAMILY_COLORS.enterprise, icon: 'user' },
  'enterprise.customer':            { label: 'Customer',           color: FAMILY_COLORS.enterprise, icon: 'users' },
  'enterprise.admin':               { label: 'Admin',              color: FAMILY_COLORS.enterprise, icon: 'user-cog' },
  'enterprise.partner_system':      { label: 'Partner System',     color: FAMILY_COLORS.enterprise, icon: 'handshake' },
  'enterprise.saas_app':            { label: 'SaaS App',           color: FAMILY_COLORS.enterprise, icon: 'cloud' },
  'enterprise.erp':                 { label: 'ERP',                color: FAMILY_COLORS.enterprise, icon: 'factory' },
  'enterprise.crm':                 { label: 'CRM',                color: FAMILY_COLORS.enterprise, icon: 'contact' },
  'enterprise.hris':                { label: 'HRIS',               color: FAMILY_COLORS.enterprise, icon: 'users' },
  'enterprise.identity_provider':   { label: 'IdP',                color: FAMILY_COLORS.enterprise, icon: 'fingerprint' },
  'enterprise.sso':                 { label: 'SSO',                color: FAMILY_COLORS.enterprise, icon: 'key-round' },
  'enterprise.directory':           { label: 'Directory',          color: FAMILY_COLORS.enterprise, icon: 'book-user' },
  'enterprise.audit':               { label: 'Audit',              color: FAMILY_COLORS.enterprise, icon: 'clipboard-check' },
  'enterprise.compliance_boundary': { label: 'Compliance',         color: FAMILY_COLORS.enterprise, icon: 'shield-check' },
  'enterprise.trust_boundary':      { label: 'Trust Boundary',     color: FAMILY_COLORS.enterprise, icon: 'shield' },
  'enterprise.business_unit':       { label: 'Business Unit',      color: FAMILY_COLORS.enterprise, icon: 'building-2' },
  'enterprise.integration_hub':     { label: 'Integration Hub',    color: FAMILY_COLORS.enterprise, icon: 'git-merge' },
  'enterprise.data_warehouse':      { label: 'Data Warehouse',     color: FAMILY_COLORS.enterprise, icon: 'warehouse' },
  'enterprise.bi_reporting':        { label: 'BI / Reporting',     color: FAMILY_COLORS.enterprise, icon: 'bar-chart-3' },
  'enterprise.etl':                 { label: 'ETL',                color: FAMILY_COLORS.enterprise, icon: 'workflow' },

  // ── Generic ─────────────────────────────────────────────────────
  'generic.server':         { label: 'Server',         color: FAMILY_COLORS.generic, icon: 'server' },
  'generic.database':       { label: 'Database',       color: FAMILY_COLORS.generic, icon: 'database' },
  'generic.queue':          { label: 'Queue',          color: FAMILY_COLORS.generic, icon: 'mail' },
  'generic.cache':          { label: 'Cache',          color: FAMILY_COLORS.generic, icon: 'memory-stick' },
  'generic.load_balancer':  { label: 'Load Balancer',  color: FAMILY_COLORS.generic, icon: 'split' },
  'generic.cdn':            { label: 'CDN',            color: FAMILY_COLORS.generic, icon: 'globe' },
  'generic.dns':            { label: 'DNS',            color: FAMILY_COLORS.generic, icon: 'signpost' },
  'generic.firewall':       { label: 'Firewall',       color: FAMILY_COLORS.generic, icon: 'shield' },
  'generic.storage':        { label: 'Storage',        color: FAMILY_COLORS.generic, icon: 'hard-drive' },
  'generic.object_storage': { label: 'Object Storage', color: FAMILY_COLORS.generic, icon: 'bucket' },
  'generic.container':      { label: 'Container',      color: FAMILY_COLORS.generic, icon: 'container' },
  'generic.api':            { label: 'API',            color: FAMILY_COLORS.generic, icon: 'plug' },
  'generic.user':           { label: 'User',           color: FAMILY_COLORS.generic, icon: 'user' },
  'generic.client':         { label: 'Client',         color: FAMILY_COLORS.generic, icon: 'monitor' },
  'generic.monitoring':     { label: 'Monitoring',     color: FAMILY_COLORS.generic, icon: 'activity' },
  'generic.service':        { label: 'Service',        color: FAMILY_COLORS.generic, icon: 'cog' },
  'generic.vpn':            { label: 'VPN',            color: FAMILY_COLORS.generic, icon: 'lock' },
  'generic.internet':       { label: 'Internet',       color: FAMILY_COLORS.generic, icon: 'globe' },
  'generic.on_prem':        { label: 'On-Prem',        color: FAMILY_COLORS.generic, icon: 'building-2' },
  'generic.webhook':        { label: 'Webhook',        color: FAMILY_COLORS.generic, icon: 'webhook' },
  'generic.browser':        { label: 'Browser',        color: FAMILY_COLORS.generic, icon: 'globe' },
  'generic.mobile_app':     { label: 'Mobile App',     color: FAMILY_COLORS.generic, icon: 'smartphone' },
  'generic.office':         { label: 'Office',         color: FAMILY_COLORS.generic, icon: 'building' },
};

// Fallback icon names per family
const FAMILY_FALLBACK_ICON: Record<string, string> = {
  aws: 'cloud', gcp: 'cloud', azure: 'cloud', k8s: 'hexagon',
  oci: 'cloud', cloudflare: 'cloud', solution: 'cog',
  network: 'network', enterprise: 'building-2', generic: 'circle',
  net: 'network', ent: 'building-2', data: 'database',
  edge: 'globe', iot: 'cpu',
};

// Derived registry from components/*.md (via build-catalog.mjs).
// Hand-curated ICON_REGISTRY wins; this fills the long tail so every
// canonical renders a real icon instead of a cloud-family fallback.
// Long-tail brand overrides — canonical → full icon id ready for
// IconRenderer. Values may use any dispatch prefix: `cdn:<path>`
// (vendor SVG via icepanel), `iconify:<id>` (vendored brand asset), or a
// bare slug (passed to iconify as `logos:<slug>` for back-compat).
// Used when the hand-curated ICON_REGISTRY doesn't already specify
// an icon. Lookup chain: ICON_REGISTRY (hand-curated wins)
// → GENERATED + this override → pure lucide fallback.
const BRAND_OVERRIDES: Record<string, string> = {
  // AWS long tail — official vendor SVGs via icepanel CDN.
  'aws.amplify':             'cdn:AWS/svg/Front-End-Web-Mobile/Amplify.svg',
  'aws.appflow':             'cdn:AWS/svg/App-Integration/AppFlow.svg',
  'aws.appsync':             'cdn:AWS/svg/App-Integration/AppSync.svg',
  'aws.athena':              'cdn:AWS/svg/Analytics/Athena.svg',
  'aws.batch':               'cdn:AWS/svg/Compute/Batch.svg',
  'aws.cloudformation':      'cdn:AWS/svg/Management-Governance/CloudFormation.svg',
  'aws.glacier':             'iconify:logos:aws-glacier',
  'aws.glue':                'cdn:AWS/svg/Analytics/Glue.svg',
  'aws.kms':                 'cdn:AWS/svg/Security-Identity-Compliance/Key-Management-Service.svg',
  'aws.lightsail':           'cdn:AWS/svg/Compute/Lightsail.svg',
  'aws.ses':                 'cdn:AWS/svg/Business-Applications/Simple-Email-Service.svg',
  'aws.xray':                'cdn:AWS/svg/Developer-Tools/X-Ray.svg',
  'aws.documentdb':          'cdn:AWS/svg/Database/DocumentDB.svg',
  'aws.neptune':             'cdn:AWS/svg/Database/Neptune.svg',
  'aws.mq':                  'cdn:AWS/svg/App-Integration/MQ.svg',
  'aws.certificate_manager': 'cdn:AWS/svg/Security-Identity-Compliance/Certificate-Manager.svg',
  'aws.quicksight':          'cdn:AWS/svg/Analytics/QuickSight.svg',
  'aws.codecommit':          'cdn:AWS/svg/Developer-Tools/CodeCommit.svg',
  'aws.codedeploy':          'cdn:AWS/svg/Developer-Tools/CodeDeploy.svg',
  'aws.api_gateway_rest':    'cdn:AWS/svg/App-Integration/API-Gateway.svg',
  'aws.api_gateway_http':    'cdn:AWS/svg/App-Integration/API-Gateway.svg',
  'aws.api_gateway_ws':      'cdn:AWS/svg/App-Integration/API-Gateway.svg',
  // GCP long tail — official vendor SVGs via icepanel CDN.
  'gcp.pubsub':              'cdn:GCP/svg/PubSub.svg',
  'gcp.cloud_tasks':         'cdn:GCP/svg/Cloud-Tasks.svg',
  'gcp.cloud_scheduler':     'cdn:GCP/svg/Cloud-Scheduler.svg',
  'gcp.dataproc':            'cdn:GCP/svg/Dataproc.svg',
  'gcp.composer':            'cdn:GCP/svg/Cloud-Composer.svg',
  'gcp.iam':                 'cdn:GCP/svg/Identity-And-Access-Management.svg',
  'gcp.cloud_endpoints':     'cdn:GCP/svg/Cloud-Endpoints.svg',
  'gcp.logging':             'cdn:GCP/svg/Cloud-Logging.svg',
  'gcp.monitoring':          'cdn:GCP/svg/Cloud-Monitoring.svg',
  // Azure / k8s family marks are vendored from the Iconify logos set.
  'azure.event_hubs':        'iconify:logos:microsoft-azure',
  'azure.service_bus':       'iconify:logos:microsoft-azure',
  'azure.data_lake':         'iconify:logos:microsoft-azure',
  'azure.synapse':           'iconify:logos:microsoft-azure',
  'azure.databricks':        'iconify:logos:microsoft-azure',
  'azure.ad':                'iconify:logos:microsoft-azure',
  'azure.monitor':           'iconify:logos:microsoft-azure',
  'k8s.replicaset':          'iconify:logos:kubernetes',
  'k8s.cronjob':             'iconify:logos:kubernetes',
  'k8s.role':                'iconify:logos:kubernetes',
  'k8s.rolebinding':         'iconify:logos:kubernetes',
  'k8s.serviceaccount':      'iconify:logos:kubernetes',
  'k8s.networkpolicy':       'iconify:logos:kubernetes',
};

function brandIcon(canonical: string, lucideName: string): string {
  const brand = BRAND_OVERRIDES[canonical];
  if (!brand) return lucideName;
  // Values already carry their dispatch prefix; pass through.
  if (brand.startsWith('cdn:') || brand.startsWith('iconify:') || brand.startsWith('lucide:')) return brand;
  // Back-compat: bare slug → iconify logos: pack.
  return `iconify:logos:${brand}`;
}

const GENERATED_ICONS: Record<string, IconDef> = (() => {
  const out: Record<string, IconDef> = {};
  for (const c of COMPONENTS) {
    out[c.canonical] = {
      label: c.display,
      color: colorFor(c.provider),
      icon: brandIcon(c.canonical, inferLucideIcon(c.canonical, c.subcategory)),
    };
  }
  return out;
})();

export function getIconDef(type: string): IconDef {
  // Coerce input to the canonical `family.component` form before lookup.
  // Handles model quirks (Opus 4.8 emits `aws:route53` with a colon) AND
  // old diagrams in IndexedDB that were saved with pre-normalization
  // types. Without this every colon-form entry hit the smart fallback
  // and rendered as a hollow circle with 'aws:route53' as the label.
  const canonical = canonicalizeType(type);

  const direct = ICON_REGISTRY[canonical];
  if (direct) return direct;
  const generated = GENERATED_ICONS[canonical];
  if (generated) return generated;

  // Smart fallback: use family color and derive label from type
  const dotIdx = canonical.indexOf('.');
  if (dotIdx !== -1) {
    const family = canonical.slice(0, dotIdx);
    const component = canonical.slice(dotIdx + 1);
    const color =
      FAMILY_COLORS[family as TypeFamily] ??
      EXTENDED_FAMILY_COLORS[family] ??
      (family === 'net' ? FAMILY_COLORS.network : undefined) ??
      (family === 'ent' ? FAMILY_COLORS.enterprise : undefined) ??
      FAMILY_COLORS.generic;
    const label = component.split('_').map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
    const icon = FAMILY_FALLBACK_ICON[family] ?? 'circle';
    return { label, color, icon };
  }

  return { label: type || 'Unknown', color: FAMILY_COLORS.generic, icon: 'circle' };
}

// canonicalizeType mirrors the backend's normalizeType shape: lowercase,
// colon→dot, hyphen→underscore, whitespace→underscore. Keeps the client
// tolerant of any type separator the model or a legacy save produces.
function canonicalizeType(raw: string): string {
  const normalized = raw
    .toLowerCase()
    .trim()
    .replaceAll(':', '.')
    .replaceAll('-', '_')
    .replaceAll(/\s+/g, '_');
  return LEGACY_CANONICAL_ALIASES[normalized] ?? normalized;
}

const LEGACY_CANONICAL_ALIASES: Record<string, string> = {
  'aws.route53': 'aws.route_53',
  'gcp.pub_sub': 'gcp.pubsub',
  'gcp.cloud_pubsub': 'gcp.pubsub',
  'azure.redis_cache': 'azure.redis',
};

export function getFamilyColor(type: string): string {
  const canonical = canonicalizeType(type);
  const family = canonical.split('.')[0] as TypeFamily;
  return FAMILY_COLORS[family] ?? FAMILY_COLORS.generic;
}

export const getPrefixColor = getFamilyColor;

// `net` and `ent` are aliases for the existing `network` / `enterprise`
// palette families — merge at display time so the sidebar doesn't split
// similar things into two columns.
const DISPLAY_FAMILY_MAP: Record<string, string> = {
  net: 'network',
  ent: 'enterprise',
};

export function getServicesByPrefix(): Record<string, Array<{ type: string; icon: IconDef }>> {
  const grouped: Record<string, Array<{ type: string; icon: IconDef }>> = {};
  const seen = new Set<string>();

  const push = (type: string, icon: IconDef) => {
    const prefix = type.split('.')[0]!;
    const family = DISPLAY_FAMILY_MAP[prefix] ?? prefix;
    (grouped[family] ??= []).push({ type, icon });
  };

  // Hand-curated first — takes precedence for icon + label.
  for (const [type, icon] of Object.entries(ICON_REGISTRY)) {
    push(type, icon);
    seen.add(type);
  }
  // Generated fills the long tail.
  for (const [type, icon] of Object.entries(GENERATED_ICONS)) {
    if (seen.has(type)) continue;
    push(type, icon);
  }
  // Alphabetical within each family for predictable palette UX.
  for (const list of Object.values(grouped)) {
    list.sort((a, b) => a.type.localeCompare(b.type));
  }
  return grouped;
}
