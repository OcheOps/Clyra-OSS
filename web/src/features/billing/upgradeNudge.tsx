import { APIError } from '@/types/api';
import { useUpgradeModalStore } from './upgradeModalStore';

// Call from API catch blocks. Returns true if the error was a quota/rate 429
// and was handled with the upgrade modal — callers can then skip their
// generic error-toast to avoid double notifications.
//
// Uses the global UpgradeModal (see GlobalUpgradeModal mounted in AppLayout)
// instead of a toast, because 429 means the user literally cannot proceed —
// a dismissable toast undersells the moment.
export function maybeShowUpgradeNudge(err: unknown): boolean {
  if (err instanceof APIError && err.status === 429) {
    useUpgradeModalStore.getState().openModal({ defaultPlan: 'pro' });
    return true;
  }
  return false;
}
