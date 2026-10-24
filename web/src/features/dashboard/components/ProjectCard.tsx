import { Link } from 'react-router-dom';
import type { Project } from '@/types/diagram';

interface ProjectCardProps {
  project: Project;
  onDelete: (id: string) => void;
}

function timeAgo(ts: number): string {
  const diff = Math.floor((Date.now() - ts) / 1000);
  if (diff < 60) return 'just now';
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

/**
 * ProjectCard — v2 restyle.
 *
 * Whole card is a navigable Link to the editor; the delete button is
 * absolute-positioned in the corner with a fade-in on hover so it
 * doesn't compete with the project name visually. A small SVG
 * "node sparkline" in the corner gives a sense of complexity at a
 * glance — three dots scaled to node tier with a connecting line whose
 * opacity reflects edge density.
 */
export function ProjectCard({ project, onDelete }: ProjectCardProps) {
  const nodeCount = project.diagram.nodes.length;
  const edgeCount = project.diagram.edges.length;

  return (
    <div className="group relative">
      <Link
        to={`/editor/${project.id}`}
        className="block overflow-hidden rounded-xl border border-c2-border bg-c2-surface/60 p-5 transition-all duration-200 hover:border-c2-primary/40 hover:bg-c2-surface hover:shadow-c2"
      >
        <div className="flex items-start justify-between gap-3">
          <h3 className="min-w-0 flex-1 truncate font-display text-[15px] font-semibold tracking-[-0.005em] text-c2-text transition-colors group-hover:text-c2-primary">
            {project.name}
          </h3>
          <NodeSparkline nodes={nodeCount} edges={edgeCount} />
        </div>

        <div className="mt-4 flex items-center gap-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
          <span>{nodeCount} nodes</span>
          <span>·</span>
          <span>{edgeCount} edges</span>
          <span>·</span>
          <span>Updated {timeAgo(project.updatedAt)}</span>
        </div>
      </Link>

      {/* Delete — absolute so it doesn't shift card layout; opacity-0 by
          default so the resting state stays clean. */}
      <button
        type="button"
        onClick={(e) => {
          e.preventDefault();
          onDelete(project.id);
        }}
        className="absolute right-3 top-3 rounded-md border border-transparent px-2 py-1 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint opacity-0 transition-all duration-150 hover:border-c2-danger/40 hover:bg-c2-danger/10 hover:text-c2-danger group-hover:opacity-100 focus-visible:opacity-100"
        aria-label="Delete project"
      >
        Delete
      </button>
    </div>
  );
}

/**
 * NodeSparkline — a tiny visual for the project's diagram complexity
 * without rendering the actual diagram. Three dots arranged in a row,
 * with their visual weight scaled to node count, plus a single
 * connecting line whose stroke opacity reflects edge density.
 *
 * For a 0-node project all three dots stay hollow; for a >15-node
 * project all three fill solid. The point is glanceability, not
 * precision.
 */
function NodeSparkline({ nodes, edges }: { nodes: number; edges: number }) {
  const tier =
    nodes === 0 ? 0
    : nodes < 5  ? 1
    : nodes < 15 ? 2
    : 3;
  const edgeOpacity = Math.min(1, 0.2 + edges / 30);

  return (
    <svg className="h-8 w-12 shrink-0" viewBox="0 0 48 24" fill="none" aria-hidden>
      <line
        x1="6" y1="12" x2="42" y2="12"
        stroke="#06B6D4"
        strokeOpacity={edgeOpacity}
        strokeWidth="1"
      />
      {[6, 24, 42].map((cx, i) => {
        const filled = tier > i;
        return (
          <circle
            key={cx}
            cx={cx}
            cy={12}
            r={filled ? 4 : 3}
            fill={filled ? '#06B6D4' : '#0F172A'}
            stroke="#06B6D4"
            strokeWidth={filled ? 0 : 1.5}
            strokeOpacity={0.7}
          />
        );
      })}
    </svg>
  );
}
