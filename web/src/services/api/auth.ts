import { api } from './client';
import type {
  AuthResponse,
  InviteCode,
  InviteRequest,
  InviteRequestStatus,
  LoginRequest,
  Plan,
  RegisterRequest,
} from '@/types/auth';

export function login(req: LoginRequest): Promise<AuthResponse> {
  return api<AuthResponse>('/auth/login', {
    method: 'POST',
    body: req,
  });
}

export function register(req: RegisterRequest): Promise<AuthResponse> {
  return api<AuthResponse>('/auth/register', {
    method: 'POST',
    body: req,
  });
}

export function me(): Promise<AuthResponse> {
  return api<AuthResponse>('/auth/me', { method: 'GET' });
}

export function logout(): Promise<void> {
  return api<void>('/auth/logout', {
    method: 'POST',
  });
}

export function forgotPassword(email: string): Promise<{ status: string; message: string }> {
  return api<{ status: string; message: string }>('/auth/forgot-password', {
    method: 'POST',
    body: { email },
  });
}

export function resetPassword(token: string, password: string): Promise<{ status: string }> {
  return api<{ status: string }>('/auth/reset-password', {
    method: 'POST',
    body: { token, password },
  });
}

export function validateInvite(code: string): Promise<{ valid: boolean }> {
  return api<{ valid: boolean }>('/invites/validate', {
    method: 'POST',
    body: { code },
  });
}

export function requestAccess(email: string, note?: string): Promise<{ status: string; id: string }> {
  return api<{ status: string; id: string }>('/invites/request', {
    method: 'POST',
    body: { email, note: note || '' },
  });
}

export function createInvite(
  email?: string,
  plan: Plan = 'free',
): Promise<{ invite: InviteCode; email_sent: boolean }> {
  return api<{ invite: InviteCode; email_sent: boolean }>('/admin/invites', {
    method: 'POST',
    body: { email: email || '', plan },
  });
}

export function listInvites(): Promise<InviteCode[]> {
  return api<InviteCode[]>('/admin/invites', {
    method: 'GET',
  });
}

export function deleteInvite(id: string): Promise<void> {
  return api<void>('/admin/invites', {
    method: 'DELETE',
    body: { id },
  });
}

export function listInviteRequests(status?: InviteRequestStatus): Promise<InviteRequest[]> {
  const qs = status ? `?status=${encodeURIComponent(status)}` : '';
  return api<InviteRequest[]>(`/admin/invite-requests${qs}`, { method: 'GET' });
}

export function approveInviteRequest(
  id: string,
  plan: Plan = 'free',
): Promise<{ invite: InviteCode; email_sent: boolean; email_to: string }> {
  return api(`/admin/invite-requests/${id}/approve`, {
    method: 'POST',
    body: { plan },
  });
}

export function declineInviteRequest(id: string): Promise<{ status: string }> {
  return api(`/admin/invite-requests/${id}/decline`, {
    method: 'POST',
    body: {},
  });
}
