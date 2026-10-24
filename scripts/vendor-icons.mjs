#!/usr/bin/env node
// Vendor the official provider architecture-icon packs into
// web/public/icons/<provider>/ for offline-first, no-CDN serving.
//
// Run:  node scripts/vendor-icons.mjs            # all providers
//       node scripts/vendor-icons.mjs aws gcp    # specific subset
//       node scripts/vendor-icons.mjs --dry-run  # show what would download
//
// Sources:
//   - AWS  : icon.icepanel.io/AWS/svg   (same CDN behind awsicons.dev)
//   - GCP  : icon.icepanel.io/GCP/svg   (same CDN behind gcpicons.com)
//   - Azure: github.com/benc-uk/icon-collection (mirrors Microsoft's
//            official architecture icons; Microsoft's own URL rotates)
//   - OCI  : github.com/oracle/oci-svg-icons (Oracle's official set)
//   - Brand marks: Iconify's open-source logos collection
//
// Output:
//   web/public/icons/<provider>/<category>/<name>.svg
//   web/public/icons/manifest.json  ← canonical → relative path
//
// The manifest is the contract the registry consumes. After running
// this, the icons.ts BRAND_OVERRIDES / ICON_REGISTRY can be rewired
// to `local:<provider>/<category>/<name>.svg` paths — IconRenderer
// already has a `cdn:` dispatch; we add a sibling `local:` path that
// resolves against /icons/.

import { mkdir, writeFile, readFile } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import https from 'node:https';

const ROOT = resolve(fileURLToPath(import.meta.url), '..', '..', 'web', 'public', 'icons');

const args = process.argv.slice(2);
const dryRun = args.includes('--dry-run');
const onlyProviders = args.filter((a) => !a.startsWith('--'));
const wants = (p) => onlyProviders.length === 0 || onlyProviders.includes(p);

