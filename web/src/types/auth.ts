export type Plan = 'free' | 'pro' | 'byo';

export interface User {
  id: string;
  email: string;
  role: string;
  // Raw plan from the users table — kept post-expiry for display context.
  plan: Plan;
  plan_expiry?: string | null;
  // Effective plan the backend actually enforces right now (expired = free).
  // Gate UI badges and disable/enable checks off this, not `plan`.
  effective_plan: Plan;
}

export interface AuthResponse {
  user: User;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  invite_code: string;
}

export interface InviteCode {
  id: string;
  code: string;
  created_by: string;
  used_by?: string;
  used_at?: string;
  expires_at?: string;
  plan: Plan;
  created_at: string;
}

export type InviteRequestStatus = 'pending' | 'approved' | 'declined';

export interface InviteRequest {
  id: string;
  email: string;
  note?: string;
  status: InviteRequestStatus;
  handled_by?: string;
  handled_at?: string;
  created_at: string;
}
