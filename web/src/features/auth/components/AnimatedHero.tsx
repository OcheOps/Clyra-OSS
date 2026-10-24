import { useEffect, useRef, useState } from 'react';

/**
 * AnimatedHero — the landing page's one orchestrated motion moment.
 *
 * Cycles through a small set of prompts: each prompt types out
 * character-by-character at terminal speed, then the diagram materializes
 * (nodes stagger-fade in, edges draw via SVG dashoffset). After a held
 * beat, everything fades and the next prompt begins.
 *
 * The whole thing is one well-rehearsed loop. Per design-tokens.md rule:
 * "use glows sparingly — only on primary CTA and the hero canvas." This
 * IS that hero canvas, so the cyan glow lives here and nowhere else.
 *
 * Hand-tuned coordinates per scene rather than auto-layout — three
 * small scenes are easier to choreograph this way than to feed through a
 * layout engine.
 */

type Node = {
  id: string;
  label: string;
  kind: string; // shown as the small line under the label
  x: number;
  y: number;
};

type Edge = {
  from: string;
  to: string;
  // Path is intentionally manual so we can shape elbows nicely.
  d: string;
};

type Scene = {
  prompt: string;
  cost: string;
  nodes: Node[];
  edges: Edge[];
};

const SCENES: Scene[] = [
  {
    prompt: 'A three-tier web app on AWS — ALB, ECS Fargate, RDS Postgres, S3.',
    cost: '$342 / mo',
    nodes: [
      { id: 'r53',  label: 'Route 53',     kind: 'DNS',           x: 28,  y: 36 },
      { id: 'cf',   label: 'CloudFront',   kind: 'CDN',           x: 28,  y: 168 },
      { id: 'alb',  label: 'ALB',          kind: 'Load balancer', x: 196, y: 36 },
      { id: 'ecs',  label: 'ECS Fargate',  kind: '3 × app',       x: 364, y: 36 },
      { id: 'rds',  label: 'RDS',          kind: 'Postgres 16',   x: 532, y: 36 },
      { id: 's3',   label: 'S3',           kind: 'Static assets', x: 364, y: 168 },
    ],
    edges: [
      { from: 'r53', to: 'alb',  d: 'M 158 64 H 196' },
      { from: 'alb', to: 'ecs',  d: 'M 326 64 H 364' },
      { from: 'ecs', to: 'rds',  d: 'M 494 64 H 532' },
      { from: 'cf',  to: 's3',   d: 'M 158 196 H 364' },
      { from: 'ecs', to: 's3',   d: 'M 414 92 V 168' },
    ],
  },
  {
    prompt: 'A RAG assistant — API Gateway, Lambda, Bedrock, S3 vectors.',
    cost: '$89 / mo',
    nodes: [
      { id: 'usr', label: 'User',         kind: 'Browser',         x: 28,  y: 36 },
      { id: 'apg', label: 'API Gateway',  kind: 'REST',            x: 196, y: 36 },
      { id: 'lam', label: 'Lambda',       kind: 'Query handler',   x: 364, y: 36 },
      { id: 'br',  label: 'Bedrock',      kind: 'Claude · Titan',  x: 532, y: 36 },
      { id: 's3v', label: 'S3 Vectors',   kind: 'Knowledge base',  x: 364, y: 168 },
      { id: 'cw',  label: 'CloudWatch',   kind: 'Logs · metrics',  x: 532, y: 168 },
    ],
    edges: [
      { from: 'usr', to: 'apg', d: 'M 158 64 H 196' },
      { from: 'apg', to: 'lam', d: 'M 326 64 H 364' },
      { from: 'lam', to: 'br',  d: 'M 494 64 H 532' },
      { from: 'lam', to: 's3v', d: 'M 414 92 V 168' },
      { from: 'lam', to: 'cw',  d: 'M 494 92 Q 540 130 562 168' },
    ],
  },
  {
    prompt: 'An event pipeline — EventBridge, Kinesis, Lambda, DynamoDB.',
    cost: '$148 / mo',
    nodes: [
      { id: 'src', label: 'Sources',      kind: 'SaaS · IoT',      x: 28,  y: 36 },
      { id: 'eb',  label: 'EventBridge',  kind: 'Router',          x: 196, y: 36 },
      { id: 'kn',  label: 'Kinesis',      kind: 'Stream',          x: 364, y: 36 },
      { id: 'lam', label: 'Lambda',       kind: 'Stream consumer', x: 532, y: 36 },
      { id: 'dd',  label: 'DynamoDB',     kind: 'On-demand',       x: 364, y: 168 },
      { id: 's3a', label: 'S3 archive',   kind: 'Parquet',         x: 532, y: 168 },
    ],
    edges: [
      { from: 'src', to: 'eb',  d: 'M 158 64 H 196' },
      { from: 'eb',  to: 'kn',  d: 'M 326 64 H 364' },
      { from: 'kn',  to: 'lam', d: 'M 494 64 H 532' },
      { from: 'lam', to: 'dd',  d: 'M 562 92 Q 540 130 414 168' },
      { from: 'lam', to: 's3a', d: 'M 582 92 V 168' },
    ],
  },
];

