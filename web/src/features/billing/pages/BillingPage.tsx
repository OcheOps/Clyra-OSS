import { useEffect, useMemo, useState } from 'react';
import toast from 'react-hot-toast';
import { AppLayout } from '@/components/layout/AppLayout';
import { PageHeader } from '@/components/layout/PageHeader';
import { useAuthStore } from '@/features/auth/store';
import { getPricing } from '@/services/api/pricing';
import { initializePayment, cancelPlan } from '@/services/api/payments';
import { detectClientCountry } from '@/services/geo';
import { PaymentHistory } from '@/features/billing/components/PaymentHistory';
import { ConfirmDialog } from '@/features/billing/components/ConfirmDialog';
import type { PricingCatalog, PricingPlan, PricingPlanCard } from '@/types/pricing';

/**
 * BillingPage — v2 restyle.
 *
 * Three sections: current-plan summary card, plan-picker grid (mirrors
 * the landing pricing visual language), and a payment history table.
 * Cancel flow lives in a ConfirmDialog rendered at the page root.
 *
 * Plan-switching logic unchanged per round-3 guardrails — recommended
 * plan glows, current plan shows an emerald 'CURRENT' chip, in-flight
 * checkout shows 'Redirecting…' on the chosen card, locked plans
 * (switching while paid) get a disabled CTA explaining how to switch.
 */
export function BillingPage() {
  const user = useAuthStore((s) => s.user);
  const setUser = useAuthStore((s) => s.setUser);
  // Display the effective plan the backend actually enforces, not the raw
  // DB value — otherwise an expired Pro shows "Pro" while every Pro API
  // 402s ("shows Pro but blocks Pro features" bug). Raw plan + expiry are
  // still available for "your Pro lapsed on Aug 5" style copy.
  const currentPlan = user?.effective_plan ?? user?.plan ?? 'free';
  const expiry = useMemo(
    () => (user?.plan_expiry ? new Date(user.plan_expiry) : null),
    [user?.plan_expiry],
  );
  const hasActivePaid = currentPlan !== 'free';

  // Refund is offered only within 24h of activation. We derive activation date
  // from expiry - 30d — this matches the backend's PeriodDays default.
  const refundEligible = useMemo(() => {
    if (!expiry || !hasActivePaid) return false;
    const activatedMs = expiry.getTime() - 30 * 24 * 60 * 60 * 1000;
    return Date.now() - activatedMs <= 24 * 60 * 60 * 1000;
  }, [expiry, hasActivePaid]);

  const [catalog, setCatalog] = useState<PricingCatalog | null>(null);
  const [loadingPlan, setLoadingPlan] = useState<PricingPlan | null>(null);
  const [fetchError, setFetchError] = useState(false);
  const [cancelOpen, setCancelOpen] = useState(false);
  const [canceling, setCanceling] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    let cancelled = false;
    (async () => {
      const country = await detectClientCountry(controller.signal);
      try {
        const c = await getPricing(country || undefined);
        if (!cancelled) setCatalog(c);
      } catch {
        if (!cancelled) setFetchError(true);
      }
    })();
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, []);

  const handleCheckout = async (plan: Exclude<PricingPlan, 'free'>) => {
    if (!catalog) return;
    setLoadingPlan(plan);
    try {
      const res = await initializePayment({ plan, region: catalog.region || undefined });
      window.location.href = res.authorization_url;
    } catch (err) {
      toast.error((err as Error).message || 'Could not start checkout');
      setLoadingPlan(null);
    }
  };

  const handleCancel = async () => {
    setCanceling(true);
    try {
      const result = await cancelPlan();
      if (user) setUser({ ...user, plan: 'free', plan_expiry: null });
      if (result.refunded) {
        toast.success('Plan canceled and refund queued. Funds return to your card in a few days.');
      } else {
        toast.success('Plan canceled. You can resubscribe anytime.');
      }
      setCancelOpen(false);
    } catch (err) {
      toast.error((err as Error).message || 'Cancel failed');
    } finally {
      setCanceling(false);
    }
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

        <div className="relative mx-auto max-w-5xl px-6 py-12">
          <PageHeader
            eyebrow="Subscription"
            title="Billing"
            description="Pick a plan that fits your workflow. Cancel anytime."
            backTo="/dashboard"
          />

          <CurrentPlanCard
            plan={currentPlan}
            hasActivePaid={hasActivePaid}
            expiry={expiry}
            onCancel={() => setCancelOpen(true)}
          />

          {fetchError && (
            <p className="mt-8 rounded-2xl border border-c2-danger/30 bg-c2-danger/10 p-4 text-petit text-c2-danger">
              Couldn't load pricing. Refresh and try again.
            </p>
          )}

          {catalog && (
            <section className="mt-10">
              <div className="flex items-center gap-2">
                <span className="inline-block h-1 w-1 rounded-full bg-c2-primary" />
                <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
                  Plans
                </span>
              </div>
              <h2 className="mt-3 font-display text-2xl font-semibold tracking-[-0.02em] text-c2-text">
                {catalog.region === 'NG' ? 'Priced in naira.' : 'Priced in USD.'}
              </h2>
              {hasActivePaid && (
                <p className="mt-2 text-petit text-c2-text-muted">
                  Cancel your current plan to switch or resubscribe.
                </p>
              )}

              <div className="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
                {catalog.plans.map((plan) => (
                  <BillingPlanCard
                    key={plan.plan}
                    plan={plan}
                    current={
                      plan.plan === currentPlan &&
                      (plan.plan === 'free' || hasActivePaid)
                    }
                    locked={
                      hasActivePaid && plan.plan !== currentPlan && plan.plan !== 'free'
                    }
                    loading={loadingPlan === plan.plan}
                    onSelect={() => {
                      if (
                        plan.plan === 'free' ||
                        plan.plan === currentPlan ||
                        hasActivePaid
                      )
                        return;
                      handleCheckout(plan.plan);
                    }}
                  />
                ))}
              </div>

              <p className="mt-6 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
                {catalog.region === 'NG'
                  ? 'Payments handled by Paystack'
                  : 'Payments by Polar (merchant of record) · VAT included where applicable'}{' '}
                · By subscribing you agree to the{' '}
                <a
                  href="/terms"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-c2-primary underline decoration-c2-primary/30 underline-offset-2 transition-colors hover:text-c2-info"
                >
                  Terms
                </a>{' '}
                and{' '}
                <a
                  href="/privacy"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-c2-primary underline decoration-c2-primary/30 underline-offset-2 transition-colors hover:text-c2-info"
                >
                  Privacy Policy
                </a>
              </p>
            </section>
          )}

          <section className="mt-12">
            <div className="flex items-center gap-2">
              <span className="inline-block h-1 w-1 rounded-full bg-c2-primary" />
              <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
                History
              </span>
            </div>
            <h2 className="mt-3 font-display text-xl font-semibold tracking-[-0.02em] text-c2-text">
              Billing history
            </h2>
            <div className="mt-5 rounded-2xl border border-c2-border bg-c2-surface/60 p-6">
              <PaymentHistory />
            </div>
          </section>
        </div>
      </main>

      <ConfirmDialog
        open={cancelOpen}
        title="Cancel your plan?"
        body={
          refundEligible
            ? "You're within the 24-hour grace window, so we'll queue a full refund to your card after downgrading to Free. Settlement takes a few days on your bank's side."
            : "Your plan will drop back to Free immediately. Past the 24-hour window we don't refund the partial month, so you'll lose any remaining paid days. You can resubscribe anytime."
        }
        confirmLabel={refundEligible ? 'Cancel & refund' : 'Cancel plan'}
        confirmVariant="danger"
        loading={canceling}
        onConfirm={handleCancel}
        onClose={() => setCancelOpen(false)}
      />
    </AppLayout>
  );
}

