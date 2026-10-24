import { Link } from 'react-router-dom';
import { AppLayout } from '@/components/layout/AppLayout';
import { PageHeader } from '@/components/layout/PageHeader';
import { useAuthStore } from '@/features/auth/store';
import { useLogout } from '@/features/auth/hooks/useAuth';
import { BYOKeysSection } from '@/features/settings/components/BYOKeysSection';
import { BYOLockedCard } from '@/features/settings/components/BYOLockedCard';
import { MCPTokensSection } from '@/features/settings/components/MCPTokensSection';
import { ThemeToggle } from '@/components/ui/ThemeToggle';

/**
 * SettingsPage — v2 restyle.
 *
 * Section-per-card layout with mono small-caps eyebrows so the page
 * reads as discrete groups instead of one continuous list. Each card
 * uses the same surface treatment as the rest of the v2 signed-in
 * surfaces — bordered c2-surface with hover-strong border on
 * interactive ones, flat on read-only meta.
 *
 * Functionality is unchanged per the round-3 guardrails: pricing
 * numbers locked, account flow locked, BYO key storage logic locked.
 * Visual only.
 */
export function SettingsPage() {
  const user = useAuthStore((s) => s.user);
  const { mutate: handleLogout, isPending } = useLogout();

  // effective_plan already respects expiry (server-side EffectivePlan
  // returns free for a past plan_expiry), so the settings surface picks
  // the same bucket the backend enforces. isPaid still cross-checks the
  // expiry explicitly for the PlanCard's 'until <date>' render.
  const currentPlan = user?.effective_plan ?? user?.plan ?? 'free';
  const expiry = user?.plan_expiry ? new Date(user.plan_expiry) : null;
  const isPaid = currentPlan !== 'free' && (!expiry || expiry.getTime() > Date.now());

  return (
    <AppLayout>
      <main className="relative flex-1 overflow-y-auto">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 opacity-60"
          style={{
            backgroundImage:
              'radial-gradient(circle at 1px 1px, rgba(148,963,184,0.05) 1px, transparent 0)',
            backgroundSize: '24px 24px',
          }}
        />

        <div className="relative mx-auto max-w-2xl px-6 py-12">
          <PageHeader
            eyebrow="Account"
            title="Settings"
            description="Manage your account, plan, and AI keys."
            backTo="/dashboard"
          />

          <div className="space-y-4">
            <AccountCard
              email={user?.email ?? ''}
              role={user?.role ?? 'user'}
            />

            <PlanCard
              plan={currentPlan}
              isPaid={isPaid}
              expiry={expiry}
            />

            {currentPlan === 'byo' && isPaid ? <BYOKeysSection /> : <BYOLockedCard />}

            {/* MCP is a Pro/BYO feature — see docs/mcp.md § Plan gate.
                Free users don't get the section (backend would 402 the
                first call anyway; hiding the surface avoids the
                'why can't I click Generate' confusion). */}
            {isPaid && <MCPTokensSection />}

            <AppearanceCard />

            <DangerZone onLogout={() => handleLogout()} loggingOut={isPending} />
          </div>
        </div>
      </main>
    </AppLayout>
  );
}

// ────────────────────────────────────────────────────────────────────
// Account — read-only identity panel.

function AccountCard({ email, role }: { email: string; role: string }) {
  return (
    <Card eyebrow="Identity" title="Account">
      <div className="grid gap-3 sm:grid-cols-2">
        <Field label="Email">
          <p className="font-mono text-[13px] text-c2-text">{email}</p>
        </Field>
        <Field label="Role">
          <p className="font-mono text-[13px] uppercase tracking-[0.18em] text-c2-text">
            {role}
          </p>
        </Field>
      </div>
    </Card>
  );
}

// ────────────────────────────────────────────────────────────────────
// Plan — current plan + status pill + upgrade/manage CTA.

