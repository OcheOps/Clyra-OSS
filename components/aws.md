# AWS — Component Catalog

Every meaningful AWS service and primitive architects use in production.
Beginner → Expert. Includes obscure surfaces (Ground Station, Nitro Enclaves,
VPC Lattice, IoT FleetWise, Omics, Resilience Hub).

## Compute

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.ec2` | EC2 | vm | B | Elastic Compute Cloud — VM on Nitro hypervisor. ~500 instance types (t/m/c/r/i/d/x/z/p/g/inf/trn/hpc). | ec2, instance, vm, amazon ec2, nitro vm, virtual machine | 3× t3.medium auto-scaled behind ALB |
| `aws.ec2_spot` | Spot Instance | vm | I | Spare capacity at up to 90% discount; can be reclaimed with 2-min notice. | spot, spot fleet, spot instance, ec2 spot | Batch workers on Spot + on-demand base |
| `aws.ec2_dedicated_host` | Dedicated Host | vm | A | Physical server reserved for one customer; required for BYOL licenses (Oracle, Windows Server). | dedicated host, dedicated hardware, bring your own license | Oracle DB licensed by socket on Dedicated Host |
| `aws.ec2_capacity_reservation` | Capacity Reservation | vm | A | Reserved capacity in a specific AZ without commitment; protects against stockout. | odcr, capacity reservation, on-demand capacity reservation | Reserve 50 p4d in us-east-1a for training |
| `aws.ec2_savings_plan` | Savings Plan | billing | I | 1 or 3-year commit to $/hr of compute for 40–72% discount. Compute SP applies across EC2/Fargate/Lambda. | savings plan, compute savings plan, ec2 sp | 3-yr compute SP covering Fargate baseline |
| `aws.nitro_enclave` | Nitro Enclave | security | A | Isolated VM carved from parent EC2; no persistent storage, no external network; cryptographic attestation. | nitro enclave, enclave, confidential compute ec2 | PCI decrypt inside an enclave, KMS attests |
| `aws.nitro_tpm` | NitroTPM | security | E | Virtual TPM 2.0 exposed to EC2 — BitLocker, attestation, secure boot. | nitro tpm, virtual tpm, tpm | Windows BitLocker rooted in NitroTPM |
| `aws.graviton` | Graviton | vm | I | ARM64 (Neoverse) instances (`*g`/`*gd`) — 20–40% cheaper per request vs x86. | graviton, arm, aws arm, neoverse | m7g.large saving 30% over m7i |
| `aws.ec2_instance_connect` | Instance Connect | security | I | Short-lived SSH keypair delivered via IAM; no long-lived keys to rotate. | instance connect, eic, ec2 instance connect | SSH to prod via EIC, no shared keys |
| `aws.ec2_image_builder` | EC2 Image Builder | vm | I | Managed AMI/container image pipeline with testing + publishing. | image builder, ami pipeline, ec2 image builder | Weekly hardened Ubuntu AMI with STIG baseline |
| `aws.placement_group` | Placement Group | vm | A | Cluster/Partition/Spread groupings to control blast radius or inter-node latency. | placement group, cluster placement, spread placement | HPC nodes in Cluster PG for 10 Gbps/1 μs |
| `aws.ecs` | ECS | container | B | Elastic Container Service — scheduler for Docker on EC2/Fargate. | ecs, container service, docker service | Blue-green ECS service on Fargate |
| `aws.ecs_anywhere` | ECS Anywhere | hybrid | A | ECS agent running on-prem servers managed from the AWS control plane. | ecs anywhere, ecs on-prem | Factory floor nodes registered to ECS Anywhere |
| `aws.fargate` | Fargate | container | B | Serverless compute for ECS/EKS tasks — no EC2 to manage. | fargate, serverless containers | API on ECS+Fargate, scale to zero on off-hours |
| `aws.fargate_spot` | Fargate Spot | container | I | Fargate capacity at 70% discount, 2-min interruption notice. | fargate spot | Batch ETL workers on Fargate Spot |
| `aws.eks` | EKS | container | B | Managed Kubernetes control plane. Data plane via EC2 / Fargate / Karpenter. | eks, kubernetes, k8s on aws | EKS cluster with Karpenter + managed node groups |
| `aws.eks_anywhere` | EKS Anywhere | hybrid | A | Install upstream K8s (EKS-D) on customer hardware with the EKS tooling. | eks anywhere, eks on-prem | On-prem bank cluster using EKS Anywhere |
| `aws.eks_distro` | EKS Distro | container | A | Kubernetes source that AWS ships — identical bits to EKS control plane. | eks-d, eks distro | Custom Raspberry Pi cluster from EKS-D |
| `aws.karpenter` | Karpenter | container | A | EKS node autoscaler that provisions right-sized nodes in seconds, respects Spot. | karpenter, eks autoscaler | Karpenter picks Graviton+Spot for burst |
| `aws.lambda` | Lambda | serverless | B | Event-driven functions; 128MB–10GB RAM; up to 15 min; zip or container image. | lambda, serverless function, faas, aws function | Image resize Lambda triggered by S3 put |
| `aws.lambda_snapstart` | Lambda SnapStart | serverless | A | Snapshot-restore for Java/.NET/Python: sub-second cold starts. | snapstart, lambda snapstart, cold start fix | Spring Boot on Lambda with SnapStart |
| `aws.lambda_layer` | Lambda Layer | serverless | I | Shared library/runtime packaged separately; mounted at /opt. | lambda layer, layer | Playwright Chromium in a Lambda Layer |
| `aws.lambda_edge` | Lambda@Edge | edge | A | Lambda running at CloudFront PoPs — viewer/origin request/response hooks. | lambda@edge, lambda at edge, cloudfront lambda | A/B test cookie rewrite at Lambda@Edge |
| `aws.cloudfront_functions` | CloudFront Functions | edge | A | Lightweight JS at PoP edge, sub-ms, cheaper than Lambda@Edge. No network I/O. | cloudfront functions, cff, viewer function | Header normalization on viewer request |
| `aws.app_runner` | App Runner | serverless | I | Container-to-URL in one step; source from ECR or GitHub. | app runner, simple container | Stateless Go service deployed from GitHub |
| `aws.elastic_beanstalk` | Elastic Beanstalk | serverless | I | PaaS wrapping EC2+ASG+ALB with runtime managers (Node, Python, Java…). | elastic beanstalk, beanstalk, eb | Rails app on EB multi-container |
| `aws.batch` | Batch | compute | I | Managed job scheduler across EC2 / Fargate / Spot; DAG via Step Functions. | aws batch, batch jobs | Nightly genome pipeline on Batch+Spot |
| `aws.parallelcluster` | ParallelCluster | hpc | A | Open-source HPC cluster provisioning (Slurm + auto-scaling EC2). | parallelcluster, hpc cluster | Fluid dynamics sims on ParallelCluster with Hpc7g |
| `aws.lightsail` | Lightsail | vm | B | Simplified EC2 + networking bundle for small workloads. | lightsail | WordPress blog on Lightsail |
| `aws.outposts` | Outposts | hybrid | A | Rack of AWS hardware in the customer's data center; same APIs as the cloud. | outposts, aws on-prem rack | Regulated bank: RDS on Outposts in own DC |
| `aws.local_zone` | Local Zone | edge | A | Metro-level AWS extension (Las Vegas, Lagos, Bogotá…) with EC2+EBS+ELB. | local zone, metro zone | Low-latency gaming in Lagos Local Zone |
| `aws.wavelength` | Wavelength | edge | E | AWS compute embedded inside a 5G carrier's network for <10 ms to devices. | wavelength, 5g edge, carrier edge | AR app on Verizon Wavelength zone |

## Storage

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.s3` | S3 | object | B | Object storage — 11 9s durability. Classes: Standard/IA/One-Zone-IA/Intelligent-Tiering/Glacier IR/Flexible/Deep Archive. | s3, bucket, object storage, amazon s3 | Static site + CloudFront origin |
| `aws.s3_express` | S3 Express One Zone | object | A | Directory-bucket, single-AZ, 10× faster PUT/GET for analytics/ML. | s3 express, one zone express, express one zone | Spark shuffle to S3 Express |
| `aws.s3_access_point` | S3 Access Point | object | I | Named hostname + policy wrapped around a bucket; multi-tenant isolation. | s3 access point, access point | Per-team access points on a shared bucket |
| `aws.s3_multiregion_ap` | S3 Multi-Region AP | object | A | Global endpoint routed to the nearest replicated bucket. | s3 multi-region access point, mrap | Active-active objects across us/eu |
| `aws.s3_object_lambda` | S3 Object Lambda | object | A | Transform objects in-flight on GET with a Lambda. | s3 object lambda, object lambda | Redact PII on GET for analysts |
| `aws.ebs` | EBS | block | B | Block volumes — `gp3`/`io2 Block Express`/`st1`/`sc1`. | ebs, block storage, amazon ebs, volume | gp3 boot disk + io2 DB volume |
| `aws.ebs_snapshot` | EBS Snapshot | block | I | Incremental S3-backed snapshots; cross-region/cross-account copy. | ebs snapshot, snapshot | Nightly snapshots with lifecycle |
| `aws.efs` | EFS | file | B | NFSv4 elastic file system; Standard / One-Zone / IA. | efs, nfs, amazon efs | Shared WordPress media on EFS |
| `aws.fsx_lustre` | FSx for Lustre | file | A | HPC parallel filesystem; integrates with S3 as backing store. | fsx lustre, lustre, hpc filesystem | ML training reading from FSx Lustre |
| `aws.fsx_ontap` | FSx for NetApp ONTAP | file | A | Managed ONTAP — SnapMirror, SnapVault, multi-protocol (NFS/SMB/iSCSI). | fsx ontap, netapp ontap | Enterprise file server replacement |
| `aws.fsx_openzfs` | FSx for OpenZFS | file | A | OpenZFS file system — instant snapshots, clones, compression. | fsx openzfs, zfs | Database backing with ZFS clones |
| `aws.fsx_windows` | FSx for Windows | file | I | SMB file server with AD integration, DFS-N. | fsx windows, windows file server | Windows VDI home directories |
| `aws.storage_gateway` | Storage Gateway | hybrid | A | On-prem appliance backing to S3/Glacier: File / Tape / Volume modes. | storage gateway, volume gateway, tape gateway | Backup tape replacement via Tape Gateway |
| `aws.datasync` | DataSync | migration | I | Managed data transfer between on-prem/cloud file systems, S3, EFS, FSx. | datasync | 50 TB NFS → S3 weekly with DataSync |
| `aws.transfer_family` | Transfer Family | storage | I | Managed SFTP/FTPS/FTP/AS2 endpoint in front of S3/EFS. | transfer family, sftp service, managed sftp | Partner SFTP drop → S3 |
| `aws.snowcone` | Snowcone | migration | E | 8 TB rugged device (2.1 kg) for field/edge collection. | snowcone, snowcone sbe | Oil rig telemetry collected on Snowcone |
| `aws.snowball_edge` | Snowball Edge | migration | A | 80 TB rackmount with on-device compute (EC2 sbe instances). | snowball, snowball edge | Vessel-to-shore data via Snowball Edge |
| `aws.snowmobile` | Snowmobile | migration | E | Shipping container + truck; 100 PB per trip. | snowmobile | Exabyte data-center exit via Snowmobile |
| `aws.backup` | AWS Backup | backup | I | Policy-driven backups across EBS/EFS/FSx/RDS/DynamoDB/Storage Gateway/Neptune/DocumentDB. | aws backup, backup plan, backup vault | Cross-region vault with 7-year retention |
| `aws.drs` | Elastic DR | dr | A | Continuous block-level replication → warm stand-by in AWS; 1-min RPO. | elastic drs, cloudendure, disaster recovery | VMware → AWS DR via Elastic DR |

