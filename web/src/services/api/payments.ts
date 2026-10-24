import { api } from './client';
import type {
  InitializePaymentRequest,
  InitializePaymentResponse,
  Payment,
} from '@/types/payments';

export function initializePayment(
  req: InitializePaymentRequest,
): Promise<InitializePaymentResponse> {
  return api<InitializePaymentResponse>('/payments/initialize', {
    method: 'POST',
    body: req,
  });
}

// Polar appends ?checkout_id=<uuid> to our success_url after payment.
// Passing it through lets the backend actively verify with Polar's API
// instead of waiting for the webhook — essential in local dev where
// webhooks can't reach localhost without a tunnel.
export function verifyPayment(reference: string, checkoutId?: string): Promise<Payment> {
  const qs = checkoutId ? `?checkout_id=${encodeURIComponent(checkoutId)}` : '';
  return api<Payment>(`/payments/verify/${encodeURIComponent(reference)}${qs}`, {
    method: 'GET',
  });
}

export function listPayments(): Promise<Payment[]> {
  return api<Payment[]>('/payments', { method: 'GET' });
}

export interface CancelPlanResult {
  refunded: boolean;
  refund_ref?: string;
  prev_plan: string;
}

export function cancelPlan(): Promise<CancelPlanResult> {
  return api<CancelPlanResult>('/payments/cancel', {
    method: 'POST',
    body: {},
  });
}
