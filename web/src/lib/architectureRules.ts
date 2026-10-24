// Architecture validation rules. Think "ESLint for cloud diagrams."
//
// Every rule takes the current graph (nodes + edges) and returns zero or more
// findings. Findings carry a pillar + severity so the UI can group and sort.
//
// Rules are intentionally presence-based — they fire on "you have X but not Y"
// patterns, since node-level metadata (encryption flags, etc.) isn't yet
// captured by the generation model. When it is, we'll add config-level rules.

import type { Edge, Node } from '@xyflow/react';

export type Pillar =
  | 'reliability'
  | 'security'
  | 'cost'
  | 'operational'
  | 'performance';

export type Severity = 'high' | 'medium' | 'low' | 'info';

export interface Finding {
  id: string;
  pillar: Pillar;
  severity: Severity;
  title: string;
  detail: string;
  nodeIds: string[];
  doc?: { label: string; url: string };
}

interface Graph {
  nodes: Node[];
  edges: Edge[];
}

type RuleFn = (g: Graph) => Finding[];

const DB_TYPES = new Set([
  'aws.rds', 'aws.aurora', 'aws.dynamodb', 'aws.redshift',
  'gcp.cloud_sql', 'gcp.spanner', 'gcp.firestore', 'gcp.bigquery',
  'azure.sql', 'azure.cosmos_db',
]);

const PUBLIC_EDGE_TYPES = new Set([
  'aws.alb', 'aws.nlb', 'aws.api_gateway', 'aws.cloudfront',
  'gcp.load_balancer', 'gcp.cloud_cdn',
  'azure.load_balancer', 'azure.application_gateway', 'azure.cdn',
]);

const STORAGE_TYPES = new Set([
  'aws.s3', 'gcp.cloud_storage', 'azure.blob_storage',
]);

const LAMBDA_TYPES = new Set([
  'aws.lambda', 'gcp.cloud_functions', 'azure.functions',
]);

const MONITORING_TYPES = new Set([
  'aws.cloudwatch', 'gcp.cloud_monitoring', 'gcp.cloud_logging', 'azure.monitor',
]);

const AUDIT_TYPES = new Set([
  'aws.cloudtrail', 'aws.config',
]);

const WAF_TYPES = new Set([
  'aws.waf', 'aws.shield',
  'azure.application_gateway', // WAF baked in
  // GCP Cloud Armor isn't in our icon set yet; leave unchecked for GCP.
]);

const CDN_TYPES = new Set([
  'aws.cloudfront', 'gcp.cloud_cdn', 'azure.cdn',
]);

const NAT_TYPES = new Set(['aws.nat_gateway']);

const SECRETS_TYPES = new Set([
  'aws.secrets_manager', 'gcp.secret_manager', 'azure.key_vault',
]);

const IAM_TYPES = new Set([
  'aws.iam', 'aws.cognito', 'gcp.iam', 'azure.ad',
]);

const BACKUP_TYPES = new Set(['aws.backup']);

const CACHE_TYPES = new Set([
  'aws.elasticache', 'gcp.memorystore', 'azure.redis',
]);

const QUEUE_TYPES = new Set([
  'aws.sqs', 'aws.sns', 'aws.eventbridge',
  'gcp.pubsub',
  'azure.service_bus', 'azure.event_hubs',
]);

function nodeType(n: Node): string {
  const d = n.data as { nodeType?: string } | undefined;
  return d?.nodeType ?? n.type ?? '';
}

function hasAnyOf(nodes: Node[], types: Set<string>): boolean {
  return nodes.some((n) => types.has(nodeType(n)));
}

function nodesOfTypes(nodes: Node[], types: Set<string>): Node[] {
  return nodes.filter((n) => types.has(nodeType(n)));
}

// ── Rule library ───────────────────────────────────────────────────
// Each rule is one function with one concern. Cheap to add, cheap to remove.

