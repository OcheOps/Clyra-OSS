import type { Diagram } from '@/types/diagram';

export interface ValidationResult {
  valid: boolean;
  errors: string[];
}

/**
 * Validates a diagram structure. Mirrors backend diagram/validator.go logic.
 */
export function validateDiagram(diagram: Diagram): ValidationResult {
  const errors: string[] = [];

  if (!diagram.nodes || diagram.nodes.length === 0) {
    errors.push('diagram must have at least one node');
    return { valid: false, errors };
  }

  // Unique node IDs
  const nodeIds = new Set<string>();
  for (const node of diagram.nodes) {
    if (!node.id) {
      errors.push('node is missing an id');
      continue;
    }
    if (nodeIds.has(node.id)) {
      errors.push(`duplicate node id: ${node.id}`);
    }
    nodeIds.add(node.id);

    if (!node.label) {
      errors.push(`node ${node.id} is missing a label`);
    }
    if (!node.type) {
      errors.push(`node ${node.id} is missing a type`);
    }
  }

  // Unique group IDs
  const groupIds = new Set<string>();
  for (const group of diagram.groups ?? []) {
    if (!group.id) {
      errors.push('group is missing an id');
      continue;
    }
    if (groupIds.has(group.id)) {
      errors.push(`duplicate group id: ${group.id}`);
    }
    groupIds.add(group.id);
  }

  // Validate group references on nodes
  for (const node of diagram.nodes) {
    if (node.group && !groupIds.has(node.group)) {
      errors.push(`node ${node.id} references unknown group: ${node.group}`);
    }
  }

  // Unique edge IDs + valid source/target
  const edgeIds = new Set<string>();
  for (const edge of diagram.edges ?? []) {
    if (!edge.id) {
      errors.push('edge is missing an id');
      continue;
    }
    if (edgeIds.has(edge.id)) {
      errors.push(`duplicate edge id: ${edge.id}`);
    }
    edgeIds.add(edge.id);

    if (!nodeIds.has(edge.source)) {
      errors.push(`edge ${edge.id} has invalid source: ${edge.source}`);
    }
    if (!nodeIds.has(edge.target)) {
      errors.push(`edge ${edge.id} has invalid target: ${edge.target}`);
    }
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Checks if raw JSON can be parsed as a Diagram.
 * Returns the parsed diagram or null with errors.
 */
export function parseDiagramJSON(json: string): { diagram: Diagram | null; errors: string[] } {
  let parsed: unknown;
  try {
    parsed = JSON.parse(json);
  } catch {
    return { diagram: null, errors: ['invalid JSON'] };
  }

  if (typeof parsed !== 'object' || parsed === null) {
    return { diagram: null, errors: ['expected a JSON object'] };
  }

  const obj = parsed as Record<string, unknown>;

  const diagram: Diagram = {
    nodes: Array.isArray(obj.nodes) ? obj.nodes : [],
    edges: Array.isArray(obj.edges) ? obj.edges : [],
    groups: Array.isArray(obj.groups) ? obj.groups : [],
  };

  const result = validateDiagram(diagram);
  return { diagram: result.valid ? diagram : null, errors: result.errors };
}
