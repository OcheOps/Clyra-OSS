import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { router } from '@/router';
import { useAuthStore } from '@/features/auth/store';
import { useThemeStore } from '@/stores/themeStore';
import { useEditorStore } from '@/features/editor/store';
import { usePromptSessionStore } from '@/features/editor/promptSessionStore';

// A minimal, hand-rolled command palette. Cmd/Ctrl+K toggles. Typeahead
// filters by substring across every command's label + keywords. No heavy
// dep — the whole thing is a portal, an input, and a list.

interface Command {
  id: string;
  label: string;
  group: 'Navigate' | 'Diagram' | 'Account' | 'View';
  keywords?: string[];
  when?: () => boolean;
  run: () => void;
}

function matches(cmd: Command, q: string): boolean {
  if (!q) return true;
  const hay = [cmd.label, cmd.group, ...(cmd.keywords ?? [])].join(' ').toLowerCase();
  const needles = q.toLowerCase().split(/\s+/).filter(Boolean);
  return needles.every((n) => hay.includes(n));
}

export function CommandPalette() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [active, setActive] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const navigate = (path: string) => {
    void router.navigate(path);
  };
  const logout = useAuthStore((s) => s.logout);
  const user = useAuthStore((s) => s.user);
  const setPreference = useThemeStore((s) => s.setPreference);
  const clearCanvas = useEditorStore((s) => s.clearCanvas);
  const showEmptyCard = usePromptSessionStore((s) => s.showEmptyCard);
  const resetSession = usePromptSessionStore((s) => s.resetSession);

  const close = useCallback(() => {
    setOpen(false);
    setQuery('');
    setActive(0);
  }, []);

  // Global hotkey
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const k = e.key.toLowerCase();
      if ((e.metaKey || e.ctrlKey) && k === 'k') {
        // Don't steal ⌘K when the user is already typing in an input
        // whose parent markup opts out — for now, always intercept.
        e.preventDefault();
        setOpen((o) => !o);
      } else if (e.key === 'Escape' && open) {
        close();
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open, close]);

  const commands: Command[] = useMemo(() => {
    const runAndClose = (fn: () => void) => () => {
      fn();
      close();
    };
    const list: Command[] = [
      {
        id: 'nav.dashboard',
        label: 'Go to Dashboard',
        group: 'Navigate',
        keywords: ['home'],
        run: runAndClose(() => navigate('/dashboard')),
      },
      {
        id: 'nav.projects',
        label: 'Open Projects',
        group: 'Navigate',
        run: runAndClose(() => navigate('/projects')),
      },
      {
        id: 'nav.editor',
        label: 'New Diagram',
        group: 'Diagram',
        keywords: ['create', 'blank', 'empty'],
        run: runAndClose(() => {
          clearCanvas();
          resetSession();
          showEmptyCard();
          navigate('/editor');
        }),
      },
      {
        id: 'diagram.show-welcome',
        label: 'Show welcome prompt card',
        group: 'Diagram',
        keywords: ['empty', 'card', 'welcome', 'reopen'],
        run: runAndClose(() => showEmptyCard()),
      },
      {
        id: 'nav.billing',
        label: 'Billing & Plans',
        group: 'Account',
        keywords: ['upgrade', 'pro', 'byo', 'subscription'],
        run: runAndClose(() => navigate('/billing')),
      },
      {
        id: 'nav.settings',
        label: 'Settings',
        group: 'Account',
        keywords: ['profile', 'password'],
        run: runAndClose(() => navigate('/settings')),
      },
      {
        id: 'nav.admin',
        label: 'Admin',
        group: 'Account',
        keywords: ['users', 'grant', 'metrics', 'audit', 'feedback'],
        when: () => user?.role === 'admin',
        run: runAndClose(() => navigate('/admin')),
      },
      {
        id: 'nav.support',
        label: 'Support / Feedback',
        group: 'Account',
        keywords: ['help', 'bug', 'contact', 'feature'],
        run: runAndClose(() => navigate('/support')),
      },
      {
        id: 'view.theme.dark',
        label: 'Switch to Dark',
        group: 'View',
        keywords: ['theme', 'mode'],
        run: runAndClose(() => setPreference('dark')),
      },
      {
        id: 'view.theme.light',
        label: 'Switch to Light',
        group: 'View',
        keywords: ['theme', 'mode', 'paper'],
        run: runAndClose(() => setPreference('light')),
      },
      {
        id: 'view.theme.system',
        label: 'Match System Theme',
        group: 'View',
        keywords: ['auto', 'os'],
        run: runAndClose(() => setPreference('system')),
      },
      {
        id: 'account.logout',
        label: 'Sign out',
        group: 'Account',
        keywords: ['log out', 'exit'],
        when: () => !!user,
        run: runAndClose(() => void logout()),
      },
    ];
    return list.filter((c) => (c.when ? c.when() : true));
  }, [close, user, setPreference, clearCanvas, logout, resetSession, showEmptyCard]);

  const filtered = useMemo(
    () => commands.filter((c) => matches(c, query)),
    [commands, query],
  );

  // Clamp active index when the filter set shrinks
  useEffect(() => {
    if (active >= filtered.length) setActive(0);
  }, [filtered, active]);

  useEffect(() => {
    if (open) {
      // Focus the input on the next tick so React has rendered the portal.
      setTimeout(() => inputRef.current?.focus(), 0);
    }
  }, [open]);

  if (!open) return null;

  const grouped = filtered.reduce<Record<string, Command[]>>((acc, c) => {
    (acc[c.group] ||= []).push(c);
    return acc;
  }, {});

  const flatIds = filtered.map((c) => c.id);

  return createPortal(
    <div
      role="dialog"
      aria-modal="true"
      aria-label="Command palette"
      className="fixed inset-0 z-[70] flex items-start justify-center bg-black/40 backdrop-blur-sm pt-[15vh] px-4"
      onClick={close}
    >
      <div
        className="w-full max-w-lg overflow-hidden rounded-xl border border-line-default bg-clyra-elevated shadow-clyra-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <input
          ref={inputRef}
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setActive(0);
          }}
          onKeyDown={(e) => {
            if (e.key === 'ArrowDown') {
              e.preventDefault();
              setActive((i) => Math.min(i + 1, filtered.length - 1));
            } else if (e.key === 'ArrowUp') {
              e.preventDefault();
              setActive((i) => Math.max(i - 1, 0));
            } else if (e.key === 'Enter') {
              e.preventDefault();
              filtered[active]?.run();
            }
          }}
          placeholder="Type a command or search…"
          className="w-full border-b border-line-subtle bg-transparent px-4 py-3.5 text-sm text-ink-primary placeholder:text-ink-secondary/50 outline-none"
        />

        <div className="max-h-[50vh] overflow-y-auto py-1">
          {filtered.length === 0 && (
            <p className="px-4 py-8 text-center text-tiny text-ink-faint">
              No commands match.
            </p>
          )}
          {Object.entries(grouped).map(([group, items]) => (
            <div key={group} className="py-1">
              <p className="px-4 pb-1 pt-2 text-micro font-medium uppercase tracking-wider text-ink-faint">
                {group}
              </p>
              {items.map((c) => {
                const idx = flatIds.indexOf(c.id);
                const isActive = idx === active;
                return (
                  <button
                    key={c.id}
                    type="button"
                    onMouseEnter={() => setActive(idx)}
                    onClick={() => c.run()}
                    className={`flex w-full items-center justify-between px-4 py-2 text-left text-sm transition-colors ${
                      isActive
                        ? 'bg-brand-500/10 text-ink-primary'
                        : 'text-ink-secondary hover:bg-clyra-overlay/50'
                    }`}
                  >
                    <span>{c.label}</span>
                    {isActive && (
                      <span className="text-micro text-ink-faint">↵</span>
                    )}
                  </button>
                );
              })}
            </div>
          ))}
        </div>

        <div className="border-t border-line-subtle bg-clyra-base/60 px-4 py-2 text-micro text-ink-faint flex items-center justify-between">
          <span>↑↓ navigate · ↵ select · esc close</span>
          <span>⌘K</span>
        </div>
      </div>
    </div>,
    document.body,
  );
}
