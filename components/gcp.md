# Google Cloud Platform — Component Catalog

Deep GCP catalog including Anthos, Vertex, Spanner internals, VPC Service
Controls, BeyondCorp, Distributed Cloud, Chronicle.

## Compute

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `gcp.compute_engine` | Compute Engine | vm | B | VMs across E2/N2/N2D/C3/C3D/M3/T2D/T2A families. | gce, compute engine, vm, google vm | N2D 16 vCPU |
| `gcp.preemptible_vm` | Preemptible VM | vm | I | 60–80% discount VMs; preempted within 24 h. | preemptible, preemptible vm | Batch workers |
| `gcp.spot_vm` | Spot VM | vm | I | New-gen preemptible — no 24 h cap, best-effort. | spot vm | Spot instead of preemptible |
| `gcp.sole_tenant` | Sole-Tenant Node | vm | A | Physical host dedicated to one project; BYOL-friendly. | sole tenant, sole-tenant node | Oracle on sole-tenant |
| `gcp.confidential_vm` | Confidential VM | security | A | AMD SEV / Intel TDX / SEV-SNP encrypted memory VMs. | confidential vm, cvm | CVM for health data |
| `gcp.gke` | GKE | container | B | Managed Kubernetes; Standard, Autopilot, Enterprise editions. | gke, kubernetes, google kubernetes engine | GKE Autopilot cluster |
| `gcp.gke_autopilot` | GKE Autopilot | container | I | Hands-off GKE: node mgmt hidden, pod-billed. | autopilot, gke autopilot | Autopilot for small teams |
| `gcp.gke_enterprise` | GKE Enterprise | container | A | Anthos-branded fleet: multi-cluster, service mesh, config mgmt. | anthos, gke enterprise | Enterprise K8s fleet |
| `gcp.cloud_run` | Cloud Run | serverless | B | Serverless containers; scale to zero; requests or jobs mode. | cloud run, run | Stateless API |
| `gcp.cloud_run_jobs` | Cloud Run Jobs | serverless | I | Batch containers; run to completion, retries. | cloud run jobs | Nightly data job |
| `gcp.cloud_functions_gen2` | Cloud Functions Gen2 | serverless | I | Functions on Cloud Run runtime — longer timeouts, concurrency. | cloud functions, gen2 | HTTP function |
| `gcp.app_engine_standard` | App Engine Standard | paas | I | Sandboxed runtimes (Python, Node, Java, Go, PHP, Ruby). | gae, app engine | Legacy GAE |
| `gcp.app_engine_flex` | App Engine Flex | paas | I | Docker-based GAE — more control, more cost. | app engine flex | Container on GAE Flex |
| `gcp.batch` | Batch | compute | I | Managed batch orchestrator (Slurm-style + GPU). | google batch | Protein folding batch |
| `gcp.anthos_bare_metal` | Anthos on Bare Metal | hybrid | A | Anthos on customer servers, no hypervisor. | anthos bare metal, abm | Bare-metal Anthos |
| `gcp.anthos_vmware` | Anthos on VMware | hybrid | A | Anthos on vSphere clusters. | anthos on vmware, avs | vSphere Anthos |
| `gcp.anthos_aws` | Anthos on AWS | hybrid | A | GKE control plane in GCP, EC2 data plane. | anthos on aws | Multi-cloud k8s |
| `gcp.anthos_azure` | Anthos on Azure | hybrid | A | GKE on Azure VMs. | anthos on azure | Cross-cloud k8s |

