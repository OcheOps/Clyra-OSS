import { useCallback, useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AppLayout } from '@/components/layout/AppLayout';
import { useAuthStore } from '@/features/auth/store';
import { useEditorStore } from '@/features/editor/store';
import { ProjectCard } from '../components/ProjectCard';
import { TemplateGallery } from '@/features/templates/TemplateGallery';
import * as projectsDB from '@/services/persistence/projects';
import type { Project } from '@/types/diagram';

/**
 * Dashboard — v2 signed-in landing.
 *
 * Reframed from "list of projects with a New Diagram button at the top"
 * into a proper signed-in landing page: a welcome eyebrow + first-name
 * greeting, the primary CTA (always visible), then a recent-projects
 * strip (top 6, capped) with a link off to /projects for the full list,
 * and templates below.
 *
 * Atmospheric background matches the marketing surfaces (dot grid + a
 * subtle cyan haze) so users feel they're inside the same product, not
 * a different app once they log in.
 */
export function DashboardPage() {
  const user = useAuthStore((s) => s.user);
  const navigate = useNavigate();
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);

  const userId = user?.id ?? '';

  // "New diagram" / "blank canvas" entries must clear the store before
  // navigating; the editor no longer auto-clears on mount (that used to
  // wipe pre-loaded cloud projects).
  const startBlankDiagram = useCallback(() => {
    useEditorStore.getState().clearCanvas();
    navigate('/editor');
  }, [navigate]);

  const loadProjects = useCallback(async () => {
    if (!userId) return;
    setLoading(true);
    const list = await projectsDB.listProjects(userId);
    setProjects(list);
    setLoading(false);
  }, [userId]);

  useEffect(() => {
    loadProjects();
  }, [loadProjects]);

  const handleDelete = useCallback(
    async (id: string) => {
      await projectsDB.deleteProject(id, userId);
      setProjects((prev) => prev.filter((p) => p.id !== id));
    },
    [userId],
  );

  const greetingName = user?.email?.split('@')[0] ?? 'there';
  const recent = projects.slice(0, 6);
  const empty = !loading && projects.length === 0;

  return (
    <AppLayout>
      <main className="relative flex-1 overflow-y-auto">
        {/* Atmospheric layer — soft cyan glow at the top to make the
            page feel like a "home" rather than a list page. Quieter
            than the landing's; we're inside the product now. */}
        <div
          aria-hidden
          className="pointer-events-none absolute inset-x-0 top-0 h-[400px]"
          style={{
            background:
              'radial-gradient(ellipse at 30% 0%, rgba(6,182,212,0.12) 0%, transparent 60%)',
          }}
        />
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 opacity-60"
          style={{
            backgroundImage:
              'radial-gradient(circle at 1px 1px, rgba(148,163,184,0.05) 1px, transparent 0)',
            backgroundSize: '24px 24px',
          }}
        />

        <div className="relative mx-auto max-w-6xl px-6 py-12">
          <Greeting name={greetingName} onNewDiagram={startBlankDiagram} />

          {loading ? (
            <SkeletonStrip />
          ) : empty ? (
            <EmptyState onNewDiagram={startBlankDiagram} />
          ) : (
            <RecentProjects
              projects={recent}
              total={projects.length}
              onDelete={handleDelete}
            />
          )}

          <div className="mt-16">
            <SectionEyebrow>
              {empty ? 'Start from a template' : 'Or start from a template'}
            </SectionEyebrow>
            <div className="mt-6">
              <TemplateGallery title="" />
            </div>
          </div>
        </div>
      </main>
    </AppLayout>
  );
}

// ────────────────────────────────────────────────────────────────────
// Greeting — top section. Eyebrow + headline + sub + primary CTA.

