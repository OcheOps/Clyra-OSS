import { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import toast from 'react-hot-toast';
import { APIError } from '@/types/api';
import { ingestTerraform } from '@/services/api/iacReverse';
import { useEditorStore } from '../store';
import { useUpgradeModalStore } from '@/features/billing/upgradeModalStore';

interface Props {
  open: boolean;
  onClose: () => void;
}

// Generous client cap. The backend hard-limits at 50KB; this is the soft
// limit we surface in the helper text so the textarea visibly fills up
// before the user hits the wire.
const MAX_CHARS = 50 * 1024;

// The "Reverse from IaC" entry point. Single textarea, paste TF, submit,
// the editor canvas reflects the parsed architecture. Naming chosen to
// future-proof CloudFormation reverse later without rewording the menu
// (locked decision in reverse-ingestion-spec.md).
export function ReverseFromIaCDialog({ open, onClose }: Props) {
  const [source, setSource] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [warnings, setWarnings] = useState<string[]>([]);
  const [unsupported, setUnsupported] = useState<string[]>([]);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const loadDiagram = useEditorStore((s) => s.loadDiagram);

  // Auto-focus the textarea when opening — pasting is the only path here,
  // skipping the click saves a step every time.
  useEffect(() => {
    if (open) {
      // requestAnimationFrame: the portal mounts after open flips, focusing
      // synchronously hits a not-yet-attached node.
      requestAnimationFrame(() => textareaRef.current?.focus());
    } else {
      // Reset on close so reopening doesn't show last run's notes.
      setSource('');
      setWarnings([]);
      setUnsupported([]);
      setSubmitting(false);
    }
  }, [open]);

  // ESC to dismiss matches the rest of the editor dialogs.
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !submitting) onClose();
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [open, submitting, onClose]);

  const onSubmit = async () => {
    if (!source.trim() || submitting) return;
    setSubmitting(true);
    setWarnings([]);
    setUnsupported([]);
    try {
      const res = await ingestTerraform(source);
      loadDiagram(res.diagram);
      setWarnings(res.warnings ?? []);
      setUnsupported(res.unsupported ?? []);
      // If the ingest produced *no* approximations, close immediately —
      // there's nothing for the user to read. If there are notes, leave
      // the modal open so they can review them before dismissing.
      const hasNotes = (res.warnings?.length ?? 0) > 0 || (res.unsupported?.length ?? 0) > 0;
      if (!hasNotes) {
        toast.success(
          `Imported ${res.diagram.nodes.length} node${res.diagram.nodes.length === 1 ? '' : 's'}`,
        );
        onClose();
      } else {
        toast.success(`Imported ${res.diagram.nodes.length} nodes (with notes)`);
      }
    } catch (err) {
      handleIngestError(err);
    } finally {
      setSubmitting(false);
    }
  };

  // Error mapping mirrors the backend's status catalog in handler.go.
  // 429 routes to the upgrade modal because the practical fix is the same
  // as for the AI generation quota — buy more daily generations.
  const handleIngestError = (err: unknown) => {
    if (err instanceof APIError) {
      switch (err.status) {
        case 401:
          toast.error('Sign in to import');
          return;
        case 413:
          toast.error('That TF file is too large. Trim it under 50KB.');
          return;
        case 429:
          useUpgradeModalStore.getState().openModal({
            defaultPlan: 'pro',
            headline: 'Daily import limit reached',
            subline: 'Imports share the same daily quota as AI generations. Upgrade for a bigger pool.',
          });
          onClose();
          return;
        case 400:
          // Parser errors include line numbers — show the raw message.
          toast.error(`Invalid Terraform: ${err.message}`);
          return;
      }
    }
    toast.error('Import failed. Try again or report it.');
  };

  if (!open) return null;

  const over = source.length > MAX_CHARS;

  return createPortal(
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm px-4"
      onClick={() => !submitting && onClose()}
    >
      <div
        className="relative w-full max-w-2xl rounded-2xl border border-c2-border bg-c2-bg shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-start justify-between border-b border-c2-hairline p-5">
          <div>
            <h2 className="text-base font-semibold text-c2-text">Reverse from IaC</h2>
            <p className="mt-1 text-petit text-c2-text-muted">
              Paste Terraform and we'll render the architecture.
              {' '}<span className="text-c2-text-faint">Modules show as opaque boxes; unsupported resources become placeholders you can rename.</span>
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            disabled={submitting}
            className="rounded p-1 text-c2-text-faint hover:bg-c2-surface/95 hover:text-c2-text disabled:opacity-40"
            aria-label="Close"
          >
            <svg className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
              <path d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" />
            </svg>
          </button>
        </div>

        <div className="p-5 space-y-3">
          <textarea
            ref={textareaRef}
            value={source}
            onChange={(e) => setSource(e.target.value)}
            disabled={submitting}
            placeholder={'resource "aws_vpc" "main" {\n  cidr_block = "10.0.0.0/16"\n}\n\nresource "aws_subnet" "public" {\n  vpc_id     = aws_vpc.main.id\n  cidr_block = "10.0.1.0/24"\n}'}
            spellCheck={false}
            className={`h-72 w-full resize-none rounded-lg border bg-c2-surface/95/40 p-3 font-mono text-petit text-c2-text outline-none transition-colors ${
              over
                ? 'border-danger/60 focus:border-danger'
                : 'border-c2-border focus:border-c2-primary/60'
            }`}
          />

          <div className="flex items-center justify-between text-micro">
            <span className={over ? 'text-danger' : 'text-c2-text-faint'}>
              {source.length.toLocaleString()} / {MAX_CHARS.toLocaleString()} chars
            </span>
            <span className="text-c2-text-faint">Daily imports share your AI generation quota.</span>
          </div>

          {(warnings.length > 0 || unsupported.length > 0) && (
            <div className="rounded-lg border border-c2-hairline bg-c2-surface/95/30 p-3 space-y-2">
              {warnings.length > 0 && (
                <div>
                  <p className="text-micro font-medium uppercase tracking-wide text-c2-text-muted">
                    Notes
                  </p>
                  <ul className="mt-1 space-y-0.5">
                    {warnings.map((w, i) => (
                      <li key={i} className="text-petit text-c2-text-muted">
                        · {w}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
              {unsupported.length > 0 && (
                <div>
                  <p className="text-micro font-medium uppercase tracking-wide text-c2-text-muted">
                    Rendered as placeholders
                  </p>
                  <div className="mt-1 flex flex-wrap gap-1">
                    {unsupported.map((t) => (
                      <span
                        key={t}
                        className="rounded border border-c2-hairline bg-c2-bg px-1.5 py-0.5 font-mono text-micro text-c2-text-muted"
                      >
                        {t}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        <div className="flex items-center justify-end gap-2 border-t border-c2-hairline bg-c2-surface/95/30 p-5">
          <button
            type="button"
            onClick={onClose}
            disabled={submitting}
            className="rounded-lg border border-c2-border bg-transparent px-3 py-1.5 text-tiny font-medium text-c2-text-muted hover:text-c2-text hover:border-c2-border-strong transition-colors disabled:opacity-50"
          >
            {warnings.length > 0 || unsupported.length > 0 ? 'Done' : 'Cancel'}
          </button>
          <button
            type="button"
            onClick={onSubmit}
            disabled={submitting || over || !source.trim()}
            className="rounded-lg bg-c2-primary px-3 py-1.5 text-tiny font-medium text-white shadow-sm hover:bg-brand-400 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {submitting ? 'Parsing…' : 'Import'}
          </button>
        </div>
      </div>
    </div>,
    document.body,
  );
}
