// Monthly cost estimates in USD for common cloud SKUs.
//
// These are BALLPARK numbers, not spec-accurate billing. They assume
// "small production" defaults (730h/mo, modest storage, typical traffic)
// so architects can see order-of-magnitude cost as they design — not a
// finance-grade invoice forecast. Pricing fluctuates; revisit quarterly.
//
// Keyed by the same node-type strings the icon library uses (aws.ec2,
// gcp.cloud_sql, azure.vm, …). Unknown node types get $0 with a "no
// estimate" flag so the UI can distinguish "free" from "unpriced".
//
// When you do a refresh, update LAST_REVIEWED to today's date in
// YYYY-MM-DD form. `npm run cost:refresh` reads it to tell you whether
// the catalog is stale (>90 days). Don't bump the date without actually
// spot-checking — the freshness signal is only useful if it's honest.

export const LAST_REVIEWED = '2026-09-16'; // YYYY-MM-DD, UTC
export const REFRESH_INTERVAL_DAYS = 90;

export interface CostEntry {
  /** USD / month. Rounded to whole dollars — precision here is false comfort. */
  usd: number;
  /** One-line assumption behind the number. Shown in the tooltip. */
  basis: string;
  /** Coarse family for breakdown grouping. */
  family: 'compute' | 'storage' | 'database' | 'network' | 'security' | 'ops' | 'ai' | 'messaging' | 'analytics' | 'cdn';
}

