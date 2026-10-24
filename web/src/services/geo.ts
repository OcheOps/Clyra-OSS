/**
 * Client-side country detection. Used only to give the server a hint for the
 * pricing endpoint — the source of truth is still the server's geo detector.
 *
 * In local dev the server can't see the user's real IP (loopback), so this
 * client-side call lets a VPN / proxy change the region correctly.
 *
 * No tokens, no PII, no auth — this service returns `{ip, country}` for the
 * caller's public IP only. If it fails (offline, ad-blocker, etc.) we just
 * return "" and let the server decide.
 */

const CACHE_KEY = 'clyra-client-country';
const CACHE_TTL_MS = 30 * 60 * 1000; // 30 min

interface CacheEntry {
  country: string;
  expiresAt: number;
}

function readCache(): string | null {
  try {
    const raw = localStorage.getItem(CACHE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as CacheEntry;
    if (!parsed.country || Date.now() > parsed.expiresAt) return null;
    return parsed.country;
  } catch {
    return null;
  }
}

function writeCache(country: string) {
  try {
    const entry: CacheEntry = { country, expiresAt: Date.now() + CACHE_TTL_MS };
    localStorage.setItem(CACHE_KEY, JSON.stringify(entry));
  } catch {
    // no-op
  }
}

export async function detectClientCountry(signal?: AbortSignal): Promise<string> {
  const cached = readCache();
  if (cached) return cached;

  try {
    const res = await fetch('https://api.country.is/', { signal });
    if (!res.ok) return '';
    const data = (await res.json()) as { country?: string };
    const country = (data.country ?? '').toUpperCase();
    if (country) writeCache(country);
    return country;
  } catch {
    return '';
  }
}

export function clearCountryCache() {
  try {
    localStorage.removeItem(CACHE_KEY);
  } catch {
    // no-op
  }
}
