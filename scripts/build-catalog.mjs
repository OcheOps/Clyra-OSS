#!/usr/bin/env node
// Generates web/src/lib/generated/componentCatalog.ts from components/*.md.
// Prints coverage diff against icons.ts and costCatalog.ts so it's obvious
// what still needs hand-curated icons / pricing.
//
// Run: node scripts/build-catalog.mjs
//
// Rules:
//   * One file, all ~660 entries, deduped by canonical.
//   * Markdown tables must have the header `| Canonical | Display | Sub | Lvl | Description | Aliases | Example |`.
//   * Canonical can be wrapped in backticks; backticks are stripped.
//   * Aliases are comma-separated; commas inside aliases aren't supported.
//   * Section H2 becomes the category; "Sub" column becomes subcategory.

import { readFileSync, writeFileSync, readdirSync, mkdirSync, existsSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const repoRoot = join(__dirname, '..');
const componentsDir = join(repoRoot, 'components');
const outputDir = join(repoRoot, 'web/src/lib/generated');
const goOutputPath = join(repoRoot, 'internal/diagram/catalog_generated.go');
const componentIdsPath = join(outputDir, 'componentIds.ts');
const lucideImportsPath = join(outputDir, 'lucideIconImports.ts');
const costCatalogPath = join(repoRoot, 'web/src/lib/costCatalog.ts');
const iconsPath = join(repoRoot, 'web/src/lib/icons.ts');
const iconHeuristicsPath = join(repoRoot, 'web/src/lib/iconHeuristics.ts');
const publicIconsRoot = join(repoRoot, 'web/public/icons');

const LEVEL_MAP = { B: 'beginner', I: 'intermediate', A: 'advanced', E: 'expert' };

// Skip index / patterns files; they're not data tables.
const SKIP = new Set(['README.md', 'patterns.md']);

function normalizeCategory(heading) {
  return heading
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '_')
    .replace(/^_+|_+$/g, '');
}

function parseTables(content) {
  const lines = content.split('\n');
  const rows = [];
  let inTable = false;
  let sawSeparator = false;
  let currentCategory = 'misc';

  for (const raw of lines) {
    const line = raw.trimEnd();

    if (line.startsWith('## ')) {
      currentCategory = normalizeCategory(line.slice(3));
      inTable = false;
      sawSeparator = false;
      continue;
    }

    if (!line.startsWith('|')) {
      inTable = false;
      sawSeparator = false;
      continue;
    }

    const cells = line.split('|').slice(1, -1).map((c) => c.trim());

    if (!inTable) {
      if (cells[0]?.toLowerCase() === 'canonical') {
        inTable = true;
        sawSeparator = false;
      }
      continue;
    }

    if (!sawSeparator) {
      // The --- separator row follows the header.
      if (cells[0]?.startsWith('---')) sawSeparator = true;
      continue;
    }

    if (cells.length < 7) continue;
    const [rawCanonical, display, sub, lvl, desc, aliases, example] = cells;
    const canonical = rawCanonical.replace(/^`|`$/g, '').trim();
    if (!canonical || !canonical.includes('.')) continue;

    const provider = canonical.split('.')[0];
    const level = LEVEL_MAP[lvl] ?? 'intermediate';
    const aliasArr = aliases
      .split(',')
      .map((a) => a.trim())
      .filter(Boolean);

    rows.push({
      canonical,
      display,
      provider,
      category: currentCategory,
      subcategory: sub,
      level,
      description: desc,
      aliases: aliasArr,
      example,
    });
  }

  return rows;
}

function generateTs(entries) {
  // Pretty but compact.
  const body = entries
    .map(
      (e) =>
        `  { canonical: ${JSON.stringify(e.canonical)}, display: ${JSON.stringify(e.display)}, provider: ${JSON.stringify(e.provider)}, category: ${JSON.stringify(e.category)}, subcategory: ${JSON.stringify(e.subcategory)}, level: ${JSON.stringify(e.level)}, description: ${JSON.stringify(e.description)}, aliases: ${JSON.stringify(e.aliases)}, example: ${JSON.stringify(e.example)} }`,
    )
    .join(',\n');

  return `// GENERATED FILE — do not edit by hand.
// Regenerate: \`node scripts/build-catalog.mjs\` from repo root.
// Source of truth: components/*.md

export type ComponentProvider =
  | 'aws'
  | 'azure'
  | 'gcp'
  | 'k8s'
  | 'net'
  | 'data'
  | 'ent'
  | 'edge'
  | 'iot';

export type ComponentLevel = 'beginner' | 'intermediate' | 'advanced' | 'expert';

export interface ComponentEntry {
  canonical: string;
  display: string;
  provider: ComponentProvider;
  category: string;
  subcategory: string;
  level: ComponentLevel;
  description: string;
  aliases: string[];
  example: string;
}

// Cast bypasses TS2590 — with ~1 000 literal rows the inferred union
// becomes "too complex to represent". We widen to ComponentEntry[] so
// downstream code keeps full type safety without the O(n) inference work.
export const COMPONENTS: ComponentEntry[] = ([
${body}
] as unknown) as ComponentEntry[];

export const BY_CANONICAL: Record<string, ComponentEntry> = (() => {
  const m: Record<string, ComponentEntry> = {};
  for (const c of COMPONENTS) m[c.canonical] = c;
  return m;
})();

// Flat alias → canonical lookup used by the prompt normalizer as a
// first-pass exact-match (case-insensitive). Trigram similarity is the
// fallback when nothing in here hits.
export const ALIAS_INDEX: Record<string, string> = (() => {
  const out: Record<string, string> = {};
  for (const c of COMPONENTS) {
    const keys = [c.canonical, c.display, ...c.aliases];
    for (const k of keys) {
      const key = k.toLowerCase().trim();
      if (!key || out[key]) continue;
      out[key] = c.canonical;
    }
  }
  return out;
})();

export function lookupAlias(phrase: string): string | undefined {
  return ALIAS_INDEX[phrase.toLowerCase().trim()];
}

export const COMPONENTS_BY_PROVIDER: Record<ComponentProvider, ComponentEntry[]> = (() => {
  const g: Record<string, ComponentEntry[]> = {};
  for (const c of COMPONENTS) (g[c.provider] ??= []).push(c);
  return g as Record<ComponentProvider, ComponentEntry[]>;
})();
`;
}

