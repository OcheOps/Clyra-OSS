export type BYOProvider = 'openai' | 'anthropic' | 'gemini';
export type BYOKeyBag = Partial<Record<BYOProvider, string>>;

const LEGACY_STORAGE_KEY = 'clyra-byo-keys';

function storageKey(userId: string): string {
  return `clyra-byo-keys:${userId || 'anonymous'}`;
}

export function readStoredBYOKeys(userId: string): BYOKeyBag {
  if (typeof window === 'undefined') return {};
  try {
    const raw = window.localStorage.getItem(storageKey(userId));
    if (!raw) return {};
    const parsed: unknown = JSON.parse(raw);
    if (parsed && typeof parsed === 'object') return parsed as BYOKeyBag;
  } catch {
    // Ignore corrupt browser state.
  }
  return {};
}

export function writeStoredBYOKeys(userId: string, bag: BYOKeyBag) {
  if (typeof window === 'undefined') return;
  window.localStorage.setItem(storageKey(userId), JSON.stringify(bag));
}

export function storedBYOKeyName(userId: string): string {
  return storageKey(userId);
}

export function clearStoredBYOKeys(userId: string) {
  if (typeof window === 'undefined') return;
  window.localStorage.removeItem(storageKey(userId));
  window.localStorage.removeItem(LEGACY_STORAGE_KEY);
}