const rules: RuleFn[] = [
  // Reliability ────────────────────────────────────────────────────
  ({ nodes }) => {
    const dbs = nodesOfTypes(nodes, DB_TYPES);
    if (dbs.length === 0) return [];
    if (hasAnyOf(nodes, BACKUP_TYPES)) return [];
    return [{
      id: 'REL001',
      pillar: 'reliability',
      severity: 'medium',
      title: 'Database present but no backup resource',
      detail: 'You have a database but no AWS Backup / equivalent in the design. Snapshot strategy should be explicit.',
      nodeIds: dbs.map((n) => n.id),
      doc: {
        label: 'AWS Well-Architected — Reliability',
        url: 'https://docs.aws.amazon.com/wellarchitected/latest/reliability-pillar/',
      },
    }];
  },

  // Security ───────────────────────────────────────────────────────
  ({ nodes }) => {
    const publicEdges = nodesOfTypes(nodes, PUBLIC_EDGE_TYPES);
    if (publicEdges.length === 0) return [];
    if (hasAnyOf(nodes, WAF_TYPES)) return [];
    return [{
      id: 'SEC001',
      pillar: 'security',
      severity: 'high',
      title: 'Internet-facing entry points without WAF',
      detail: 'Public load balancer or API gateway without a WAF leaves Layer 7 attacks unmitigated (SQL injection, XSS, credential stuffing).',
      nodeIds: publicEdges.map((n) => n.id),
      doc: {
        label: 'AWS Well-Architected — Security (SEC5)',
        url: 'https://docs.aws.amazon.com/wellarchitected/latest/security-pillar/',
      },
    }];
  },

  ({ nodes }) => {
    const dbs = nodesOfTypes(nodes, DB_TYPES);
    if (dbs.length === 0) return [];
    if (hasAnyOf(nodes, SECRETS_TYPES)) return [];
    return [{
      id: 'SEC002',
      pillar: 'security',
      severity: 'medium',
      title: 'No secrets manager / key vault',
      detail: 'Database credentials almost certainly live somewhere. Add Secrets Manager / Key Vault so rotation and access control are first-class.',
      nodeIds: dbs.map((n) => n.id),
    }];
  },

  ({ nodes }) => {
    const publicFacing = nodesOfTypes(nodes, PUBLIC_EDGE_TYPES);
    if (publicFacing.length === 0) return [];
    if (hasAnyOf(nodes, IAM_TYPES)) return [];
    return [{
      id: 'SEC003',
      pillar: 'security',
      severity: 'medium',
      title: 'Public endpoints without an identity layer',
      detail: 'Public-facing surface without IAM / Cognito / Azure AD means authentication is implicit or ad hoc. Make it explicit.',
      nodeIds: publicFacing.map((n) => n.id),
    }];
  },

  // Cost ───────────────────────────────────────────────────────────
  ({ nodes }) => {
    const nats = nodesOfTypes(nodes, NAT_TYPES);
    if (nats.length === 0) return [];
    // VPC endpoints aren't modelled as separate nodes yet — flag as info
    // so the user considers it. Bump to low once we detect endpoints.
    return [{
      id: 'COST001',
      pillar: 'cost',
      severity: 'low',
      title: 'NAT Gateway present — verify VPC endpoints',
      detail: 'NAT Gateways are billed per-hour and per-GB. Route S3 and DynamoDB traffic through VPC endpoints to avoid NAT data-transfer charges.',
      nodeIds: nats.map((n) => n.id),
    }];
  },

  // Operational excellence ─────────────────────────────────────────
  ({ nodes }) => {
    if (nodes.length < 2) return [];
    if (hasAnyOf(nodes, MONITORING_TYPES)) return [];
    return [{
      id: 'OPS001',
      pillar: 'operational',
      severity: 'medium',
      title: 'No monitoring component',
      detail: 'CloudWatch / Cloud Monitoring / Azure Monitor is missing. Without metrics + logs you cannot detect or debug production incidents.',
      nodeIds: [],
    }];
  },

  ({ nodes }) => {
    const awsNodes = nodes.filter((n) => nodeType(n).startsWith('aws.'));
    if (awsNodes.length === 0) return [];
    if (hasAnyOf(nodes, AUDIT_TYPES)) return [];
    return [{
      id: 'OPS002',
      pillar: 'operational',
      severity: 'low',
      title: 'No audit trail for AWS resources',
      detail: 'CloudTrail is absent. Every compliance framework (SOC2, HIPAA, PCI) expects an audit log of API activity.',
      nodeIds: [],
    }];
  },

  ({ nodes }) => {
    const lambdas = nodesOfTypes(nodes, LAMBDA_TYPES);
    if (lambdas.length === 0) return [];
    if (hasAnyOf(nodes, QUEUE_TYPES)) return [];
    return [{
      id: 'OPS003',
      pillar: 'operational',
      severity: 'low',
      title: 'Serverless functions without a queue / DLQ',
      detail: 'Async Lambda / Functions / Cloud Functions need an SQS / Pub/Sub / Service Bus DLQ to avoid silent message loss on failure.',
      nodeIds: lambdas.map((n) => n.id),
    }];
  },

  // Performance ────────────────────────────────────────────────────
  ({ nodes }) => {
    const buckets = nodesOfTypes(nodes, STORAGE_TYPES);
    if (buckets.length === 0) return [];
    if (hasAnyOf(nodes, CDN_TYPES)) return [];
    return [{
      id: 'PERF001',
      pillar: 'performance',
      severity: 'low',
      title: 'Object storage without a CDN in front',
      detail: 'If the bucket serves public assets, fronting it with CloudFront / Cloud CDN / Azure CDN reduces latency and egress cost.',
      nodeIds: buckets.map((n) => n.id),
    }];
  },

  ({ nodes }) => {
    const dbs = nodesOfTypes(nodes, DB_TYPES);
    if (dbs.length === 0) return [];
    if (hasAnyOf(nodes, CACHE_TYPES)) return [];
    return [{
      id: 'PERF002',
      pillar: 'performance',
      severity: 'info',
      title: 'No caching layer in front of the database',
      detail: 'Consider ElastiCache / Memorystore / Azure Cache for read-heavy workloads to reduce DB load and latency.',
      nodeIds: dbs.map((n) => n.id),
    }];
  },
];

export interface ValidationReport {
  findings: Finding[];
  countsBySeverity: Record<Severity, number>;
}

const SEVERITY_ORDER: Record<Severity, number> = {
  high: 0,
  medium: 1,
  low: 2,
  info: 3,
};

export function validate(nodes: Node[], edges: Edge[]): ValidationReport {
  const findings = rules.flatMap((fn) => fn({ nodes, edges }));
  findings.sort((a, b) => SEVERITY_ORDER[a.severity] - SEVERITY_ORDER[b.severity]);

  const countsBySeverity: Record<Severity, number> = { high: 0, medium: 0, low: 0, info: 0 };
  for (const f of findings) countsBySeverity[f.severity]++;

  return { findings, countsBySeverity };
}
