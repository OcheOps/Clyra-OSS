export const EDGE_STYLES = ['solid', 'dashed', 'dotted'] as const;

export const PROVIDERS = ['managed_bedrock', 'openai', 'anthropic', 'gemini'] as const;

export const ARCHITECTURE_MODES = ['solution_architecture', 'network_architecture', 'enterprise_architecture'] as const;

export const TYPE_FAMILIES = [
  'aws', 'gcp', 'azure', 'k8s', 'oci', 'cloudflare',
  'solution', 'network', 'enterprise', 'generic',
] as const;

export type TypeFamily = (typeof TYPE_FAMILIES)[number];

// Legacy export for backward compat
export const CLOUD_PREFIXES = TYPE_FAMILIES;
export type CloudPrefix = TypeFamily;
