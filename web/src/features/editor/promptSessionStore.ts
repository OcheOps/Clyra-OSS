import { create } from 'zustand';

// Shared state for the two prompt entry points (empty-canvas card and
// topbar PromptBar). A generation in either surface counts as the
// "initial" prompt, unlocking the topbar's 2 follow-up slots.
//
// Intentionally NOT persisted: a fresh reload should start a fresh
// conversation. The empty-card dismissal is session-local too — user
// who reloads expects to see the card again.

interface PromptSessionState {
  lastPrompt: string;
  followupCount: number;
  emptyCardDismissed: boolean;
  setLastPrompt: (p: string) => void;
  incrementFollowup: () => void;
  resetSession: () => void;
  dismissEmptyCard: () => void;
  showEmptyCard: () => void;
}

export const usePromptSessionStore = create<PromptSessionState>((set) => ({
  lastPrompt: '',
  followupCount: 0,
  emptyCardDismissed: false,
  setLastPrompt: (p) => set({ lastPrompt: p, followupCount: 0 }),
  incrementFollowup: () => set((s) => ({ followupCount: s.followupCount + 1 })),
  resetSession: () => set({ lastPrompt: '', followupCount: 0 }),
  dismissEmptyCard: () => set({ emptyCardDismissed: true }),
  showEmptyCard: () => set({ emptyCardDismissed: false }),
}));
