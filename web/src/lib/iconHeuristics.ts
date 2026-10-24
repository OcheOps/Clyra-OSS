// Lucide icon inference for canonical components that don't have a
// hand-curated ICON_REGISTRY entry. Two-pass:
//   1) subcategory → icon lookup (authoritative when we have one)
//   2) keyword regex over canonical name as fallback
// Keep this list tight — the hand-curated ICON_REGISTRY should still own
// anything visually distinctive; this is the long tail.

// Icon names must exist in lucide-react OR in web/src/lib/customIcons.
const SUB_ICON: Record<string, string> = {
  // compute
  vm: 'server',
  container: 'container',
  serverless: 'zap',
  paas: 'sprout',
  hpc: 'cpu',
  registry: 'package',
  games: 'gamepad-2',
  // network / edge
  network: 'network',
  lb: 'split',
  dns: 'signpost',
  cdn: 'globe',
  api: 'door-open',
  service_discovery: 'radar',
  nat: 'arrow-right-left',
  routing: 'route',
  switching: 'network',
  transport: 'truck',
  overlay: 'layers',
  fabric: 'grid-2x2',
  addressing: 'hash',
  qos: 'gauge',
  kernel: 'terminal',
  proxy: 'split',
  backbone: 'cable',
  peering: 'users',
  // storage
  object: 'bucket',
  block: 'hard-drive',
  file: 'files',
  backup: 'archive',
  storage: 'bucket',
  // database
  relational: 'database',
  nosql: 'table',
  cache: 'memory-stick',
  graph: 'share-2',
  warehouse: 'warehouse',
  timeseries: 'bar-chart-3',
  ledger: 'scroll',
  db: 'database',
  // messaging / streaming / events
  queue: 'mail',
  pubsub: 'megaphone',
  broker: 'radio-tower',
  events: 'bus',
  streaming: 'waves',
  messaging: 'megaphone',
  cdc: 'refresh-cw',
  orchestration: 'workflow',
  integration: 'git-merge',
  // observability / ops
  observability: 'activity',
  tracing: 'git-branch',
  metrics: 'activity',
  logs: 'file-text',
  audit: 'footprints',
  ops: 'terminal',
  uptime: 'heart-pulse',
  experiments: 'flask-conical',
  resilience: 'shield-check',
  dr: 'life-buoy',
  // security
  identity: 'key-round',
  security: 'shield',
  crypto: 'lock-keyhole',
  compliance: 'gavel',
  governance: 'gavel',
  sdwan: 'network',
  vpn: 'lock',
  // ai
  ai: 'bot',
  search: 'search',
  quantum: 'atom',
  robotics: 'bot',
  ml: 'brain',
  // iot / edge / space
  iot: 'cpu',
  space: 'satellite-dish',
  hybrid: 'building-2',
  mec: 'antenna',
  industrial: 'factory',
  protocol: 'radio',
  wireless: 'wifi',
  twin: 'box',
  edge_runtime: 'zap',
  edge_hw: 'cpu',
  telco: 'antenna',
  // k8s
  control: 'cog',
  node: 'hexagon',
  workload: 'boxes',
  config: 'settings',
  autoscale: 'scaling',
  mesh: 'network',
  packaging: 'package',
  platform: 'layers',
  infra: 'server',
  tooling: 'wrench',
  // data / analytics
  analytics: 'bar-chart-3',
  etl: 'git-merge',
  catalog: 'book-open',
  quality: 'check-circle',
  table_format: 'table',
  format: 'file',
  engine: 'cog',
  lakehouse: 'database',
  // enterprise
  erp: 'building-2',
  crm: 'users',
  hcm: 'user-check',
  itsm: 'wrench',
  incident: 'alert-triangle',
  siem: 'eye',
  endpoint: 'shield',
  pam: 'key',
  secrets: 'lock',
  scm: 'git-branch',
  ci: 'hammer',
  cd: 'rocket',
  collab: 'message-circle',
  productivity: 'layout-grid',
  payments: 'credit-card',
  communications: 'phone',
  directory: 'folder-tree',
  iac: 'file-code-2',
  bi: 'bar-chart-3',
  batch: 'list-checks',
  xr: 'cuboid',
  clipboard: 'clipboard-list',
  procurement: 'shopping-cart',
  cost: 'wallet',
  billing: 'credit-card',
  geo: 'map',
  vdi: 'monitor',
  frontend: 'layout-panel-top',
  contact_center: 'headphones',
  blockchain: 'link',
  health: 'heart-pulse',
  migration: 'move-right',
  entity: 'user',
  auth: 'shield',
};

