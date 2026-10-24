import { RouterProvider } from 'react-router-dom';
import { Suspense, useEffect } from 'react';
import { router } from './router';
import { UpgradePromptModal } from '@/features/billing/UpgradePromptModal';
import { DarkModeNudge } from '@/features/onboarding/DarkModeNudge';
import { CommandPalette } from '@/components/CommandPalette';
import { useThemeStore } from '@/stores/themeStore';
import { useAuthStore } from '@/features/auth/store';
import { me as fetchMe } from '@/services/api/auth';

function App() {
  // Apply the resolved theme to <body> globally, and keep it in sync
  // with the OS media-query when the user picks 'system'.
  const resolved = useThemeStore((s) => s.resolved);
  const setSystemResolved = useThemeStore((s) => s._setSystemResolved);

  useEffect(() => {
    if (resolved === 'light') {
      document.body.classList.add('paper');
    } else {
      document.body.classList.remove('paper');
    }
  }, [resolved]);

  // One-shot session hydration. OAuth callbacks set HttpOnly cookies and
  // redirect to /dashboard; without this probe the Zustand store stays empty
  // and AuthGuard would bounce the user straight back to /login.
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const { user } = await fetchMe();
        if (!cancelled) useAuthStore.getState().setUser(user);
      } catch {
        if (!cancelled) useAuthStore.getState().setUser(null);
      } finally {
        if (!cancelled) useAuthStore.getState().setHydrated(true);
      }
    })();
    return () => { cancelled = true; };
  }, []);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    // Initial sync (the store was rehydrated from localStorage before
    // the media query was observed).
    setSystemResolved(mq.matches ? 'dark' : 'light');
    const onChange = (e: MediaQueryListEvent) => {
      setSystemResolved(e.matches ? 'dark' : 'light');
    };
    mq.addEventListener('change', onChange);
    return () => mq.removeEventListener('change', onChange);
  }, [setSystemResolved]);

  return (
    <>
      <Suspense fallback={<div className="min-h-screen bg-c2-bg p-8 text-c2-text-muted">Loading…</div>}>
        <RouterProvider router={router} />
      </Suspense>
      <CommandPalette />
      <UpgradePromptModal />
      <DarkModeNudge />
    </>
  );
}

export default App;
