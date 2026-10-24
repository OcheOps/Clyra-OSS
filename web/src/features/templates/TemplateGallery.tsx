import { useNavigate } from 'react-router-dom';
import { TEMPLATES, type Template } from './data';
import { useEditorStore } from '@/features/editor/store';

/**
 * TemplateGallery — v2 restyle.
 *
 * Surfaces starter architectures on the dashboard. Click → loads the
 * diagram into the editor store and navigates to /editor with a
 * preserveCanvas signal. The prompt drawer is seeded with the
 * template's description so follow-up tweaks feel natural ("add a CDN",
 * "swap ECS for Lambda").
 *
 * Visual: dark surface cards with a categorical tag pill in the corner
 * (cyan / accent / info / warning) and a node/edge count in mono small
 * caps. The "Open →" CTA picks up the cyan link voice we use across
 * the rest of v2.
 */

interface Props {
  title?: string;
}

export function TemplateGallery({ title = 'Start from a template' }: Props) {
  const loadDiagram = useEditorStore((s) => s.loadDiagram);
  const navigate = useNavigate();

  const open = (tpl: Template) => {
    loadDiagram(tpl.diagram);
    try {
      sessionStorage.setItem('clyra-initial-prompt', tpl.prompt);
    } catch {
      // sessionStorage quota exceeded is fine — the template still loads.
    }
    navigate('/editor');
  };

  return (
    <section>
      {title && (
        <div className="mb-6 flex items-baseline justify-between">
          <div>
            <h2 className="font-display text-lg font-semibold tracking-[-0.015em] text-c2-text">
              {title}
            </h2>
            <p className="mt-1 text-petit text-c2-text-muted">
              Open one, then edit — these are real diagrams, not stubs.
            </p>
          </div>
        </div>
      )}

      <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
        {TEMPLATES.map((t) => (
          <button
            key={t.id}
            type="button"
            onClick={() => open(t)}
            className="group relative overflow-hidden rounded-xl border border-c2-border bg-c2-surface/60 p-5 text-left transition-all duration-200 hover:border-c2-primary/40 hover:bg-c2-surface hover:shadow-c2"
          >
            <div className="flex items-center justify-between">
              <TagPill tag={t.tag} />
              <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
                {t.diagram.nodes.length} nodes · {t.diagram.edges.length} edges
              </span>
            </div>
            <h3 className="mt-4 font-display text-[15px] font-semibold tracking-[-0.005em] text-c2-text">
              {t.name}
            </h3>
            <p className="mt-1.5 text-petit leading-relaxed text-c2-text-muted">
              {t.summary}
            </p>
            <div className="mt-4 inline-flex items-center gap-1.5 text-tiny font-medium text-c2-primary transition-all group-hover:text-c2-info">
              Open
              <span className="transition-transform group-hover:translate-x-0.5">→</span>
            </div>
          </button>
        ))}
      </div>
    </section>
  );
}

// Tag color is categorical, mapped to the v2 palette. Each pill gets a
// soft tinted bg + matching foreground to read as a small status chip,
// not a button.
const TAG_STYLES: Record<Template['tag'], string> = {
  starter: 'bg-c2-primary-soft text-c2-primary border-c2-primary/30',
  data:    'bg-amber-500/15 text-amber-300 border-amber-500/30',
  app:     'bg-c2-accent-soft text-c2-accent border-c2-accent/30',
  ai:      'bg-violet-500/15 text-violet-300 border-violet-500/30',
};

function TagPill({ tag }: { tag: Template['tag'] }) {
  return (
    <span
      className={`rounded-full border px-2 py-0.5 font-mono text-[9px] font-semibold uppercase tracking-[0.18em] ${TAG_STYLES[tag]}`}
    >
      {tag}
    </span>
  );
}