## Database

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.rds` | RDS | relational | B | Managed Postgres/MySQL/MariaDB/Oracle/SQL Server; Multi-AZ, read replicas, automated backup. | rds, relational database, managed database, amazon rds | Postgres RDS Multi-AZ + 2 read replicas |
| `aws.rds_proxy` | RDS Proxy | relational | A | Connection-pool in front of RDS; protects against Lambda connection storms. | rds proxy, connection pool | Lambda at scale via RDS Proxy |
| `aws.aurora` | Aurora | relational | I | MySQL/Postgres-compatible with 6-way replicated storage; auto-scaling storage. | aurora, amazon aurora, aurora postgres, aurora mysql | Aurora Postgres Global Database, us+eu |
| `aws.aurora_serverless` | Aurora Serverless v2 | relational | I | ACU-scaled Aurora; scales per-second, zero to many. | aurora serverless, aurora v2 | Dev/QA DB on Aurora Serverless v2 |
| `aws.aurora_limitless` | Aurora Limitless | relational | A | Horizontally sharded Aurora Postgres for petabyte workloads. | aurora limitless, limitless database | Sharded Aurora for 100M-row table |
| `aws.dynamodb` | DynamoDB | nosql | B | Key-value / document store, SS-table-backed. On-demand or provisioned RCU/WCU. | dynamodb, amazon dynamodb, key-value store, ddb | Session store, global table across 3 regions |
| `aws.dynamodb_streams` | DynamoDB Streams | nosql | I | Ordered change-data stream, 24 h retention; triggers Lambda. | dynamodb streams, ddb streams, cdc | DDB streams → Lambda → OpenSearch |
| `aws.dynamodb_accelerator` | DAX | nosql | A | In-memory cache in front of DDB, microsecond reads. | dax, dynamodb accelerator | DAX cluster for high-throughput gaming leaderboards |
| `aws.elasticache_redis` | ElastiCache Redis | cache | B | Managed Redis / Valkey clusters. | elasticache, redis, amazon redis, elasticache redis | Session cache + rate-limit store |
| `aws.elasticache_memcached` | ElastiCache Memcached | cache | I | Managed Memcached clusters. | memcached, elasticache memcached | Object cache for web tier |
| `aws.memorydb` | MemoryDB | cache | A | Durable Redis — Multi-AZ, transaction log, microsecond reads. | memorydb, durable redis | Primary datastore for financial tick data |
| `aws.neptune` | Neptune | graph | A | Graph DB — Gremlin (property graph) + SPARQL (RDF). | neptune, graph database | Fraud ring detection on Neptune |
| `aws.documentdb` | DocumentDB | nosql | I | MongoDB API-compatible managed service (pre-5.0 semantics). | documentdb, mongo compatible | CMS content on DocumentDB |
| `aws.keyspaces` | Keyspaces | nosql | A | Cassandra (CQL) compatible managed service. | keyspaces, managed cassandra | IoT time-partitioned events on Keyspaces |
| `aws.timestream` | Timestream | timeseries | A | Time-series DB with tiered memory/magnetic storage + scheduled queries. | timestream, time-series database | Industrial sensor data on Timestream |
| `aws.qldb` | QLDB | ledger | E | Immutable ledger with cryptographic verification (⚠ retiring 2025). | qldb, ledger database | Financial ledger with hash chain |
| `aws.redshift` | Redshift | warehouse | I | Columnar MPP data warehouse. RA3 nodes separate storage/compute. | redshift, data warehouse, amazon redshift | Finance mart on Redshift RA3 |
| `aws.redshift_serverless` | Redshift Serverless | warehouse | I | On-demand Redshift billed per second of compute. | redshift serverless | Ad-hoc queries, no cluster mgmt |
| `aws.redshift_spectrum` | Redshift Spectrum | warehouse | A | Query S3 directly with Redshift SQL (parquet/ORC/JSON). | redshift spectrum, external tables | Query S3 parquet from Redshift |

## Networking

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.vpc` | VPC | network | B | Isolated virtual network — CIDR block, route tables, ACLs, DHCP options. | vpc, virtual network, amazon vpc | 10.0.0.0/16 VPC with 6 subnets |
| `aws.subnet` | Subnet | network | B | CIDR slice of a VPC pinned to one AZ; public/private/isolated. | subnet, vpc subnet | 10.0.1.0/24 public subnet |
| `aws.route_table` | Route Table | network | I | Per-subnet map of destination CIDRs → target (IGW/NAT/TGW/VGW/VPC endpoint). | route table, rtb | Private subnet RTB → NAT Gateway |
| `aws.internet_gateway` | Internet Gateway | network | B | Egress/ingress door for public subnets (NAT-less SNAT for instances with EIP). | igw, internet gateway | Public subnet route 0.0.0.0/0 → IGW |
| `aws.nat_gateway` | NAT Gateway | network | I | Managed NAT; IPv4 egress for private subnets; HA within one AZ. | nat gateway, nat gw, natgw | Private subnet egress via NAT GW |
| `aws.egress_only_igw` | Egress-only IGW | network | A | IPv6-only egress door for private subnets. | egress-only, egress only igw, eigw | IPv6 private subnet egress |
| `aws.elastic_ip` | Elastic IP | network | B | Static IPv4 you own (per account, per region). | eip, elastic ip | Bastion with stable public IP |
| `aws.network_acl` | Network ACL | security | I | Stateless subnet-level packet filter (numbered rules). | nacl, network acl | Deny-list NACL on DMZ subnet |
| `aws.security_group` | Security Group | security | B | Stateful ENI-level firewall. | security group, sg | App SG allows 5432 from DB SG |
| `aws.vpc_endpoint_gateway` | Gateway VPC Endpoint | network | I | Route-table entry for S3 / DynamoDB — traffic stays on AWS backbone. | gateway endpoint, s3 endpoint, dynamodb endpoint | S3 gateway endpoint saves NAT fees |
| `aws.vpc_endpoint_interface` | Interface VPC Endpoint | network | I | ENI + PrivateLink tunnel to an AWS service without internet. | interface endpoint, privatelink endpoint | Interface endpoint for SSM, no NAT |
| `aws.privatelink` | PrivateLink | network | A | Service-consumer model where consumer ENI reaches provider NLB over AWS backbone. | privatelink, aws privatelink, vpc service | SaaS vendor exposed to my VPC via PrivateLink |
| `aws.vpc_lattice` | VPC Lattice | network | A | Service-mesh-as-a-service across VPCs/accounts with L7 auth (SigV4/OAuth). | vpc lattice, lattice, service network | Microservices across 5 VPCs via Lattice |
| `aws.transit_gateway` | Transit Gateway | network | I | Hub for VPC-to-VPC, VPN, DX transit with route tables. | tgw, transit gateway | 200-VPC hub on TGW + DX |
| `aws.transit_gateway_connect` | TGW Connect | network | A | GRE + BGP attachment for SD-WAN appliances into TGW. | tgw connect, gre attachment | Viptela cEdge peered to TGW via Connect |
| `aws.cloud_wan` | Cloud WAN | network | A | Managed global WAN policy language spanning regions/accounts/on-prem. | cloud wan, aws cloud wan | Global WAN policy with network segments |
| `aws.direct_connect` | Direct Connect | network | I | Dedicated 1/10/100 Gbps link to an AWS DX location. | direct connect, dx, dedicated line | 10 Gbps DX from Equinix LD5 |
| `aws.direct_connect_gateway` | DX Gateway | network | A | Global object connecting DX VIFs to TGWs/VGWs across regions. | dx gateway, direct connect gateway | DX Gateway linking one link to 3 TGWs |
| `aws.dx_macsec` | DX MACsec | network | E | Layer-2 encryption on 10/100 Gbps DX ports. | macsec, dx macsec, layer2 encryption | MACsec-encrypted 100 G DX for healthcare |
| `aws.vpn_site_to_site` | Site-to-Site VPN | network | I | IPsec tunnels (2 per connection) terminating on VGW or TGW. | site-to-site vpn, s2s vpn, ipsec | Branch office IPsec + BGP to TGW |
| `aws.vpn_client` | Client VPN | network | I | Managed OpenVPN endpoint with mTLS / AD / SAML auth. | client vpn, aws client vpn | Remote dev client VPN with Okta SAML |
| `aws.network_firewall` | Network Firewall | security | A | Managed stateful Suricata-based firewall as a VPC attachment. | network firewall, aws network firewall, suricata | Egress URL allow-list via Network Firewall |
| `aws.network_firewall_flow_inspection` | TLS Inspection | security | E | Deep-inspect TLS flows on Network Firewall with customer cert. | tls inspection, tls decryption | TLS inspection for outbound compliance |
| `aws.firewall_manager` | Firewall Manager | security | A | Org-wide policy enforcement across WAF / Shield / Network Firewall / SG. | firewall manager, fms, fw manager | WAF baseline across 40 accounts |
| `aws.route_53` | Route 53 | dns | B | Authoritative DNS with routing policies (failover/latency/geo/weighted). | route 53, route53, dns | Weighted A records for canary |
| `aws.route_53_resolver` | Route 53 Resolver | dns | A | Hybrid DNS resolution across on-prem ↔ VPC with inbound/outbound endpoints. | route 53 resolver, dns resolver, inbound resolver | On-prem → VPC DNS via outbound resolver |
| `aws.route_53_profile` | Resolver Profile | dns | A | Reusable DNS config applied to many VPCs/accounts. | resolver profile, r53 profile | Org-wide DNS config via Profile |
| `aws.route_53_application_recovery` | ARC | dr | E | Zonal shift + multi-region recovery readiness checks. | arc, application recovery controller, zonal shift | ARC to shift traffic off impaired AZ |
| `aws.global_accelerator` | Global Accelerator | network | A | Anycast front door — 2 static anycast IPs over AWS backbone. | global accelerator, ga, anycast | Mobile clients routed via GA to closest region |
| `aws.cloudfront` | CloudFront | cdn | B | Global CDN with ~600 PoPs, signed URLs, origin shielding, Lambda@Edge. | cloudfront, cf, amazon cloudfront, aws cdn | Static + API acceleration behind CloudFront |
| `aws.cloudfront_origin_shield` | Origin Shield | cdn | A | Mid-tier cache layer to reduce origin load. | origin shield | Shield reduces origin QPS 80% |
| `aws.cloudfront_kvs` | CloudFront KVS | cdn | E | Edge-local key-value store (≤5 MB) read from CF Functions. | cloudfront kvs, edge kv | Edge A/B bucketing via CF KVS |
| `aws.alb` | ALB | lb | B | Application Load Balancer — L7, host/path routing, gRPC, WebSocket. | alb, application load balancer, load balancer, l7 lb | ALB → ECS target groups with path routing |
| `aws.nlb` | NLB | lb | B | Network Load Balancer — L4, static IP per AZ, ultra-low latency. | nlb, network load balancer, l4 lb | NLB fronting gRPC mesh |
| `aws.gwlb` | Gateway Load Balancer | lb | A | L3 GENEVE-tunneled LB for inline security appliances. | gwlb, gateway load balancer, genvee | GWLB chaining Palo Alto appliances |
| `aws.elb_classic` | Classic ELB | lb | I | Legacy L4/L7 LB (⚠ prefer ALB/NLB). | elb, classic elb, clb | Legacy EC2-Classic app |
| `aws.api_gateway_rest` | API Gateway REST | api | I | REST API w/ resource/method model, usage plans, API keys. | api gateway, apigw, rest api gateway | REST API with usage plans for partners |
| `aws.api_gateway_http` | API Gateway HTTP | api | I | Lighter, cheaper HTTP API; JWT authorizer; VPC Link. | http api, apigw v2 | HTTP API + Lambda auth, 50% cheaper |
| `aws.api_gateway_ws` | API Gateway WebSocket | api | A | Managed WebSocket gateway with route keys. | websocket api, ws apigw | Chat backplane via WebSocket API |
| `aws.appsync` | AppSync | api | A | Managed GraphQL with resolvers to DDB/Lambda/HTTP/RDS. | appsync, graphql, managed graphql | GraphQL mobile API with real-time subs |
| `aws.cloud_map` | Cloud Map | service_discovery | A | Service-name to resource registry (DNS + API). | cloud map, service discovery | ECS service discovery via Cloud Map |

