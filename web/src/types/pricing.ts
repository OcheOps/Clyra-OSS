export type PricingPlan = 'free' | 'pro' | 'byo';

export interface PricingPlanCard {
  plan: PricingPlan;
  name: string;
  tagline: string;
  amount_minor: number;
  amount_major: string;
  currency: string;
  currency_symbol: string;
  period: string;
  daily_limit: number;
  features: string[];
  cta: string;
  recommended?: boolean;
  requires_keys?: boolean;
}

export interface PricingCatalog {
  region: string;
  country?: string;
  plans: PricingPlanCard[];
}
