import { useCallback, useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { listFeedback, setFeedbackStatus, type FeedbackItem } from '@/services/api/admin';

const PAGE_SIZE = 50;
const STATUSES = ['', 'new', 'open', 'closed'] as const;

function timeAgo(ts: string): string {
  const diff = Math.floor((Date.now() - new Date(ts).getTime()) / 1000);
  if (diff < 60) return 'just now';
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

const STATUS_BADGE: Record<string, string> = {
  new:    'border-c2-primary/30 bg-c2-primary-soft text-c2-primary',
  open:   'border-amber-500/30 bg-amber-500/15 text-amber-300',
  closed: 'border-c2-border bg-c2-bg/40 text-c2-text-muted',
};

const CATEGORY_BADGE: Record<string, string> = {
  bug:      'border-c2-danger/30 bg-c2-danger/10 text-c2-danger',
  feature:  'border-violet-500/30 bg-violet-500/15 text-violet-300',
  billing:  'border-c2-accent/30 bg-c2-accent-soft text-c2-accent',
  security: 'border-amber-500/30 bg-amber-500/15 text-amber-300',
  general:  'border-c2-border bg-c2-bg/40 text-c2-text-muted',
  other:    'border-c2-border bg-c2-bg/40 text-c2-text-muted',
};

/**
 * FeedbackPanel — v2 restyle.
 *
 * Inbox for the public support form submissions. Filter pills at the
 * top (All / New / Open / Closed), then rows showing status + category
 * badges + sender info + a truncated message that expands on click.
 * Triage actions ('Open' / 'Close') sit on the right; optimistic local
 * mutation drops the row when the current filter excludes the new
 * status.
 */
export function FeedbackPanel() {
  const [items, setItems] = useState<FeedbackItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [status, setStatus] = useState<string>('new');
  const [offset, setOffset] = useState(0);
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await listFeedback({
        status: status || undefined,
        limit: PAGE_SIZE,
        offset,
      });
      setItems(data.items);
      setTotal(data.total);
    } catch {
      toast.error('Failed to load feedback');
    } finally {
      setLoading(false);
    }
  }, [status, offset]);

  useEffect(() => {
    load();
  }, [load]);

  const handleSetStatus = async (id: string, next: 'new' | 'open' | 'closed') => {
    try {
      await setFeedbackStatus(id, next);
      // Optimistic local mutation; drop the row if the new status leaves
      // the current filter.
      setItems((prev) => {
        if (status && status !== next) return prev.filter((i) => i.id !== id);
        return prev.map((i) => (i.id === id ? { ...i, status: next } : i));
      });
      toast.success(`Marked ${next}`);
    } catch {
      toast.error('Update failed');
    }
  };

  const page = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div>
      {/* Filter pills */}
      <div className="inline-flex gap-1 rounded-full border border-c2-border bg-c2-surface/40 p-1">
        {STATUSES.map((s) => (
          <button
            key={s || 'all'}
            type="button"
            onClick={() => {
              setStatus(s);
              setOffset(0);
            }}
            className={`rounded-full px-3 py-1.5 text-tiny font-medium transition-colors ${
              status === s
                ? 'bg-c2-primary text-c2-bg'
                : 'text-c2-text-muted hover:text-c2-text'
            }`}
          >
            {s ? s.charAt(0).toUpperCase() + s.slice(1) : 'All'}
          </button>
        ))}
      </div>

      {/* List */}
      <div className="mt-4 space-y-2">
        {loading ? (
          <div className="rounded-2xl border border-c2-border bg-c2-surface/40 px-4 py-12 text-center font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            Loading…
          </div>
        ) : items.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-c2-border bg-c2-surface/40 px-4 py-12 text-center">
            <p className="font-display text-base font-semibold text-c2-text">
              No feedback in this view.
            </p>
            <p className="mt-1.5 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
              Switch filter or wait
            </p>
          </div>
        ) : (
          items.map((it) => {
            const isOpen = !!expanded[it.id];
            return (
              <div
                key={it.id}
                className="overflow-hidden rounded-2xl border border-c2-border bg-c2-surface/60"
              >
                <div className="flex items-start justify-between gap-3 p-4">
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <Badge style={STATUS_BADGE[it.status] ?? STATUS_BADGE.new!}>
                        {it.status}
                      </Badge>
                      <Badge
                        style={CATEGORY_BADGE[it.category] ?? CATEGORY_BADGE.general!}
                      >
                        {it.category}
                      </Badge>
                      <span className="truncate text-tiny text-c2-text-muted">{it.name}</span>
                      <span className="text-tiny text-c2-text-faint">·</span>
                      <a
                        href={`mailto:${it.email}`}
                        className="truncate font-mono text-[11px] text-c2-primary transition-colors hover:text-c2-info"
                      >
                        {it.email}
                      </a>
                      <span
                        className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint"
                        title={new Date(it.created_at).toLocaleString()}
                      >
                        {timeAgo(it.created_at)}
                      </span>
                    </div>
                    <p
                      className={`mt-3 whitespace-pre-wrap text-tiny leading-relaxed text-c2-text-muted ${
                        isOpen ? '' : 'line-clamp-2'
                      }`}
                    >
                      {it.message}
                    </p>
                  </div>
                  <div className="flex shrink-0 flex-col items-end gap-1.5">
                    <button
                      type="button"
                      onClick={() =>
                        setExpanded((m) => ({ ...m, [it.id]: !isOpen }))
                      }
                      className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint transition-colors hover:text-c2-text"
                    >
                      {isOpen ? 'Collapse' : 'Expand'}
                    </button>
                    {it.status !== 'open' && (
                      <button
                        onClick={() => handleSetStatus(it.id, 'open')}
                        className="font-mono text-[10px] uppercase tracking-[0.18em] text-amber-300 transition-colors hover:text-amber-200"
                      >
                        Open
                      </button>
                    )}
                    {it.status !== 'closed' && (
                      <button
                        onClick={() => handleSetStatus(it.id, 'closed')}
                        className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint transition-colors hover:text-c2-text-muted"
                      >
                        Close
                      </button>
                    )}
                  </div>
                </div>
                {isOpen && it.user_id && (
                  <div className="border-t border-c2-hairline bg-c2-bg/40 px-4 py-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
                    user · <span className="text-c2-text-muted">{it.user_id}</span>
                    {it.ip && (
                      <>
                        {' '}
                        · ip · <span className="text-c2-text-muted">{it.ip}</span>
                      </>
                    )}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>

      <div className="mt-4 flex items-center justify-between font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
        <span>
          {total} item{total === 1 ? '' : 's'}
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

function Badge({ children, style }: { children: React.ReactNode; style: string }) {
  return (
    <span
      className={`rounded-full border px-2 py-0.5 font-mono text-[9px] font-semibold uppercase tracking-[0.18em] ${style}`}
    >
      {children}
    </span>
  );
}
