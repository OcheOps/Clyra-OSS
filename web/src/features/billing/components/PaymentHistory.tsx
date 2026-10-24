import { useEffect, useState } from 'react';
import { listPayments } from '@/services/api/payments';
import type { Payment } from '@/types/payments';

/**
 * PaymentHistory — v2 restyle.
 *
 * Compact list of past payments. Each row shows the plan as a Switzer
 * label + a tiny status pill (paid / pending / failed / refunded), the
 * date + truncated provider reference as mono small-caps meta, and the
 * formatted amount on the right.
 */
export function PaymentHistory() {
  const [payments, setPayments] = useState<Payment[] | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const list = await listPayments();
        if (!cancelled) setPayments(list);
      } catch {
        if (!cancelled) setError(true);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  if (error) {
    return (
      <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
        Couldn't load billing history.
      </p>
    );
  }

  if (!payments) {
    return (
      <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
        Loading…
      </p>
    );
  }

  if (payments.length === 0) {
    return (
      <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
        No payments yet.
      </p>
    );
  }

  return (
    <div className="divide-y divide-c2-hairline">
      {payments.map((p) => (
        <div key={p.id} className="flex items-center justify-between py-3 first:pt-0 last:pb-0">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <span className="font-display text-[14px] font-semibold capitalize text-c2-text">
                {p.plan}
              </span>
              <StatusBadge status={p.status} />
            </div>
            <p className="mt-1 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
              {new Date(p.created_at).toLocaleDateString()} · ref{' '}
              {p.provider_ref.slice(0, 12)}…
            </p>
          </div>
          <span className="font-mono text-[13px] text-c2-text">
            {formatAmount(p.amount_minor, p.currency)}
          </span>
        </div>
      ))}
    </div>
  );
}

function formatAmount(minor: number, currency: string): string {
  const major = minor / 100;
  const sym = currency === 'NGN' ? '₦' : currency === 'USD' ? '$' : '';
  return `${sym}${major.toLocaleString(undefined, { maximumFractionDigits: 2 })}`;
}

function StatusBadge({ status }: { status: Payment['status'] }) {
  const style =
    status === 'paid'
      ? 'border-c2-accent/30 bg-c2-accent-soft text-c2-accent'
      : status === 'failed'
        ? 'border-c2-danger/30 bg-c2-danger/10 text-c2-danger'
        : status === 'refunded'
          ? 'border-c2-info/30 bg-c2-info/10 text-c2-info'
          : 'border-c2-border bg-c2-bg/40 text-c2-text-muted';
  return (
    <span
      className={`rounded-full border px-2 py-0.5 font-mono text-[9px] font-semibold uppercase tracking-[0.18em] ${style}`}
    >
      {status}
    </span>
  );
}
