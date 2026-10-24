import { useCallback } from 'react';
import { getIconDef, getServicesByPrefix } from '@/lib/icons';
import { IconRenderer } from '@/components/ui/IconRenderer';
import { EDGE_STYLES } from '@/lib/constants';
import { useEditorStore } from '../store';
import { useBlastRadius } from '../lib/blastRadiusContext';
import type { EdgeStyle } from '@/types/diagram';

const allServices = Object.entries(getServicesByPrefix()).flatMap(
  ([, items]) => items,
);

/**
 * NodeProperties — v2 restyle.
 *
 * Right-sidebar inspector. Switches between an empty state (shortcut
 * tips), a node panel, and an edge panel based on selection. All form
 * surfaces use the shared INPUT_CLASSES; section dividers are c2-
 * hairline; primary action chips pick up the v2 voice (cyan-soft for
 * blast/duplicate, danger-soft for delete).
 */
export function NodeProperties() {
  const selectedNodeId = useEditorStore((s) => s.selectedNodeId);
  const selectedEdgeId = useEditorStore((s) => s.selectedEdgeId);
  const nodes = useEditorStore((s) => s.nodes);
  const edges = useEditorStore((s) => s.edges);
  const updateNodeData = useEditorStore((s) => s.updateNodeData);
  const removeNode = useEditorStore((s) => s.removeNode);
  const duplicateNode = useEditorStore((s) => s.duplicateNode);
  const updateEdgeLabel = useEditorStore((s) => s.updateEdgeLabel);
  const updateEdgeStyle = useEditorStore((s) => s.updateEdgeStyle);
  const removeEdge = useEditorStore((s) => s.removeEdge);

  const node = selectedNodeId ? nodes.find((n) => n.id === selectedNodeId) : null;
  const edge = selectedEdgeId ? edges.find((e) => e.id === selectedEdgeId) : null;

  if (edge) {
    return (
      <EdgeProperties
        edge={edge}
        onLabelChange={(label) => updateEdgeLabel(edge.id, label)}
        onStyleChange={(style) => updateEdgeStyle(edge.id, style)}
        onDelete={() => removeEdge(edge.id)}
      />
    );
  }

  if (!node || !selectedNodeId) {
    return <EmptyState />;
  }

  return (
    <NodePanel
      node={node}
      onLabelChange={(label) => updateNodeData(selectedNodeId, { label })}
      onTypeChange={(nodeType) => updateNodeData(selectedNodeId, { nodeType })}
      onDuplicate={() => duplicateNode(selectedNodeId)}
      onDelete={() => removeNode(selectedNodeId)}
    />
  );
}

// ────────────────────────────────────────────────────────────────────

function EmptyState() {
  return (
    <div className="p-4">
      <p className="text-petit text-c2-text-muted">
        Select a node or edge to edit its properties.
      </p>
      <div className="mt-5 space-y-1.5 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
        <p>Ctrl+Z — Undo</p>
        <p>Ctrl+Shift+Z — Redo</p>
        <p>Ctrl+D — Duplicate</p>
        <p>Delete — Remove selected</p>
      </div>
    </div>
  );
}

// ────────────────────────────────────────────────────────────────────
// Blast-radius controls — selected-node trigger + post-trigger
// affected-count summary, plus the SPOF chip if applicable.

