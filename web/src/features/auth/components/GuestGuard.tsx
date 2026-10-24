import { type ReactNode } from 'react';
import { Navigate, useSearchParams } from 'react-router-dom';
import { useAuthStore } from '@/features/auth/store';

// Honors `?redirect=<path>` so a user bounced from a protected page by an
// expired-session 401 lands back where they were after re-auth.
function safeRedirect(target: string | null): string {
  if (!target) return '/dashboard';
  // Only allow same-origin paths — never let the query string smuggle in an
  // external host or a protocol-relative URL.
  if (!target.startsWith('/') || target.startsWith('//')) return '/dashboard';
  return target;
}

export function GuestGuard({ children }: { children: ReactNode }) {
  const user = useAuthStore((s) => s.user);
  const hydrated = useAuthStore((s) => s.hydrated);
  const [params] = useSearchParams();

  // Wait for the boot-time /auth/me probe — otherwise a logged-in user with
  // a valid session cookie briefly sees /login before getting redirected.
  if (!hydrated) {
    return (
      <div className="flex h-screen items-center justify-center bg-clyra-bg">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-ink-secondary border-t-transparent" />
      </div>
    );
  }

  if (user) {
    return <Navigate to={safeRedirect(params.get('redirect'))} replace />;
  }

  return children;
}
