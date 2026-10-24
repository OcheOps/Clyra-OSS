import { useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { AnimatedHero } from './AnimatedHero';

// HeroDemo — the interactive replacement for the pure-loop AnimatedHero
// on the landing page.
//
// Default state: renders the existing 3-scene loop (ambient marketing
// moment). When the visitor types a prompt and hits Generate, the loop
// swaps for a live render of what our backend returned via
// POST /api/v1/public/demo/generate — no signup, no signup wall.
//
// After a successful render we surface the "Like it? Sign up..." CTA
// under the canvas. That's the whole conversion loop: prove the product
// works on THEIR prompt in <15 seconds, then ask for the email.
//
// Failures (network, upstream Bedrock hiccup) get the backend's canned
// fallback diagram (3-tier web on AWS). Same UI treatment — we never
// show a raw error on the marketing surface.

const MAX_PROMPT = 400;

const SEEDS = [
  'A three-tier web app on AWS',
  'A serverless API with DynamoDB and Lambda',
  'A RAG assistant with S3, Bedrock, and OpenSearch',
];

type DemoNode = { id: string; label: string; type?: string };
type DemoEdge = { source: string; target: string };
type DemoResponse = {
  nodes: DemoNode[];
  edges: DemoEdge[];
  fallback?: boolean;
  message?: string;
};

type DemoState =
  | { kind: 'idle' }
  | { kind: 'loading' }
  | { kind: 'result'; data: DemoResponse; prompt: string }
  | { kind: 'error'; message: string };

export function HeroDemo() {
  const [prompt, setPrompt] = useState('');
  const [state, setState] = useState<DemoState>({ kind: 'idle' });

  const onPickSeed = (s: string) => {
    setPrompt(s);
  };

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const trimmed = prompt.trim();
    if (!trimmed) return;
    if (trimmed.length > MAX_PROMPT) return; // input maxLength handles it too
    setState({ kind: 'loading' });
    try {
      const res = await fetch('/api/v1/public/demo/generate', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ prompt: trimmed }),
      });
      if (res.status === 429) {
        setState({
          kind: 'error',
          message: 'Slow down — one demo per minute. Try again shortly.',
        });
        return;
      }
      const envelope = (await res.json()) as { data?: DemoResponse; error?: string };
      const data = envelope.data;
      // 503 fallback still carries a valid diagram in envelope.data.
      if (!data?.nodes?.length) {
        setState({ kind: 'error', message: 'Generation failed — try again.' });
        return;
      }
      setState({ kind: 'result', data, prompt: trimmed });
    } catch {
      setState({ kind: 'error', message: 'Network hiccup — try again.' });
    }
  };

  const reset = () => {
    setPrompt('');
    setState({ kind: 'idle' });
  };

  return (
    <div className="relative">
      {state.kind === 'idle' && <AnimatedHero />}
      {state.kind === 'loading' && <DemoSkeleton prompt={prompt} />}
      {state.kind === 'result' && (
        <DemoResult data={state.data} prompt={state.prompt} onReset={reset} />
      )}
      {state.kind === 'error' && (
        <DemoError message={state.message} onReset={reset} />
      )}

      {/* Prompt panel — pinned below the hero canvas so it's the first
          thing a visitor's eye lands on after the visual demo. Kept
          quiet by default: single input + button, seed prompts as
          subtle pills. */}
      <form onSubmit={onSubmit} className="mt-4">
        <div className="rounded-2xl border border-c2-border bg-c2-surface/60 p-3 shadow-c2 backdrop-blur">
          <div className="flex flex-wrap gap-1.5">
            {SEEDS.map((s) => (
              <button
                key={s}
                type="button"
                onClick={() => onPickSeed(s)}
                disabled={state.kind === 'loading'}
                className="rounded-full border border-c2-hairline bg-c2-bg/60 px-3 py-1 text-[11px] text-c2-text-muted transition-colors hover:border-c2-border-strong hover:text-c2-text disabled:opacity-40"
              >
                {s}
              </button>
            ))}
          </div>
          <div className="mt-2 flex items-center gap-2">
            <input
              type="text"
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              maxLength={MAX_PROMPT}
              placeholder="Try your own — describe a system in one sentence"
              disabled={state.kind === 'loading'}
              className="flex-1 rounded-lg border border-c2-hairline bg-c2-bg/70 px-3 py-2 text-sm text-c2-text placeholder:text-c2-text-muted/60 outline-none focus:border-c2-primary/50 focus:bg-c2-bg disabled:opacity-50"
            />
            <button
              type="submit"
              disabled={!prompt.trim() || state.kind === 'loading'}
              className="rounded-lg bg-c2-primary px-4 py-2 text-sm font-medium text-c2-bg shadow-c2-glow transition-all hover:bg-c2-primary-hover disabled:opacity-40 disabled:cursor-not-allowed"
            >
              {state.kind === 'loading' ? 'Generating…' : 'Generate'}
            </button>
          </div>
          <p className="mt-1.5 px-1 text-[10px] uppercase tracking-[0.16em] text-c2-text-faint">
            no signup · one demo per minute · powered by Clyra
          </p>
        </div>
      </form>
    </div>
  );
}

