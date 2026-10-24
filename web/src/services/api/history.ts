import { api } from './client';
import type { LoadProjectResponse } from '@/types/cloudsync';

export interface ProjectVersion {
  id: string;
  title: string;
  byte_size: number;
  created_at: string;
}

export function listVersions(projectID: string): Promise<ProjectVersion[]> {
  return api<ProjectVersion[]>(
    `/sync/projects/${encodeURIComponent(projectID)}/versions`,
    { method: 'GET' },
  );
}

export function restoreVersion(
  projectID: string,
  versionID: string,
): Promise<LoadProjectResponse> {
  return api<LoadProjectResponse>(
    `/sync/projects/${encodeURIComponent(projectID)}/versions/${encodeURIComponent(versionID)}/restore`,
    { method: 'POST' },
  );
}
