import { useState } from 'react';
import { useArchitectureInsights } from '../hooks/useArchitectureInsights';
import { useEditorStore } from '../store';
import { useInsightsStore } from '../insightsStore';
import type { Finding, Severity } from '@/lib/architectureRules';
import type { CostEntry, NodeCost } from '@/lib/costCatalog';

type Tab = 'cost' | 'review';

// Floating panel pinned to the right of the canvas. Always-on architecture
// intelligence: live monthly cost + rule violations. Collapsible so it stays
// out of the way when you're drafting.
export function InsightsPanel() {
  const [open, setOpen] = useState(true);
  const [tab, setTab] = useState<Tab>('cost');
  const { cost, validation } = useArchitectureInsights();
  const visible = useInsightsStore((s) => s.visible);
  const hide = useInsightsStore((s) => s.hide);

  const highCount = validation.countsBySeverity.high;
  const mediumCount = validation.countsBySeverity.medium;

  if (!visible) return null;

  return (
    <div className="pointer-events-auto absolute right-3 top-3 z-10 w-[320px] max-w-[calc(100vw-48px)]">
      <div className="overflow-hidden rounded-clyra-lg border border-line-default bg-clyra-raised/95 shadow-clyra backdrop-blur-md">
        <div className="flex items-stretch border-b border-line-subtle bg-clyra-raised/80">
          <button
            type="button"
            onClick={() => setOpen((v) => !v)}
            className="flex flex-1 items-center justify-between gap-2 px-3 py-2 text-left hover:bg-clyra-overlay/50 transition-colors"
          >
            <div className="flex items-center gap-2">
              <svg className="h-3.5 w-3.5 text-brand-400" viewBox="0 0 20 20" fill="currentColor">
                <path d="M10 2a1 1 0 011 1v.5a6.5 6.5 0 11-4.9 11.3 1 1 0 011.4-1.4 4.5 4.5 0 104.5-7.9V6a1 1 0 01-2 0V3a1 1 0 011-1z" />
              </svg>
              <span className="text-petit font-semibold text-ink-primary">Insights</span>
              <span className="text-micro text-ink-faint">
                ${cost.total}/mo
              </span>
              {(highCount + mediumCount) > 0 && (
                <span className="ml-1 rounded-full border border-danger/30 bg-danger/10 px-1.5 py-[1px] text-micro text-danger">
                  {highCount + mediumCount} issue{highCount + mediumCount === 1 ? '' : 's'}
                </span>
              )}
            </div>
            <svg
              className={`h-3.5 w-3.5 text-ink-muted transition-transform ${open ? 'rotate-180' : ''}`}
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" />
            </svg>
          </button>
          <button
            type="button"
            onClick={hide}
            title="Hide Insights (reopen from the ⋯ menu)"
            aria-label="Hide Insights"
            className="flex w-8 items-center justify-center border-l border-line-subtle text-ink-muted hover:text-ink-primary hover:bg-clyra-overlay/50 transition-colors"
          >
            <svg className="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round">
              <path d="M6 6l12 12M18 6L6 18" />
            </svg>
          </button>
        </div>

        {open && (
          <>
            <div className="flex border-b border-line-subtle px-1 pt-1">
              <TabButton active={tab === 'cost'} onClick={() => setTab('cost')}>
                Cost
              </TabButton>
              <TabButton active={tab === 'review'} onClick={() => setTab('review')}>
                Review
                {validation.findings.length > 0 && (
                  <span className="ml-1 text-micro text-ink-faint">
                    {validation.findings.length}
                  </span>
                )}
              </TabButton>
            </div>

            <div className="max-h-[60vh] overflow-y-auto">
              {tab === 'cost' ? (
                <CostView priced={cost.priced} unpriced={cost.unpriced} total={cost.total} byFamily={cost.byFamily} />
              ) : (
                <ReviewView findings={validation.findings} />
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
}

function TabButton({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex items-center gap-1 rounded-clyra px-2.5 py-1 text-tiny font-medium transition-colors ${
        active
          ? 'bg-clyra-bg text-ink-primary'
          : 'text-ink-muted hover:text-ink-primary'
      }`}
    >
      {children}
    </button>
  );
}

function CostView({
  priced,
  unpriced,
  total,
  byFamily,
}: {
  priced: NodeCost[];
  unpriced: NodeCost[];
  total: number;
  byFamily: Array<{ family: CostEntry['family']; total: number }>;
}) {
  if (priced.length === 0 && unpriced.length === 0) {
    return (
      <div className="p-4 text-petit text-ink-muted">
        Add nodes to see a cost estimate.
      </div>
    );
  }

  return (
    <div className="p-3 space-y-4">
      <div>
        <div className="flex items-baseline justify-between">
          <span className="text-micro uppercase tracking-wider text-ink-faint">Estimated monthly</span>
          <span className="text-lg font-semibold text-ink-primary">${total.toLocaleString()}</span>
        </div>
        <p className="mt-1 text-micro text-ink-faint">
          Small-production assumptions. Not a billing forecast.
        </p>
      </div>

      {byFamily.length > 0 && (
        <div>
          <p className="mb-1.5 text-micro uppercase tracking-wider text-ink-faint">By family</p>
          <div className="space-y-1">
            {byFamily.map((f) => (
              <div key={f.family} className="flex items-center justify-between gap-2">
                <div className="flex-1">
                  <div className="h-1 overflow-hidden rounded-full bg-clyra-bg">
                    <div
                      className="h-full bg-brand-500/60"
                      style={{ width: `${total > 0 ? (f.total / total) * 100 : 0}%` }}
                    />
                  </div>
                  <span className="mt-0.5 inline-block text-micro capitalize text-ink-muted">{f.family}</span>
                </div>
                <span className="text-micro text-ink-secondary">${f.total}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      <div>
        <p className="mb-1.5 text-micro uppercase tracking-wider text-ink-faint">Per node</p>
        <div className="space-y-1">
          {priced.map((p) => (
            <CostRow key={p.nodeId} entry={p} />
          ))}
          {unpriced.map((p) => (
            <CostRow key={p.nodeId} entry={p} />
          ))}
        </div>
      </div>
    </div>
  );
}

function CostRow({ entry }: { entry: NodeCost }) {
  const setSelectedNodeId = useEditorStore((s) => s.setSelectedNodeId);
  const priced = !!entry.estimate;

  return (
    <button
      type="button"
      onClick={() => setSelectedNodeId(entry.nodeId)}
      className="group flex w-full items-center justify-between gap-2 rounded-clyra px-1.5 py-1 text-left hover:bg-clyra-overlay/40"
      title={entry.estimate ? `${entry.estimate.basis}` : 'No pricing data — treat as $0'}
    >
      <div className="min-w-0 flex-1">
        <p className="truncate text-petit text-ink-primary">{entry.nodeLabel}</p>
        <p className="truncate text-micro text-ink-faint">
          {entry.nodeType}
        </p>
      </div>
      <span className={`shrink-0 text-petit ${priced ? 'text-ink-secondary' : 'text-ink-faint'}`}>
        {priced ? `$${entry.estimate!.usd}` : '—'}
      </span>
    </button>
  );
}

function ReviewView({ findings }: { findings: Finding[] }) {
  if (findings.length === 0) {
    return (
      <div className="p-4">
        <div className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded-full bg-success/15 text-success">
            <svg className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
              <path d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" />
            </svg>
          </div>
          <div>
            <p className="text-petit font-medium text-ink-primary">No issues found</p>
            <p className="text-micro text-ink-muted">Design passes the built-in checks.</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="divide-y divide-line-subtle">
      {findings.map((f) => (
        <FindingRow key={f.id} finding={f} />
      ))}
    </div>
  );
}

const SEVERITY_STYLES: Record<Severity, { badge: string; label: string }> = {
  high: { badge: 'border-danger/30 bg-danger/10 text-danger', label: 'High' },
  medium: { badge: 'border-amber-500/30 bg-amber-500/10 text-amber-300', label: 'Medium' },
  low: { badge: 'border-ink-muted/30 bg-ink-muted/10 text-ink-muted', label: 'Low' },
  info: { badge: 'border-brand-500/30 bg-brand-500/10 text-brand-300', label: 'Info' },
};

function FindingRow({ finding }: { finding: Finding }) {
  const setSelectedNodeId = useEditorStore((s) => s.setSelectedNodeId);
  const style = SEVERITY_STYLES[finding.severity];

  const handleFocus = () => {
    if (finding.nodeIds[0]) setSelectedNodeId(finding.nodeIds[0]);
  };

  return (
    <div className="p-3">
      <div className="flex items-start justify-between gap-2">
        <p className="text-petit font-medium text-ink-primary">{finding.title}</p>
        <span className={`shrink-0 rounded-full border px-1.5 py-0.5 text-micro ${style.badge}`}>
          {style.label}
        </span>
      </div>
      <p className="mt-1 text-micro text-ink-secondary">{finding.detail}</p>

      <div className="mt-2 flex flex-wrap items-center gap-2 text-micro">
        <span className="capitalize text-ink-faint">{finding.pillar}</span>
        <span className="text-ink-faint">·</span>
        <span className="font-mono text-ink-faint">{finding.id}</span>
        {finding.nodeIds.length > 0 && (
          <>
            <span className="text-ink-faint">·</span>
            <button
              type="button"
              onClick={handleFocus}
              className="text-brand-400 hover:text-brand-300"
            >
              Focus {finding.nodeIds.length === 1 ? 'node' : `${finding.nodeIds.length} nodes`}
            </button>
          </>
        )}
        {finding.doc && (
          <>
            <span className="text-ink-faint">·</span>
            <a
              href={finding.doc.url}
              target="_blank"
              rel="noreferrer"
              className="text-brand-400 hover:text-brand-300"
            >
              {finding.doc.label} ↗
            </a>
          </>
        )}
      </div>
    </div>
  );
}
