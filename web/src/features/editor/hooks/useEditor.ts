import { useMemo } from 'react';
import { useEditorStore } from '../store';

/**
 * Convenience hook that returns commonly-used editor state and actions
 * without needing individual selectors everywhere.
 */
export function useEditor() {
  const nodes = useEditorStore((s) => s.nodes);
  const edges = useEditorStore((s) => s.edges);
  const selectedNodeId = useEditorStore((s) => s.selectedNodeId);
  const selectedEdgeId = useEditorStore((s) => s.selectedEdgeId);
  const canUndo = useEditorStore((s) => s.canUndo);
  const canRedo = useEditorStore((s) => s.canRedo);

  const selectedNode = useMemo(
    () => (selectedNodeId ? nodes.find((n) => n.id === selectedNodeId) ?? null : null),
    [nodes, selectedNodeId],
  );

  const selectedEdge = useMemo(
    () => (selectedEdgeId ? edges.find((e) => e.id === selectedEdgeId) ?? null : null),
    [edges, selectedEdgeId],
  );

  const stats = useMemo(() => {
    const cloudNodes = nodes.filter((n) => n.type === 'cloud');
    const groupNodes = nodes.filter((n) => n.type === 'group');
    return {
      totalNodes: cloudNodes.length,
      totalGroups: groupNodes.length,
      totalEdges: edges.length,
    };
  }, [nodes, edges]);

  return {
    nodes,
    edges,
    selectedNode,
    selectedEdge,
    selectedNodeId,
    selectedEdgeId,
    canUndo,
    canRedo,
    stats,

    // Actions (stable references from zustand)
    addNode: useEditorStore.getState().addNode,
    updateNodeData: useEditorStore.getState().updateNodeData,
    removeNode: useEditorStore.getState().removeNode,
    duplicateNode: useEditorStore.getState().duplicateNode,
    updateEdgeLabel: useEditorStore.getState().updateEdgeLabel,
    updateEdgeStyle: useEditorStore.getState().updateEdgeStyle,
    removeEdge: useEditorStore.getState().removeEdge,
    loadDiagram: useEditorStore.getState().loadDiagram,
    getDiagram: useEditorStore.getState().getDiagram,
    clearCanvas: useEditorStore.getState().clearCanvas,
    undo: useEditorStore.getState().undo,
    redo: useEditorStore.getState().redo,
  };
}
