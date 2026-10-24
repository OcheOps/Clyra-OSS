import type { Diagram, DiagramNode, DiagramEdge, DiagramGroup, EdgeStyle } from '@/types/diagram';
import type { Node, Edge } from '@xyflow/react';
import { COMPONENT_IDS } from './generated/componentIds';

// --- Known type registry (mirrors backend normalizer.go) ---

const KNOWN_TYPES = new Set([
  // AWS
  'aws.vpc', 'aws.subnet', 'aws.ec2', 'aws.ecs', 'aws.eks',
  'aws.lambda', 'aws.alb', 'aws.nlb', 'aws.elb', 'aws.api_gateway',
  'aws.cloudfront', 'aws.s3', 'aws.rds', 'aws.dynamodb',
  'aws.elasticache', 'aws.route_53', 'aws.iam', 'aws.waf',
  'aws.nat_gateway', 'aws.internet_gateway', 'aws.sqs', 'aws.sns',
  'aws.eventbridge', 'aws.step_functions', 'aws.cloudwatch',
  'aws.secrets_manager', 'aws.systems_manager', 'aws.ecr',
  'aws.cognito', 'aws.kinesis', 'aws.transit_gateway',
  'aws.direct_connect', 'aws.config', 'aws.guardduty',
  // GCP
  'gcp.vpc', 'gcp.compute_engine', 'gcp.gke', 'gcp.cloud_run',
  'gcp.app_engine', 'gcp.cloud_functions', 'gcp.cloud_load_balancing',
  'gcp.cloud_cdn', 'gcp.cloud_storage', 'gcp.cloud_sql',
  'gcp.bigquery', 'gcp.memorystore', 'gcp.pubsub', 'gcp.cloud_dns',
  'gcp.secret_manager', 'gcp.firestore', 'gcp.cloud_armor',
  // Azure
  'azure.vnet', 'azure.subnet', 'azure.vm', 'azure.aks',
  'azure.app_service', 'azure.functions', 'azure.application_gateway',
  'azure.load_balancer', 'azure.front_door', 'azure.blob_storage',
  'azure.sql_database', 'azure.cosmos_db', 'azure.redis',
  'azure.dns', 'azure.key_vault', 'azure.monitor',
  'azure.active_directory', 'azure.service_bus',
  // Kubernetes
  'k8s.cluster', 'k8s.node', 'k8s.namespace', 'k8s.deployment',
  'k8s.pod', 'k8s.service', 'k8s.ingress', 'k8s.configmap',
  'k8s.secret', 'k8s.statefulset', 'k8s.daemonset', 'k8s.job',
  // OCI
  'oci.vcn', 'oci.compute', 'oci.oke', 'oci.functions',
  'oci.load_balancer', 'oci.object_storage', 'oci.autonomous_database',
  'oci.dns', 'oci.vault',
  // Cloudflare
  'cloudflare.dns', 'cloudflare.waf', 'cloudflare.cdn',
  'cloudflare.tunnel', 'cloudflare.load_balancer', 'cloudflare.r2',
  'cloudflare.workers', 'cloudflare.access',
  // Solution
  'solution.user', 'solution.browser', 'solution.mobile_app',
  'solution.frontend', 'solution.backend_service', 'solution.microservice',
  'solution.worker', 'solution.function', 'solution.api_gateway',
  'solution.webhook', 'solution.event_bus', 'solution.queue',
  'solution.cache', 'solution.relational_database', 'solution.nosql_database',
  'solution.object_storage', 'solution.search', 'solution.analytics',
  'solution.observability', 'solution.vector_store', 'solution.ai_service',
  'solution.ci_cd', 'solution.registry',
  // Network
  'network.internet', 'network.wan', 'network.lan', 'network.vpn',
  'network.firewall', 'network.waf', 'network.ids_ips', 'network.nat',
  'network.subnet', 'network.private_subnet', 'network.public_subnet',
  'network.route_table', 'network.dns', 'network.cdn',
  'network.reverse_proxy', 'network.load_balancer_l4', 'network.load_balancer_l7',
  'network.bastion', 'network.transit_hub', 'network.peering',
  'network.zero_trust_proxy', 'network.office', 'network.datacenter',
  'network.branch_office', 'network.on_prem',
  'network.mpls', 'network.siem', 'network.file_server',
  'network.active_directory', 'network.proxy', 'network.switch',
  'network.router', 'network.endpoint', 'network.ztna',
  // Enterprise
  'enterprise.employee', 'enterprise.customer', 'enterprise.admin',
  'enterprise.partner_system', 'enterprise.saas_app', 'enterprise.erp',
  'enterprise.crm', 'enterprise.hris', 'enterprise.identity_provider',
  'enterprise.sso', 'enterprise.directory', 'enterprise.audit',
  'enterprise.compliance_boundary', 'enterprise.trust_boundary',
  'enterprise.business_unit', 'enterprise.integration_hub',
  'enterprise.data_warehouse', 'enterprise.bi_reporting', 'enterprise.etl',
  // Generic
  'generic.server', 'generic.database', 'generic.queue',
  'generic.cache', 'generic.load_balancer', 'generic.cdn',
  'generic.dns', 'generic.firewall', 'generic.storage',
  'generic.object_storage', 'generic.container', 'generic.api',
  'generic.user', 'generic.client', 'generic.monitoring',
  'generic.service', 'generic.vpn', 'generic.internet',
  'generic.on_prem', 'generic.webhook', 'generic.browser',
  'generic.mobile_app', 'generic.office',
]);

