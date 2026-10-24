import { useEffect, useState } from 'react';

// Step 2 of the two-step onboarding: after a first-run user's very first
// generation completes and the canvas has content, point them at the
// FloatingPromptDock and teach the tweak-refine loop. Skippable, keyed
// off localStorage so it never fires again once dismissed.
//
// Step 1 (the prompt itself) lives in EmptyCanvasPrompt — that surface
// already teaches "describe your system". This one teaches "now change
// it in words" — the differentiator most new users miss because they
// think the diagram is the terminal artifact.

const SEEN_KEY = 'clyra-tweak-coach-seen';

interface Props {
  /** True when the editor has at least one node — enables coach visibility. */
  hasContent: boolean;
}

export function TweakCoach({ hasContent }: Props) {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (!hasContent) return;
    if (localStorage.getItem(SEEN_KEY) === '1') return;
    // Small delay so it doesn't collide with the toast + reveal animation.
    const t = setTimeout(() => setVisible(true), 900);
    return () => clearTimeout(t);
  }, [hasContent]);

  const dismiss = () => {
    localStorage.setItem(SEEN_KEY, '1');
    setVisible(false);
  };

  if (!visible) return null;

  return (
    <div
      role="dialog"
      aria-label="Tip: tweak your diagram in words"
      className="pointer-events-auto absolute bottom-24 left-1/2 z-40 -translate-x-1/2 animate-in fade-in slide-in-from-bottom-2 duration-300"
    >
      <div className="relative max-w-sm rounded-2xl border border-c2-primary/40 bg-c2-elevated shadow-c2-lg p-4">
        {/* pointer arrow toward the dock */}
        <div className="absolute -bottom-1.5 left-1/2 h-3 w-3 -translate-x-1/2 rotate-45 border-b border-r border-c2-primary/40 bg-c2-elevated" />

        <div className="flex items-start gap-3">
          <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-c2-primary/15 text-c2-primary">
            <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M12 20h9" />
              <path d="M16.5 3.5a2.121 2.121 0 013 3L7 19l-4 1 1-4L16.5 3.5z" />
            </svg>
          </div>
          <div className="min-w-0">
            <p className="text-tiny font-medium text-c2-text">
              Now tweak it in plain English
            </p>
            <p className="mt-1 text-micro leading-relaxed text-c2-text-muted">
              Say things like <span className="text-c2-text">&ldquo;add a Redis cache&rdquo;</span> or
              {' '}
              <span className="text-c2-text">&ldquo;move Postgres to a private subnet&rdquo;</span>
              . Clyra edits the diagram — you don&rsquo;t redraw it.
            </p>
            <div className="mt-3 flex items-center justify-end gap-2">
              <button
                type="button"
                onClick={dismiss}
                className="rounded-md px-3 py-1 text-micro font-medium text-c2-text-muted hover:text-c2-text transition-colors"
              >
                Got it
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
