import { api } from './client';
import type { PricingCatalog } from '@/types/pricing';

/**
 * Fetches the pricing catalog. An optional ISO-3166-1 alpha-2 country code
 * is passed as `?region=XX` so the server honors it over its default.
 * This lets the browser's own IP check (via services/geo.ts) beat localhost.
 */
export function getPricing(region?: string): Promise<PricingCatalog> {
  const query = region ? `?region=${encodeURIComponent(region)}` : '';
  return api<PricingCatalog>(`/pricing${query}`, { method: 'GET' });
}
