import { create } from 'zustand';
import { useEffect } from 'react';
import { getDailyUsage, type DailyUsage } from '@/services/api/usage';
import { useAuthStore } from '@/features/auth/store';

interface UsageState {
  usage: DailyUsage | null;
  loading: boolean;
  fetchedAt: number;
  refresh: () => Promise<void>;
  clear: () => void;
}

// Module-level in-flight guard prevents thundering-herd when many components
// mount simultaneously and all call refresh().
let inflight: Promise<void> | null = null;

export const useUsageStore = create<UsageState>((set) => ({
  usage: null,
  loading: false,
  fetchedAt: 0,
  refresh: async () => {
    if (inflight) return inflight;
    const p = (async () => {
      set({ loading: true });
      try {
        const u = await getDailyUsage();
        set({ usage: u, loading: false, fetchedAt: Date.now() });
      } catch {
        set({ loading: false });
      } finally {
        inflight = null;
      }
    })();
    inflight = p;
    return p;
  },
  clear: () => set({ usage: null, fetchedAt: 0 }),
}));

// React hook — auto-fetches on mount and refreshes when user changes (login /
// logout). Components that trigger usage-changing actions (generate, cancel)
// should call `useUsageStore.getState().refresh()` after their mutation to
// keep the topbar + modal state accurate.
export function useDailyUsage() {
  const user = useAuthStore((s) => s.user);
  const { usage, loading, refresh, clear } = useUsageStore();

  useEffect(() => {
    if (!user) {
      clear();
      return;
    }
    void refresh();
  }, [user, refresh, clear]);

  return { usage, loading, refresh };
}
