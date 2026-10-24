import { Link } from 'react-router-dom';

/**
 * AuthShell — shared page chrome for every v2 auth surface (login,
 * signup, forgot-password, reset-password).
 *
 * One source of truth for: dark substrate + dot grid + cyan halo
 * atmosphere, top nav (wordmark left, contextual link + chip right),
 * centered max-w-md column, eyebrow + headline + sub, the halo'd
 * auth card itself, and the footer-link prompt.
 *
 * Each auth page passes its own copy slots (eyebrow / heading / sub /
 * nav action / footer prompt) and renders the form inside `children`.
 * The visual rhythm stays identical across all four pages — only the
 * inner form + microcopy changes — which makes the auth surfaces read
 * as a coherent system instead of four parallel-but-different forms.
 */

interface AuthShellProps {
  /** Mono small-caps text in the eyebrow above the headline. */
  eyebrow: string;
  /** Big display headline. Can be a string or JSX (e.g. for a colored span). */
  heading: React.ReactNode;
  /** One-line supporting copy under the headline. */
  sub: React.ReactNode;
  /** Top-right nav: a mono label + a chip link. e.g. { hint: 'New here?', to: '/signup', label: 'Create account' } */
  navAction?: { hint: string; to: string; label: string };
  /** Optional footer prompt under the auth card. */
  footer?: React.ReactNode;
  /** The form itself (or any focal content) renders inside the halo'd card. */
  children: React.ReactNode;
}

export function AuthShell({
  eyebrow,
  heading,
  sub,
  navAction,
  footer,
  children,
}: AuthShellProps) {
  return (
    <div className="relative min-h-screen overflow-hidden bg-c2-bg font-brand text-c2-text antialiased">
      {/* Engineering substrate */}
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0"
        style={{
          backgroundImage:
            'radial-gradient(circle at 1px 1px, rgba(148,163,184,0.06) 1px, transparent 0)',
          backgroundSize: '24px 24px',
        }}
      />
      {/* Cyan haze pulling the eye to the form below */}
      <div
        aria-hidden
        className="pointer-events-none absolute left-1/2 top-0 h-[600px] w-[1000px] -translate-x-1/2 opacity-40 blur-[100px]"
        style={{
          background:
            'radial-gradient(ellipse at center top, rgba(6,182,212,0.22) 0%, transparent 60%)',
        }}
      />

      <div className="relative">
        <nav className="mx-auto flex max-w-6xl items-center justify-between px-6 py-6">
          <Wordmark />
          {navAction && (
            <div className="flex items-center gap-1">
              <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
                {navAction.hint}
              </span>
              <Link
                to={navAction.to}
                className="ml-1 rounded-full bg-c2-text px-3 py-1.5 text-tiny font-medium text-c2-bg transition-colors hover:bg-c2-text-muted"
              >
                {navAction.label}
              </Link>
            </div>
          )}
        </nav>

        <main className="mx-auto flex max-w-md flex-col px-6 pt-12 pb-20">
          <div className="mb-8 text-center">
            <Eyebrow>{eyebrow}</Eyebrow>
            <h1 className="mt-5 font-display text-3xl font-semibold leading-[1.05] tracking-[-0.025em] text-c2-text md:text-[36px]">
              {heading}
            </h1>
            <p className="mt-3 text-[14px] leading-relaxed text-c2-text-muted">
              {sub}
            </p>
          </div>

          {/* Halo'd auth card — the focal moment of every auth page. */}
          <div className="relative">
            <div
              aria-hidden
              className="pointer-events-none absolute -inset-8 rounded-3xl opacity-60 blur-2xl"
              style={{
                background:
                  'radial-gradient(circle at 50% 30%, rgba(6,182,212,0.18) 0%, transparent 60%)',
              }}
            />
            <div className="relative rounded-2xl border border-c2-border bg-c2-surface/80 p-6 shadow-c2-lg backdrop-blur-sm">
              {children}
            </div>
          </div>

          {footer && (
            <div className="mt-6 flex flex-col items-center gap-3 text-center">
              {footer}
            </div>
          )}
        </main>
      </div>
    </div>
  );
}

function Eyebrow({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-center gap-2">
      <span className="inline-block h-1.5 w-1.5 rounded-full bg-c2-primary shadow-[0_0_8px_rgba(6,182,212,0.6)]" />
      <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
        {children}
      </span>
    </div>
  );
}

function Wordmark() {
  return (
    <Link
      to="/"
      className="text-lg font-display font-semibold tracking-[-0.015em] text-c2-text"
    >
      clyra<span className="text-c2-primary">.</span>
    </Link>
  );
}
