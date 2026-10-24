# Clyra Component & Icon Catalog

Exhaustive, expert-level catalog of cloud, networking, data, enterprise, and
edge infrastructure components used in real production architectures.
Source of truth for:

- the icon registry (`web/src/lib/icons.ts`)
- the cost catalog (`web/src/lib/costCatalog.ts`)
- the validation rule engine (`web/src/lib/architectureRules.ts`)
- IaC generation service-name mapping
- prompt → component normalization (fuzzy match against `aliases`)

## Files

| File | Scope | Depth |
|---|---|---|
| [aws.md](aws.md) | AWS — every major + obscure service | very deep |
| [azure.md](azure.md) | Azure / Entra / Fabric | deep |
| [gcp.md](gcp.md) | Google Cloud / Anthos / Vertex | deep |
| [k8s.md](k8s.md) | Kubernetes internals, CRDs, mesh, add-ons | internal |
| [networking.md](networking.md) | BGP, MPLS, SD-WAN, DNS, zero trust | wire-level |
| [data.md](data.md) | Kafka, Spark, Databricks, Iceberg, etc. | expert |
| [enterprise.md](enterprise.md) | ERP, CRM, IdP, SIEM, observability | enterprise |
| [edge.md](edge.md) | IoT, 5G MEC, satellite, hybrid | edge/space |

## Entry format

Every domain file uses the same table format. One row per component.

```
| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
```

- **Canonical** — the `provider.snake_case` key (e.g. `aws.ec2`, `gcp.cloud_run`).
  This is what `icons.ts` / `costCatalog.ts` / rules reference.
- **Display** — label shown on nodes.
- **Sub** — subcategory inside the provider (compute, network, storage, ai, …).
- **Lvl** — experience tier: **B**eginner / **I**ntermediate / **A**dvanced / **E**xpert.
  Used by prompt normalizer to bias toward familiar components in general
  queries, expose obscure ones only on explicit mention.
- **Description** — short but precise. Mentions positioning relative to
  siblings (e.g. "ALB vs NLB vs Classic ELB").
- **Aliases** — comma-separated strings the LLM / user might type. These
  power fuzzy matching. Include full names, abbreviations, marketing names,
  common typos, colloquial names.
- **Example** — one concise prompt fragment showing realistic usage.

## Levels explained

- **B (Beginner)** — every cloud user will touch this (EC2, S3, RDS, Lambda).
  Safe to surface by default in suggestions and templates.
- **I (Intermediate)** — architects know these (Transit Gateway, PrivateLink,
  EventBridge, Step Functions).
- **A (Advanced)** — senior engineers (Nitro Enclaves, Graviton, VPC Lattice,
  MACsec, Neptune, Timestream, MSK Connect, EKS Anywhere).
- **E (Expert)** — the "I didn't know that existed" tier. Surface only on
  explicit mention: Ground Station, Braket, Snowcone SBE, DeepRacer, Omics,
  Assured Workloads, Confidential Space, Azure Orbital, Distributed Cloud
  Air-Gapped.

## Naming convention

```
<provider>.<service_slug>
```

- `<provider>` ∈ {`aws`, `azure`, `gcp`, `k8s`, `net`, `data`, `ent`, `edge`, `iot`}
- `<service_slug>` — lowercase snake_case, no spaces, no marketing articles.
  `aws.api_gateway`, not `aws.amazon_api_gateway`.
- Sub-resources use a dot: `k8s.workload.deployment`, `k8s.control.etcd`.
- Variants use underscores: `azure.sql_managed_instance`, `aws.lambda_snapstart`.

When AWS renames a service (e.g. "AWS SSO" → "IAM Identity Center"),
keep the canonical name stable and add the old name as an alias.

## Patterns library

See [patterns.md](patterns.md) for common architectural patterns expressed
as combinations of canonical components. Patterns are what the prompt
normalizer collapses multi-sentence specs into (e.g. "3-tier web app" →
`{ALB → ECS Fargate → RDS Multi-AZ}` + observability + secrets + WAF).

## Icon system build

1. Every entry in these files gets a Lucide icon (or a custom SVG) in
   `web/src/lib/icons.ts`, keyed by the canonical name.
2. The icon registry has five columns: `canonical`, `display`, `color`,
   `icon_name`, `family`. Color is inherited from the provider's brand color
   (see the `FAMILY_COLORS` map).
3. When a canonical is missing an icon, the registry falls back to a generic
   family icon (the cloud logo). Safer than a 404.
4. The palette/search UI uses the `aliases` field to make "loadbalancer",
   "load balancer", "LB", "alb" all hit `aws.alb`.

## Prompt normalization

The `normalize_prompt` routine (to be added to the diagram service) runs
this pipeline:

1. **Tokenize** the prompt into candidate service phrases (2-4 word windows).
2. **Fuzzy match** each candidate against the union of every `aliases` list
   using trigram similarity (threshold 0.72). Higher weight for longer matches.
3. **Context bias** — if the prompt contains "aws", "amazon", "nitro",
   raise AWS matches by 1.3×. Same for "azure", "gcp"/"google cloud".
4. **Level bias** — expert-level entries require a direct token hit; they
   don't fire from generic phrasing.
5. **Pattern collapse** — phrases like "3-tier web app", "event-driven
   microservices", "data lake", "streaming ETL" expand to multi-component
   patterns from `patterns.md`.
6. **Deduplicate** — if both `aws.load_balancer` and `aws.alb` match, keep
   the more specific one.

## Counts

| Domain | Entries |
|---|---|
| AWS | ~180 |
| Azure | ~110 |
| GCP | ~100 |
| Kubernetes | ~60 |
| Networking | ~70 |
| Data | ~55 |
| Enterprise | ~45 |
| Edge / IoT | ~40 |
| **Total** | **~660** |

## Maintenance

- Quarterly pricing review for `costCatalog.ts` → regenerate `usd` values.
- New services: first add here, then propagate to icons, cost catalog,
  and rules. Don't let the three drift.
- Deprecations: mark the canonical with `⚠ deprecated` in the description
  but keep it — existing diagrams reference it.