// ─── AWS — icepanel.io ──────────────────────────────────────────────
// Manifest is (category, name, canonical). Canonical maps back to
// the `aws.X` ids in components/*.md so the registry can wire it.
const AWS_MANIFEST = [
  ['Compute', 'EC2', 'aws.ec2'],
  ['Compute', 'Lambda', 'aws.lambda'],
  ['Compute', 'Fargate', 'aws.fargate'],
  ['Compute', 'Batch', 'aws.batch'],
  ['Compute', 'Lightsail', 'aws.lightsail'],
  ['Compute', 'Elastic-Beanstalk', 'aws.elastic_beanstalk'],
  ['Containers', 'Elastic-Container-Service', 'aws.ecs'],
  ['Containers', 'Elastic-Kubernetes-Service', 'aws.eks'],
  ['Containers', 'Elastic-Container-Registry', 'aws.ecr'],
  ['Storage', 'Simple-Storage-Service', 'aws.s3'],
  ['Storage', 'EFS', 'aws.efs'],
  ['Storage', 'Backup', 'aws.backup'],
  ['Storage', 'Storage-Gateway', 'aws.storage_gateway'],
  ['Storage', 'S3-on-Outposts', 'aws.s3_outposts'],
  ['Database', 'RDS', 'aws.rds'],
  ['Database', 'DynamoDB', 'aws.dynamodb'],
  ['Database', 'Aurora', 'aws.aurora'],
  ['Database', 'ElastiCache', 'aws.elasticache'],
  ['Database', 'Neptune', 'aws.neptune'],
  ['Database', 'DocumentDB', 'aws.documentdb'],
  ['Database', 'Timestream', 'aws.timestream'],
  ['Database', 'Keyspaces', 'aws.keyspaces'],
  ['Networking-Content-Delivery', 'CloudFront', 'aws.cloudfront'],
  ['Networking-Content-Delivery', 'Route-53', 'aws.route_53'],
  ['Networking-Content-Delivery', 'Virtual-Private-Cloud', 'aws.vpc'],
  ['Networking-Content-Delivery', 'Elastic-Load-Balancing', 'aws.elb'],
  ['Networking-Content-Delivery', 'Direct-Connect', 'aws.direct_connect'],
  ['Networking-Content-Delivery', 'Transit-Gateway', 'aws.transit_gateway'],
  ['App-Integration', 'API-Gateway', 'aws.api_gateway'],
  ['App-Integration', 'Simple-Queue-Service', 'aws.sqs'],
  ['App-Integration', 'Simple-Notification-Service', 'aws.sns'],
  ['App-Integration', 'Step-Functions', 'aws.step_functions'],
  ['App-Integration', 'EventBridge', 'aws.eventbridge'],
  ['App-Integration', 'AppFlow', 'aws.appflow'],
  ['App-Integration', 'AppSync', 'aws.appsync'],
  ['App-Integration', 'MQ', 'aws.mq'],
  ['Security-Identity-Compliance', 'Identity-and-Access-Management', 'aws.iam'],
  ['Security-Identity-Compliance', 'Cognito', 'aws.cognito'],
  ['Security-Identity-Compliance', 'Secrets-Manager', 'aws.secrets_manager'],
  ['Security-Identity-Compliance', 'Shield', 'aws.shield'],
  ['Security-Identity-Compliance', 'WAF', 'aws.waf'],
  ['Security-Identity-Compliance', 'GuardDuty', 'aws.guardduty'],
  ['Security-Identity-Compliance', 'Key-Management-Service', 'aws.kms'],
  ['Security-Identity-Compliance', 'Certificate-Manager', 'aws.certificate_manager'],
  ['Analytics', 'Kinesis', 'aws.kinesis'],
  ['Analytics', 'Redshift', 'aws.redshift'],
  ['Analytics', 'Athena', 'aws.athena'],
  ['Analytics', 'OpenSearch-Service', 'aws.opensearch'],
  ['Analytics', 'Glue', 'aws.glue'],
  ['Analytics', 'Managed-Streaming-for-Apache-Kafka', 'aws.msk'],
  ['Analytics', 'QuickSight', 'aws.quicksight'],
  ['Analytics', 'EMR', 'aws.emr'],
  ['Analytics', 'Lake-Formation', 'aws.lake_formation'],
  ['Machine-Learning', 'SageMaker', 'aws.sagemaker'],
  ['Management-Governance', 'CloudWatch', 'aws.cloudwatch'],
  ['Management-Governance', 'CloudTrail', 'aws.cloudtrail'],
  ['Management-Governance', 'Systems-Manager', 'aws.systems_manager'],
  ['Management-Governance', 'CloudFormation', 'aws.cloudformation'],
  ['Management-Governance', 'Config', 'aws.config'],
  ['Management-Governance', 'Auto-Scaling', 'aws.auto_scaling'],
  ['Front-End-Web-Mobile', 'Amplify', 'aws.amplify'],
  ['Developer-Tools', 'CodeBuild', 'aws.codebuild'],
  ['Developer-Tools', 'CodeCommit', 'aws.codecommit'],
  ['Developer-Tools', 'CodeDeploy', 'aws.codedeploy'],
  ['Developer-Tools', 'CodePipeline', 'aws.codepipeline'],
  ['Developer-Tools', 'X-Ray', 'aws.xray'],
  ['Business-Applications', 'Simple-Email-Service', 'aws.ses'],
];

