import { create } from 'zustand';

// A shared store for "feature requires Pro/BYO" prompts. Any surface that
// hits a 402 from the backend calls `show(reason)` and the global
// UpgradeForSyncModal renders on top.
//
// Keeping this minimal: a single open-reason string, no queue. Two
// near-simultaneous 402s just collapse to the same modal.

export type UpgradeReason = 'cloud_sync' | 'iac_export' | 'generic';

interface UpgradePromptState {
  reason: UpgradeReason | null;
  show: (reason: UpgradeReason) => void;
  hide: () => void;
}

export const useUpgradePromptStore = create<UpgradePromptState>((set) => ({
  reason: null,
  show: (reason) => set({ reason }),
  hide: () => set({ reason: null }),
}));
