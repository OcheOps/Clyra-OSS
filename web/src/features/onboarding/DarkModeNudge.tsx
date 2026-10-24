import { useEffect, useState } from 'react';
import { useThemeStore } from '@/stores/themeStore';

// DarkModeNudge — one-time friendly prompt that asks users whose OS
// preference disagrees with our default (dark) whether they'd like to
// flip. Nudges only if:
//   - the user has never explicitly set a theme (still on dark default), and
//   - their OS actually prefers light (so the default is "wrong" for them)
// Otherwise it's silent.

const SEEN_KEY = 'clyra-theme-nudge';

export function DarkModeNudge() {
  const preference = useThemeStore((s) => s.preference);
  const setPreference = useThemeStore((s) => s.setPreference);
  const [show, setShow] = useState(false);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    if (localStorage.getItem(SEEN_KEY) === 'seen') return;
    // If user explicitly chose light or system, don't ask.
    if (preference !== 'dark') {
      localStorage.setItem(SEEN_KEY, 'seen');
      return;
    }
    const osLight = window.matchMedia('(prefers-color-scheme: light)').matches;
    if (!osLight) {
      localStorage.setItem(SEEN_KEY, 'seen');
      return;
    }
    const t = window.setTimeout(() => setShow(true), 4000);
    return () => window.clearTimeout(t);
  }, [preference]);

  const dismiss = (nextPref: 'light' | 'dark') => {
    setPreference(nextPref);
    localStorage.setItem(SEEN_KEY, 'seen');
    setShow(false);
  };

  if (!show) return null;

  return (
    <div
      role="dialog"
      aria-labelledby="theme-nudge-title"
      className="fixed bottom-6 right-6 z-40 w-[min(340px,calc(100vw-24px))] animate-[fadeIn_180ms_ease-out]"
    >
      <div className="rounded-xl border border-line-default bg-clyra-elevated p-4 shadow-clyra-lg [body.paper_&]:border-stone-subtle [body.paper_&]:bg-paper-raised [body.paper_&]:shadow-paper-lg">
        <p id="theme-nudge-title" className="text-sm font-semibold text-ink-primary [body.paper_&]:text-charcoal">
          Lights on?
        </p>
        <p className="mt-1 text-petit text-ink-secondary [body.paper_&]:text-charcoal-muted">
          We defaulted to dark — your OS is on light. Want to match? You can change it anytime from Settings.
        </p>
        <div className="mt-4 flex items-center justify-end gap-2">
          <button
            type="button"
            onClick={() => dismiss('dark')}
            className="rounded-md border border-line-default px-3 py-1.5 text-tiny font-medium text-ink-secondary hover:border-line-strong [body.paper_&]:border-stone-default [body.paper_&]:text-charcoal-soft [body.paper_&]:hover:border-stone-strong"
          >
            No, keep dark
          </button>
          <button
            type="button"
            onClick={() => dismiss('light')}
            className="rounded-md bg-brand-600 px-3 py-1.5 text-tiny font-medium text-white hover:bg-brand-500 [body.paper_&]:bg-charcoal [body.paper_&]:text-paper [body.paper_&]:hover:bg-charcoal-soft"
          >
            Switch to light
          </button>
        </div>
      </div>
    </div>
  );
}
