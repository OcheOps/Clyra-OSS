import { useCallback, useEffect, useState, Fragment } from 'react';
import toast from 'react-hot-toast';
import { listAuditEvents, type AuditEvent } from '@/services/api/admin';

const PAGE_SIZE = 50;

const KNOWN_EVENTS = [
  'auth.login_success',
  'auth.password_reset_requested',
  'auth.password_reset_completed',
  'billing.plan_upgraded',
  'billing.plan_cancelled',
  'billing.plan_refunded',
  'admin.role_changed',
  'share.created',
  'share.revoked',
  'iac.exported',
  'webhook.signature_failed',
] as const;

function timeAgo(ts: string): string {
  const diff = Math.floor((Date.now() - new Date(ts).getTime()) / 1000);
  if (diff < 60) return 'just now';
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

/**
 * AuditPanel — v2 restyle.
 *
 * Security event log with filtering + pagination. Each row expands to
 * show the raw metadata JSON so an investigator can see every field the
 * backend emitted without us having to schema-match here.
 *
 * Filter bar uses the same input treatment as the rest of the v2 admin
 * surface (c2-bg input + cyan focus ring). Table header is mono small-
 * caps; cells use mono for IDs and ink-readable text for the event name.
 * Expanded metadata renders as a code block on the dark substrate.
 */
export function AuditPanel() {
  const [items, setItems] = useState<AuditEvent[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [event, setEvent] = useState('');
  const [actor, setActor] = useState('');
  const [offset, setOffset] = useState(0);
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await listAuditEvents({
        event: event || undefined,
        actor: actor.trim() || undefined,
        limit: PAGE_SIZE,
        offset,
      });
      setItems(data.items);
      setTotal(data.total);
    } catch {
      toast.error('Failed to load audit log');
    } finally {
      setLoading(false);
    }
  }, [event, actor, offset]);

  useEffect(() => {
    load();
  }, [load]);

  const page = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div>
      {/* Filter bar */}
      <div className="flex flex-wrap items-end gap-3 rounded-2xl border border-c2-border bg-c2-surface/60 p-4">
        <div className="min-w-[200px] flex-1 space-y-2">
          <label className="block font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            Event
          </label>
          <select
            value={event}
            onChange={(e) => {
              setEvent(e.target.value);
              setOffset(0);
            }}
            className="w-full rounded-lg border border-c2-border bg-c2-bg/40 px-3 py-2 text-sm text-c2-text transition-colors focus:border-c2-primary/60 focus:bg-c2-surface focus:outline-none focus:ring-2 focus:ring-c2-primary-ring/40"
          >
            <option value="">All events</option>
            {KNOWN_EVENTS.map((e) => (
              <option key={e} value={e}>
                {e}
              </option>
            ))}
          </select>
        </div>
        <div className="min-w-[200px] flex-1 space-y-2">
          <label className="block font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            Actor user id
          </label>
          <input
            value={actor}
            onChange={(e) => setActor(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                setOffset(0);
                load();
              }
            }}
            placeholder="usr_…"
            className="w-full rounded-lg border border-c2-border bg-c2-bg/40 px-3 py-2 font-mono text-sm text-c2-text placeholder:text-c2-text-faint transition-colors focus:border-c2-primary/60 focus:bg-c2-surface focus:outline-none focus:ring-2 focus:ring-c2-primary-ring/40"
          />
        </div>
        <button
          type="button"
          onClick={() => {
            setOffset(0);
            load();
          }}
          className="rounded-full bg-c2-primary-soft px-3 py-2 text-tiny font-medium text-c2-primary transition-colors hover:bg-c2-primary/25"
        >
          Refresh
        </button>
      </div>

      {/* Table */}
      <div className="mt-4 overflow-hidden rounded-2xl border border-c2-border">
        {loading ? (
          <div className="px-4 py-12 text-center font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            Loading…
          </div>
        ) : items.length === 0 ? (
          <div className="px-4 py-12 text-center font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            No events match.
          </div>
        ) : (
          <table className="w-full text-tiny">
            <thead>
              <tr className="border-b border-c2-hairline bg-c2-bg/40 text-left">
                <Th>When</Th>
                <Th>Event</Th>
                <Th>Actor</Th>
                <Th>Target</Th>
                <Th>IP</Th>
                <th className="px-3 py-2" />
              </tr>
            </thead>
            <tbody>
              {items.map((it) => {
                const isOpen = !!expanded[it.id];
                return (
                  <Fragment key={it.id}>
                    <tr className="border-b border-c2-hairline transition-colors hover:bg-c2-bg/30">
                      <td
                        className="whitespace-nowrap px-3 py-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint"
                        title={new Date(it.occurred_at).toLocaleString()}
                      >
                        {timeAgo(it.occurred_at)}
                      </td>
                      <td className="px-3 py-2 font-mono text-c2-text">{it.event}</td>
                      <td className="px-3 py-2 font-mono text-c2-text-muted">{it.actor_user_id ?? '—'}</td>
                      <td className="px-3 py-2 font-mono text-c2-text-muted">{it.target_id ?? '—'}</td>
                      <td className="px-3 py-2 font-mono text-c2-text-faint">{it.ip ?? '—'}</td>
                      <td className="px-3 py-2 text-right">
                        <button
                          type="button"
                          onClick={() =>
                            setExpanded((m) => ({ ...m, [it.id]: !isOpen }))
                          }
                          className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-primary transition-colors hover:text-c2-info"
                        >
                          {isOpen ? 'Hide' : 'Details'}
                        </button>
                      </td>
                    </tr>
                    {isOpen && (
                      <tr>
                        <td colSpan={6} className="bg-c2-bg/30 px-3 pb-4">
                          <pre className="overflow-x-auto rounded-lg border border-c2-hairline bg-c2-bg p-3 font-mono text-[11px] leading-relaxed text-c2-text-muted">
                            {JSON.stringify(it.metadata, null, 2)}
                          </pre>
                          {it.user_agent && (
                            <p
                              className="mt-2 truncate font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint"
                              title={it.user_agent}
                            >
                              UA · {it.user_agent}
                            </p>
                          )}
                        </td>
                      </tr>
                    )}
                  </Fragment>
                );
              })}
            </tbody>
          </table>
        )}
      </div>

      {/* Pagination */}
      <div className="mt-4 flex items-center justify-between font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
        <span>
          {total} event{total === 1 ? '' : 's'}
        </span>
        <div className="flex items-center gap-2">
          <button
            type="button"
            disabled={offset === 0}
            onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
            className="rounded px-2 py-1 transition-colors hover:text-c2-text disabled:cursor-not-allowed disabled:opacity-40"
          >
            Prev
          </button>
          <span>
            Page {page} of {totalPages}
          </span>
          <button
            type="button"
            disabled={page >= totalPages}
            onClick={() => setOffset(offset + PAGE_SIZE)}
            className="rounded px-2 py-1 transition-colors hover:text-c2-text disabled:cursor-not-allowed disabled:opacity-40"
          >
            Next
          </button>
        </div>
      </div>
    </div>
  );
}

function Th({ children }: { children: React.ReactNode }) {
  return (
    <th className="px-3 py-2.5 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
      {children}
    </th>
  );
}
