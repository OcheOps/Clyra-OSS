import { useCallback, useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import {
  type AdminUser,
  emailUser,
  grantUserPlan,
  listAdminUsers,
  setUserRole,
} from '@/services/api/admin';
import { me as fetchMe } from '@/services/api/auth';
import { useAuthStore } from '@/features/auth/store';

/**
 * UsersPanel — v2 restyle.
 *
 * Admin-side user management tab. Search-by-email at the top, then a
 * table of users with three per-row actions (Plan / Email / Promote-
 * Demote). Two modal dialogs: grant a plan with an optional expiry +
 * note, and send a one-off plain-text email from the admin address.
 *
 * Functionality unchanged — only visual tokens swapped to c2-*.
 */
export function UsersPanel() {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [total, setTotal] = useState(0);
  const [q, setQ] = useState('');
  const [loading, setLoading] = useState(true);
  const [active, setActive] = useState<
    { kind: 'grant' | 'email'; user: AdminUser } | null
  >(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await listAdminUsers({ q: q.trim() || undefined, limit: 50 });
      setUsers(res.users ?? []);
      setTotal(res.total ?? 0);
    } catch {
      toast.error('Failed to load users');
    } finally {
      setLoading(false);
    }
  }, [q]);

  useEffect(() => {
    void load();
  }, [load]);

  const onToggleRole = async (u: AdminUser) => {
    const next = u.role === 'admin' ? 'user' : 'admin';
    const word = next === 'admin' ? 'Promote' : 'Demote';
    if (!confirm(`${word} ${u.email} to ${next}?`)) return;
    try {
      await setUserRole(u.id, next);
      toast.success(`${word}d ${u.email}`);
      void load();
    } catch (err) {
      toast.error((err as Error).message || 'Role change failed');
    }
  };

  return (
    <div>
      <div className="flex items-center gap-3">
        <input
          type="search"
          placeholder="Search by email…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          className="flex-1 rounded-lg border border-c2-border bg-c2-bg/40 px-3 py-2 text-sm text-c2-text placeholder:text-c2-text-faint transition-colors focus:border-c2-primary/60 focus:bg-c2-surface focus:outline-none focus:ring-2 focus:ring-c2-primary-ring/40"
        />
        <span className="shrink-0 font-mono text-[10px] uppercase tracking-[0.18em] tabular-nums text-c2-text-faint">
          {total.toLocaleString()} {total === 1 ? 'user' : 'users'}
        </span>
      </div>

      {loading ? (
        <p className="mt-10 text-center font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
          Loading…
        </p>
      ) : users.length === 0 ? (
        <p className="mt-10 text-center font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
          No users match.
        </p>
      ) : (
        <div className="mt-4 overflow-hidden rounded-2xl border border-c2-border">
          <table className="w-full text-sm">
            <thead className="bg-c2-bg/40 text-left">
              <tr>
                <Th>Email</Th>
                <Th>Role</Th>
                <Th>Plan</Th>
                <Th>Joined</Th>
                <th className="px-4 py-2.5" />
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr
                  key={u.id}
                  className="border-t border-c2-hairline transition-colors hover:bg-c2-bg/20"
                >
                  <td className="px-4 py-2.5 font-mono text-[13px] text-c2-text">
                    {u.email}
                  </td>
                  <td className="px-4 py-2.5">
                    <span
                      className={`rounded-full border px-2 py-0.5 font-mono text-[9px] font-semibold uppercase tracking-[0.18em] ${
                        u.role === 'admin'
                          ? 'border-c2-primary/30 bg-c2-primary-soft text-c2-primary'
                          : 'border-c2-border bg-c2-bg/40 text-c2-text-muted'
                      }`}
                    >
                      {u.role}
                    </span>
                  </td>
                  <td className="px-4 py-2.5">
                    {(() => {
                      const expired =
                        u.plan !== 'free' &&
                        !!u.plan_expiry &&
                        new Date(u.plan_expiry) < new Date();
                      return (
                        <>
                          <span
                            className={`font-mono text-[10px] uppercase tracking-[0.18em] ${
                              expired ? 'text-c2-text-faint line-through' : 'text-c2-text'
                            }`}
                          >
                            {u.plan}
                          </span>
                          {expired && (
                            <span className="ml-2 rounded-full border border-c2-danger/40 bg-c2-danger/10 px-2 py-0.5 font-mono text-[9px] uppercase tracking-[0.18em] text-c2-danger">
                              expired
                            </span>
                          )}
                          {u.plan_expiry && u.plan !== 'free' && !expired && (
                            <span className="ml-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
                              until {new Date(u.plan_expiry).toLocaleDateString()}
                            </span>
                          )}
                          {expired && (
                            <span className="ml-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
                              lapsed {new Date(u.plan_expiry!).toLocaleDateString()}
                            </span>
                          )}
                        </>
                      );
                    })()}
                  </td>
                  <td className="px-4 py-2.5 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
                    {new Date(u.created_at).toLocaleDateString()}
                  </td>
                  <td className="whitespace-nowrap px-4 py-2.5 text-right">
                    <button
                      onClick={() => setActive({ kind: 'grant', user: u })}
                      className="mr-4 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-primary transition-colors hover:text-c2-info"
                    >
                      Plan
                    </button>
                    <button
                      onClick={() => setActive({ kind: 'email', user: u })}
                      className="mr-4 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-muted transition-colors hover:text-c2-text"
                    >
                      Email
                    </button>
                    <button
                      onClick={() => onToggleRole(u)}
                      className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint transition-colors hover:text-c2-text"
                    >
                      {u.role === 'admin' ? 'Demote' : 'Promote'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {active?.kind === 'grant' && (
        <GrantPlanDialog
          user={active.user}
          onClose={() => setActive(null)}
          onDone={() => {
            setActive(null);
            void load();
            // If the admin just changed their OWN plan, the auth store
            // still holds the pre-grant plan — topbar keeps showing the
            // stale badge until a hard reload. Refetch /auth/me so the
            // pill matches reality immediately.
            const self = useAuthStore.getState().user;
            if (self && self.id === active.user.id) {
              void fetchMe()
                .then((res) => useAuthStore.getState().setUser(res.user))
                .catch(() => { /* non-fatal — reload will pick it up */ });
            }
          }}
        />
      )}
      {active?.kind === 'email' && (
        <EmailUserDialog
          user={active.user}
          onClose={() => setActive(null)}
          onDone={() => setActive(null)}
        />
      )}
    </div>
  );
}

function Th({ children }: { children: React.ReactNode }) {
  return (
    <th className="px-4 py-2.5 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
      {children}
    </th>
  );
}

// ────────────────────────────────────────────────────────────────────

function GrantPlanDialog({
  user,
  onClose,
  onDone,
}: {
  user: AdminUser;
  onClose: () => void;
  onDone: () => void;
}) {
  const [plan, setPlan] = useState<'free' | 'pro' | 'byo'>(user.plan);
  const [days, setDays] = useState('30');
  const [note, setNote] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const onSubmit = async () => {
    setSubmitting(true);
    try {
      let expires_at: string | undefined;
      if (plan !== 'free') {
        const d = parseInt(days, 10);
        if (Number.isFinite(d) && d > 0) {
          expires_at = new Date(Date.now() + d * 86400_000).toISOString();
        }
      }
      await grantUserPlan(user.id, plan, { expires_at, note: note.trim() || undefined });
      toast.success(`Granted ${plan} to ${user.email}`);
      onDone();
    } catch (err) {
      toast.error((err as Error).message || 'Grant failed');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <ModalShell title={`Grant plan — ${user.email}`} onClose={onClose}>
      <div className="space-y-5">
        <Field label="Plan">
          <div className="flex rounded-full border border-c2-border bg-c2-bg/40 p-1">
            {(['free', 'pro', 'byo'] as const).map((p) => (
              <button
                key={p}
                type="button"
                onClick={() => setPlan(p)}
                className={`flex-1 rounded-full px-3 py-1.5 font-mono text-[10px] font-semibold uppercase tracking-[0.18em] transition-colors ${
                  plan === p
                    ? 'bg-c2-primary text-c2-bg'
                    : 'text-c2-text-muted hover:text-c2-text'
                }`}
              >
                {p}
              </button>
            ))}
          </div>
        </Field>
        {plan !== 'free' && (
          <Field label="Duration · days">
            <input
              type="number"
              min={1}
              value={days}
              onChange={(e) => setDays(e.target.value)}
              className={INPUT_CLASSES}
            />
          </Field>
        )}
        <Field label="Note · optional · logged">
          <input
            type="text"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder="e.g. partnership comp — 90 days"
            className={INPUT_CLASSES}
          />
        </Field>
      </div>

      <ModalActions
        onCancel={onClose}
        onConfirm={onSubmit}
        confirmLabel={plan === 'free' ? 'Revert to Free' : 'Grant plan'}
        loading={submitting}
      />
    </ModalShell>
  );
}

function EmailUserDialog({
  user,
  onClose,
  onDone,
}: {
  user: AdminUser;
  onClose: () => void;
  onDone: () => void;
}) {
  const [subject, setSubject] = useState('');
  const [body, setBody] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const onSubmit = async () => {
    if (!subject.trim() || !body.trim()) return;
    setSubmitting(true);
    try {
      await emailUser(user.id, subject.trim(), body.trim());
      toast.success(`Emailed ${user.email}`);
      onDone();
    } catch (err) {
      toast.error((err as Error).message || 'Send failed');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <ModalShell title={`Email — ${user.email}`} onClose={onClose}>
      <div className="space-y-5">
        <Field label="Subject">
          <input
            type="text"
            maxLength={160}
            value={subject}
            onChange={(e) => setSubject(e.target.value)}
            placeholder="Quick follow-up"
            className={INPUT_CLASSES}
          />
        </Field>
        <Field label="Message">
          <textarea
            rows={7}
            maxLength={8000}
            value={body}
            onChange={(e) => setBody(e.target.value)}
            placeholder="Plain text. Newlines become paragraphs."
            className={`${INPUT_CLASSES} resize-y`}
          />
          <p className="mt-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            Sent from{' '}
            <code className="rounded border border-c2-border bg-c2-bg/40 px-1 py-0.5 normal-case tracking-normal text-c2-text">
              admin@example.com
            </code>{' '}
            · HTML is stripped
          </p>
        </Field>
      </div>

      <ModalActions
        onCancel={onClose}
        onConfirm={onSubmit}
        confirmLabel="Send email"
        loading={submitting}
        disabled={!subject.trim() || !body.trim()}
      />
    </ModalShell>
  );
}

const INPUT_CLASSES =
  'w-full rounded-lg border border-c2-border bg-c2-bg/40 px-3 py-2 text-sm text-c2-text ' +
  'placeholder:text-c2-text-faint focus:border-c2-primary/60 focus:bg-c2-surface focus:outline-none ' +
  'focus:ring-2 focus:ring-c2-primary-ring/40 transition-colors';

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-2">
      <label className="block font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
        {label}
      </label>
      {children}
    </div>
  );
}

// Shared dialog skeleton — same halo'd treatment as the auth pages so
// admin modals stay on-brand.
function ModalShell({
  title,
  onClose,
  children,
}: {
  title: string;
  onClose: () => void;
  children: React.ReactNode;
}) {
  return (
    <div
      role="dialog"
      aria-modal="true"
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="relative w-full max-w-md"
        onClick={(e) => e.stopPropagation()}
      >
        <div
          aria-hidden
          className="pointer-events-none absolute -inset-8 rounded-3xl opacity-60 blur-2xl"
          style={{
            background:
              'radial-gradient(circle at 50% 30%, rgba(6,182,212,0.18) 0%, transparent 60%)',
          }}
        />
        <div className="relative rounded-2xl border border-c2-border bg-c2-surface/95 p-6 shadow-c2-lg backdrop-blur-sm">
          <h2 className="font-display text-base font-semibold tracking-[-0.005em] text-c2-text">
            {title}
          </h2>
          <div className="mt-5">{children}</div>
        </div>
      </div>
    </div>
  );
}

function ModalActions({
  onCancel,
  onConfirm,
  confirmLabel,
  loading,
  disabled,
}: {
  onCancel: () => void;
  onConfirm: () => void;
  confirmLabel: string;
  loading?: boolean;
  disabled?: boolean;
}) {
  return (
    <div className="mt-6 flex justify-end gap-2">
      <button
        type="button"
        onClick={onCancel}
        className="rounded-full border border-c2-border bg-c2-bg/40 px-4 py-2 text-tiny font-medium text-c2-text-muted transition-colors hover:border-c2-border-strong hover:text-c2-text"
      >
        Cancel
      </button>
      <button
        type="button"
        onClick={onConfirm}
        disabled={loading || disabled}
        className="rounded-full bg-c2-primary px-4 py-2 text-tiny font-medium text-c2-bg shadow-c2-glow transition-all hover:bg-c2-primary-hover hover:shadow-[0_0_32px_-6px_rgba(6,182,212,0.55)] disabled:cursor-not-allowed disabled:opacity-50"
      >
        {loading ? 'Working…' : confirmLabel}
      </button>
    </div>
  );
}