## Storage

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `gcp.cloud_storage` | Cloud Storage | object | B | GCS buckets — Standard / Nearline / Coldline / Archive. | gcs, cloud storage, bucket | Data lake on GCS |
| `gcp.gcs_autoclass` | GCS Autoclass | object | I | Auto-tier objects to cheapest storage class. | autoclass | Cost-optimized bucket |
| `gcp.gcs_turbo_replication` | GCS Turbo Replication | object | A | 15-min RPO dual-region replication SLA. | turbo replication | Financial bucket |
| `gcp.persistent_disk` | Persistent Disk | block | B | Zonal / regional; pd-standard / balanced / ssd / extreme. | pd, persistent disk | Boot + DB disks |
| `gcp.hyperdisk` | Hyperdisk | block | A | Next-gen block (Balanced/Extreme/ML/Throughput). | hyperdisk | Hyperdisk for SQL |
| `gcp.local_ssd` | Local SSD | block | I | NVMe attached to host, ephemeral. | local ssd | Ephemeral cache |
| `gcp.filestore` | Filestore | file | I | Managed NFSv3 — Basic/Zonal/Enterprise. | filestore, nfs | Shared NFS |
| `gcp.netapp_volumes` | NetApp Volumes | file | A | Managed ONTAP on GCP. | netapp volumes | ONTAP file server |
| `gcp.storage_transfer` | Storage Transfer Service | migration | I | Managed transfers from S3 / Azure / HTTP / on-prem to GCS. | sts, storage transfer | Cross-cloud copy |
| `gcp.transfer_appliance` | Transfer Appliance | migration | A | Shippable 40 TB / 300 TB appliance. | transfer appliance | Offline upload |
| `gcp.backup_dr` | Backup & DR | backup | A | Managed data protection (Actifio-based). | backup dr, actifio | VM + DB backups |

## Database

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `gcp.cloud_sql` | Cloud SQL | relational | B | MySQL / Postgres / SQL Server managed. | cloud sql, cloud sql postgres | Managed Postgres |
| `gcp.cloud_sql_enterprise_plus` | Cloud SQL Enterprise Plus | relational | A | Near-zero maintenance + data cache tier. | cloud sql enterprise plus | Low-downtime tier |
| `gcp.alloydb` | AlloyDB | relational | A | Postgres-compat with columnar engine — 4× OLTP, 100× analytics. | alloydb, alloy db | OLTP + HTAP |
| `gcp.alloydb_omni` | AlloyDB Omni | hybrid | A | AlloyDB runnable anywhere (on-prem, other cloud). | alloydb omni | On-prem AlloyDB |
| `gcp.spanner` | Spanner | relational | A | Horizontally-scalable strongly-consistent SQL, TrueTime. | spanner, google spanner | Global bank ledger |
| `gcp.firestore` | Firestore | nosql | B | Document DB (Native + Datastore modes); real-time sync. | firestore | Mobile sync |
| `gcp.bigtable` | Bigtable | nosql | A | HBase-compat wide-column at PB scale. | bigtable | Ad-tech events |
| `gcp.memorystore_redis` | Memorystore Redis | cache | B | Managed Redis / Redis Cluster. | memorystore, memorystore redis | Session cache |
| `gcp.memorystore_memcached` | Memorystore Memcached | cache | I | Managed Memcached. | memorystore memcached | Object cache |
| `gcp.memorystore_valkey` | Memorystore Valkey | cache | A | Managed Valkey (Redis OSS fork). | valkey | Open-source cache |
| `gcp.firebase_realtime_db` | Firebase RTDB | nosql | I | Real-time JSON tree. | firebase rtdb | Mobile real-time |

