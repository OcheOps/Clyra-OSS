import { describe, it, expect } from 'vitest';
import { validateDiagram, parseDiagramJSON } from './validate';
import type { Diagram } from '@/types/diagram';

describe('validateDiagram', () => {
  it('passes for a valid diagram with new type format', () => {
    const diagram: Diagram = {
      nodes: [
        { id: 'n1', label: 'Cluster', type: 'k8s.cluster' },
        { id: 'n2', label: 'Deploy', type: 'k8s.deployment', group: 'g1' },
        { id: 'n3', label: 'FW', type: 'network.firewall' },
        { id: 'n4', label: 'ERP', type: 'enterprise.erp' },
      ],
      edges: [
        { id: 'e1', source: 'n1', target: 'n2' },
        { id: 'e2', source: 'n3', target: 'n4', style: 'dashed' },
      ],
      groups: [{ id: 'g1', label: 'Namespace' }],
    };
    const result = validateDiagram(diagram);
    expect(result.valid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it('rejects empty nodes', () => {
    const result = validateDiagram({ nodes: [], edges: [], groups: [] });
    expect(result.valid).toBe(false);
  });

  it('detects duplicate node ids', () => {
    const diagram: Diagram = {
      nodes: [
        { id: 'n1', label: 'A', type: 'solution.microservice' },
        { id: 'n1', label: 'B', type: 'solution.worker' },
      ],
      edges: [], groups: [],
    };
    expect(validateDiagram(diagram).valid).toBe(false);
  });

  it('detects invalid edge references', () => {
    const diagram: Diagram = {
      nodes: [{ id: 'n1', label: 'A', type: 'network.bastion' }],
      edges: [{ id: 'e1', source: 'n1', target: 'missing' }],
      groups: [],
    };
    expect(validateDiagram(diagram).errors.some((e) => e.includes('invalid target'))).toBe(true);
  });
});

describe('parseDiagramJSON', () => {
  it('parses valid JSON with new types', () => {
    const json = JSON.stringify({
      nodes: [
        { id: 'n1', label: 'Tunnel', type: 'cloudflare.tunnel' },
        { id: 'n2', label: 'ERP', type: 'enterprise.erp' },
      ],
      edges: [{ id: 'e1', source: 'n1', target: 'n2' }],
      groups: [],
    });
    const result = parseDiagramJSON(json);
    expect(result.diagram).not.toBeNull();
    expect(result.errors).toHaveLength(0);
  });

  it('rejects invalid JSON', () => {
    expect(parseDiagramJSON('not json{{{').diagram).toBeNull();
  });

  it('handles missing arrays gracefully', () => {
    const result = parseDiagramJSON('{}');
    expect(result.errors.some((e) => e.includes('at least one node'))).toBe(true);
  });
});
