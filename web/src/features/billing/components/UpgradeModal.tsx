import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { Button } from '@/components/ui/Button';
import { getPricing } from '@/services/api/pricing';
import { detectClientCountry } from '@/services/geo';
import { initializePayment } from '@/services/api/payments';
import type { PricingCatalog, PricingPlan, PricingPlanCard } from '@/types/pricing';

interface UpgradeModalProps {
  open: boolean;
  onClose: () => void;
  defaultPlan?: Exclude<PricingPlan, 'free'>;
  // Optional context-aware copy passed through `openModal({ headline, subline })`.
  // When set, replaces the generic "Upgrade plan" / region tagline so callers
  // can say things like "You've hit 10/10 today" right inside the modal.
  headline?: string;
  subline?: string;
}

export function UpgradeModal({ open, onClose, defaultPlan = 'pro', headline, subline }: UpgradeModalProps) {
  const [catalog, setCatalog] = useState<PricingCatalog | null>(null);
  const [selected, setSelected] = useState<Exclude<PricingPlan, 'free'>>(defaultPlan);
  const [loading, setLoading] = useState(false);
  const [fetchError, setFetchError] = useState(false);

  useEffect(() => {
    if (!open) return;
    setSelected(defaultPlan);
    setFetchError(false);
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
  }, [open, defaultPlan]);

  const handleUpgrade = async () => {
    if (!catalog) return;
    setLoading(true);
    try {
      const res = await initializePayment({
        plan: selected,
        region: catalog.region || undefined,
      });
      // Polar doesn't append checkout_id to custom success_urls — stash it
      // here so /billing/return can active-verify via the provider's API.
      if (res.external_id || res.provider) {
        sessionStorage.setItem(
          'clyra-pending-checkout',
          JSON.stringify({
            reference: res.reference,
            external_id: res.external_id ?? '',
            provider: res.provider ?? '',
          }),
        );
      }
      // Bounce to Paystack / Polar (or the mock return page in mock mode).
      window.location.href = res.authorization_url;
    } catch (err) {
      toast.error((err as Error).message || 'Could not start checkout');
      setLoading(false);
    }
  };

  if (!open) return null;

  const paidPlans = (catalog?.plans ?? []).filter(
    (p): p is PricingPlanCard & { plan: Exclude<PricingPlan, 'free'> } => p.plan !== 'free',
  );

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm px-4"
      onClick={onClose}
    >
      <div
        className="relative w-full max-w-xl rounded-clyra-lg border border-line-default bg-clyra-bg shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-start justify-between border-b border-line-subtle p-5">
          <div>
            <h2 className="text-base font-semibold text-ink-primary">
              {headline || 'Upgrade plan'}
            </h2>
            <p className="mt-1 text-petit text-ink-muted">
              {subline || (
                <>
                  {catalog?.region === 'NG' ? 'Pricing in naira.' : 'Pricing in USD.'}
                  {' '}You can cancel or downgrade anytime.
                </>
              )}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded p-1 text-ink-faint hover:bg-clyra-raised hover:text-ink-primary"
            aria-label="Close"
          >
            <svg className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
              <path d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" />
            </svg>
          </button>
        </div>

        <div className="p-5">
          {fetchError && (
            <p className="text-petit text-danger">Couldn't load pricing. Try again in a moment.</p>
          )}

          {catalog && (
            <div className="grid gap-3 sm:grid-cols-2">
              {paidPlans.map((plan) => (
                <PlanOption
                  key={plan.plan}
                  plan={plan}
                  active={selected === plan.plan}
                  onSelect={() => setSelected(plan.plan)}
                />
              ))}
            </div>
          )}
        </div>

        <div className="flex flex-col gap-3 border-t border-line-subtle bg-clyra-raised/30 p-5 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-micro text-ink-faint">
            {catalog?.region === 'NG'
              ? 'Payments handled by Paystack. No card on file.'
              : 'Payments handled by Polar (Merchant of Record). No card on file.'}
            {' '}By continuing you agree to our{' '}
            <a href="/terms" target="_blank" rel="noopener noreferrer" className="text-brand-400 hover:text-brand-300">Terms</a>
            {' '}and{' '}
            <a href="/privacy" target="_blank" rel="noopener noreferrer" className="text-brand-400 hover:text-brand-300">Privacy Policy</a>.
          </p>
          <div className="flex shrink-0 gap-2">
            <Button variant="secondary" size="sm" onClick={onClose} disabled={loading}>
              Cancel
            </Button>
            <Button size="sm" onClick={handleUpgrade} loading={loading} disabled={!catalog}>
              Continue to checkout
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

function PlanOption({
  plan,
  active,
  onSelect,
}: {
  plan: PricingPlanCard;
  active: boolean;
  onSelect: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onSelect}
      className={`flex flex-col rounded-clyra-lg border p-4 text-left transition-all ${
        active
          ? 'border-brand-500/60 bg-brand-500/[0.05] ring-1 ring-brand-500/30'
          : 'border-line-default hover:border-line-strong'
      }`}
    >
      <div className="flex items-center justify-between">
        <span className="text-sm font-semibold text-ink-primary">{plan.name}</span>
        {plan.recommended && (
          <span className="rounded-full border border-brand-500/40 bg-brand-500/10 px-1.5 py-0.5 text-micro text-brand-300">
            Recommended
          </span>
        )}
      </div>
      <p className="mt-1 text-micro text-ink-faint">{plan.tagline}</p>

      <div className="mt-3 flex items-end gap-1">
        <span className="text-xl font-semibold text-ink-primary">{plan.amount_major}</span>
        <span className="pb-1 text-micro text-ink-faint">/ {plan.period}</span>
      </div>

      <p className="mt-2 text-micro text-ink-secondary">
        {plan.daily_limit.toLocaleString()} generations/day
      </p>

      {plan.requires_keys && (
        <p className="mt-2 text-micro text-ink-faint">Uses your own AI keys.</p>
      )}
    </button>
  );
}
