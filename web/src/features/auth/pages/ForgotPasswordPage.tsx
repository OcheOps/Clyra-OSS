import { type FormEvent, useState } from 'react';
import { Link } from 'react-router-dom';
import toast from 'react-hot-toast';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { forgotPassword } from '@/services/api/auth';
import { AuthShell } from '@/features/auth/components/AuthShell';

export function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const trimmed = email.trim();
    if (!trimmed) return;
    setLoading(true);
    try {
      await forgotPassword(trimmed);
      setSent(true);
    } catch (err) {
      toast.error((err as Error).message || 'Something went wrong');
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthShell
      eyebrow="Reset access"
      heading="Forgot your password?"
      sub="Enter your email and we'll send a reset link."
      navAction={{ hint: 'Remembered?', to: '/login', label: 'Sign in' }}
      footer={
        <p className="text-petit text-c2-text-muted">
          Remembered it?{' '}
          <Link to="/login" className="text-c2-primary transition-colors hover:text-c2-info">
            Sign in
          </Link>
        </p>
      }
    >
      {sent ? (
        // Confirmation panel. Success-tinted icon + the actual email
        // echoed back so the user can immediately see whether they typed
        // the right address.
        <div className="space-y-3 text-center">
          <div className="mx-auto flex h-11 w-11 items-center justify-center rounded-full border border-c2-accent/30 bg-c2-accent-soft text-c2-accent">
            <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" />
              <polyline points="22,6 12,13 2,6" />
            </svg>
          </div>
          <p className="text-sm font-semibold text-c2-text">Check your inbox.</p>
          <p className="text-petit leading-relaxed text-c2-text-muted">
            If <span className="font-mono text-[11px] text-c2-text">{email}</span> is registered,
            a reset link is on its way. The link is good for one hour.
          </p>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="Email"
            type="email"
            placeholder="you@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
            autoFocus
          />
          <Button type="submit" className="w-full" loading={loading} disabled={!email.trim()}>
            Send reset link
          </Button>
        </form>
      )}
    </AuthShell>
  );
}
