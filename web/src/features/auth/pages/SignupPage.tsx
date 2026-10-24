import { Navigate, useSearchParams } from 'react-router-dom';
import { Link } from 'react-router-dom';
import { SignupForm } from '@/features/auth/components/SignupForm';
import { OAuthButtons } from '@/features/auth/components/OAuthButtons';
import { AuthShell } from '@/features/auth/components/AuthShell';
import { usePublicConfig } from '@/features/auth/hooks/usePublicConfig';

/**
 * SignupPage — v2.
 *
 * Same shell as the rest of the auth surfaces. Open-registration check
 * stays in this component because it's signup-specific: closed-registration
 * deployments without an `?invite=` query bounce back to the landing page
 * (the link from email carries the invite, so a valid invite link still
 * lands here cleanly).
 */
export function SignupPage() {
  const [searchParams] = useSearchParams();
  const invite = searchParams.get('invite');
  const { mode, loading } = usePublicConfig();

  // Hold render until config resolves to avoid a flash of the form on a
  // closed-registration deployment.
  if (loading) {
    return <div className="min-h-screen bg-c2-bg" />;
  }

  if (mode !== 'open' && !invite) {
    return <Navigate to="/" replace />;
  }

  return (
    <AuthShell
      eyebrow="Create account"
      heading={
        <>
          Start with
          <br />
          <span className="bg-gradient-to-r from-c2-primary to-c2-accent bg-clip-text text-transparent">
            ten free diagrams.
          </span>
        </>
      }
      sub="Ten diagrams per account, no credit card."
      navAction={{ hint: 'Have an account?', to: '/login', label: 'Sign in' }}
      footer={
        <p className="text-petit text-c2-text-muted">
          Already have an account?{' '}
          <Link to="/login" className="text-c2-primary transition-colors hover:text-c2-info">
            Sign in
          </Link>
        </p>
      }
    >
      <OAuthButtons label="Sign up with" />
      <div className="mt-4">
        <SignupForm />
      </div>
    </AuthShell>
  );
}
