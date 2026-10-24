import { describe, it, expect } from 'vitest';
import { getIconDef, getFamilyColor, getServicesByPrefix } from './icons';

describe('getIconDef', () => {
  it('returns icon for known AWS type', () => {
    const icon = getIconDef('aws.ec2');
    expect(icon.label).toBe('EC2');
    expect(icon.color).toBe('#FF9900');
    expect(icon.icon).toBe('cdn:AWS/svg/Compute/EC2.svg');
  });

  it('returns icon for known GCP type', () => {
    expect(getIconDef('gcp.cloud_run').label).toBe('Cloud Run');
  });

  it('returns icon for known Azure type', () => {
    expect(getIconDef('azure.cosmos_db').label).toBe('Cosmos DB');
  });

  it('returns icon for known K8s type', () => {
    const icon = getIconDef('k8s.deployment');
    expect(icon.label).toBe('Deployment');
    expect(icon.color).toBe('#326CE5');
  });

  it('returns icon for known OCI type', () => {
    const icon = getIconDef('oci.compute');
    expect(icon.label).toBe('Compute');
    expect(icon.color).toBe('#F80000');
  });

  it('returns icon for known Cloudflare type', () => {
    const icon = getIconDef('cloudflare.workers');
    expect(icon.label).toBe('Workers');
    expect(icon.color).toBe('#F6821F');
  });

  it('returns icon for known Solution type', () => {
    const icon = getIconDef('solution.microservice');
    expect(icon.label).toBe('Microservice');
    expect(icon.color).toBe('#10B981');
  });

  it('returns icon for known Network type', () => {
    const icon = getIconDef('network.firewall');
    expect(icon.label).toBe('Firewall');
    expect(icon.color).toBe('#6366F1');
  });

  it('returns icon for known Enterprise type', () => {
    const icon = getIconDef('enterprise.erp');
    expect(icon.label).toBe('ERP');
    expect(icon.color).toBe('#8B5CF6');
  });

  it('returns smart fallback for unknown type', () => {
    const icon = getIconDef('foo.bar_baz');
    expect(icon.label).toBe('Bar Baz');
    expect(icon.icon).toBe('circle');
  });

  it('returns smart fallback with family color', () => {
    const icon = getIconDef('aws.some_new_service');
    expect(icon.label).toBe('Some New Service');
    expect(icon.color).toBe('#FF9900');
    expect(icon.icon).toBe('cloud');
  });

  // Spot-check new services
  const newServices = [
    'aws.subnet', 'aws.nlb', 'aws.nat_gateway', 'aws.ecr',
    'gcp.app_engine', 'gcp.memorystore', 'gcp.secret_manager',
    'azure.vnet', 'azure.application_gateway', 'azure.key_vault',
    'k8s.cluster', 'k8s.pod', 'k8s.ingress', 'k8s.statefulset',
    'oci.vcn', 'oci.oke', 'oci.object_storage',
    'cloudflare.dns', 'cloudflare.tunnel', 'cloudflare.r2',
    'solution.backend_service', 'solution.vector_store', 'solution.ai_service',
    'network.bastion', 'network.zero_trust_proxy', 'network.load_balancer_l7',
    'enterprise.sso', 'enterprise.identity_provider', 'enterprise.etl',
    'generic.browser', 'generic.mobile_app', 'generic.webhook',
  ];

  it.each(newServices)('has icon for %s', (type) => {
    expect(getIconDef(type).label).not.toBe('Unknown');
  });
});

describe('getFamilyColor', () => {
  it('returns correct color for each family', () => {
    expect(getFamilyColor('aws.ec2')).toBe('#FF9900');
    expect(getFamilyColor('gcp.gke')).toBe('#4285F4');
    expect(getFamilyColor('azure.vm')).toBe('#0078D4');
    expect(getFamilyColor('k8s.pod')).toBe('#326CE5');
    expect(getFamilyColor('oci.compute')).toBe('#F80000');
    expect(getFamilyColor('cloudflare.cdn')).toBe('#F6821F');
    expect(getFamilyColor('solution.microservice')).toBe('#10B981');
    expect(getFamilyColor('network.firewall')).toBe('#6366F1');
    expect(getFamilyColor('enterprise.erp')).toBe('#8B5CF6');
    expect(getFamilyColor('generic.server')).toBe('#8892a8');
  });

  it('returns generic color for unknown family', () => {
    expect(getFamilyColor('unknown.thing')).toBe('#8892a8');
  });
});

describe('getServicesByPrefix', () => {
  const grouped = getServicesByPrefix();

  it('has all 10 family groups', () => {
    expect(Object.keys(grouped)).toEqual(
      expect.arrayContaining([
        'aws', 'gcp', 'azure', 'k8s', 'oci', 'cloudflare',
        'solution', 'network', 'enterprise', 'generic',
      ]),
    );
  });

  it('keeps representative catalog entries in every core family', () => {
    expect(grouped.aws?.some((item) => item.type === 'aws.ec2')).toBe(true);
    expect(grouped.gcp?.some((item) => item.type === 'gcp.pubsub')).toBe(true);
    expect(grouped.azure?.some((item) => item.type === 'azure.redis')).toBe(true);
    expect(grouped.k8s?.some((item) => item.type === 'k8s.deployment')).toBe(true);
    expect(grouped.solution?.some((item) => item.type === 'solution.microservice')).toBe(true);
    expect(grouped.network?.some((item) => item.type === 'net.dns_authoritative')).toBe(true);
    expect(grouped.enterprise?.some((item) => item.type === 'ent.active_directory')).toBe(true);
  });

  it('contains no duplicate canonical types', () => {
    const types = Object.values(grouped).flat().map((item) => item.type);
    expect(new Set(types).size).toBe(types.length);
  });

  it('each entry has type and icon', () => {
    for (const items of Object.values(grouped)) {
      for (const item of items!) {
        expect(item.type).toBeTruthy();
        expect(item.icon.label).toBeTruthy();
        expect(item.icon.color).toBeTruthy();
        expect(item.icon.icon).toBeTruthy();
      }
    }
  });
});