const TABLE: Record<string, CostEntry> = {
  // ── AWS ────────────────────────────────────────────────────────
  'aws.ec2':               { usd: 30,  basis: 't3.small, 730h/mo',               family: 'compute' },
  'aws.ecs':               { usd: 50,  basis: '2×small Fargate tasks',            family: 'compute' },
  'aws.eks':               { usd: 75,  basis: 'cluster fee + 2 nodes',            family: 'compute' },
  'aws.fargate':           { usd: 50,  basis: '2×small tasks continuous',         family: 'compute' },
  'aws.lambda':            { usd: 5,   basis: '1M invocations, 200ms avg',        family: 'compute' },
  'aws.app_runner':        { usd: 25,  basis: '0.25 vCPU, always-on',             family: 'compute' },
  'aws.elastic_beanstalk': { usd: 30,  basis: 'wraps t3.small',                   family: 'compute' },

  'aws.alb':               { usd: 20,  basis: 'ALB, low traffic',                 family: 'network' },
  'aws.nlb':               { usd: 20,  basis: 'NLB, low traffic',                 family: 'network' },
  'aws.elb':               { usd: 18,  basis: 'classic LB',                       family: 'network' },
  'aws.api_gateway':       { usd: 10,  basis: '1M REST calls',                    family: 'network' },
  'aws.nat_gateway':       { usd: 35,  basis: '1 AZ, minimal data',               family: 'network' },
  'aws.internet_gateway':  { usd: 0,   basis: 'free',                             family: 'network' },
  'aws.transit_gateway':   { usd: 40,  basis: '1 attachment',                     family: 'network' },
  'aws.direct_connect':    { usd: 200, basis: '1 Gbps port, shared',              family: 'network' },
  'aws.vpc':               { usd: 0,   basis: 'free',                             family: 'network' },
  'aws.subnet':            { usd: 0,   basis: 'free',                             family: 'network' },
  'aws.route_53':          { usd: 1,   basis: '1 hosted zone',                    family: 'network' },
  'aws.cloudfront':        { usd: 15,  basis: '100GB egress',                     family: 'cdn' },

  'aws.s3':                { usd: 5,   basis: '100GB standard + mild requests',   family: 'storage' },
  'aws.efs':               { usd: 15,  basis: '50GB standard',                    family: 'storage' },
  'aws.backup':            { usd: 10,  basis: '100GB incremental',                family: 'storage' },
  'aws.ecr':               { usd: 1,   basis: '10GB images',                      family: 'storage' },

  'aws.rds':               { usd: 50,  basis: 'db.t3.small, Single-AZ',           family: 'database' },
  'aws.aurora':            { usd: 90,  basis: '2× db.r6g.large writer+reader',    family: 'database' },
  'aws.dynamodb':          { usd: 15,  basis: '25 RCU/WCU on-demand',             family: 'database' },
  'aws.elasticache':       { usd: 50,  basis: 'cache.t3.small Redis',             family: 'database' },
  'aws.redshift':          { usd: 180, basis: 'dc2.large single node',            family: 'analytics' },
  'aws.opensearch':        { usd: 90,  basis: 't3.small.search',                  family: 'analytics' },

  'aws.sqs':               { usd: 1,   basis: '1M messages',                      family: 'messaging' },
  'aws.sns':               { usd: 1,   basis: '1M notifications',                 family: 'messaging' },
  'aws.eventbridge':       { usd: 1,   basis: '1M events',                        family: 'messaging' },
  'aws.kinesis':           { usd: 15,  basis: '1 shard',                          family: 'messaging' },
  'aws.msk':               { usd: 220, basis: 'kafka.t3.small 3-broker cluster',  family: 'messaging' },
  'aws.step_functions':    { usd: 3,   basis: '10k state transitions',            family: 'messaging' },

  'aws.cloudwatch':        { usd: 10,  basis: 'logs + custom metrics',            family: 'ops' },
  'aws.cloudtrail':        { usd: 2,   basis: 'management events only',           family: 'ops' },
  'aws.config':            { usd: 5,   basis: 'configuration items',              family: 'ops' },
  'aws.systems_manager':   { usd: 0,   basis: 'free at small scale',              family: 'ops' },

  'aws.iam':               { usd: 0,   basis: 'free',                             family: 'security' },
  'aws.cognito':           { usd: 0,   basis: 'free below 50k MAU',               family: 'security' },
  'aws.secrets_manager':   { usd: 2,   basis: '5 secrets',                        family: 'security' },
  'aws.waf':               { usd: 10,  basis: '1 web ACL',                        family: 'security' },
  'aws.shield':            { usd: 0,   basis: 'Standard is free',                 family: 'security' },
  'aws.guardduty':         { usd: 15,  basis: 'single-account small org',         family: 'security' },

  'aws.sagemaker':         { usd: 70,  basis: 'ml.t3.medium notebook',            family: 'ai' },
  'aws.bedrock':           { usd: 30,  basis: 'on-demand, moderate usage',        family: 'ai' },

  // ── GCP ────────────────────────────────────────────────────────
  'gcp.compute_engine':    { usd: 25,  basis: 'e2-small, 730h',                   family: 'compute' },
  'gcp.gke':               { usd: 75,  basis: 'cluster + 2 nodes',                family: 'compute' },
  'gcp.cloud_run':         { usd: 10,  basis: '1M requests, 256MB',               family: 'compute' },
  'gcp.cloud_functions':   { usd: 5,   basis: '1M invocations',                   family: 'compute' },
  'gcp.app_engine':        { usd: 30,  basis: 'F2 instance, always-on',           family: 'compute' },

  'gcp.vpc':               { usd: 0,   basis: 'free',                             family: 'network' },
  'gcp.load_balancer':     { usd: 20,  basis: 'HTTP(S) LB',                       family: 'network' },
  'gcp.cloud_cdn':         { usd: 15,  basis: '100GB cache egress',               family: 'cdn' },
  'gcp.cloud_dns':         { usd: 1,   basis: '1 zone',                           family: 'network' },

  'gcp.cloud_storage':     { usd: 5,   basis: '100GB standard',                   family: 'storage' },
  'gcp.cloud_sql':         { usd: 50,  basis: 'db-custom-1-3840, HA off',         family: 'database' },
  'gcp.bigquery':          { usd: 30,  basis: '100GB storage + 1TB query',        family: 'analytics' },
  'gcp.firestore':         { usd: 5,   basis: 'small collection, free tier+',    family: 'database' },
  'gcp.spanner':           { usd: 260, basis: '1 node regional',                  family: 'database' },
  'gcp.memorystore':       { usd: 40,  basis: 'M1 standard, 1GB',                 family: 'database' },

  'gcp.pubsub':            { usd: 2,   basis: '1M messages',                      family: 'messaging' },

  'gcp.cloud_logging':     { usd: 5,   basis: 'first 50GiB is free',              family: 'ops' },
  'gcp.cloud_monitoring':  { usd: 5,   basis: 'first 150MB free',                 family: 'ops' },
  'gcp.iam':               { usd: 0,   basis: 'free',                             family: 'security' },
  'gcp.secret_manager':    { usd: 1,   basis: '10 secrets',                       family: 'security' },

  'gcp.vertex_ai':         { usd: 60,  basis: 'small endpoint',                   family: 'ai' },

  // ── Azure ──────────────────────────────────────────────────────
  'azure.vm':              { usd: 30,  basis: 'B2s, 730h',                        family: 'compute' },
  'azure.aks':             { usd: 75,  basis: 'cluster + 2 nodes',                family: 'compute' },
  'azure.functions':       { usd: 5,   basis: '1M executions',                    family: 'compute' },
  'azure.app_service':     { usd: 55,  basis: 'B1 plan',                          family: 'compute' },
  'azure.container_apps':  { usd: 20,  basis: '2 replicas',                       family: 'compute' },

  'azure.vnet':            { usd: 0,   basis: 'free',                             family: 'network' },
  'azure.load_balancer':   { usd: 20,  basis: 'Standard',                         family: 'network' },
  'azure.application_gateway':{ usd: 35, basis: 'v2 small',                       family: 'network' },
  'azure.cdn':             { usd: 15,  basis: '100GB egress',                     family: 'cdn' },
  'azure.dns':             { usd: 1,   basis: '1 zone',                           family: 'network' },

  'azure.blob_storage':    { usd: 5,   basis: '100GB hot',                        family: 'storage' },
  'azure.sql':             { usd: 50,  basis: 'Basic 5 DTU',                      family: 'database' },
  'azure.cosmos_db':       { usd: 30,  basis: '400 RU/s provisioned',             family: 'database' },
  'azure.redis':           { usd: 50,  basis: 'Basic C1',                         family: 'database' },

  'azure.service_bus':     { usd: 10,  basis: 'Standard',                         family: 'messaging' },
  'azure.event_hubs':      { usd: 15,  basis: 'Standard 1 TU',                    family: 'messaging' },

  'azure.monitor':         { usd: 10,  basis: 'logs + metrics',                   family: 'ops' },
  'azure.key_vault':       { usd: 1,   basis: 'standard, small usage',            family: 'security' },
  'azure.ad':              { usd: 0,   basis: 'free tier',                        family: 'security' },

  // ── Beginner-tier fill (Aug 2026) ──────────────────────────────
  // Adds the 70 canonical entries that had no estimate. These are the
  // most-hit services in generated diagrams; without them the cost
  // panel showed "—" on obvious things (S3, security groups, KMS)
  // which made the whole feature feel half-built.

  // AWS — beginner
  'aws.athena':            { usd: 5,   basis: '~500 queries, 5GB scan each',       family: 'analytics' },
  'aws.budgets':           { usd: 0,   basis: 'free',                              family: 'ops' },
  'aws.certificate_manager':{ usd: 0,  basis: 'public certs free',                 family: 'security' },
  'aws.cloudshell':        { usd: 0,   basis: 'free w/ 1GB storage',               family: 'ops' },
  'aws.cloudwatch_alarms': { usd: 1,   basis: '10 alarms',                         family: 'ops' },
  'aws.cloudwatch_logs':   { usd: 5,   basis: '5GB ingested, 30d retention',       family: 'ops' },
  'aws.cloudwatch_metrics':{ usd: 3,   basis: '10 custom metrics',                 family: 'ops' },
  'aws.cost_explorer':     { usd: 0,   basis: 'free view; APIs metered',           family: 'ops' },
  'aws.ebs':               { usd: 8,   basis: '100GB gp3',                         family: 'storage' },
  'aws.elastic_ip':        { usd: 0,   basis: 'free while attached',               family: 'network' },
  'aws.elasticache_redis': { usd: 50,  basis: 'cache.t3.small Redis',              family: 'database' },
  'aws.iot_core':          { usd: 1,   basis: '1M messages',                       family: 'messaging' },
  'aws.kms':               { usd: 1,   basis: '$1/key/mo baseline',                family: 'security' },
  'aws.lightsail':         { usd: 10,  basis: 'small instance bundle',             family: 'compute' },
  'aws.marketplace':       { usd: 0,   basis: 'catalog only; product varies',      family: 'ops' },
  'aws.polly':             { usd: 5,   basis: '1M characters neural',              family: 'ai' },
  'aws.sagemaker_canvas':  { usd: 40,  basis: 'session hours, small model',        family: 'ai' },
  'aws.security_group':    { usd: 0,   basis: 'free',                              family: 'security' },
  'aws.translate':         { usd: 5,   basis: '1M characters',                     family: 'ai' },

  // Azure — beginner
  'azure.advisor':         { usd: 0,   basis: 'free',                              family: 'ops' },
  'azure.arm':             { usd: 0,   basis: 'orchestrator; free',                family: 'ops' },
  'azure.blob':            { usd: 5,   basis: '100GB hot (alias)',                 family: 'storage' },
  'azure.cost_management': { usd: 0,   basis: 'free',                              family: 'ops' },
  'azure.entra_id':        { usd: 0,   basis: 'free tier; P1 $6/user',             family: 'security' },
  'azure.files':           { usd: 10,  basis: '100GB standard',                    family: 'storage' },
  'azure.iot_hub':         { usd: 10,  basis: 'basic tier',                        family: 'messaging' },
  'azure.managed_disk':    { usd: 8,   basis: '100GB Premium SSD',                 family: 'storage' },
  'azure.nsg':             { usd: 0,   basis: 'free',                              family: 'security' },
  'azure.openai':          { usd: 15,  basis: 'GPT-4o mini, ~2M tokens',           family: 'ai' },
  'azure.power_bi':        { usd: 10,  basis: 'Pro $10/user',                      family: 'analytics' },
  'azure.public_ip':       { usd: 4,   basis: 'Standard static',                   family: 'network' },
  'azure.sql_database':    { usd: 50,  basis: 'Basic 5 DTU (alias)',               family: 'database' },
  'azure.static_web_apps': { usd: 0,   basis: 'free tier',                         family: 'compute' },
  'azure.storage_account': { usd: 5,   basis: '100GB blob hot',                    family: 'storage' },
  'azure.subnet':          { usd: 0,   basis: 'free',                              family: 'network' },
  'azure.translator':      { usd: 10,  basis: 'S1 tier, 2M chars',                 family: 'ai' },

  // GCP — beginner
  'gcp.cloud_billing':     { usd: 0,   basis: 'free',                              family: 'ops' },
  'gcp.cloud_shell':       { usd: 0,   basis: 'free w/ 5GB',                       family: 'ops' },
  'gcp.firebase':          { usd: 0,   basis: 'Spark plan (free tier)',            family: 'compute' },
  'gcp.kms':               { usd: 1,   basis: '$0.06/key/mo baseline',             family: 'security' },
  'gcp.memorystore_redis': { usd: 45,  basis: '1GB basic tier',                    family: 'database' },
  'gcp.persistent_disk':   { usd: 10,  basis: '100GB pd-balanced',                 family: 'storage' },

  // Kubernetes primitives — cluster carries the real cost (priced
  // separately as k8s.cluster / gke / aks / eks). These are free
  // abstractions inside a cluster.
  'k8s.configmap':         { usd: 0,   basis: 'free',                              family: 'ops' },
  'k8s.coredns':           { usd: 0,   basis: 'free (in-cluster)',                 family: 'network' },
  'k8s.dashboard':         { usd: 0,   basis: 'free (in-cluster)',                 family: 'ops' },
  'k8s.deployment':        { usd: 0,   basis: 'free (pods carry cost)',            family: 'compute' },
  'k8s.helm':              { usd: 0,   basis: 'free',                              family: 'ops' },
  'k8s.ingress':           { usd: 0,   basis: 'free (LB carries cost)',            family: 'network' },
  'k8s.pod':               { usd: 0,   basis: 'free (node carries cost)',          family: 'compute' },
  'k8s.secret':            { usd: 0,   basis: 'free',                              family: 'security' },
  'k8s.service_clusterip': { usd: 0,   basis: 'free (internal)',                   family: 'network' },
  'k8s.service_loadbalancer':{ usd: 20,basis: 'cloud LB per service',              family: 'network' },

  // Enterprise SaaS — external tools. Per-seat approximations for
  // small teams; scaling is linear in headcount, so this understates
  // for a 200-person co (also over-estimates for 3-person indie).
  // Better to show *something* than nothing.
  'ent.active_directory':  { usd: 0,   basis: 'baked into Windows Server',         family: 'security' },
  'ent.github':            { usd: 4,   basis: 'Team plan, per user',               family: 'ops' },
  'ent.google_workspace':  { usd: 6,   basis: 'Starter, per user',                 family: 'ops' },
  'ent.microsoft_365':     { usd: 13,  basis: 'Business Basic, per user',          family: 'ops' },
  'ent.microsoft_teams':   { usd: 4,   basis: 'Essentials, per user',              family: 'ops' },
  'ent.paypal':            { usd: 0,   basis: 'transaction fees, not fixed',       family: 'ops' },
  'ent.salesforce':        { usd: 25,  basis: 'Starter, per user',                 family: 'ops' },
  'ent.slack':             { usd: 7,   basis: 'Pro, per user',                     family: 'ops' },
  'ent.zoom':              { usd: 15,  basis: 'Pro, per user',                     family: 'ops' },

  // Data + protocol primitives — abstractions, not billed services.
  'data.parquet':          { usd: 0,   basis: 'file format',                       family: 'storage' },
  'iot.http2':             { usd: 0,   basis: 'protocol',                          family: 'network' },
  'net.dns_a_record':      { usd: 0,   basis: 'DNS-provider dependent',            family: 'network' },
  'net.dns_aaaa_record':   { usd: 0,   basis: 'DNS-provider dependent',            family: 'network' },
  'net.dns_cname':         { usd: 0,   basis: 'DNS-provider dependent',            family: 'network' },
  'net.dns_mx':            { usd: 0,   basis: 'DNS-provider dependent',            family: 'network' },
  'net.lb_round_robin':    { usd: 0,   basis: 'routing pattern',                   family: 'network' },
  'net.rfc1918':           { usd: 0,   basis: 'private CIDR concept',              family: 'network' },

  // ── Intermediate-tier fill (Sep 2026) ──────────────────────────
  // Popular level-I canonicals that showed up bare on real diagrams.
  // 36 entries chosen from the ~333 intermediate slots — the ones a
  // customer is most likely to drop onto a canvas. Rest of the tier
  // stays open on purpose; ballparks for niche services do more harm
  // than good.

  // AWS — intermediate
  'aws.aurora_serverless': { usd: 45,  basis: '0.5 ACU baseline, sporadic use',    family: 'database' },
  'aws.glue':              { usd: 20,  basis: '10 DPU-hours/mo',                   family: 'analytics' },
  'aws.batch':             { usd: 30,  basis: '~50 job-hours on Fargate',          family: 'compute' },
  'aws.documentdb':        { usd: 200, basis: 'db.t3.medium single-instance',      family: 'database' },
  'aws.codebuild':         { usd: 5,   basis: '~100 build-min general1.small',     family: 'ops' },
  'aws.codepipeline':      { usd: 1,   basis: '1 active pipeline',                 family: 'ops' },
  'aws.codedeploy':        { usd: 0,   basis: 'free for EC2/Lambda',               family: 'ops' },
  'aws.codecommit':        { usd: 0,   basis: 'first 5 users free',                family: 'ops' },
  'aws.codeartifact':      { usd: 2,   basis: '10GB storage + light egress',       family: 'storage' },
  'aws.xray':              { usd: 5,   basis: '1M traces recorded',                family: 'ops' },
  'aws.inspector':         { usd: 10,  basis: 'small EC2/ECR fleet',               family: 'security' },
  'aws.ssm_session_manager':{ usd: 0,  basis: 'free',                              family: 'ops' },
  'aws.organizations':     { usd: 0,   basis: 'free',                              family: 'ops' },
  'aws.trusted_advisor':   { usd: 0,   basis: 'core checks free',                  family: 'ops' },
  'aws.service_catalog':   { usd: 0,   basis: 'free at small scale',               family: 'ops' },

  // Azure — intermediate
  'azure.data_factory':    { usd: 30,  basis: '~50 pipeline runs, low DIU',        family: 'analytics' },
  'azure.logic_apps':      { usd: 10,  basis: 'Consumption, 100k actions',         family: 'messaging' },
  'azure.app_configuration':{ usd: 2,  basis: 'Standard, small store',             family: 'ops' },
  'azure.application_insights':{ usd: 15,basis: '5GB data/mo',                     family: 'ops' },
  'azure.log_analytics':   { usd: 10,  basis: '5GB ingested pay-as-you-go',        family: 'ops' },
  'azure.ai_search':       { usd: 75,  basis: 'Basic tier, 15GB',                  family: 'ai' },

  // GCP — intermediate
  'gcp.composer':          { usd: 350, basis: 'small env (Airflow, always-on)',    family: 'ops' },
  'gcp.cloud_nat':         { usd: 35,  basis: '1 gateway, minimal data',           family: 'network' },
  'gcp.artifact_registry': { usd: 2,   basis: '10GB images',                       family: 'storage' },
  'gcp.cloud_build':       { usd: 5,   basis: '~100 build-min standard',           family: 'ops' },
  'gcp.cloud_deploy':      { usd: 10,  basis: 'few pipelines, monthly runs',       family: 'ops' },
  'gcp.error_reporting':   { usd: 0,   basis: 'free',                              family: 'ops' },

  // Kubernetes — intermediate primitives. Free abstractions inside a
  // cluster; the workload (pods on nodes) is what actually costs.
  'k8s.statefulset':       { usd: 0,   basis: 'free (pods carry cost)',            family: 'compute' },
  'k8s.daemonset':         { usd: 0,   basis: 'free (pods carry cost)',            family: 'compute' },
  'k8s.hpa':               { usd: 0,   basis: 'free (autoscaler)',                 family: 'ops' },
  'k8s.pvc':               { usd: 0,   basis: 'free (PV carries storage cost)',    family: 'storage' },
  'k8s.pv':                { usd: 10,  basis: '~100GB backing disk',               family: 'storage' },
  'k8s.job':               { usd: 0,   basis: 'free (pods carry cost)',            family: 'compute' },
  'k8s.cronjob':           { usd: 0,   basis: 'free (pods carry cost)',            family: 'compute' },

  // Data formats — file formats, not billed.
  'data.avro':             { usd: 0,   basis: 'file format',                       family: 'storage' },
  'data.orc':              { usd: 0,   basis: 'file format',                       family: 'storage' },
};