## Messaging & Streaming

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.sqs` | SQS | queue | B | Managed queues — Standard / FIFO. Up to 14-day retention. | sqs, queue, amazon sqs, message queue | Order pipeline decoupled via SQS |
| `aws.sns` | SNS | pubsub | B | Pub/sub topics — email/SMS/HTTP/Lambda/SQS/Firehose/Kinesis. | sns, pub/sub, amazon sns | SNS fan-out to SQS + Lambda |
| `aws.sns_fifo` | SNS FIFO | pubsub | I | Ordered, dedup'd SNS matched with SQS FIFO. | sns fifo, fifo topic | Ordered payment events |
| `aws.eventbridge` | EventBridge | events | I | Event bus + rules (content-based routing) with schema registry. | eventbridge, event bus, cloudwatch events | SaaS partner events via EventBridge |
| `aws.eventbridge_scheduler` | EB Scheduler | events | I | Universal cron with one-time and recurring rules, 270 targets. | eventbridge scheduler, cron, scheduler | Cron firing SQS/Lambda at 10 req/s |
| `aws.eventbridge_pipes` | EB Pipes | events | A | Source → Filter → Enrich → Target pipeline (Kinesis/MSK/SQS/DDB Streams). | eventbridge pipes, pipes | Pipe DDB Stream → filter → Step Functions |
| `aws.step_functions_standard` | Step Functions Std | orchestration | I | Workflow orchestrator — long-running (1 y), exact-once. | step functions, sfn, workflow | Order fulfillment saga |
| `aws.step_functions_express` | Step Functions Express | orchestration | I | High-throughput (100k/s), 5-min max, at-least-once. | step functions express | Short-lived IoT event routing |
| `aws.mq_activemq` | MQ ActiveMQ | broker | A | Managed ActiveMQ for JMS/AMQP legacy apps. | amazon mq, activemq | Legacy Java JMS on Amazon MQ |
| `aws.mq_rabbit` | MQ RabbitMQ | broker | A | Managed RabbitMQ with AMQP 0-9-1, MQTT, STOMP. | rabbitmq, amazon mq rabbit | Microservices broker on Amazon MQ |
| `aws.msk` | MSK | streaming | A | Managed Apache Kafka clusters (classic). | msk, managed kafka, apache kafka | 20-partition order topic on MSK |
| `aws.msk_serverless` | MSK Serverless | streaming | A | Per-throughput Kafka — no broker sizing. | msk serverless | Burst streaming without broker mgmt |
| `aws.msk_connect` | MSK Connect | streaming | A | Managed Kafka Connect workers. | msk connect, kafka connect | Debezium CDC on MSK Connect |
| `aws.kinesis_data_streams` | Kinesis Streams | streaming | I | Sharded ordered streams, millisecond replay, 1 y retention. | kinesis, kinesis streams, kds | Clickstream at 10 k/s sharded |
| `aws.kinesis_firehose` | Kinesis Firehose | streaming | I | Near-real-time delivery Stream → S3/Redshift/OpenSearch with transforms. | firehose, kinesis firehose | Logs → S3 parquet via Firehose |
| `aws.kinesis_analytics` | Managed Flink | streaming | A | Apache Flink on managed infra (formerly Kinesis Analytics). | managed flink, kinesis analytics | Anomaly detection on Managed Flink |
| `aws.kinesis_video` | Kinesis Video | streaming | A | Video ingestion SDK + HLS/WebRTC playback. | kinesis video, kvs | Home-cam stream on KVS → Rekognition |

## Observability

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.cloudwatch_metrics` | CW Metrics | metrics | B | Metrics + alarms + composite alarms + anomaly detection. | cloudwatch, cw metrics, metrics | CPU alarm triggers SNS |
| `aws.cloudwatch_logs` | CW Logs | logs | B | Log groups/streams, subscription filters → Firehose/Lambda. | cloudwatch logs, cw logs, logs | App logs → CW Logs Insights |
| `aws.cloudwatch_logs_insights` | Logs Insights | logs | I | Purpose-built query language for CW Logs. | logs insights, cwli | `stats count() by @severity` |
| `aws.cloudwatch_alarms` | CW Alarms | metrics | B | Threshold / anomaly / composite alarms → SNS/SSM/OpsCenter. | alarm, cloudwatch alarm | P99 latency alarm for canary |
| `aws.cloudwatch_synthetics` | Synthetics | uptime | A | Canaries that hit URLs / APIs with Puppeteer/Selenium. | synthetics, canary, cw synthetics | Login flow canary every 5 min |
| `aws.cloudwatch_rum` | RUM | observability | A | Real User Monitoring SDK — page load, JS errors, user journeys. | rum, real user monitoring | React RUM tracking p75 LCP |
| `aws.cloudwatch_evidently` | Evidently | experiments | A | Feature flags + experiments with statistical significance. | evidently, feature flag, experiments | A/B with Evidently on checkout |
| `aws.cloudwatch_contributor_insights` | Contributor Insights | observability | A | Top-N analyses over CW Logs / DDB. | contributor insights | Top-100 IPs hitting WAF |
| `aws.xray` | X-Ray | tracing | I | Distributed tracing with service map. | x-ray, xray, tracing | Cold-start traces for Lambda |
| `aws.cloudtrail` | CloudTrail | audit | B | API audit log; org trails write to central S3. | cloudtrail, audit log | Org trail with 7-yr S3 retention |
| `aws.cloudtrail_lake` | CloudTrail Lake | audit | A | SQL-queryable CloudTrail data store. | cloudtrail lake | Ad-hoc SQL over 90 days of events |
| `aws.config` | Config | governance | I | Resource configuration recorder + rules (managed/custom). | aws config, config rule, compliance | Config rule: S3 not public |
| `aws.audit_manager` | Audit Manager | compliance | A | Automated evidence collection for compliance frameworks. | audit manager | SOC2 Type II evidence collector |
| `aws.fis` | Fault Injection | resilience | A | Managed chaos engineering — inject CPU/net/EBS failures. | fault injection simulator, fis, chaos | FIS gameday taking AZ offline |
| `aws.resilience_hub` | Resilience Hub | resilience | A | RTO/RPO assessment + recommendations per app. | resilience hub | Hub flags non-HA RDS |

