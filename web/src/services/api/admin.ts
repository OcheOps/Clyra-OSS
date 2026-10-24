import { api } from './client';

// Admin surface. All endpoints require role=admin on the session; the
// backend enforces, we don't double-check here.

export interface AdminUser {
  id: string;
  email: string;
  role: 'user' | 'admin';
  plan: 'free' | 'pro' | 'byo';
  plan_expiry?: string | null;
  created_at: string;
  updated_at: string;
}

export interface UsersListResponse {
  users: AdminUser[];
  total: number;
}

export interface AdminMetrics {
  total_users: number;
  users_by_plan: Record<string, number>;
  signups_last_7d: number;
  signups_last_30d: number;
  activation_funnel: Record<string, number>;
}

export function listAdminUsers(params: { q?: string; limit?: number; offset?: number } = {}): Promise<UsersListResponse> {
  const qs = new URLSearchParams();
  if (params.q) qs.set('q', params.q);
  if (params.limit) qs.set('limit', String(params.limit));
  if (params.offset) qs.set('offset', String(params.offset));
  const suffix = qs.toString() ? `?${qs}` : '';
  return api<UsersListResponse>(`/admin/users${suffix}`, { method: 'GET' });
}

export function getAdminMetrics(): Promise<AdminMetrics> {
  return api<AdminMetrics>('/admin/metrics', { method: 'GET' });
}

export function setUserRole(userId: string, role: 'user' | 'admin'): Promise<void> {
  return api<void>(`/admin/users/${encodeURIComponent(userId)}/role`, {
    method: 'POST',
    body: { role },
  });
}

export function grantUserPlan(
  userId: string,
  plan: 'free' | 'pro' | 'byo',
  opts: { expires_at?: string; note?: string } = {},
): Promise<void> {
  return api<void>(`/admin/users/${encodeURIComponent(userId)}/plan`, {
    method: 'POST',
    body: { plan, ...opts },
  });
}

export function emailUser(userId: string, subject: string, body: string): Promise<void> {
  return api<void>(`/admin/users/${encodeURIComponent(userId)}/email`, {
    method: 'POST',
    body: { subject, body },
  });
}

// Audit log read-side. Filters mirror the backend.
export interface AuditEvent {
  id: string;
  occurred_at: string;
  event: string;
  actor_user_id?: string | null;
  target_id?: string | null;
  ip?: string | null;
  user_agent?: string | null;
  metadata: Record<string, unknown>;
}

export interface AuditListResponse {
  items: AuditEvent[];
  total: number;
}

export function listAuditEvents(params: { event?: string; actor?: string; limit?: number; offset?: number } = {}): Promise<AuditListResponse> {
  const qs = new URLSearchParams();
  if (params.event) qs.set('event', params.event);
  if (params.actor) qs.set('actor', params.actor);
  if (params.limit) qs.set('limit', String(params.limit));
  if (params.offset) qs.set('offset', String(params.offset));
  const suffix = qs.toString() ? `?${qs}` : '';
  return api<AuditListResponse>(`/admin/audit${suffix}`, { method: 'GET' });
}

// Feedback inbox.
export interface FeedbackItem {
  id: string;
  user_id?: string | null;
  name: string;
  email: string;
  category: string;
  message: string;
  status: 'new' | 'open' | 'closed';
  ip?: string | null;
  user_agent?: string | null;
  created_at: string;
  updated_at: string;
}

export interface FeedbackListResponse {
  items: FeedbackItem[];
  total: number;
}

export function listFeedback(params: { status?: string; limit?: number; offset?: number } = {}): Promise<FeedbackListResponse> {
  const qs = new URLSearchParams();
  if (params.status) qs.set('status', params.status);
  if (params.limit) qs.set('limit', String(params.limit));
  if (params.offset) qs.set('offset', String(params.offset));
  const suffix = qs.toString() ? `?${qs}` : '';
  return api<FeedbackListResponse>(`/admin/feedback${suffix}`, { method: 'GET' });
}

export function setFeedbackStatus(id: string, status: 'new' | 'open' | 'closed'): Promise<void> {
  return api<void>(`/admin/feedback/${encodeURIComponent(id)}/status`, {
    method: 'POST',
    body: { status },
  });
}