function BlastRadiusControls({ nodeId }: { nodeId: string }) {
  const setBlastRadiusFocus = useEditorStore((s) => s.setBlastRadiusFocus);
  const { focus, affected, spofs } = useBlastRadius();
  const isActive = focus === nodeId;
  const isSPOF = spofs.has(nodeId);

  // affected is now an AffectedNode[] ranked list (was a Set). Downstream
  // count = list length minus the epicenter itself.
  const downstreamCount = isActive ? Math.max(0, affected.length - 1) : 0;

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <FieldLabel>Blast radius</FieldLabel>
        {isSPOF && (
          <span
            className="inline-flex items-center gap-1 rounded-full border border-amber-500/40 bg-amber-500/15 px-1.5 py-0.5 font-mono text-[9px] font-semibold uppercase tracking-[0.18em] text-amber-300"
            title="Failure cascades to >25% of nodes"
          >
            SPOF
          </span>
        )}
      </div>
      {!isActive ? (
        <button
          type="button"
          onClick={() => setBlastRadiusFocus(nodeId)}
          className="w-full rounded-full border border-c2-border bg-c2-bg/40 px-3 py-1.5 text-tiny font-medium text-c2-text transition-colors hover:border-c2-border-strong hover:bg-c2-surface"
        >
          What breaks if this fails?
        </button>
      ) : (
        <>
          <p className="text-petit text-c2-text-muted">
            <span className="font-semibold text-c2-danger">{downstreamCount}</span>{' '}
            downstream {downstreamCount === 1 ? 'node breaks' : 'nodes break'} if this fails.
          </p>
          <button
            type="button"
            onClick={() => setBlastRadiusFocus(null)}
            className="w-full rounded-full border border-c2-border bg-c2-bg/40 px-3 py-1.5 text-tiny font-medium text-c2-text transition-colors hover:border-c2-border-strong hover:bg-c2-surface"
          >
            Exit blast view
          </button>
        </>
      )}
    </div>
  );
}

// ────────────────────────────────────────────────────────────────────
// Node panel

interface NodePanelProps {
  node: {
    id: string;
    type?: string;
    position: { x: number; y: number };
    data: Record<string, unknown>;
  };
  onLabelChange: (label: string) => void;
  onTypeChange: (type: string) => void;
  onDuplicate: () => void;
  onDelete: () => void;
}

function NodePanel({
  node,
  onLabelChange,
  onTypeChange,
  onDuplicate,
  onDelete,
}: NodePanelProps) {
  const label = (node.data.label as string) ?? '';
  const nodeType = (node.data.nodeType as string) ?? 'generic.server';
  const icon = getIconDef(nodeType);
  const isGroup = node.type === 'group';

  const handleLabel = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => onLabelChange(e.target.value),
    [onLabelChange],
  );
  const handleType = useCallback(
    (e: React.ChangeEvent<HTMLSelectElement>) => onTypeChange(e.target.value),
    [onTypeChange],
  );

  return (
    <div className="flex flex-col gap-4 p-3">
      <div className="flex items-center gap-2.5">
        <span
          className="flex h-9 w-9 items-center justify-center rounded-lg"
          style={{ backgroundColor: `${icon.color}15`, color: icon.color }}
        >
          <IconRenderer name={icon.icon} size={18} strokeWidth={1.5} />
        </span>
        <div className="min-w-0 flex-1">
          <p className="truncate font-display text-sm font-semibold tracking-[-0.005em] text-c2-text">
            {label}
          </p>
          <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            {node.id}
          </p>
        </div>
      </div>

      <Divider />

      <Field label="Label">
        <input
          type="text"
          value={label}
          onChange={handleLabel}
          className={INPUT_CLASSES}
        />
      </Field>

      {!isGroup && (
        <Field label="Service type">
          <select value={nodeType} onChange={handleType} className={INPUT_CLASSES}>
            {allServices.map((s) => (
              <option key={s.type} value={s.type}>
                {s.icon.label} — {s.type}
              </option>
            ))}
          </select>
        </Field>
      )}

      <Field label="Position">
        <div className="flex gap-3 font-mono text-petit tabular-nums text-c2-text-muted">
          <span>x · {Math.round(node.position.x)}</span>
          <span>y · {Math.round(node.position.y)}</span>
        </div>
      </Field>

      {!isGroup && (
        <>
          <Divider />
          <BlastRadiusControls nodeId={node.id} />
        </>
      )}

      <Divider />

      <div className="flex gap-2">
        {!isGroup && (
          <button
            type="button"
            onClick={onDuplicate}
            className="flex-1 rounded-full bg-c2-primary-soft px-3 py-1.5 text-tiny font-medium text-c2-primary transition-colors hover:bg-c2-primary/25"
          >
            Duplicate
          </button>
        )}
        <button
          type="button"
          onClick={onDelete}
          className="flex-1 rounded-full border border-c2-danger/40 bg-c2-danger/10 px-3 py-1.5 text-tiny font-medium text-c2-danger transition-colors hover:border-c2-danger/60 hover:bg-c2-danger/20"
        >
          Delete
        </button>
      </div>
    </div>
  );
}

