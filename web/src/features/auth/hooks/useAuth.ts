import { useMutation } from '@tanstack/react-query';
import { useNavigate, useSearchParams } from 'react-router-dom';
import toast from 'react-hot-toast';
import * as authApi from '@/services/api/auth';
import { useAuthStore } from '@/features/auth/store';
import type { APIError } from '@/types/api';
import { clearStoredBYOKeys } from '@/features/editor/lib/byoKeyStorage';

// Honors ?redirect=<same-origin-path> so login/signup buttons on
// specific-intent surfaces (landing "Import your Terraform" → /editor?
// reverse=1, expired-session bounces, etc.) land the user exactly where
// they were headed instead of the generic /dashboard.
function safeRedirect(target: string | null): string {
  if (!target) return '/dashboard';
  if (!target.startsWith('/') || target.startsWith('//')) return '/dashboard';
  return target;
}

export function useLogin() {
  const setUser = useAuthStore((s) => s.setUser);
  const navigate = useNavigate();
  const [params] = useSearchParams();

  return useMutation({
    mutationFn: authApi.login,
    onSuccess: (data) => {
      setUser(data.user);
      navigate(safeRedirect(params.get('redirect')));
    },
    onError: (err: APIError) => {
      toast.error(err.message);
    },
  });
}

export function useRegister() {
  const setUser = useAuthStore((s) => s.setUser);
  const navigate = useNavigate();
  const [params] = useSearchParams();

  return useMutation({
    mutationFn: authApi.register,
    onSuccess: (data) => {
      setUser(data.user);
      navigate(safeRedirect(params.get('redirect')));
    },
    onError: (err: APIError) => {
      toast.error(err.message);
    },
  });
}

export function useLogout() {
  const logout = useAuthStore((s) => s.logout);
  const userId = useAuthStore((s) => s.user?.id) ?? '';
  const navigate = useNavigate();

  return useMutation({
    mutationFn: authApi.logout,
    onSuccess: () => {
      clearStoredBYOKeys(userId);
      logout();
      navigate('/login');
    },
    onError: () => {
      // Clear local state even if server call fails
      clearStoredBYOKeys(userId);
      logout();
      navigate('/login');
    },
  });
}
