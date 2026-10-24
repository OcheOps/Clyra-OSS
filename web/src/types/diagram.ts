// --- API schema (matches backend provider.* structs) ---

export interface DiagramNode {
  id: string;
  label: string;
  type: string;
  group?: string;
  metadata?: Record<string, string>;
}

export interface DiagramEdge {
  id: string;
  source: string;
  target: string;
  label?: string;
  style?: EdgeStyle;
}

export interface DiagramGroup {
  id: string;
  label: string;
  type?: string;
}

export type EdgeStyle = 'solid' | 'dashed' | 'dotted';

export interface Diagram {
  nodes: DiagramNode[];
  edges: DiagramEdge[];
  groups: DiagramGroup[];
}

// --- API request/response ---

export type ProviderName = 'managed_bedrock' | 'openai' | 'anthropic' | 'gemini';

export type ArchitectureMode = 'solution_architecture' | 'network_architecture' | 'enterprise_architecture';

export interface GenerateRequest {
  prompt: string;
  provider: ProviderName;
  api_key?: string;
  mode?: ArchitectureMode;
  config?: ProviderConfig;
}

export interface ProviderConfig {
  model?: string;
  max_tokens?: number;
  temperature?: number;
}

export interface StreamChunk {
  delta: string;
  done: boolean;
  error?: string;
}

// --- Local project ---

export interface Project {
  id: string;
  name: string;
  diagram: Diagram;
  createdAt: number;
  updatedAt: number;
}
