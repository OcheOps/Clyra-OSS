import { useEditorStore } from '../store';

export function EditorStatusBar() {
  const nodes = useEditorStore((s) => s.nodes);
  const edges = useEditorStore((s) => s.edges);
  const canUndo = useEditorStore((s) => s.canUndo);
  const canRedo = useEditorStore((s) => s.canRedo);
  const undo = useEditorStore((s) => s.undo);
  const redo = useEditorStore((s) => s.redo);

  const cloudNodes = nodes.filter((n) => n.type === 'cloud').length;
  const groups = nodes.filter((n) => n.type === 'group').length;

  return (
    <div className="flex h-6 items-center justify-between border-t border-c2-hairline bg-c2-bg/50 px-3 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint backdrop-blur">
      <div className="flex items-center gap-3 tabular-nums">
        <span>{cloudNodes} nodes</span>
        {groups > 0 && <span>{groups} groups</span>}
        <span>{edges.length} edges</span>
      </div>

      <div className="flex items-center gap-0.5">
        <button
          onClick={undo}
          disabled={!canUndo}
          className="rounded px-1.5 py-0.5 uppercase tracking-[0.18em] transition-colors hover:bg-c2-bg/60 hover:text-c2-text disabled:cursor-default disabled:opacity-30"
        >
          Undo
        </button>
        <button
          onClick={redo}
          disabled={!canRedo}
          className="rounded px-1.5 py-0.5 uppercase tracking-[0.18em] transition-colors hover:bg-c2-bg/60 hover:text-c2-text disabled:cursor-default disabled:opacity-30"
        >
          Redo
        </button>
      </div>
    </div>
  );
}