## Security & Identity

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.iam` | IAM | identity | B | Users, groups, roles, policies; 6,144-char policy limit. | iam, role, policy, amazon iam | Service role for Lambda with least priv |
| `aws.iam_identity_center` | IAM Identity Center | identity | I | SSO for AWS accounts + SaaS (formerly AWS SSO). | iam identity center, aws sso, identity center | Okta SAML to IAM Identity Center |
| `aws.iam_access_analyzer` | Access Analyzer | security | A | Finds resources shared externally + policy validation + generator. | access analyzer | Detect public S3 or shared KMS |
| `aws.iam_roles_anywhere` | Roles Anywhere | identity | A | IAM roles for on-prem workloads via X.509 from private CA. | roles anywhere, iam anywhere | On-prem server assuming IAM role |
| `aws.sts` | STS | identity | I | Short-lived creds, AssumeRole*, Federation. | sts, assume role, federation | Cross-account STS AssumeRole |
| `aws.organizations` | Organizations | governance | I | Multi-account root + OUs + SCPs + consolidated billing. | organizations, aws org, multi-account | OU hierarchy with SCP guardrails |
| `aws.scp` | SCPs | governance | A | Service Control Policies — org-level deny boundaries. | scp, service control policy | Deny region usage except us-east-1 |
| `aws.control_tower` | Control Tower | governance | A | Opinionated multi-account landing zone (LZ) with guardrails. | control tower, landing zone | Control Tower baseline with 3 OUs |
| `aws.service_catalog` | Service Catalog | governance | I | Curated CFN/Terraform products end-users self-serve. | service catalog, catalog | Approved VPC template self-service |
| `aws.resource_access_manager` | RAM | governance | I | Share subnets/TGWs/PrefixLists/License Configs across accounts. | ram, resource access manager | TGW shared via RAM to 20 accounts |
| `aws.license_manager` | License Manager | governance | A | Track BYOL (Microsoft/Oracle/SAP) entitlements. | license manager | MS SQL BYOL via License Manager |
| `aws.cognito_user_pools` | Cognito User Pools | identity | I | User directory + sign-in (OAuth/OIDC) with hosted UI. | cognito, cognito user pool, user directory | Consumer app sign-in via Cognito |
| `aws.cognito_identity_pools` | Cognito Identity Pools | identity | A | Exchange user pool JWT for temporary AWS creds. | identity pool, cognito identity | Browser-direct S3 via Cognito identity |
| `aws.verified_access` | Verified Access | security | A | ZTNA — app-level access without VPN, signals from IdP + MDM. | verified access, ava, ztna | Internal app via Verified Access, no VPN |
| `aws.verified_permissions` | Verified Permissions | security | A | Cedar policy engine for application authz. | verified permissions, cedar, avp | Fine-grained doc permissions via Cedar |
| `aws.waf` | WAF | security | I | L7 web ACLs — managed/custom rule groups, bot control. | waf, web application firewall, aws waf | WAF rate limit 2000/5min |
| `aws.shield` | Shield Standard | security | B | Always-on DDoS protection (L3/L4) for CF + Route 53 + ELB. | shield, ddos protection | Shield auto-protects CF |
| `aws.shield_advanced` | Shield Advanced | security | E | Paid DDoS mitigation + 24/7 SRT + cost protection. | shield advanced | Game launch protected by Shield Advanced |
| `aws.guardduty` | GuardDuty | security | I | Threat detection on VPC Flow / DNS / CloudTrail / S3 / EKS / Lambda. | guardduty | Alert on crypto-mining EC2 |
| `aws.detective` | Detective | security | A | Graph-based incident investigation across GuardDuty/CT/VPC-Flow. | detective | Pivot a GuardDuty finding |
| `aws.macie` | Macie | security | A | S3 sensitive-data discovery (PII/PHI/PCI) + classification. | macie | PII scan of data-lake bucket |
| `aws.inspector` | Inspector | security | I | Continuous vuln scanning for EC2/ECR/Lambda (CVSS + EPSS). | inspector, amazon inspector | ECR image vulnerability scan |
| `aws.security_hub` | Security Hub | security | A | Aggregates GuardDuty/Macie/Inspector/Config/IAM Analyzer with CSPM frameworks. | security hub, cspm | CIS / PCI dashboard |
| `aws.security_lake` | Security Lake | security | A | OCSF-formatted S3 data lake of AWS + 3rd-party security data. | security lake, ocsf lake | SIEM on top of Security Lake |
| `aws.kms` | KMS | crypto | B | Envelope encryption (CMKs); symmetric/asymmetric; 4k req/sec per key. | kms, key management, cmk | S3 SSE-KMS with customer key |
| `aws.cloudhsm` | CloudHSM | crypto | A | Dedicated FIPS 140-2 L3 HSM cluster. | cloudhsm, hsm | FIPS L3 HSM for code signing |
| `aws.private_ca` | Private CA | crypto | A | Managed X.509 private CA hierarchy. | private ca, acm pca | Mesh mTLS CA on ACM PCA |
| `aws.certificate_manager` | ACM | crypto | B | Free public TLS certs for CF/ALB/API GW; autorenew. | acm, certificate manager, tls cert | Free public cert on ALB |
| `aws.signer` | AWS Signer | crypto | A | Code signing for Lambda, IoT, container images. | signer, code signing | Sign Lambda zip via Signer |
| `aws.secrets_manager` | Secrets Manager | crypto | I | Managed secret store with rotation Lambdas for RDS / Redshift. | secrets manager, secrets | RDS password rotated every 30 d |
| `aws.systems_manager_parameter_store` | Parameter Store | config | I | Config + SecureString; cheaper than Secrets Manager. | parameter store, ssm parameter | Flat config in Parameter Store |
| `aws.artifact` | Artifact | compliance | I | Self-serve compliance reports (SOC, ISO, PCI). | aws artifact, compliance reports | Download SOC2 Type II |

## AI / ML

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.bedrock` | Bedrock | ai | B | Managed foundation models (Anthropic, Meta, Mistral, Amazon Titan, Stability). | bedrock, foundation models, aws llm | Claude on Bedrock for support bot |
| `aws.bedrock_agents` | Bedrock Agents | ai | A | Action-oriented LLM agents with knowledge bases + tools. | bedrock agent, agent | Support agent with RAG over S3 |
| `aws.bedrock_guardrails` | Bedrock Guardrails | ai | A | Content filters, PII detection, denied topics for prompts/responses. | bedrock guardrails, guardrail | PII-redacting guardrail on production agent |
| `aws.bedrock_knowledge_base` | Bedrock KB | ai | A | Managed RAG pipeline S3/Confluence → vector DB (Aurora pgvector/OpenSearch). | bedrock knowledge base, kb, rag | KB over S3 docs for sales agent |
| `aws.q_developer` | Amazon Q Developer | ai | I | In-IDE / CLI coding assistant (formerly CodeWhisperer). | q developer, codewhisperer, copilot | Q Developer in VSCode |
| `aws.q_business` | Amazon Q Business | ai | I | Enterprise chat grounded on connectors (SharePoint/Slack/Jira/S3). | q business | Q Business enterprise chat |
| `aws.sagemaker` | SageMaker | ai | I | ML platform — Studio, notebooks, training jobs, endpoints, pipelines. | sagemaker, aws sagemaker, ml platform | SageMaker training on p4d |
| `aws.sagemaker_studio` | SageMaker Studio | ai | I | Hosted JupyterLab IDE for ML. | sagemaker studio | Studio notebooks for DS team |
| `aws.sagemaker_canvas` | SageMaker Canvas | ai | B | No-code ML for analysts. | sagemaker canvas, no-code ml | Sales forecast in Canvas |
| `aws.sagemaker_ground_truth` | Ground Truth | ai | A | Managed data labeling with humans-in-the-loop. | ground truth, sagemaker ground truth, labeling | Bounding boxes via GT+Mechanical Turk |
| `aws.sagemaker_feature_store` | Feature Store | ai | A | Online + offline feature store with time-travel. | feature store, sagemaker feature store | Real-time fraud features |
| `aws.sagemaker_clarify` | Clarify | ai | A | Bias + explainability (SHAP) for datasets + models. | clarify, sagemaker clarify | Bias audit on underwriting model |
| `aws.sagemaker_neo` | Neo | ai | E | Compile models for target hardware (Jetson, Cortex-A, Intel). | sagemaker neo | Optimize ResNet for Jetson Nano |
| `aws.sagemaker_hyperpod` | HyperPod | ai | E | Resilient GPU clusters for multi-week foundation-model training. | hyperpod, sagemaker hyperpod | HyperPod for 70B LLM pretraining |
| `aws.rekognition` | Rekognition | ai | I | Vision API — faces, objects, labels, moderation, celebrities. | rekognition, computer vision | Brand-logo moderation at scale |
| `aws.textract` | Textract | ai | I | OCR + form extraction + table detection + signature detection. | textract | Invoice parsing → DynamoDB |
| `aws.comprehend` | Comprehend | ai | I | NLP — entity/PII, sentiment, topic modeling, custom classifiers. | comprehend, nlp | PII redaction pipeline |
| `aws.comprehend_medical` | Comprehend Medical | ai | A | Medical NLP — ICD-10, RxNorm, SNOMED extraction. | comprehend medical | Clinical note entity extraction |
| `aws.kendra` | Kendra | search | A | Enterprise semantic search w/ 50+ connectors. | kendra, enterprise search | Intranet search on Kendra |
| `aws.polly` | Polly | ai | B | Managed TTS with neural + generative voices. | polly, text to speech, tts | IVR prompts via Polly |
| `aws.transcribe` | Transcribe | ai | I | ASR with speaker diarization, call analytics, medical. | transcribe, asr, speech to text | Call-center transcripts |
| `aws.translate` | Translate | ai | B | NMT for 75+ languages. | translate, machine translation | Product titles for 10 locales |
| `aws.lex` | Lex | ai | I | Conversational bots (ASR+NLU+DM) on same tech as Alexa. | lex, amazon lex, chatbot | IVR bot on Lex + Connect |
| `aws.forecast` | Forecast | ai | A | Managed time-series forecasting (DeepAR, Prophet, ETS). | forecast | Demand forecast on Forecast |
| `aws.personalize` | Personalize | ai | A | Real-time recommender (HRNN, personalized ranking). | personalize | Home-page recs via Personalize |
| `aws.fraud_detector` | Fraud Detector | ai | A | Online fraud scoring with tabular ML + rules. | fraud detector | Account-takeover scoring |
| `aws.panorama` | Panorama | ai | E | Computer-vision appliance for factory floors (⚠ EoS 2026). | panorama | Shop-floor defect detection |
| `aws.braket` | Braket | quantum | E | Quantum computing: Rigetti, IonQ, QuEra, D-Wave, on-demand. | braket, quantum | QAOA on IonQ via Braket |
| `aws.deepracer` | DeepRacer | ai | E | RL 1/18 race car + league for training drivers. | deepracer | Team gameday on DeepRacer |
| `aws.omics` | HealthOmics | ai | E | Genomic data lake + workflows (WDL/Nextflow/CWL) + variant/annotation stores. | omics, genomics, healthomics | WGS pipeline on HealthOmics |