// ─── GCP — icepanel.io ──────────────────────────────────────────────
const GCP_MANIFEST = [
  ['', 'Cloud-Run', 'gcp.cloud_run'],
  ['', 'Cloud-Functions', 'gcp.cloud_functions'],
  ['', 'Compute-Engine', 'gcp.compute_engine'],
  ['', 'Google-Kubernetes-Engine', 'gcp.gke'],
  ['', 'App-Engine', 'gcp.app_engine'],
  ['', 'Cloud-Storage', 'gcp.cloud_storage'],
  ['', 'Cloud-SQL', 'gcp.cloud_sql'],
  ['', 'BigQuery', 'gcp.bigquery'],
  ['', 'Bigtable', 'gcp.bigtable'],
  ['', 'Cloud-Spanner', 'gcp.spanner'],
  ['', 'Memorystore', 'gcp.memorystore'],
  ['', 'Firestore', 'gcp.firestore'],
  ['', 'Cloud-CDN', 'gcp.cloud_cdn'],
  ['', 'Cloud-Load-Balancing', 'gcp.cloud_load_balancing'],
  ['', 'Cloud-DNS', 'gcp.cloud_dns'],
  ['', 'Cloud-NAT', 'gcp.cloud_nat'],
  ['', 'Cloud-Armor', 'gcp.cloud_armor'],
  ['', 'Vertex-AI', 'gcp.vertex_ai'],
  ['', 'PubSub', 'gcp.pubsub'],
  ['', 'Secret-Manager', 'gcp.secret_manager'],
  ['', 'Cloud-Build', 'gcp.cloud_build'],
  ['', 'Artifact-Registry', 'gcp.artifact_registry'],
  ['', 'Container-Registry', 'gcp.container_registry'],
  ['', 'Cloud-Tasks', 'gcp.cloud_tasks'],
  ['', 'Cloud-Scheduler', 'gcp.cloud_scheduler'],
  ['', 'Cloud-Endpoints', 'gcp.cloud_endpoints'],
  ['', 'Dataflow', 'gcp.dataflow'],
  ['', 'Dataproc', 'gcp.dataproc'],
  ['', 'Cloud-Composer', 'gcp.composer'],
  ['', 'Cloud-Logging', 'gcp.logging'],
  ['', 'Cloud-Monitoring', 'gcp.monitoring'],
  ['', 'Identity-And-Access-Management', 'gcp.iam'],
  ['', 'Eventarc', 'gcp.eventarc'],
];

const BRAND_MANIFEST = [
  ['logos', 'aws-glacier'],
  ['logos', 'microsoft-azure'],
  ['logos', 'kubernetes'],
  ['logos', 'cloudflare'],
  ['logos', 'cloudflare-workers'],
];

// ─── Azure / CF / OCI ───────────────────────────────────────────────
// These don't have a clean icon.icepanel.io tree, so we mirror upstream
// GitHub zips. The script downloads + unzips them under
// public/icons/<provider>/raw/, then you (or a follow-up commit) write
// the canonical→file map manually since their naming is irregular.
const ZIP_MIRRORS = {
  azure: {
    url: 'https://github.com/benc-uk/icon-collection/archive/refs/heads/master.zip',
    extractGlob: 'icon-collection-master/azure-patterns/**/*.svg',
    note: 'Mirror of Microsoft Azure architecture icons. Browse benc-uk/icon-collection on GitHub for layout.',
  },
  oci: {
    url: 'https://github.com/oracle-quickstart/oci-data-science-ai-samples/archive/refs/heads/main.zip',
    note: 'OCI symbol set — fallback. Oracle does not publish a stable zip URL.',
  },
  cloudflare: {
    url: 'https://github.com/cloudflare/cf-ui/archive/refs/heads/master.zip',
    note: 'No first-party Cloudflare architecture icon pack exists; you may prefer to keep iconify:logos:cloudflare for these.',
  },
};

// ─── Runtime ────────────────────────────────────────────────────────

const CDN = 'https://icon.icepanel.io';

function get(url) {
  return new Promise((resolve, reject) => {
    https
      .get(url, (res) => {
        if (res.statusCode === 301 || res.statusCode === 302) {
          return resolve(get(res.headers.location));
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`${res.statusCode} ${url}`));
        }
        const chunks = [];
        res.on('data', (c) => chunks.push(c));
        res.on('end', () => resolve(Buffer.concat(chunks)));
        res.on('error', reject);
      })
      .on('error', reject);
  });
}

async function ensureDir(p) {
  if (!existsSync(p)) await mkdir(p, { recursive: true });
}

