import { useCallback, useEffect, useState } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { PageHeader } from '@/components/layout/PageHeader';
import { UsersPanel } from '@/features/admin/components/UsersPanel';
import { MetricsPanel } from '@/features/admin/components/MetricsPanel';
import { AuditPanel } from '@/features/admin/components/AuditPanel';
import { FeedbackPanel } from '@/features/admin/components/FeedbackPanel';

type AdminTab = 'invites' | 'users' | 'metrics' | 'audit' | 'feedback';
import {
  approveInviteRequest,
  createInvite,
  declineInviteRequest,
  deleteInvite,
  listInviteRequests,
  listInvites,
} from '@/services/api/auth';
import type { InviteCode, InviteRequest, Plan } from '@/types/auth';
import toast from 'react-hot-toast';

const PLAN_OPTIONS: { value: Plan; label: string; description: string }[] = [
  { value: 'free', label: 'Free', description: '10/day' },
  { value: 'pro', label: 'Pro', description: '200/day' },
  { value: 'byo', label: 'BYO', description: '1000/day' },
];

const PLAN_BADGE: Record<Plan, string> = {
  free: 'border-c2-border bg-c2-bg/40 text-c2-text-muted',
  pro:  'border-c2-primary/30 bg-c2-primary-soft text-c2-primary',
  byo:  'border-c2-accent/30 bg-c2-accent-soft text-c2-accent',
};

