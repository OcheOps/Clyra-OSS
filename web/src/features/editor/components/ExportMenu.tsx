import { useState, useRef, useEffect, useCallback } from 'react';
import { createPortal } from 'react-dom';
import { Button } from '@/components/ui/Button';
import { useExport } from '../hooks/useExport';

interface ExportMenuProps {
  projectName: string;
}

export function ExportMenu({ projectName }: ExportMenuProps) {
  const [open, setOpen] = useState(false);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const [pos, setPos] = useState({ top: 0, right: 0 });
  const { handleExportJSON, handleExportSVG, handleExportPNG, handleExportPDF } = useExport(projectName);

  // Position the dropdown relative to the button
  useEffect(() => {
    if (!open || !buttonRef.current) return;
    const rect = buttonRef.current.getBoundingClientRect();
    setPos({
      top: rect.bottom + 4,
      right: window.innerWidth - rect.right,
    });
  }, [open]);

  // Close on click outside
  useEffect(() => {
    if (!open) return;
    function handleClick(e: MouseEvent) {
      const target = e.target as Node;
      if (
        buttonRef.current && !buttonRef.current.contains(target) &&
        menuRef.current && !menuRef.current.contains(target)
      ) {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [open]);

  // Close on escape
  useEffect(() => {
    if (!open) return;
    const handleKey = (e: KeyboardEvent) => { if (e.key === 'Escape') setOpen(false); };
    document.addEventListener('keydown', handleKey);
    return () => document.removeEventListener('keydown', handleKey);
  }, [open]);

  const doExport = useCallback(
    (fn: () => void | Promise<void>) => {
      setOpen(false);
      fn();
    },
    [],
  );

  return (
    <>
      <Button
        ref={buttonRef}
        variant="secondary"
        size="sm"
        onClick={() => setOpen((s) => !s)}
      >
        Export
      </Button>

      {open && createPortal(
        <div
          ref={menuRef}
          className="fixed z-50 w-44 rounded-2xl border border-c2-border bg-c2-elevated/95 backdrop-blur-md py-1.5 shadow-c2-lg"
          style={{ top: pos.top, right: pos.right }}
        >
          <ExportItem
            label="JSON"
            description="Diagram data"
            onClick={() => doExport(handleExportJSON)}
          />
          <ExportItem
            label="SVG"
            description="Vector image"
            onClick={() => doExport(handleExportSVG)}
          />
          <ExportItem
            label="PNG"
            description="Raster image (2x)"
            onClick={() => doExport(handleExportPNG)}
          />
          <ExportItem
            label="PDF"
            description="For docs & decks"
            onClick={() => doExport(handleExportPDF)}
          />
        </div>,
        document.body,
      )}
    </>
  );
}

function ExportItem({ label, description, onClick }: { label: string; description: string; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="flex w-full items-center justify-between px-3 py-2 text-left text-xs text-c2-text hover:bg-c2-primary-soft transition-colors"
    >
      <span className="font-medium">{label}</span>
      <span className="text-micro text-c2-text-muted">{description}</span>
    </button>
  );
}
