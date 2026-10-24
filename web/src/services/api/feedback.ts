import { api } from './client';

export type FeedbackCategory = 'general' | 'bug' | 'feature' | 'billing' | 'security' | 'other';

export interface SubmitFeedbackInput {
  name: string;
  email: string;
  category: FeedbackCategory;
  message: string;
}

export interface SubmitFeedbackResponse {
  id: string;
  status: string;
}

// Public endpoint — works for signed-out and signed-in users. The backend
// overrides the email + attaches user_id when there's an active session,
// so the form can't be used to spoof a signed-in user's identity.
export function submitFeedback(input: SubmitFeedbackInput): Promise<SubmitFeedbackResponse> {
  return api<SubmitFeedbackResponse>('/feedback', { method: 'POST', body: input });
}