## Data / Analytics / ML

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `gcp.bigquery` | BigQuery | warehouse | B | Serverless MPP SQL; separates storage + compute; BI Engine. | bigquery, bq | Event analytics |
| `gcp.bigquery_omni` | BigQuery Omni | warehouse | A | Query Azure Blob / S3 from BigQuery with Anthos. | bigquery omni, bq omni | Multi-cloud analytics |
| `gcp.bigquery_ml` | BigQuery ML | ai | I | Train models with SQL; native embeddings, remote models to Vertex. | bq ml, bigquery ml | Churn model in SQL |
| `gcp.bigquery_studio` | BigQuery Studio | analytics | I | Integrated notebooks + SQL + DataFrames. | bq studio | Data-science notebook |
| `gcp.bigquery_continuous` | Continuous Queries | streaming | A | Streaming SQL over BigQuery tables. | continuous queries | Stream enrichment |
| `gcp.dataflow` | Dataflow | streaming | A | Serverless Apache Beam (stream + batch). | dataflow | ETL pipeline |
| `gcp.dataflow_prime` | Dataflow Prime | streaming | A | Vertical autoscaling + smart diagnostics. | dataflow prime | Auto-tuning pipelines |
| `gcp.dataproc` | Dataproc | analytics | A | Managed Spark / Hadoop / Hive / Flink / Presto clusters. | dataproc | Spark cluster |
| `gcp.dataproc_serverless` | Dataproc Serverless | analytics | A | Serverless Spark jobs. | dataproc serverless | PySpark job |
| `gcp.dataplex` | Dataplex | catalog | A | Data mesh: zones, discovery, governance over GCS + BigQuery. | dataplex | Data mesh |
| `gcp.dataplex_catalog` | Dataplex Catalog | catalog | A | Next-gen Data Catalog (mandatory 2025+). | data catalog, dataplex catalog | Table lineage |
| `gcp.datastream` | Datastream | cdc | A | Serverless CDC from Oracle / MySQL / Postgres → BigQuery / GCS. | datastream, cdc | Postgres → BQ CDC |
| `gcp.data_fusion` | Data Fusion | etl | A | Visual ETL on CDAP (open-source). | data fusion, cdap | No-code ETL |
| `gcp.pubsub` | Pub/Sub | messaging | B | Managed pub/sub w/ ordering + dead letters + schemas. | pubsub, pub/sub | Event bus |
| `gcp.pubsub_lite` | Pub/Sub Lite | messaging | A | Zonal, partitioned, cheaper than Pub/Sub. | pubsub lite | Cost-sensitive streaming |
| `gcp.analytics_hub` | Analytics Hub | analytics | A | Marketplace of BigQuery datasets inside + outside the org. | analytics hub | Share dataset externally |
| `gcp.composer` | Composer | orchestration | I | Managed Apache Airflow. | composer | DAG-based ETL |
| `gcp.dataform` | Dataform | etl | I | SQL-based ELT (similar to dbt) inside BQ Studio. | dataform | Modeled marts |
| `gcp.vertex_ai` | Vertex AI | ai | I | Unified ML — Pipelines, Workbench, Endpoints, Matching Engine, Gemini. | vertex, vertex ai, gemini | Foundation models |
| `gcp.vertex_matching_engine` | Vector Search | ai | A | ANN vector search (formerly Matching Engine). | matching engine, vector search | RAG vectors |
| `gcp.vertex_feature_store` | Vertex Feature Store | ai | A | Online + offline features for ML serving. | feature store | Real-time feats |
| `gcp.vertex_agent_builder` | Agent Builder | ai | A | Multi-agent orchestration (Agentspace). | agent builder, agentspace | Enterprise agent |
| `gcp.vertex_model_garden` | Model Garden | ai | I | Catalog of 130+ foundation models (Gemini, Llama, Claude via partners). | model garden | Claude via Vertex |
| `gcp.document_ai` | Document AI | ai | I | Doc parsing, OCR, form extraction, custom processors. | document ai | Invoice processor |
| `gcp.speech_to_text` | Speech-to-Text | ai | I | ASR with streaming + Chirp universal model. | stt, speech to text | Voicemail transcribe |
| `gcp.text_to_speech` | Text-to-Speech | ai | I | TTS w/ WaveNet + Studio voices. | tts, text to speech | IVR prompts |
| `gcp.translation` | Translation | ai | I | NMT v3; Document + Glossary translation. | translate | Localize product |
| `gcp.vision_api` | Vision API | ai | I | Label, face, OCR, logo, safe search. | vision api | Image moderation |
| `gcp.video_intelligence` | Video Intelligence | ai | A | Video labeling, shot change, object tracking. | video intelligence | Video labeling |
| `gcp.contact_center_ai` | Contact Center AI | ai | A | CCAI platform — agent assist + insights + Dialogflow. | ccai, contact center ai | Contact-center ai |
| `gcp.recaptcha_enterprise` | reCAPTCHA Enterprise | security | I | Bot detection, ML-based. | recaptcha, recaptcha enterprise | Fraud-form bot detect |
| `gcp.web_risk` | Web Risk | security | A | Google Safe Browsing as API. | web risk | Phishing URL check |