async function fetchOne(provider, category, name) {
  const cdnPath = category ? `${provider}/svg/${category}/${name}.svg` : `${provider}/svg/${name}.svg`;
  const dest = category
    ? join(ROOT, provider.toLowerCase(), category.toLowerCase(), `${name}.svg`)
    : join(ROOT, provider.toLowerCase(), `${name}.svg`);
  if (existsSync(dest)) return { skipped: true, dest };
  if (dryRun) return { dry: true, dest, cdnPath };
  await ensureDir(dirname(dest));
  const body = await get(`${CDN}/${cdnPath}`);
  await writeFile(dest, body);
  return { written: body.length, dest };
}

async function sweepIcepanel(provider, manifest, canonicalToPath) {
  const results = { ok: 0, fail: 0, skipped: 0, dry: 0 };
  for (const [category, name, canonical] of manifest) {
    try {
      const r = await fetchOne(provider, category, name);
      if (r.skipped) results.skipped++;
      else if (r.dry) {
        results.dry++;
        console.log(`  [dry] ${provider}/${category}/${name}`);
      } else {
        results.ok++;
        process.stdout.write(`.`);
      }
      const rel = category
        ? `/icons/${provider.toLowerCase()}/${category.toLowerCase()}/${name}.svg`
        : `/icons/${provider.toLowerCase()}/${name}.svg`;
      canonicalToPath[canonical] = rel;
    } catch (e) {
      results.fail++;
      console.warn(`\n  [fail] ${provider}/${category}/${name}: ${e.message}`);
    }
  }
  console.log(`\n${provider}: ${results.ok} new, ${results.skipped} cached, ${results.fail} failed, ${results.dry} dry`);
  return results;
}

async function sweepBrands() {
  const results = { ok: 0, skipped: 0, dry: 0 };
  for (const [prefix, name] of BRAND_MANIFEST) {
    const dest = join(ROOT, 'brands', prefix, `${name}.svg`);
    if (existsSync(dest)) {
      results.skipped++;
      continue;
    }
    if (dryRun) {
      results.dry++;
      console.log(`  [dry] ${prefix}/${name}`);
      continue;
    }
    await ensureDir(dirname(dest));
    const body = await get(`https://api.iconify.design/${prefix}/${name}.svg`);
    await writeFile(dest, body);
    results.ok++;
    process.stdout.write('.');
  }
  console.log(`\nBrands: ${results.ok} new, ${results.skipped} cached, ${results.dry} dry`);
}

async function main() {
  await ensureDir(ROOT);
  const manifestPath = join(ROOT, 'manifest.json');
  const canonicalToPath = existsSync(manifestPath)
    ? JSON.parse(await readFile(manifestPath, 'utf8'))
    : {};

  if (wants('aws')) {
    console.log('AWS via icepanel CDN…');
    await sweepIcepanel('AWS', AWS_MANIFEST, canonicalToPath);
  }
  if (wants('gcp')) {
    console.log('GCP via icepanel CDN…');
    await sweepIcepanel('GCP', GCP_MANIFEST, canonicalToPath);
  }
  if (wants('brands')) {
    console.log('Brand marks via Iconify…');
    await sweepBrands();
  }
  for (const provider of ['azure', 'oci', 'cloudflare']) {
    if (!wants(provider)) continue;
    const m = ZIP_MIRRORS[provider];
    console.log(`\n${provider.toUpperCase()}: no icepanel coverage.`);
    console.log(`  Manual step — download ${m.url}`);
    console.log(`  Place SVGs under web/public/icons/${provider}/`);
    if (m.note) console.log(`  Note: ${m.note}`);
  }

  if (!dryRun) {
    await writeFile(manifestPath, JSON.stringify(canonicalToPath, null, 2));
    console.log(`\nWrote ${manifestPath}`);
    console.log(`${Object.keys(canonicalToPath).length} canonicals mapped.`);
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
