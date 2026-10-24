import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type ThemePreference = 'light' | 'dark' | 'system';

interface ThemeState {
  preference: ThemePreference;
  // resolved is what's actually applied ('light' | 'dark') after
  // collapsing 'system' against the OS preference.
  resolved: 'light' | 'dark';
  setPreference: (p: ThemePreference) => void;
  // Called by the root App to sync the OS media-query — keeps the store
  // as single source of truth for the whole app.
  _setSystemResolved: (resolved: 'light' | 'dark') => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set, get) => ({
      preference: 'dark',
      resolved: 'dark',
      setPreference: (p) => {
        const sysDark =
          typeof window !== 'undefined' &&
          window.matchMedia('(prefers-color-scheme: dark)').matches;
        const resolved =
          p === 'system' ? (sysDark ? 'dark' : 'light') : p;
        set({ preference: p, resolved });
      },
      _setSystemResolved: (resolved) => {
        if (get().preference === 'system') set({ resolved });
      },
    }),
    {
      name: 'clyra-theme',
      partialize: (s) => ({ preference: s.preference }),
    },
  ),
);
