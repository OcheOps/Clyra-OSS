# Azure — Component Catalog

Microsoft Azure with real depth. Covers core, Entra (formerly AD), Fabric,
Arc hybrid surface, niche services (Orbital, Quantum, Sphere, HBv4 HPC).

## Compute

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `azure.vm` | Virtual Machine | vm | B | Hyper-V based VMs across D/E/F/B/M/HB/NC/ND/NV/Lsv3 families. | azure vm, virtual machine, vm | D4s_v5 web tier |
| `azure.vmss` | VM Scale Set | vm | I | Orchestration mode (Uniform vs Flexible) for N similar VMs. | vmss, scale set | VMSS Flex with 3 AZs |
| `azure.dedicated_host` | Dedicated Host | vm | A | Physical server reserved for one Azure sub; BYOL for Windows/SQL. | azure dedicated host, dedicated host | SQL BYOL on Dedicated Host |
| `azure.confidential_vm` | Confidential VM | security | A | AMD SEV-SNP / Intel TDX VMs w/ guest attestation. | confidential vm, amd sev-snp, tdx vm | Sealed key material in Confidential VM |
| `azure.spot_vm` | Spot VM | vm | I | Spare capacity at discount; eviction by price or capacity. | spot vm, azure spot | Batch workers on Spot |
| `azure.azure_vmware_solution` | Azure VMware Solution | hybrid | E | Bare-metal ESXi SDDC inside Azure, managed by Microsoft/VMware. | avs, azure vmware solution | Lift 500 VMs to AVS |
| `azure.aks` | AKS | container | B | Managed Kubernetes; AKS Automatic for hands-off mode. | aks, azure kubernetes service | AKS + Cilium dataplane |
| `azure.aks_fleet` | AKS Fleet Manager | container | A | Multi-cluster AKS management & upgrade orchestration. | aks fleet, fleet manager | Fleet of 40 AKS clusters |
| `azure.aks_automatic` | AKS Automatic | container | A | AKS with opinionated defaults, auto-upgrades, auto-scale. | aks automatic | Managed K8s no ops |
| `azure.aci` | Container Instances | container | I | Run a single container with no orchestrator. | aci, container instance | Burst ACI from AKS via virtual-kubelet |
| `azure.acr` | Container Registry | registry | I | OCI registry w/ geo-replication, tasks, content trust. | acr, azure container registry | Geo-replicated ACR |
| `azure.container_apps` | Container Apps | serverless | I | Serverless containers on KEDA + Dapr; scale to zero. | container apps, aca | Event-driven API on ACA |
| `azure.functions` | Functions | serverless | B | Event-driven functions; Consumption / Premium / Dedicated / Flex Consumption. | azure functions, function app | Queue-triggered Function app |
| `azure.functions_durable` | Durable Functions | serverless | A | Stateful orchestrations (orchestrator / activity / entity / sub-orchestrator). | durable functions | Fan-out/fan-in in Durable |
| `azure.app_service` | App Service | paas | B | PaaS for web apps (Windows/Linux, .NET/Java/Node/Python/PHP). | app service, web app, paas | ASP.NET app on App Service |
| `azure.app_service_environment` | ASE v3 | paas | A | Dedicated App Service in customer VNet, single-tenant. | ase, app service environment | Regulated ASE for finance |
| `azure.spring_apps` | Spring Apps | paas | A | Managed Spring Boot / Steeltoe (⚠ retiring 2028). | spring apps, azure spring | Spring microservices on Spring Apps |
| `azure.batch` | Batch | compute | I | Massively parallel batch scheduler (pool + job + task). | azure batch, batch | Monte-Carlo pricing on Batch |
| `azure.hpc_hbv4` | HBv4 HPC | hpc | E | InfiniBand HDR-200 Gb/s HPC SKUs for CFD / MD. | hbv4, hpc vm, infiniband | Weather sim on HBv4 |
| `azure.service_fabric` | Service Fabric | container | A | Microsoft microservice platform (stateful actors). | service fabric | Legacy SF cluster |
| `azure.static_web_apps` | Static Web Apps | paas | B | Static frontend + Functions API + auth, GitHub flow. | static web apps, swa | Next.js + Functions API |

