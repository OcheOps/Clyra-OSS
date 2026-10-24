import { memo, useState, useCallback, useRef, useEffect } from 'react';
import { Handle, Position, type NodeProps } from '@xyflow/react';
import { getIconDef } from '@/lib/icons';
import { IconRenderer } from '@/components/ui/IconRenderer';
import { useEditorStore } from '../store';
import { useBlastRadius } from '../lib/blastRadiusContext';

interface CloudNodeData {
  label: string;
  nodeType: string;
  metadata?: Record<string, string>;
  [key: string]: unknown;
}

function CloudNodeComponent({ id, data, selected }: NodeProps) {
  const { label, nodeType } = data as CloudNodeData;
  const icon = getIconDef(nodeType);
  const updateNodeData = useEditorStore((s) => s.updateNodeData);
  const { focus, affectedIds, spofs } = useBlastRadius();

  // Three modes for the node visual:
  //   - epicenter: this is the failed node — strong red, pulsing.
  //   - affected: in the blast set — red tint, no pulse.
  //   - dimmed: blast mode active but this node is unaffected — fade out.
  //   - normal: blast mode off.
  // The SPOF badge ⚠ is independent of blast focus; it always shows on
  // nodes whose failure would cascade to >25% of the diagram.
  const isEpicenter = focus === id;
  const isAffected = focus !== null && affectedIds.has(id) && !isEpicenter;
  const isDimmed = focus !== null && !affectedIds.has(id);
  const isSPOF = spofs.has(id);

  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(label);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (editing) {
      inputRef.current?.focus();
      inputRef.current?.select();
    }
  }, [editing]);

  const commitLabel = useCallback(() => {
    const trimmed = draft.trim();
    if (trimmed && trimmed !== label) {
      updateNodeData(id, { label: trimmed });
    } else {
      setDraft(label);
    }
    setEditing(false);
  }, [draft, label, id, updateNodeData]);

  const handleDoubleClick = useCallback(() => {
    setDraft(label);
    setEditing(true);
  }, [label]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') commitLabel();
      else if (e.key === 'Escape') { setDraft(label); setEditing(false); }
    },
    [commitLabel, label],
  );

  // Compose the outer ring + background. Blast styling takes precedence
  // over plain selection — when blast mode is on, the regular cyan
  // selected ring would fight the red epicenter signal.
  let outerClasses = 'border-c2-border bg-c2-surface/95 hover:border-c2-border-strong shadow-c2-sm';
  if (isEpicenter) {
    outerClasses =
      'border-c2-danger ring-2 ring-c2-danger/40 shadow-[0_0_24px_-4px_rgba(248,113,113,0.5)] animate-pulse bg-c2-surface/95';
  } else if (isAffected) {
    outerClasses = 'border-c2-danger/40 ring-1 ring-c2-danger/20 bg-c2-danger/[0.06]';
  } else if (isDimmed) {
    outerClasses = 'border-c2-border/30 opacity-30 grayscale bg-c2-surface/60';
  } else if (selected) {
    outerClasses = 'border-c2-primary/60 shadow-c2-glow ring-1 ring-c2-primary/20 bg-c2-surface/95';
  }

  return (
    <div
      className={`relative flex min-w-[110px] max-w-[160px] flex-col items-center rounded-xl border px-4 py-3 backdrop-blur-sm transition-all duration-150 ${outerClasses}`}
    >
      {isSPOF && !isEpicenter && !isAffected && (
        <span
          title="Single Point of Failure — many things depend on this"
          className="absolute -right-2 -top-2 z-10 flex h-5 w-5 items-center justify-center rounded-full bg-amber-500 text-[10px] font-bold text-white shadow-sm"
        >
          !
        </span>
      )}
      <Handle
        id="top-target"
        type="target"
        position={Position.Top}
        className="!-top-[3px] !h-1.5 !w-1.5 !border-0 !bg-c2-primary/70"
      />
      <Handle
        id="top-source"
        type="source"
        position={Position.Top}
        className="!-top-[3px] !h-1.5 !w-1.5 !border-0 !bg-c2-primary/70"
      />
      <Handle
        id="left-target"
        type="target"
        position={Position.Left}
        className="!-left-[3px] !h-1.5 !w-1.5 !border-0 !bg-c2-primary/70"
      />
      <Handle
        id="left-source"
        type="source"
        position={Position.Left}
        className="!-left-[3px] !h-1.5 !w-1.5 !border-0 !bg-c2-primary/70"
      />

      <div
        className="flex h-9 w-9 items-center justify-center rounded-lg"
        style={{ backgroundColor: `${icon.color}15`, color: icon.color }}
      >
        <IconRenderer name={icon.icon} size={20} strokeWidth={1.5} />
      </div>

      {editing ? (
        <input
          ref={inputRef}
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onBlur={commitLabel}
          onKeyDown={handleKeyDown}
          className="mt-1.5 w-[120px] rounded-md border border-c2-primary/40 bg-c2-bg px-1.5 py-0.5 text-center text-tiny font-medium text-c2-text outline-none focus:border-c2-primary/70 focus:ring-1 focus:ring-c2-primary-ring/30"
        />
      ) : (
        <span
          onDoubleClick={handleDoubleClick}
          className="mt-1.5 max-w-[130px] cursor-text truncate text-center text-tiny font-medium leading-tight text-c2-text"
          title={label}
        >
          {label}
        </span>
      )}

      <span
        className="mt-0.5 max-w-[130px] truncate text-micro font-medium"
        style={{ color: `${icon.color}aa` }}
      >
        {icon.label}
      </span>

      <Handle
        id="bottom-target"
        type="target"
        position={Position.Bottom}
        className="!-bottom-[3px] !h-1.5 !w-1.5 !border-0 !bg-c2-primary/70"
      />
      <Handle
        id="bottom-source"
        type="source"
        position={Position.Bottom}
        className="!-bottom-[3px] !h-1.5 !w-1.5 !border-0 !bg-c2-primary/70"
      />
      <Handle
        id="right-target"
        type="target"
        position={Position.Right}
        className="!-right-[3px] !h-1.5 !w-1.5 !border-0 !bg-c2-primary/70"
      />
      <Handle
        id="right-source"
        type="source"
        position={Position.Right}
        className="!-right-[3px] !h-1.5 !w-1.5 !border-0 !bg-c2-primary/70"
      />
    </div>
  );
}

export const CloudNode = memo(CloudNodeComponent);
