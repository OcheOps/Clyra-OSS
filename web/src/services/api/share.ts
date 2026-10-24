import { api } from './client';
import type { CloudProject } from '@/types/cloudsync';
import type { Diagram } from '@/types/diagram';

export interface ShareToken {
  token: string;
  project_id: string;
  created_at: string;
}

export function createShare(projectID: string): Promise<ShareToken> {
  return api<ShareToken>(
    `/sync/projects/${encodeURIComponent(projectID)}/share`,
    { method: 'POST' },
  );
}

export function listShares(projectID: string): Promise<ShareToken[]> {
  return api<ShareToken[]>(
    `/sync/projects/${encodeURIComponent(projectID)}/share`,
    { method: 'GET' },
  );
}

export function revokeShare(projectID: string, token: string): Promise<void> {
  return api<void>(
    `/sync/projects/${encodeURIComponent(projectID)}/share/${encodeURIComponent(token)}`,
    { method: 'DELETE' },
  );
}

// Public — called by the unauth viewer page.
export function loadSharedProject(token: string): Promise<{ project: CloudProject; diagram: Diagram }> {
  return api<{ project: CloudProject; diagram: Diagram }>(
    `/public/share/${encodeURIComponent(token)}`,
    { method: 'GET' },
  );
}

export function shareURL(token: string): string {
  return `${window.location.origin}/share/${token}`;
}