function PlanCard({
  plan,
  isPaid,
  expiry,
}: {
  plan: string;
  isPaid: boolean;
  expiry: Date | null;
}) {
  return (
    <Card eyebrow="Subscription" title="Plan">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-display text-2xl font-semibold capitalize tracking-[-0.02em] text-c2-text">
              {plan}
            </span>
            {isPaid && (
              <span className="inline-flex items-center gap-1.5 rounded-full border border-c2-accent/30 bg-c2-accent-soft px-2.5 py-0.5 font-mono text-[9px] font-semibold uppercase tracking-[0.18em] text-c2-accent">
                <span className="h-1 w-1 rounded-full bg-c2-accent shadow-[0_0_6px_rgba(16,185,129,0.7)]" />
                Active
              </span>
            )}
          </div>
          {expiry && isPaid && (
            <p className="mt-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
              Renews · {expiry.toLocaleDateString()}
            </p>
          )}
          {!isPaid && (
            <p className="mt-2 text-petit text-c2-text-muted">
              Upgrade for higher daily limits and managed AI.
            </p>
          )}
        </div>
        <Link
          to="/billing"
          className={
            isPaid
              ? 'inline-flex items-center gap-2 rounded-full border border-c2-border bg-c2-surface/40 px-4 py-2 text-tiny font-medium text-c2-text transition-colors hover:border-c2-border-strong hover:bg-c2-surface'
              : 'group inline-flex items-center gap-2 rounded-full bg-c2-primary px-4 py-2 text-tiny font-medium text-c2-bg shadow-c2-glow transition-all hover:bg-c2-primary-hover hover:shadow-[0_0_32px_-6px_rgba(6,182,212,0.55)]'
          }
        >
          {isPaid ? 'Manage' : 'Upgrade'}
          {!isPaid && <span className="transition-transform group-hover:translate-x-0.5">→</span>}
        </Link>
      </div>
    </Card>
  );
}

// ────────────────────────────────────────────────────────────────────
// Appearance — theme toggle.

function AppearanceCard() {
  return (
    <Card eyebrow="Display" title="Appearance">
      <p className="text-petit text-c2-text-muted">
        Choose how Clyra looks. System matches your OS setting.
      </p>
      <div className="mt-4">
        <ThemeToggle variant="labeled" />
      </div>
    </Card>
  );
}

// ────────────────────────────────────────────────────────────────────
// Danger Zone — log-out lives here for now; account deletion will land
// in a future session with a confirm-typing modal.

function DangerZone({
  onLogout,
  loggingOut,
}: {
  onLogout: () => void;
  loggingOut: boolean;
}) {
  return (
    <section className="overflow-hidden rounded-2xl border border-c2-danger/20 bg-c2-danger/[0.04] p-6">
      <div className="flex items-center gap-2">
        <span className="inline-block h-1 w-1 rounded-full bg-c2-danger" />
        <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-danger/80">
          Danger zone
        </span>
      </div>
      <div className="mt-4 flex items-center justify-between gap-4">
        <div>
          <h3 className="font-display text-base font-semibold text-c2-text">
            Sign out of this device
          </h3>
          <p className="mt-1 text-petit text-c2-text-muted">
            You'll need to sign in again to access your projects.
          </p>
        </div>
        <button
          type="button"
          onClick={onLogout}
          disabled={loggingOut}
          className="shrink-0 rounded-full border border-c2-danger/40 bg-c2-danger/10 px-4 py-2 text-tiny font-medium text-c2-danger transition-colors hover:border-c2-danger/60 hover:bg-c2-danger/20 disabled:opacity-50"
        >
          {loggingOut ? 'Signing out…' : 'Log out'}
        </button>
      </div>
    </section>
  );
}

// ────────────────────────────────────────────────────────────────────
// Card — shared section primitive. Title slot is on the same row as the
// eyebrow chip so each card reads as a discrete labeled group.

function Card({
  eyebrow,
  title,
  children,
}: {
  eyebrow: string;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="overflow-hidden rounded-2xl border border-c2-border bg-c2-surface/60 p-6 transition-colors hover:border-c2-border-strong">
      <div className="flex items-center gap-2">
        <span className="inline-block h-1 w-1 rounded-full bg-c2-primary" />
        <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
          {eyebrow}
        </span>
      </div>
      <h2 className="mt-3 font-display text-base font-semibold tracking-[-0.005em] text-c2-text">
        {title}
      </h2>
      <div className="mt-5">{children}</div>
    </section>
  );
}

// Read-only key/value field — used by the Account card. Same voice as
// the support page Field but designed for static text rather than
// inputs.
function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <span className="block font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
        {label}
      </span>
      {children}
    </div>
  );
}
