import { create } from 'zustand';
import { persist } from 'zustand/middleware';

// Whether the floating Insights panel is shown at all. Separate from
// its expanded/collapsed state, which lives inside the panel. A user
// who dismisses via the × shouldn't see it pop back on next reload.

interface InsightsState {
  visible: boolean;
  show: () => void;
  hide: () => void;
  toggle: () => void;
}

export const useInsightsStore = create<InsightsState>()(
  persist(
    (set) => ({
      visible: true,
      show: () => set({ visible: true }),
      hide: () => set({ visible: false }),
      toggle: () => set((s) => ({ visible: !s.visible })),
    }),
    { name: 'clyra-insights-visible' },
  ),
);
