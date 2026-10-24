import { useEffect, useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { verifyPayment } from '@/services/api/payments';
import { useAuthStore } from '@/features/auth/store';
import type { Payment } from '@/types/payments';

type Phase = 'verifying' | 'success' | 'failed' | 'not_found';

/**
 * BillingReturnPage — v2 restyle.
 *
 * Post-checkout landing. User comes here from Paystack or Polar's
 * success_url; we verify the reference + Polar checkout id, then show
 * one of four phases: verifying, success, failed, not_found.
 * Verification logic unchanged per round-3 guardrails.
 *
 * Visual: centered halo'd card (same focal-moment treatment as the
 * auth pages) on the dark substrate. Each phase has its own
 * orb-and-headline composition so the outcome reads in a glance.
 */
export function BillingReturnPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const setUser = useAuthStore((s) => s.setUser);
  const user = useAuthStore((s) => s.user);

  // Paystack sends either `reference` or `trxref`; our mock returns `reference`.
  const reference = params.get('reference') || params.get('trxref') || '';
  // Polar doesn't echo `checkout_id` on custom success_url, so we read the
  // ID from sessionStorage (stashed at initialize time). URL wins if the
  // provider ever does append it — keeps the door open for future changes.
  const polarCheckoutId = (() => {
    const fromUrl = params.get('checkout_id');
    if (fromUrl) return fromUrl;
    try {
      const raw = sessionStorage.getItem('clyra-pending-checkout');
      if (!raw) return undefined;
      const parsed = JSON.parse(raw) as {
        reference: string;
        external_id: string;
        provider: string;
      };
      if (
        parsed.reference === reference &&
        parsed.provider === 'polar' &&
        parsed.external_id
      ) {
        return parsed.external_id;
      }
    } catch {
      /* ignore */
    }
    return undefined;
  })();

  const [phase, setPhase] = useState<Phase>('verifying');
  const [payment, setPayment] = useState<Payment | null>(null);

  useEffect(() => {
    if (!reference) {
      setPhase('not_found');
      return;
    }
    let cancelled = false;
    (async () => {
      try {
        const p = await verifyPayment(reference, polarCheckoutId);
        if (cancelled) return;
        setPayment(p);
        if (p.status === 'paid') {
          setPhase('success');
          try {
            sessionStorage.removeItem('clyra-pending-checkout');
          } catch {
            /* ignore */
          }
          if (user) {
            setUser({ ...user, plan: p.plan, plan_expiry: plannedExpiry(p) });
          }
        } else {
          setPhase('failed');
        }
      } catch {
        if (!cancelled) setPhase('not_found');
      }
    })();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [reference]);

  return (
    <div className="relative flex min-h-screen items-center justify-center bg-c2-bg font-brand text-c2-text antialiased">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0"
        style={{
          backgroundImage:
            'radial-gradient(circle at 1px 1px, rgba(148,163,184,0.06) 1px, transparent 0)',
          backgroundSize: '24px 24px',
        }}
      />
      <div
        aria-hidden
        className="pointer-events-none absolute left-1/2 top-0 h-[600px] w-[1000px] -translate-x-1/2 opacity-40 blur-[100px]"
        style={{
          background:
            'radial-gradient(ellipse at center top, rgba(6,182,212,0.22) 0%, transparent 60%)',
        }}
      />

      <div className="relative w-full max-w-md px-4">
        <div className="relative">
          <div
            aria-hidden
            className="pointer-events-none absolute -inset-8 rounded-3xl opacity-60 blur-2xl"
            style={{
              background:
                'radial-gradient(circle at 50% 30%, rgba(6,182,212,0.18) 0%, transparent 60%)',
            }}
          />
          <div className="relative rounded-2xl border border-c2-border bg-c2-surface/80 p-8 text-center shadow-c2-lg backdrop-blur-sm">
            {phase === 'verifying' && (
              <>
                <Spinner />
                <h1 className="mt-5 font-display text-xl font-semibold tracking-[-0.02em] text-c2-text">
                  Confirming your payment…
                </h1>
                <p className="mt-2 text-petit text-c2-text-muted">
                  This only takes a second.
                </p>
              </>
            )}

            {phase === 'success' && payment && (
              <>
                <SuccessOrb />
                <h1 className="mt-5 font-display text-xl font-semibold tracking-[-0.02em] text-c2-text">
                  You're on {planLabel(payment.plan)}.
                </h1>
                <p className="mt-2 text-petit leading-relaxed text-c2-text-muted">
                  Thanks for the upgrade. Your new daily limit kicks in right away.
                </p>
                <div className="mt-7 flex justify-center gap-2">
                  <button
                    type="button"
                    onClick={() => navigate('/dashboard')}
                    className="group inline-flex items-center gap-2 rounded-full bg-c2-primary px-4 py-2 text-tiny font-medium text-c2-bg shadow-c2-glow transition-all hover:bg-c2-primary-hover hover:shadow-[0_0_32px_-6px_rgba(6,182,212,0.55)]"
                  >
                    Back to dashboard
                    <span className="transition-transform group-hover:translate-x-0.5">→</span>
                  </button>
                  <button
                    type="button"
                    onClick={() => navigate('/billing')}
                    className="rounded-full border border-c2-border bg-c2-surface/40 px-4 py-2 text-tiny font-medium text-c2-text transition-colors hover:border-c2-border-strong hover:bg-c2-surface"
                  >
                    View billing
                  </button>
                </div>
              </>
            )}

            {phase === 'failed' && (
              <>
                <WarnOrb />
                <h1 className="mt-5 font-display text-xl font-semibold tracking-[-0.02em] text-c2-text">
                  Payment didn't go through.
                </h1>
                <p className="mt-2 text-petit leading-relaxed text-c2-text-muted">
                  No charges were captured. You can try again from billing.
                </p>
                <div className="mt-7 flex items-center justify-center gap-3">
                  <button
                    type="button"
                    onClick={() => navigate('/billing')}
                    className="rounded-full bg-c2-primary px-4 py-2 text-tiny font-medium text-c2-bg shadow-c2-glow transition-all hover:bg-c2-primary-hover"
                  >
                    Try again
                  </button>
                  <Link
                    to="/dashboard"
                    className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint transition-colors hover:text-c2-text"
                  >
                    Back to dashboard
                  </Link>
                </div>
              </>
            )}

            {phase === 'not_found' && (
              <>
                <WarnOrb />
                <h1 className="mt-5 font-display text-xl font-semibold tracking-[-0.02em] text-c2-text">
                  We couldn't find that payment.
                </h1>
                <p className="mt-2 text-petit leading-relaxed text-c2-text-muted">
                  The link may have expired. Start a new upgrade from billing.
                </p>
                <div className="mt-7 flex justify-center">
                  <button
                    type="button"
                    onClick={() => navigate('/billing')}
                    className="rounded-full bg-c2-primary px-4 py-2 text-tiny font-medium text-c2-bg shadow-c2-glow transition-all hover:bg-c2-primary-hover"
                  >
                    Go to billing
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function plannedExpiry(p: Payment): string {
  if (p.verified_at) {
    const d = new Date(p.verified_at);
    d.setDate(d.getDate() + p.period_days);
    return d.toISOString();
  }
  const d = new Date();
  d.setDate(d.getDate() + p.period_days);
  return d.toISOString();
}

function planLabel(plan: string): string {
  if (plan === 'pro') return 'Pro';
  if (plan === 'byo') return 'BYO';
  return plan;
}

function Spinner() {
  return (
    <svg className="mx-auto h-8 w-8 animate-spin text-c2-primary" viewBox="0 0 24 24" fill="none">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeOpacity="0.2" strokeWidth="3" />
      <path fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" opacity="0.9" />
    </svg>
  );
}

function SuccessOrb() {
  return (
    <div className="relative mx-auto flex h-14 w-14 items-center justify-center">
      <div
        aria-hidden
        className="absolute inset-0 rounded-full opacity-60 blur-md"
        style={{
          background: 'radial-gradient(circle, rgba(16,185,129,0.5) 0%, transparent 60%)',
        }}
      />
      <div className="relative flex h-14 w-14 items-center justify-center rounded-full border border-c2-accent/40 bg-c2-accent-soft text-c2-accent">
        <svg
          className="h-7 w-7"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <polyline points="20 6 9 17 4 12" />
        </svg>
      </div>
    </div>
  );
}

function WarnOrb() {
  return (
    <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full border border-c2-danger/40 bg-c2-danger/15 text-c2-danger">
      <svg
        className="h-6 w-6"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="8" x2="12" y2="12" />
        <line x1="12" y1="16" x2="12.01" y2="16" />
      </svg>
    </div>
  );
}