## Analytics

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.glue` | Glue | etl | I | Serverless Spark ETL + crawlers + data catalog. | glue, etl, aws glue | Glue job S3 → Redshift |
| `aws.glue_databrew` | DataBrew | etl | A | Visual data prep, 250 transformations. | databrew, glue databrew | DataBrew profiling of raw CSV |
| `aws.glue_data_catalog` | Glue Data Catalog | catalog | I | Hive-metastore-compatible catalog for Athena/EMR/Redshift Spectrum. | glue catalog, data catalog, hive metastore | Central catalog across S3 + Redshift |
| `aws.lake_formation` | Lake Formation | governance | A | Fine-grained row/column security over Glue Catalog + S3. | lake formation, lf | Lake Formation row-level security |
| `aws.athena` | Athena | query | B | Serverless Presto/Trino SQL over S3 (CSV/Parquet/ORC/JSON/Iceberg). | athena, s3 sql, serverless query | Athena CTAS to Parquet |
| `aws.athena_workgroup` | Athena Workgroup | query | A | Per-team isolation + query limits + encryption. | workgroup, athena workgroup | Finance workgroup with $/day cap |
| `aws.emr` | EMR | analytics | A | Managed Hadoop/Spark/Hive/HBase/Trino clusters. | emr, hadoop, spark, elastic map reduce | EMR Spark on Graviton |
| `aws.emr_serverless` | EMR Serverless | analytics | A | Run Spark/Hive without provisioning nodes. | emr serverless | Spark burst without cluster mgmt |
| `aws.emr_studio` | EMR Studio | analytics | A | Notebook IDE for EMR/Serverless. | emr studio | DS notebooks against EMR |
| `aws.msk` | MSK | streaming | A | See Messaging & Streaming — included here as analytics source. | msk | MSK → Flink → Iceberg |
| `aws.opensearch` | OpenSearch | search | I | Elasticsearch-compatible search + analytics (fork @ 7.10.2). | opensearch, elasticsearch, search | Log + APM search on OpenSearch |
| `aws.opensearch_serverless` | OpenSearch Serverless | search | A | On-demand vector + text search; sub-second. | opensearch serverless | Vector RAG store serverless |
| `aws.opensearch_dashboards` | OS Dashboards | search | I | Kibana-like UI. | opensearch dashboards, kibana | Trace exploration in Dashboards |
| `aws.quicksight` | QuickSight | bi | I | Self-serve BI with SPICE in-memory; Q for NLQ. | quicksight, bi, dashboards | Exec dashboards with QuickSight Q |
| `aws.data_pipeline` | Data Pipeline | etl | E | Legacy orchestration (⚠ deprecated 2025). | data pipeline | Legacy ETL on Data Pipeline |
| `aws.cloudsearch` | CloudSearch | search | E | Older search service (⚠ prefer OpenSearch). | cloudsearch | Legacy CloudSearch domain |
| `aws.finspace` | FinSpace | analytics | E | Kx/KDB+ time-series for capital markets. | finspace | FinSpace for intraday quant |
| `aws.iceberg` | Iceberg on S3 | tableformat | A | ACID table format with snapshots + schema evolution on S3 (via Glue/Athena/EMR). | iceberg, apache iceberg, lakehouse | Iceberg tables managed by Glue |

## Application Integration & Business Apps

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.amplify` | Amplify | frontend | I | Full-stack web/mobile framework + hosting + auth/data codegen. | amplify, aws amplify | Next.js app on Amplify Hosting |
| `aws.appflow` | AppFlow | integration | A | No-code SaaS↔AWS connectors (Salesforce/Slack/ServiceNow/Marketo). | appflow | Salesforce → S3 via AppFlow |
| `aws.connect` | Amazon Connect | contact_center | I | Managed omnichannel contact center (voice/chat/tasks). | amazon connect, contact center | Connect IVR + Lex agent |
| `aws.chime_sdk` | Chime SDK | communications | I | Build meeting/messaging/voice into apps. | chime sdk, meetings sdk | In-app video via Chime SDK |
| `aws.workspaces` | WorkSpaces | vdi | I | Managed Windows/Linux desktops. | workspaces, vdi | 500 contractors on WorkSpaces |
| `aws.appstream` | AppStream 2.0 | vdi | A | Stream Windows apps, not full desktops. | appstream, application streaming | CAD app via AppStream |
| `aws.workmail` | WorkMail | productivity | I | Managed email + calendar on AWS. | workmail | Small-biz email via WorkMail |
| `aws.workdocs` | WorkDocs | productivity | E | Managed document collab (⚠ EoS 2025). | workdocs | Legacy WorkDocs integration |

