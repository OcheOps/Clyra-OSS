import { type MutableRefObject, useCallback, useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { HistoryPanel } from './HistoryPanel';
import { ShareDialog } from './ShareDialog';
import { ReverseFromIaCDialog } from './ReverseFromIaCDialog';
import { useExport } from '../hooks/useExport';
import { useEditorStore } from '../store';
import { useInsightsStore } from '../insightsStore';

// Single overflow button ("⋯") that houses low-use topbar actions so
// the primary row stays readable: history, share, add group, and the
// raster export submenu. Keeps the editor's topbar short enough that
// the prompt input gets real estate at mid widths.

interface Props {
  projectIdRef: MutableRefObject<string | null>;
  projectName: string;
  onAddGroup: () => void;
}

export function EditorOverflowMenu({ projectIdRef, projectName, onAddGroup }: Props) {
  const [open, setOpen] = useState(false);
  const [showExports, setShowExports] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [shareOpen, setShareOpen] = useState(false);
  const [reverseOpen, setReverseOpen] = useState(false);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const [pos, setPos] = useState({ top: 0, right: 0 });
  const loadDiagram = useEditorStore((s) => s.loadDiagram);
  const insightsVisible = useInsightsStore((s) => s.visible);
  const toggleInsights = useInsightsStore((s) => s.toggle);
  const { handleExportJSON, handleExportSVG, handleExportPNG, handleExportPDF } = useExport(projectName);

  useEffect(() => {
    if (!open || !buttonRef.current) return;
    const rect = buttonRef.current.getBoundingClientRect();
    setPos({ top: rect.bottom + 4, right: window.innerWidth - rect.right });
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const onClick = (e: MouseEvent) => {
      const t = e.target as Node;
      if (
        buttonRef.current && !buttonRef.current.contains(t) &&
        menuRef.current && !menuRef.current.contains(t)
      ) {
        setOpen(false);
        setShowExports(false);
      }
    };
    document.addEventListener('mousedown', onClick);
    return () => document.removeEventListener('mousedown', onClick);
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setOpen(false);
        setShowExports(false);
      }
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [open]);

  const run = useCallback((fn: () => void | Promise<void>) => {
    setOpen(false);
    setShowExports(false);
    void fn();
  }, []);

  const hasProject = !!projectIdRef.current;

  return (
    <>
      <button
        ref={buttonRef}
        type="button"
        onClick={() => setOpen((s) => !s)}
        title="More actions"
        className="shrink-0 rounded-lg border border-c2-border bg-c2-bg/40 px-2 py-1 text-tiny font-medium text-c2-text-muted hover:text-c2-text hover:border-c2-border-strong transition-colors"
      >
        ⋯
      </button>

      {open && createPortal(
        <div
          ref={menuRef}
          className="fixed z-50 w-52 rounded-2xl border border-c2-border bg-c2-elevated/95 backdrop-blur-md py-1.5 shadow-c2-lg [body.paper_&]:border-c2-border [body.paper_&]:bg-c2-surface"
          style={{ top: pos.top, right: pos.right }}
        >
          {!showExports && (
            <>
              <Item
                label="Reverse from IaC"
                hint="Paste Terraform → diagram"
                onClick={() => {
                  setOpen(false);
                  setReverseOpen(true);
                }}
              />
              <div className="my-1 border-t border-c2-hairline [body.paper_&]:border-c2-hairline" />
              <Item
                label="Add group"
                hint="Container node"
                onClick={() => run(() => onAddGroup())}
              />
              <Item
                label="Version history"
                hint={hasProject ? 'Last 10 saves' : 'Save to cloud first'}
                disabled={!hasProject}
                onClick={() => {
                  if (!hasProject) return;
                  setOpen(false);
                  setHistoryOpen(true);
                }}
              />
              <Item
                label="Share link"
                hint={hasProject ? 'Read-only URL' : 'Save to cloud first'}
                disabled={!hasProject}
                onClick={() => {
                  if (!hasProject) return;
                  setOpen(false);
                  setShareOpen(true);
                }}
              />
              <div className="my-1 border-t border-c2-hairline [body.paper_&]:border-c2-hairline" />
              <Item
                label={insightsVisible ? 'Hide Insights panel' : 'Show Insights panel'}
                hint="Cost · review"
                onClick={() => run(() => toggleInsights())}
              />
              <Item
                label="Export…"
                hint="PNG · SVG · PDF · JSON"
                trailingArrow
                onClick={() => setShowExports(true)}
              />
            </>
          )}

          {showExports && (
            <>
              <button
                onClick={() => setShowExports(false)}
                className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-micro text-c2-text-muted hover:text-c2-text"
              >
                ← Back
              </button>
              <div className="my-0.5 border-t border-c2-hairline [body.paper_&]:border-c2-hairline" />
              <Item label="PNG" hint="Raster (2×)" onClick={() => run(handleExportPNG)} />
              <Item label="SVG" hint="Vector" onClick={() => run(handleExportSVG)} />
              <Item label="PDF" hint="Docs & decks" onClick={() => run(handleExportPDF)} />
              <Item label="JSON" hint="Diagram data" onClick={() => run(handleExportJSON)} />
            </>
          )}
        </div>,
        document.body,
      )}

      <ShareDialog
        projectID={projectIdRef.current}
        open={shareOpen}
        onClose={() => setShareOpen(false)}
      />
      <HistoryPanel
        projectID={projectIdRef.current}
        open={historyOpen}
        onClose={() => setHistoryOpen(false)}
        onRestored={(diagram) => loadDiagram(diagram as Parameters<typeof loadDiagram>[0])}
      />
      <ReverseFromIaCDialog open={reverseOpen} onClose={() => setReverseOpen(false)} />
    </>
  );
}

function Item({
  label,
  hint,
  onClick,
  disabled,
  trailingArrow,
}: {
  label: string;
  hint?: string;
  onClick: () => void;
  disabled?: boolean;
  trailingArrow?: boolean;
}) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`flex w-full items-center justify-between px-3 py-2 text-left text-xs transition-colors ${
        disabled
          ? 'text-c2-text-faint/50 cursor-not-allowed'
          : 'text-c2-text hover:bg-c2-primary-soft [body.paper_&]:hover:bg-teal-50'
      }`}
    >
      <span className="font-medium">{label}</span>
      <span className="flex items-center gap-1 text-micro text-c2-text-muted">
        {hint}
        {trailingArrow && <span>›</span>}
      </span>
    </button>
  );
}
