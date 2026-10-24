import { type FormEvent, useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import toast from 'react-hot-toast';
import { Button } from '@/components/ui/Button';
import { PasswordInput } from '@/components/ui/PasswordInput';
import { resetPassword } from '@/services/api/auth';
import { AuthShell } from '@/features/auth/components/AuthShell';

export function ResetPasswordPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get('token') ?? '';

  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [confirmError, setConfirmError] = useState('');
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (password.length < 8) {
      setConfirmError('Password must be at least 8 characters');
      return;
    }
    if (password !== confirm) {
      setConfirmError('Passwords do not match');
      return;
    }
    setConfirmError('');
    setLoading(true);
    try {
      await resetPassword(token, password);
      setDone(true);
      setTimeout(() => navigate('/login'), 1500);
    } catch (err) {
      toast.error((err as Error).message || 'Reset failed');
    } finally {
      setLoading(false);
    }
  };

  const invalidToken = !token || token.length < 32;

  return (
    <AuthShell
      eyebrow="New password"
      heading="Reset your password"
      sub="Choose a new password for your account."
      navAction={{ hint: 'Remembered?', to: '/login', label: 'Sign in' }}
      footer={
        <Link
          to="/login"
          className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint transition-colors hover:text-c2-text-muted"
        >
          Back to sign in
        </Link>
      }
    >
      {invalidToken ? (
        <div className="space-y-2 text-center">
          <p className="text-sm font-semibold text-c2-text">This reset link looks broken.</p>
          <p className="text-petit leading-relaxed text-c2-text-muted">
            Request a new one from the{' '}
            <Link
              to="/forgot-password"
              className="text-c2-primary transition-colors hover:text-c2-info"
            >
              forgot-password page
            </Link>
            .
          </p>
        </div>
      ) : done ? (
        <div className="space-y-2 text-center">
          <div className="mx-auto flex h-11 w-11 items-center justify-center rounded-full border border-c2-accent/30 bg-c2-accent-soft text-c2-accent">
            <svg
              className="h-5 w-5"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <polyline points="20 6 9 17 4 12" />
            </svg>
          </div>
          <p className="text-sm font-semibold text-c2-text">Password updated.</p>
          <p className="text-petit text-c2-text-muted">Taking you to sign in…</p>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="space-y-4">
          <PasswordInput
            label="New password"
            placeholder="Min. 8 characters"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
            autoComplete="new-password"
            autoFocus
          />
          <PasswordInput
            label="Confirm password"
            placeholder="Repeat your password"
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            error={confirmError}
            required
            minLength={8}
            autoComplete="new-password"
          />
          <Button
            type="submit"
            className="w-full"
            loading={loading}
            disabled={!password || !confirm}
          >
            Set new password
          </Button>
        </form>
      )}
    </AuthShell>
  );
}
