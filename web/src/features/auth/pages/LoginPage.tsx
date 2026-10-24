import { useEffect } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import toast from 'react-hot-toast';
import { LoginForm } from '@/features/auth/components/LoginForm';
import { OAuthButtons } from '@/features/auth/components/OAuthButtons';
import { AuthShell } from '@/features/auth/components/AuthShell';

// oauth_error codes match internal/auth/oauth_handler.go#mapOAuthError.
const OAUTH_ERROR_MESSAGES: Record<string, string> = {
  denied: 'Sign-in was canceled. Try again when ready.',
  unverified: "That provider account's email isn't verified. Verify it and retry.",
  invite_required:
    'New accounts are invite-only right now. Sign in with an existing account or request access.',
  state: 'Sign-in link expired. Please try again.',
  disabled: 'That sign-in option is currently disabled.',
  upstream: "Couldn't reach the sign-in provider. Try again in a moment.",
  server: 'Something went wrong. Try again.',
};

export function LoginPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const oauthError = searchParams.get('oauth_error');
  // `?reason=expired` is set by the API client when a request 401s on a
  // protected route — surface a clearer toast so the user knows to just
  // sign back in rather than thinking they typed bad creds.
  const expired = searchParams.get('reason') === 'expired';

  useEffect(() => {
    if (!oauthError && !expired) return;
    if (expired) {
      toast('Your session expired — sign in to continue.', { icon: '🔒' });
    } else if (oauthError) {
      toast.error(OAUTH_ERROR_MESSAGES[oauthError] ?? OAUTH_ERROR_MESSAGES.server!);
    }
    const next = new URLSearchParams(searchParams);
    next.delete('oauth_error');
    next.delete('reason');
    setSearchParams(next, { replace: true });
  }, [oauthError, expired, searchParams, setSearchParams]);

  return (
    <AuthShell
      eyebrow="Sign in"
      heading="Welcome back."
      sub="Continue where you left off."
      navAction={{ hint: 'New here?', to: '/signup', label: 'Create account' }}
      footer={
        <>
          <Link
            to="/forgot-password"
            className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint transition-colors hover:text-c2-text-muted"
          >
            Forgot your password?
          </Link>
          <p className="text-petit text-c2-text-muted">
            New here?{' '}
            <Link to="/signup" className="text-c2-primary transition-colors hover:text-c2-info">
              Create an account
            </Link>
          </p>
        </>
      }
    >
      <OAuthButtons />
      <div className="mt-4">
        <LoginForm />
      </div>
    </AuthShell>
  );
}
