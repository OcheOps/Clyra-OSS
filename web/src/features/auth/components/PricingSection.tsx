import { useEffect, useState } from 'react';
import { getPricing } from '@/services/api/pricing';
import { detectClientCountry } from '@/services/geo';
import type { PricingCatalog, PricingPlanCard } from '@/types/pricing';

// FALLBACK_CATALOG mirrors the server's USD catalog (see
// internal/pricing/catalog.go). Rendered when the live pricing fetch
// fails so the section never collapses into a sad "Pricing unavailable"
// stub on a partner pitch. Numbers stay in sync with the backend by
// convention — if those change, change these.
const FALLBACK_CATALOG: PricingCatalog = {
  region: 'GLOBAL',
  plans: [
    {
      plan: 'free',
      name: 'Free',
      tagline: 'Get started, no card',
      amount_minor: 0,
      amount_major: '$0',
      currency: 'USD',
      currency_symbol: '$',
      period: 'month',
      daily_limit: 10,
      cta: 'Start free',
      features: [
        '10 AI generations per account',
        'Solution, Network & Enterprise modes',
        '1 Terraform / CloudFormation export',
        '2,000-character prompts',
        'Local-only — projects stay on this device',
      ],
    },
    {
      plan: 'pro',
      name: 'Pro',
      tagline: 'Managed AI, no setup',
      amount_minor: 1000,
      amount_major: '$10',
      currency: 'USD',
      currency_symbol: '$',
      period: 'month',
      daily_limit: 200,
      cta: 'Upgrade to Pro',
      recommended: true,
      features: [
        '200 AI generations per account',
        'Managed Bedrock — no API keys needed',
        'Priority model routing',
        'Unlimited IaC exports',
        '8,000-character prompts',
        'Cloud sync across devices',
        'Everything in Free',
      ],
    },
    {
      plan: 'byo',
      name: 'BYO',
      tagline: 'Bring your own keys',
      amount_minor: 400,
      amount_major: '$4',
      currency: 'USD',
      currency_symbol: '$',
      period: 'month',
      daily_limit: 1000,
      cta: 'Go BYO',
      requires_keys: true,
      features: [
        '1,000 AI generations per account',
        'Use your OpenAI / Anthropic / Gemini keys',
        'Keys stored in your browser only',
        '8,000-character prompts',
        'Cloud sync across devices',
        'Everything in Free',
      ],
    },
  ],
};

/**
 * PricingSection — restyled in v2 dark slate to match the landing page.
 *
 * Content + pricing numbers are unchanged (guardrail from round-3 lock).
 * Visual treatment: dark surface cards, cyan accent on the recommended
 * plan, mono labels for currency / "RECOMMENDED" badge / scale meta.
 *
 * The recommended plan card is the only place on the whole pricing
 * section that uses a glow — keeps the eye magnetised to the right
 * choice without painting the page red.
 */
export function PricingSection() {
  // Start with the fallback catalog so the section paints immediately.
  // The live fetch then upgrades it with region-correct numbers (NGN
  // pricing for NG, etc.). If the fetch fails, the user keeps seeing
  // the fallback — no empty 'Pricing unavailable' state on a pitch.
  const [catalog, setCatalog] = useState<PricingCatalog>(FALLBACK_CATALOG);
  const [usingFallback, setUsingFallback] = useState(true);

  useEffect(() => {
    const controller = new AbortController();
    let cancelled = false;

    (async () => {
      const country = await detectClientCountry(controller.signal);
      try {
        const c = await getPricing(country || undefined);
        if (!cancelled) {
          setCatalog(c);
          setUsingFallback(false);
        }
      } catch {
        // Silent — fallback stays painted. No alarming banner.
        if (!cancelled) setUsingFallback(true);
      }
    })();

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, []);

  return (
    <section id="pricing" className="relative border-t border-c2-hairline">
      <div className="mx-auto max-w-6xl px-6 py-24">
        <div className="max-w-2xl">
          <div className="flex items-center gap-2">
            <span className="inline-block h-1.5 w-1.5 rounded-full bg-c2-primary shadow-[0_0_8px_rgba(6,182,212,0.6)]" />
            <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
              Pricing
            </span>
          </div>
          <h2 className="mt-5 font-display text-3xl font-semibold leading-[1.1] tracking-[-0.025em] text-c2-text md:text-[40px]">
            Honest plans.
            <br />
            <span className="text-c2-text-muted">Local where it matters.</span>
          </h2>
          <p className="mt-5 max-w-xl text-[15px] leading-relaxed text-c2-text-muted">
            {catalog.region === 'NG'
              ? 'Priced in naira for Nigerian customers. Bring your own AI keys if you want to save more.'
              : 'Flat monthly plans in USD. Bring your own AI keys and pay only the card fee.'}
          </p>
        </div>

        {/* Cards: 1-col mobile, 2-col tablet, 3-col only at xl (1280px+).
            At 1024-1279px the grid renders 2 cards on top + 1 below
            instead of three cramped 330px columns — the right card was
            being cropped at the viewport edge on laptop-class widths
            because letter-spacing on the mono uppercase "X AI GENERATIONS
            · PER ACCOUNT" line ate the breathing room. */}
        <div className="mt-14 grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {catalog.plans.map((p) => (
            <PlanCard key={p.plan} plan={p} />
          ))}
        </div>

        <p className="mt-10 text-center font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
          {usingFallback
            ? 'Payments by Polar (merchant of record) · cancel anytime'
            : catalog.region === 'NG'
              ? 'Payments handled by Paystack · cancel anytime'
              : 'Payments by Polar (merchant of record) · cancel anytime'}
        </p>
      </div>
    </section>
  );
}

