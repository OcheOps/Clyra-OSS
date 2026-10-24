import { useCallback, useEffect, useState } from 'react';
import { useAuthStore } from '@/features/auth/store';
import {
  readStoredBYOKeys,
  storedBYOKeyName,
  writeStoredBYOKeys,
  type BYOKeyBag,
  type BYOProvider,
} from '../lib/byoKeyStorage';

export type { BYOProvider } from '../lib/byoKeyStorage';

/**
 * useBYOKeys manages per-provider API keys stored in localStorage.
 * Keys never leave the browser except in-flight on provider requests.
 */
export function useBYOKeys() {
  const userId = useAuthStore((s) => s.user?.id) ?? '';
  const [keys, setKeys] = useState<BYOKeyBag>(() => readStoredBYOKeys(userId));

  useEffect(() => {
    setKeys(readStoredBYOKeys(userId));
  }, [userId]);

  // Keep tabs in sync — another tab editing keys should propagate.
  useEffect(() => {
    if (typeof window === 'undefined') return;
    const onStorage = (e: StorageEvent) => {
      if (e.key !== storedBYOKeyName(userId)) return;
      setKeys(readStoredBYOKeys(userId));
    };
    window.addEventListener('storage', onStorage);
    return () => window.removeEventListener('storage', onStorage);
  }, [userId]);

  const setKey = useCallback((provider: BYOProvider, key: string) => {
    setKeys((prev) => {
      const next: BYOKeyBag = { ...prev };
      const trimmed = key.trim();
      if (trimmed) next[provider] = trimmed;
      else delete next[provider];
      writeStoredBYOKeys(userId, next);
      return next;
    });
  }, [userId]);

  const clearKey = useCallback((provider: BYOProvider) => {
    setKey(provider, '');
  }, [setKey]);

  const getKey = useCallback((provider: BYOProvider) => keys[provider] ?? '', [keys]);
  const hasKey = useCallback((provider: BYOProvider) => Boolean(keys[provider]), [keys]);

  return { keys, setKey, clearKey, getKey, hasKey };
}