const VALID_EDGE_STYLES = new Set<EdgeStyle>(['solid', 'dashed', 'dotted']);

// --- Type helpers ---

export function parseNodeType(type: string): { prefix: string; service: string } {
  // Support both dot (new) and colon (legacy) separators
  const dotIdx = type.indexOf('.');
  if (dotIdx !== -1) {
    return { prefix: type.slice(0, dotIdx), service: type.slice(dotIdx + 1) };
  }
  const colonIdx = type.indexOf(':');
  if (colonIdx !== -1) {
    return { prefix: type.slice(0, colonIdx), service: type.slice(colonIdx + 1) };
  }
  return { prefix: 'generic', service: type };
}

export function isKnownType(type: string): boolean {
  return KNOWN_TYPES.has(type) || COMPONENT_IDS.has(type);
}

// --- Normalize edge style ---

function normalizeEdgeStyle(style: string | undefined): EdgeStyle {
  if (style && VALID_EDGE_STYLES.has(style as EdgeStyle)) {
    return style as EdgeStyle;
  }
  return 'solid';
}

// --- API Diagram → React Flow conversion ---

interface FlowData {
  nodes: Node[];
  edges: Edge[];
}

const GRID_COLS = 3;
const NODE_SPACING_X = 240;
const NODE_SPACING_Y = 180;
const GROUP_PADDING = 60;

