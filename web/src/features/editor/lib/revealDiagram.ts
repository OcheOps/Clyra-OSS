import { diagramToFlow } from '@/lib/normalize';
import type { Diagram } from '@/types/diagram';
import { useEditorStore } from '../store';

// revealDiagram replaces the canvas with `diagram` but stages node
// insertion so the final result appears to paint in instead of pop.
// Edges are appended after all nodes mount, because React Flow throws
// "source/target not found" warnings if an edge lands before its
// endpoints. Combined with the `.react-flow__node { animation }` rule
// in index.css, each node fades + rises as it enters.
//
// Fallback to a direct load if the diagram is tiny (<=2 nodes) — the
// stagger is pointless there and just delays display.

const STAGGER_MS = 55;
const BATCH = 1;
const MIN_NODES_FOR_STAGGER = 3;

export function revealDiagram(diagram: Diagram): void {
  const { nodes, edges } = diagramToFlow(diagram);
  const store = useEditorStore.getState();

  if (nodes.length < MIN_NODES_FOR_STAGGER) {
    store.loadDiagram(diagram);
    return;
  }

  // Clear canvas + history, then feed nodes in order.
  useEditorStore.setState({
    nodes: [],
    edges: [],
    selectedNodeId: null,
    selectedEdgeId: null,
    past: [],
    future: [],
    canUndo: false,
    canRedo: false,
  });

  let i = 0;
  const step = () => {
    const next = nodes.slice(0, Math.min(nodes.length, i + BATCH));
    useEditorStore.setState({ nodes: next });
    i += BATCH;
    if (i < nodes.length) {
      setTimeout(step, STAGGER_MS);
    } else {
      // All nodes in — drop edges in one shot.
      useEditorStore.setState({ edges });
    }
  };
  step();
}