export interface NodeCost {
  nodeId: string;
  nodeLabel: string;
  nodeType: string;
  estimate: CostEntry | null;
}

export interface CostReport {
  total: number;
  currency: 'USD';
  priced: NodeCost[];
  unpriced: NodeCost[];
  byFamily: Array<{ family: CostEntry['family']; total: number }>;
}

export function estimateFor(nodeType: string): CostEntry | null {
  return TABLE[nodeType] ?? null;
}

export function summarize(
  nodes: Array<{ id: string; type: string; label: string }>,
): CostReport {
  const priced: NodeCost[] = [];
  const unpriced: NodeCost[] = [];
  const familyTotals = new Map<CostEntry['family'], number>();

  for (const n of nodes) {
    const entry = estimateFor(n.type);
    const nc: NodeCost = { nodeId: n.id, nodeLabel: n.label, nodeType: n.type, estimate: entry };
    if (entry) {
      priced.push(nc);
      familyTotals.set(entry.family, (familyTotals.get(entry.family) ?? 0) + entry.usd);
    } else {
      unpriced.push(nc);
    }
  }

  const total = priced.reduce((sum, p) => sum + (p.estimate?.usd ?? 0), 0);
  const byFamily = [...familyTotals.entries()]
    .map(([family, total]) => ({ family, total }))
    .sort((a, b) => b.total - a.total);

  return { total, currency: 'USD', priced, unpriced, byFamily };
}