function generateGo(entries) {
  const aliases = new Map();
  for (const entry of entries) {
    for (const value of [entry.display, ...entry.aliases]) {
      const key = value.toLowerCase().trim();
      if (key && !aliases.has(key)) aliases.set(key, entry.canonical);
    }
  }
  const canonicalLiterals = entries.map((entry) => JSON.stringify(entry.canonical));
  const canonicalWidth = Math.max(...canonicalLiterals.map((value) => value.length));
  const canonicalRows = canonicalLiterals
    .map((value) => `\t${value}:${' '.repeat(canonicalWidth - value.length + 1)}{},`)
    .join('\n');
  const aliasEntries = [...aliases.entries()].sort(([a], [b]) => a.localeCompare(b));
  const aliasWidth = Math.max(...aliasEntries.map(([alias]) => JSON.stringify(alias).length));
  const aliasRows = aliasEntries
    .map(([alias, canonical]) => {
      const value = JSON.stringify(alias);
      return `\t${value}:${' '.repeat(aliasWidth - value.length + 1)}${JSON.stringify(canonical)},`;
    })
    .join('\n');
  return `// Code generated by scripts/build-catalog.mjs; DO NOT EDIT.\n\npackage diagram\n\nvar generatedCatalogTypes = map[string]struct{}{\n${canonicalRows}\n}\n\nvar generatedCatalogAliases = map[string]string{\n${aliasRows}\n}\n`;
}

function generateComponentIds(entries) {
  const rows = entries.map((entry) => `  ${JSON.stringify(entry.canonical)},`).join('\n');
  return `// GENERATED FILE — do not edit by hand.\nexport const COMPONENT_IDS = new Set<string>([\n${rows}\n]);\n`;
}

function generateLucideImports() {
  const sources = [readFileSync(iconsPath, 'utf8'), readFileSync(iconHeuristicsPath, 'utf8')].join('\n');
  const candidates = new Set();
  for (const pattern of [/icon:\s*'([^']+)'/g, /:\s*'([^']+)'/g, /,\s*'([^']+)'\]/g]) {
    for (const match of sources.matchAll(pattern)) candidates.add(match[1]);
  }
  const iconRoot = join(repoRoot, 'web/node_modules/lucide-react/dist/esm/icons');
  const icons = [...candidates]
    .filter((name) => existsSync(join(iconRoot, `${name}.js`)))
    .sort((a, b) => a.localeCompare(b));
  const rows = icons
    .map((name) => `  ${JSON.stringify(name)}: () => import('lucide-react/dist/esm/icons/${name}.js'),`)
    .join('\n');
  return `// GENERATED FILE — do not edit by hand.\nexport const lucideIconImports = {\n${rows}\n};\n`;
}

