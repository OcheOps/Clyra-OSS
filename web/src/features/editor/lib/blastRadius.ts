import type { Edge, Node } from '@xyflow/react';

// Node types where a single failure of one instance = full-service down.
// Databases, DNS, and auth are the classic offenders; caches usually
// degrade rather than kill.
const HARD_DEPENDENCY_TYPES = new Set<string>([
  'aws.rds', 'aws.aurora', 'aws.dynamodb', 'aws.documentdb',
  'aws.neptune', 'aws.elasticache', 'aws.route_53', 'aws.iam',
  'aws.cognito', 'aws.secrets_manager', 'aws.kms',
  'gcp.cloud_sql', 'gcp.spanner', 'gcp.firestore', 'gcp.cloud_dns', 'gcp.iam',
  'azure.sql_database', 'azure.cosmos_db', 'azure.dns', 'azure.ad',
  'k8s.deployment', 'k8s.service',
  'cloudflare.dns',
]);

// Node types that typically run replicated / multi-AZ by default. Their
// failure of a single instance is usually recoverable within seconds.
// Downgrades severity in the ranking pass.
const REPLICATED_TYPES = new Set<string>([
  'aws.alb', 'aws.nlb', 'aws.elb', 'aws.cloudfront', 'aws.s3',
  'aws.sqs', 'aws.sns', 'aws.eventbridge',
  'gcp.cloud_load_balancing', 'gcp.cloud_cdn', 'gcp.cloud_storage',
  'gcp.pubsub',
  'azure.load_balancer', 'azure.front_door', 'azure.blob_storage',
  'cloudflare.cdn', 'cloudflare.load_balancer',
]);

// Rough dollar/hour of full-service downtime, per node type. Numbers are
// intentionally conservative — designed to prompt attention, not to be a
// financial forecast. Multiplied by minutes-to-recovery when displayed.
//
// These are per-NODE-TYPE, not per-instance. They represent the business
// value at risk when that layer of the stack goes down. Adjust when we
// learn from actual outages.
const DOWNTIME_COST_USD_PER_HOUR: Record<string, number> = {
  // Compute layer — direct request handlers
  'aws.ec2': 500, 'aws.ecs': 500, 'aws.eks': 500, 'aws.lambda': 500,
  'aws.fargate': 500, 'aws.elastic_beanstalk': 500,
  'gcp.compute_engine': 500, 'gcp.gke': 500, 'gcp.cloud_run': 500,
  'gcp.cloud_functions': 500, 'gcp.app_engine': 500,
  'azure.vm': 500, 'azure.aks': 500, 'azure.app_service': 500,
  'azure.functions': 500,
  'k8s.deployment': 500, 'k8s.pod': 500,
  // Database layer — highest impact
  'aws.rds': 1500, 'aws.aurora': 1500, 'aws.dynamodb': 1500,
  'aws.documentdb': 1500, 'aws.neptune': 1500,
  'gcp.cloud_sql': 1500, 'gcp.spanner': 1500, 'gcp.firestore': 1500,
  'azure.sql_database': 1500, 'azure.cosmos_db': 1500,
  // Cache layer — degradation, not total outage
  'aws.elasticache': 800,
  'gcp.memorystore': 800,
  'azure.redis': 800,
  // Edge / DNS layer — full-service outage
  'aws.cloudfront': 1000, 'aws.route_53': 1500,
  'gcp.cloud_cdn': 1000, 'gcp.cloud_dns': 1500,
  'azure.front_door': 1000, 'azure.dns': 1500,
  'cloudflare.dns': 1500, 'cloudflare.cdn': 1000,
  // Auth layer
  'aws.iam': 1200, 'aws.cognito': 1200, 'aws.secrets_manager': 800,
  'gcp.iam': 1200,
  'azure.ad': 1200,
  // Storage — degraded reads/writes but usually not full down
  'aws.s3': 700, 'aws.efs': 500,
  'gcp.cloud_storage': 700,
  'azure.blob_storage': 700,
  // Queue / event — jobs stall
  'aws.sqs': 400, 'aws.sns': 400, 'aws.eventbridge': 400,
  'gcp.pubsub': 400,
  // Default when type unknown
};

const DEFAULT_DOWNTIME_COST_USD_PER_HOUR = 200;

export type Severity = 'critical' | 'high' | 'medium' | 'low';

export interface AffectedNode {
  id: string;
  distance: number;      // BFS hop count from epicenter (0 = epicenter itself)
  severity: Severity;    // computed from type + distance + replication
  nodeType: string;
  label: string;
}

export interface Mitigation {
  /** Node id this mitigation applies to. */
  nodeId: string;
  /** Short imperative — "Enable multi-AZ on RDS Postgres". */
  title: string;
  /** One-line explanation of the risk being mitigated. */
  rationale: string;
}

