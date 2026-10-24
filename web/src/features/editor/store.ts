import { create } from 'zustand';
import {
  type Node,
  type Edge,
  applyNodeChanges,
  applyEdgeChanges,
  type NodeChange,
  type EdgeChange,
} from '@xyflow/react';
import type { Diagram, EdgeStyle, ArchitectureMode } from '@/types/diagram';
import { diagramToFlow, flowToDiagram } from '@/lib/normalize';

interface HistoryEntry {
  nodes: Node[];
  edges: Edge[];
}

interface EditorState {
  // Data
  nodes: Node[];
  edges: Edge[];
  selectedNodeId: string | null;
  selectedEdgeId: string | null;

  // Architecture mode
  mode: ArchitectureMode;
  setMode: (mode: ArchitectureMode) => void;

  // Blast-radius focus mode. When non-null, the canvas tints every node
  // that would break if this one fails. Toggled from the NodeProperties
  // panel ("What breaks if this fails?"); cleared by clicking the focused
  // node again, pressing Escape, or selecting a different node entirely.
  blastRadiusFocus: string | null;
  setBlastRadiusFocus: (id: string | null) => void;

  // Pending cloud-project metadata. Set by ProjectsPage.handleOpen before
  // navigating to /editor; consumed (and cleared) by useAutosave on mount so
  // the editor title bar reflects the project the user just opened instead
  // of stuck on "Untitled Diagram".
  pendingProjectId: string | null;
  pendingProjectName: string | null;
  setPendingProject: (id: string | null, name: string | null) => void;
  loadVersion: number;

  // Undo/redo
  past: HistoryEntry[];
  future: HistoryEntry[];
  canUndo: boolean;
  canRedo: boolean;

  // React Flow callbacks
  onNodesChange: (changes: NodeChange[]) => void;
  onEdgesChange: (changes: EdgeChange[]) => void;

  // Node actions
  setNodes: (nodes: Node[] | ((prev: Node[]) => Node[])) => void;
  setEdges: (edges: Edge[] | ((prev: Edge[]) => Edge[])) => void;
  setSelectedNodeId: (id: string | null) => void;
  setSelectedEdgeId: (id: string | null) => void;

  addNode: (node: Node) => void;
  updateNodeData: (id: string, data: Partial<Record<string, unknown>>) => void;
  removeNode: (id: string) => void;
  duplicateNode: (id: string) => void;

  // Edge actions
  updateEdgeLabel: (id: string, label: string) => void;
  updateEdgeStyle: (id: string, style: EdgeStyle) => void;
  removeEdge: (id: string) => void;

  // Diagram operations
  loadDiagram: (diagram: Diagram) => void;
  getDiagram: () => Diagram;
  clearCanvas: () => void;

  // History
  pushHistory: () => void;
  undo: () => void;
  redo: () => void;
}

const MAX_HISTORY = 50;

const EDGE_STYLE_MAP: Record<EdgeStyle, { strokeDasharray?: string; animated: boolean }> = {
  solid: { animated: false },
  dashed: { strokeDasharray: '6 3', animated: true },
  dotted: { strokeDasharray: '2 2', animated: false },
};

