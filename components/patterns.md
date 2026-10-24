# Common Architecture Patterns

Named compositions that the prompt normalizer expands into multi-component
skeletons. Each pattern lists **anchor** components (required) and **optional**
components (added when prompt suggests them).

Pattern-matching rule: if the user prompt contains any pattern's **trigger**
phrases AND does not enumerate components manually, expand to the anchor set.
Trigger phrases are trigram-fuzzy.

## 3-Tier Web App (cloud-agnostic)

**Triggers**: "3-tier", "3 tier web app", "web app with database", "classic web app"
**Anchors (AWS)**: `aws.cloudfront` → `aws.alb` → `aws.ecs` + `aws.fargate` → `aws.rds`
**Anchors (Azure)**: `azure.front_door_std` → `azure.application_gateway` → `azure.container_apps` → `azure.sql_database`
**Anchors (GCP)**: `gcp.cloud_cdn` → `gcp.load_balancer_global_external` → `gcp.cloud_run` → `gcp.cloud_sql`
**Optional**: WAF, Secrets Manager / Key Vault / Secret Manager, CloudWatch / Monitor / Cloud Monitoring, backup service, CDN

## Event-Driven Microservices

**Triggers**: "event-driven", "event driven microservices", "pub/sub", "message bus architecture"
**Anchors (AWS)**: `aws.eventbridge` / `aws.msk` + several `aws.lambda` + `aws.dynamodb` + `aws.sqs` DLQ
**Anchors (Azure)**: `azure.event_grid` or `azure.event_hubs` + `azure.functions` + `azure.cosmos_nosql` + `azure.service_bus` DLQ
**Anchors (GCP)**: `gcp.pubsub` + `gcp.cloud_run` + `gcp.firestore` + `gcp.pubsub` DLQ
**Optional**: Schema registry, CDC source (DB Streams), distributed tracing

## Streaming Analytics Pipeline

**Triggers**: "streaming analytics", "clickstream", "real-time pipeline", "kafka pipeline"
**Anchors**: `data.kafka_broker` (or `aws.msk`, `azure.event_hubs`, `gcp.pubsub`) → `data.flink_job_manager` (or `data.spark_structured_streaming` / `aws.kinesis_analytics` / `azure.stream_analytics`) → data lake + warehouse
**Lake layer**: `aws.s3` / `azure.datalake_gen2` / `gcp.cloud_storage` + table format (`data.iceberg` / `data.delta_lake` / `data.hudi`)
**Warehouse**: `aws.redshift` / `azure.synapse` or `azure.fabric_warehouse` / `gcp.bigquery`
**Orchestration**: `data.airflow` / `gcp.composer` / `data.dagster`

## Lakehouse / Modern Data Stack

**Triggers**: "lakehouse", "data lakehouse", "bronze silver gold", "medallion architecture"
**Anchors**: ingest (`data.fivetran`/`data.airbyte`) → object store + table format → `data.databricks` or `azure.databricks` with Unity Catalog → `data.dbt` transforms → `ent.looker` or `aws.quicksight` or `azure.power_bi` BI
**Cloud substrate**: any object store + catalog (Glue / Unity / Dataplex Catalog)

## Serverless SaaS Backend

**Triggers**: "serverless saas", "startup api", "mvp api", "serverless backend"
**Anchors (AWS)**: `aws.cloudfront` → `aws.api_gateway_http` → `aws.lambda` → `aws.dynamodb` + `aws.cognito_user_pools` + `aws.secrets_manager`
**Anchors (Azure)**: `azure.front_door_std` → `azure.apim` → `azure.functions` → `azure.cosmos_nosql` + `azure.entra_external`
**Anchors (GCP)**: `gcp.cloud_cdn` → `gcp.apigee` → `gcp.cloud_functions_gen2` or `gcp.cloud_run` → `gcp.firestore` + Firebase Auth

## Zero-Trust Internal Access

**Triggers**: "zero trust", "replace vpn", "ztna", "beyondcorp"
**Anchors**: IdP (`ent.okta` / `azure.entra_id` / `gcp.iam`) → ZTNA proxy (`aws.verified_access` / `gcp.iap` / `ent.zscaler_zpa` / `ent.hashicorp_boundary`) → internal app targets
**Optional**: device posture (MDM), SSE (`net.sse`), DLP

## ML / GenAI RAG Stack

**Triggers**: "rag stack", "retrieval-augmented", "rag app", "ai chatbot with knowledge"
**Anchors (AWS)**: data in `aws.s3` → `aws.bedrock_knowledge_base` (or `aws.opensearch_serverless` with embeddings) → `aws.bedrock` model + `aws.bedrock_agents` → API via `aws.api_gateway_http` → `aws.lambda`
**Anchors (Azure)**: `azure.blob` → `azure.ai_search` (vector) → `azure.openai` → `azure.ai_foundry` agent → App Service
**Anchors (GCP)**: `gcp.cloud_storage` → `gcp.vertex_matching_engine` → `gcp.vertex_ai` model (Gemini) → Cloud Run
**Observability**: LLM tracing, eval harness, guardrails (`aws.bedrock_guardrails` / `azure.content_safety`)

## Multi-Region Active-Active

**Triggers**: "multi-region active active", "global active-active", "geo-redundant"
**Anchors (AWS)**: `aws.route_53` latency routing + `aws.global_accelerator` + `aws.aurora` Global Database + cross-region `aws.s3` replication
**Anchors (Azure)**: `azure.traffic_manager` + `azure.front_door_std` + `azure.cosmos_nosql` multi-region writes + RA-GZRS storage
**Anchors (GCP)**: `gcp.load_balancer_global_external` + `gcp.spanner` multi-regional + `gcp.gcs_turbo_replication`