## Networking

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `azure.vnet` | VNet | network | B | Isolated IP space with subnets, peering, service endpoints. | vnet, virtual network | Hub VNet 10.0.0.0/16 |
| `azure.subnet` | Subnet | network | B | Segment inside a VNet; delegation for services (SQL MI, Functions, ASE). | subnet | SQL MI delegated subnet |
| `azure.nsg` | Network Security Group | security | B | Stateful subnet/NIC ACLs. | nsg, network security group | NSG deny 3389 from internet |
| `azure.asg` | Application Security Group | security | I | Named group of NICs referenced in NSG rules. | asg, app security group | ASG "web", "db" in NSG |
| `azure.route_table` | Route Table | network | I | Subnet UDRs for next-hop injection. | route table, udr | UDR to NVA next-hop |
| `azure.public_ip` | Public IP | network | B | Static/dynamic PIPs; Standard SKU required for AZs. | public ip, pip | Standard PIP for LB |
| `azure.nat_gateway` | NAT Gateway | network | I | Outbound NAT at subnet level (no SNAT-port exhaustion). | azure nat gateway | Subnet NAT avoiding LB SNAT |
| `azure.load_balancer` | Load Balancer | lb | B | L4 LB — Standard SKU, zonal / zone-redundant. | azure lb, load balancer | Public Std LB front of VMSS |
| `azure.internal_lb` | Internal Load Balancer | lb | I | Private L4 LB in VNet. | ilb, internal load balancer | ILB in front of SQL AG |
| `azure.cross_region_lb` | Cross-Region LB | lb | A | Global anycast front of regional Std LBs. | cross region lb, global lb | Active-active regional tier |
| `azure.application_gateway` | App Gateway | lb | I | L7 LB + WAF v2 (OWASP CRS 3.2 / DRS). | application gateway, app gw, agw | AGW WAF for public portal |
| `azure.application_gateway_for_containers` | AGC | lb | A | ALB Controller-class L7 for AKS, Gateway API compatible. | application gateway for containers, agc | AGC Gateway API in AKS |
| `azure.front_door_classic` | Front Door Classic | cdn | A | Legacy global L7 + CDN (⚠ prefer Standard/Premium). | front door classic | Legacy AFD |
| `azure.front_door_std` | Front Door Standard | cdn | I | Global L7 reverse proxy + CDN + WAF. | front door, afd | Multi-region web app via AFD |
| `azure.front_door_premium` | Front Door Premium | cdn | A | AFD + Private Link origins + bot protection. | front door premium, afd premium | AFD Premium w/ Private Link origin |
| `azure.cdn_akamai` | Azure CDN | cdn | I | Microsoft / Verizon / Akamai profile options (⚠ Verizon retired). | azure cdn | MS CDN for static assets |
| `azure.traffic_manager` | Traffic Manager | dns | I | DNS-based GSLB (priority/perf/geo/weight/subnet). | traffic manager, gslb | Active-passive DNS failover |
| `azure.virtual_wan` | Virtual WAN | network | A | Managed SD-WAN hub with ExpressRoute + VPN + NVA support. | virtual wan, vwan | vWAN hub w/ 50 branches |
| `azure.virtual_hub` | Virtual Hub | network | A | Spoke-and-hub routing center inside vWAN. | virtual hub | vWAN hub in West US 2 |
| `azure.route_server` | Route Server | network | A | BGP endpoint that peers with NVAs in VNet. | route server | Route Server w/ Cisco CSR |
| `azure.azure_firewall` | Azure Firewall | security | A | Managed stateful firewall (Basic/Std/Premium); IDPS + TLS inspect. | azure firewall, firewall | Premium AFW w/ TLS inspection |
| `azure.firewall_manager` | Firewall Manager | security | A | Policy + network topology orchestration for Azure Firewall + DDoS. | firewall manager | Org firewall policy |
| `azure.ddos_protection` | DDoS Protection | security | A | Network-layer DDoS for public IPs (IP plan or Network plan). | ddos protection | Protect 200 public IPs |
| `azure.waf_policy` | WAF Policy | security | I | Shared policy attached to AGW / AFD / APIM. | waf policy | DRS 2.1 WAF policy |
| `azure.expressroute` | ExpressRoute | network | A | Private L2 connection to Microsoft via carrier. | expressroute, er, private circuit | 10 G ER to UK South |
| `azure.expressroute_direct` | ExpressRoute Direct | network | E | Dual 10/100 Gbps ER at Microsoft peering (no telco middleman). | er direct, expressroute direct | ER Direct for hyperscaler |
| `azure.expressroute_global_reach` | ER Global Reach | network | A | Connect two ER circuits worldwide over Microsoft backbone. | global reach | Branch-to-branch via MS backbone |
| `azure.vpn_gateway` | VPN Gateway | network | I | IPsec/IKEv2 gateway (route-based/policy-based) + P2S OpenVPN/IKE. | vpn gateway | S2S + P2S VPN |
| `azure.bastion` | Bastion | security | I | Managed jump host (RDP/SSH) via Azure portal, no public IP. | azure bastion, bastion | Bastion to private VMs |
| `azure.private_link` | Private Link | network | I | Private endpoint into PaaS services via your VNet. | private link, private endpoint | Private Link to Storage |
| `azure.private_endpoint` | Private Endpoint | network | I | NIC in your subnet projecting a PaaS service. | private endpoint | PE for Cosmos DB |
| `azure.service_endpoint` | Service Endpoint | network | I | VNet-scoped short-path to PaaS (less isolating than PE). | service endpoint | SE for Storage firewall |
| `azure.vnet_peering` | VNet Peering | network | I | Transitive-less private link between VNets (same/cross-region). | vnet peering | Hub ↔ spoke peering |
| `azure.dns` | Azure DNS | dns | B | Public + private DNS zones. | azure dns, public dns, private dns | Private zone for internal apps |
| `azure.private_dns` | Private DNS | dns | I | Per-VNet DNS zones + DNS records for Private Link. | private dns zone | privatelink.blob.core.windows.net zone |
| `azure.network_watcher` | Network Watcher | observability | I | Packet capture, flow logs, connection troubleshooter, NSG analytics. | network watcher | Flow logs to Log Analytics |
| `azure.connection_monitor` | Connection Monitor | observability | A | Continuous reachability + latency probes across sources. | connection monitor | Hub-spoke reachability monitor |