// ── Loading skeleton ──────────────────────────────────────────────

function DemoSkeleton({ prompt }: { prompt: string }) {
  return (
    <div className="relative overflow-hidden rounded-2xl border border-c2-border bg-c2-surface shadow-c2-lg">
      <div className="flex items-center justify-between border-b border-c2-hairline bg-c2-bg/40 px-4 py-2.5">
        <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
          clyra · generating…
        </span>
        <span className="inline-block h-2 w-2 animate-pulse rounded-full bg-c2-primary shadow-[0_0_10px_rgba(6,182,212,0.7)]" />
      </div>
      <div className="border-b border-c2-hairline bg-c2-bg/30 px-4 py-3 text-petit text-c2-text">
        <span className="mr-2 inline-block h-2 w-2 rounded-full bg-c2-primary/70 align-middle" />
        {prompt}
      </div>
      <div
        className="relative bg-c2-bg"
        style={{
          backgroundImage:
            'radial-gradient(circle at 1px 1px, rgba(148,163,184,0.10) 1px, transparent 0)',
          backgroundSize: '18px 18px',
        }}
      >
        <div className="grid h-[240px] grid-cols-4 gap-6 p-6">
          {Array.from({ length: 6 }).map((_, i) => (
            <div
              key={i}
              className="h-14 animate-pulse rounded-lg bg-c2-elevated/60 border border-c2-hairline"
              style={{ animationDelay: `${i * 120}ms` }}
            />
          ))}
        </div>
      </div>
    </div>
  );
}

// ── Result render (BFS-layered SVG) ───────────────────────────────
// Rather than pull in @xyflow/react on the marketing page, do a small
// BFS layer layout and draw the graph as inline SVG matching the
// AnimatedHero's visual system. Same node shape, edge stroke, and
// grid background so the demo blends into the loop it replaced.

function DemoResult({
  data,
  prompt,
  onReset,
}: {
  data: DemoResponse;
  prompt: string;
  onReset: () => void;
}) {
  const layout = layoutGraph(data.nodes, data.edges);
  const width = 720;
  const height = Math.max(240, layout.cols * 120 - 40);

  return (
    <div className="relative">
      <div className="relative overflow-hidden rounded-2xl border border-c2-border bg-c2-surface shadow-c2-lg">
        <div className="flex items-center justify-between border-b border-c2-hairline bg-c2-bg/40 px-4 py-2.5">
          <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            clyra · your prompt
          </span>
          <button
            type="button"
            onClick={onReset}
            className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-muted hover:text-c2-text"
          >
            ← reset
          </button>
        </div>
        <div className="border-b border-c2-hairline bg-c2-bg/30 px-4 py-3 text-petit text-c2-text">
          <span className="mr-2 inline-block h-2 w-2 rounded-full bg-c2-primary align-middle" />
          {prompt}
        </div>
        <div
          className="relative bg-c2-bg"
          style={{
            backgroundImage:
              'radial-gradient(circle at 1px 1px, rgba(148,163,184,0.10) 1px, transparent 0)',
            backgroundSize: '18px 18px',
          }}
        >
          <svg viewBox={`0 0 ${width} ${height}`} className="block h-auto w-full" aria-hidden>
            <defs>
              <linearGradient id="demoEdge" x1="0" y1="0" x2="1" y2="0">
                <stop offset="0%" stopColor="#06B6D4" />
                <stop offset="100%" stopColor="#10B981" />
              </linearGradient>
              <marker id="demoArrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="5" markerHeight="5" orient="auto">
                <path d="M0,0 L10,5 L0,10 z" fill="#10B981" />
              </marker>
            </defs>
            {data.edges.map((e, i) => {
              const from = layout.pos.get(e.source);
              const to = layout.pos.get(e.target);
              if (!from || !to) return null;
              return (
                <path
                  key={i}
                  d={pathFor(from, to)}
                  fill="none"
                  stroke="url(#demoEdge)"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                  markerEnd="url(#demoArrow)"
                />
              );
            })}
            {data.nodes.map((n) => {
              const p = layout.pos.get(n.id);
              if (!p) return null;
              return (
                <g key={n.id}>
                  <rect x={p.x} y={p.y} width="130" height="56" rx="10" fill="#1E293B" stroke="#334155" strokeWidth="1" />
                  <circle cx={p.x + 14} cy={p.y + 24} r="4" fill="#06B6D4" />
                  <text x={p.x + 26} y={p.y + 28} style={{ font: '600 12px Switzer, Inter, system-ui' }} fill="#F8FAFC">
                    {truncate(n.label, 14)}
                  </text>
                  <text x={p.x + 26} y={p.y + 44} style={{ font: '500 10px Switzer, Inter, system-ui' }} fill="#64748B">
                    {n.type ?? ''}
                  </text>
                </g>
              );
            })}
          </svg>
        </div>
        <div className="flex items-center justify-between border-t border-c2-hairline bg-c2-bg/40 px-4 py-2">
          <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            {data.nodes.length} nodes · {data.edges.length} edges
            {data.fallback ? ' · canned example' : ''}
          </span>
          <Link
            to="/signup"
            className="rounded-full bg-c2-primary px-3 py-1 font-mono text-[10px] font-medium uppercase tracking-[0.18em] text-c2-bg shadow-c2-glow hover:bg-c2-primary-hover"
          >
            {data.fallback ? 'Sign up to run yours' : 'Sign up · 10/day free'}
          </Link>
        </div>
      </div>
      {data.fallback && (
        <p className="mt-2 text-center text-micro text-c2-warning">
          {data.message || "Couldn't reach the generator — showing a canned example."}
          {' '}Sign up and try your own prompt.
        </p>
      )}
      {!data.fallback && data.message && (
        <p className="mt-2 text-center text-micro text-c2-text-faint">{data.message}</p>
      )}
    </div>
  );
}