## Disaster Recovery (Pilot-Light / Warm / Hot)

**Triggers**: "dr", "disaster recovery", "pilot light", "warm standby"
**Anchors (AWS)**: `aws.drs` (continuous replication) + cross-region `aws.backup` vault + `aws.route_53_application_recovery`
**Anchors (Azure)**: `azure.site_recovery` + `azure.backup` immutable vault
**Anchors (GCP)**: `gcp.backup_dr` + multi-region `gcp.cloud_storage`

## Hybrid Cloud (3 tiers)

**Triggers**: "hybrid cloud", "on-prem to cloud", "cloud burst"
**Anchors (AWS)**: on-prem DC + `aws.direct_connect` + `aws.transit_gateway` + `aws.outposts` or `aws.eks_anywhere`
**Anchors (Azure)**: on-prem DC + `azure.expressroute` + `azure.virtual_wan` + `azure.arc`
**Anchors (GCP)**: on-prem DC + `gcp.cloud_interconnect` + `gcp.ncc` + `gcp.anthos_bare_metal`

## Data Mesh

**Triggers**: "data mesh", "domain-oriented data", "data products"
**Anchors**: per-domain lakehouse (Iceberg/Delta on GCS/ADLS/S3) + central catalog (`gcp.dataplex` / `azure.purview` / `aws.lake_formation`) + contract-driven transfer + federated query (`data.trino`, `gcp.bigquery_omni`)

## IoT Fleet / Digital Twin

**Triggers**: "iot fleet", "industrial iot", "connected devices", "digital twin"
**Anchors (AWS)**: devices → `aws.iot_core` → `aws.iot_sitewise` + `aws.iot_twinmaker` + `aws.timestream` + `aws.cloudwatch` + `aws.iot_greengrass` on the edge
**Anchors (Azure)**: devices → `azure.iot_hub` → `azure.digital_twins` + `azure.data_explorer` + `azure.iot_edge` on edge
**Anchors (GCP)**: devices → `gcp.pubsub` → `gcp.dataflow` → `gcp.bigtable` / `gcp.bigquery`

## AI Training on HPC

**Triggers**: "train llm", "pretraining", "foundation model", "distributed training"
**Anchors (AWS)**: `aws.sagemaker_hyperpod` + `aws.fsx_lustre` + `aws.ec2` (p5/trn1) clusters + `aws.parallelcluster`
**Anchors (Azure)**: `azure.hpc_hbv4` (or NDv5) + `azure.high_performance_fsx` (Managed Lustre) + `azure.cyclecloud` + `azure.ml_compute_cluster`
**Anchors (GCP)**: `gcp.gke` + TPU v5p Pod + `gcp.filestore` Enterprise + `gcp.vertex_ai` custom training

## CI/CD Golden Path

**Triggers**: "ci/cd", "devops pipeline", "golden path", "internal developer platform"
**Anchors**: `ent.github` or `ent.gitlab` → CI (`ent.github_actions` or cloud CI) → registry (`aws.ecr` / `azure.acr` / `gcp.artifact_registry`) → deploy (`aws.codedeploy` / `azure.deployment_env` / `gcp.cloud_deploy`) → runtime (ECS/AKS/GKE)
**Optional**: IaC (`aws.cdk`, `azure.bicep`, Terraform), policy (`k8s.kyverno`, Open Policy Agent)

## Observability Stack (OSS)

**Triggers**: "observability stack", "open source observability", "prometheus grafana"
**Anchors**: `k8s.prometheus` or `k8s.mimir` + `k8s.grafana` + `k8s.loki` + `k8s.tempo` + `k8s.open_telemetry` collector
**Optional**: `k8s.thanos` for long-term storage, `k8s.victoria_metrics` swap, `k8s.falco` runtime security

## Contact Center (Modern)

**Triggers**: "contact center", "call center", "ivr", "customer service platform"
**Anchors (AWS)**: `aws.connect` + `aws.lex` + `aws.polly` + `aws.transcribe` + `aws.bedrock` (agent assist) + `aws.kinesis_video` (for screen recording)
**Anchors (GCP)**: `gcp.contact_center_ai` + `gcp.vertex_ai` + Dialogflow

## Compliance Baseline (HIPAA / PCI)

**Triggers**: "hipaa", "pci", "compliance baseline", "regulated workload"
**Anchors**: encryption at rest (KMS/Key Vault), encryption in transit (TLS), WAF, private networking (no public IPs where avoidable), audit log (CloudTrail / Activity Log / Cloud Audit Logs), SIEM (`aws.security_hub` / `azure.sentinel` / `gcp.chronicle_security`), data classification (`aws.macie` / `azure.purview` / `gcp.dataplex`)

---

## Pattern expansion rules for the normalizer

1. Match pattern by **trigger set** fuzzily. Score = max trigram similarity of any trigger phrase to any 2–5-word window in the prompt.
2. Choose cloud from explicit mention ("aws", "amazon", "azure", "gcp"/"google cloud"); if multiple, require explicit enumeration and don't auto-expand.
3. Always ask the normalizer to **render anchors first**, optional only if the prompt hints at them or mentions operational concerns ("secure", "with logging", "backup").
4. If the prompt enumerates ≥ 4 specific components, **do not expand a pattern** — assume the architect already knows the shape.
5. After expansion, re-run the validator. Any high-severity rule firing (no WAF, no IAM, no observability) should drive additional optional components to be added silently, so patterns don't fail best practices out of the box.
