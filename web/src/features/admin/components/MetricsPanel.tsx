import { useEffect, useState } from 'react';
import { type AdminMetrics, getAdminMetrics } from '@/services/api/admin';

/**
 * MetricsPanel — v2 restyle.
 *
 * Activation dashboard for the admin tab. Three sections: top-line user
 * counts, plan split, and an activation funnel (signup → first diagram
 * → first IaC export). Stat tiles render as v2 cards; the funnel rows
 * pick up a cyan→emerald progress gradient that doubles as a visual
 * health indicator (you want the bars long).
 */
export function MetricsPanel() {
  const [data, setData] = useState<AdminMetrics | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    getAdminMetrics().then(setData).catch(() => setError(true));
  }, []);

  if (error) {
    return (
      <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-danger">
        Failed to load metrics.
      </p>
    );
  }
  if (!data) {
    return (
      <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
        Loading…
      </p>
    );
  }

  const funnel = data.activation_funnel ?? {};
  const signups = funnel.signup ?? data.total_users;
  const firstDiagram = funnel.first_diagram ?? 0;
  const firstExport = funnel.first_iac_export ?? 0;
  const pctDiagram = signups > 0 ? Math.round((firstDiagram / signups) * 100) : 0;
  const pctExport  = signups > 0 ? Math.round((firstExport / signups) * 100) : 0;

  return (
    <div className="space-y-10">
      <Section title="Users">
        <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
          <Stat label="Total users" value={data.total_users} />
          <Stat label="Signups · 7d" value={data.signups_last_7d} />
          <Stat label="Signups · 30d" value={data.signups_last_30d} />
          <Stat
            label="Paid users"
            value={(data.users_by_plan.pro ?? 0) + (data.users_by_plan.byo ?? 0)}
            highlight
          />
        </div>
      </Section>

      <Section title="Plan split">
        <div className="grid grid-cols-3 gap-3">
          <Stat label="Free" value={data.users_by_plan.free ?? 0} />
          <Stat label="Pro"  value={data.users_by_plan.pro ?? 0} />
          <Stat label="BYO"  value={data.users_by_plan.byo ?? 0} />
        </div>
      </Section>

      <Section title="Activation funnel">
        <div className="space-y-2">
          <FunnelRow label="Signed up" value={signups} pct={100} />
          <FunnelRow label="Created first diagram" value={firstDiagram} pct={pctDiagram} />
          <FunnelRow label="Exported first IaC" value={firstExport} pct={pctExport} />
        </div>
        <p className="mt-4 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
          Activation tracks the first time each user reaches a step · lifetime cumulative, not period counts
        </p>
      </Section>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section>
      <div className="flex items-center gap-2">
        <span className="inline-block h-1 w-1 rounded-full bg-c2-primary" />
        <h2 className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
          {title}
        </h2>
      </div>
      <div className="mt-4">{children}</div>
    </section>
  );
}

function Stat({
  label,
  value,
  highlight,
}: {
  label: string;
  value: number;
  highlight?: boolean;
}) {
  return (
    <div
      className={`rounded-xl border p-4 transition-colors ${
        highlight
          ? 'border-c2-primary/40 bg-c2-primary-soft'
          : 'border-c2-border bg-c2-surface/60 hover:border-c2-border-strong'
      }`}
    >
      <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
        {label}
      </p>
      <p
        className={`mt-2 font-display text-2xl font-semibold tracking-[-0.025em] tabular-nums ${
          highlight ? 'text-c2-primary' : 'text-c2-text'
        }`}
      >
        {value.toLocaleString()}
      </p>
    </div>
  );
}

function FunnelRow({
  label,
  value,
  pct,
}: {
  label: string;
  value: number;
  pct: number;
}) {
  return (
    <div className="rounded-xl border border-c2-border bg-c2-surface/60 px-4 py-3">
      <div className="flex items-center justify-between text-sm">
        <span className="text-c2-text">{label}</span>
        <span className="font-mono tabular-nums text-c2-text-muted">
          {value.toLocaleString()}{' '}
          <span className="text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            · {pct}%
          </span>
        </span>
      </div>
      <div className="mt-2.5 h-1.5 overflow-hidden rounded-full bg-c2-bg/60">
        <div
          className="h-full rounded-full bg-gradient-to-r from-c2-primary to-c2-accent transition-all duration-700"
          style={{ width: `${Math.min(100, pct)}%` }}
        />
      </div>
    </div>
  );
}
