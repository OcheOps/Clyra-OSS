import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '@/types/auth';

interface AuthState {
  user: User | null;
  // hydrated flips true once the boot-time /auth/me probe finishes (either
  // a user came back or 401 confirmed guest). Guards block redirect until
  // then so an OAuth-callback landing on /dashboard doesn't bounce to /login
  // before the cookie-backed session has a chance to hydrate the store.
  hydrated: boolean;
  setUser: (user: User | null) => void;
  setHydrated: (v: boolean) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      hydrated: false,
      setUser: (user) => set({ user }),
      setHydrated: (hydrated) => set({ hydrated }),
      logout: () => set({ user: null }),
    }),
    {
      name: 'clyra-auth',
      partialize: (s) => ({ user: s.user }),
    },
  ),
);