export interface CostImpact {
  /** Rough dollar/hour of downtime if this failure isn't mitigated. */
  usdPerHour: number;
  /** Human-readable breakdown ranked by contribution. */
  breakdown: Array<{ nodeId: string; nodeType: string; usdPerHour: number }>;
}

export interface BlastRadiusResult {
  /** All affected nodes including the epicenter, ranked by severity + distance. */
  affected: AffectedNode[];
  /** Set variant of `affected` node ids — kept for O(1) canvas rendering checks. */
  affectedIds: Set<string>;
  /** Diagram-wide SPOFs (>25% of graph reachable through this node). */
  spofs: Set<string>;
  /** Concrete "do this" list based on affected node types. */
  mitigations: Mitigation[];
  /** Downtime cost estimate for the affected set. */
  cost: CostImpact;
}

/**
 * computeBlastRadius answers "what breaks if this node fails?" and now
 * returns rich per-node severity, mitigation suggestions, and a rough
 * downtime cost estimate.
 *
 * Edge convention in Clyra diagrams: an edge points from the dependent
 * to its dependency. A node's *incoming* edges therefore point at the
 * things that would break if it dies. BFS along incoming.
 *
 * Severity ranking factors:
 *   - Node type criticality (databases > compute > queues > caches)
 *   - Distance from epicenter (direct dependents worse than transitive)
 *   - Whether the node type has native redundancy (dampened)
 *
 * O(V + E) per call including the SPOF sweep. Diagrams are sub-100 nodes;
 * no memoization needed.
 */
export function computeBlastRadius(
  nodes: Node[],
  edges: Edge[],
  sourceId: string,
): BlastRadiusResult {
  const nodeIndex = new Map<string, Node>();
  for (const n of nodes) nodeIndex.set(n.id, n);

  const incoming = new Map<string, string[]>();
  for (const edge of edges) {
    const list = incoming.get(edge.target);
    if (list) list.push(edge.source);
    else incoming.set(edge.target, [edge.source]);
  }

  // Empty result for no-focus mode — SPOFs still get computed below.
  let affected: AffectedNode[] = [];
  let affectedIds = new Set<string>();

  if (sourceId && nodeIndex.has(sourceId)) {
    const bfs = bfsDependentsWithDistance(sourceId, incoming);
    affectedIds = new Set(bfs.keys());
    affected = Array.from(bfs.entries()).map(([id, distance]) => {
      const n = nodeIndex.get(id)!;
      const nodeType = extractType(n);
      return {
        id,
        distance,
        severity: rankSeverity(nodeType, distance, id === sourceId),
        nodeType,
        label: extractLabel(n),
      };
    });
    // Severity first, then distance, then id for stable sort.
    affected.sort((a, b) => {
      const severityDelta = severityRank(a.severity) - severityRank(b.severity);
      if (severityDelta !== 0) return severityDelta;
      if (a.distance !== b.distance) return a.distance - b.distance;
      return a.id.localeCompare(b.id);
    });
  }

  const spofs = detectSpofs(nodes, incoming);
  const mitigations = suggestMitigations(affected, nodeIndex);
  const cost = estimateCost(affected);

  return { affected, affectedIds, spofs, mitigations, cost };
}

// ---------- internals ----------

function bfsDependentsWithDistance(
  from: string,
  incoming: Map<string, string[]>,
): Map<string, number> {
  const distances = new Map<string, number>([[from, 0]]);
  const queue: string[] = [from];
  while (queue.length > 0) {
    const current = queue.shift()!;
    const currentDistance = distances.get(current)!;
    const dependents = incoming.get(current);
    if (!dependents) continue;
    for (const dep of dependents) {
      if (!distances.has(dep)) {
        distances.set(dep, currentDistance + 1);
        queue.push(dep);
      }
    }
  }
  return distances;
}

function bfsCount(from: string, incoming: Map<string, string[]>): number {
  const visited = new Set([from]);
  const queue = [from];
  while (queue.length > 0) {
    const c = queue.shift()!;
    const deps = incoming.get(c);
    if (!deps) continue;
    for (const d of deps) {
      if (!visited.has(d)) {
        visited.add(d);
        queue.push(d);
      }
    }
  }
  return visited.size;
}

function detectSpofs(
  nodes: Node[],
  incoming: Map<string, string[]>,
): Set<string> {
  const threshold = nodes.length * 0.25;
  const spofs = new Set<string>();
  for (const n of nodes) {
    if (bfsCount(n.id, incoming) - 1 > threshold) spofs.add(n.id);
  }
  return spofs;
}

function extractType(node: Node): string {
  const data = node.data as { nodeType?: string } | undefined;
  return data?.nodeType ?? 'generic.service';
}

function extractLabel(node: Node): string {
  const data = node.data as { label?: string } | undefined;
  return data?.label ?? node.id;
}