function validateVendoredIcons(iconsSrc) {
  const required = new Set();
  const manifest = JSON.parse(readFileSync(join(publicIconsRoot, 'manifest.json'), 'utf8'));
  for (const path of Object.values(manifest)) required.add(path.replace(/^\/icons\//, ''));
  for (const match of iconsSrc.matchAll(/['"]cdn:([^'"]+)['"]/g)) {
    const parts = match[1].split('/');
    const provider = parts.shift().toLowerCase();
    if (parts[0]?.toLowerCase() === 'svg') parts.shift();
    if (provider === 'aws' && parts.length > 1) parts[0] = parts[0].toLowerCase();
    required.add([provider, ...parts].join('/'));
  }
  for (const match of iconsSrc.matchAll(/['"]iconify:([^'"]+)['"]/g)) {
    required.add(`brands/${match[1].replace(':', '/')}.svg`);
  }
  const missing = [...required].filter((path) => !existsSync(join(publicIconsRoot, path)));
  if (missing.length > 0) {
    throw new Error(`missing vendored icon files:\n${missing.map((path) => `  ${path}`).join('\n')}`);
  }
  return required.size;
}

function coverage(entries, iconsSrc, costsSrc) {
  const missingIcons = [];
  const missingCosts = [];
  for (const e of entries) {
    // icons.ts uses single-quote keys, costCatalog.ts too.
    const needle = `'${e.canonical}'`;
    if (!iconsSrc.includes(needle)) missingIcons.push(e);
    if (!costsSrc.includes(needle)) missingCosts.push(e);
  }
  return { missingIcons, missingCosts };
}

function groupBy(arr, key) {
  const out = {};
  for (const a of arr) (out[a[key]] ??= []).push(a);
  return out;
}

function main() {
  const files = readdirSync(componentsDir).filter((f) => f.endsWith('.md') && !SKIP.has(f));

  const seen = new Set();
  const all = [];
  const perFile = {};

  for (const file of files) {
    const content = readFileSync(join(componentsDir, file), 'utf8');
    const rows = parseTables(content);
    perFile[file] = rows.length;
    for (const row of rows) {
      if (seen.has(row.canonical)) continue;
      seen.add(row.canonical);
      all.push(row);
    }
  }

  // Sort for stable output
  all.sort((a, b) => a.canonical.localeCompare(b.canonical));

  mkdirSync(outputDir, { recursive: true });
  writeFileSync(join(outputDir, 'componentCatalog.ts'), generateTs(all));
  writeFileSync(componentIdsPath, generateComponentIds(all));
  writeFileSync(lucideImportsPath, generateLucideImports());
  writeFileSync(goOutputPath, generateGo(all));

  const iconsSrc = existsSync(iconsPath) ? readFileSync(iconsPath, 'utf8') : '';
  const costsSrc = existsSync(costCatalogPath) ? readFileSync(costCatalogPath, 'utf8') : '';
  const { missingIcons, missingCosts } = coverage(all, iconsSrc, costsSrc);
  const vendoredIconCount = validateVendoredIcons(iconsSrc);

  const byLevel = groupBy(all, 'level');
  const byProvider = groupBy(all, 'provider');

  console.log(`✓ Parsed ${all.length} components from ${files.length} files`);
  for (const [f, n] of Object.entries(perFile)) console.log(`    ${f.padEnd(18)} ${n}`);
  console.log(`\n→ wrote web/src/lib/generated/componentCatalog.ts`);
  console.log(`→ wrote web/src/lib/generated/componentIds.ts`);
  console.log(`→ wrote web/src/lib/generated/lucideIconImports.ts`);
  console.log(`→ wrote internal/diagram/catalog_generated.go`);

  console.log(`\nBy provider:`);
  for (const [p, cs] of Object.entries(byProvider).sort((a, b) => b[1].length - a[1].length)) {
    console.log(`    ${p.padEnd(6)} ${cs.length}`);
  }

  console.log(`\nBy level:`);
  for (const lvl of ['beginner', 'intermediate', 'advanced', 'expert']) {
    console.log(`    ${lvl.padEnd(14)} ${(byLevel[lvl] ?? []).length}`);
  }

  console.log(`\nCoverage vs existing hand-curated files:`);
  console.log(`    vendored files  ${vendoredIconCount} verified`);
  console.log(`    icons.ts       ${all.length - missingIcons.length}/${all.length}  (missing ${missingIcons.length})`);
  console.log(`    costCatalog.ts ${all.length - missingCosts.length}/${all.length}  (missing ${missingCosts.length})`);

  if (missingIcons.length > 0) {
    const g = groupBy(missingIcons, 'provider');
    console.log(`\n    Icons missing by provider:`);
    for (const [p, cs] of Object.entries(g).sort((a, b) => b[1].length - a[1].length)) {
      console.log(`      ${p.padEnd(6)} ${cs.length}`);
    }
  }
  if (missingCosts.length > 0) {
    const g = groupBy(missingCosts, 'provider');
    console.log(`\n    Costs missing by provider:`);
    for (const [p, cs] of Object.entries(g).sort((a, b) => b[1].length - a[1].length)) {
      console.log(`      ${p.padEnd(6)} ${cs.length}`);
    }
  }

  // Write a machine-readable coverage report for CI or follow-up scripts.
  const report = {
    total: all.length,
    byProvider: Object.fromEntries(Object.entries(byProvider).map(([k, v]) => [k, v.length])),
    byLevel: Object.fromEntries(Object.entries(byLevel).map(([k, v]) => [k, v.length])),
    missingIcons: missingIcons.map((c) => c.canonical),
    missingCosts: missingCosts.map((c) => c.canonical),
  };
  writeFileSync(join(outputDir, 'coverage.json'), JSON.stringify(report, null, 2));
  console.log(`\n→ wrote web/src/lib/generated/coverage.json`);
}

main();