// ────────────────────────────────────────────────────────────────────
// Edge panel

interface EdgePropertiesProps {
  edge: {
    id: string;
    label?: string | React.ReactNode;
    style?: React.CSSProperties;
    animated?: boolean;
  };
  onLabelChange: (label: string) => void;
  onStyleChange: (style: EdgeStyle) => void;
  onDelete: () => void;
}

function EdgeProperties({
  edge,
  onLabelChange,
  onStyleChange,
  onDelete,
}: EdgePropertiesProps) {
  const currentLabel = typeof edge.label === 'string' ? edge.label : '';

  const currentStyle: EdgeStyle = (() => {
    const dash = edge.style?.strokeDasharray;
    if (typeof dash === 'string') {
      if (dash.startsWith('2')) return 'dotted';
      if (dash.startsWith('6')) return 'dashed';
    }
    if (edge.animated) return 'dashed';
    return 'solid';
  })();

  return (
    <div className="flex flex-col gap-4 p-3">
      <div>
        <p className="font-display text-sm font-semibold tracking-[-0.005em] text-c2-text">
          Edge
        </p>
        <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
          {edge.id}
        </p>
      </div>

      <Divider />

      <Field label="Label">
        <input
          type="text"
          value={currentLabel}
          onChange={(e) => onLabelChange(e.target.value)}
          placeholder="e.g. HTTPS, gRPC, events…"
          className={INPUT_CLASSES}
        />
      </Field>

      <Field label="Line style">
        <div className="flex gap-1">
          {EDGE_STYLES.map((style) => (
            <button
              key={style}
              onClick={() => onStyleChange(style)}
              className={`flex-1 rounded-lg border px-2 py-1.5 text-petit capitalize transition-colors ${
                currentStyle === style
                  ? 'border-c2-primary/40 bg-c2-primary-soft text-c2-primary'
                  : 'border-c2-border bg-c2-bg/30 text-c2-text-muted hover:border-c2-border-strong hover:text-c2-text'
              }`}
            >
              {style}
            </button>
          ))}
        </div>
      </Field>

      <Divider />

      <button
        type="button"
        onClick={onDelete}
        className="rounded-full border border-c2-danger/40 bg-c2-danger/10 px-3 py-1.5 text-tiny font-medium text-c2-danger transition-colors hover:border-c2-danger/60 hover:bg-c2-danger/20"
      >
        Delete edge
      </button>
    </div>
  );
}

// ────────────────────────────────────────────────────────────────────
// Inline primitives

const INPUT_CLASSES =
  'w-full rounded-lg border border-c2-border bg-c2-bg/40 px-3 py-2 text-sm text-c2-text ' +
  'placeholder:text-c2-text-faint focus:border-c2-primary/60 focus:bg-c2-surface focus:outline-none ' +
  'focus:ring-2 focus:ring-c2-primary-ring/40 transition-colors';

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-2">
      <FieldLabel>{label}</FieldLabel>
      {children}
    </div>
  );
}

function FieldLabel({ children }: { children: React.ReactNode }) {
  return (
    <label className="block font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
      {children}
    </label>
  );
}

function Divider() {
  return <hr className="border-c2-hairline" />;
}
