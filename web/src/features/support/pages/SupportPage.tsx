import { useState } from 'react';
import { Link } from 'react-router-dom';
import toast from 'react-hot-toast';
import { useAuthStore } from '@/features/auth/store';
import { submitFeedback, type FeedbackCategory } from '@/services/api/feedback';

const CATEGORIES: { value: FeedbackCategory; label: string; hint: string }[] = [
  { value: 'general',  label: 'General',         hint: 'Anything else.' },
  { value: 'bug',      label: 'Bug',             hint: 'Something broke or behaved unexpectedly.' },
  { value: 'feature',  label: 'Feature request', hint: 'You wish Clyra did this thing.' },
  { value: 'billing',  label: 'Billing',         hint: 'Payments, plans, refunds.' },
  { value: 'security', label: 'Security',        hint: 'Vulnerability or suspicious activity. We respond quickly.' },
];

/**
 * SupportPage — v2.
 *
 * Public contact form available without login so visitors who hit a
 * snag during signup can still reach us. When signed in, name + email
 * pre-fill from the session and the backend ties the submission to the
 * user_id. Form behavior is unchanged from v1.
 *
 * Visual treatment: same dark substrate + dot grid + soft cyan haze
 * as the rest of the marketing surfaces. Inputs hand-built in v2
 * tokens rather than going through the shared Input component because
 * this page wants a slightly heavier form style (larger touch targets,
 * inline category pills, character counter) that doesn't generalize.
 */