export const useEditorStore = create<EditorState>((set, get) => ({
  nodes: [],
  edges: [],
  mode: 'solution_architecture' as ArchitectureMode,
  setMode: (mode) => set({ mode }),

  blastRadiusFocus: null,
  setBlastRadiusFocus: (id) => set({ blastRadiusFocus: id }),

  pendingProjectId: null,
  pendingProjectName: null,
  setPendingProject: (id, name) => set({ pendingProjectId: id, pendingProjectName: name }),

  // Bumped every time loadDiagram() runs. Canvas watches this to call
  // fitView() after a fresh load — ReactFlow's own `fitView` prop only
  // fires on first mount, so without this a cloud-open lands with the
  // diagram crammed into whatever viewport the user last left behind.
  loadVersion: 0,
  selectedNodeId: null,
  selectedEdgeId: null,
  past: [],
  future: [],
  canUndo: false,
  canRedo: false,

  onNodesChange: (changes) => {
    set((state) => ({
      nodes: applyNodeChanges(changes, state.nodes),
    }));
  },

  onEdgesChange: (changes) => {
    set((state) => ({
      edges: applyEdgeChanges(changes, state.edges),
    }));
  },

  setNodes: (updater) => {
    set((state) => ({
      nodes: typeof updater === 'function' ? updater(state.nodes) : updater,
    }));
  },

  setEdges: (updater) => {
    set((state) => ({
      edges: typeof updater === 'function' ? updater(state.edges) : updater,
    }));
  },

  setSelectedNodeId: (id) => set({ selectedNodeId: id, selectedEdgeId: null }),
  setSelectedEdgeId: (id) => set({ selectedEdgeId: id, selectedNodeId: null }),

  // --- Node actions ---

  addNode: (node) => {
    const state = get();
    state.pushHistory();
    set({ nodes: [...state.nodes, node] });
  },

  updateNodeData: (id, data) => {
    const state = get();
    state.pushHistory();
    set({
      nodes: state.nodes.map((n) =>
        n.id === id ? { ...n, data: { ...n.data, ...data } } : n,
      ),
    });
  },

  removeNode: (id) => {
    const state = get();
    state.pushHistory();
    set({
      nodes: state.nodes.filter((n) => n.id !== id),
      edges: state.edges.filter((e) => e.source !== id && e.target !== id),
      selectedNodeId: state.selectedNodeId === id ? null : state.selectedNodeId,
    });
  },

  duplicateNode: (id) => {
    const state = get();
    const original = state.nodes.find((n) => n.id === id);
    if (!original) return;

    state.pushHistory();
    const newId = `node-${Date.now()}-dup`;
    const clone: Node = {
      ...original,
      id: newId,
      position: {
        x: original.position.x + 40,
        y: original.position.y + 40,
      },
      selected: false,
      data: { ...original.data },
    };
    set({ nodes: [...state.nodes, clone], selectedNodeId: newId });
  },

  // --- Edge actions ---

  updateEdgeLabel: (id, label) => {
    const state = get();
    state.pushHistory();
    set({
      edges: state.edges.map((e) =>
        e.id === id ? { ...e, label } : e,
      ),
    });
  },

  updateEdgeStyle: (id, style) => {
    const mapping = EDGE_STYLE_MAP[style];
    const state = get();
    state.pushHistory();
    set({
      edges: state.edges.map((e) =>
        e.id === id
          ? {
              ...e,
              style: { ...e.style, strokeDasharray: mapping.strokeDasharray },
              animated: mapping.animated,
            }
          : e,
      ),
    });
  },

  removeEdge: (id) => {
    const state = get();
    state.pushHistory();
    set({
      edges: state.edges.filter((e) => e.id !== id),
      selectedEdgeId: state.selectedEdgeId === id ? null : state.selectedEdgeId,
    });
  },

  // --- Diagram operations ---

  loadDiagram: (diagram) => {
    const { nodes, edges } = diagramToFlow(diagram);
    set((state) => ({
      nodes,
      edges,
      selectedNodeId: null,
      selectedEdgeId: null,
      past: [],
      future: [],
      canUndo: false,
      canRedo: false,
      loadVersion: state.loadVersion + 1,
    }));
  },

  getDiagram: () => {
    const { nodes, edges } = get();
    return flowToDiagram(nodes, edges);
  },

  clearCanvas: () => {
    const state = get();
    state.pushHistory();
    set({ nodes: [], edges: [], selectedNodeId: null, selectedEdgeId: null });
  },

  // --- History ---

  pushHistory: () => {
    set((state) => {
      const newPast = [...state.past.slice(-MAX_HISTORY + 1), { nodes: state.nodes, edges: state.edges }];
      return {
        past: newPast,
        future: [],
        canUndo: true,
        canRedo: false,
      };
    });
  },

  undo: () => {
    const { past, nodes, edges } = get();
    if (past.length === 0) return;
    const prev = past[past.length - 1]!;
    const newPast = past.slice(0, -1);
    set({
      past: newPast,
      future: [{ nodes, edges }, ...get().future],
      nodes: prev.nodes,
      edges: prev.edges,
      canUndo: newPast.length > 0,
      canRedo: true,
    });
  },

  redo: () => {
    const { future, nodes, edges } = get();
    if (future.length === 0) return;
    const next = future[0]!;
    const newFuture = future.slice(1);
    const newPast = [...get().past, { nodes, edges }];
    set({
      future: newFuture,
      past: newPast,
      nodes: next.nodes,
      edges: next.edges,
      canUndo: true,
      canRedo: newFuture.length > 0,
    });
  },
}));
