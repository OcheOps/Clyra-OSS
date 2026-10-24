import { useCallback } from 'react';
import { STARTER_TEMPLATES, type StarterTemplate } from '@/lib/templates';
import { useEditorStore } from '../store';

interface OnboardingModalProps {
  open: boolean;
  onDismiss: () => void;
}

export function OnboardingModal({ open, onDismiss }: OnboardingModalProps) {
  const loadDiagram = useEditorStore((s) => s.loadDiagram);
  const setMode = useEditorStore((s) => s.setMode);

  const pick = useCallback(
    (tpl: StarterTemplate) => {
      setMode(tpl.mode);
      loadDiagram(tpl.diagram);
      onDismiss();
    },
    [loadDiagram, setMode, onDismiss],
  );

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-c2-bg/80 backdrop-blur-sm px-4">
      <div className="w-full max-w-3xl rounded-2xl border border-c2-border bg-c2-surface/95/95 shadow-c2-lg">
        <div className="flex items-center justify-between border-b border-c2-hairline px-6 py-4">
          <div>
            <h2 className="text-base font-semibold text-c2-text">Start with a template</h2>
            <p className="mt-0.5 text-micro text-c2-text-muted">
              Pick a starting point, or dismiss and describe your system.
            </p>
          </div>
          <button
            onClick={onDismiss}
            className="rounded-lg px-2 py-1 text-tiny text-c2-text-muted hover:text-c2-text hover:bg-c2-bg/60 transition-colors"
          >
            Skip
          </button>
        </div>

        <div className="grid gap-3 p-6 sm:grid-cols-2">
          {STARTER_TEMPLATES.map((tpl) => (
            <button
              key={tpl.id}
              type="button"
              onClick={() => pick(tpl)}
              className="group rounded-2xl border border-c2-border bg-c2-bg/40 p-4 text-left transition-all hover:border-c2-primary/40 hover:bg-c2-primary-soft"
            >
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-c2-text">{tpl.title}</h3>
                <span className="rounded-full bg-c2-bg/60/60 px-2 py-0.5 text-micro text-c2-text-faint uppercase tracking-wider">
                  {tpl.mode.split('_')[0]}
                </span>
              </div>
              <p className="mt-1.5 text-tiny text-c2-text-muted leading-relaxed">{tpl.summary}</p>
              <p className="mt-3 text-micro text-c2-text-faint">
                {tpl.diagram.nodes.length} nodes · {tpl.diagram.edges.length} edges
              </p>
            </button>
          ))}
        </div>

        <div className="border-t border-c2-hairline px-6 py-3 text-center text-micro text-c2-text-faint">
          You can always start over from the editor.
        </div>
      </div>
    </div>
  );
}
