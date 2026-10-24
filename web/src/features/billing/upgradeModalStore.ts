import { create } from 'zustand';
import type { PricingPlan } from '@/types/pricing';

interface UpgradeModalState {
  open: boolean;
  defaultPlan: Exclude<PricingPlan, 'free'>;
  headline?: string;
  subline?: string;
  openModal: (opts?: {
    defaultPlan?: Exclude<PricingPlan, 'free'>;
    headline?: string;
    subline?: string;
  }) => void;
  closeModal: () => void;
}

// Global modal store so any feature can trigger the upgrade flow without
// threading props. Mount <GlobalUpgradeModal /> once at the app shell.
export const useUpgradeModalStore = create<UpgradeModalState>((set) => ({
  open: false,
  defaultPlan: 'pro',
  openModal: (opts) =>
    set({
      open: true,
      defaultPlan: opts?.defaultPlan ?? 'pro',
      headline: opts?.headline,
      subline: opts?.subline,
    }),
  closeModal: () => set({ open: false }),
}));