export function diagramToFlow(diagram: Diagram): FlowData {
  const groupMap = new Map(diagram.groups.map((g) => [g.id, g]));

  // Build group membership: groupId → nodeIds
  const groupMembers = new Map<string, DiagramNode[]>();
  for (const node of diagram.nodes) {
    if (node.group && groupMap.has(node.group)) {
      const members = groupMembers.get(node.group) ?? [];
      members.push(node);
      groupMembers.set(node.group, members);
    }
  }

  const flowNodes: Node[] = [];
  const positioned = new Set<string>();

  // Lay out group nodes first
  let groupOffsetY = 0;
  for (const [groupId, members] of groupMembers) {
    const group = groupMap.get(groupId)!;
    const cols = Math.min(members.length, GRID_COLS);
    const rows = Math.ceil(members.length / GRID_COLS);

    flowNodes.push({
      id: groupId,
      type: 'group',
      position: { x: 0, y: groupOffsetY },
      data: { label: group.label, nodeType: group.type ?? 'generic.container' },
      style: {
        width: cols * NODE_SPACING_X + GROUP_PADDING,
        height: rows * NODE_SPACING_Y + GROUP_PADDING + 30,
      },
    });

    members.forEach((node, i) => {
      const col = i % GRID_COLS;
      const row = Math.floor(i / GRID_COLS);
      flowNodes.push({
        id: node.id,
        type: 'cloud',
        position: {
          x: GROUP_PADDING / 2 + col * NODE_SPACING_X,
          y: 30 + GROUP_PADDING / 2 + row * NODE_SPACING_Y,
        },
        parentId: groupId,
        data: {
          label: node.label,
          nodeType: node.type,
          metadata: node.metadata ?? {},
        },
      });
      positioned.add(node.id);
    });

    groupOffsetY += rows * NODE_SPACING_Y + GROUP_PADDING + 30 + 60;
  }

  // Lay out ungrouped nodes in a grid
  let ungroupedIdx = 0;
  const ungroupedOffsetX = groupMembers.size > 0 ? (GRID_COLS * NODE_SPACING_X + GROUP_PADDING + 80) : 0;

  for (const node of diagram.nodes) {
    if (positioned.has(node.id)) continue;

    const col = ungroupedIdx % GRID_COLS;
    const row = Math.floor(ungroupedIdx / GRID_COLS);

    flowNodes.push({
      id: node.id,
      type: 'cloud',
      position: {
        x: ungroupedOffsetX + col * NODE_SPACING_X,
        y: row * NODE_SPACING_Y,
      },
      data: {
        label: node.label,
        nodeType: node.type,
        metadata: node.metadata ?? {},
      },
    });
    ungroupedIdx++;
  }

  // Edges
  const nodeIds = new Set(diagram.nodes.map((n) => n.id));
  const flowEdges: Edge[] = diagram.edges
    .filter((e) => nodeIds.has(e.source) && nodeIds.has(e.target))
    .map((e) => toFlowEdge(e));

  return { nodes: flowNodes, edges: flowEdges };
}

function toFlowEdge(edge: DiagramEdge): Edge {
  const style = normalizeEdgeStyle(edge.style);
  return {
    id: edge.id,
    source: edge.source,
    target: edge.target,
    label: edge.label,
    type: 'smoothstep',
    style: {
      stroke: '#4f46e5',
      strokeWidth: 1.5,
      opacity: style === 'dotted' ? 0.35 : 0.55,
      strokeDasharray: style === 'dashed' ? '8 4' : style === 'dotted' ? '3 3' : undefined,
    },
    animated: false,
  };
}

// --- React Flow → API Diagram conversion ---

export function flowToDiagram(flowNodes: Node[], flowEdges: Edge[]): Diagram {
  const groups: DiagramGroup[] = [];
  const nodes: DiagramNode[] = [];

  for (const n of flowNodes) {
    if (n.type === 'group') {
      groups.push({
        id: n.id,
        label: (n.data as Record<string, string>).label ?? n.id,
        type: (n.data as Record<string, string>).nodeType,
      });
    } else {
      const data = n.data as Record<string, unknown>;
      nodes.push({
        id: n.id,
        label: (data.label as string) ?? n.id,
        type: (data.nodeType as string) ?? 'generic.server',
        group: n.parentId,
        metadata: (data.metadata as Record<string, string>) ?? undefined,
      });
    }
  }

  const edges: DiagramEdge[] = flowEdges.map((e) => ({
    id: e.id,
    source: e.source,
    target: e.target,
    label: typeof e.label === 'string' ? e.label : undefined,
    style: edgeStyleFromFlow(e),
  }));

  return { nodes, edges, groups };
}

function edgeStyleFromFlow(edge: Edge): EdgeStyle {
  const dash = edge.style?.strokeDasharray;
  if (typeof dash === 'string') {
    if (dash.startsWith('3') || dash.startsWith('2')) return 'dotted';
    if (dash.startsWith('8') || dash.startsWith('6')) return 'dashed';
  }
  if (edge.animated) return 'dashed';
  return 'solid';
}
