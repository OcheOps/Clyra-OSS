# Data Engineering & Analytics — Component Catalog

Third-party and OSS platforms engineers pair with cloud primitives.
Kafka, Spark, Flink, Databricks, table formats, orchestration, ML platforms.

## Streaming platforms

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `data.kafka_broker` | Kafka Broker | streaming | A | Apache Kafka node; stores partitions on log segments. | kafka, apache kafka, broker | 6-broker MSK cluster |
| `data.kafka_controller_kraft` | Kafka Controller (KRaft) | streaming | A | Kafka 3.3+ Zookeeper replacement — Raft quorum. | kraft, kafka controller | Zk-less cluster |
| `data.kafka_zookeeper` | Zookeeper (legacy) | streaming | A | Pre-KRaft coordination service. | zookeeper, kafka zk | Legacy |
| `data.kafka_connect` | Kafka Connect | streaming | A | Distributed connector runtime; Source + Sink. | kafka connect | Debezium CDC |
| `data.kafka_streams` | Kafka Streams | streaming | A | Library for stateful stream processing on Kafka. | kafka streams | Windowed join in JVM |
| `data.ksqldb` | ksqlDB | streaming | A | SQL interface over Kafka Streams. | ksqldb, ksql | SQL-on-stream |
| `data.schema_registry` | Schema Registry | streaming | A | Versioned Avro / JSON / Protobuf schemas with compatibility. | schema registry, confluent sr | Avro evolution |
| `data.kafka_mirrormaker2` | MirrorMaker 2 | streaming | A | Kafka → Kafka replication across clusters/regions. | mirrormaker, mm2 | Active-active regions |
| `data.strimzi` | Strimzi | streaming | A | K8s operator for Kafka. | strimzi | Kafka on K8s |
| `data.apache_pulsar` | Apache Pulsar | streaming | A | Multi-tenant pub/sub with tiered storage (BookKeeper). | pulsar | Geo-replication |
| `data.redpanda` | Redpanda | streaming | A | Kafka-compat broker in C++; Raft-based; no ZK. | redpanda | Low-latency Kafka |
| `data.nats` | NATS | messaging | A | Lightweight pub/sub + JetStream streams. | nats | Edge messaging |
| `data.rabbitmq` | RabbitMQ | broker | I | AMQP 0-9-1 broker (+MQTT/STOMP). | rabbitmq, amqp | Task queue |
| `data.activemq_artemis` | ActiveMQ Artemis | broker | A | Next-gen AMQ (JMS 2.0 + AMQP + MQTT). | activemq, artemis | Enterprise JMS |
| `data.warpstream` | WarpStream | streaming | E | Kafka on S3 — no disks, cheaper at rest. | warpstream | Cheap streaming |

## Batch / distributed compute

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `data.spark_driver` | Spark Driver | analytics | A | Driver program — SparkContext, DAG scheduler, task coordinator. | spark driver | Driver on EMR master |
| `data.spark_executor` | Spark Executor | analytics | A | JVM runtime for tasks on workers; shuffle service. | spark executor | 100 executors |
| `data.spark_catalyst` | Spark Catalyst | analytics | A | Query optimizer — logical + physical plans, AQE. | catalyst, spark optimizer | Pushdown predicates |
| `data.spark_tungsten` | Spark Tungsten | analytics | E | Off-heap memory + whole-stage codegen. | tungsten | Perf engine |
| `data.spark_structured_streaming` | Structured Streaming | streaming | A | Micro-batch / continuous streaming on Spark SQL. | structured streaming | Sliding window |
| `data.pyspark` | PySpark | analytics | I | Python DataFrame API for Spark. | pyspark | DS notebook |
| `data.spark_connect` | Spark Connect | analytics | A | Decoupled client/server protocol (Spark 3.4+). | spark connect | Thin client + remote cluster |
| `data.flink_job_manager` | Flink JobManager | streaming | A | Coordinator; checkpoints + schedules. | flink jobmanager | JM HA |
| `data.flink_task_manager` | Flink TaskManager | streaming | A | Runs task slots; manages state backend. | flink taskmanager | TM slots |
| `data.flink_checkpoint` | Flink Checkpoint | streaming | A | Periodic snapshots of operator state for fault tolerance. | checkpoint, flink checkpoint | Exactly-once |
| `data.flink_savepoint` | Flink Savepoint | streaming | A | Manual snapshot for rolling upgrades / state migration. | savepoint | Upgrade to v1.18 |
| `data.flink_sql` | Flink SQL | streaming | A | SQL over bounded/unbounded tables. | flink sql | SQL on stream |
| `data.apache_beam` | Apache Beam | analytics | A | Unified batch/stream SDK with runner abstraction (Dataflow/Flink/Spark). | beam, apache beam | Portable pipeline |
| `data.apache_storm` | Apache Storm | streaming | E | Legacy streaming (⚠ superseded). | storm | Legacy |
| `data.ray` | Ray | compute | A | Distributed Python — actors + tasks + serve + data + tune. | ray | ML training |
| `data.dask` | Dask | compute | A | Parallel Python — arrays, dataframes, delayed. | dask | Scale pandas |

## Table formats & lakehouse

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `data.iceberg` | Apache Iceberg | table_format | A | Snapshots, schema evolution, branching, time travel. | iceberg, apache iceberg | Iceberg + Glue + Athena |
| `data.delta_lake` | Delta Lake | table_format | A | Parquet + transaction log (Databricks-origin). | delta lake, delta | DLT pipelines |
| `data.hudi` | Apache Hudi | table_format | A | Upsert-friendly format (COW + MOR). | hudi, apache hudi | CDC into Hudi |
| `data.parquet` | Parquet | format | B | Columnar file format (Apache). | parquet, apache parquet | Data-lake default |
| `data.orc` | ORC | format | I | Columnar format (Hive-origin). | orc | Hive warehouses |
| `data.avro` | Avro | format | I | Row-oriented w/ embedded schema. | avro | Kafka topics |
| `data.arrow` | Apache Arrow | format | A | In-memory columnar standard; IPC, Flight. | arrow, apache arrow | Zero-copy IPC |
| `data.flight` | Arrow Flight | format | A | Arrow IPC over gRPC for fast data transfer. | flight, arrow flight | BI tool backend |

