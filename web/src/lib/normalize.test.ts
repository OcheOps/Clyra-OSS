import { describe, it, expect } from 'vitest';
import { parseNodeType, isKnownType, diagramToFlow, flowToDiagram } from './normalize';
import type { Diagram } from '@/types/diagram';

describe('parseNodeType', () => {
  it('splits dot-separated types correctly', () => {
    expect(parseNodeType('aws.ec2')).toEqual({ prefix: 'aws', service: 'ec2' });
    expect(parseNodeType('k8s.deployment')).toEqual({ prefix: 'k8s', service: 'deployment' });
    expect(parseNodeType('solution.microservice')).toEqual({ prefix: 'solution', service: 'microservice' });
    expect(parseNodeType('network.firewall')).toEqual({ prefix: 'network', service: 'firewall' });
    expect(parseNodeType('enterprise.erp')).toEqual({ prefix: 'enterprise', service: 'erp' });
  });

  it('handles legacy colon-separated types', () => {
    expect(parseNodeType('aws:ec2')).toEqual({ prefix: 'aws', service: 'ec2' });
    expect(parseNodeType('k8s:pod')).toEqual({ prefix: 'k8s', service: 'pod' });
  });

  it('defaults to generic prefix when no separator', () => {
    expect(parseNodeType('server')).toEqual({ prefix: 'generic', service: 'server' });
  });
});

describe('isKnownType', () => {
  it('returns true for new dot-format types', () => {
    expect(isKnownType('aws.ec2')).toBe(true);
    expect(isKnownType('gcp.cloud_run')).toBe(true);
    expect(isKnownType('azure.aks')).toBe(true);
    expect(isKnownType('generic.server')).toBe(true);
    expect(isKnownType('aws.bedrock_agents')).toBe(true);
    expect(isKnownType('gcp.cloud_deploy')).toBe(true);
  });

  it('returns true for new architecture types', () => {
    expect(isKnownType('solution.microservice')).toBe(true);
    expect(isKnownType('solution.vector_store')).toBe(true);
    expect(isKnownType('network.firewall')).toBe(true);
    expect(isKnownType('network.bastion')).toBe(true);
    expect(isKnownType('enterprise.erp')).toBe(true);
    expect(isKnownType('enterprise.sso')).toBe(true);
  });

  it('returns true for K8s, OCI, Cloudflare', () => {
    expect(isKnownType('k8s.deployment')).toBe(true);
    expect(isKnownType('oci.compute')).toBe(true);
    expect(isKnownType('cloudflare.workers')).toBe(true);
  });

  it('returns false for unknown types', () => {
    expect(isKnownType('aws.made_up')).toBe(false);
    expect(isKnownType('foo.bar')).toBe(false);
    expect(isKnownType('')).toBe(false);
  });

  it('returns false for legacy colon format', () => {
    // Frontend KNOWN_TYPES only has dot format now
    expect(isKnownType('aws:ec2')).toBe(false);
  });
});

describe('diagramToFlow', () => {
  it('converts nodes from all families', () => {
    const diagram: Diagram = {
      nodes: [
        { id: 'n1', label: 'Web', type: 'aws.ec2' },
        { id: 'n2', label: 'Cache', type: 'cloudflare.cdn' },
        { id: 'n3', label: 'App', type: 'k8s.deployment' },
        { id: 'n4', label: 'FW', type: 'network.firewall' },
        { id: 'n5', label: 'ERP', type: 'enterprise.erp' },
      ],
      edges: [{ id: 'e1', source: 'n1', target: 'n2' }],
      groups: [],
    };

    const { nodes, edges } = diagramToFlow(diagram);
    expect(nodes).toHaveLength(5);
    expect(edges).toHaveLength(1);

    for (const node of nodes) {
      expect(node.type).toBe('cloud');
    }

    expect((nodes[0]!.data as Record<string, unknown>).nodeType).toBe('aws.ec2');
    expect((nodes[3]!.data as Record<string, unknown>).nodeType).toBe('network.firewall');
    expect((nodes[4]!.data as Record<string, unknown>).nodeType).toBe('enterprise.erp');
  });
});

describe('flowToDiagram', () => {
  it('round-trips through diagramToFlow and back', () => {
    const original: Diagram = {
      nodes: [
        { id: 'n1', label: 'Worker', type: 'cloudflare.workers' },
        { id: 'n2', label: 'Service', type: 'solution.microservice' },
      ],
      edges: [{ id: 'e1', source: 'n1', target: 'n2', label: 'calls', style: 'dashed' }],
      groups: [],
    };

    const flow = diagramToFlow(original);
    const result = flowToDiagram(flow.nodes, flow.edges);

    expect(result.nodes).toHaveLength(2);
    expect(result.nodes[0]!.type).toBe('cloudflare.workers');
    expect(result.nodes[1]!.type).toBe('solution.microservice');
    expect(result.edges[0]!.style).toBe('dashed');
  });
});