// ────────────────────────────────────────────────────────────────────
// Current plan summary panel — shows what the user is on right now,
// renewal date, and the cancel chip when applicable.

function CurrentPlanCard({
  plan,
  hasActivePaid,
  expiry,
  onCancel,
}: {
  plan: string;
  hasActivePaid: boolean;
  expiry: Date | null;
  onCancel: () => void;
}) {
  return (
    <section className="mt-8 overflow-hidden rounded-2xl border border-c2-border bg-c2-surface/60 p-6">
      <div className="flex items-center gap-2">
        <span className="inline-block h-1 w-1 rounded-full bg-c2-primary" />
        <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
          Current plan
        </span>
      </div>
      <div className="mt-3 flex flex-wrap items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-3">
            <span className="font-display text-3xl font-semibold capitalize tracking-[-0.02em] text-c2-text">
              {plan}
            </span>
            {hasActivePaid && (
              <span className="inline-flex items-center gap-1.5 rounded-full border border-c2-accent/30 bg-c2-accent-soft px-2.5 py-0.5 font-mono text-[9px] font-semibold uppercase tracking-[0.18em] text-c2-accent">
                <span className="h-1 w-1 rounded-full bg-c2-accent shadow-[0_0_6px_rgba(16,185,129,0.7)]" />
                Active
              </span>
            )}
          </div>
          {expiry && hasActivePaid && (
            <p className="mt-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
              Renews · {expiry.toLocaleDateString()}
            </p>
          )}
        </div>
        {hasActivePaid && (
          <button
            type="button"
            onClick={onCancel}
            className="rounded-full border border-c2-danger/40 bg-c2-danger/10 px-4 py-2 text-tiny font-medium text-c2-danger transition-colors hover:border-c2-danger/60 hover:bg-c2-danger/20"
          >
            Cancel plan
          </button>
        )}
      </div>
    </section>
  );
}

