import { useUpgradeModalStore } from '@/features/billing/upgradeModalStore';

/**
 * BYOLockedCard — v2 restyle.
 *
 * Shown on Settings when the user isn't on the BYO plan. Keeps the
 * section slot visible so users know the feature exists, while making
 * clear the keys input is only unlocked by upgrading. Removes the old
 * footgun where free users typed an OpenAI key expecting unlimited
 * runs.
 */
export function BYOLockedCard() {
  const openUpgrade = useUpgradeModalStore((s) => s.openModal);

  return (
    <section className="overflow-hidden rounded-2xl border border-c2-border bg-c2-surface/60 p-6 transition-colors hover:border-c2-border-strong">
      <div className="flex items-center gap-2">
        <span className="inline-block h-1 w-1 rounded-full bg-c2-primary" />
        <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
          AI keys
        </span>
      </div>
      <h2 className="mt-3 font-display text-base font-semibold tracking-[-0.005em] text-c2-text">
        Your AI keys
      </h2>

      <div className="mt-5 flex items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="text-[13px] leading-relaxed text-c2-text-muted">
            Available on the{' '}
            <strong className="font-semibold text-c2-text">BYO</strong> plan.
            Bring your own OpenAI / Anthropic / Gemini keys and get{' '}
            <strong className="font-semibold text-c2-text">
              1,000 generations per account
            </strong>{' '}
            — the cheapest path if you already pay an AI provider.
          </p>
          <p className="mt-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            Keys stored in your browser only · never transmitted
          </p>
        </div>
        <button
          type="button"
          onClick={() => openUpgrade({ defaultPlan: 'byo' })}
          className="group inline-flex shrink-0 items-center gap-2 rounded-full bg-c2-primary px-4 py-2 text-tiny font-medium text-c2-bg shadow-c2-glow transition-all hover:bg-c2-primary-hover hover:shadow-[0_0_32px_-6px_rgba(6,182,212,0.55)]"
        >
          Go BYO
          <span className="transition-transform group-hover:translate-x-0.5">→</span>
        </button>
      </div>
    </section>
  );
}
