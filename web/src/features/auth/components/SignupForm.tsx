import { type FormEvent, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { PasswordInput } from '@/components/ui/PasswordInput';
import { useRegister } from '@/features/auth/hooks/useAuth';

export function SignupForm() {
  const [searchParams] = useSearchParams();
  const inviteCode = searchParams.get('invite') ?? '';
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [confirmError, setConfirmError] = useState('');
  const register = useRegister();

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (password !== confirm) {
      setConfirmError('Passwords do not match');
      return;
    }
    setConfirmError('');
    register.mutate({ email, password, invite_code: inviteCode });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {inviteCode && (
        <div className="rounded-clyra bg-brand-500/5 border border-brand-500/15 px-3 py-2 text-tiny text-brand-400">
          Invite code applied
        </div>
      )}
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
      <PasswordInput
        label="Password"
        placeholder="Min. 8 characters"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        required
        minLength={8}
        autoComplete="new-password"
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
        loading={register.isPending}
      >
        Create account
      </Button>
    </form>
  );
}
