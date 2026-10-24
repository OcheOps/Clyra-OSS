import { describe, it, expect } from 'vitest';
import type { Edge, Node } from '@xyflow/react';
import { computeBlastRadius } from './blastRadius';

// Tiny helper to build a node — we only care about ID for traversal so
// position + data get reasonable defaults.
function n(id: string): Node {
  return {
    id,
    position: { x: 0, y: 0 },
    data: { label: id, nodeType: 'generic.service' },
  };
}

function e(source: string, target: string): Edge {
  return { id: `${source}-${target}`, source, target };
}

describe('computeBlastRadius', () => {
  // Convention: edges go from caller → callee (subnet → vpc, api → db).
  // So if a node fails, the things that BREAK are everything reachable by
  // following edges BACKWARDS from it. computeBlastRadius models that
  // dependency-failure relationship.

  it('returns just the focused node when nothing depends on it', () => {
    const nodes = [n('a'), n('b'), n('c')];
    const edges = [e('a', 'b'), e('a', 'c')];
    // a depends on b and c. If a fails, nothing else breaks (no incoming).
    const r = computeBlastRadius(nodes, edges, 'a');
    expect([...r.affectedIds]).toEqual(['a']);
  });

  it('finds direct dependents (one hop incoming)', () => {
    const nodes = [n('vpc'), n('subnet1'), n('subnet2'), n('alb')];
    const edges = [e('subnet1', 'vpc'), e('subnet2', 'vpc'), e('alb', 'subnet1')];
    // vpc fails → subnet1, subnet2 break (direct dep). alb depends on subnet1
    // → also breaks (transitive).
    const r = computeBlastRadius(nodes, edges, 'vpc');
    expect(new Set(r.affectedIds)).toEqual(new Set(['vpc', 'subnet1', 'subnet2', 'alb']));
  });

  it('handles transitive multi-hop chains', () => {
    // user → api → db → cache (each depends on the next)
    const nodes = [n('user'), n('api'), n('db'), n('cache')];
    const edges = [e('user', 'api'), e('api', 'db'), e('db', 'cache')];
    // cache fails: db, api, user all break.
    const r = computeBlastRadius(nodes, edges, 'cache');
    expect(new Set(r.affectedIds)).toEqual(new Set(['cache', 'db', 'api', 'user']));
  });

  it('does not follow outgoing edges', () => {
    // db → secrets_manager (db depends on secrets). If secrets fails, db
    // breaks — secrets is what we're killing. But if db fails, secrets is
    // fine — it has no upstream dependency on db. Catches the
    // directionality bug.
    const nodes = [n('db'), n('secrets_manager')];
    const edges = [e('db', 'secrets_manager')];
    const fromDB = computeBlastRadius(nodes, edges, 'db');
    expect([...fromDB.affectedIds]).toEqual(['db']);
    const fromSecrets = computeBlastRadius(nodes, edges, 'secrets_manager');
    expect(new Set(fromSecrets.affectedIds)).toEqual(new Set(['secrets_manager', 'db']));
  });

  it('survives cycles without infinite-looping', () => {
    // Mutual dependency (rare but legal — multi-master DBs etc).
    const nodes = [n('a'), n('b'), n('c')];
    const edges = [e('a', 'b'), e('b', 'c'), e('c', 'a')];
    const r = computeBlastRadius(nodes, edges, 'a');
    // Whole cycle is in the same blast radius.
    expect(new Set(r.affectedIds)).toEqual(new Set(['a', 'b', 'c']));
  });

  it('returns empty result for an unknown source ID', () => {
    const r = computeBlastRadius([n('a')], [], 'does-not-exist');
    expect(r.affectedIds.size).toBe(0);
  });

  it('flags SPOFs as nodes whose blast radius covers ≥25% of the diagram', () => {
    // 4-node graph; if any one node takes down 1 (25%) including itself,
    // it qualifies. Build a star: vpc has 3 direct dependents.
    const nodes = [n('vpc'), n('s1'), n('s2'), n('s3')];
    const edges = [e('s1', 'vpc'), e('s2', 'vpc'), e('s3', 'vpc')];
    const r = computeBlastRadius(nodes, edges, 'vpc');
    // vpc kills everything → SPOF.
    expect(r.spofs.has('vpc')).toBe(true);
    // leaves only kill themselves (1/4 = 25%, threshold is >25% so not SPOF).
    expect(r.spofs.has('s1')).toBe(false);
  });

  it('counts only the affected size for SPOF detection, not the focused node', () => {
    // Two leaves that depend on a shared parent. Parent's blast = self + 2
    // leaves = 3/3 nodes, definitely SPOF. Leaves kill only themselves.
    const nodes = [n('p'), n('a'), n('b')];
    const edges = [e('a', 'p'), e('b', 'p')];
    const r = computeBlastRadius(nodes, edges, 'p');
    expect(r.affectedIds.size).toBe(3);
    expect(r.spofs.has('p')).toBe(true);
    expect(r.spofs.has('a')).toBe(false);
    expect(r.spofs.has('b')).toBe(false);
  });
});
