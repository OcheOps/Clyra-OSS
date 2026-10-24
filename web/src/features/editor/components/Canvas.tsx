import { useCallback, useEffect } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  BackgroundVariant,
  useReactFlow,
  type OnNodesChange,
  type OnEdgesChange,
  type OnConnect,
  type Node,
  type Edge,
  addEdge,
  type Connection,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { nodeTypes } from '../nodeTypes';
import { edgeTypes } from '../edgeTypes';
import { useEditorStore } from '../store';
import { useDragDrop } from '../hooks/useDragDrop';
import { InsightsPanel } from './InsightsPanel';
import { BlastRadiusInsights } from './BlastRadiusInsights';

export function Canvas() {
  const nodes = useEditorStore((s) => s.nodes);
  const edges = useEditorStore((s) => s.edges);
  const onNodesChange = useEditorStore((s) => s.onNodesChange);
  const onEdgesChange = useEditorStore((s) => s.onEdgesChange);
  const setEdges = useEditorStore((s) => s.setEdges);
  const setSelectedNodeId = useEditorStore((s) => s.setSelectedNodeId);
  const setSelectedEdgeId = useEditorStore((s) => s.setSelectedEdgeId);
  const loadVersion = useEditorStore((s) => s.loadVersion);
  const blastFocus = useEditorStore((s) => s.blastRadiusFocus);
  const setBlastFocus = useEditorStore((s) => s.setBlastRadiusFocus);
  const { onDragOver, onDrop } = useDragDrop();
  const { fitView } = useReactFlow();

  // Escape exits blast mode. Only active WHILE in blast so it doesn't
  // steal Escape from ReactFlow's selection clear when nothing is
  // focused. Removed on cleanup so mounted/unmounted views don't stack
  // listeners.
  useEffect(() => {
    if (!blastFocus) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setBlastFocus(null);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [blastFocus, setBlastFocus]);

  // Re-fit the viewport whenever loadDiagram() bumps loadVersion (cloud
  // open, template open, AI generate, history restore). The `fitView` prop
  // on <ReactFlow> only fires on initial mount, so without this a cloud-
  // open inherits whatever viewport the user last left and the diagram
  // ends up crammed into a corner. requestAnimationFrame defers the fit
  // until after nodes have actually rendered + measured.
  useEffect(() => {
    if (loadVersion === 0 || nodes.length === 0) return;
    const id = requestAnimationFrame(() => {
      fitView({ padding: 0.15, duration: 250 });
    });
    return () => cancelAnimationFrame(id);
    // nodes intentionally omitted — we only refit on a fresh load, not on
    // every node change (that would fight the user's manual panning).
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loadVersion, fitView]);

  const handleConnect: OnConnect = useCallback(
    (connection: Connection) => {
      setEdges((eds: Edge[]) =>
        addEdge(
          {
            ...connection,
            id: `e-${connection.source}-${connection.target}-${Date.now()}`,
            type: 'smoothstep',
            style: { stroke: '#1eaaff', strokeWidth: 1.5, opacity: 0.45 },
          },
          eds,
        ),
      );
    },
    [setEdges],
  );

  const handleSelectionChange = useCallback(
    ({ nodes: selectedNodes, edges: selectedEdges }: { nodes: Node[]; edges: Edge[] }) => {
      if (selectedNodes.length === 1 && selectedNodes[0]) {
        setSelectedNodeId(selectedNodes[0].id);
      } else if (selectedEdges.length === 1 && selectedEdges[0]) {
        setSelectedEdgeId(selectedEdges[0].id);
      } else {
        setSelectedNodeId(null);
      }
    },
    [setSelectedNodeId, setSelectedEdgeId],
  );

  const handlePaneClick = useCallback(() => {
    setSelectedNodeId(null);
  }, [setSelectedNodeId]);

  const handleEdgeClick = useCallback(
    (_: React.MouseEvent, edge: Edge) => {
      setSelectedEdgeId(edge.id);
    },
    [setSelectedEdgeId],
  );

  return (
    <div className="relative flex-1 h-full" onDragOver={onDragOver} onDrop={onDrop}>
      <InsightsPanel />
      <BlastRadiusInsights />
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange as OnNodesChange}
        onEdgesChange={onEdgesChange as OnEdgesChange}
        onConnect={handleConnect}
        onSelectionChange={handleSelectionChange}
        onPaneClick={handlePaneClick}
        onEdgeClick={handleEdgeClick}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        fitView
        fitViewOptions={{ padding: 0.15 }}
        snapToGrid
        snapGrid={[20, 20]}
        minZoom={0.1}
        maxZoom={2}
        deleteKeyCode={['Backspace', 'Delete']}
        multiSelectionKeyCode="Shift"
        className="bg-c2-bg"
        defaultEdgeOptions={{
          type: 'smoothstep',
          style: { stroke: '#1eaaff', strokeWidth: 1.5, opacity: 0.35 },
          labelStyle: { fill: '#9ba4b8', fontSize: 9, fontWeight: 500 },
          labelBgStyle: { fill: '#0f1219', fillOpacity: 0.95 },
          labelBgPadding: [6, 3] as [number, number],
          labelBgBorderRadius: 6,
        }}
      >
        <Background
          variant={BackgroundVariant.Dots}
          gap={24}
          size={0.6}
          color="#1e2433"
        />
        <Controls
          className="!bg-c2-surface/95 !border-c2-hairline !rounded-2xl !shadow-c2 [&>button]:!bg-c2-surface/95 [&>button]:!border-c2-hairline [&>button]:!text-c2-text-muted [&>button:hover]:!bg-c2-bg/60 [&>button]:!rounded-lg"
          showInteractive={false}
        />
        <MiniMap
          className="!bg-c2-surface/95/90 !border-c2-hairline !rounded-2xl !backdrop-blur-sm"
          nodeColor={(node) => {
            if (node.type === 'group') return '#1eaaff10';
            return '#1eaaff';
          }}
          maskColor="rgba(10, 13, 20, 0.85)"
          pannable
          zoomable
        />
      </ReactFlow>
    </div>
  );
}
