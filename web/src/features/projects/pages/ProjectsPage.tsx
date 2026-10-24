import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { AppLayout } from '@/components/layout/AppLayout';
import { PageHeader } from '@/components/layout/PageHeader';
import {
  deleteCloudProject,
  listCloudProjects,
  loadCloudProject,
} from '@/services/api/cloudsync';
import type { CloudProject } from '@/types/cloudsync';
import { APIError } from '@/types/api';
import { useUpgradePromptStore } from '@/features/billing/upgradePromptStore';
import { useEditorStore } from '@/features/editor/store';

function timeAgo(ts: string): string {
  const diff = Math.floor((Date.now() - new Date(ts).getTime()) / 1000);
  if (diff < 60) return 'just now';
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / 1024 / 1024).toFixed(2)} MB`;
}

/**
 * ProjectsPage — v2 restyle.
 *
 * Cloud-saved projects only (local IndexedDB lives on the Dashboard
 * empty state and on the editor's autosave path). Renders each project
 * as a card with the preview image as a 16:9 thumbnail when available,
 * and meta data — title, last-updated, byte size, storage backend — in
 * the v2 mono small-caps voice underneath.
 */
export function ProjectsPage() {
  const [projects, setProjects] = useState<CloudProject[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const loadDiagram = useEditorStore((s) => s.loadDiagram);
  const clearCanvas = useEditorStore((s) => s.clearCanvas);

  const handleNewDiagram = useCallback(() => {
    clearCanvas();
    navigate('/editor');
  }, [clearCanvas, navigate]);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const list = await listCloudProjects();
      setProjects(list);
    } catch (err) {
      if (err instanceof APIError && err.status === 402) {
        useUpgradePromptStore.getState().show('cloud_sync');
      } else {
        toast.error('Could not load your cloud projects');
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const handleOpen = useCallback(
    async (p: CloudProject) => {
      try {
        const resp = await loadCloudProject(p.id);
        loadDiagram(resp.diagram);
        // Carry the project identity into the editor — useAutosave reads
        // these on mount so the title bar opens on the correct name
        // instead of "Untitled Diagram", and so subsequent saves update
        // *this* project instead of forking a duplicate cloud entry.
        useEditorStore.getState().setPendingProject(resp.project.id, resp.project.title);
        navigate('/editor');
      } catch {
        toast.error('Failed to load project');
      }
    },
    [loadDiagram, navigate],
  );

  const handleDelete = useCallback(async (p: CloudProject) => {
    if (!confirm(`Delete "${p.title}"? This removes it from the cloud.`)) return;
    try {
      await deleteCloudProject(p.id);
      setProjects((prev) => prev.filter((r) => r.id !== p.id));
      toast.success('Deleted');
    } catch {
      toast.error('Failed to delete');
    }
  }, []);

  return (
    <AppLayout>
      <main className="relative flex-1 overflow-y-auto">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 opacity-60"
          style={{
            backgroundImage:
              'radial-gradient(circle at 1px 1px, rgba(148,163,184,0.05) 1px, transparent 0)',
            backgroundSize: '24px 24px',
          }}
        />

        <div className="relative mx-auto max-w-5xl px-6 py-12">
          <PageHeader
            eyebrow="Cloud sync"
            title="My projects"
            description="Diagrams saved to your account. Local IndexedDB is still the default."
            backTo="/dashboard"
            actions={
              <button
                onClick={handleNewDiagram}
                className="rounded-full border border-c2-border bg-c2-surface/40 px-4 py-2 text-tiny font-medium text-c2-text transition-colors hover:border-c2-border-strong hover:bg-c2-surface"
              >
                New diagram
              </button>
            }
          />

          {loading ? (
            <div className="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-2">
              {[0, 1, 2, 3].map((i) => (
                <div
                  key={i}
                  className="h-[160px] animate-pulse rounded-2xl border border-c2-border bg-c2-surface/40"
                />
              ))}
            </div>
          ) : projects.length === 0 ? (
            <EmptyState />
          ) : (
            <div className="grid gap-4 sm:grid-cols-2">
              {projects.map((p) => (
                <CloudProjectCard
                  key={p.id}
                  project={p}
                  onOpen={() => handleOpen(p)}
                  onDelete={() => handleDelete(p)}
                />
              ))}
            </div>
          )}
        </div>
      </main>
    </AppLayout>
  );

  function EmptyState() {
    return (
      <div className="relative mt-8 overflow-hidden rounded-2xl border border-dashed border-c2-border bg-c2-surface/40 p-12 text-center">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0"
          style={{
            background:
              'radial-gradient(circle at 50% 0%, rgba(6,182,212,0.06) 0%, transparent 60%)',
          }}
        />
        <div className="relative">
          <div className="mx-auto mb-5 flex h-12 w-12 items-center justify-center rounded-full border border-c2-border bg-c2-bg/40 text-c2-text-faint">
            <svg
              className="h-5 w-5"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.75"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 00-9.78 1.292A4 4 0 003 15z" />
            </svg>
          </div>
          <h3 className="font-display text-lg font-semibold text-c2-text">
            No cloud-saved projects yet
          </h3>
          <p className="mx-auto mt-2 max-w-sm text-[13px] leading-relaxed text-c2-text-muted">
            Use{' '}
            <span className="font-mono text-[12px] text-c2-text">Save to cloud</span>{' '}
            in the editor to keep a snapshot accessible from any device.
          </p>
        </div>
      </div>
    );
  }
}

function CloudProjectCard({
  project,
  onOpen,
  onDelete,
}: {
  project: CloudProject;
  onOpen: () => void;
  onDelete: () => void;
}) {
  return (
    <article className="group relative overflow-hidden rounded-2xl border border-c2-border bg-c2-surface/60 transition-all duration-200 hover:border-c2-primary/40 hover:bg-c2-surface hover:shadow-c2">
      {project.preview_url && (
        <button
          type="button"
          onClick={onOpen}
          className="block aspect-[16/9] w-full overflow-hidden border-b border-c2-hairline bg-c2-bg"
        >
          <img
            src={project.preview_url}
            alt={project.title}
            className="h-full w-full object-cover opacity-90 transition-opacity group-hover:opacity-100"
          />
        </button>
      )}
      <div className="p-5">
        <div className="flex items-start justify-between gap-3">
          <button
            type="button"
            onClick={onOpen}
            className="min-w-0 flex-1 text-left"
          >
            <h3 className="truncate font-display text-[15px] font-semibold tracking-[-0.005em] text-c2-text transition-colors group-hover:text-c2-primary">
              {project.title}
            </h3>
            {project.source === 'mcp' && (
              <span className="mt-2 inline-flex rounded-full border border-c2-primary/30 bg-c2-primary-soft px-2 py-0.5 font-mono text-[9px] uppercase tracking-[0.14em] text-c2-primary">
                Created via MCP
              </span>
            )}
            <p className="mt-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
              Updated {timeAgo(project.updated_at)} · {formatBytes(project.byte_size)} ·{' '}
              {project.storage === 's3' ? 'cloud' : 'server'}
            </p>
          </button>
          <div className="flex shrink-0 items-center gap-2">
            <button
              onClick={onOpen}
              className="rounded-full bg-c2-primary-soft px-3 py-1 text-tiny font-medium text-c2-primary transition-colors hover:bg-c2-primary/25"
            >
              Open
            </button>
            <button
              onClick={onDelete}
              className="rounded-md px-2 py-1 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint transition-colors hover:text-c2-danger"
              aria-label="Delete project"
            >
              Delete
            </button>
          </div>
        </div>
      </div>
    </article>
  );
}