function timeAgo(ts: string): string {
  const diff = Math.floor((Date.now() - new Date(ts).getTime()) / 1000);
  if (diff < 60) return 'just now';
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

/**
 * AdminPage — v2 chrome restyle.
 *
 * Main page scaffold + tabs + invites view restyled to c2-* tokens.
 * Sub-panels (UsersPanel, MetricsPanel, AuditPanel, FeedbackPanel)
 * keep their legacy `ink-*` / `clyra-*` token usage for a follow-up
 * session — they're internal data tables, lower pitch visibility,
 * and a token-by-token sweep is fine without redesigning the layout.
 */
export function AdminPage() {
  const [tab, setTab] = useState<AdminTab>('invites');
  const [invites, setInvites] = useState<InviteCode[]>([]);
  const [requests, setRequests] = useState<InviteRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [email, setEmail] = useState('');
  const [plan, setPlan] = useState<Plan>('free');

  const load = useCallback(async () => {
    try {
      const [inviteData, requestData] = await Promise.all([
        listInvites(),
        listInviteRequests('pending').catch(() => [] as InviteRequest[]),
      ]);
      setInvites(inviteData);
      setRequests(requestData);
    } catch {
      toast.error('Failed to load invites');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  const handleApprove = async (req: InviteRequest, approvePlan: Plan) => {
    try {
      const result = await approveInviteRequest(req.id, approvePlan);
      setInvites((prev) => [result.invite, ...prev]);
      setRequests((prev) => prev.filter((r) => r.id !== req.id));
      toast.success(result.email_sent ? `Invite sent to ${req.email}` : 'Invite created');
    } catch {
      toast.error('Failed to approve request');
    }
  };

  const handleDecline = async (req: InviteRequest) => {
    try {
      await declineInviteRequest(req.id);
      setRequests((prev) => prev.filter((r) => r.id !== req.id));
      toast.success('Request declined');
    } catch {
      toast.error('Failed to decline request');
    }
  };

  const handleCreate = async () => {
    setCreating(true);
    try {
      const result = await createInvite(email.trim() || undefined, plan);
      setInvites((prev) => [result.invite, ...prev]);
      if (result.email_sent) {
        toast.success(`Invite sent to ${email}`);
      } else if (email.trim()) {
        toast.success('Invite created (email not configured)');
      } else {
        toast.success('Invite created');
      }
      setEmail('');
      setPlan('free');
    } catch {
      toast.error('Failed to create invite');
    } finally {
      setCreating(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteInvite(id);
      setInvites((prev) => prev.filter((i) => i.id !== id));
      toast.success('Invite deleted');
    } catch {
      toast.error('Failed to delete invite');
    }
  };

  const handleCopy = (code: string) => {
    navigator.clipboard.writeText(code);
    toast.success('Copied to clipboard');
  };

  return (
    <AppLayout>
      <main className="relative flex-1 overflow-y-auto">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 opacity-60"
          style={{
            backgroundImage:
              'radial-gradient(circle at 1px 1px, rgba(148,163,184,0.05) 1px, transparent 0)',
            backgroundSize: '24px 24px',
          }}
        />

        <div className="relative mx-auto max-w-4xl px-6 py-12">
          <PageHeader
            eyebrow="Operator"
            title="Admin"
            description="Manage invites, users, and see activation metrics."
            backTo="/dashboard"
          />

          {/* Tabs — pill row instead of underline so the active state
              reads as a chip, matching the rest of the v2 voice. */}
          <div className="mb-8 flex gap-1 rounded-full border border-c2-border bg-c2-surface/40 p-1 text-tiny font-medium">
            {(['invites', 'users', 'metrics', 'audit', 'feedback'] as const).map((t) => (
              <button
                key={t}
                type="button"
                onClick={() => setTab(t)}
                className={`flex-1 rounded-full px-3 py-1.5 capitalize transition-all ${
                  tab === t
                    ? 'bg-c2-primary text-c2-bg shadow-c2-glow-soft'
                    : 'text-c2-text-muted hover:text-c2-text'
                }`}
              >
                {t}
              </button>
            ))}
          </div>

          {tab === 'users' && <UsersPanel />}
          {tab === 'metrics' && <MetricsPanel />}
          {tab === 'audit' && <AuditPanel />}
          {tab === 'feedback' && <FeedbackPanel />}
          {tab === 'invites' && <InvitesContent
            invites={invites} requests={requests} loading={loading}
            creating={creating} email={email} plan={plan}
            setEmail={setEmail} setPlan={setPlan}
            onCreate={handleCreate} onDelete={handleDelete} onCopy={handleCopy}
            onApprove={handleApprove} onDecline={handleDecline}
          />}
        </div>
      </main>
    </AppLayout>
  );
}

// Extract existing invites UI into its own component to fit the tab layout.
// Props are passed through from AdminPage so state stays up there.
interface InvitesProps {
  invites: InviteCode[];
  requests: InviteRequest[];
  loading: boolean;
  creating: boolean;
  email: string;
  plan: Plan;
  setEmail: (v: string) => void;
  setPlan: (v: Plan) => void;
  onCreate: () => void;
  onDelete: (id: string) => void;
  onCopy: (code: string) => void;
  onApprove: (req: InviteRequest, plan: Plan) => void;
  onDecline: (req: InviteRequest) => void;
}

function InvitesContent({
  invites, requests, loading, creating, email, plan,
  setEmail, setPlan, onCreate, onDelete, onCopy, onApprove, onDecline,
}: InvitesProps) {
  return (
    <>
      {/* Pending requests inbox */}
      {requests.length > 0 && (
        <div className="mb-6">
          <div className="flex items-baseline gap-2">
            <h2 className="font-display text-base font-semibold text-c2-text">
              Pending requests
            </h2>
            <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
              ({requests.length})
            </span>
          </div>
          <div className="mt-3 space-y-2">
            {requests.map((req) => (
              <RequestRow
                key={req.id}
                req={req}
                onApprove={onApprove}
                onDecline={onDecline}
              />
            ))}
          </div>
        </div>
      )}

      {/* Create invite with optional email + plan */}
      <div className="rounded-2xl border border-c2-border bg-c2-surface/60 p-5">
        <div className="flex items-end gap-3">
          <div className="flex-1 space-y-2">
            <label className="block font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
              Recipient email · optional
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="colleague@company.com"
              className="w-full rounded-lg border border-c2-border bg-c2-bg/40 px-3 py-2 text-sm text-c2-text placeholder:text-c2-text-faint transition-colors focus:border-c2-primary/60 focus:bg-c2-surface focus:outline-none focus:ring-2 focus:ring-c2-primary-ring/40"
              onKeyDown={(e) => {
                if (e.key === 'Enter') onCreate();
              }}
            />
          </div>
          <button
            type="button"
            onClick={onCreate}
            disabled={creating}
            className="group inline-flex items-center gap-2 rounded-full bg-c2-primary px-4 py-2 text-tiny font-medium text-c2-bg shadow-c2-glow transition-all hover:bg-c2-primary-hover hover:shadow-[0_0_32px_-6px_rgba(6,182,212,0.55)] disabled:opacity-60"
          >
            {creating
              ? 'Creating…'
              : email.trim()
                ? 'Send invite'
                : 'Generate code'}
            {!creating && (
              <span className="transition-transform group-hover:translate-x-0.5">→</span>
            )}
          </button>
        </div>

        {/* Plan selector */}
        <div className="mt-4 space-y-2">
          <label className="block font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            Plan · applied on signup
          </label>
          <div className="flex rounded-full border border-c2-border bg-c2-bg/40 p-1">
            {PLAN_OPTIONS.map((p) => (
              <button
                key={p.value}
                type="button"
                onClick={() => setPlan(p.value)}
                className={`flex-1 rounded-full px-3 py-1.5 text-tiny font-medium transition-all ${
                  plan === p.value
                    ? 'bg-c2-primary text-c2-bg'
                    : 'text-c2-text-muted hover:text-c2-text'
                }`}
              >
                {p.label}
                <span
                  className={`ml-1.5 font-mono text-[9px] tracking-[0.12em] ${
                    plan === p.value ? 'text-c2-bg/70' : 'text-c2-text-faint'
                  }`}
                >
                  {p.description}
                </span>
              </button>
            ))}
          </div>
        </div>

        <p className="mt-3 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
          {email.trim()
            ? 'Email will include the invite code + a direct signup link'
            : 'A code will be generated · copy and share manually'}
        </p>
      </div>

      {/* Invite list */}
      {loading ? (
        <div className="mt-8 grid gap-2">
          {[0, 1, 2].map((i) => (
            <div
              key={i}
              className="h-[60px] animate-pulse rounded-xl border border-c2-border bg-c2-surface/40"
            />
          ))}
        </div>
      ) : invites.length === 0 ? (
        <div className="mt-6 rounded-2xl border border-dashed border-c2-border bg-c2-surface/40 p-12 text-center">
          <p className="font-display text-base font-semibold text-c2-text">
            No invites yet
          </p>
          <p className="mt-1.5 text-petit text-c2-text-muted">
            Generate one to invite a beta tester.
          </p>
        </div>
      ) : (
        <div className="mt-6 space-y-2">
          {invites.map((inv) => (
            <div
              key={inv.id}
              className={`flex items-center justify-between rounded-xl border p-3 transition-colors ${
                inv.used_by
                  ? 'border-c2-hairline bg-c2-bg/40 opacity-70'
                  : 'border-c2-border bg-c2-surface/60'
              }`}
            >
              <div className="flex items-center gap-3">
                <code
                  className={`rounded-md px-2.5 py-1 font-mono text-sm ${
                    inv.used_by
                      ? 'bg-c2-bg/40 text-c2-text-faint line-through'
                      : 'cursor-pointer bg-c2-primary-soft text-c2-primary transition-colors hover:bg-c2-primary/25'
                  }`}
                  onClick={() => !inv.used_by && onCopy(inv.code)}
                  title={inv.used_by ? 'Used' : 'Click to copy'}
                >
                  {inv.code}
                </code>
                <span
                  className={`rounded-full border px-2 py-0.5 font-mono text-[9px] font-semibold uppercase tracking-[0.18em] ${
                    PLAN_BADGE[inv.plan] ?? PLAN_BADGE.free
                  }`}
                >
                  {inv.plan}
                </span>
                <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
                  {inv.used_by ? (
                    <>Used {inv.used_at ? timeAgo(inv.used_at) : ''}</>
                  ) : (
                    <>Created {timeAgo(inv.created_at)}</>
                  )}
                </span>
              </div>
              <div className="flex items-center gap-2">
                {inv.used_by ? (
                  <span className="rounded-full border border-c2-hairline bg-c2-bg/40 px-2 py-0.5 font-mono text-[9px] uppercase tracking-[0.18em] text-c2-text-faint">
                    used
                  </span>
                ) : (
                  <>
                    <button
                      onClick={() => onCopy(inv.code)}
                      className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-muted transition-colors hover:text-c2-text"
                    >
                      Copy
                    </button>
                    <button
                      onClick={() => onDelete(inv.id)}
                      className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint transition-colors hover:text-c2-danger"
                    >
                      Delete
                    </button>
                  </>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </>
  );
}

interface RequestRowProps {
  req: InviteRequest;
  onApprove: (req: InviteRequest, plan: Plan) => void;
  onDecline: (req: InviteRequest) => void;
}

function RequestRow({ req, onApprove, onDecline }: RequestRowProps) {
  const [plan, setPlan] = useState<Plan>('free');
  return (
    <div className="rounded-xl border border-c2-border bg-c2-surface/60 p-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate font-display text-sm font-semibold text-c2-text">
            {req.email}
          </p>
          <p className="mt-1 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            {timeAgo(req.created_at)}
          </p>
          {req.note && (
            <p className="mt-3 whitespace-pre-wrap text-tiny text-c2-text-muted">
              {req.note}
            </p>
          )}
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <select
            value={plan}
            onChange={(e) => setPlan(e.target.value as Plan)}
            className="rounded-md border border-c2-border bg-c2-bg/40 px-2 py-1 text-tiny text-c2-text focus:border-c2-primary/60 focus:outline-none focus:ring-2 focus:ring-c2-primary-ring/40"
          >
            <option value="free">Free</option>
            <option value="pro">Pro</option>
            <option value="byo">BYO</option>
          </select>
          <button
            onClick={() => onApprove(req, plan)}
            className="rounded-full bg-c2-primary-soft px-3 py-1 text-tiny font-medium text-c2-primary transition-colors hover:bg-c2-primary/25"
          >
            Approve
          </button>
          <button
            onClick={() => onDecline(req)}
            className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint transition-colors hover:text-c2-danger"
          >
            Decline
          </button>
        </div>
      </div>
    </div>
  );
}