## Storage

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `azure.storage_account` | Storage Account | storage | B | Namespace for blobs/queues/tables/files (StorageV2). | storage account, account | Standard LRS SA |
| `azure.blob` | Blob Storage | object | B | Object storage; tiers Hot/Cool/Cold/Archive; immutable blobs. | blob, azure blob, object storage | Parquet lake in Blob |
| `azure.datalake_gen2` | Data Lake Gen2 | object | I | Hierarchical namespace on Blob for Hadoop/Spark. | adls, adls gen2, data lake | ADLS for Databricks |
| `azure.files` | Azure Files | file | B | SMB / NFSv4.1 shares; AD Auth; Azure File Sync to on-prem. | azure files, smb share | AD-joined shares |
| `azure.netapp_files` | NetApp Files | file | A | Bare-metal NetApp ONTAP volumes at HPC throughput. | anf, netapp files | ANF for SAP HANA |
| `azure.managed_disk` | Managed Disk | block | B | Standard HDD / Std SSD / Premium SSD / Premium v2 / Ultra. | managed disk, disk | Ultra disk for SQL tempdb |
| `azure.queues` | Storage Queues | queue | I | Simple REST queue (not Service Bus). | azure queue, storage queue | Legacy queue |
| `azure.tables` | Storage Tables | nosql | I | NoSQL key-value (cheap). | storage table | Small partition key store |
| `azure.data_box` | Data Box | migration | A | 100 TB shippable appliance for offline ingress/egress. | data box | 100 TB → Azure via Data Box |
| `azure.site_recovery` | Site Recovery | dr | A | DR for on-prem / Azure VMs; RPO ~30 s via continuous replication. | site recovery, asr | VMware DR into Azure |
| `azure.backup` | Azure Backup | backup | I | Policy-driven VM/SQL/File/AzFiles/DB backups w/ immutable vault. | azure backup | 7 y immutable vault |

