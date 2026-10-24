import type { Diagram, ArchitectureMode } from '@/types/diagram';

export interface StarterTemplate {
  id: string;
  title: string;
  summary: string;
  mode: ArchitectureMode;
  diagram: Diagram;
}

export const STARTER_TEMPLATES: StarterTemplate[] = [
  {
    id: 'aws-web-app',
    title: 'AWS web app',
    summary: 'ALB → EC2 → RDS with a managed cache.',
    mode: 'solution_architecture',
    diagram: {
      nodes: [
        { id: 'users', label: 'Users', type: 'generic.users' },
        { id: 'alb', label: 'ALB', type: 'aws.elb' },
        { id: 'web', label: 'Web tier', type: 'aws.ec2' },
        { id: 'cache', label: 'ElastiCache', type: 'aws.elasticache' },
        { id: 'db', label: 'Postgres', type: 'aws.rds' },
      ],
      edges: [
        { id: 'e1', source: 'users', target: 'alb', label: 'HTTPS' },
        { id: 'e2', source: 'alb', target: 'web', label: 'HTTP' },
        { id: 'e3', source: 'web', target: 'cache', label: 'Redis' },
        { id: 'e4', source: 'web', target: 'db', label: 'SQL' },
      ],
      groups: [],
    },
  },
  {
    id: 'k8s-system',
    title: 'Kubernetes system',
    summary: 'Ingress in front of a service mesh with a managed database.',
    mode: 'solution_architecture',
    diagram: {
      nodes: [
        { id: 'users', label: 'Users', type: 'generic.users' },
        { id: 'ingress', label: 'Ingress', type: 'k8s.ingress' },
        { id: 'svc-a', label: 'Service A', type: 'k8s.service' },
        { id: 'svc-b', label: 'Service B', type: 'k8s.service' },
        { id: 'queue', label: 'Queue', type: 'generic.queue' },
        { id: 'db', label: 'Database', type: 'generic.database' },
      ],
      edges: [
        { id: 'e1', source: 'users', target: 'ingress', label: 'HTTPS' },
        { id: 'e2', source: 'ingress', target: 'svc-a' },
        { id: 'e3', source: 'svc-a', target: 'svc-b', label: 'gRPC' },
        { id: 'e4', source: 'svc-a', target: 'queue', label: 'publish' },
        { id: 'e5', source: 'svc-b', target: 'db' },
      ],
      groups: [],
    },
  },
  {
    id: 'microservices',
    title: 'Microservices',
    summary: 'API gateway fanning out to independent services with an event bus.',
    mode: 'solution_architecture',
    diagram: {
      nodes: [
        { id: 'users', label: 'Users', type: 'generic.users' },
        { id: 'gw', label: 'API Gateway', type: 'aws.api_gateway' },
        { id: 'auth', label: 'Auth service', type: 'generic.service' },
        { id: 'orders', label: 'Orders service', type: 'generic.service' },
        { id: 'billing', label: 'Billing service', type: 'generic.service' },
        { id: 'bus', label: 'Event bus', type: 'generic.queue' },
        { id: 'db-orders', label: 'Orders DB', type: 'generic.database' },
        { id: 'db-billing', label: 'Billing DB', type: 'generic.database' },
      ],
      edges: [
        { id: 'e1', source: 'users', target: 'gw' },
        { id: 'e2', source: 'gw', target: 'auth' },
        { id: 'e3', source: 'gw', target: 'orders' },
        { id: 'e4', source: 'orders', target: 'db-orders' },
        { id: 'e5', source: 'orders', target: 'bus', label: 'order.created' },
        { id: 'e6', source: 'bus', target: 'billing', label: 'order.created' },
        { id: 'e7', source: 'billing', target: 'db-billing' },
      ],
      groups: [],
    },
  },
  {
    id: 'enterprise',
    title: 'Enterprise system',
    summary: 'CRM, ERP, data warehouse and analytics — business-level view.',
    mode: 'enterprise_architecture',
    diagram: {
      nodes: [
        { id: 'crm', label: 'CRM', type: 'enterprise.app' },
        { id: 'erp', label: 'ERP', type: 'enterprise.app' },
        { id: 'esb', label: 'Integration bus', type: 'enterprise.integration' },
        { id: 'dwh', label: 'Data warehouse', type: 'enterprise.data' },
        { id: 'bi', label: 'BI / Analytics', type: 'enterprise.app' },
        { id: 'iam', label: 'Identity', type: 'enterprise.security' },
      ],
      edges: [
        { id: 'e1', source: 'crm', target: 'esb' },
        { id: 'e2', source: 'erp', target: 'esb' },
        { id: 'e3', source: 'esb', target: 'dwh', label: 'ETL' },
        { id: 'e4', source: 'dwh', target: 'bi' },
        { id: 'e5', source: 'iam', target: 'crm', style: 'dashed' },
        { id: 'e6', source: 'iam', target: 'erp', style: 'dashed' },
      ],
      groups: [],
    },
  },
];