## Networking

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `gcp.vpc` | VPC | network | B | Global by default; auto-mode or custom; cross-region subnets in one VPC. | vpc, google vpc | Global VPC |
| `gcp.shared_vpc` | Shared VPC | network | A | Host project shares VPC subnets with service projects. | shared vpc, xpn | Platform team host VPC |
| `gcp.vpc_peering` | VPC Peering | network | I | Non-transitive peering between VPCs. | vpc peering | Peering to partner |
| `gcp.cloud_nat` | Cloud NAT | network | I | Regional egress NAT with port allocation and logging. | cloud nat, nat | Egress NAT |
| `gcp.cloud_router` | Cloud Router | network | A | BGP router for Interconnect / VPN / Cloud NAT. | cloud router | BGP peer for DX |
| `gcp.private_service_connect` | Private Service Connect | network | A | Consumer-to-producer private endpoints (PSC). | psc, private service connect | PSC to SaaS provider |
| `gcp.vpc_service_controls` | VPC Service Controls | security | A | Service perimeter — prevents data exfil across project/API boundary. | vpc sc, vpc service controls | Data perimeter |
| `gcp.service_directory` | Service Directory | service_discovery | A | Central registry of services (DNS + API). | service directory | Cross-cluster discovery |
| `gcp.cloud_interconnect` | Cloud Interconnect | network | A | Dedicated (10/100 G) or Partner (50 Mb–10 G) to a Google PoP. | interconnect | DX into GCP |
| `gcp.cross_cloud_interconnect` | Cross-Cloud Interconnect | network | E | Direct private link to AWS / Azure / Oracle / Alibaba via Google. | cross cloud interconnect | CCI to AWS |
| `gcp.ncc` | Network Connectivity Center | network | A | Hub for Interconnect / VPN / 3rd-party router attachments. | ncc | NCC hub + spokes |
| `gcp.ha_vpn` | HA VPN | network | I | 2-tunnel IPsec w/ BGP; 99.99% SLA. | ha vpn, cloud vpn | Branch VPN |
| `gcp.cloud_vpn_classic` | Classic VPN | network | E | Legacy VPN (⚠ prefer HA VPN). | classic vpn | Legacy VPN |
| `gcp.load_balancer_global_external` | Global External LB | lb | I | Anycast L7/L4 — premium tier. | global load balancer, glb | Global web tier |
| `gcp.load_balancer_regional_external` | Regional External LB | lb | I | Regional L7/L4 (envoy-based). | regional lb | Regional API |
| `gcp.load_balancer_internal` | Internal LB | lb | I | L4 ILB (regional). | internal lb | Intra-VPC LB |
| `gcp.load_balancer_internal_l7` | Internal HTTP(S) LB | lb | A | Envoy-based L7 internal. | internal http lb | Intra-service mesh entry |
| `gcp.load_balancer_ncc` | Cross-Region LB | lb | A | Multi-region internal LB. | cross region lb | Multi-region internal |
| `gcp.cloud_cdn` | Cloud CDN | cdn | I | CDN tied to HTTP(S) LB; signed URLs, cache invalidation. | cloud cdn | CDN for statics |
| `gcp.media_cdn` | Media CDN | cdn | A | Anycast CDN for video; 100 ms origin latency target. | media cdn | Live OTT |
| `gcp.cloud_dns` | Cloud DNS | dns | B | Authoritative public + private DNS. | cloud dns | Private zones |
| `gcp.cloud_armor` | Cloud Armor | security | A | L7 WAF + DDoS + bot management + adaptive protection. | cloud armor, gcp waf | WAF + DDoS |
| `gcp.cloud_ids` | Cloud IDS | security | A | Palo Alto-powered IDS for VPC traffic mirroring. | cloud ids | IDS for VPC |
| `gcp.packet_mirroring` | Packet Mirroring | observability | A | Copy flow to collectors (IDS, forensics). | packet mirroring | Mirror to IDS |
| `gcp.apigee` | Apigee | api | A | Enterprise API management (X, Hybrid). | apigee | API platform |
| `gcp.apigee_integration` | Apigee Integration | api | A | iPaaS / integration flows on Apigee. | apigee integration | Low-code integration |
| `gcp.application_integration` | Application Integration | integration | A | iPaaS w/ 200+ connectors (not Apigee). | application integration | SaaS connector flows |