## Database

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `azure.sql_database` | Azure SQL DB | relational | B | Managed SQL Server — Single, Elastic Pool, Serverless, Hyperscale. | azure sql, sql database | Hyperscale with 100 TB |
| `azure.sql_managed_instance` | SQL Managed Instance | relational | A | Near-100% SQL Server feature parity in VNet. | sql mi, managed instance | MI for legacy ERP |
| `azure.sql_edge` | SQL Edge | relational | E | SQL Server optimized for IoT gateways (⚠ retired 2025). | sql edge | Factory-floor SQL Edge |
| `azure.postgres_flexible` | Postgres Flexible | relational | I | Managed Postgres w/ AZ-redundant HA, near-zero-downtime patch. | postgres, azure postgres, postgres flex | Flexible Postgres 16 |
| `azure.mysql_flexible` | MySQL Flexible | relational | I | Managed MySQL w/ burstable + GP + business-critical tiers. | azure mysql, mysql flex | MySQL 8 flexible |
| `azure.mariadb` | MariaDB | relational | E | Managed MariaDB (⚠ retiring 2025). | azure mariadb | Legacy MariaDB |
| `azure.cosmos_nosql` | Cosmos NoSQL | nosql | I | Document/key-value API, single-digit-ms, multi-region writes. | cosmos db, cosmos nosql, documentdb azure | Globally distributed profile store |
| `azure.cosmos_mongo_vcore` | Cosmos Mongo vCore | nosql | A | Mongo 6.0+ compatibility with dedicated vCore scaling (distinct from RU). | cosmos mongo, mongo vcore | Mongo 6 on vCore |
| `azure.cosmos_cassandra` | Cosmos Cassandra | nosql | A | Managed Cassandra API. | cosmos cassandra | CQL workload, no ops |
| `azure.cosmos_gremlin` | Cosmos Gremlin | nosql | A | Graph API over Cosmos. | cosmos gremlin | Social graph |
| `azure.cosmos_table` | Cosmos Table | nosql | I | Premium TableStorage-compatible. | cosmos table | Premium tables |
| `azure.cosmos_postgres` | Cosmos PostgreSQL | nosql | A | Citus-powered distributed Postgres (formerly Hyperscale Citus). | cosmos postgres, citus | Multi-tenant SaaS shards |
| `azure.redis` | Cache for Redis | cache | B | Managed Redis (⚠ migrating to Redis Enterprise). | azure redis, cache for redis | Session cache |
| `azure.redis_enterprise` | Redis Enterprise | cache | A | Geo-replicated enterprise Redis w/ RediSearch + JSON modules. | redis enterprise | Active-active cache |
| `azure.data_explorer` | Data Explorer (Kusto) | analytics | A | Time-series / log / IoT store; KQL; 10× faster than ES for logs. | adx, kusto, data explorer | Telemetry search |

## Messaging & Events

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `azure.service_bus` | Service Bus | broker | I | Queues / topics / sessions / DLQ / transactions. | service bus, asb | Ordered order queue |
| `azure.event_grid` | Event Grid | events | I | Pub/sub event routing w/ filter + push. | event grid, eg | Resource events → Logic App |
| `azure.event_grid_mqtt` | Event Grid MQTT | iot | A | MQTT 3.1.1 / 5 broker for IoT use-cases. | event grid mqtt | MQTT broker for fleet |
| `azure.event_hubs` | Event Hubs | streaming | I | High-throughput ingest; Kafka endpoint; Capture to Blob/ADLS. | event hubs, azure kafka | 1 M msg/s into EH |
| `azure.event_hubs_dedicated` | EH Dedicated | streaming | A | Single-tenant EH cluster; 100 MB/s or more. | event hubs dedicated | Dedicated EH cluster |
| `azure.relay` | Relay | network | A | Hybrid Connections — expose on-prem HTTP/WebSockets to cloud. | service bus relay, hybrid connections | On-prem API via Relay |
| `azure.notification_hubs` | Notification Hubs | messaging | I | Cross-platform push (APNs/FCM/WNS) fan-out. | notification hubs | App push fan-out |
| `azure.logic_apps` | Logic Apps | integration | I | iPaaS workflow engine (Consumption + Standard). | logic apps | Approval workflow |
| `azure.apim` | API Management | api | I | Managed API gateway (Consumption/Developer/Basic/Std/Premium). | apim, api management | APIM facade for microservices |
| `azure.integration_services_environment` | ISE | integration | E | Single-tenant Logic Apps / APIM (⚠ retired, moved to Std). | ise | Legacy ISE |