// Tunable per-cycle timings. Keep them in one place so the dance is easy
// to nudge.
const TIMING = {
  typingMsPerChar: 28,
  holdAfterType:   400,
  nodesStaggerMs:  90,
  edgesStaggerMs:  160,
  holdAfterDiagram: 3600,
  fadeOutMs:       500,
};

// Phase order: TYPING → DIAGRAM → HOLD → FADE → next scene.
type Phase = 'typing' | 'diagram' | 'hold' | 'fade';

export function AnimatedHero() {
  const [sceneIdx, setSceneIdx] = useState(0);
  const [phase, setPhase] = useState<Phase>('typing');
  const [typedChars, setTypedChars] = useState(0);
  const [costShown, setCostShown] = useState(false);
  const timeoutsRef = useRef<ReturnType<typeof setTimeout>[]>([]);

  // `!` is safe: sceneIdx is always wrapped via `% SCENES.length` and
  // SCENES is a non-empty const, so this index always resolves. The
  // bang appeases TS strict-mode without adding a runtime guard.
  const scene = SCENES[sceneIdx]!;

  // Clear pending timeouts on scene change / unmount so a slow loop never
  // double-fires a phase transition.
  const schedule = (ms: number, fn: () => void) => {
    const t = setTimeout(fn, ms);
    timeoutsRef.current.push(t);
  };

  useEffect(() => {
    timeoutsRef.current.forEach(clearTimeout);
    timeoutsRef.current = [];

    if (!scene) return;

    setPhase('typing');
    setTypedChars(0);
    setCostShown(false);

    // Type out the prompt char-by-char.
    const total = scene.prompt.length;
    for (let i = 1; i <= total; i++) {
      schedule(i * TIMING.typingMsPerChar, () => setTypedChars(i));
    }

    const typingEndsAt = total * TIMING.typingMsPerChar + TIMING.holdAfterType;
    schedule(typingEndsAt, () => setPhase('diagram'));

    // Cost pill appears after nodes finish staggering in.
    const lastNodeAt = typingEndsAt + scene.nodes.length * TIMING.nodesStaggerMs;
    schedule(lastNodeAt + 200, () => setCostShown(true));

    const diagramEndsAt =
      lastNodeAt + scene.edges.length * TIMING.edgesStaggerMs + TIMING.holdAfterDiagram;
    schedule(diagramEndsAt, () => setPhase('fade'));

    schedule(diagramEndsAt + TIMING.fadeOutMs, () => {
      setSceneIdx((i) => (i + 1) % SCENES.length);
    });

    return () => {
      timeoutsRef.current.forEach(clearTimeout);
      timeoutsRef.current = [];
    };
  }, [sceneIdx, scene]);

  const typed = scene.prompt.slice(0, typedChars);
  const showDiagram = phase !== 'typing';
  const showFade = phase === 'fade';

  return (
    <div className="relative">
      {/* Subtle cyan glow ambient behind the frame */}
      <div
        aria-hidden
        className="pointer-events-none absolute -inset-16 opacity-40 blur-3xl"
        style={{
          background:
            'radial-gradient(60% 60% at 60% 40%, rgba(6,182,212,0.20) 0%, rgba(15,23,42,0) 70%)',
        }}
      />

      <div
        className={`relative overflow-hidden rounded-2xl border border-c2-border bg-c2-surface shadow-c2-lg transition-opacity duration-500 ${
          showFade ? 'opacity-0' : 'opacity-100'
        }`}
      >
        {/* Editor chrome */}
        <div className="flex items-center justify-between border-b border-c2-hairline bg-c2-bg/40 px-4 py-2.5">
          <div className="flex items-center gap-3">
            <span className="flex items-center gap-1.5">
              <span className="h-2 w-2 rounded-full bg-c2-danger/70" />
              <span className="h-2 w-2 rounded-full bg-c2-warning/70" />
              <span className="h-2 w-2 rounded-full bg-c2-accent/70" />
            </span>
            <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
              clyra · editor · example
            </span>
          </div>
          <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            solution · aws
          </span>
        </div>

        {/* Prompt bar — auto-typing */}
        <div className="flex items-start gap-3 border-b border-c2-hairline bg-c2-bg/30 px-4 py-3">
          <span className="mt-[3px] inline-block h-2 w-2 rounded-full bg-c2-primary shadow-[0_0_10px_rgba(6,182,212,0.7)]" />
          <span className="flex-1 font-brand text-petit leading-relaxed text-c2-text">
            {typed}
            <span
              aria-hidden
              className={`ml-0.5 inline-block h-[14px] w-[2px] -translate-y-[1px] bg-c2-primary align-middle ${
                phase === 'typing' ? 'animate-pulse' : 'opacity-0'
              }`}
            />
          </span>
        </div>

        {/* Canvas */}
        <div
          className="relative bg-c2-bg"
          style={{
            backgroundImage:
              'radial-gradient(circle at 1px 1px, rgba(148,163,184,0.10) 1px, transparent 0)',
            backgroundSize: '18px 18px',
          }}
        >
          <svg viewBox="0 0 720 240" className="block h-auto w-full" aria-hidden>
            <defs>
              <linearGradient id="edgeGradient" x1="0" y1="0" x2="1" y2="0">
                <stop offset="0%" stopColor="#06B6D4" />
                <stop offset="100%" stopColor="#10B981" />
              </linearGradient>
              <marker
                id="edgeArrow"
                viewBox="0 0 10 10"
                refX="8"
                refY="5"
                markerWidth="5"
                markerHeight="5"
                orient="auto"
              >
                <path d="M0,0 L10,5 L0,10 z" fill="#10B981" />
              </marker>
            </defs>

            {/* Edges — animated stroke draw */}
            {showDiagram &&
              scene.edges.map((e, i) => (
                <path
                  key={`${scene.prompt}-edge-${i}`}
                  d={e.d}
                  fill="none"
                  stroke="url(#edgeGradient)"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                  markerEnd="url(#edgeArrow)"
                  style={{
                    strokeDasharray: 240,
                    strokeDashoffset: 240,
                    animation: `edge-draw 0.8s ${
                      scene.nodes.length * TIMING.nodesStaggerMs + i * TIMING.edgesStaggerMs
                    }ms cubic-bezier(0.4,0,0.2,1) forwards`,
                  }}
                />
              ))}

            {/* Nodes */}
            {showDiagram &&
              scene.nodes.map((n, i) => (
                <g
                  key={`${scene.prompt}-node-${n.id}`}
                  style={{
                    opacity: 0,
                    animation: `node-pop 0.45s ${i * TIMING.nodesStaggerMs}ms cubic-bezier(0.34,1.4,0.64,1) forwards`,
                    transformOrigin: `${n.x + 65}px ${n.y + 28}px`,
                  }}
                >
                  <rect
                    x={n.x}
                    y={n.y}
                    width="130"
                    height="56"
                    rx="10"
                    fill="#1E293B"
                    stroke="#334155"
                    strokeWidth="1"
                  />
                  {/* Subtle inner top highlight */}
                  <rect
                    x={n.x + 1}
                    y={n.y + 1}
                    width="128"
                    height="1"
                    fill="rgba(255,255,255,0.04)"
                  />
                  <circle
                    cx={n.x + 14}
                    cy={n.y + 24}
                    r="4"
                    fill={i % 3 === 0 ? '#06B6D4' : i % 3 === 1 ? '#10B981' : '#67E8F9'}
                  />
                  <text
                    x={n.x + 26}
                    y={n.y + 28}
                    style={{
                      font: '600 12px Switzer, Inter, system-ui',
                      letterSpacing: '-0.005em',
                    }}
                    fill="#F8FAFC"
                  >
                    {n.label}
                  </text>
                  <text
                    x={n.x + 26}
                    y={n.y + 44}
                    style={{ font: '500 10px Switzer, Inter, system-ui' }}
                    fill="#64748B"
                  >
                    {n.kind}
                  </text>
                </g>
              ))}
          </svg>

          {/* Cost pill — counts in after diagram materializes */}
          <div
            className={`pointer-events-none absolute right-4 top-4 flex items-center gap-2 rounded-full border border-c2-accent/30 bg-c2-accent-soft px-3 py-1.5 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-accent backdrop-blur-sm transition-all duration-500 ${
              costShown ? 'translate-y-0 opacity-100' : '-translate-y-2 opacity-0'
            }`}
          >
            <span className="inline-block h-1.5 w-1.5 rounded-full bg-c2-accent shadow-[0_0_8px_rgba(16,185,129,0.7)]" />
            est. {scene.cost}
          </div>
        </div>

        {/* Footer chrome */}
        <div className="flex items-center justify-between border-t border-c2-hairline bg-c2-bg/40 px-4 py-2">
          <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            {scene.nodes.length} nodes · {scene.edges.length} edges · solution
          </span>
          <div className="flex items-center gap-2">
            <span className="rounded-md border border-c2-border bg-c2-bg/40 px-2 py-0.5 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-muted">
              Terraform
            </span>
            <span className="rounded-md border border-c2-border bg-c2-bg/40 px-2 py-0.5 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-muted">
              Save
            </span>
          </div>
        </div>
      </div>

      {/* Scene progress dots — quiet indicator at the bottom */}
      <div className="mt-4 flex justify-center gap-2">
        {SCENES.map((_, i) => (
          <span
            key={i}
            className={`h-[3px] rounded-full transition-all duration-500 ${
              i === sceneIdx
                ? 'w-8 bg-c2-primary shadow-[0_0_8px_rgba(6,182,212,0.5)]'
                : 'w-3 bg-c2-border'
            }`}
          />
        ))}
      </div>

      {/* Inline keyframes — kept local because they only matter to this
          component and we don't want to balloon the global stylesheet. */}
      <style>{`
        @keyframes edge-draw {
          from { stroke-dashoffset: 240; }
          to   { stroke-dashoffset: 0; }
        }
        @keyframes node-pop {
          from { opacity: 0; transform: scale(0.85) translateY(6px); }
          to   { opacity: 1; transform: scale(1)    translateY(0);    }
        }
      `}</style>
    </div>
  );
}
