import { Link, NavLink } from 'react-router-dom';
import { useAuthStore } from '@/features/auth/store';
import { useLogout } from '@/features/auth/hooks/useAuth';
import { useDailyUsage } from '@/features/billing/hooks/useDailyUsage';
import { useUpgradeModalStore } from '@/features/billing/upgradeModalStore';
import { ThemeToggle } from '@/components/ui/ThemeToggle';

interface TopbarProps {
  children?: React.ReactNode;
}

/**
 * Topbar — v2 restyle.
 *
 * Layout: left = brand + primary nav (Dashboard, Projects);
 * center = page-contextual children (editor prompt bar, etc.);
 * right = plan state, account controls.
 *
 * Substrate matches the rest of the v2 signed-in app (c2-surface/60 +
 * c2-hairline). Nav items render as soft chips; the usage pill is the
 * one place glow appears in this bar, and only for free users at the
 * danger threshold.
 */
export function Topbar({ children }: TopbarProps) {
  const user = useAuthStore((s) => s.user);
  const { mutate: handleLogout, isPending } = useLogout();
  const { usage } = useDailyUsage();
  const openUpgrade = useUpgradeModalStore((s) => s.openModal);

  return (
    <header className="relative shrink-0 border-b border-c2-hairline bg-c2-surface/60 px-4 py-1.5 backdrop-blur-md">
      <div className="flex items-center gap-3">
        <Link
          to="/dashboard"
          className="shrink-0 transition-opacity hover:opacity-80"
        >
          <span className="font-display text-base font-semibold tracking-[-0.015em] text-c2-text">
            clyra<span className="text-c2-primary">.</span>
          </span>
        </Link>

        {user && (
          <nav className="hidden shrink-0 items-center gap-0.5 pl-2 md:flex">
            <NavItem to="/dashboard">Dashboard</NavItem>
            <NavItem to="/projects">Projects</NavItem>
          </nav>
        )}

        <div className="flex min-w-0 flex-1 items-center gap-3 overflow-x-auto [&::-webkit-scrollbar]:hidden [scrollbar-width:none]">
          {children}
        </div>

        <div className="flex shrink-0 items-center gap-2">
          {user && (
            <>
              {usage && (
                <UsagePill
                  plan={user.effective_plan ?? user.plan}
                  used={usage.daily_used}
                  limit={usage.daily_limit}
                />
              )}
              {/* Upgrade nudge is a SEPARATE control from the usage pill.
                  Previously fused into '1/10 · UPGRADE' — usage should
                  read as neutral information, upgrade should read as an
                  optional next step. Split so each carries one meaning.
                  Only shows for free users at >=80% usage (or when we
                  don't have usage data yet, so brand-new signups still
                  see the CTA on their first visit). effective_plan
                  respects expired PRO/BYO — a lapsed subscription lands
                  the user back in the free upgrade prompt. */}
              {(user.effective_plan ?? user.plan) === 'free' &&
                (!usage || usage.daily_used >= usage.daily_limit * 0.8) && (
                <button
                  type="button"
                  onClick={() => openUpgrade({ defaultPlan: 'pro' })}
                  className="rounded-full border border-c2-primary/40 bg-c2-primary-soft px-3 py-1 text-tiny font-medium text-c2-primary transition-colors hover:bg-c2-primary/20"
                  title="Upgrade for higher daily limits"
                >
                  Upgrade
                </button>
              )}
              {user.role === 'admin' && (
                <Link
                  to="/admin"
                  className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-primary transition-colors hover:text-c2-info"
                >
                  Admin
                </Link>
              )}
              <ThemeToggle variant="compact" />
              <Link
                to="/settings"
                className="hidden font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-muted transition-colors hover:text-c2-text lg:inline-block"
                title="Settings, plan, AI keys"
              >
                Settings
              </Link>
              <span
                className="hidden max-w-[14ch] truncate font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint xl:inline-block"
                title={user.email}
              >
                {user.email}
              </span>
              <button
                type="button"
                onClick={() => handleLogout()}
                disabled={isPending}
                className="rounded-full px-2.5 py-1 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-muted transition-colors hover:bg-c2-bg/40 hover:text-c2-text disabled:opacity-50"
              >
                {isPending ? 'Out…' : 'Log out'}
              </button>
            </>
          )}
        </div>
      </div>
    </header>
  );
}

function UsagePill({
  plan,
  used,
  limit,
}: {
  plan: 'free' | 'pro' | 'byo';
  used: number;
  limit: number;
}) {
  const pct = limit > 0 ? Math.min(100, Math.round((used / limit) * 100)) : 0;
  // Neutral tone always. Previously the pill doubled as an upgrade CTA
  // and shouted 'UPGRADE' at 100% — that lives in a separate button now.
  // Free/Pro/BYO all see the same calm chip; the tone shifts to amber
  // when approaching the cap purely as a heads-up, not a marketing hit.
  const tone =
    plan === 'free' && pct >= 100
      ? 'border-c2-danger/40 bg-c2-danger/10 text-c2-danger'
      : plan === 'free' && pct >= 80
        ? 'border-amber-500/40 bg-amber-500/15 text-amber-300'
        : 'border-c2-border bg-c2-surface/60 text-c2-text-muted';
  const planLabel = plan === 'pro' ? 'Pro' : plan === 'byo' ? 'BYO' : 'Free';
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-tiny font-medium ${tone}`}
      title={`Daily generations on ${planLabel}`}
    >
      <span className="tabular-nums">
        {used}/{limit}
      </span>
      <span className="text-c2-text-faint">·</span>
      <span className="font-mono text-[10px] uppercase tracking-[0.14em]">
        {planLabel}
      </span>
    </span>
  );
}

function NavItem({ to, children }: { to: string; children: React.ReactNode }) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        `rounded-full px-2.5 py-1 text-tiny font-medium transition-colors ${
          isActive
            ? 'bg-c2-primary-soft text-c2-primary'
            : 'text-c2-text-muted hover:bg-c2-bg/40 hover:text-c2-text'
        }`
      }
      end
    >
      {children}
    </NavLink>
  );
}