## Warehouses & engines

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `data.snowflake` | Snowflake | warehouse | I | Multi-cluster shared-data warehouse. | snowflake | Finance mart |
| `data.databricks` | Databricks | lakehouse | I | Spark + Delta + Unity Catalog + MLflow. | databricks | Lakehouse platform |
| `data.databricks_unity_catalog` | Unity Catalog | catalog | A | Row-level + column-level + lineage across lakehouse. | unity catalog | Multi-workspace governance |
| `data.databricks_dlt` | Delta Live Tables | etl | A | Declarative streaming ETL on Delta. | dlt, delta live tables | Bronze/silver/gold DLT |
| `data.databricks_photon` | Photon | engine | A | Native (C++) vectorized engine for Spark SQL. | photon | Faster SQL |
| `data.trino` | Trino | engine | A | Distributed MPP SQL over many connectors (formerly PrestoSQL). | trino, presto | Federated SQL |
| `data.presto` | Presto (old fork) | engine | E | Original Presto (Facebook). | presto legacy | Legacy Presto |
| `data.clickhouse` | ClickHouse | warehouse | A | Columnar OLAP DB; vectorized execution. | clickhouse | Real-time analytics |
| `data.druid` | Apache Druid | warehouse | A | Time-series OLAP with pre-aggregated segments. | druid, apache druid | Event analytics |
| `data.pinot` | Apache Pinot | warehouse | A | Low-latency OLAP (LinkedIn). | pinot, apache pinot | Feed ranking |
| `data.starrocks` | StarRocks | warehouse | A | MPP OLAP fork of Apache Doris. | starrocks | Self-hosted BigQuery |
| `data.dremio` | Dremio | engine | A | Lakehouse SQL over Iceberg + Delta + Parquet. | dremio | Self-service analytics |
| `data.firebolt` | Firebolt | warehouse | A | Low-latency warehouse w/ indexes + Parquet. | firebolt | Dashboard backend |
| `data.materialize` | Materialize | streaming | A | Real-time SQL views on streams (timely dataflow). | materialize | Real-time analytics |
| `data.risingwave` | RisingWave | streaming | A | Streaming DB (Materialize competitor). | risingwave | Streaming SQL |
| `data.singlestore` | SingleStore | db | A | HTAP DB (formerly MemSQL). | singlestore, memsql | OLTP + OLAP |

## Orchestration

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `data.airflow` | Apache Airflow | orchestration | I | DAG orchestrator w/ scheduler + workers + DB. | airflow | Daily DAGs |
| `data.prefect` | Prefect | orchestration | I | Modern workflow (Python-first, Cloud v3). | prefect | ML pipelines |
| `data.dagster` | Dagster | orchestration | A | Asset-oriented pipelines with types + testing. | dagster | Data-product pipelines |
| `data.temporal` | Temporal | orchestration | A | Durable execution (workflows + activities, infinite retries). | temporal | Long-running sagas |
| `data.luigi` | Luigi | orchestration | E | Spotify original (⚠ less used). | luigi | Legacy |
| `data.dbt` | dbt | etl | I | SQL-based ELT with tests + lineage. | dbt | Analytics engineering |
| `data.fivetran` | Fivetran | etl | I | SaaS ELT connectors. | fivetran | Zero-code ingest |
| `data.airbyte` | Airbyte | etl | I | Open-source ELT. | airbyte | Self-hosted ELT |
| `data.nifi` | Apache NiFi | etl | A | Flow-based programming. | nifi, apache nifi | Flow routing |

## Governance & quality

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `data.great_expectations` | Great Expectations | quality | A | Data assertions + docs. | great expectations, ge | Column expectations |
| `data.soda` | Soda | quality | A | Data observability + quality tests. | soda | SodaCL checks |
| `data.monte_carlo` | Monte Carlo | observability | A | Data observability SaaS (freshness, volume, schema). | monte carlo | Data SLO |
| `data.collibra` | Collibra | governance | E | Enterprise data governance platform. | collibra | Data steward |
| `data.alation` | Alation | governance | E | Data catalog + governance. | alation | Catalog |
| `data.datahub` | DataHub | governance | A | Open-source data catalog (LinkedIn). | datahub | Lineage + catalog |
| `data.atlan` | Atlan | governance | A | Modern data workspace / catalog. | atlan | Active metadata |
| `data.open_lineage` | OpenLineage | governance | A | Vendor-neutral lineage spec. | open lineage | Lineage events |

## ML platforms

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `data.mlflow` | MLflow | ml | A | Tracking, model registry, deployment. | mlflow | Experiment tracking |
| `data.kubeflow` | Kubeflow | ml | A | K8s-native ML platform. | kubeflow | Training on k8s |
| `data.feast` | Feast | ml | A | Feature store (offline + online). | feast | Features for serving |
| `data.weights_and_biases` | Weights & Biases | ml | I | Experiment tracking SaaS. | wandb, w&b | Model tracking |
| `data.pachyderm` | Pachyderm | ml | E | Data versioning + pipelines. | pachyderm | DVC-like pipelines |

---

**Entry count: ~55.** Core OSS + key commercial platforms engineers compose alongside cloud services.
