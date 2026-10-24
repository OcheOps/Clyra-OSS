#!/usr/bin/env node
// scripts/cost-coverage.mjs — snapshot how much of the component catalog
// has a cost estimate, grouped by level + provider. Run before + after
// a refresh to see what you moved. Not a CI gate; a signal.
//
//   node scripts/cost-coverage.mjs
//
// Produces a compact table so we can see the beginner tier is at 100%
// coverage, intermediate at X%, etc.

import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

const HERE = dirname(fileURLToPath(import.meta.url));
const REPO = resolve(HERE, '..');

const catalog = readFileSync(resolve(REPO, 'web/src/lib/generated/componentCatalog.ts'), 'utf8');
const cost = readFileSync(resolve(REPO, 'web/src/lib/costCatalog.ts'), 'utf8');

const priced = new Set(
  Array.from(cost.matchAll(/'([a-z0-9]+\.[a-z_0-9]+)'/g), (m) => m[1]),
);

const rows = [];
const rowRx = /\{\s*canonical:\s*"([^"]+)"[^}]*provider:\s*"([^"]+)"[^}]*level:\s*"([^"]+)"/g;
let m;
while ((m = rowRx.exec(catalog))) {
  rows.push({ c: m[1], p: m[2], l: m[3] });
}

const levels = ['beginner', 'intermediate', 'advanced', 'expert'];
const providers = [...new Set(rows.map((r) => r.p))].sort();

console.log(`Total canonicals: ${rows.length}  ·  priced: ${priced.size}  ·  coverage: ${((priced.size / rows.length) * 100).toFixed(1)}%\n`);

// By level
console.log('BY LEVEL');
for (const level of levels) {
  const total = rows.filter((r) => r.l === level);
  const covered = total.filter((r) => priced.has(r.c));
  const pct = total.length ? ((covered.length / total.length) * 100).toFixed(0) : '—';
  const bar = '█'.repeat(Math.round((covered.length / (total.length || 1)) * 20)).padEnd(20, '·');
  console.log(`  ${level.padEnd(14)} ${bar} ${covered.length}/${total.length} (${pct}%)`);
}
console.log();

// By provider × level (unpriced count)
console.log('UNPRICED BY PROVIDER × LEVEL');
console.log(`  ${'provider'.padEnd(10)}${levels.map((l) => l.slice(0, 4).padStart(6)).join('')}`);
for (const p of providers) {
  const cells = levels.map((l) => {
    const n = rows.filter((r) => r.p === p && r.l === l && !priced.has(r.c)).length;
    return String(n || '·').padStart(6);
  });
  console.log(`  ${p.padEnd(10)}${cells.join('')}`);
}