## AI / ML

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `azure.openai` | Azure OpenAI | ai | B | GPT-4/4o/o1/embedding/DALL·E with enterprise SLAs + data boundary. | azure openai, aoai, gpt | Chatbot on AOAI |
| `azure.ai_foundry` | AI Foundry | ai | I | Unified Studio for model catalog + agents + evaluation (formerly AI Studio). | ai foundry, azure ai foundry | Build + eval agents |
| `azure.ai_search` | AI Search | search | I | Vector + keyword hybrid search with semantic ranker. | ai search, cognitive search | RAG index over KB |
| `azure.document_intelligence` | Document Intelligence | ai | I | Form + doc extraction (Layout, Read, Receipt, Invoice, ID, Custom). | document intelligence, form recognizer | Invoice parsing |
| `azure.speech` | Speech | ai | I | STT + TTS + translation + voice cloning + keyword. | speech service | Call-center STT |
| `azure.language` | Language | ai | I | Text analytics, PII, entities, sentiment, summarization. | language service | PII redaction |
| `azure.translator` | Translator | ai | B | 130+ language translation. | translator | i18n pipeline |
| `azure.vision` | Vision | ai | I | Image analysis, OCR, spatial analysis. | computer vision, azure vision | Shelf recognition |
| `azure.face` | Face | ai | A | Face detection + recognition (restricted availability). | face api | Auth via Face (gated) |
| `azure.content_safety` | Content Safety | ai | A | Classify harmful content; promptshield + groundedness detection. | content safety | Moderation for UGC |
| `azure.video_indexer` | Video Indexer | ai | A | Multi-modal video insights (OCR, speakers, topics). | video indexer | Media archive search |
| `azure.bot_service` | Bot Service | ai | I | Managed hosting for Bot Framework. | bot service | Teams bot |
| `azure.ml_workspace` | Azure ML | ai | I | ML platform — designer, notebooks, pipelines, endpoints, models. | azure ml, aml | Training + endpoints |
| `azure.ml_prompt_flow` | Prompt Flow | ai | A | Graph-based LLM pipelines with eval + tracing. | prompt flow, promptflow | Agent pipeline evaluation |
| `azure.immersive_reader` | Immersive Reader | ai | E | Accessibility text reader SDK. | immersive reader | Accessibility feature |
| `azure.ml_compute_cluster` | ML Compute Cluster | ai | A | Autoscaling training cluster for AML. | ml compute cluster | GPU cluster for training |
| `azure.ai_content_understanding` | Content Understanding | ai | A | Multi-modal document understanding (2025+). | content understanding | Structured extraction |

