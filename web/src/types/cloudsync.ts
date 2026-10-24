import type { Diagram } from './diagram';

export interface CloudProject {
  id: string;
  user_id: string;
  title: string;
  version: number;
  storage: 'db' | 's3';
  source: 'app' | 'mcp';
  byte_size: number;
  preview_url?: string;
  preview_bytes?: number;
  created_at: string;
  updated_at: string;
}

export interface SaveProjectRequest {
  id?: string;
  title: string;
  diagram: Diagram;
  preview_png_b64?: string;
}

export interface LoadProjectResponse {
  project: CloudProject;
  diagram: Diagram;
}