## Developer Tooling & IaC

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.cloudformation` | CloudFormation | iac | I | YAML/JSON template engine w/ StackSets, drift detect, change sets. | cloudformation, cfn, iac | StackSet across 20 accounts |
| `aws.cdk` | CDK | iac | I | Code-first IaC (TS/Python/Java/Go) that synthesizes to CFN. | cdk, aws cdk | CDK app with HttpApi + Lambda |
| `aws.cdktf` | CDK for Terraform | iac | A | CDK that outputs Terraform HCL/JSON. | cdktf, cdk for terraform | Multi-cloud constructs via CDKTF |
| `aws.sam` | SAM | iac | I | Serverless Application Model — CFN dialect + CLI for Lambda. | sam, serverless application model | SAM template for API+DDB |
| `aws.proton` | Proton | paas | A | Platform-engineering tool — environment + service templates. | proton | Proton templates for golden path |
| `aws.codecommit` | CodeCommit | scm | I | Managed Git hosting (⚠ feature-frozen). | codecommit | Legacy mirror of GitHub |
| `aws.codebuild` | CodeBuild | ci | I | Container-based CI build service. | codebuild | CodeBuild running integration tests |
| `aws.codedeploy` | CodeDeploy | cd | I | Deployment automation for EC2/ECS/Lambda (canary/linear/blue-green). | codedeploy | Canary deploys to Lambda |
| `aws.codepipeline` | CodePipeline | cd | I | DAG pipeline across Code* and 3rd-party tools. | codepipeline | Pipeline GitHub → Prod |
| `aws.codeartifact` | CodeArtifact | registry | I | Package registry (npm/pip/Maven/Gradle/NuGet) with upstream mirror. | codeartifact, package registry | Internal npm mirror |
| `aws.cloud9` | Cloud9 | ide | I | Cloud IDE (⚠ deprecated new accounts). | cloud9 | Shared debug env |
| `aws.cloudshell` | CloudShell | cli | B | Browser shell w/ AWS CLI, credentials preloaded. | cloudshell | Quick `aws` CLI in browser |
| `aws.systems_manager` | Systems Manager | ops | I | Umbrella: Session Mgr, Patch Mgr, Inventory, Automation, Run Command. | systems manager, ssm, patch manager | Session Manager bastion-less SSH |
| `aws.ssm_incident_manager` | Incident Manager | ops | A | Paging + runbooks + chat + postmortem. | incident manager, ssm incident | PagerDuty-style alerts via IM |
| `aws.app_config` | AppConfig | config | I | Runtime feature flags + config with validators + deploy strategies. | appconfig, app config | Gradual rollout of new flag |

## Migration

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.migration_hub` | Migration Hub | migration | I | Portfolio + progress tracker across MGN/DMS/DataSync/SMS. | migration hub | Portfolio view of migration |
| `aws.mgn` | Application Migration Service | migration | A | Block-level replication of on-prem VMs to AWS (formerly CloudEndure). | mgn, cloudendure, application migration | Lift-and-shift 200 VMs via MGN |
| `aws.dms` | Database Migration Service | migration | I | Homo/heterogeneous DB migration + CDC. | dms, db migration service | Oracle → Aurora Postgres via DMS + SCT |
| `aws.sct` | Schema Conversion Tool | migration | I | Convert schemas + code (PL/SQL → PL/pgSQL). | sct, schema conversion | SCT for Oracle → Aurora |
| `aws.application_discovery` | Application Discovery | migration | I | Agentless + agent discovery of on-prem inventory. | application discovery | DC inventory before migration |