## IoT / Hybrid / Arc

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `azure.iot_hub` | IoT Hub | iot | B | Device gateway w/ MQTT/AMQP; twins, direct methods, file upload. | iot hub | 50 k devices on IoT Hub |
| `azure.iot_central` | IoT Central | iot | I | SaaS app platform for IoT. | iot central | Branded customer IoT app |
| `azure.iot_edge` | IoT Edge | iot | A | Edge runtime for Docker modules + custom modules. | iot edge, azure iot edge | Edge inference |
| `azure.iot_dps` | Device Provisioning Service | iot | A | Just-in-time enrollment to the right IoT Hub. | dps, device provisioning | 1 M-device provisioning |
| `azure.defender_for_iot` | Defender for IoT | iot | A | OT/IoT asset inventory + threat detect (network + agent). | defender for iot | OT monitoring in plant |
| `azure.digital_twins` | Digital Twins | iot | A | DTDL-modeled graph of physical world + live telemetry. | digital twins, adt | Smart building twin |
| `azure.sphere` | Sphere | iot | E | Secured microcontroller platform (⚠ EoS 2027). | sphere | MCU w/ attested OS |
| `azure.percept` | Percept | iot | E | Edge AI kit (⚠ retired). | percept | Legacy Percept |
| `azure.arc` | Arc | hybrid | A | Project Azure control plane onto on-prem / other-cloud resources. | azure arc, arc | GKE cluster into Arc |
| `azure.arc_servers` | Arc Servers | hybrid | A | Azure management for Windows/Linux anywhere. | arc servers, connected machine | Azure Policy on on-prem VMs |
| `azure.arc_kubernetes` | Arc-enabled K8s | hybrid | A | K8s clusters (any distro) projected into Azure. | arc k8s, arc kubernetes | On-prem k8s managed via Arc |
| `azure.arc_data_services` | Arc Data | hybrid | A | SQL MI / Postgres Hyperscale outside Azure. | arc data | SQL MI on-prem |
| `azure.stack_hci` | Stack HCI | hybrid | A | Hyperconverged on customer hardware, Azure-aligned. | stack hci | HCI cluster in branch |
| `azure.stack_hub` | Stack Hub | hybrid | E | Full Azure region in customer data center (⚠ niche). | stack hub | Sovereign region |
| `azure.stack_edge` | Stack Edge | hybrid | A | Rugged appliance with GPU for edge ML + data transfer. | stack edge | Edge ML appliance |
| `azure.orbital` | Orbital | space | E | Satellite ground-station as a service (commercial). | orbital, azure orbital | LEO contacts via Orbital |

## Security & Identity

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `azure.entra_id` | Entra ID | identity | B | Cloud IdP (formerly Azure AD) — SSO, SAML, OIDC, SCIM. | azure ad, entra id, aad | SSO for SaaS |
| `azure.entra_external` | Entra External ID | identity | I | Customer identity (CIAM) — formerly B2C + External Identities. | b2c, entra external, external id | Consumer app signup |
| `azure.entra_domain_services` | Entra Domain Services | identity | I | Managed DC for legacy AD-bound apps. | aad ds, entra domain services | AD-joined VMs via AAD-DS |
| `azure.entra_verified_id` | Entra Verified ID | identity | A | Decentralized identity (W3C VC). | verified id, decentralized identity | Employee credential |
| `azure.entra_pim` | PIM | identity | A | Just-in-time privileged role activation. | pim, privileged identity | JIT Global Admin |
| `azure.entra_conditional_access` | Conditional Access | identity | A | Risk + signal-based policy engine. | conditional access, ca | Block legacy auth, enforce MFA |
| `azure.entra_identity_protection` | Identity Protection | identity | A | Detect risky sign-ins + users (atypical travel, leaked creds). | identity protection | Risk-based auto-remediation |
| `azure.entra_permissions_management` | Permissions Management | identity | A | CIEM — cross-cloud entitlement governance (formerly CloudKnox). | entra permissions management, ciem | Identify over-permissioned roles |
| `azure.key_vault` | Key Vault | crypto | B | Keys, secrets, certificates in FIPS 140-2 L2 HSMs. | key vault, kv | App secrets in KV |
| `azure.managed_hsm` | Managed HSM | crypto | A | Dedicated FIPS 140-2 L3 HSM pool. | managed hsm | Healthcare HSM |
| `azure.dedicated_hsm` | Dedicated HSM | crypto | E | Bare-metal Thales Luna HSM you rent. | dedicated hsm | Crypto root of trust |
| `azure.sentinel` | Sentinel | security | A | Cloud-native SIEM + SOAR on Log Analytics. | sentinel, siem | SOC analytics on Sentinel |
| `azure.defender_cloud` | Defender for Cloud | security | I | CSPM + CWPP across VMs/containers/DBs/storage/API. | defender for cloud, mdc | CSPM + vuln on Azure |
| `azure.defender_xdr` | Defender XDR | security | A | XDR across endpoints + identity + email + cloud apps. | defender xdr | Unified SOC pane |
| `azure.defender_endpoint` | Defender for Endpoint | security | I | EDR + MDR for Windows/Mac/Linux/Android/iOS. | mde, defender for endpoint | Enterprise EDR |
| `azure.defender_identity` | Defender for Identity | security | A | On-prem AD + Entra-ID attack detection. | defender for identity, atp | Lateral-movement detection |
| `azure.defender_servers` | Defender for Servers | security | A | Server-plan w/ vulnerability scan + JIT + FIM. | defender for servers | Vuln scan on Arc servers |
| `azure.defender_api` | Defender for APIs | security | A | API Management inline protection + posture. | defender for apis | API abuse detection |
| `azure.defender_dsbm` | Data Security Posture Mgmt | security | A | Classify + prioritize sensitive data risk. | dspm, data posture | PII discovery |
| `azure.purview` | Microsoft Purview | governance | A | Data catalog + data map + DLP + eDiscovery + info protection. | purview | Data mesh catalog |
| `azure.attestation` | Attestation | crypto | A | Remote attestation for Confidential VMs / enclaves / TPM. | attestation | Enclave attestation service |
| `azure.app_proxy` | Entra Application Proxy | security | A | Publish on-prem HTTP apps via Entra ID. | application proxy, app proxy | Internal SharePoint over net |

