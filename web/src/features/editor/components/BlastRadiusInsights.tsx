import { useBlastRadius } from '../lib/blastRadiusContext';
import { useEditorStore } from '../store';
import type { Severity } from '../lib/blastRadius';

// Floating insights panel that appears whenever blast-radius focus is set.
// Sits on the right edge of the canvas so it doesn't fight the right-side
// NodeProperties inspector for real estate — the two show at different
// times (Properties on selection, Insights on blast focus). Escape closes.
//
// The panel does three things the canvas alone can't:
//   1. Ranks affected nodes by severity + distance so the user's eye
//      lands on the worst-hit part first ("your DB is on fire, THEN look
//      at the app tier").
//   2. Suggests concrete mitigations tied to node types (multi-AZ,
//      autoscaling, secondary DNS). Not "consider redundancy" hand-waving.
//   3. Estimates rough downtime cost per hour so the business case for
//      fixing the flagged SPOF has a number attached, not just a vibe.

export function BlastRadiusInsights() {
  const { focus, affected, mitigations, cost } = useBlastRadius();
  const setBlastRadiusFocus = useEditorStore((s) => s.setBlastRadiusFocus);

  if (!focus || affected.length === 0) return null;

  const epicenter = affected.find((a) => a.distance === 0);
  const downstream = affected.filter((a) => a.distance > 0);

  return (
    <aside
      // Positioned inside the canvas frame so it inherits the editor's
      // dark substrate + hairline border language. Fixed max-height so
      // very large blast radii scroll rather than push the panel off-
      // screen on short viewports.
      className="pointer-events-auto absolute right-4 top-4 z-20 flex w-80 max-h-[calc(100vh-8rem)] flex-col overflow-hidden rounded-2xl border border-c2-border bg-c2-elevated/95 backdrop-blur-md shadow-c2-lg"
      role="complementary"
      aria-label="Blast radius insights"
    >
      <header className="flex items-start justify-between gap-3 border-b border-c2-hairline px-4 py-3">
        <div className="min-w-0">
          <div className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            Blast radius
          </div>
          <div className="mt-1 truncate text-sm font-semibold text-c2-text" title={epicenter?.label}>
            {epicenter?.label ?? focus}
          </div>
          <div className="mt-0.5 font-mono text-micro text-c2-text-faint">
            {epicenter?.nodeType}
          </div>
        </div>
        <button
          type="button"
          onClick={() => setBlastRadiusFocus(null)}
          className="shrink-0 rounded-lg p-1 text-c2-text-muted transition-colors hover:bg-c2-bg/60 hover:text-c2-text"
          title="Exit blast view (Esc)"
          aria-label="Close blast radius panel"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
        </button>
      </header>

      <div className="flex-1 overflow-y-auto">
        {/* Headline stats — 3 numbers the presenter can read out loud in
            a demo without pausing to think. */}
        <div className="grid grid-cols-3 gap-px bg-c2-hairline">
          <Stat label="Nodes" value={affected.length.toString()} />
          <Stat label="Direct" value={downstream.filter((d) => d.distance === 1).length.toString()} />
          <Stat
            label="$/hr"
            value={cost.usdPerHour > 0 ? `$${cost.usdPerHour.toLocaleString()}` : '—'}
            tone={cost.usdPerHour > 2000 ? 'danger' : cost.usdPerHour > 500 ? 'warn' : 'neutral'}
          />
        </div>

        {/* Downstream ranked by severity. */}
        {downstream.length > 0 && (
          <Section title="Also breaks">
            <ol className="space-y-1.5">
              {downstream.map((a) => (
                <li key={a.id} className="flex items-center gap-2">
                  <SeverityDot severity={a.severity} />
                  <span className="min-w-0 flex-1 truncate text-tiny text-c2-text" title={a.label}>
                    {a.label}
                  </span>
                  <span className="font-mono text-[9px] uppercase tracking-[0.15em] text-c2-text-faint">
                    {a.distance === 1 ? 'direct' : `${a.distance} hops`}
                  </span>
                </li>
              ))}
            </ol>
          </Section>
        )}

        {/* Concrete fixes based on affected node types. */}
        {mitigations.length > 0 && (
          <Section title="Mitigations">
            <ul className="space-y-2.5">
              {mitigations.map((m) => (
                <li key={m.nodeId + m.title} className="rounded-lg border border-c2-hairline bg-c2-bg/30 p-2.5">
                  <div className="flex items-start gap-2">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="mt-0.5 shrink-0 text-c2-accent">
                      <path d="M20 6L9 17l-5-5" />
                    </svg>
                    <div className="min-w-0">
                      <p className="text-tiny font-medium text-c2-text">{m.title}</p>
                      <p className="mt-1 text-micro leading-relaxed text-c2-text-muted">
                        {m.rationale}
                      </p>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          </Section>
        )}

        {/* Cost breakdown — only render when we actually have numbers. */}
        {cost.breakdown.length > 0 && (
          <Section title="Downtime cost">
            <ul className="space-y-1">
              {cost.breakdown.slice(0, 6).map((b) => (
                <li key={b.nodeId} className="flex items-baseline justify-between gap-2">
                  <span className="min-w-0 flex-1 truncate font-mono text-[10px] uppercase tracking-[0.14em] text-c2-text-faint" title={b.nodeType}>
                    {b.nodeType}
                  </span>
                  <span className="font-mono text-tiny tabular-nums text-c2-text-muted">
                    ${b.usdPerHour.toLocaleString()}/hr
                  </span>
                </li>
              ))}
            </ul>
            <p className="mt-2 font-mono text-[9px] uppercase tracking-[0.14em] text-c2-text-faint">
              Rough estimates. Adjust in blastRadius.ts to match your unit economics.
            </p>
          </Section>
        )}
      </div>
    </aside>
  );
}

// ─── primitives ─────────────────────────────────────────────────────

function Stat({
  label,
  value,
  tone = 'neutral',
}: {
  label: string;
  value: string;
  tone?: 'neutral' | 'warn' | 'danger';
}) {
  const toneClass =
    tone === 'danger' ? 'text-c2-danger'
    : tone === 'warn' ? 'text-amber-300'
    : 'text-c2-text';
  return (
    <div className="bg-c2-elevated px-3 py-2.5">
      <div className="font-mono text-[9px] uppercase tracking-[0.18em] text-c2-text-faint">
        {label}
      </div>
      <div className={`mt-0.5 font-display text-lg font-semibold tabular-nums ${toneClass}`}>
        {value}
      </div>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="border-b border-c2-hairline px-4 py-3 last:border-b-0">
      <h3 className="mb-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
        {title}
      </h3>
      {children}
    </section>
  );
}

function SeverityDot({ severity }: { severity: Severity }) {
  const color =
    severity === 'critical' ? 'bg-c2-danger'
    : severity === 'high' ? 'bg-amber-400'
    : severity === 'medium' ? 'bg-c2-primary'
    : 'bg-c2-text-faint';
  return (
    <span
      className={`inline-block h-2 w-2 shrink-0 rounded-full ${color}`}
      title={severity}
      aria-label={`Severity: ${severity}`}
    />
  );
}