## Edge / IoT / Robotics / Space

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.iot_core` | IoT Core | iot | B | Device gateway with MQTT/HTTP/WSS, device shadow, rules engine. | iot core, iot, mqtt broker | 50 k devices on IoT Core |
| `aws.iot_device_defender` | Device Defender | iot | A | Continuous audit + detect for device fleet. | iot device defender | Device Defender audit job |
| `aws.iot_device_management` | Device Management | iot | I | Bulk onboarding, OTA jobs, fleet indexing. | iot device management | OTA firmware rollout |
| `aws.iot_analytics` | IoT Analytics | iot | A | Time-series pipeline (pipelines + datastore) for device data. | iot analytics | Fleet analytics pipeline |
| `aws.iot_events` | IoT Events | iot | A | Detector model state-machine for device events. | iot events | Alarm state-machine per device |
| `aws.iot_sitewise` | IoT SiteWise | iot | A | Industrial data modeling (assets + properties + alarms) + SiteWise Edge. | sitewise, iot sitewise | Factory OEE dashboard |
| `aws.iot_twinmaker` | IoT TwinMaker | iot | A | Digital twin composer — entities, scenes, KG data connectors. | twinmaker, digital twin | 3D twin of warehouse |
| `aws.iot_fleetwise` | IoT FleetWise | iot | E | Automotive-specific telemetry with campaigns + DBC signal decoding. | fleetwise | Vehicle fleet data to S3 |
| `aws.iot_greengrass` | Greengrass | iot | A | Edge runtime for local compute, shadow sync, ML inference. | greengrass, iot greengrass, edge runtime | Greengrass v2 on gateway |
| `aws.iot_expresslink` | ExpressLink | iot | E | Pre-certified modules that speak IoT Core with AT-style commands. | expresslink, iot expresslink | ExpressLink module on sensor |
| `aws.iot_roborunner` | RoboRunner | iot | E | Multi-vendor robot fleet orchestration. | roborunner | AMR fleet mgmt in warehouse |
| `aws.robomaker` | RoboMaker | robotics | E | Cloud robotics sim (ROS Gazebo) + fleet mgmt (⚠ EoS 2025). | robomaker | ROS sim for drone software |
| `aws.ground_station` | Ground Station | space | E | Managed satellite antenna-as-a-service — pay per minute of contact. | ground station, satellite ground station | LEO telemetry via Ground Station |

## Media

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.mediaconvert` | MediaConvert | media | I | File-based video transcoding, profile presets, QVBR. | mediaconvert, transcoder | Ingest → MediaConvert → ABR ladder |
| `aws.medialive` | MediaLive | media | A | Live encoder with failover pipelines. | medialive | 24×7 linear channel MediaLive |
| `aws.mediapackage` | MediaPackage | media | A | Just-in-time packaging (HLS/DASH/CMAF) + origin. | mediapackage | Live + VOD packager |
| `aws.mediastore` | MediaStore | media | E | Low-latency object store for live video (⚠ deprecated 2025). | mediastore | MediaStore origin for HLS |
| `aws.mediatailor` | MediaTailor | media | A | SSAI server-side ad insertion. | mediatailor, ssai | Ad-inserted HLS playback |
| `aws.ivs` | IVS | media | A | Amazon IVS — ultra-low-latency streaming + chat. | ivs, interactive video | Live shopping with IVS |
| `aws.elemental_link` | Elemental Link | media | E | $995 SDI/HDMI appliance pushing to MediaLive. | elemental link | Stadium Link → MediaLive |

