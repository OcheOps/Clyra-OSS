#!/usr/bin/env node
// Cost catalog staleness checker.
//
// Reads LAST_REVIEWED from web/src/lib/costCatalog.ts, compares against
// REFRESH_INTERVAL_DAYS, and prints either:
//   - "OK, next refresh due YYYY-MM-DD" and exits 0, OR
//   - "STALE" with a grouped spot-check URL list and exits 1.
//
// Run with:  npm run cost:refresh
//
// Designed to be wired into a weekly CI cron later — exit code is the
// signal. No AWS API auth, no scraping, no fragile selectors. The script
// is a forcing function for a human to spot-check the ~30 entries
// flagged by family, not an attempt to derive prices automatically.

import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const HERE = dirname(fileURLToPath(import.meta.url));
const CATALOG = join(HERE, '..', 'web', 'src', 'lib', 'costCatalog.ts');

const src = readFileSync(CATALOG, 'utf8');

const lastReviewedMatch = src.match(/LAST_REVIEWED\s*=\s*'(\d{4}-\d{2}-\d{2})'/);
const intervalMatch = src.match(/REFRESH_INTERVAL_DAYS\s*=\s*(\d+)/);
if (!lastReviewedMatch || !intervalMatch) {
  console.error('Could not find LAST_REVIEWED / REFRESH_INTERVAL_DAYS in costCatalog.ts');
  process.exit(2);
}

const lastReviewed = new Date(lastReviewedMatch[1] + 'T00:00:00Z');
const intervalDays = Number(intervalMatch[1]);
const now = new Date();
const ageDays = Math.floor((now.getTime() - lastReviewed.getTime()) / 86_400_000);
const dueDate = new Date(lastReviewed.getTime() + intervalDays * 86_400_000);
const isStale = ageDays >= intervalDays;

// Count entries per family for the spot-check report.
const familyCounts = {};
for (const m of src.matchAll(/family:\s*'([a-z]+)'/g)) {
  familyCounts[m[1]] = (familyCounts[m[1]] ?? 0) + 1;
}

// One canonical pricing URL per family — the SRE doing the refresh
// opens these, ctrl-F the SKUs they care about, updates the dollar
// figures in costCatalog.ts. The list is opinionated: AWS-first since
// our coverage is AWS-heaviest. Add more if a future SKU type lands.
const SPOT_CHECK_URLS = {
  compute: [
    'https://aws.amazon.com/ec2/pricing/on-demand/',
    'https://aws.amazon.com/fargate/pricing/',
    'https://aws.amazon.com/lambda/pricing/',
    'https://cloud.google.com/compute/all-pricing',
    'https://azure.microsoft.com/en-us/pricing/details/virtual-machines/linux/',
  ],
  storage: [
    'https://aws.amazon.com/s3/pricing/',
    'https://aws.amazon.com/efs/pricing/',
    'https://cloud.google.com/storage/pricing',
  ],
  database: [
    'https://aws.amazon.com/rds/pricing/',
    'https://aws.amazon.com/dynamodb/pricing/',
    'https://aws.amazon.com/elasticache/pricing/',
    'https://cloud.google.com/sql/pricing',
  ],
  network: [
    'https://aws.amazon.com/elasticloadbalancing/pricing/',
    'https://aws.amazon.com/api-gateway/pricing/',
    'https://aws.amazon.com/vpc/pricing/',
  ],
  cdn: [
    'https://aws.amazon.com/cloudfront/pricing/',
    'https://www.cloudflare.com/plans/',
  ],
  messaging: [
    'https://aws.amazon.com/sqs/pricing/',
    'https://aws.amazon.com/msk/pricing/',
    'https://aws.amazon.com/kinesis/data-streams/pricing/',
  ],
  analytics: [
    'https://aws.amazon.com/redshift/pricing/',
    'https://aws.amazon.com/athena/pricing/',
    'https://cloud.google.com/bigquery/pricing',
  ],
  ai: [
    'https://aws.amazon.com/bedrock/pricing/',
    'https://aws.amazon.com/sagemaker/pricing/',
    'https://openai.com/api/pricing/',
    'https://www.anthropic.com/pricing#api',
  ],
  security: [
    'https://aws.amazon.com/waf/pricing/',
    'https://aws.amazon.com/guardduty/pricing/',
    'https://aws.amazon.com/secrets-manager/pricing/',
  ],
  ops: [
    'https://aws.amazon.com/cloudwatch/pricing/',
    'https://aws.amazon.com/config/pricing/',
  ],
};

const DIM = '\x1b[2m';
const RED = '\x1b[31m';
const GREEN = '\x1b[32m';
const YELLOW = '\x1b[33m';
const RESET = '\x1b[0m';

const fmtDate = (d) => d.toISOString().slice(0, 10);

if (!isStale) {
  console.log(`${GREEN}OK${RESET} cost catalog reviewed ${ageDays} days ago`);
  console.log(`${DIM}  last reviewed:  ${fmtDate(lastReviewed)}${RESET}`);
  console.log(`${DIM}  next due:       ${fmtDate(dueDate)} (in ${intervalDays - ageDays} days)${RESET}`);
  process.exit(0);
}

console.log(`${RED}STALE${RESET} cost catalog hasn't been reviewed in ${ageDays} days (threshold: ${intervalDays})`);
console.log(`${DIM}  last reviewed:  ${fmtDate(lastReviewed)}${RESET}`);
console.log(`${DIM}  was due:        ${fmtDate(dueDate)}${RESET}`);
console.log();
console.log(`${YELLOW}Refresh procedure:${RESET}`);
console.log(`  1. For each family below, open the linked pricing pages.`);
console.log(`  2. ctrl-F each SKU in web/src/lib/costCatalog.ts under that family.`);
console.log(`  3. Adjust the 'usd' values where the public price has moved >10%.`);
console.log(`  4. Update LAST_REVIEWED in costCatalog.ts to today's date.`);
console.log(`  5. Commit with a message listing which SKUs moved and why.`);
console.log();
console.log(`${YELLOW}Spot-check URLs by family${RESET} (entry counts in parens):`);
console.log();

const families = Object.keys(familyCounts).sort();
for (const family of families) {
  const urls = SPOT_CHECK_URLS[family] ?? [];
  console.log(`  ${family} (${familyCounts[family]})`);
  if (urls.length === 0) {
    console.log(`    ${DIM}— no canonical URLs yet, search vendor docs${RESET}`);
  } else {
    for (const u of urls) console.log(`    ${DIM}- ${u}${RESET}`);
  }
}

process.exit(1);
