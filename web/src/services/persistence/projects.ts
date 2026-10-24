import { getDB } from './db';
import type { Project, Diagram } from '@/types/diagram';

export async function listProjects(userId: string): Promise<Project[]> {
  const db = await getDB();
  const all = await db.getAllFromIndex('projects', 'by-updated');
  // Filter by user and sort most recent first
  return (all.filter((p) => p.userId === userId).reverse()) as Project[];
}

export async function getProject(id: string, userId: string): Promise<Project | undefined> {
  const db = await getDB();
  const project = (await db.get('projects', id)) as (Project & { userId?: string }) | undefined;
  // Only return if it belongs to this user
  if (project && project.userId !== userId) return undefined;
  return project;
}

export async function saveProject(project: Project, userId: string): Promise<void> {
  const db = await getDB();
  await db.put('projects', {
    ...project,
    userId,
    updatedAt: Date.now(),
  });
}

export async function createProject(name: string, diagram: Diagram, userId: string): Promise<Project> {
  const now = Date.now();
  const project: Project = {
    id: `proj-${now}-${Math.random().toString(36).slice(2, 8)}`,
    name,
    diagram,
    createdAt: now,
    updatedAt: now,
  };
  const db = await getDB();
  await db.put('projects', { ...project, userId });
  return project;
}

export async function deleteProject(id: string, userId: string): Promise<void> {
  // Verify ownership before deleting
  const project = await getProject(id, userId);
  if (!project) return;
  const db = await getDB();
  await db.delete('projects', id);
}