## Data / Analytics / Fabric

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `azure.synapse` | Synapse Analytics | warehouse | A | Integrated DW + Spark + Pipelines + SQL serverless. | synapse, synapse analytics | Lakehouse pattern |
| `azure.synapse_dedicated_sql` | Synapse Dedicated SQL | warehouse | A | MPP DW (former SQL DW / Gen2). | dedicated sql pool, azure dw | Finance mart |
| `azure.synapse_serverless_sql` | Synapse Serverless SQL | warehouse | I | Pay-per-query SQL over ADLS. | serverless sql | Parquet over ADLS |
| `azure.databricks` | Databricks | analytics | A | Unified Lakehouse (Delta + Unity Catalog + MLflow). | databricks, azure databricks, adb | Bronze/Silver/Gold pipelines |
| `azure.data_factory` | Data Factory | etl | I | Cloud ETL/ELT; mapping + wrangling dataflows; 100+ connectors. | adf, data factory | On-prem SQL → ADLS |
| `azure.hdinsight` | HDInsight | analytics | E | Managed Hadoop/Spark/Kafka (⚠ legacy). | hdinsight | Legacy Hadoop |
| `azure.stream_analytics` | Stream Analytics | streaming | I | SQL-over-streams w/ windowing; outputs to Synapse/ADLS/EH. | asa, stream analytics | IoT alert rule |
| `azure.fabric` | Microsoft Fabric | analytics | A | Unified lakehouse SaaS — OneLake, Warehouse, KQL DB, Real-Time. | fabric, microsoft fabric | OneLake for all analytics |
| `azure.fabric_onelake` | OneLake | analytics | A | Single data lake behind Fabric (one copy via shortcuts). | onelake | Shortcut to ADLS |
| `azure.fabric_warehouse` | Fabric Warehouse | warehouse | A | T-SQL warehouse on OneLake. | fabric warehouse | Exec BI warehouse |
| `azure.fabric_kql` | Fabric KQL DB | analytics | A | ADX built into Fabric. | fabric kql | Event analytics |
| `azure.fabric_real_time` | Fabric Real-Time | streaming | A | Event streams, routing, KQL analytics in Fabric. | fabric real-time | Clickstream routing |
| `azure.data_share` | Data Share | integration | A | Share snapshots or in-place w/ partners (ADLS / Blob / SQL DB). | data share | Partner data share |
| `azure.power_bi` | Power BI | bi | B | BI reporting + Semantic Models + Direct Lake on OneLake. | power bi, pbi | Exec dashboards |

