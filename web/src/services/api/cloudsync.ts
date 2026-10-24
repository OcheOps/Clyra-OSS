import { api } from './client';
import type { CloudProject, LoadProjectResponse, SaveProjectRequest } from '@/types/cloudsync';

export function listCloudProjects(): Promise<CloudProject[]> {
  return api<CloudProject[]>('/sync/projects', { method: 'GET' });
}

export function loadCloudProject(id: string): Promise<LoadProjectResponse> {
  return api<LoadProjectResponse>(`/sync/projects/${encodeURIComponent(id)}`, { method: 'GET' });
}

export function saveCloudProject(req: SaveProjectRequest): Promise<CloudProject> {
  const id = req.id || 'new';
  return api<CloudProject>(`/sync/projects/${encodeURIComponent(id)}`, {
    method: 'PUT',
    body: req,
  });
}

export function deleteCloudProject(id: string): Promise<void> {
  return api<void>(`/sync/projects/${encodeURIComponent(id)}`, { method: 'DELETE' });
}
