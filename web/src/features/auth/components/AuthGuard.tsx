import { type ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/features/auth/store';

// Redirects unauthenticated users to /login with a `redirect` param so they
// land back on the protected page after sign-in. Critical for the post-pay
// /billing/return flow: if the session expires mid-checkout, this hop keeps
// the reference query intact so verification still completes.
export function AuthGuard({ children }: { children: ReactNode }) {
  const user = useAuthStore((s) => s.user);
  const hydrated = useAuthStore((s) => s.hydrated);
  const location = useLocation();

  // Hold the redirect until the boot-time /auth/me probe finishes. Without
  // this, OAuth callbacks (which land on /dashboard with a fresh session
  // cookie but an empty client store) would flash back to /login.
  if (!hydrated) return <HydrationFallback />;

  if (!user) {
    const here = location.pathname + location.search;
    const target = `/login?redirect=${encodeURIComponent(here)}&reason=expired`;
    return <Navigate to={target} replace />;
  }

  return children;
}

function HydrationFallback() {
  return (
    <div className="flex h-screen items-center justify-center bg-clyra-bg">
      <div className="h-6 w-6 animate-spin rounded-full border-2 border-ink-secondary border-t-transparent" />
    </div>
  );
}