function DemoError({ message, onReset }: { message: string; onReset: () => void }) {
  return (
    <div className="relative overflow-hidden rounded-2xl border border-c2-border bg-c2-surface shadow-c2-lg">
      <div className="flex flex-col items-center justify-center gap-3 px-6 py-12 text-center">
        <p className="text-tiny text-c2-text">{message}</p>
        <button
          type="button"
          onClick={onReset}
          className="rounded-lg border border-c2-border bg-c2-bg/60 px-3 py-1.5 text-tiny font-medium text-c2-text hover:border-c2-border-strong"
        >
          Reset
        </button>
      </div>
    </div>
  );
}

// ── BFS layer layout ──────────────────────────────────────────────
// Chosen over dagre to keep the landing bundle lean. Roots first
// (in-degree 0), then breadth-first; each layer becomes a column,
// nodes distributed vertically within their column with a small
// gutter. Enough structure for a 6-12 node demo.

function layoutGraph(
  nodes: DemoNode[],
  edges: DemoEdge[],
): { pos: Map<string, { x: number; y: number }>; cols: number } {
  const inDeg = new Map<string, number>();
  nodes.forEach((n) => inDeg.set(n.id, 0));
  edges.forEach((e) => {
    if (inDeg.has(e.target)) inDeg.set(e.target, (inDeg.get(e.target) ?? 0) + 1);
  });
  const adj = new Map<string, string[]>();
  edges.forEach((e) => {
    const list = adj.get(e.source) ?? [];
    list.push(e.target);
    adj.set(e.source, list);
  });

  const layer = new Map<string, number>();
  const queue: string[] = [];
  nodes.forEach((n) => {
    if ((inDeg.get(n.id) ?? 0) === 0) {
      layer.set(n.id, 0);
      queue.push(n.id);
    }
  });
  // Fallback for cyclic — assign layer 0 to any unlayered node.
  while (queue.length) {
    const id = queue.shift()!;
    const l = layer.get(id) ?? 0;
    for (const t of adj.get(id) ?? []) {
      const cur = layer.get(t);
      if (cur === undefined || cur <= l) {
        layer.set(t, l + 1);
        queue.push(t);
      }
    }
  }
  nodes.forEach((n) => {
    if (!layer.has(n.id)) layer.set(n.id, 0);
  });

  const cols = Math.max(...Array.from(layer.values())) + 1;
  const byCol: Record<number, string[]> = {};
  nodes.forEach((n) => {
    const l = layer.get(n.id)!;
    (byCol[l] ??= []).push(n.id);
  });

  const pos = new Map<string, { x: number; y: number }>();
  const colWidth = 168;
  const rowHeight = 90;
  const startX = 24;
  Object.entries(byCol).forEach(([col, ids]) => {
    const c = Number(col);
    const totalHeight = ids.length * rowHeight;
    const startY = Math.max(30, (240 - totalHeight) / 2);
    ids.forEach((id, row) => {
      pos.set(id, { x: startX + c * colWidth, y: startY + row * rowHeight });
    });
  });
  return { pos, cols };
}

function pathFor(from: { x: number; y: number }, to: { x: number; y: number }): string {
  const fx = from.x + 130;
  const fy = from.y + 28;
  const tx = to.x;
  const ty = to.y + 28;
  if (Math.abs(fy - ty) < 4) return `M ${fx} ${fy} H ${tx}`;
  const mid = (fx + tx) / 2;
  return `M ${fx} ${fy} H ${mid} V ${ty} H ${tx}`;
}

function truncate(s: string, n: number): string {
  return s.length <= n ? s : s.slice(0, n - 1) + '…';
}