function PlanCard({ plan }: { plan: PricingPlanCard }) {
  const highlight = plan.recommended;

  return (
    // min-w-0 lets the card actually shrink inside its grid track —
    // without it, content like the long mono-uppercase usage line
    // forces the column wider than (1fr 1fr 1fr) wants to allocate.
    //
    // overflow-hidden was removed here because the 'RECOMMENDED' badge
    // sits at -top-2.5 (sticking 10px above the card edge) and got
    // clipped. The interior content doesn't need clipping at the card
    // boundary — children stay inside via natural flow.
    <div
      className={`relative flex min-w-0 flex-col rounded-2xl border p-5 transition-all duration-300 ${
        highlight
          ? 'border-c2-primary/50 bg-c2-surface shadow-c2-glow'
          : 'border-c2-border bg-c2-surface/60 hover:border-c2-border-strong hover:bg-c2-surface'
      }`}
    >
      {highlight && (
        <>
          {/* Recommended halo — the only glowing card on the page */}
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
        <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
          {plan.currency || ''}
        </span>
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

      {/* Tracking dialed back to 0.10em (from 0.18em) so the line fits a
          330px card width on laptop displays. Still reads as a tracked-
          out mono eyebrow; just stops blowing past the column edge. */}
      <p className="mt-3 truncate font-mono text-[10px] uppercase tracking-[0.10em] text-c2-text-muted">
        {plan.daily_limit.toLocaleString()} AI generations · per account
      </p>

      <ul className="mt-6 space-y-2.5">
        {plan.features.map((f) => (
          <li key={f} className="flex items-start gap-2.5 text-[13px] leading-relaxed text-c2-text-muted">
            <span
              className={`mt-[7px] h-1 w-1 shrink-0 rounded-full ${
                highlight ? 'bg-c2-primary' : 'bg-c2-accent'
              }`}
            />
            <span>{f}</span>
          </li>
        ))}
      </ul>

      <div className="mt-7">
        <a
          href="/signup"
          className={`group block w-full rounded-full px-4 py-2.5 text-center text-tiny font-medium transition-all ${
            highlight
              ? 'bg-c2-primary text-c2-bg shadow-c2-glow hover:bg-c2-primary-hover hover:shadow-[0_0_32px_-6px_rgba(6,182,212,0.55)]'
              : 'border border-c2-border bg-c2-bg/40 text-c2-text hover:border-c2-border-strong hover:bg-c2-surface'
          }`}
        >
          <span className="inline-flex items-center gap-2">
            {plan.amount_minor === 0 ? 'Start free' : `Choose ${plan.name}`}
            <span className="transition-transform group-hover:translate-x-0.5">→</span>
          </span>
        </a>
        {plan.requires_keys && (
          <p className="mt-2.5 font-mono text-[9px] uppercase tracking-[0.18em] text-c2-text-faint">
            Needs your own openai / anthropic / gemini key
          </p>
        )}
      </div>
    </div>
  );
}
