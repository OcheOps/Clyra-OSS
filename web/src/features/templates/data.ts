import type { Diagram } from '@/types/diagram';

// Template catalog — five reference architectures a visitor would expect
// to see in an AWS-focused tool. Keeping them deliberately small (5-8
// nodes each) so auto-layout produces a readable canvas on first open.
// The `type` field matches node-type strings produced by the generator
// (normalize.ts looks them up in the component catalog to pick an icon).

export interface Template {
  id: string;
  name: string;
  summary: string;
  prompt: string;        // Pre-filled into the editor's prompt drawer.
  tag: 'starter' | 'data' | 'app' | 'ai';
  diagram: Diagram;
}

export const TEMPLATES: Template[] = [
  {
    id: 'three-tier-aws',
    name: 'Three-tier web app',
    summary: 'ALB → ECS Fargate → RDS Postgres with an S3 asset bucket.',
    tag: 'starter',
    prompt: 'A three-tier web app on AWS — ALB, ECS Fargate, RDS Postgres, Redis cache, and S3 for assets.',
    diagram: {
      nodes: [
        { id: 'route53', label: 'Route 53', type: 'aws:route53' },
        { id: 'cloudfront', label: 'CloudFront', type: 'aws:cloudfront' },
        { id: 'alb', label: 'ALB', type: 'aws:alb' },
        { id: 'ecs', label: 'ECS Fargate', type: 'aws:ecs' },
        { id: 'rds', label: 'RDS Postgres', type: 'aws:rds' },
        { id: 'redis', label: 'ElastiCache', type: 'aws:elasticache' },
        { id: 's3', label: 'S3 assets', type: 'aws:s3' },
      ],
      edges: [
        { id: 'e1', source: 'route53', target: 'cloudfront', label: 'DNS' },
        { id: 'e2', source: 'cloudfront', target: 'alb', label: 'HTTPS' },
        { id: 'e3', source: 'cloudfront', target: 's3', label: 'static' },
        { id: 'e4', source: 'alb', target: 'ecs', label: '443' },
        { id: 'e5', source: 'ecs', target: 'rds', label: '5432' },
        { id: 'e6', source: 'ecs', target: 'redis', label: '6379' },
      ],
      groups: [],
    },
  },
  {
    id: 'nextjs-postgres',
    name: 'Next.js + Postgres',
    summary: 'SSR frontend, managed Postgres, background jobs.',
    tag: 'app',
    prompt: 'A Next.js app deployed on AWS with RDS Postgres, Redis for cache, and a background worker pulling from SQS.',
    diagram: {
      nodes: [
        { id: 'cf', label: 'CloudFront', type: 'aws:cloudfront' },
        { id: 'alb', label: 'ALB', type: 'aws:alb' },
        { id: 'nextjs', label: 'Next.js (ECS)', type: 'aws:ecs' },
        { id: 'worker', label: 'Worker (ECS)', type: 'aws:ecs' },
        { id: 'sqs', label: 'SQS queue', type: 'aws:sqs' },
        { id: 'rds', label: 'RDS Postgres', type: 'aws:rds' },
        { id: 'redis', label: 'ElastiCache', type: 'aws:elasticache' },
      ],
      edges: [
        { id: 'e1', source: 'cf', target: 'alb', label: 'HTTPS' },
        { id: 'e2', source: 'alb', target: 'nextjs' },
        { id: 'e3', source: 'nextjs', target: 'rds' },
        { id: 'e4', source: 'nextjs', target: 'redis' },
        { id: 'e5', source: 'nextjs', target: 'sqs', label: 'enqueue' },
        { id: 'e6', source: 'worker', target: 'sqs', label: 'poll', style: 'dashed' },
        { id: 'e7', source: 'worker', target: 'rds' },
      ],
      groups: [],
    },
  },
  {
    id: 'lambda-dynamodb',
    name: 'Lambda + DynamoDB',
    summary: 'API Gateway → Lambda → DynamoDB, serverless all the way.',
    tag: 'starter',
    prompt: 'A serverless API on AWS: API Gateway in front of Lambda handlers, DynamoDB for state, S3 for uploads.',
    diagram: {
      nodes: [
        { id: 'apigw', label: 'API Gateway', type: 'aws:apigateway' },
        { id: 'lambda-read', label: 'Read Lambda', type: 'aws:lambda' },
        { id: 'lambda-write', label: 'Write Lambda', type: 'aws:lambda' },
        { id: 'ddb', label: 'DynamoDB', type: 'aws:dynamodb' },
        { id: 's3', label: 'S3 uploads', type: 'aws:s3' },
      ],
      edges: [
        { id: 'e1', source: 'apigw', target: 'lambda-read', label: 'GET /*' },
        { id: 'e2', source: 'apigw', target: 'lambda-write', label: 'POST /*' },
        { id: 'e3', source: 'lambda-read', target: 'ddb' },
        { id: 'e4', source: 'lambda-write', target: 'ddb' },
        { id: 'e5', source: 'lambda-write', target: 's3' },
      ],
      groups: [],
    },
  },
  {
    id: 'go-redis',
    name: 'Go service + Redis',
    summary: 'Stateless Go API fronted by an NLB with Redis cache.',
    tag: 'app',
    prompt: 'A Go microservice behind a network load balancer with Redis cache and Postgres for durable state.',
    diagram: {
      nodes: [
        { id: 'nlb', label: 'NLB', type: 'aws:nlb' },
        { id: 'go', label: 'Go API (ECS)', type: 'aws:ecs' },
        { id: 'redis', label: 'ElastiCache', type: 'aws:elasticache' },
        { id: 'rds', label: 'RDS Postgres', type: 'aws:rds' },
        { id: 'cw', label: 'CloudWatch', type: 'aws:cloudwatch' },
      ],
      edges: [
        { id: 'e1', source: 'nlb', target: 'go', label: 'TCP 8080' },
        { id: 'e2', source: 'go', target: 'redis' },
        { id: 'e3', source: 'go', target: 'rds' },
        { id: 'e4', source: 'go', target: 'cw', label: 'logs', style: 'dashed' },
      ],
      groups: [],
    },
  },
  {
    id: 'rag-app',
    name: 'RAG assistant',
    summary: 'Bedrock + Pinecone-style vectors, S3 for docs, Lambda orchestration.',
    tag: 'ai',
    prompt: 'A retrieval-augmented assistant on AWS: S3 stores docs, an ingest Lambda chunks + embeds, a vector store holds the embeddings, an API Gateway + chat Lambda uses Bedrock to answer.',
    diagram: {
      nodes: [
        { id: 'apigw', label: 'API Gateway', type: 'aws:apigateway' },
        { id: 'chat', label: 'Chat Lambda', type: 'aws:lambda' },
        { id: 'ingest', label: 'Ingest Lambda', type: 'aws:lambda' },
        { id: 's3', label: 'S3 documents', type: 'aws:s3' },
        { id: 'vec', label: 'Vector store', type: 'aws:opensearch' },
        { id: 'bedrock', label: 'Bedrock', type: 'aws:bedrock' },
      ],
      edges: [
        { id: 'e1', source: 'apigw', target: 'chat', label: 'POST /chat' },
        { id: 'e2', source: 'chat', target: 'vec', label: 'nearest-k' },
        { id: 'e3', source: 'chat', target: 'bedrock', label: 'invoke' },
        { id: 'e4', source: 's3', target: 'ingest', label: 'put event', style: 'dashed' },
        { id: 'e5', source: 'ingest', target: 'bedrock', label: 'embed' },
        { id: 'e6', source: 'ingest', target: 'vec', label: 'upsert' },
      ],
      groups: [],
    },
  },
];
