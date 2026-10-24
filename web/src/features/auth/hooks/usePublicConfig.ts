import { useEffect, useState } from 'react';
import { getPublicConfig, type PublicConfig, type RegistrationMode } from '@/services/api/config';

// Cache the public config for the life of the page load — it's small and
// doesn't change without a backend redeploy, so a single fetch per session
// is enough. Fail closed on network error: invite-only + no OAuth buttons
// shown until the fetch resolves.
let cached: PublicConfig | null = null;

const fallback: PublicConfig = {
  registration_mode: 'invite',
  google_oauth_enabled: false,
  github_oauth_enabled: false,
};

export function usePublicConfig(): {
  mode: RegistrationMode;
  googleEnabled: boolean;
  githubEnabled: boolean;
  loading: boolean;
} {
  const [cfg, setCfg] = useState<PublicConfig>(cached ?? fallback);
  const [loading, setLoading] = useState<boolean>(cached === null);

  useEffect(() => {
    if (cached !== null) return;
    let cancelled = false;
    getPublicConfig()
      .then((next) => {
        if (cancelled) return;
        cached = next;
        setCfg(next);
      })
      .catch(() => {
        // Network blip — stay on the safe fallback.
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return {
    mode: cfg.registration_mode,
    googleEnabled: cfg.google_oauth_enabled,
    githubEnabled: cfg.github_oauth_enabled,
    loading,
  };
}