## Security / Identity

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `gcp.iam` | IAM | identity | B | Role-based + resource policy; predefined/custom roles; conditions. | gcp iam, iam, identity access management | IAM role for service account |
| `gcp.workload_identity_federation` | Workload Identity Federation | identity | A | OIDC/SAML federation → short-lived GCP tokens (no service-account keys). | wif, workload identity federation | GitHub Actions → GCP via WIF |
| `gcp.workload_identity` | Workload Identity | identity | A | Bind GKE service accounts to Google service accounts. | workload identity | GKE pods use service account |
| `gcp.beyondcorp_enterprise` | BeyondCorp Enterprise | security | A | Zero-trust access: IAP + context-aware + threat + DLP. | beyondcorp, bce, ztna gcp | ZTNA for SaaS |
| `gcp.iap` | IAP | security | I | Identity-Aware Proxy for GCE / GKE / App Engine / on-prem. | iap, identity-aware proxy | Internal apps behind IAP |
| `gcp.context_aware_access` | Context-Aware Access | security | A | Access conditions (device, IP, geo) for GCP + SaaS. | context aware access | Device + IP gate |
| `gcp.security_command_center` | Security Command Center | security | A | CSPM + CWPP aggregator; Standard/Premium/Enterprise tiers. | scc, security command center | CSPM dashboard |
| `gcp.scc_enterprise` | SCC Enterprise | security | A | Adds CIEM + DSPM + attack-path analysis. | scc enterprise | Attack-path graph |
| `gcp.chronicle_security` | Chronicle SecOps | security | A | Petabyte-scale SIEM (formerly Chronicle Backstory) + SOAR. | chronicle, secops, chronicle siem | Threat hunting |
| `gcp.chronicle_soar` | Chronicle SOAR | security | A | Siemplify-based SOAR inside SecOps. | chronicle soar | Automate triage |
| `gcp.mandiant_threat_intel` | Mandiant Threat Intel | security | E | Curated threat intel from Mandiant researchers. | mandiant, threat intel | IoC enrichment |
| `gcp.kms` | Cloud KMS | crypto | B | Managed cryptographic keys (software-protected). | kms, cloud kms | SSE-KMS for GCS |
| `gcp.cloud_hsm` | Cloud HSM | crypto | A | FIPS 140-2 L3 dedicated HSM (cluster). | cloud hsm | Regulated crypto |
| `gcp.cloud_external_km` | External Key Manager | crypto | A | Keys held on-prem / 3rd party KMS; GCP calls out to encrypt. | ekm, external key manager | Hold your own key |
| `gcp.confidential_space` | Confidential Space | security | E | Enclave for multi-party computation (attested GCE VMs). | confidential space | MPC across orgs |
| `gcp.confidential_gke_nodes` | Confidential GKE Nodes | security | A | Confidential VMs as GKE workers. | confidential gke | Pods on confidential nodes |
| `gcp.secret_manager` | Secret Manager | crypto | I | Versioned secret store with CMEK. | secret manager | App secrets |
| `gcp.certificate_authority_service` | CAS | crypto | A | Managed private CA hierarchy. | cas, certificate authority service | mTLS root |
| `gcp.certificate_manager` | Certificate Manager | crypto | I | Managed public + private certs for LBs. | certificate manager | Wildcard cert |
| `gcp.assured_workloads` | Assured Workloads | compliance | E | Sovereignty controls for FedRAMP / CJIS / ITAR / IL4. | assured workloads | FedRAMP high |
| `gcp.assured_oss` | Assured OSS | security | E | Curated, scanned, signed OSS packages. | assured oss, aoss | Hardened Maven deps |
| `gcp.security_key_enforcement` | Titan Key Enforcement | identity | A | Force hardware 2FA keys for users. | titan key, security key enforcement | Phishing-resistant MFA |

