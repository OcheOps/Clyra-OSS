import { memo, useState, useCallback, useRef, useEffect } from 'react';
import { type NodeProps } from '@xyflow/react';
import { getFamilyColor } from '@/lib/icons';
import { useEditorStore } from '../store';

interface GroupNodeData {
  label: string;
  nodeType: string;
  [key: string]: unknown;
}

function GroupNodeComponent({ id, data, selected }: NodeProps) {
  const { label, nodeType } = data as GroupNodeData;
  const color = getFamilyColor(nodeType);
  const updateNodeData = useEditorStore((s) => s.updateNodeData);

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

  return (
    <div
      className={`h-full w-full rounded-2xl border transition-colors duration-150 ${
        selected ? 'border-c2-primary/50' : ''
      }`}
      style={{
        borderColor: selected ? undefined : `${color}20`,
        backgroundColor: `${color}05`,
      }}
    >
      {editing ? (
        <input
          ref={inputRef}
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onBlur={commitLabel}
          onKeyDown={handleKeyDown}
          className="absolute left-3 top-2.5 rounded-md border border-c2-primary/40 bg-c2-bg px-2 py-0.5 font-mono text-[10px] font-semibold uppercase tracking-[0.18em] outline-none focus:ring-1 focus:ring-c2-primary-ring/30"
          style={{ color }}
        />
      ) : (
        <div
          onDoubleClick={handleDoubleClick}
          className="absolute left-3 top-2.5 cursor-text rounded-md px-2 py-0.5 font-mono text-[10px] font-semibold uppercase tracking-[0.18em]"
          style={{ color, backgroundColor: `${color}0a` }}
          title="Double-click to edit"
        >
          {label}
        </div>
      )}
    </div>
  );
}

export const GroupNode = memo(GroupNodeComponent);