export function SupportPage() {
  const user = useAuthStore((s) => s.user);

  const [name, setName] = useState(user?.email?.split('@')[0] ?? '');
  const [email, setEmail] = useState(user?.email ?? '');
  const [category, setCategory] = useState<FeedbackCategory>('general');
  const [message, setMessage] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);

  const isLoggedIn = !!user;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (message.trim().length < 10) {
      toast.error('Tell us a little more — at least 10 characters.');
      return;
    }
    setSubmitting(true);
    try {
      await submitFeedback({
        name: name.trim() || (isLoggedIn ? user!.email : 'Anonymous'),
        // For logged-in users the backend overrides this anyway; we still send
        // it so the form works for visitors.
        email: email.trim(),
        category,
        message: message.trim(),
      });
      setSubmitted(true);
      setMessage('');
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Could not send — please try again.';
      toast.error(msg);
    } finally {
      setSubmitting(false);
    }
  };

  const activeHint = CATEGORIES.find((c) => c.value === category)?.hint;

  return (
    <div className="relative min-h-screen overflow-x-hidden bg-c2-bg font-brand text-c2-text antialiased">
      <div
        aria-hidden
        className="pointer-events-none fixed inset-0"
        style={{
          backgroundImage:
            'radial-gradient(circle at 1px 1px, rgba(148,163,184,0.06) 1px, transparent 0)',
          backgroundSize: '24px 24px',
        }}
      />
      <div
        aria-hidden
        className="pointer-events-none fixed left-1/2 top-0 h-[500px] w-[1000px] -translate-x-1/2 opacity-30 blur-[120px]"
        style={{
          background:
            'radial-gradient(ellipse at center top, rgba(6,182,212,0.18) 0%, transparent 60%)',
        }}
      />

      <div className="relative">
        <nav className="mx-auto flex max-w-4xl items-center justify-between px-6 py-6">
          <Link
            to="/"
            className="text-lg font-display font-semibold tracking-[-0.015em] text-c2-text"
          >
            clyra<span className="text-c2-primary">.</span>
          </Link>
          <Link
            to={isLoggedIn ? '/dashboard' : '/login'}
            className="rounded-full px-3 py-1.5 text-tiny text-c2-text-muted transition-colors hover:text-c2-text"
          >
            {isLoggedIn ? 'Back to app' : 'Sign in'}
          </Link>
        </nav>

        <main className="mx-auto max-w-2xl px-6 pb-32 pt-12">
          <div className="flex items-center gap-2">
            <span className="inline-block h-1.5 w-1.5 rounded-full bg-c2-primary shadow-[0_0_8px_rgba(6,182,212,0.6)]" />
            <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
              Support
            </span>
          </div>

          <h1 className="mt-5 font-display text-4xl font-semibold leading-[1.05] tracking-[-0.03em] text-c2-text md:text-5xl">
            Tell us what's
            <br />
            <span className="text-c2-text-muted">on your mind.</span>
          </h1>
          <p className="mt-5 max-w-xl text-[15px] leading-relaxed text-c2-text-muted">
            A real person reads every message. Bug reports, feature requests,
            billing questions, security disclosures — they all land in the same
            inbox. Average response is under a day.
          </p>

          {submitted ? (
            <div className="mt-12 rounded-2xl border border-c2-accent/30 bg-c2-accent-soft p-6">
              <div className="flex items-start gap-4">
                <div className="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-c2-accent/30 bg-c2-accent-soft text-c2-accent">
                  <svg
                    className="h-5 w-5"
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
                <div className="min-w-0 flex-1">
                  <p className="text-sm font-semibold text-c2-text">
                    Thanks — we got it.
                  </p>
                  <p className="mt-1.5 text-petit text-c2-text-muted">
                    We'll reply to{' '}
                    <span className="font-mono text-[11px] text-c2-text">{email}</span>{' '}
                    as soon as we can.
                  </p>
                  <button
                    type="button"
                    onClick={() => setSubmitted(false)}
                    className="mt-4 inline-flex items-center gap-1 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-accent transition-colors hover:text-c2-text"
                  >
                    Send another message →
                  </button>
                </div>
              </div>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="mt-12 space-y-6">
              <div className="grid gap-5 md:grid-cols-2">
                <Field id="sup-name" label="Your name">
                  <input
                    id="sup-name"
                    type="text"
                    required
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    className={INPUT_CLASSES}
                  />
                </Field>
                <Field
                  id="sup-email"
                  label={
                    <>
                      Email{' '}
                      {isLoggedIn && (
                        <span className="font-mono text-[10px] uppercase tracking-[0.14em] text-c2-text-faint">
                          (from your account)
                        </span>
                      )}
                    </>
                  }
                >
                  <input
                    id="sup-email"
                    type="email"
                    required
                    disabled={isLoggedIn}
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className={`${INPUT_CLASSES} disabled:cursor-not-allowed disabled:bg-c2-elevated/30 disabled:text-c2-text-faint`}
                  />
                </Field>
              </div>

              <Field id="sup-cat" label="Category">
                <div className="grid grid-cols-2 gap-2 sm:grid-cols-5">
                  {CATEGORIES.map((c) => {
                    const active = category === c.value;
                    return (
                      <button
                        key={c.value}
                        type="button"
                        onClick={() => setCategory(c.value)}
                        title={c.hint}
                        className={`rounded-lg border px-3 py-2 text-tiny font-medium transition-colors ${
                          active
                            ? 'border-c2-primary/50 bg-c2-primary-soft text-c2-primary'
                            : 'border-c2-border bg-c2-surface/40 text-c2-text-muted hover:border-c2-border-strong hover:text-c2-text'
                        }`}
                      >
                        {c.label}
                      </button>
                    );
                  })}
                </div>
                <p className="mt-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
                  {activeHint}
                </p>
              </Field>

              <Field id="sup-msg" label="Message">
                <textarea
                  id="sup-msg"
                  required
                  rows={8}
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  placeholder={
                    category === 'bug'
                      ? 'What did you do, what did you expect, what happened? Include browser + diagram URL if relevant.'
                      : category === 'security'
                        ? 'A short summary is enough — we will reply with our PGP key for full details.'
                        : 'Take your time.'
                  }
                  className={`${INPUT_CLASSES} resize-y`}
                  minLength={10}
                  maxLength={4000}
                />
                <div className="mt-2 flex justify-between font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
                  <span>Minimum 10 characters</span>
                  <span>{message.length} / 4000</span>
                </div>
              </Field>

              <div className="flex flex-col items-stretch justify-between gap-4 sm:flex-row sm:items-center">
                <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
                  By submitting you agree to our{' '}
                  <Link
                    to="/privacy"
                    className="text-c2-primary underline decoration-c2-primary/30 underline-offset-2 transition-colors hover:text-c2-info"
                  >
                    privacy policy
                  </Link>
                </p>
                <button
                  type="submit"
                  disabled={submitting}
                  className="group inline-flex items-center justify-center gap-2 rounded-full bg-c2-primary px-5 py-2.5 text-sm font-medium text-c2-bg shadow-c2-glow transition-all hover:bg-c2-primary-hover hover:shadow-[0_0_32px_-6px_rgba(6,182,212,0.55)] disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {submitting ? 'Sending…' : 'Send message'}
                  {!submitting && (
                    <span className="transition-transform group-hover:translate-x-0.5">→</span>
                  )}
                </button>
              </div>
            </form>
          )}

          <div className="mt-20 border-t border-c2-hairline pt-8">
            <Link
              to="/"
              className="inline-flex items-center gap-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint transition-colors hover:text-c2-text"
            >
              ← Back to home
            </Link>
          </div>
        </main>
      </div>
    </div>
  );
}

// ────────────────────────────────────────────────────────────────────
// Local form primitives — tuned for the support page, not exported.

const INPUT_CLASSES =
  'w-full rounded-lg border border-c2-border bg-c2-surface/40 px-3 py-2.5 text-sm text-c2-text ' +
  'placeholder:text-c2-text-faint focus:border-c2-primary/60 focus:bg-c2-surface focus:outline-none ' +
  'focus:ring-2 focus:ring-c2-primary-ring/40 transition-colors';

function Field({
  id,
  label,
  children,
}: {
  id: string;
  label: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-2">
      <label
        htmlFor={id}
        className="block font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint"
      >
        {label}
      </label>
      {children}
    </div>
  );
}