## DevOps / Management

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `gcp.cloud_build` | Cloud Build | ci | I | Container-native CI builds, GitHub triggers, Terraform integrations. | cloud build | GitHub PR build |
| `gcp.cloud_deploy` | Cloud Deploy | cd | I | Progressive delivery for GKE / Cloud Run / Anthos. | cloud deploy | Canary GKE |
| `gcp.artifact_registry` | Artifact Registry | registry | I | Multi-format (OCI, Maven, npm, APT, yum, Python). | artifact registry | Mirror npm + Docker |
| `gcp.binary_authorization` | Binary Authorization | security | A | Policy-signed container deploys. | binary auth, binauthz | Attested deploys only |
| `gcp.config_connector` | Config Connector | iac | A | Represent GCP resources as K8s CRDs. | config connector, kcc | GKE-managed GCP resources |
| `gcp.deployment_manager` | Deployment Manager | iac | E | Legacy IaC (⚠ prefer Terraform). | deployment manager | Old stack |
| `gcp.cloud_logging` | Cloud Logging | observability | B | Log router + sinks to BigQuery/GCS/Pub/Sub; formerly Stackdriver. | cloud logging, stackdriver | Central log sink |
| `gcp.cloud_monitoring` | Cloud Monitoring | observability | B | Metrics + alerts + dashboards + uptime checks. | cloud monitoring, stackdriver | Alert policies |
| `gcp.cloud_trace` | Cloud Trace | tracing | I | Distributed tracing (OTLP). | cloud trace | Latency debugging |
| `gcp.cloud_profiler` | Cloud Profiler | observability | A | Continuous CPU / memory profiling. | cloud profiler | Perf profile |
| `gcp.error_reporting` | Error Reporting | observability | I | Groups and alerts on errors from logs + instrumentation. | error reporting | Sentry-lite |
| `gcp.ops_agent` | Ops Agent | observability | I | Unified telemetry agent (metrics + logs + fluent-bit). | ops agent | Install on VMs |
| `gcp.cloud_asset_inventory` | Cloud Asset Inventory | governance | A | Time-series view of all resources + IAM across org. | cloud asset, cai | Daily snapshot |
| `gcp.recommender` | Recommender | cost | I | Cost, security, reliability, perf recommendations. | recommender, active assist | IAM recommendations |
| `gcp.cloud_billing` | Cloud Billing | cost | B | Billing account + budgets + cost breakdown. | cloud billing, billing | Multi-project billing |
| `gcp.cloud_shell` | Cloud Shell | devops | B | Browser shell w/ gcloud preinstalled. | cloud shell | Quick gcloud |
| `gcp.cloud_code` | Cloud Code | devops | I | IDE plugin for K8s / Cloud Run. | cloud code | VS Code integration |
| `gcp.firebase` | Firebase | paas | B | Mobile/web BaaS — Hosting, Auth, Firestore, Storage, FCM, Remote Config. | firebase | Mobile backend |
| `gcp.firebase_extensions` | Firebase Extensions | paas | I | Packaged reusable Firebase functions. | firebase extensions | Stripe ext |
| `gcp.agones` | Agones | games | E | Game-server orchestration on K8s. | agones | Multiplayer lobby |
| `gcp.game_servers` | Game Servers | games | E | (⚠ retired — migrated to Agones on GKE). | game servers | Legacy |

## Edge / Hybrid / Distributed

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `gcp.distributed_cloud_edge` | Distributed Cloud Edge | hybrid | E | GCP-managed hardware at customer edge (5G, retail). | gdc edge, distributed cloud edge | Edge compute in store |
| `gcp.distributed_cloud_air_gapped` | Distributed Cloud Air-Gapped | hybrid | E | Fully air-gapped GCP region for gov / intel. | gdc air gapped | Sovereign cloud |
| `gcp.distributed_cloud_hosted` | Distributed Cloud Hosted | hybrid | E | Partner-hosted GCP region. | gdc hosted | Carrier-hosted GCP |
| `gcp.migrate_to_vms` | Migrate to VMs | migration | A | VMware / AWS / Azure / Hyper-V → GCE (formerly Velostrata). | migrate to vms, m2vm | Lift-and-shift VMs |
| `gcp.migrate_to_containers` | Migrate to Containers | migration | A | VMs → GKE via modernization analysis. | migrate to containers, m4a | Legacy app modernization |
| `gcp.database_migration_service` | Database Migration | migration | I | Continuous MySQL / Postgres / Oracle → CloudSQL / AlloyDB. | dms, database migration service | Postgres live migration |
| `gcp.transfer_cloud` | Cloud-to-Cloud Transfer | migration | A | Managed S3 / Azure → GCS. | cloud to cloud transfer | S3 → GCS |

---

**Entry count: ~110.** Expect additions as GCP ships; canonical names stable.