const KEYWORD_ICON: Array<[RegExp, string]> = [
  [/firewall|waf|shield/, 'shield'],
  [/vpn\b|wireguard|ipsec/, 'lock'],
  [/cache|redis|memcached|valkey|memorystore/, 'memory-stick'],
  [/bucket|blob|\bs3\b/, 'bucket'],
  [/database|_db\b|_sql\b|postgres|mysql|oracle|mariadb/, 'database'],
  [/log/, 'file-text'],
  [/metric|monitor/, 'activity'],
  [/trace|tracing|jaeger|zipkin/, 'git-branch'],
  [/queue|\bsqs\b/, 'mail'],
  [/pubsub|topic|event_grid|event_hub|event_bridge/, 'megaphone'],
  [/kafka|stream|kinesis|pulsar/, 'waves'],
  [/lambda|function(s)?\b/, 'zap'],
  [/container|docker|fargate/, 'container'],
  [/kubernetes|\bgke\b|\beks\b|\baks\b/, 'hexagon'],
  [/load_balancer|\blb\b|alb|nlb|elb|gwlb/, 'split'],
  [/\bdns\b|route_53|cloud_dns/, 'signpost'],
  [/\bcdn\b|cloudfront/, 'globe'],
  [/iam|identity|auth/, 'key-round'],
  [/secret/, 'lock-keyhole'],
  [/\bkey\b|crypto|\bhsm\b|\bkms\b/, 'key'],
  [/route|routing/, 'route'],
  [/subnet|vpc|vnet/, 'square-dashed'],
  [/\bnat\b/, 'arrow-right-left'],
  [/backup|snapshot|archive/, 'archive'],
  [/disk|volume|\bssd\b|ebs|\bpd\b/, 'hard-drive'],
  [/quantum|braket/, 'atom'],
  [/satellite|ground_station|orbital|kuiper|starlink|iridium|oneweb|globalstar|leo|meo|geo/, 'satellite-dish'],
  [/iot|device|sensor|gateway/, 'cpu'],
  [/openai|bedrock|gemini|vertex|sagemaker|claude|anthropic|llm|ai_/, 'bot'],
  [/robot|robomaker/, 'bot'],
  [/mail|smtp|\bses\b|email/, 'mail'],
  [/chat|slack|teams|discord|zoom/, 'message-circle'],
  [/phone|voice|call|ivr/, 'phone'],
  [/video|media|livestream/, 'video'],
  [/search|elastic|opensearch|kendra/, 'search'],
  [/workflow|step_functions|orchestrat/, 'workflow'],
  [/\bgit\b|repo|scm/, 'git-branch'],
  [/\bbuild\b|\bci\b/, 'hammer'],
  [/deploy|release/, 'rocket'],
  [/registry|artifact|ecr|acr|harbor/, 'package'],
  [/api|apigee|apim|gateway/, 'door-open'],
  [/proxy/, 'split'],
  [/payment|billing|stripe|paystack|paypal|adyen/, 'credit-card'],
  [/calendar|schedule|cron/, 'calendar'],
  [/factory|plant|industrial|opc|modbus|profinet|ethercat|canbus|profibus/, 'factory'],
  [/wifi|lorawan|zigbee|zwave|thread|matter|bluetooth|ble/, 'wifi'],
  [/submarine|fiber|ixp/, 'cable'],
  [/twin/, 'box'],
  [/chart|analytics|dashboard|quicksight|power_bi|tableau/, 'bar-chart-3'],
  [/bastion|jump/, 'key'],
  [/sso|saml|oauth|oidc|okta|auth0|entra|keycloak|cognito/, 'key-round'],
  [/ticket|itsm|servicenow|jira|zendesk/, 'wrench'],
  [/crm|salesforce|hubspot/, 'users'],
  [/hcm|workday/, 'user-check'],
  [/erp|sap|oracle_fusion|netsuite|dynamics|infor/, 'building-2'],
  [/backup|veeam|commvault|rubrik|cohesity|druva/, 'archive'],
  [/siem|splunk|sentinel|chronicle|datadog|new_relic|dynatrace/, 'eye'],
  [/edr|crowdstrike|sentinelone|carbon_black|mde/, 'shield'],
  [/pam|cyberark|vault|boundary|beyondtrust/, 'key'],
  [/spark|flink|hadoop|hive|trino|presto|clickhouse|druid|pinot|dbt|airflow|prefect|dagster|temporal/, 'git-merge'],
  [/iceberg|delta|hudi|parquet|orc|avro|arrow/, 'table'],
];

export function inferLucideIcon(canonical: string, subcategory: string): string {
  const sub = subcategory.toLowerCase().trim();
  if (sub && SUB_ICON[sub]) return SUB_ICON[sub];

  const probe = `${canonical.toLowerCase()} ${sub}`;
  for (const [re, icon] of KEYWORD_ICON) {
    if (re.test(probe)) return icon;
  }
  return 'box';
}

// Additional colors for the new provider codes in the generated catalog.
// The existing FAMILY_COLORS (icons.ts) already handles aws/azure/gcp/k8s
// and maps `net`→network + `ent`→enterprise. These extend the palette so
// data/edge/iot get distinct looks.
export const EXTENDED_FAMILY_COLORS: Record<string, string> = {
  net: '#6366F1',       // indigo (matches network)
  ent: '#8B5CF6',       // purple (matches enterprise)
  data: '#14B8A6',      // teal
  edge: '#F43F5E',      // rose
  iot: '#F97316',       // orange
};
