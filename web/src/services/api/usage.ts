import { api } from './client';
import type { Plan } from '@/types/auth';

export interface DailyUsage {
  plan: Plan;
  effective_plan: Plan;
  plan_expiry?: string | null;
  daily_limit: number;
  daily_used: number;
  daily_remaining: number;
}

export function getDailyUsage(): Promise<DailyUsage> {
  return api<DailyUsage>('/me/usage', { method: 'GET' });
}