// ────────────────────────────────────────────────────────────────────
// BillingPlanCard — mirrors the landing PricingSection plan card with
// extra states: current / locked / loading.

function BillingPlanCard({
  plan,
  current,
  locked,
  loading,
  onSelect,
}: {
  plan: PricingPlanCard;
  current: boolean;
  locked: boolean;
  loading: boolean;
  onSelect: () => void;
}) {
  const highlight = plan.recommended;
  const isFree = plan.plan === 'free';

  return (
    <div
      className={`relative flex min-w-0 flex-col rounded-2xl border p-5 transition-all duration-300 ${
        highlight
          ? 'border-c2-primary/50 bg-c2-surface shadow-c2-glow'
          : 'border-c2-border bg-c2-surface/60 hover:border-c2-border-strong hover:bg-c2-surface'
      } ${locked ? 'opacity-50' : ''}`}
    >
      {highlight && (
        <>
          <div
            aria-hidden
            className="pointer-events-none absolute -inset-px -z-10 rounded-2xl"
            style={{
              background:
                'radial-gradient(circle at 50% 0%, rgba(6,182,212,0.18) 0%, transparent 50%)',
            }}
          />
          <span className="absolute -top-2.5 left-6 inline-flex items-center gap-1.5 rounded-full bg-c2-primary px-2.5 py-0.5 font-mono text-[9px] font-semibold uppercase tracking-[0.18em] text-c2-bg">
            <span className="inline-block h-1 w-1 rounded-full bg-c2-bg" />
            Recommended
          </span>
        </>
      )}

      <div className="flex items-baseline justify-between">
        <h3 className="font-display text-base font-semibold tracking-[-0.005em] text-c2-text">
          {plan.name}
        </h3>
        {current && (
          <span className="inline-flex items-center gap-1 rounded-full border border-c2-accent/30 bg-c2-accent-soft px-2 py-0.5 font-mono text-[9px] font-semibold uppercase tracking-[0.18em] text-c2-accent">
            <span className="h-1 w-1 rounded-full bg-c2-accent" />
            Current
          </span>
        )}
      </div>
      <p className="mt-1 text-petit text-c2-text-muted">{plan.tagline}</p>

      <div className="mt-6 flex items-end gap-1.5">
        <span className="font-display text-4xl font-semibold tracking-[-0.025em] text-c2-text">
          {plan.amount_major}
        </span>
        {plan.amount_minor > 0 && (
          <span className="pb-1.5 text-tiny text-c2-text-faint">/ {plan.period}</span>
        )}
      </div>

      <p className="mt-3 truncate font-mono text-[10px] uppercase tracking-[0.10em] text-c2-text-muted">
        {plan.daily_limit.toLocaleString()} AI generations · per account
      </p>

      <ul className="mt-5 space-y-2 text-[13px] leading-relaxed text-c2-text-muted">
        {plan.features.map((f) => (
          <li key={f} className="flex items-start gap-2.5">
            <span
              className={`mt-[7px] h-1 w-1 shrink-0 rounded-full ${
                highlight ? 'bg-c2-primary' : 'bg-c2-accent'
              }`}
            />
            <span>{f}</span>
          </li>
        ))}
      </ul>

      <div className="mt-6">
        {current ? (
          <DisabledCTA>You're on this plan</DisabledCTA>
        ) : isFree ? (
          <DisabledCTA>Included by default</DisabledCTA>
        ) : locked ? (
          <DisabledCTA>Cancel current plan to switch</DisabledCTA>
        ) : (
          <button
            type="button"
            onClick={onSelect}
            disabled={loading}
            className={`group block w-full rounded-full px-4 py-2.5 text-center text-tiny font-medium transition-all disabled:opacity-60 ${
              highlight
                ? 'bg-c2-primary text-c2-bg shadow-c2-glow hover:bg-c2-primary-hover hover:shadow-[0_0_32px_-6px_rgba(6,182,212,0.55)]'
                : 'border border-c2-border bg-c2-bg/40 text-c2-text hover:border-c2-border-strong hover:bg-c2-surface'
            }`}
          >
            <span className="inline-flex items-center gap-2">
              {loading ? 'Redirecting…' : plan.cta}
              {!loading && (
                <span className="transition-transform group-hover:translate-x-0.5">→</span>
              )}
            </span>
          </button>
        )}
        {plan.requires_keys && (
          <p className="mt-2.5 font-mono text-[9px] uppercase tracking-[0.18em] text-c2-text-faint">
            Uses your own openai / anthropic / gemini key
          </p>
        )}
      </div>
    </div>
  );
}

function DisabledCTA({ children }: { children: React.ReactNode }) {
  return (
    <div className="block w-full cursor-not-allowed rounded-full border border-c2-border bg-c2-bg/30 px-4 py-2.5 text-center text-tiny font-medium text-c2-text-faint">
      {children}
    </div>
  );
}