function Greeting({
  name,
  onNewDiagram,
}: {
  name: string;
  onNewDiagram: () => void;
}) {
  return (
    <section className="flex flex-col gap-8 md:flex-row md:items-end md:justify-between">
      <div>
        <div className="flex items-center gap-2">
          <span className="inline-block h-1.5 w-1.5 rounded-full bg-c2-primary shadow-[0_0_8px_rgba(6,182,212,0.6)]" />
          <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
            Workbench
          </span>
        </div>
        <h1 className="mt-4 font-display text-3xl font-semibold leading-[1.05] tracking-[-0.025em] text-c2-text md:text-[40px]">
          Welcome back,
          <br />
          <span className="text-c2-text-muted">{name}.</span>
        </h1>
      </div>
      <div className="flex items-center gap-2">
        <button
          onClick={onNewDiagram}
          className="group inline-flex items-center gap-2 rounded-full bg-c2-primary px-5 py-2.5 text-sm font-medium text-c2-bg shadow-c2-glow transition-all hover:bg-c2-primary-hover hover:shadow-[0_0_40px_-6px_rgba(6,182,212,0.55)]"
        >
          New diagram
          <span className="transition-transform group-hover:translate-x-0.5">→</span>
        </button>
        <Link
          to="/projects"
          className="rounded-full border border-c2-border bg-c2-surface/40 px-4 py-2.5 text-sm font-medium text-c2-text transition-colors hover:border-c2-border-strong hover:bg-c2-surface"
        >
          All projects
        </Link>
      </div>
    </section>
  );
}

// ────────────────────────────────────────────────────────────────────
// Recent projects — capped at 6 with a footer link to the full list.

function RecentProjects({
  projects,
  total,
  onDelete,
}: {
  projects: Project[];
  total: number;
  onDelete: (id: string) => void;
}) {
  const moreCount = Math.max(0, total - projects.length);
  return (
    <section className="mt-14">
      <div className="flex items-baseline justify-between">
        <SectionEyebrow>Recent projects</SectionEyebrow>
        {moreCount > 0 && (
          <Link
            to="/projects"
            className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint transition-colors hover:text-c2-text"
          >
            + {moreCount} more →
          </Link>
        )}
      </div>
      <div className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {projects.map((project) => (
          <ProjectCard key={project.id} project={project} onDelete={onDelete} />
        ))}
      </div>
    </section>
  );
}

// ────────────────────────────────────────────────────────────────────
// Empty state — friendly, encourages templates or blank canvas.

function EmptyState({ onNewDiagram }: { onNewDiagram: () => void }) {
  return (
    <section className="mt-14">
      <div className="relative overflow-hidden rounded-2xl border border-dashed border-c2-border bg-c2-surface/40 px-8 py-14 text-center">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0"
          style={{
            background:
              'radial-gradient(circle at 50% 0%, rgba(6,182,212,0.08) 0%, transparent 60%)',
          }}
        />
        <div className="relative">
          <SectionEyebrow center>Empty workbench</SectionEyebrow>
          <h2 className="mt-5 font-display text-2xl font-semibold tracking-[-0.02em] text-c2-text">
            Nothing here yet.
          </h2>
          <p className="mx-auto mt-3 max-w-md text-[14px] leading-relaxed text-c2-text-muted">
            Start from a template below, or describe what you're building in
            plain English on a blank canvas.
          </p>
          <button
            onClick={onNewDiagram}
            className="mt-7 inline-flex items-center gap-2 rounded-full border border-c2-border bg-c2-bg/40 px-5 py-2.5 text-sm font-medium text-c2-text transition-colors hover:border-c2-border-strong hover:bg-c2-surface"
          >
            Start from a blank canvas
            <span>→</span>
          </button>
        </div>
      </div>
    </section>
  );
}

// ────────────────────────────────────────────────────────────────────
// Loading state — three pulsing skeletons in the recent-projects shape.

function SkeletonStrip() {
  return (
    <section className="mt-14">
      <SectionEyebrow>Recent projects</SectionEyebrow>
      <div className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {[0, 1, 2].map((i) => (
          <div
            key={i}
            className="h-[110px] animate-pulse rounded-xl border border-c2-border bg-c2-surface/40"
          />
        ))}
      </div>
    </section>
  );
}

// ────────────────────────────────────────────────────────────────────
// SectionEyebrow — small consistent header for each subsection.

function SectionEyebrow({
  children,
  center,
}: {
  children: React.ReactNode;
  center?: boolean;
}) {
  return (
    <div className={`flex items-center gap-2 ${center ? 'justify-center' : ''}`}>
      <span className="inline-block h-1 w-1 rounded-full bg-c2-primary" />
      <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
        {children}
      </span>
    </div>
  );
}