## Cost & Billing

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.cost_explorer` | Cost Explorer | cost | B | Slice/filter cost history + forecasts + anomaly detection. | cost explorer, cost | CE drilldown by service |
| `aws.budgets` | Budgets | cost | B | Per-service/tag budgets with SNS alerts + actions. | budgets, aws budgets | $ budget → disable EC2 RunInstances |
| `aws.cost_and_usage_reports` | CUR | cost | I | Hourly granular CSV/parquet in S3 (billing-source of truth). | cur, cost and usage report | CUR → Athena for chargeback |
| `aws.compute_optimizer` | Compute Optimizer | cost | I | Right-sizing for EC2/EBS/Lambda/ASG/ECS Fargate. | compute optimizer | Downsize 300 EC2 via CO |
| `aws.trusted_advisor` | Trusted Advisor | ops | I | Best-practice checks across cost/security/fault tol/limits. | trusted advisor | TA: unused EIPs |
| `aws.well_architected_tool` | WA Tool | governance | I | Self-assessment against the 6 pillars. | well architected, wa tool | WA review for prod workload |
| `aws.marketplace` | Marketplace | procurement | B | Third-party AMIs / containers / SaaS billed via AWS invoice. | marketplace, aws marketplace | Palo Alto VM-Series from MP |
| `aws.service_quotas` | Service Quotas | ops | I | View/request quota increases programmatically. | service quotas, limits | Auto-request ENI quota bump |

## Obscure / Specialized

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `aws.wickr` | Wickr | collab | E | End-to-end-encrypted messaging + file transfer + rooms. | wickr | E2EE IR comms on Wickr |
| `aws.chatbot` | Chatbot | ops | I | Chat (Slack/Teams/Chime) integration with CloudWatch + SNS + Lambda. | chatbot, slack bot | Slack ops alerts via Chatbot |
| `aws.supply_chain` | Supply Chain | business | E | Supply-chain visibility + demand planning SaaS. | aws supply chain | Inventory forecasting |
| `aws.healthlake` | HealthLake | health | E | FHIR R4 store with NLP + analytics. | healthlake | FHIR store for EHR data |
| `aws.robomaker` | RoboMaker | robotics | E | (Listed above, legacy) |  |  |
| `aws.ecs_exec` | ECS Exec | ops | I | `kubectl exec`-like shell into running ECS tasks via SSM. | ecs exec | Debug prod task securely |
| `aws.ssm_session_manager` | Session Manager | ops | I | Shell into EC2 via SSM, no SSH key or inbound port. | session manager | SM shell to private EC2 |
| `aws.application_insights` | Application Insights | observability | A | Auto-discover .NET/SAP/Oracle/SQL Server workloads + dashboards. | application insights | Observability for .NET farm |
| `aws.cloud_directory` | Cloud Directory | directory | E | Hierarchical directory service (⚠ niche). | cloud directory | Multi-schema org data |
| `aws.managed_blockchain` | Managed Blockchain | blockchain | E | Hyperledger Fabric + Ethereum nodes. | managed blockchain | Supply-chain chaincode |

---

**Entry count: ~180.** Expect additions as AWS launches; canonical names stable.
