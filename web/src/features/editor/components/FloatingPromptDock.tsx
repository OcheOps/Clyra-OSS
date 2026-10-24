import { useEffect, useRef, useState } from 'react';
import { PromptBar } from './PromptBar';
import { usePromptSessionStore } from '../promptSessionStore';

interface Props {
  onProjectRename?: (name: string) => void;
}

// FloatingPromptDock — adaptive Cursor-style bottom pill.
//
// State machine:
//   collapsed  — compact 'Refine your diagram · ⌘K' chip. Default when
//                idle (no focus, no text). Doesn't compete visually with
//                the canvas.
//   expanded   — full PromptBar mounted inside the pill. Triggered by
//                click, focus on the compact chip, or ⌘K shortcut.
//                Stays expanded as long as the input is focused OR has
//                content OR the user is in the middle of a follow-up.
//
// Why not always-expanded: the previous version put the full two-row
// prompt bar on screen permanently, which crowded the canvas and made
// the eye jump between diagram + bar. Cursor's cmd-K popover pattern
// keeps the dock small until you actively need it — the canvas remains
// the star.
export function FloatingPromptDock({ onProjectRename }: Props) {
  const [expanded, setExpanded] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const lastPrompt = usePromptSessionStore((s) => s.lastPrompt);
  const followupCount = usePromptSessionStore((s) => s.followupCount);

  // Keep expanded whenever a follow-up loop is mid-flight — collapsing
  // then would hide the '2 left' + refine input the user is watching.
  const followupActive = Boolean(lastPrompt) && followupCount < 2;

  // Global keyboard shortcuts: ⌘K / Ctrl+K expands the dock and focuses
  // the input. Esc collapses. Mirrors Cursor / Linear / most modern
  // command palettes so power users don't have to reach for the mouse.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const isMac = navigator.platform.toLowerCase().includes('mac');
      const shortcut = (isMac ? e.metaKey : e.ctrlKey) && e.key === 'k';
      if (shortcut) {
        e.preventDefault();
        setExpanded(true);
      } else if (e.key === 'Escape' && expanded) {
        // Only collapse on Escape when the input is empty — a user
        // actively typing shouldn't lose their draft. PromptBar's own
        // input owns the current text; we check the DOM directly since
        // lifting text state into this component would require passing
        // controlled props through PromptBar's signature.
        const input = containerRef.current?.querySelector<HTMLInputElement>('input[type="text"]');
        if (!input || !input.value.trim()) {
          setExpanded(false);
        }
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [expanded]);

  // Click outside collapses — but only when the input is empty and not
  // mid-followup. Prevents accidentally hiding a half-written prompt.
  useEffect(() => {
    if (!expanded) return;
    const onClick = (e: MouseEvent) => {
      const target = e.target as Node;
      if (containerRef.current?.contains(target)) return;
      const input = containerRef.current?.querySelector<HTMLInputElement>('input[type="text"]');
      if (input && input.value.trim()) return;
      if (followupActive) return;
      setExpanded(false);
    };
    document.addEventListener('mousedown', onClick);
    return () => document.removeEventListener('mousedown', onClick);
  }, [expanded, followupActive]);

  // Auto-expand when a follow-up becomes available so the tweak input
  // is visible without the user having to click first — the moment
  // after generation is exactly when they want to refine.
  useEffect(() => {
    if (followupActive) setExpanded(true);
  }, [followupActive]);

  // Focus the newly-mounted input on expand so ⌘K → type is one motion.
  useEffect(() => {
    if (!expanded) return;
    const t = setTimeout(() => {
      containerRef.current
        ?.querySelector<HTMLInputElement>('input[type="text"]')
        ?.focus();
    }, 30);
    return () => clearTimeout(t);
  }, [expanded]);

  return (
    <div
      className="pointer-events-none absolute inset-x-0 bottom-6 z-30 flex justify-center px-4"
      aria-label="Prompt bar"
    >
      <div
        ref={containerRef}
        className={`pointer-events-auto transition-all duration-200 ease-out ${
          expanded ? 'w-full max-w-2xl' : 'w-auto'
        }`}
      >
        {expanded ? (
          <div className="rounded-2xl border border-c2-border bg-c2-elevated/85 px-3 py-2 shadow-c2-lg backdrop-blur-xl">
            <PromptBar onProjectRename={onProjectRename} />
          </div>
        ) : (
          <CollapsedPill onClick={() => setExpanded(true)} hasPrior={Boolean(lastPrompt)} />
        )}
      </div>
    </div>
  );
}

// CollapsedPill — the compact idle state. Reads different copy depending
// on whether the user has generated once (refine framing) vs. is starting
// fresh (generate framing). Same visual affordance either way.
function CollapsedPill({
  onClick,
  hasPrior,
}: {
  onClick: () => void;
  hasPrior: boolean;
}) {
  const isMac = typeof navigator !== 'undefined' && navigator.platform.toLowerCase().includes('mac');
  const shortcut = isMac ? '⌘K' : 'Ctrl K';
  return (
    <button
      type="button"
      onClick={onClick}
      className="group flex items-center gap-2.5 rounded-full border border-c2-border bg-c2-elevated/85 py-2 pl-3 pr-2 text-tiny font-medium text-c2-text-muted shadow-c2-lg backdrop-blur-xl transition-colors hover:border-c2-primary/40 hover:text-c2-text"
    >
      <svg
        className="h-3.5 w-3.5 text-c2-primary"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M15 3h6v6" />
        <path d="M10 14L21 3" />
        <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
      </svg>
      <span>{hasPrior ? 'Refine or extend' : 'Prompt a new diagram'}</span>
      <span className="ml-1 rounded border border-c2-border bg-c2-bg/40 px-1.5 py-0.5 font-mono text-[9px] font-semibold uppercase tracking-[0.14em] text-c2-text-faint group-hover:border-c2-primary/40 group-hover:text-c2-primary">
        {shortcut}
      </span>
    </button>
  );
}