## Management, DevOps, Cost

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `azure.arm` | Resource Manager | control_plane | B | Azure's global control plane + deployment engine. | arm, azure resource manager | ARM template |
| `azure.bicep` | Bicep | iac | I | DSL that transpiles to ARM JSON. | bicep | Bicep for VNet + AKS |
| `azure.terraform` | Terraform AzureRM | iac | I | AzureRM provider — defacto IaC outside Microsoft. | terraform, azurerm | TF for AKS |
| `azure.policy` | Policy | governance | I | Guardrails + remediation tasks. | policy, azure policy | Deny non-standard SKUs |
| `azure.blueprints` | Blueprints | governance | E | Composable bundle of policy + RBAC + ARM (⚠ retiring → TF). | blueprints | Legacy landing zone |
| `azure.lighthouse` | Lighthouse | governance | A | Delegated management across customer tenants (MSP). | lighthouse | MSP managing 30 tenants |
| `azure.resource_graph` | Resource Graph | governance | I | KQL over Azure resource metadata. | resource graph | Inventory query |
| `azure.cost_management` | Cost Management | cost | B | Cross-cloud cost analysis + budgets + exports. | cost management | Chargeback by tag |
| `azure.advisor` | Advisor | cost | B | Recommendations (cost, reliability, security, perf, ops). | advisor | Advisor weekly digest |
| `azure.monitor` | Monitor | observability | B | Metrics + logs + alerts + workbooks + Application Insights. | azure monitor, monitor | Alert on CPU anomaly |
| `azure.log_analytics` | Log Analytics | observability | I | KQL log store behind Monitor + Sentinel. | log analytics, la workspace | Central logs workspace |
| `azure.application_insights` | App Insights | observability | I | APM (now part of Monitor) — transactions, dep map, live metrics. | application insights, app insights | .NET APM |
| `azure.action_groups` | Action Groups | observability | I | Alert fan-out (email/SMS/webhook/Logic Apps/ITSM/Functions). | action group | Page on-call on alert |
| `azure.workbooks` | Workbooks | observability | I | Parameterized notebooks over Monitor data. | workbook | Incident workbook |
| `azure.devops` | Azure DevOps | devops | I | Repos + Pipelines + Boards + Artifacts + Test Plans. | ado, azure devops | CI/CD on ADO |
| `azure.github_enterprise` | GitHub EMU | devops | I | Enterprise Managed User GitHub linked to Entra. | github emu | Enterprise GitHub via Entra |
| `azure.deployment_env` | Deployment Environments | devops | A | Per-PR env templating on IaC. | deployment environments | Ephemeral PR envs |
| `azure.dev_box` | Dev Box | devops | A | Cloud dev workstation (like Codespaces). | dev box, microsoft dev box | VS dev workstation |
| `azure.app_configuration` | App Configuration | config | I | Feature flags + key-value store. | app configuration | Shared feature flags |
| `azure.chaos_studio` | Chaos Studio | resilience | A | Managed fault injection across VM/AKS/App Service. | chaos studio, chaos | Zone-failure drill |

## Communications & Maps

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `azure.communication_services` | Communication Services | communications | I | Voice / chat / SMS / email SDK w/ Teams interop. | acs, communication services | In-app video |
| `azure.maps` | Maps | geo | I | Maps, routing, geocoding, weather, creator. | azure maps | Fleet routing |
| `azure.health_data_services` | Health Data Services | health | E | FHIR + DICOM + MedTech (IoT→FHIR). | health data services, fhir service | Hospital EHR |

## Quantum & Specialized

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `azure.quantum` | Azure Quantum | quantum | E | IonQ / Quantinuum / Rigetti / Pasqal / QCI hardware + sim. | azure quantum | VQE on IonQ |
| `azure.playfab` | PlayFab | games | E | Game backend (data, LiveOps, player mgmt). | playfab | MMO backend |
| `azure.remote_rendering` | Remote Rendering | xr | E | Cloud GPU rendering for HoloLens + head-mounted. | remote rendering | Field-service AR |
| `azure.spatial_anchors` | Spatial Anchors | xr | E | Cross-device persisted AR anchors. | spatial anchors | Shared AR experience |
| `azure.high_performance_fsx` | Managed Lustre | hpc | A | Parallel filesystem (Lustre) for HPC / AI. | managed lustre | ML training feed |
| `azure.batch_ai` | Batch AI | ai | E | (⚠ folded into AML). | batch ai | Legacy |

---

**Entry count: ~110.** Canonical names stay stable across MS renames; aliases carry old names.