function rankSeverity(
  nodeType: string,
  distance: number,
  isEpicenter: boolean,
): Severity {
  // The epicenter is by definition down. Always critical.
  if (isEpicenter) return 'critical';
  // Direct dependents of a hard dependency = critical (DB down → app down).
  if (distance === 1 && HARD_DEPENDENCY_TYPES.has(nodeType)) return 'critical';
  // Direct dependents of anything else = high (some feature broken).
  if (distance === 1) return REPLICATED_TYPES.has(nodeType) ? 'medium' : 'high';
  // 2 hops away — noticeable but not total outage.
  if (distance === 2) return REPLICATED_TYPES.has(nodeType) ? 'low' : 'medium';
  // 3+ hops = degraded experience, not visible failure.
  return 'low';
}

function severityRank(s: Severity): number {
  switch (s) {
    case 'critical': return 0;
    case 'high':     return 1;
    case 'medium':   return 2;
    case 'low':      return 3;
  }
}

function suggestMitigations(
  affected: AffectedNode[],
  nodeIndex: Map<string, Node>,
): Mitigation[] {
  const out: Mitigation[] = [];
  const seen = new Set<string>(); // dedupe on (nodeId + title)

  const add = (m: Mitigation) => {
    const key = m.nodeId + '::' + m.title;
    if (!seen.has(key)) {
      seen.add(key);
      out.push(m);
    }
  };

  for (const a of affected) {
    const label = nodeIndex.get(a.id)?.data && (nodeIndex.get(a.id)!.data as any).label
      ? (nodeIndex.get(a.id)!.data as any).label
      : a.id;

    switch (a.nodeType) {
      case 'aws.rds':
      case 'aws.aurora':
      case 'gcp.cloud_sql':
      case 'azure.sql_database':
        add({
          nodeId: a.id,
          title: `Enable multi-AZ on ${label}`,
          rationale: 'Single-AZ databases lose ~30s to minutes of writes on instance failure. Multi-AZ automates failover to a hot standby.',
        });
        break;
      case 'aws.elasticache':
      case 'gcp.memorystore':
      case 'azure.redis':
        add({
          nodeId: a.id,
          title: `Add a replica to ${label}`,
          rationale: 'Cache warmth is expensive to rebuild. A read replica keeps hot data available if the primary drops.',
        });
        break;
      case 'aws.dynamodb':
      case 'gcp.firestore':
      case 'azure.cosmos_db':
        add({
          nodeId: a.id,
          title: `Enable global tables on ${label}`,
          rationale: 'Regional outages take down single-region NoSQL. Multi-region replication survives a full AWS region loss.',
        });
        break;
      case 'aws.route_53':
      case 'gcp.cloud_dns':
      case 'azure.dns':
      case 'cloudflare.dns':
        add({
          nodeId: a.id,
          title: `Configure a secondary DNS provider alongside ${label}`,
          rationale: 'DNS is your kill switch. A single provider means one BGP incident takes you fully offline.',
        });
        break;
      case 'aws.ec2':
      case 'gcp.compute_engine':
      case 'azure.vm':
        add({
          nodeId: a.id,
          title: `Run ${label} behind an autoscaling group across ≥2 AZs`,
          rationale: 'Single-instance compute is a needless SPOF. Autoscaling replaces failed instances and spreads across zones.',
        });
        break;
      case 'k8s.deployment':
        add({
          nodeId: a.id,
          title: `Set replicas >= 2 and add pod anti-affinity for ${label}`,
          rationale: 'A single-pod deployment is a rolling deploy away from downtime. Anti-affinity keeps replicas on different nodes.',
        });
        break;
      case 'aws.lambda':
      case 'aws.fargate':
      case 'aws.ecs':
      case 'aws.eks':
        // Compute-tier already replicated by default — flag only if it's
        // the epicenter of a chain to something worse.
        if (a.severity === 'critical' && a.distance === 0) {
          add({
            nodeId: a.id,
            title: `Add cross-region failover for ${label}`,
            rationale: 'The compute layer is redundant within a region, but a full-region outage still kills it. Warm standby in a second region.',
          });
        }
        break;
    }
  }
  return out;
}

function estimateCost(affected: AffectedNode[]): CostImpact {
  const breakdown: CostImpact['breakdown'] = [];
  let total = 0;
  for (const a of affected) {
    const rate = DOWNTIME_COST_USD_PER_HOUR[a.nodeType] ?? DEFAULT_DOWNTIME_COST_USD_PER_HOUR;
    // Weight by severity — 'critical' = full rate, degrading down.
    const weight = a.severity === 'critical' ? 1
      : a.severity === 'high' ? 0.6
      : a.severity === 'medium' ? 0.3
      : 0.1;
    const cost = Math.round(rate * weight);
    if (cost > 0) {
      breakdown.push({ nodeId: a.id, nodeType: a.nodeType, usdPerHour: cost });
      total += cost;
    }
  }
  breakdown.sort((a, b) => b.usdPerHour - a.usdPerHour);
  return { usdPerHour: total, breakdown };
}
