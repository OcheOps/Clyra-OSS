import type { Plan } from './auth';

export type PaymentStatus = 'pending' | 'paid' | 'failed' | 'refunded';

export interface Payment {
  id: string;
  user_id: string;
  provider: string;
  provider_ref: string;
  amount_minor: number;
  currency: string;
  region: string;
  plan: Plan;
  period_days: number;
  status: PaymentStatus;
  created_at: string;
  verified_at?: string | null;
}

export interface InitializePaymentResponse {
  authorization_url: string;
  reference: string;
  // Provider's native ID (Polar's checkout UUID). Absent for Paystack since
  // its reference and ID are the same thing.
  external_id?: string;
  provider?: string;
  mock?: boolean;
}

export interface InitializePaymentRequest {
  plan: Exclude<Plan, 'free'>;
  region?: string;
}
