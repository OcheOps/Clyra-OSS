import type { ReactElement } from 'react';
import { useThemeStore, type ThemePreference } from '@/stores/themeStore';

// ThemeToggle comes in two flavours:
//   - `variant="compact"` — icon-only segmented group for the topbar
//   - `variant="labeled"` — larger pill-selector with labels, for Settings

interface Props {
  variant?: 'compact' | 'labeled';
}

const OPTIONS: { value: ThemePreference; label: string; title: string; icon: ReactElement }[] = [
  { value: 'light', label: 'Light', title: 'Light', icon: <SunIcon /> },
  { value: 'dark', label: 'Dark', title: 'Dark', icon: <MoonIcon /> },
  { value: 'system', label: 'System', title: 'Match OS', icon: <SystemIcon /> },
];

export function ThemeToggle({ variant = 'compact' }: Props) {
  const preference = useThemeStore((s) => s.preference);
  const setPreference = useThemeStore((s) => s.setPreference);

  // Compact: a single icon button that cycles light → dark → system.
  // Keeping three inline buttons ate ~90px of topbar width and wasn't
  // worth it given that users pick a preference once. The labeled
  // variant in Settings remains a full radio group for discoverability.
  if (variant === 'compact') {
    const current = OPTIONS.find((o) => o.value === preference) ?? OPTIONS[1]!;
    const nextIdx = (OPTIONS.findIndex((o) => o.value === preference) + 1) % OPTIONS.length;
    const next = OPTIONS[nextIdx]!;
    return (
      <button
        type="button"
        aria-label={`Appearance: ${current.label}. Click to switch to ${next.label}.`}
        title={`Appearance: ${current.label} (click for ${next.label})`}
        onClick={() => setPreference(next.value)}
        className={
          'flex h-7 w-7 items-center justify-center rounded-md border border-line-subtle bg-clyra-base/60 text-ink-muted transition-colors hover:text-ink-primary hover:bg-clyra-raised ' +
          '[body.paper_&]:border-stone-subtle [body.paper_&]:bg-paper-raised [body.paper_&]:text-charcoal-muted [body.paper_&]:hover:text-charcoal'
        }
      >
        {current.icon}
      </button>
    );
  }

  return (
    <div
      role="radiogroup"
      aria-label="Appearance"
      className={
        'inline-flex items-center gap-0.5 rounded-md border border-line-subtle bg-clyra-base/60 p-0.5 ' +
        '[body.paper_&]:border-stone-subtle [body.paper_&]:bg-paper-raised'
      }
    >
      {OPTIONS.map((opt) => {
        const active = opt.value === preference;
        return (
          <button
            key={opt.value}
            type="button"
            role="radio"
            aria-checked={active}
            title={opt.title}
            onClick={() => setPreference(opt.value)}
            className={
              'flex items-center gap-1.5 rounded px-2 py-1 text-tiny font-medium transition-colors ' +
              (active
                ? 'bg-clyra-raised text-ink-primary [body.paper_&]:bg-charcoal [body.paper_&]:text-paper'
                : 'text-ink-muted hover:text-ink-primary [body.paper_&]:text-charcoal-muted [body.paper_&]:hover:text-charcoal')
            }
          >
            {opt.icon}
            <span>{opt.label}</span>
          </button>
        );
      })}
    </div>
  );
}

function SunIcon() {
  return (
    <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41" />
    </svg>
  );
}
function MoonIcon() {
  return (
    <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
    </svg>
  );
}
function SystemIcon() {
  return (
    <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="2" y="4" width="20" height="14" rx="2" />
      <path d="M8 22h8M12 18v4" />
    </svg>
  );
}
